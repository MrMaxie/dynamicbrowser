package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func TestLauncherProcess(t *testing.T) {
	for path, want := range map[string]string{
		"/usr/lib/xdg-desktop-portal-hyprland": "",
		"/usr/bin/bash":                        "",
		"/usr/bin/xdg-open":                    "",
		"/opt/slack/slack (deleted)":           "slack",
	} {
		if got := launcherProcess(path); got != want {
			t.Errorf("launcherProcess(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestLinuxProcessRouting(t *testing.T) {
	const expectedEnv = "DYNAMICBROWSER_TEST_LAUNCHER"
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if want := os.Getenv(expectedEnv); want != "" {
		source := captureSource()
		if source.Process != want || source.Window != "" {
			t.Fatalf("captured %+v, want only process %q", source, want)
		}
		base := t.TempDir()
		path := filepath.Join(base, "config.yaml")
		yaml := fmt.Sprintf("default: fallback\nbrowsers:\n  fallback: {exe: %q}\n  matched: {exe: %q}\nrules:\n  - browser: matched\n    source: {window: %q}\n  - browser: matched\n    source: {process: '*'}\n", executable, executable, want)
		if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		launched := make(chan []string, 1)
		cfg.Browsers[0].Args = []string{"fallback"}
		cfg.Browsers[1].Args = []string{"matched"}
		router := routing.New(cfg, base, func(_ string, arguments []string) error {
			launched <- arguments
			return nil
		})
		t.Setenv("XDG_CACHE_HOME", filepath.Join(base, "cache"))
		stop, err := startRoutingControl(router.Open)
		if err != nil {
			t.Fatal(err)
		}
		defer stop()
		for _, route := range []func(routing.Request) error{router.Open, forwardURLs} {
			for _, tc := range []struct {
				source routing.Source
				want   string
			}{{source, "matched"}, {routing.Source{RequireKnown: true}, "fallback"}} {
				err := route(routing.Request{URLs: []string{"https://example.test/"}, Source: tc.source})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case args := <-launched:
					if len(args) != 2 || args[0] != tc.want {
						t.Fatalf("source %+v selected %v; want %q", tc.source, args, tc.want)
					}
				default:
					t.Fatal("no browser launched")
				}
			}
		}
		return
	}
	child := exec.Command(executable, "-test.run=^TestLinuxProcessRouting$")
	child.Env = append(os.Environ(), expectedEnv+"="+filepath.Base(executable))
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("child: %v\n%s", err, out)
	}
}
