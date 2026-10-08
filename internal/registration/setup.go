package registration

import (
	"errors"
	"fmt"
	"strings"
)

var ErrCancelled = errors.New("default-browser setup cancelled")

type defaults struct {
	HTTP  string `json:"http"`
	HTTPS string `json:"https"`
}

func (d defaults) any(other defaults) bool {
	return strings.EqualFold(d.HTTP, other.HTTP) || strings.EqualFold(d.HTTP, other.HTTPS) || strings.EqualFold(d.HTTPS, other.HTTP) || strings.EqualFold(d.HTTPS, other.HTTPS)
}

func (d defaults) matches(other defaults) bool {
	return strings.EqualFold(d.HTTP, other.HTTP) && strings.EqualFold(d.HTTPS, other.HTTPS)
}

type platform interface {
	current() (defaults, error)
	previous() (defaults, bool, error)
	savePrevious(defaults) error
	install() error
	remove() error
	describe(string) (string, error)
	assist(string, bool) (bool, error)
}

type setup struct {
	system platform
	own    defaults
}

func (s setup) owns(id string) bool {
	return strings.EqualFold(id, s.own.HTTP) || strings.EqualFold(id, s.own.HTTPS)
}

func (s setup) register() error {
	current, err := s.system.current()
	if err != nil {
		return err
	}
	previous, saved, err := s.system.previous()
	if err != nil {
		return err
	}
	if saved && previous.any(s.own) {
		return errors.New("saved previous browser settings refer to dynamicbrowser; choose another browser in Windows Settings before setting up dynamicbrowser again")
	}
	if !saved {
		if current.any(s.own) {
			return errors.New("previous browser settings are missing; choose another browser for HTTP and HTTPS in Windows Settings before registering again")
		}
		if err := s.system.savePrevious(current); err != nil {
			return fmt.Errorf("save previous browser settings: %w", err)
		}
	}
	if err := s.system.install(); err != nil {
		return fmt.Errorf("register browser: %w", err)
	}
	current, err = s.system.current()
	if err != nil {
		return err
	}
	if current.matches(s.own) {
		return nil
	}
	confirmed, err := s.system.assist("In Windows Settings, choose dynamicbrowser for HTTP and HTTPS. Then click OK here to verify the selection. Cancel leaves dynamicbrowser available without changing your browser choice.", true)
	if err != nil {
		return err
	}
	if !confirmed {
		return ErrCancelled
	}
	current, err = s.system.current()
	if err != nil {
		return err
	}
	if !current.matches(s.own) {
		return errors.New("dynamicbrowser is not the default for both HTTP and HTTPS; select it for both in Windows Settings, then run --register again")
	}
	return nil
}

func (s setup) restore() error {
	for {
		current, err := s.system.current()
		if err != nil {
			return err
		}
		if !current.any(s.own) {
			return s.system.remove()
		}
		previous, saved, err := s.system.previous()
		if err != nil {
			return err
		}
		if !saved || previous.any(s.own) {
			return errors.New("previous browser settings are missing or invalid; choose another browser for HTTP and HTTPS in Windows Settings, then run --restore again to remove dynamicbrowser")
		}
		var instructions []string
		for _, target := range []struct{ scheme, current, previous string }{
			{"HTTP", current.HTTP, previous.HTTP},
			{"HTTPS", current.HTTPS, previous.HTTPS},
		} {
			if !s.owns(target.current) {
				continue
			}
			name, err := s.system.describe(target.previous)
			if err != nil {
				return fmt.Errorf("previous %s browser is unavailable; choose a browser for %s in Windows Settings, then run --restore again: %w", target.scheme, target.scheme, err)
			}
			instructions = append(instructions, target.scheme+": "+name)
		}
		latest, err := s.system.current()
		if err != nil {
			return err
		}
		if !latest.matches(current) {
			continue
		}
		message := "In Windows Settings, restore these link types:\n\n" + strings.Join(instructions, "\n") + "\n\nLeave other link types unchanged. Then click OK here to verify and remove dynamicbrowser's registration."
		confirmed, err := s.system.assist(message, false)
		if err != nil {
			return err
		}
		if !confirmed {
			return ErrCancelled
		}
		after, err := s.system.current()
		if err != nil {
			return err
		}
		if (s.owns(current.HTTP) && !strings.EqualFold(after.HTTP, previous.HTTP)) || (s.owns(current.HTTPS) && !strings.EqualFold(after.HTTPS, previous.HTTPS)) {
			return errors.New("the saved browser choices have not been restored; review HTTP and HTTPS in Windows Settings and run --restore again")
		}
		if (!s.owns(current.HTTP) && !strings.EqualFold(after.HTTP, current.HTTP)) || (!s.owns(current.HTTPS) && !strings.EqualFold(after.HTTPS, current.HTTPS)) {
			return errors.New("a link type outside this restoration changed in Windows Settings; choose your preferred browser for that link type, then run --restore again")
		}
		if after.any(s.own) {
			return errors.New("dynamicbrowser is still selected for a link type; select another browser in Windows Settings before removing its registration")
		}
		return s.system.remove()
	}
}
