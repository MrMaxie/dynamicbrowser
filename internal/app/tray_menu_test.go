package app

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func TestConcurrentTraySelections(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Browsers: config.Browsers{{Name: "first", Exe: executable}, {Name: "second", Exe: executable}}}
	router := routing.New(cfg, t.TempDir(), func(string, []string) error { return nil })
	view := newTrayMenu(context.Background(), router, func() error { return nil }, func() {})
	choices := []string{"", "first", "second"}
	for range 100 {
		var group sync.WaitGroup
		for i := range 30 {
			group.Go(func() { view.choose(choices[i%len(choices)]) })
		}
		group.Wait()
		selected := router.Forced()
		for name, item := range view.checks {
			if item.IsChecked() != (name == selected) {
				t.Fatalf("check for %q inconsistent with selection %q", name, selected)
			}
		}
	}
}
