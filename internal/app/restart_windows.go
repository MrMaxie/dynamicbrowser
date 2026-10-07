package app

import (
	"errors"
	"sync"

	"golang.org/x/sys/windows"
)

func restartEventName() (*uint16, error) {
	return windows.UTF16PtrFromString(`Local\` + instanceName + ".shutdown")
}

func startInstanceControl(quit func()) (func(), error) {
	name, err := restartEventName()
	if err != nil {
		return nil, err
	}
	requestEvent, err := windows.CreateEvent(nil, 0, 0, name)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return nil, err
	}
	stopEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		_ = windows.CloseHandle(requestEvent)
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		request, stop := restartAction(quit)
		defer stop()
		for {
			event, err := windows.WaitForMultipleObjects([]windows.Handle{stopEvent, requestEvent}, false, windows.INFINITE)
			if err != nil || event == windows.WAIT_OBJECT_0 {
				return
			}
			if event == windows.WAIT_OBJECT_0+1 {
				request()
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			_ = windows.SetEvent(stopEvent)
			<-done
			// Never close a handle while WaitForMultipleObjects is still borrowing it.
			_ = windows.CloseHandle(requestEvent)
			_ = windows.CloseHandle(stopEvent)
		})
	}, nil
}

func requestInstanceRestart() error {
	name, err := restartEventName()
	if err != nil {
		return err
	}
	event, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, name)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return errRestartNotReady
	}
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(event) }()
	return windows.SetEvent(event)
}
