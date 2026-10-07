package app

import (
	"context"
	"runtime"
	"sync"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/gogpu/systray"
)

type Options struct {
	Force bool
}

// Run starts the singleton tray application and blocks until it exits.
func Run(options Options) error {
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

	// The native window and its message loop must use the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	icon, err := trayIcon()
	if err != nil {
		return err
	}
	tray := systray.New().SetIcon(icon).SetTooltip("dynamicbrowser — running")
	menu := systray.NewMenu()
	status := menu.Add(store.Status(), nil)
	status.SetDisabled(true)
	statusDone := make(chan struct{})
	quit := trayQuit(cancel, statusDone, tray.Remove)
	menu.AddSeparator()
	menu.Add("Quit", quit)
	tray.SetMenu(menu)
	if err := bindTrayMenu(tray); err != nil {
		tray.Remove()
		return err
	}
	go func() {
		defer close(statusDone)
		for {
			select {
			case <-ctx.Done():
				return
			case <-changes:
				status.SetLabel(store.Status())
			}
		}
	}()
	defer func() { cancel(); <-statusDone }()
	tray.Show()
	stopControl, err := startInstanceControl(quit)
	if err != nil {
		// Let pending native updates finish before destroying the UI event loop.
		quit()
		_ = tray.Run()
		return err
	}
	defer stopControl()
	return tray.Run()
}

func trayQuit(cancel context.CancelFunc, updatesDone <-chan struct{}, remove func()) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			// Keep the UI loop pumping while an in-flight native menu update completes.
			go func() {
				<-updatesDone
				remove()
			}()
		})
	}
}
