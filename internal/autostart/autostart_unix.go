//go:build linux || darwin

package autostart

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Set(identity, executable string, enabled bool) error {
	var path string
	var content []byte
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, "Library", "LaunchAgents", identity+".plist")
		content = launchAgent(identity, executable)
	} else {
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		path = filepath.Join(dir, "autostart", identity+".desktop")
		content = desktopEntry(executable)
	}
	if !enabled {
		err := os.Remove(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0600)
}

func desktopEntry(executable string) []byte {
	escape := strings.NewReplacer("\\", "\\\\\\\\", "\"", "\\\\\\\"", "$", "\\\\$", "`", "\\\\`", "%", "%%", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	return []byte("[Desktop Entry]\nType=Application\nName=dynamicbrowser\nExec=\"" + escape.Replace(executable) + "\"\nTerminal=false\n")
}

func launchAgent(identity, executable string) []byte {
	var id, exe bytes.Buffer
	_ = xml.EscapeText(&id, []byte(identity))
	_ = xml.EscapeText(&exe, []byte(executable))
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>Label</key><string>` + id.String() + `</string><key>ProgramArguments</key><array><string>` + exe.String() + `</string></array><key>RunAtLoad</key><true/></dict></plist>`)
}
