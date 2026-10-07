//go:build linux || darwin

package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gogpu/systray"
	"golang.org/x/sys/unix"
)

func acquireInstance() (func(), bool, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, false, err
	}
	dir := filepath.Join(cache, "dynamicbrowser")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, false, err
	}
	file, err := os.OpenFile(filepath.Join(dir, instanceName+".lock"), os.O_CREATE|os.O_RDWR, 0600)
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
	_, _ = fmt.Fprintf(os.Stderr, "dynamicbrowser: %v\n", err)
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
