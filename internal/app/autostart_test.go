package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRoutingSurvivesAutostartFailure(t *testing.T) {
	failure := errors.New("startup registration denied")
	opened := false
	var warnings bytes.Buffer
	err := openWithAutostart(true, func(bool) error { return failure }, &warnings, func() error { opened = true; return nil })
	if !opened || err != nil || !strings.Contains(warnings.String(), failure.Error()) {
		t.Fatalf("opened=%t, error=%v, warnings=%q", opened, err, warnings.String())
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
