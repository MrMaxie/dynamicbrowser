package app

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Test subprocesses must not contend with the user's running application.
	instanceName = fmt.Sprintf("dynamicbrowser.test.%d", os.Getpid())
	terminateInstance = func() { panic("restart watchdog must not terminate the test runner") }
	os.Exit(m.Run())
}

func TestSingleton(t *testing.T) {
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("first instance: acquired=%v, err=%v", acquired, err)
	}
	_, acquired, err = acquireInstance()
	if err != nil || acquired {
		release()
		t.Fatalf("duplicate instance: acquired=%v, err=%v", acquired, err)
	}
	release()
	release, acquired, err = acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("after release: acquired=%v, err=%v", acquired, err)
	}
	release()
}
