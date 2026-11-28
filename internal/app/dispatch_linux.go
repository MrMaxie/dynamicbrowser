package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/gogpu/systray"
)

func checkTrayHost() error {
	bus, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("tray is unavailable on Linux: %w", err)
	}
	defer func() { _ = bus.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	const watcher = "org.kde.StatusNotifierWatcher"
	var status dbus.Variant
	err = bus.Object(watcher, "/StatusNotifierWatcher").CallWithContext(ctx,
		"org.freedesktop.DBus.Properties.Get", dbus.FlagNoAutoStart,
		watcher, "IsStatusNotifierHostRegistered").Store(&status)
	if err != nil || status.Value() != true {
		return errors.New("tray is unavailable on Linux: no StatusNotifierItem host is ready")
	}
	return nil
}

func newTrayDispatch(_ *systray.SystemTray, ctx context.Context) (func(func()) error, func(), error) {
	return func(job func()) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		job()
		return nil
	}, func() {}, nil
}
