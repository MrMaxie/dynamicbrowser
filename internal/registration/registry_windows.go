package registration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type windowsPlatform struct{}

type snapshot struct {
	Version  int      `json:"version"`
	Previous defaults `json:"previous"`
}

func (windowsPlatform) current() (defaults, error) { return currentDefaults() }

func (windowsPlatform) previous() (defaults, bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, stateKey, registry.QUERY_VALUE)
	if errors.Is(err, os.ErrNotExist) {
		return defaults{}, false, nil
	}
	if err != nil {
		return defaults{}, false, err
	}
	defer func() { _ = key.Close() }()
	value, _, err := key.GetStringValue("Previous")
	if errors.Is(err, os.ErrNotExist) {
		return defaults{}, false, nil
	}
	if err != nil {
		return defaults{}, false, err
	}
	var saved snapshot
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&saved); err != nil {
		return defaults{}, false, fmt.Errorf("saved browser settings are invalid; choose another browser for HTTP and HTTPS in Windows Settings, then run --restore to remove the stale registration: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return defaults{}, false, errors.New("saved browser settings contain unexpected data; choose another browser for HTTP and HTTPS in Windows Settings, then run --restore")
	}
	if saved.Version != 1 || saved.Previous.HTTP == "" || saved.Previous.HTTPS == "" {
		return defaults{}, false, errors.New("saved browser settings are incomplete or unsupported; choose another browser for HTTP and HTTPS in Windows Settings, then run --restore")
	}
	return saved.Previous, true, nil
}

func (windowsPlatform) savePrevious(previous defaults) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, stateKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = key.Close() }()
	if _, _, err := key.GetStringValue("Previous"); !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return err
		}
		return errors.New("previous browser settings already exist and were not replaced")
	}
	value, err := json.Marshal(snapshot{Version: 1, Previous: previous})
	if err != nil {
		return err
	}
	return key.SetStringValue("Previous", string(value))
}

func writeRegistrationValues(path string, values map[string]string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = key.Close() }()
	for name, value := range values {
		if err := key.SetStringValue(name, value); err != nil {
			return err
		}
	}
	return nil
}

func (windowsPlatform) install() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if strings.Contains(executable, "%") {
		return errors.New("move dynamicbrowser to a path without '%' before registering it")
	}
	published, err := publication()
	if err != nil {
		return err
	}
	if published != "" && !strings.EqualFold(published, capabilitiesKey) {
		return errors.New("the dynamicbrowser registration name is owned by another application; its registration was left unchanged")
	}
	command := windows.EscapeArg(executable) + ` -- "%1"`
	icon := `"` + executable + `",0`
	defer notifyAssociations()
	for _, id := range []string{ownDefaults.HTTP, ownDefaults.HTTPS} {
		path := `Software\Classes\` + id
		for _, entry := range []struct {
			path   string
			values map[string]string
		}{
			{path, map[string]string{"": "dynamicbrowser web link", "URL Protocol": ""}},
			{path + `\DefaultIcon`, map[string]string{"": icon}},
			{path + `\shell\open\command`, map[string]string{"": command}},
		} {
			if err := writeRegistrationValues(entry.path, entry.values); err != nil {
				return err
			}
		}
	}
	for _, entry := range []struct {
		path   string
		values map[string]string
	}{
		{clientKey, map[string]string{"": applicationName}},
		{clientKey + `\DefaultIcon`, map[string]string{"": icon}},
		{clientKey + `\shell\open\command`, map[string]string{"": windows.EscapeArg(executable)}},
		{capabilitiesKey, map[string]string{"ApplicationName": applicationName, "ApplicationDescription": "Open web links in the configured browser profile.", "ApplicationIcon": icon}},
		{capabilitiesKey + `\URLAssociations`, map[string]string{"http": ownDefaults.HTTP, "https": ownDefaults.HTTPS}},
		{registeredApps, map[string]string{applicationName: capabilitiesKey}},
	} {
		if err := writeRegistrationValues(entry.path, entry.values); err != nil {
			return err
		}
	}
	return nil
}

func publication() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, registeredApps, registry.QUERY_VALUE)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer func() { _ = key.Close() }()
	value, _, err := key.GetStringValue(applicationName)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return value, err
}
