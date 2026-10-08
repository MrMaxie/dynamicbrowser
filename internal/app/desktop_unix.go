//go:build linux || darwin

package app

import (
	"os/exec"
	"runtime"

	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func captureSource() routing.Source {
	return routing.Source{}
}

func startCommand(executable string, args []string) error {
	command := exec.Command(executable, args...)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}

func editConfiguration(path string) error {
	if runtime.GOOS == "darwin" {
		return startCommand("open", []string{path})
	}
	return startCommand("xdg-open", []string{path})
}
