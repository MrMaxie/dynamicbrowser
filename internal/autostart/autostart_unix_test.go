//go:build linux || darwin

package autostart

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStartupFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	identity := "dynamicbrowser-test"
	if err := Set(identity, "/example/app with spaces", true); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "autostart", identity+".desktop")
	if runtime.GOOS == "darwin" {
		path = filepath.Join(home, "Library", "LaunchAgents", identity+".plist")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Set(identity, "", false); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestStartupEscaping(t *testing.T) {
	entry := string(desktopEntry("/example/quotes\"$`%\\\napp"))
	if strings.Contains(entry, "\napp") || !strings.Contains(entry, "%%") {
		t.Fatalf("unsafe desktop entry: %q", entry)
	}
	plist := launchAgent("test<&", "/example/a&b\" app")
	var doc any
	if err := xml.NewDecoder(bytes.NewReader(plist)).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(plist, []byte("a&b")) {
		t.Fatal("unescaped XML")
	}
}
