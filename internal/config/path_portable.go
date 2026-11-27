//go:build !linux

package config

import (
	"os"
	"path/filepath"
)

// Path returns config.yaml next to the running executable.
func Path() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "config.yaml"), nil
}
