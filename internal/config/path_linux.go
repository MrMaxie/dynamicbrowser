package config

import "path/filepath"

// Path returns the per-user XDG configuration path.
func Path() (string, error) {
	base, err := LinuxConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "dynamicbrowser", "config.yaml"), nil
}
