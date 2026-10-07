package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

const helperOwnerEnv = "DYNAMICBROWSER_TEST_OWNER"

func TestUnresponsiveInstanceHelper(t *testing.T) {
	identity := os.Getenv(helperOwnerEnv)
	if identity == "" {
		return
	}
	instanceName = identity
	terminateInstance = func() { os.Exit(0) }
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("helper acquire: %v, %v", acquired, err)
	}
	defer release()
	stop, err := startInstanceControl(func() {})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	fmt.Println("ready")
	<-time.After(15 * time.Second)
	t.Fatal("unresponsive helper was not terminated")
}

func TestForceStartupTerminatesUnresponsiveOwner(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), restartTimeout+3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestUnresponsiveInstanceHelper$", "-test.timeout=15s")
	cmd.Env = append(os.Environ(), helperOwnerEnv+"="+instanceName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || ready != "ready\n" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("helper readiness: %v, %q", err, ready)
	}
	started := time.Now()
	release, acquired, err := acquireForStartup(true)
	if err != nil || !acquired {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("force replacement: %v, %v", acquired, err)
	}
	defer release()
	if err := cmd.Wait(); err != nil {
		t.Fatalf("unresponsive owner exit: %v", err)
	}
	if time.Since(started) < restartGrace {
		t.Fatal("unresponsive owner exited before the watchdog deadline")
	}
	_, duplicate, err := acquireInstance()
	if err != nil || duplicate {
		t.Fatal("replacement did not retain exclusive ownership")
	}
}
