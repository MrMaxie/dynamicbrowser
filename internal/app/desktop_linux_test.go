package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLinuxRoutingWithoutTrayHost(t *testing.T) {
	busTool, err := exec.LookPath("dbus-run-session")
	if err != nil {
		t.Skip("dbus-run-session is unavailable")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "dynamicbrowser")
	flags := "-X github.com/MrMaxie/dynamicbrowser/internal/app.instanceName=" + instanceName
	if out, err := exec.Command("go", "build", "-ldflags", flags, "-o", binary, "../../cmd/dynamicbrowser").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := filepath.Join(dir, "home")
	configDir := filepath.Join(home, ".config", "dynamicbrowser")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("autostart: true\ndefault: probe\nbrowsers:\n  probe: {exe: /bin/true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_RUNTIME_DIR", "/tmp/"+strings.Repeat("r", 120))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, strings.Repeat("c", 120)))
	t.Setenv("XDG_CONFIG_HOME", "relative")
	t.Setenv("HOME", home)
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, busTool, "--", binary)
	command.WaitDelay = time.Second
	out, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "tray is unavailable on Linux") {
		t.Fatalf("headless tray: %v, %s", err, out)
	}
	out, err = exec.CommandContext(ctx, binary, "https://example.test").CombinedOutput()
	if err != nil || len(out) != 0 {
		t.Fatalf("direct routing: %v, %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "autostart", instanceName+".desktop")); err != nil {
		t.Fatal(err)
	}
}
