//go:build linux || darwin

package config

import (
	"os"
	"path/filepath"
)

// LinuxConfigHome follows XDG's fallback for unset or relative paths.
func LinuxConfigHome() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(base) {
		return base, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}
