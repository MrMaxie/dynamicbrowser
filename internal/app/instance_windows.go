package app

import (
	"errors"

	"golang.org/x/sys/windows"
)

const instanceName = `Local\dynamicbrowser.MrMaxie`

// Keeping a handle open keeps the named mutex alive, including after a crash
// until Windows closes the process handles. No lock file can become stale.
func acquireInstance() (func(), bool, error) {
	name, err := windows.UTF16PtrFromString(instanceName)
	if err != nil {
		return nil, false, err
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		_ = windows.CloseHandle(handle)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return func() { _ = windows.CloseHandle(handle) }, true, nil
}

// ShowStartupError reports a startup failure without opening a console.
func ShowStartupError(err error) {
	text, _ := windows.UTF16PtrFromString(err.Error())
	caption, _ := windows.UTF16PtrFromString("dynamicbrowser")
	_, _ = windows.MessageBox(0, text, caption, windows.MB_OK|windows.MB_ICONERROR)
}
