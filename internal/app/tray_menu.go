package app

import (
	"context"
	"sync"

	"github.com/MrMaxie/dynamicbrowser/internal/routing"
	"github.com/gogpu/systray"
)

type trayMenu struct {
	mu     sync.Mutex
	ctx    context.Context
	menu   *systray.Menu
	checks map[string]*systray.MenuItem
	router *routing.Router
}

func newTrayMenu(ctx context.Context, router *routing.Router, edit func() error, closeTray func()) *trayMenu {
	view := &trayMenu{ctx: ctx, menu: systray.NewMenu(), checks: make(map[string]*systray.MenuItem), router: router}
	view.checks[""] = view.menu.AddCheckbox("Auto", router.Forced() == "", func() { view.choose("") })
	view.menu.AddSeparator()
	for _, name := range router.Names() {
		view.checks[name] = view.menu.AddCheckbox(name, router.Forced() == name, func() { view.choose(name) })
	}
	if len(view.checks) > 1 {
		view.menu.AddSeparator()
	}
	view.menu.Add("Edit configuration", func() {
		if ctx.Err() != nil {
			return
		}
		if err := edit(); err != nil {
			ShowStartupError(err)
		}
	})
	view.menu.Add("Close tray", func() {
		if ctx.Err() == nil {
			closeTray()
		}
	})
	return view
}

func (view *trayMenu) choose(name string) {
	view.mu.Lock()
	defer view.mu.Unlock()
	if view.ctx.Err() != nil {
		return
	}
	if err := view.router.Force(name); err != nil {
		ShowStartupError(err)
		return
	}
	for candidate, check := range view.checks {
		check.SetChecked(candidate == name)
	}
}
