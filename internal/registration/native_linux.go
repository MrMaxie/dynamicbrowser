package registration

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const linuxDesktopID = "dynamicbrowser.desktop"
const desktopOwner = "\nX-Dynamicbrowser-Managed=true\n"

var linuxSchemes = [...]string{"http", "https"}

type linuxSetup struct {
	entry, snapshot, executable string
	applications                []string
	mime                        func(...string) (string, error)
	confirm                     func() error
}

func Register() error { return linuxDefaultBrowser(false) }
func Restore() error  { return linuxDefaultBrowser(true) }

func linuxDefaultBrowser(restore bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	base := func(key, fallback string) string {
		if path := os.Getenv(key); filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(home, fallback)
	}
	applications := []string{filepath.Join(base("XDG_DATA_HOME", ".local/share"), "applications")}
	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}
	for _, dir := range filepath.SplitList(dataDirs) {
		if filepath.IsAbs(dir) {
			applications = append(applications, filepath.Join(dir, "applications"))
		}
	}
	state := filepath.Join(base("XDG_STATE_HOME", ".local/state"), "dynamicbrowser")
	if err := os.MkdirAll(state, 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(state, "registration.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return fmt.Errorf("lock default-browser setup: %w", err)
	}
	tool, err := exec.LookPath("xdg-mime")
	if err != nil {
		return errors.New("install xdg-utils to change the default browser on Linux")
	}
	s := linuxSetup{
		entry: filepath.Join(applications[0], linuxDesktopID), snapshot: filepath.Join(state, "previous-handlers.json"),
		executable: executable, applications: applications,
		mime: func(args ...string) (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, tool, args...).Output()
			if err != nil {
				return "", fmt.Errorf("xdg-mime %s: %w", args[0], err)
			}
			return strings.TrimSpace(string(out)), nil
		},
		confirm: func() error {
			_, _ = fmt.Fprint(os.Stderr, "Make dynamicbrowser the default for HTTP and HTTPS? [y/N] ")
			answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
				return ErrCancelled
			}
			return nil
		},
	}
	if restore {
		return s.restore()
	}
	return s.register()
}

func (s linuxSetup) current(scheme string) (string, error) {
	return s.mime("query", "default", "x-scheme-handler/"+scheme)
}

func (s linuxSetup) set(scheme, id string) error {
	if _, err := s.mime("default", id, "x-scheme-handler/"+scheme); err != nil {
		return err
	}
	got, err := s.current(scheme)
	if err != nil {
		return err
	}
	if got != id {
		return fmt.Errorf("could not verify the %s handler; check your desktop's default applications", scheme)
	}
	return nil
}

func (s linuxSetup) previous() (map[string]string, bool, error) {
	data, err := os.ReadFile(s.snapshot)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var saved map[string]string
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, false, fmt.Errorf("read previous browser choices: %w", err)
	}
	return saved, true, nil
}

func (s linuxSetup) available(id string) bool {
	if id == linuxDesktopID || !strings.HasSuffix(id, ".desktop") || strings.HasPrefix(id, "-") || strings.ContainsAny(id, "/\\\t\r\n ") {
		return false
	}
	for _, dir := range s.applications {
		if info, err := os.Stat(filepath.Join(dir, id)); err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	// shortcut: only top-level desktop files are supported; resolve nested desktop IDs if a desktop requires them.
	return false
}

func (s linuxSetup) ownedEntry() (bool, error) {
	info, err := os.Lstat(s.entry)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, errors.New("dynamicbrowser.desktop already exists and is not managed by this application")
	}
	data, err := os.ReadFile(s.entry)
	if err != nil {
		return false, err
	}
	if !strings.Contains(string(data), desktopOwner) {
		return false, errors.New("dynamicbrowser.desktop was installed manually; use your desktop settings to change its associations and remove that entry before running setup")
	}
	return true, nil
}

func linuxDesktopEntry(executable string) ([]byte, error) {
	if !filepath.IsAbs(executable) || strings.ContainsAny(executable, "=\r\n") {
		return nil, errors.New("move the executable to a path without equals signs or line breaks before registering it")
	}
	escaped := strings.NewReplacer(`\`, `\\\\`, `"`, `\\"`, `$`, `\\$`, "`", "\\\\`", "%", "%%", "\t", `\t`).Replace(executable)
	return []byte("[Desktop Entry]\nType=Application\nName=dynamicbrowser\nExec=\"" + escaped + "\" %U\nTerminal=false\nCategories=Network;WebBrowser;\nMimeType=x-scheme-handler/http;x-scheme-handler/https;" + desktopOwner), nil
}

func (s linuxSetup) register() error {
	if _, err := s.ownedEntry(); err != nil {
		return err
	}
	entry, err := linuxDesktopEntry(s.executable)
	if err != nil {
		return err
	}
	saved, exists, err := s.previous()
	if err != nil {
		return err
	}
	current := make(map[string]string)
	for _, scheme := range linuxSchemes {
		id, err := s.current(scheme)
		if err != nil {
			return err
		}
		current[scheme] = id
	}
	if !exists {
		saved = current
	}
	for _, scheme := range linuxSchemes {
		if !s.available(saved[scheme]) {
			return fmt.Errorf("no restorable previous %s handler; select another default browser in your desktop settings, then try again", scheme)
		}
	}
	if current["http"] != linuxDesktopID || current["https"] != linuxDesktopID {
		if err := s.confirm(); err != nil {
			return err
		}
		for _, scheme := range linuxSchemes {
			if id, err := s.current(scheme); err != nil || id != current[scheme] {
				return errors.New("default browser choices changed during setup; run --register again")
			}
		}
	}
	if !exists {
		data, err := json.Marshal(saved)
		if err != nil {
			return err
		}
		if err := writeLinuxSetupFile(s.snapshot, data, 0600); err != nil {
			return err
		}
	}
	if err := writeLinuxSetupFile(s.entry, entry, 0644); err != nil {
		return err
	}
	for _, scheme := range linuxSchemes {
		if current[scheme] != linuxDesktopID {
			if err := s.set(scheme, linuxDesktopID); err != nil {
				return fmt.Errorf("registration incomplete; retry --register or run --restore: %w", err)
			}
		}
	}
	for _, scheme := range linuxSchemes {
		if id, err := s.current(scheme); err != nil || id != linuxDesktopID {
			return fmt.Errorf("registration incomplete for %s; retry --register or run --restore", scheme)
		}
	}
	return nil
}

func (s linuxSetup) restore() error {
	owned, err := s.ownedEntry()
	if err != nil {
		return err
	}
	saved, _, err := s.previous()
	if err != nil {
		return err
	}
	for _, scheme := range linuxSchemes {
		id, err := s.current(scheme)
		if err != nil {
			return err
		}
		if id != linuxDesktopID {
			continue
		}
		if !s.available(saved[scheme]) {
			return fmt.Errorf("previous %s handler is missing; choose a browser in desktop settings, then run --restore again", scheme)
		}
		if err := s.set(scheme, saved[scheme]); err != nil {
			return fmt.Errorf("restoration incomplete; run --restore again: %w", err)
		}
	}
	for _, scheme := range linuxSchemes {
		if id, err := s.current(scheme); err != nil || id == linuxDesktopID {
			return fmt.Errorf("could not verify restoration for %s; run --restore again", scheme)
		}
	}
	if owned {
		if err := os.Remove(s.entry); err != nil {
			return err
		}
	}
	if err := os.Remove(s.snapshot); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func writeLinuxSetupFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".dynamicbrowser-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Chmod(mode), file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
