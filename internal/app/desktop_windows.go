package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/MrMaxie/dynamicbrowser/internal/routing"
	"golang.org/x/sys/windows"
)

func captureSource() routing.Source {
	hwnd := windows.GetForegroundWindow()
	if hwnd == 0 {
		return routing.Source{}
	}
	source := routing.Source{Window: nativeWindowText(hwnd)}
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(hwnd, &pid); err != nil {
		return source
	}
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return source
	}
	defer func() { _ = windows.CloseHandle(process) }()
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(process, 0, &buffer[0], &size); err == nil {
		name := filepath.Base(windows.UTF16ToString(buffer[:size]))
		source.Process = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return source
}

func startCommand(executable string, args []string) error {
	command := exec.Command(executable, args...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}

func editConfiguration(path string) error {
	file, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, nil, file, nil, nil, windows.SW_SHOWNORMAL)
}
