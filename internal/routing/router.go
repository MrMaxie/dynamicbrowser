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
	Process string
	Window  string
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
	destination, err := url.Parse(raw)
	if err != nil || destination.Hostname() == "" || (!strings.EqualFold(destination.Scheme, "http") && !strings.EqualFold(destination.Scheme, "https")) {
		return config.Browser{}, errors.New("only valid HTTP and HTTPS URLs can be opened")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if browser, ok := findBrowser(r.browsers, r.forced); ok {
		return browser, nil
	}
	for _, rule := range r.config.Rules {
		if rule.Source.Process.Match(source.Process) && rule.Source.Window.Match(source.Window) && rule.Target.URL.Match(raw) && rule.Target.Domain.Match(strings.ToLower(destination.Hostname())) && rule.Target.Path.Match(destination.Path) {
			if browser, ok := findBrowser(r.browsers, rule.Browser); ok {
				return browser, nil
			}
			break
		}
	}
	if browser, ok := findBrowser(r.browsers, r.config.Default); ok {
		return browser, nil
	}
	return config.Browser{}, errors.New("no valid default browser is configured; set default to an available browser name in config.yaml")
}

func (r *Router) Open(request Request) error {
	var failures []error
	for _, raw := range request.URLs {
		browser, err := r.Select(raw, request.Source)
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
