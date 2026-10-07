package app

import (
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	errRestartNotReady = errors.New("running instance has not enabled restart control")
	terminateInstance  = func() { os.Exit(0) }
)

const (
	restartTimeout = 5 * time.Second
	restartGrace   = 2 * time.Second
)

func acquireForStartup(force bool) (func(), bool, error) {
	deadline := time.Now().Add(restartTimeout)
	// Reserve the handoff so ordinary launches cannot steal a restarting owner's lock.
	for {
		release, acquired, err := acquireNamedInstance(instanceName + ".restart")
		if err != nil {
			return nil, false, err
		}
		if acquired {
			defer release()
			break
		}
		if !force {
			return nil, false, nil
		}
		if time.Now().After(deadline) {
			return nil, false, errors.New("another restart is still in progress")
		}
		time.Sleep(25 * time.Millisecond)
	}
	requested := false
	for {
		release, acquired, err := acquireInstance()
		if err != nil || acquired || !force {
			return release, acquired, err
		}
		if !requested {
			if err := requestInstanceRestart(); err == nil {
				requested = true
			} else if !errors.Is(err, errRestartNotReady) {
				return nil, false, fmt.Errorf("request restart: %w", err)
			}
		}
		if time.Now().After(deadline) {
			return nil, false, errors.New("restart timed out; close the running instance manually and try again")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// The control adapters serialize requests and stop the watchdog on their own goroutine.
func restartAction(quit func()) (request, stop func()) {
	var timer *time.Timer
	return func() {
			if timer == nil {
				// Only the existing instance terminates itself; no PID/name-based process killing.
				timer = time.AfterFunc(restartGrace, terminateInstance)
				quit()
			}
		}, func() {
			if timer != nil {
				timer.Stop()
			}
		}
}
