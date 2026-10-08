package autostart

import (
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func TestOwnedAutostartValue(t *testing.T) {
	identity := fmt.Sprintf("dynamicbrowser-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	executable := `C:\Program Files\Example\dynamicbrowser.exe`
	t.Cleanup(func() { _ = Set(identity, executable, false) })
	if err := Set(identity, executable, true); err != nil {
		t.Fatal(err)
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = key.Close() }()
	got, _, err := key.GetStringValue(identity)
	if err != nil || got != windows.EscapeArg(executable) {
		t.Fatalf("startup command = %q, %v", got, err)
	}
	for range 2 {
		if err := Set(identity, executable, false); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := key.GetStringValue(identity); err != registry.ErrNotExist {
		t.Fatalf("owned value was not removed: %v", err)
	}
}
