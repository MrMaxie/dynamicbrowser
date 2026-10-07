package app

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRestartControl(t *testing.T) {
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("acquire: %v, %v", acquired, err)
	}
	defer release()
	called := make(chan struct{}, 1)
	var calls atomic.Int32
	stop, err := startInstanceControl(func() {
		calls.Add(1)
		called <- struct{}{}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := requestInstanceRestart(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("restart did not reach the owning instance")
	}
	if err := requestInstanceRestart(); err != nil {
		t.Fatal(err)
	}
	stop()
	stop()
	if calls.Load() != 1 {
		t.Fatalf("shutdown called %d times", calls.Load())
	}
	if err := requestInstanceRestart(); !errors.Is(err, errRestartNotReady) {
		t.Fatalf("closed control: %v", err)
	}
}

func TestRestartControlIdentityIsolation(t *testing.T) {
	called := make(chan struct{}, 1)
	stop, err := startInstanceControl(func() { called <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	original := instanceName
	instanceName += ".unrelated"
	err = requestInstanceRestart()
	instanceName = original
	if !errors.Is(err, errRestartNotReady) {
		t.Fatalf("unrelated control: %v", err)
	}
	select {
	case <-called:
		t.Fatal("restart targeted an unrelated identity")
	default:
	}
}

func TestOrdinaryStartupDropsDuringRestart(t *testing.T) {
	release, acquired, err := acquireNamedInstance(instanceName + ".restart")
	if err != nil || !acquired {
		t.Fatalf("restart gate: %v, %v", acquired, err)
	}
	defer release()
	releaseOwner, acquired, err := acquireForStartup(false)
	if acquired {
		defer releaseOwner()
	}
	if err != nil || acquired {
		t.Fatalf("ordinary startup during restart: %v, %v", acquired, err)
	}
}

func TestForceStartupWithoutOwner(t *testing.T) {
	release, acquired, err := acquireForStartup(true)
	if err != nil || !acquired {
		t.Fatalf("force startup: %v, %v", acquired, err)
	}
	defer release()
	_, duplicate, err := acquireInstance()
	if err != nil || duplicate {
		t.Fatalf("forced instance lost its singleton lock: %v, %v", duplicate, err)
	}
}

func TestForceStartupWaitsForOwner(t *testing.T) {
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("acquire: %v, %v", acquired, err)
	}
	release = sync.OnceFunc(release)
	defer release()
	stop, err := startInstanceControl(release)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	replacement, acquired, err := acquireForStartup(true)
	if err != nil || !acquired {
		t.Fatalf("replacement: %v, %v", acquired, err)
	}
	defer replacement()
}

func TestForceStartupMissingControlTimesOut(t *testing.T) {
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("acquire: %v, %v", acquired, err)
	}
	defer release()
	started := time.Now()
	_, acquired, err = acquireForStartup(true)
	if err == nil || acquired {
		t.Fatal("force must not start a second instance when restart control is unavailable")
	}
	if time.Since(started) > restartTimeout+time.Second {
		t.Fatal("force startup exceeded its bounded timeout")
	}
}
