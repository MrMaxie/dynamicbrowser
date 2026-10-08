package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigPath(t *testing.T) {
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(filepath.Dir(exe), "config.yaml") {
		t.Fatalf("path = %q", got)
	}
}

func TestWatchConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	store := &Store{}
	ctx, cancel := context.WithCancel(context.Background())
	done, err := Watch(ctx, path, store)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() { cancel(); <-done })
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	write := func(path, data string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	wait := func(want string, wantErr bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			store.mu.RLock()
			got := store.values["custom"]
			hasErr := store.err != nil
			store.mu.RUnlock()
			if got == want && hasErr == wantErr {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("config did not reach %q, error=%v", want, wantErr)
	}
	write(path, "custom: first\nanything:\n  list: [1, 2]\n")
	wait("first", false)
	write(path, "invalid: [")
	wait("first", true)
	replacement := filepath.Join(filepath.Dir(path), "replacement.yaml")
	write(replacement, "custom: second\n")
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	wait("second", false)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	wait("second", true)
	write(path, "custom: third\n")
	wait("third", false)
}

func TestConfigChanges(t *testing.T) {
	store := &Store{}
	changes := store.Changes()
	path := filepath.Join(t.TempDir(), "config.yaml")
	store.reload(path)
	select {
	case <-changes:
	default:
		t.Fatal("missing error notification")
	}
	if store.Status() != "Configuration: error (last valid configuration retained)" {
		t.Fatal("missing configuration error status")
	}
	if err := os.WriteFile(path, []byte("custom: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	store.reload(path)
	store.reload(path)
	select {
	case <-changes:
	default:
		t.Fatal("missing reload notification")
	}
	select {
	case <-changes:
		t.Fatal("reload notifications were not coalesced")
	default:
	}
	if store.Status() != "Configuration: loaded" {
		t.Fatal("configuration did not recover")
	}
}

func TestConfigPreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	const content = "arbitrary: true\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	store := &Store{}
	done, err := Watch(ctx, path, store)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()
	<-done
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content || store.values["arbitrary"] != true {
		t.Fatal("existing config changed or was not loaded")
	}
}
