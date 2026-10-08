package app

import (
	"errors"
	"testing"
)

func TestRoutingSurvivesAutostartFailure(t *testing.T) {
	failure := errors.New("startup registration denied")
	opened := false
	err := openWithAutostart(true, func(bool) error { return failure }, func() error { opened = true; return nil })
	if !opened || !errors.Is(err, failure) {
		t.Fatal("registration failure blocked routing or was discarded")
	}
}

func TestAutostartRetriesFailure(t *testing.T) {
	calls := 0
	update := autostartUpdater(func(enabled bool) error {
		calls++
		if calls == 1 {
			return errors.New("startup registration denied")
		}
		return nil
	})
	if err := update(true); err == nil {
		t.Fatal("registration failure hidden")
	}
	for range 2 {
		if err := update(true); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("registration attempts = %d", calls)
	}
	if err := update(false); err != nil || calls != 3 {
		t.Fatalf("disable: %v, attempts=%d", err, calls)
	}
}
