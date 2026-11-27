package app

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func captureSource() routing.Source {
	// shortcut: procfs only identifies the direct launcher, not the app behind a desktop broker; use broker-provided metadata when available.
	source := routing.Source{RequireKnown: true}
	pid := os.Getppid()
	if pid > 1 {
		path, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "exe"))
		if err == nil && os.Getppid() == pid {
			source.Process = launcherProcess(path)
		}
	}
	return source
}

func activeWindowHint() string {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return ""
	}
	// shortcut: focus can change before the handler starts; replace this hint if activation provides reliable attribution.
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	data, err := exec.CommandContext(ctx, "hyprctl", "-j", "activewindow").Output()
	if err != nil {
		return ""
	}
	var window struct{ Class string }
	if err := json.Unmarshal(data, &window); err != nil {
		return ""
	}
	return window.Class
}

func launcherProcess(path string) string {
	name := filepath.Base(strings.TrimSuffix(path, " (deleted)"))
	switch name {
	case "dynamicbrowser", "xdg-open", "gio", "kioclient", "kioclient5", "kioclient6", "kde-open", "kde-open5", "kde-open6",
		"xdg-desktop-portal", "systemd", "dbus-daemon", "dbus-broker", "sh", "bash", "dash", "zsh", "fish", "busybox":
		return ""
	}
	if strings.HasPrefix(name, "xdg-desktop-portal-") {
		return ""
	}
	return name
}
