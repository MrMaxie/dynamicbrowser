package routing

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"go.yaml.in/yaml/v3"
)

func configuration(t *testing.T, extra string) *config.Config {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	text := fmt.Sprintf("browsers:\n  work:\n    exe: %q\n    args: ['--profile=Example Profile']\n  personal:\n    exe: %q\n", executable, executable) + extra
	var cfg config.Config
	if err := yaml.Unmarshal([]byte(text), &cfg); err != nil {
		t.Fatal(err)
	}
	return &cfg
}

func TestMatchingAndFallback(t *testing.T) {
	cfg := configuration(t, "default: personal\nrules:\n  - browser: work\n    source:\n      process: ['Editor*', 'regex:^Viewer[0-9]+$']\n      window: '*Document*'\n    target:\n      url: 'https://*'\n      domain: example.com\n      path: ['*/work/*', '/work/*']\n")
	router := New(cfg, t.TempDir(), nil)
	for _, tt := range []struct {
		url     string
		source  Source
		browser string
	}{
		{"https://example.com/work/item", Source{Process: "EditorApp", Window: "Document 1"}, "work"},
		{"https://example.com/area/work/item", Source{Process: "Viewer2", Window: "Document 2"}, "work"},
		{"http://example.com/work/item", Source{Process: "EditorApp", Window: "Document 1"}, "personal"},
		{"https://other.example/work/item", Source{Process: "EditorApp", Window: "Document 1"}, "personal"},
		{"https://example.com/work/item", Source{Process: "EditorApp", Window: "Other"}, "personal"},
		{"https://example.com/work/item", Source{}, "personal"},
	} {
		browser, err := router.Select(tt.url, tt.source)
		if err != nil || browser.Name != tt.browser {
			t.Fatalf("selection = %q, %v; want %q", browser.Name, err, tt.browser)
		}
	}
}

func TestWindowHintIsLastRuleFallback(t *testing.T) {
	cfg := configuration(t, "default: personal\nrules:\n  - browser: personal\n    source: {process: '*Slack*'}\n  - browser: work\n    source: {window: '*Slack*'}\n  - browser: personal\n    target: {domain: url.example}\n")
	router := New(cfg, t.TempDir(), nil)
	for _, tt := range []struct {
		url, process, hint, browser, reason string
	}{
		{"https://example.test/", "", "com.slack.Slack", "work", "window-hint-rule-2"},
		{"https://url.example/", "", "com.slack.Slack", "personal", "rule-3"},
		{"https://example.test/", "Editor", "com.slack.Slack", "personal", "default"},
		{"https://example.test/", "", "", "personal", "default"},
	} {
		source := Source{Process: tt.process, RequireKnown: true}
		browser, reason, err := router.selectBrowser(tt.url, &source, func() string { return tt.hint })
		if err != nil || browser.Name != tt.browser || reason != tt.reason {
			t.Errorf("source %+v: got %q (%s), %v; want %q (%s)", source, browser.Name, reason, err, tt.browser, tt.reason)
		}
	}
}

func TestLinuxSourceSettings(t *testing.T) {
	const rules = "default: personal\nrules:\n  - browser: work\n    source: {window: '*Slack*'}\n"
	for _, tt := range []struct {
		name, settings string
		source         Source
		want           string
		lookup         bool
	}{
		{"alias on by default", "", Source{Process: "Slack", RequireKnown: true}, "work", false},
		{"alias off", "window-as-process: false\n", Source{Process: "Slack", RequireKnown: true}, "personal", false},
		{"hint on with alias off", "window-as-process: false\n", Source{RequireKnown: true}, "work", true},
		{"hint off", "active-window-fallback: false\n", Source{RequireKnown: true}, "personal", false},
		{"alias on with hint off", "active-window-fallback: false\n", Source{Process: "Slack", RequireKnown: true}, "work", false},
		{"native window unaffected", "window-as-process: false\nactive-window-fallback: false\n", Source{Process: "other", Window: "Slack"}, "work", false},
		{"no hint outside Linux", "active-window-fallback: true\nwindow-as-process: true\n", Source{Process: "Slack"}, "personal", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			router := New(configuration(t, tt.settings+rules), t.TempDir(), nil)
			called := false
			browser, _, err := router.selectBrowser("https://example.test/", &tt.source, func() string {
				called = true
				return "com.slack.Slack"
			})
			if err != nil || browser.Name != tt.want || called != tt.lookup {
				t.Fatalf("got %q, looked up=%t, err=%v; want %q, looked up=%t", browser.Name, called, err, tt.want, tt.lookup)
			}
		})
	}
}

func TestWindowsSourceGuessRequiresOptIn(t *testing.T) {
	const rules = "default: personal\nrules:\n  - browser: work\n    source: {process: '*', window: 'Document*'}\n"
	source := Source{Process: "Editor", Window: "Document 1", RequireKnown: true, ForegroundGuess: true}
	for _, tt := range []struct{ setting, want string }{
		{"", "personal"},
		{"guess-windows-source: true\n", "work"},
	} {
		router := New(configuration(t, tt.setting+rules), t.TempDir(), nil)
		browser, err := router.Select("https://example.test/", source)
		if err != nil || browser.Name != tt.want {
			t.Fatalf("setting %q: browser=%q, err=%v; want %q", tt.setting, browser.Name, err, tt.want)
		}
	}
}

func TestDefaultOnlyWhenNeeded(t *testing.T) {
	cfg := configuration(t, "rules:\n  - browser: work\n    target:\n      domain: example.com\n")
	router := New(cfg, t.TempDir(), nil)
	if browser, err := router.Select("https://example.com/", Source{}); err != nil || browser.Name != "work" {
		t.Fatal("a matched browser required a default")
	}
	if _, err := router.Select("https://other.example/", Source{}); err == nil {
		t.Fatal("missing default guessed the first browser")
	}
	cfg.Default = "missing"
	router.Update(cfg)
	if _, err := router.Select("https://other.example/", Source{}); err == nil {
		t.Fatal("unknown default selected another browser")
	}
	if err := router.Force("personal"); err != nil {
		t.Fatal(err)
	}
	if browser, err := router.Select("https://other.example/", Source{}); err != nil || browser.Name != "personal" {
		t.Fatal("valid override required a default")
	}
}

func TestUnknownRuleBrowserUsesDefault(t *testing.T) {
	cfg := configuration(t, "default: personal\nrules:\n  - browser: missing\n  - browser: work\n")
	router := New(cfg, t.TempDir(), nil)
	browser, err := router.Select("https://example.com/", Source{})
	if err != nil || browser.Name != "personal" {
		t.Fatal("unknown browser did not fall back immediately")
	}
}

func TestAvailableBrowserOrderAndOverrideReload(t *testing.T) {
	cfg := configuration(t, "default: personal\n")
	cfg.Browsers = append(cfg.Browsers, config.Browser{Name: "missing", Exe: "no-such-browser-executable"})
	router := New(cfg, t.TempDir(), nil)
	if !reflect.DeepEqual(router.Names(), []string{"work", "personal"}) {
		t.Fatal("invalid browser was listed or declaration order changed")
	}
	if err := router.Force("missing"); err == nil {
		t.Fatal("unavailable override accepted")
	}
	if err := router.Force("work"); err != nil {
		t.Fatal(err)
	}
	cfg.Browsers = cfg.Browsers[1:]
	router.Update(cfg)
	if router.Forced() != "" {
		t.Fatal("removed override did not return to Auto")
	}
}

func TestLaunchArgumentsAndMultipleURLs(t *testing.T) {
	cfg := configuration(t, "default: work\n")
	var calls [][]string
	router := New(cfg, t.TempDir(), func(_ string, args []string) error {
		calls = append(calls, args)
		return nil
	})
	request := Request{URLs: []string{"https://example.com/?q=a&x=b", "file:///not-a-web-url", "https://example.com/second"}}
	if err := router.Open(request); err == nil {
		t.Fatal("invalid URL did not report an error")
	}
	if len(calls) != 2 || !reflect.DeepEqual(calls[0], []string{"--profile=Example Profile", request.URLs[0]}) || len(cfg.Browsers[0].Args) != 1 {
		t.Fatal("URL boundaries changed, valid links were dropped, or configuration mutated")
	}
}

func TestConcurrentRoutingAndOverride(t *testing.T) {
	cfg := configuration(t, "default: personal\n")
	router := New(cfg, t.TempDir(), func(string, []string) error { return nil })
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for range 30 {
				_ = router.Force("work")
				_ = router.Open(Request{URLs: []string{"https://example.com/"}})
				_ = router.Force("")
				router.Update(cfg)
			}
		})
	}
	workers.Wait()
}
