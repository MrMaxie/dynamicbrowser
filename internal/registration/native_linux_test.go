package registration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxSetup(t *testing.T) {
	for _, scenario := range []string{"restore", "preserve independent choice", "decline", "missing previous", "partial", "manual entry"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			const http, https = "x-scheme-handler/http", "x-scheme-handler/https"
			current := map[string]string{http: "first.desktop", https: "second.desktop"}
			for _, id := range current {
				if err := os.WriteFile(filepath.Join(dir, id), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			setCalls := 0
			failure := errors.New("cannot set HTTPS")
			s := linuxSetup{
				entry: filepath.Join(dir, linuxDesktopID), snapshot: filepath.Join(dir, "previous.json"),
				executable: filepath.Join(dir, "test binary"), applications: []string{dir},
				confirm: func() error { return nil },
				mime: func(args ...string) (string, error) {
					scheme := args[2]
					if args[0] == "query" {
						return current[scheme], nil
					}
					setCalls++
					if scenario == "partial" && args[1] == linuxDesktopID && scheme == https {
						return "", failure
					}
					current[scheme] = args[1]
					return "", nil
				},
			}
			switch scenario {
			case "decline":
				s.confirm = func() error { return ErrCancelled }
			case "missing previous":
				current[https] = "not-installed.desktop"
			case "manual entry":
				if err := os.WriteFile(s.entry, []byte("manual"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := s.register()
			switch scenario {
			case "decline", "missing previous", "manual entry":
				if err == nil || setCalls != 0 || (scenario == "decline" && !errors.Is(err, ErrCancelled)) {
					t.Fatalf("rejected setup: %v, setter calls %d", err, setCalls)
				}
				if _, err := os.Stat(s.snapshot); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("aborted setup created snapshot: %v", err)
				}
				return
			case "partial":
				if !errors.Is(err, failure) || current[http] != linuxDesktopID || current[https] != "second.desktop" {
					t.Fatalf("partial registration: %v, %#v", err, current)
				}
			default:
				if err != nil || current[http] != linuxDesktopID || current[https] != linuxDesktopID {
					t.Fatalf("register: %v, %#v", err, current)
				}
				if err := s.register(); err != nil || setCalls != 2 {
					t.Fatalf("repeat registration: %v, setter calls %d", err, setCalls)
				}
			}
			wantHTTPS := "second.desktop"
			if scenario == "preserve independent choice" {
				wantHTTPS = "newer.desktop"
				current[https] = wantHTTPS
			}
			if err := s.restore(); err != nil || current[http] != "first.desktop" || current[https] != wantHTTPS {
				t.Fatalf("restore: %v, %#v", err, current)
			}
			for _, path := range []string{s.entry, s.snapshot} {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("restore did not clean up %s: %v", path, err)
				}
			}
		})
	}
}
