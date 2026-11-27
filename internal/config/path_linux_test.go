package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	home := t.TempDir()
	xdg := t.TempDir()
	t.Setenv("HOME", home)
	for _, tt := range []struct{ name, value, base string }{
		{"empty", "", filepath.Join(home, ".config")},
		{"relative", "relative/config", filepath.Join(home, ".config")},
		{"absolute", xdg, xdg},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.value)
			got, err := Path()
			if want := filepath.Join(tt.base, "dynamicbrowser", "config.yaml"); err != nil || got != want {
				t.Fatalf("Path() = %q, %v; want %q", got, err, want)
			}
		})
	}
}

func TestInitializeXDGConfig(t *testing.T) {
	for _, mode := range []string{"load", "watch"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
			path, err := Path()
			if err != nil {
				t.Fatal(err)
			}
			if mode == "load" {
				if _, err := Load(path); err != nil {
					t.Fatal(err)
				}
			} else {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				store := &Store{}
				done, err := Watch(ctx, path, store)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { cancel(); <-done })
				if _, err := store.Snapshot(); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range []string{path, filepath.Join(filepath.Dir(path), "config.schema.json")} {
				if _, err := os.Stat(file); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
