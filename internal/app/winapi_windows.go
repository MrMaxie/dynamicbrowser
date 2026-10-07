package app

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	messageOnlyWindow = windows.HWND(^uintptr(2))
	wmClose           = 0x0010
	wmCancelMode      = 0x001F
	wmLeftButtonUp    = 0x0202
	wmContextMenu     = 0x007B
	systrayClass      = "GoGPUSystrayMsg"
	// Private callback protocol of gogpu/systray v0.3.0; covered by the native menu test.
	systrayCallback = 0x0400 + 100
)

var (
	user32API        = windows.NewLazySystemDLL("user32.dll")
	findWindowEx     = user32API.NewProc("FindWindowExW")
	postMessage      = user32API.NewProc("PostMessageW")
	getSystemMetrics = user32API.NewProc("GetSystemMetrics")
	getWindowDPI     = user32API.NewProc("GetWindowDpiAwarenessContext")
	getDPIAwareness  = user32API.NewProc("GetAwarenessFromDpiAwarenessContext")
	setThreadDPI     = user32API.NewProc("SetThreadDpiAwarenessContext")
)

func findNativeWindow(parent, after windows.HWND, class string) (windows.HWND, error) {
	name, err := windows.UTF16FromString(class)
	if err != nil {
		return 0, err
	}
	// FindWindowExW borrows this UTF-16 buffer only for the duration of the call.
	hwnd, _, _ := findWindowEx.Call(uintptr(parent), uintptr(after), uintptr(unsafe.Pointer(&name[0])), 0)
	runtime.KeepAlive(name)
	return windows.HWND(hwnd), nil
}

func postNativeMessage(hwnd windows.HWND, message uint32, wParam, lParam uintptr) error {
	ok, _, err := postMessage.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	if ok == 0 {
		return fmt.Errorf("PostMessageW: %w", err)
	}
	return nil
}

func nativeTrayIconSize() int {
	width, _, _ := getSystemMetrics.Call(49)
	height, _, _ := getSystemMetrics.Call(50)
	return max(int(width), int(height), 16)
}

func nativeWindowAwareness(hwnd windows.HWND) int {
	context, _, _ := getWindowDPI.Call(uintptr(hwnd))
	awareness, _, _ := getDPIAwareness.Call(context)
	return int(awareness)
}

func setPerMonitorThreadDPI() (func(), error) {
	previous, _, err := setThreadDPI.Call(^uintptr(3))
	if previous == 0 {
		return nil, err
	}
	return func() { _, _, _ = setThreadDPI.Call(previous) }, nil
}
