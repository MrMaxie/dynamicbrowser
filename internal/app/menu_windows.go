package app

import (
	"fmt"

	"github.com/gogpu/systray"
	"golang.org/x/sys/windows"
)

func ownedTrayWindow() (windows.HWND, error) {
	var after windows.HWND
	for {
		hwnd, err := findNativeWindow(messageOnlyWindow, after, systrayClass)
		if err != nil {
			return 0, err
		}
		if hwnd == 0 {
			return 0, fmt.Errorf("systray window not found on the current thread")
		}
		var pid uint32
		thread, err := windows.GetWindowThreadProcessId(hwnd, &pid)
		if err == nil && pid == windows.GetCurrentProcessId() && thread == windows.GetCurrentThreadId() {
			return hwnd, nil
		}
		after = hwnd
	}
}

func bindTrayMenu(tray *systray.SystemTray) error {
	hwnd, err := ownedTrayWindow()
	if err != nil {
		return err
	}
	tray.OnClick(func() {
		if err := postNativeMessage(hwnd, systrayCallback, 0, wmContextMenu); err != nil {
			tray.SetTooltip("dynamicbrowser — unable to open menu")
		}
	})
	return nil
}
