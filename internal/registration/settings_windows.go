package registration

import (
	"errors"
	"fmt"
	"os/exec"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func (windowsPlatform) describe(id string) (string, error) {
	if id == "" {
		return "", errors.New("no previous handler was saved")
	}
	key, err := registry.OpenKey(registry.CLASSES_ROOT, id, registry.QUERY_VALUE)
	if err != nil {
		return "", errors.New("the saved handler is no longer registered")
	}
	_ = key.Close()
	executable, executableErr := queryAssociation(assocFixedProgID|assocVerify, assocExecutable, id)
	if executableErr == nil && executable != "" {
		if _, err := exec.LookPath(executable); err != nil {
			return "", errors.New("the saved browser executable is no longer available")
		}
	} else if delegate, err := queryAssociation(assocFixedProgID, assocDelegateExecute, id); err != nil || delegate == "" {
		return "", errors.New("the saved handler cannot be launched")
	}
	name, err := queryAssociation(assocFixedProgID|assocVerify, assocFriendlyAppName, id)
	if err != nil || name == "" {
		return "", errors.New("the saved browser cannot be identified by Windows; choose a browser manually in Windows Settings")
	}
	return name, nil
}

func (windowsPlatform) assist(message string, registration bool) (bool, error) {
	location := "ms-settings:defaultapps"
	if registration {
		location += "?registeredAppUser=dynamicbrowser"
	}
	uri, err := windows.UTF16PtrFromString(location)
	if err != nil {
		return false, err
	}
	if err := windows.ShellExecute(0, nil, uri, nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return false, fmt.Errorf("open Windows Default Apps settings: %w", err)
	}
	text, err := windows.UTF16PtrFromString(message)
	if err != nil {
		return false, err
	}
	title, _ := windows.UTF16PtrFromString(applicationName)
	button, err := windows.MessageBox(0, text, title, windows.MB_OKCANCEL|windows.MB_ICONINFORMATION)
	return button == 1, err
}
