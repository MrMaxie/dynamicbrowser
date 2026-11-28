package app

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func TestTrayAppliesLatestGoodConfigWithLoadError(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	old := &config.Config{Default: "old"}
	next := &config.Config{Default: "new", Browsers: config.Browsers{{Name: "new", Exe: exe}}}
	router := routing.New(old, t.TempDir(), nil)
	apply := trayConfigUpdater(router, func(bool) error { return nil })
	if changed, err := apply(old, nil); !changed || err != nil {
		t.Fatalf("initial apply: changed=%t, err=%v", changed, err)
	}
	invalid := errors.New("latest edit is invalid")
	if changed, err := apply(next, invalid); !changed || !errors.Is(err, invalid) {
		t.Fatalf("changed=%t, err=%v", changed, err)
	}
	if browser, err := router.Select("https://example.test", routing.Source{}); err != nil || browser.Name != "new" {
		t.Fatalf("browser=%q, err=%v", browser.Name, err)
	}
	if changed, _ := apply(next, invalid); changed {
		t.Fatal("applied the same config twice")
	}
}

func TestSlowRoutingDoesNotBlockOtherClients(t *testing.T) {
	const slow = "https://example.test/slow"
	const fast = "https://example.test/fast"
	started := make(chan struct{}, routingWorkers)
	release := make(chan struct{})
	stop, err := startRoutingControl(func(req routing.Request) error {
		if req.URLs[0] == slow {
			started <- struct{}{}
			<-release
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	defer close(release)
	forward := func(url string) error { return forwardURLs(routing.Request{URLs: []string{url}}) }
	startSlow := func() {
		t.Helper()
		done := make(chan error, 1)
		go func() { done <- forward(slow) }()
		select {
		case <-started:
		case err := <-done:
			t.Fatalf("slow route returned before starting: %v", err)
		case <-time.After(time.Second):
			t.Fatal("slow route did not start")
		}
	}
	startSlow()
	if err := forward(fast); err != nil {
		t.Fatalf("fast route blocked by slow route: %v", err)
	}
	for range routingWorkers - 1 {
		startSlow()
	}
	if err := forward(fast); err == nil || !strings.Contains(err.Error(), "tray is busy") {
		t.Fatalf("overloaded route did not report busy: %v", err)
	}
}
