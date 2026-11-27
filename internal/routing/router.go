package routing

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
)

type Source struct {
	Process      string
	Window       string
	WindowHint   string `json:"-"`
	RequireKnown bool   `json:",omitempty"`
}

type Request struct {
	URLs   []string
	Source Source
}

type Router struct {
	mu       sync.RWMutex
	config   *config.Config
	browsers []config.Browser
	forced   string
	base     string
	launch   func(string, []string) error
}

func New(configuration *config.Config, base string, launch func(string, []string) error) *Router {
	router := &Router{base: base, launch: launch}
	router.Update(configuration)
	return router
}

func (r *Router) Update(configuration *config.Config) {
	var available []config.Browser
	for _, browser := range configuration.Browsers {
		if browser.Exe == "" {
			continue
		}
		path := browser.Exe
		if !filepath.IsAbs(path) {
			path = filepath.Join(r.base, path)
		}
		resolved, err := exec.LookPath(path)
		if err != nil && !strings.ContainsAny(browser.Exe, `/\`) {
			resolved, err = exec.LookPath(browser.Exe)
		}
		if err == nil {
			browser.Exe = resolved
			available = append(available, browser)
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = configuration
	r.browsers = available
	if _, ok := findBrowser(available, r.forced); !ok {
		r.forced = ""
	}
}

func (r *Router) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.browsers))
	for _, browser := range r.browsers {
		names = append(names, browser.Name)
	}
	return names
}

func (r *Router) Force(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if name != "" {
		if _, ok := findBrowser(r.browsers, name); !ok {
			return fmt.Errorf("browser configuration %q is unavailable", name)
		}
	}
	r.forced = name
	return nil
}

func (r *Router) Forced() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.forced
}

func findBrowser(browsers []config.Browser, name string) (config.Browser, bool) {
	if name == "" {
		return config.Browser{}, false
	}
	for _, browser := range browsers {
		if browser.Name == name {
			return browser, true
		}
	}
	return config.Browser{}, false
}

func (r *Router) Select(raw string, source Source) (config.Browser, error) {
	browser, _, err := r.selectBrowser(raw, &source, nil)
	return browser, err
}

func (r *Router) selectBrowser(raw string, source *Source, lookupWindow func() string) (config.Browser, string, error) {
	destination, err := url.Parse(raw)
	if err != nil || destination.Hostname() == "" || (!strings.EqualFold(destination.Scheme, "http") && !strings.EqualFold(destination.Scheme, "https")) {
		return config.Browser{}, "invalid-url", errors.New("only valid HTTP and HTTPS URLs can be opened")
	}
	r.mu.RLock()
	configuration, browsers, forced := r.config, r.browsers, r.forced
	r.mu.RUnlock()
	if browser, ok := findBrowser(browsers, forced); ok {
		return browser, "override", nil
	}
	domain := strings.ToLower(destination.Hostname())
	match := func(window string) int {
		for i, rule := range configuration.Rules {
			if rule.Source.Process.MatchSource(source.Process, source.RequireKnown) && rule.Source.Window.MatchSource(window, source.RequireKnown) && rule.Target.URL.Match(raw) && rule.Target.Domain.Match(domain) && rule.Target.Path.Match(destination.Path) {
				return i
			}
		}
		return -1
	}
	window := source.Window
	if source.RequireKnown && enabled(configuration.WindowAsProcess) {
		window = source.Process
	}
	i := match(window)
	reason := "rule"
	if i < 0 && source.RequireKnown && source.Process == "" && enabled(configuration.ActiveWindowFallback) && lookupWindow != nil {
		if source.WindowHint = lookupWindow(); source.WindowHint != "" {
			i = match(source.WindowHint)
			reason = "window-hint-rule"
		}
	}
	if i >= 0 {
		if browser, ok := findBrowser(browsers, configuration.Rules[i].Browser); ok {
			return browser, fmt.Sprintf("%s-%d", reason, i+1), nil
		}
	}
	if browser, ok := findBrowser(browsers, configuration.Default); ok {
		return browser, "default", nil
	}
	return config.Browser{}, "no-default", errors.New("no valid default browser is configured; set default to an available browser name in config.yaml")
}

func enabled(setting *bool) bool { return setting == nil || *setting }

func (r *Router) Open(request Request) error {
	return r.OpenObserved(request, nil, nil)
}

// OpenObserved reports decisions without passing URLs or launch arguments to the observer.
func (r *Router) OpenObserved(request Request, lookupWindow func() string, observe func(source Source, browser, reason string)) error {
	if lookupWindow != nil {
		lookupWindow = sync.OnceValue(lookupWindow)
	}
	var failures []error
	for _, raw := range request.URLs {
		source := request.Source
		browser, reason, err := r.selectBrowser(raw, &source, lookupWindow)
		if observe != nil {
			observe(source, browser.Name, reason)
		}
		if err == nil {
			args := append(append([]string(nil), browser.Args...), raw)
			err = r.launch(browser.Exe, args)
		}
		if err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
