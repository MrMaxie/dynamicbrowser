//go:build !linux

package config

import (
	"os"
	"path/filepath"
	"testing"
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
	if want := filepath.Join(filepath.Dir(exe), "config.yaml"); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
