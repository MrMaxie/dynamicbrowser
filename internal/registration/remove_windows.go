package registration

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func deleteRegistrationTree(path string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.ENUMERATE_SUB_KEYS)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	names, err := key.ReadSubKeyNames(-1)
	_ = key.Close()
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := deleteRegistrationTree(path + `\` + name); err != nil {
			return err
		}
	}
	err = registry.DeleteKey(registry.CURRENT_USER, path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (windowsPlatform) remove() error {
	current, err := currentDefaults()
	if err != nil {
		return err
	}
	if current.any(ownDefaults) {
		return errors.New("dynamicbrowser is still selected for HTTP or HTTPS; choose another browser in Windows Settings before removing its registration")
	}
	published, err := publication()
	if err != nil {
		return err
	}
	if published != "" && !strings.EqualFold(published, capabilitiesKey) {
		return errors.New("another application owns the dynamicbrowser registration name; it was not removed")
	}
	defer notifyAssociations()
	if published != "" {
		key, err := registry.OpenKey(registry.CURRENT_USER, registeredApps, registry.SET_VALUE)
		if err != nil {
			return err
		}
		err = key.DeleteValue(applicationName)
		_ = key.Close()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		notifyAssociations()
	}
	current, err = currentDefaults()
	if err != nil || current.any(ownDefaults) {
		if published != "" {
			err = errors.Join(err, writeRegistrationValues(registeredApps, map[string]string{applicationName: capabilitiesKey}))
		}
		if err != nil {
			return err
		}
		return errors.New("browser choices changed during removal; registration was retained, review Windows Settings and retry")
	}
	for _, path := range []string{`Software\Classes\` + ownDefaults.HTTP, `Software\Classes\` + ownDefaults.HTTPS, clientKey} {
		if err := deleteRegistrationTree(path); err != nil {
			return err
		}
	}
	return deleteRegistrationTree(stateKey)
}
