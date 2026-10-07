package app

import (
	"fmt"

	"github.com/gogpu/systray"
	"golang.org/x/sys/windows"
)

func bindTrayMenu(tray *systray.SystemTray) error {
	var after windows.HWND
	for {
		hwnd, err := findNativeWindow(messageOnlyWindow, after, systrayClass)
		if err != nil {
			return err
		}
		if hwnd == 0 {
			return fmt.Errorf("systray window not found on the current thread")
		}
		var pid uint32
		thread, err := windows.GetWindowThreadProcessId(hwnd, &pid)
		if err == nil && pid == windows.GetCurrentProcessId() && thread == windows.GetCurrentThreadId() {
			tray.OnClick(func() {
				if err := postNativeMessage(hwnd, systrayCallback, 0, wmContextMenu); err != nil {
					tray.SetTooltip("dynamicbrowser — unable to open menu")
				}
			})
			return nil
		}
		after = hwnd
	}
}
