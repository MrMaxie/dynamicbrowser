//go:build linux || darwin

package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gogpu/systray"
	"golang.org/x/sys/unix"
)

func acquireInstance() (func(), bool, error) {
	return acquireNamedInstance(instanceName)
}

func instanceDirectory() (string, error) {
	if runtime.GOOS == "linux" {
		if base := os.Getenv("XDG_RUNTIME_DIR"); filepath.IsAbs(base) {
			dir := filepath.Join(base, "dynamicbrowser")
			if len(dir)+len("/0123456789abcdef.sock.routing") <= 107 && privateRuntimeDir(base) == nil {
				return dir, ensureRuntimeDir(dir)
			}
		}
		dir := fmt.Sprintf("/tmp/dynamicbrowser-%d", os.Geteuid())
		return dir, ensureRuntimeDir(dir)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "dynamicbrowser")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func ensureRuntimeDir(dir string) error {
	if err := os.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return privateRuntimeDir(dir)
}

func privateRuntimeDir(dir string) error {
	var stat unix.Stat_t
	if err := unix.Lstat(dir, &stat); err != nil {
		return err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0077 != 0 {
		return fmt.Errorf("instance directory %q must be private and owned by the current user", dir)
	}
	return nil
}

func acquireNamedInstance(identity string) (func(), bool, error) {
	dir, err := instanceDirectory()
	if err != nil {
		return nil, false, err
	}
	file, err := os.OpenFile(filepath.Join(dir, identity+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, false, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return nil, false, nil
		}
		return nil, false, err
	}
	// Keep the inode: unlinking a lock file lets concurrent processes lock different files.
	return func() { _ = file.Close() }, true, nil
}

func ShowStartupError(err error) {
	message := err.Error()
	var commands []*exec.Cmd
	if runtime.GOOS == "darwin" {
		commands = []*exec.Cmd{exec.Command("osascript", "-e", `on run argv
 display dialog (item 1 of argv) with title "dynamicbrowser" buttons {"OK"} default button "OK" with icon stop
end run`, "--", message)}
	} else {
		commands = []*exec.Cmd{
			exec.Command("zenity", "--error", "--title=dynamicbrowser", "--no-markup", "--text="+message),
			exec.Command("kdialog", "--title", "dynamicbrowser", "--error", message),
		}
	}
	for _, command := range commands {
		if command.Run() == nil {
			return
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "dynamicbrowser: %s\n", message)
}

func trayIcon() ([]byte, error) {
	if runtime.GOOS == "darwin" {
		return trayIconForSize(44)
	}
	return trayIconForSize(24)
}

func bindTrayMenu(tray *systray.SystemTray) error {
	if runtime.GOOS == "darwin" {
		icon, err := trayIcon()
		if err != nil {
			return err
		}
		tray.SetTemplateIcon(icon)
	}
	// NSStatusItem/SNI hosts open the attached menu; there is no Win32 event bridge here.
	return nil
}
