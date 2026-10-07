package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

type testAppProcess struct {
	cmd  *exec.Cmd
	done chan struct{}
	err  error
}

func startTestApp(t *testing.T, binary string, args ...string) *testAppProcess {
	t.Helper()
	process := &testAppProcess{cmd: exec.Command(binary, args...), done: make(chan struct{})}
	if err := process.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		process.err = process.cmd.Wait()
		close(process.done)
	}()
	t.Cleanup(func() {
		_ = process.cmd.Process.Kill()
		<-process.done
	})
	return process
}

func TestForceCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "dynamicbrowser.exe")
	flags := "-H=windowsgui -X main.version=test-version -X github.com/MrMaxie/dynamicbrowser/internal/app.instanceName=" + instanceName
	if output, err := exec.Command("go", "build", "-ldflags", flags, "-o", binary, "../../cmd/dynamicbrowser").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	path := filepath.Join(filepath.Dir(binary), "config.toml")
	const config = "[arbitrary]\nname = 'preserved'\n"
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}

	t.Run("force without an owner starts normally", func(t *testing.T) {
		process := startTestApp(t, binary, "--force")
		waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(process.cmd.Process.Pid))
	})

	t.Run("force replaces the owner and preserves singleton", func(t *testing.T) {
		owner := startTestApp(t, binary)
		waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(owner.cmd.Process.Pid))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary, "--force", "--version").CombinedOutput()
		if err != nil || string(output) != "test-version\n" {
			t.Fatalf("version with force: %v, %q", err, output)
		}
		select {
		case <-owner.done:
			t.Fatal("--version unexpectedly stopped the running instance")
		default:
		}
		replacement := startTestApp(t, binary, "--force")
		waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(replacement.cmd.Process.Pid))
		select {
		case <-owner.done:
			if owner.err != nil {
				t.Fatalf("old instance exit: %v", owner.err)
			}
		case <-time.After(time.Second):
			t.Fatal("old instance remained alive after replacement startup")
		}
		output, err = exec.CommandContext(ctx, binary).CombinedOutput()
		if err != nil || len(output) != 0 {
			t.Fatalf("duplicate after force: %v, %q", err, output)
		}
		select {
		case <-replacement.done:
			t.Fatalf("replacement unexpectedly exited: %v", replacement.err)
		default:
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != config {
			t.Fatalf("force changed config.toml: %v, %q", err, data)
		}
	})

	t.Run("concurrent force requests leave one owner", func(t *testing.T) {
		owner := startTestApp(t, binary)
		waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(owner.cmd.Process.Pid))
		first := startTestApp(t, binary, "--force")
		second := startTestApp(t, binary, "--force")
		deadline := time.Now().Add(restartTimeout + 2*time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-owner.done:
				for _, pair := range [][2]*testAppProcess{{first, second}, {second, first}} {
					select {
					case <-pair[0].done:
						if pair[0].err != nil {
							t.Fatalf("replaced forced launch: %v", pair[0].err)
						}
						var after windows.HWND
						for {
							hwnd, err := findNativeWindow(messageOnlyWindow, after, systrayClass)
							if err != nil {
								t.Fatal(err)
							}
							if hwnd == 0 {
								break
							}
							var pid uint32
							_, err = windows.GetWindowThreadProcessId(hwnd, &pid)
							if err == nil && pid == uint32(pair[1].cmd.Process.Pid) {
								return
							}
							after = hwnd
						}
					default:
					}
				}
			default:
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatal("concurrent forced launches did not settle on a single live owner")
	})
}
