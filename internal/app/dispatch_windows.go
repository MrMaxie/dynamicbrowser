package app

import (
	"context"
	"errors"

	"github.com/gogpu/systray"
	"golang.org/x/sys/windows"
)

// SetMenu is not synchronized by systray on Windows; rebuild on the window thread,
// after a native popup has finished borrowing the previous HMENU.
func newTrayDispatch(_ *systray.SystemTray, ctx context.Context) (func(func()) error, func(), error) {
	hwnd, err := ownedTrayWindow()
	if err != nil {
		return nil, nil, err
	}
	queue := trayDispatchQueue{jobs: make(chan func(), 1)}
	const dispatchMessage = 0x8000 + 51
	comctl := windows.NewLazySystemDLL("comctl32.dll")
	setSubclass := comctl.NewProc("SetWindowSubclass")
	removeSubclass := comctl.NewProc("RemoveWindowSubclass")
	defSubclass := comctl.NewProc("DefSubclassProc")
	menuDepth := 0
	drain := func() {
		if menuDepth != 0 {
			return
		}
		select {
		case job := <-queue.jobs:
			job()
		default:
		}
	}
	callback := windows.NewCallback(func(window uintptr, message uint32, wParam, lParam, id, data uintptr) uintptr {
		if message == dispatchMessage {
			drain()
			return 0
		}
		popup := message == systrayCallback && (lParam == wmContextMenu || lParam == 0x0205)
		if popup {
			menuDepth++
		}
		result, _, _ := defSubclass.Call(window, uintptr(message), wParam, lParam)
		if popup {
			menuDepth--
			drain()
		}
		return result
	})
	if ok, _, _ := setSubclass.Call(uintptr(hwnd), callback, 1, 0); ok == 0 {
		return nil, nil, errors.New("unable to attach tray UI dispatcher")
	}
	dispatch := func(job func()) error {
		return dispatchTask(ctx, func(run func()) error {
			return queue.submit(ctx, run, func() error {
				return postNativeMessage(hwnd, dispatchMessage, 0, 0)
			})
		}, job)
	}
	return dispatch, func() { _, _, _ = removeSubclass.Call(uintptr(hwnd), callback, 1) }, nil
}
