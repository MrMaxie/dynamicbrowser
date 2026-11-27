package config

import (
	"os"
	"path/filepath"
)

// Path returns the per-user XDG configuration path.
func Path() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "dynamicbrowser", "config.yaml"), nil
}
