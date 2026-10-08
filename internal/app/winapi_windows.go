package app

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
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
	findWindowOnce   sync.Once
	findWindowCall   func(windows.HWND, windows.HWND, *uint16, *uint16) windows.HWND
	postMessage      = user32API.NewProc("PostMessageW")
	getWindowText    = user32API.NewProc("GetWindowTextW")
	windowTextOnce   sync.Once
	windowTextCall   func(windows.HWND, *uint16, int32) int32
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
	findWindowOnce.Do(func() { purego.RegisterFunc(&findWindowCall, findWindowEx.Addr()) })
	hwnd := findWindowCall(parent, after, &name[0], nil)
	runtime.KeepAlive(name)
	return hwnd, nil
}

func postNativeMessage(hwnd windows.HWND, message uint32, wParam, lParam uintptr) error {
	ok, _, err := postMessage.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	if ok == 0 {
		return fmt.Errorf("PostMessageW: %w", err)
	}
	return nil
}

func nativeWindowText(hwnd windows.HWND) string {
	windowTextOnce.Do(func() { purego.RegisterFunc(&windowTextCall, getWindowText.Addr()) })
	buffer := make([]uint16, 32768)
	length := windowTextCall(hwnd, &buffer[0], int32(len(buffer)))
	runtime.KeepAlive(buffer)
	if length < 0 || int(length) >= len(buffer) {
		return ""
	}
	return windows.UTF16ToString(buffer[:length])
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
