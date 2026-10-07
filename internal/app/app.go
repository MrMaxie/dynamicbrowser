package app

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"runtime"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/gogpu/systray"
)

// Run starts the singleton tray application and blocks until it exits.
func Run() error {
	release, acquired, err := acquireInstance()
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
	done, err := config.Watch(ctx, path, store)
	if err != nil {
		cancel()
		return err
	}
	defer func() { cancel(); <-done }()

	icon, err := trayIcon()
	if err != nil {
		return err
	}
	// The native window and its message loop must use the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	tray := systray.New().SetIcon(icon).SetTooltip("dynamicbrowser — running")
	menu := systray.NewMenu()
	status := func() { tray.ShowNotification("dynamicbrowser", store.Status()) }
	menu.Add("Configuration status", status)
	menu.AddSeparator()
	menu.Add("Quit", tray.Remove)
	tray.SetMenu(menu).OnClick(status).Show()
	return tray.Run()
}

func trayIcon() ([]byte, error) {
	icon := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 4; x < 28; x++ {
			icon.SetNRGBA(x, y, color.NRGBA{R: 40, G: 180, B: 100, A: 255})
		}
	}
	var data bytes.Buffer
	err := png.Encode(&data, icon)
	return data.Bytes(), err
}
