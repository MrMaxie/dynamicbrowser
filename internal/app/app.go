package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/MrMaxie/dynamicbrowser/internal/autostart"
	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/MrMaxie/dynamicbrowser/internal/routing"
	"github.com/gogpu/systray"
)

type Options struct {
	Force bool
	URLs  []string
}

func Run(options Options) error {
	request := routing.Request{URLs: options.URLs, Source: captureSource()}
	if len(options.URLs) > 0 && !options.Force {
		return routeOnDemand(request)
	}
	release, acquired, err := acquireForStartup(options.Force)
	if err != nil || !acquired {
		return err
	}
	defer release()
	path, err := config.Path()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	store := &config.Store{}
	changes := store.Changes()
	done, err := config.Watch(ctx, path, store)
	if err != nil {
		cancel()
		return err
	}
	defer func() { cancel(); <-done }()
	configuration, err := store.Snapshot()
	if err != nil {
		return err
	}
	router := routing.New(configuration, filepath.Dir(path), startCommand)
	updateAutostart := autostartUpdater(configureAutostart)
	select {
	case <-changes:
	default:
	}

	var stopControl, stopRouting func()
	defer func() {
		if stopRouting != nil {
			stopRouting()
		}
		if stopControl != nil {
			stopControl()
		}
	}()
	ready := func() error {
		if stopControl != nil {
			return nil
		}
		var err error
		stopControl, err = startInstanceControl(cancel)
		if err != nil {
			return err
		}
		stopRouting, err = startRoutingControl(router.Open)
		if err != nil {
			return err
		}
		return router.Open(request)
	}
	for {
		if err := runTray(ctx, cancel, store, router, path, updateAutostart, ready); err != nil {
			return err
		}
		if runtime.GOOS != "linux" || ctx.Err() != nil {
			return nil
		}
	}
}

func runTray(parent context.Context, closeTray context.CancelFunc, store *config.Store, router *routing.Router, path string, updateAutostart func(bool) error, ready func() error) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	configuration, loadErr := store.Snapshot()
	if loadErr == nil {
		router.Update(configuration)
		loadErr = updateAutostart(configuration.Autostart)
	}
	// The native window and its message loop must use the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	icon, err := trayIcon()
	if err != nil {
		return err
	}
	tray := systray.New().SetIcon(icon)
	updatesDone := make(chan struct{})
	remove := trayQuit(cancel, updatesDone, tray.Remove)
	edit := func() error { return editConfiguration(path) }
	tray.SetMenu(newTrayMenu(ctx, router, edit, closeTray).menu)
	setTrayTooltip(tray, loadErr)
	if err := bindTrayMenu(tray); err != nil {
		tray.Remove()
		return err
	}
	dispatch, restore, err := newTrayDispatch(tray, ctx)
	if err != nil {
		tray.Remove()
		return err
	}
	defer restore()
	tray.Show()
	go func() {
		defer close(updatesDone)
		for {
			select {
			case <-ctx.Done():
				return
			case <-store.Changes():
				next, loadErr := store.Snapshot()
				valid := loadErr == nil
				if valid {
					router.Update(next)
					loadErr = updateAutostart(next.Autostart)
					if runtime.GOOS == "linux" {
						// systray reuses positional DBus IDs on SetMenu. A new service prevents
						// clicks from an old displayed menu activating a different new action.
						remove()
						return
					}
				}
				if err := dispatchTrayUpdate(ctx, dispatch, func() {
					setTrayTooltip(tray, loadErr)
					if valid {
						tray.SetMenu(newTrayMenu(ctx, router, edit, closeTray).menu)
					}
				}); err != nil {
					return
				}
			}
		}
	}()
	go func() { <-ctx.Done(); remove() }()
	defer func() { cancel(); <-updatesDone }()
	if err := ready(); err != nil {
		remove()
		_ = tray.Run()
		return err
	}
	return tray.Run()
}

func setTrayTooltip(tray *systray.SystemTray, err error) {
	if err != nil {
		tray.SetTooltip("dynamicbrowser — configuration or autostart error; edit config.yaml")
		return
	}
	tray.SetTooltip("dynamicbrowser — running")
}

func autostartUpdater(apply func(bool) error) func(bool) error {
	var applied *bool
	return func(enabled bool) error {
		if applied != nil && *applied == enabled {
			return nil
		}
		if err := apply(enabled); err != nil {
			return err
		}
		applied = &enabled
		return nil
	}
}

func configureAutostart(enabled bool) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := autostart.Set(instanceName, executable, enabled); err != nil {
		return fmt.Errorf("configure tray autostart: %w", err)
	}
	return nil
}

func routeOnDemand(request routing.Request) error {
	deadline := time.Now().Add(restartTimeout)
	for {
		if err := forwardURLs(request); err == nil {
			return nil
		} else if !errors.Is(err, errRestartNotReady) {
			return err
		}
		release, acquired, err := acquireForStartup(false)
		if err != nil {
			return err
		}
		if acquired {
			defer release()
			path, err := config.Path()
			if err != nil {
				return err
			}
			configuration, err := config.Load(path)
			if err != nil {
				return err
			}
			return openWithAutostart(configuration.Autostart, configureAutostart, func() error {
				return routing.New(configuration, filepath.Dir(path), startCommand).Open(request)
			})
		}
		if time.Now().After(deadline) {
			return errors.New("running tray is not accepting URLs; close or restart it and try again")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func openWithAutostart(enabled bool, register func(bool) error, open func() error) error {
	registrationErr := register(enabled)
	return errors.Join(open(), registrationErr)
}

func trayQuit(cancel context.CancelFunc, updatesDone <-chan struct{}, remove func()) func() {
	var once sync.Once
	return func() { once.Do(func() { cancel(); go func() { <-updatesDone; remove() }() }) }
}
