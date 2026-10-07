package app

import "testing"

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
