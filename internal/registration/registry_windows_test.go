package registration

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func requireSandbox(t *testing.T) {
	t.Helper()
	if os.Getenv("DYNAMICBROWSER_SANDBOX_REGISTRATION") != "1" || os.Getenv("USERNAME") != "WDAGUtilityAccount" {
		t.Skip("registry integration is restricted to Windows Sandbox")
	}
}

func TestWindowsRegistrationAndSnapshot(t *testing.T) {
	requireSandbox(t)
	before, err := currentDefaults()
	if err != nil || before.any(ownDefaults) {
		t.Fatalf("unexpected initial associations: %v", err)
	}
	platform := windowsPlatform{}
	if _, saved, err := platform.previous(); err != nil || saved {
		t.Fatalf("Sandbox already has setup state: %v", err)
	}
	t.Cleanup(func() {
		if err := platform.remove(); err != nil {
			t.Error(err)
		}
	})
	if err := platform.savePrevious(before); err != nil {
		t.Fatal(err)
	}
	if err := platform.savePrevious(ownDefaults); err == nil {
		t.Fatal("overwrote an existing snapshot")
	}
	if err := platform.install(); err != nil {
		t.Fatal(err)
	}
	for _, scheme := range []string{ownDefaults.HTTP, ownDefaults.HTTPS} {
		key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\`+scheme+`\shell\open\command`, registry.QUERY_VALUE)
		if err != nil {
			t.Fatal(err)
		}
		command, _, err := key.GetStringValue("")
		_ = key.Close()
		if err != nil || !strings.HasSuffix(command, ` -- "%1"`) {
			t.Fatalf("unsafe URI activation command: %v", err)
		}
	}
	after, err := currentDefaults()
	if err != nil || !after.matches(before) {
		t.Fatalf("publication changed default choices: %v", err)
	}
	previous, saved, err := platform.previous()
	if err != nil || !saved || !previous.matches(before) {
		t.Fatalf("snapshot was not retained: %v", err)
	}
	if _, err := platform.describe(before.HTTP); err != nil {
		t.Fatal(err)
	}
	if _, err := platform.describe(before.HTTPS); err != nil {
		t.Fatal(err)
	}
	if err := platform.remove(); err != nil {
		t.Fatal(err)
	}
	if _, saved, err := platform.previous(); err != nil || saved {
		t.Fatalf("snapshot remained after removal: %v", err)
	}
	if published, err := publication(); err != nil || published != "" {
		t.Fatalf("publication remained: %v", err)
	}
	for _, path := range []string{clientKey, `Software\Classes\` + ownDefaults.HTTP, `Software\Classes\` + ownDefaults.HTTPS} {
		key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
		if err == nil {
			_ = key.Close()
			t.Fatal("owned registration remained")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
	}
}

func TestWindowsSnapshotRejectsCorruption(t *testing.T) {
	requireSandbox(t)
	valid, err := json.Marshal(snapshot{Version: 1, Previous: defaults{HTTP: "example.http", HTTPS: "example.https"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{`{`, `{"version":2}`, `{"version":1}`, `{"version":1,"previous":null}`, `{"version":1,"unexpected":true}`, string(valid) + ` {}`} {
		if err := writeRegistrationValues(stateKey, map[string]string{"Previous": input}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := (windowsPlatform{}).previous(); err == nil {
			t.Fatal("accepted corrupted snapshot")
		}
	}
	if err := (windowsPlatform{}).remove(); err != nil {
		t.Fatal(err)
	}
}
