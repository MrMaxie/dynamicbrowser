//go:build linux || darwin

package app

import (
	"os/exec"
	"runtime"
)

func startCommand(executable string, args []string) error {
	command := exec.Command(executable, args...)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}

func showAbout(version string) error {
	message := aboutText(version)
	if runtime.GOOS == "darwin" {
		return exec.Command("osascript", "-e", `on run argv
 display dialog (item 1 of argv) with title "About dynamicbrowser" buttons {"OK"} default button "OK"
end run`, "--", message).Run()
	}
	if err := exec.Command("zenity", "--info", "--title=About dynamicbrowser", "--no-markup", "--text="+message).Run(); err == nil {
		return nil
	}
	return exec.Command("kdialog", "--title", "About dynamicbrowser", "--msgbox", message).Run()
}

func editConfiguration(path string) error {
	if runtime.GOOS == "darwin" {
		return startCommand("open", []string{path})
	}
	return startCommand("xdg-open", []string{path})
}
