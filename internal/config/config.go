package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

// Path returns config.toml next to the running executable.
func Path() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "config.toml"), nil
}

// Store retains the last valid, schema-free TOML document.
type Store struct {
	mu     sync.RWMutex
	values map[string]any
	err    error
}

func (c *Store) reload(path string) {
	data, err := os.ReadFile(path)
	values := make(map[string]any)
	if err == nil {
		err = toml.Unmarshal(data, &values)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	if err == nil {
		c.values = values
	}
}

// Status describes the current configuration health for the tray UI.
func (c *Store) Status() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return "config.toml error. Fix the file; the last valid configuration remains active."
	}
	return "Application is running. config.toml loaded; watching for changes."
}

// Watch loads the configuration and watches for changes until ctx is canceled.
// The returned channel closes after the watcher has released its resources.
func Watch(ctx context.Context, path string, store *Store) (<-chan struct{}, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	// Watch the directory so editor atomic saves and file recreation keep working.
	if err = watcher.Add(filepath.Dir(path)); err != nil {
		_ = watcher.Close()
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err == nil {
		err = file.Close()
	}
	if err != nil && !errors.Is(err, os.ErrExist) {
		_ = watcher.Close()
		return nil, fmt.Errorf("create config.toml: %w", err)
	}
	store.reload(path)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = watcher.Close() }()
		var timer *time.Timer
		var pending <-chan time.Time
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.EqualFold(filepath.Clean(event.Name), filepath.Clean(path)) && event.Has(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) {
					if timer == nil {
						timer = time.NewTimer(150 * time.Millisecond)
					} else {
						timer.Reset(150 * time.Millisecond)
					}
					pending = timer.C
				}
			case <-pending:
				pending = nil
				store.reload(path)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				store.mu.Lock()
				store.err = err
				store.mu.Unlock()
			}
		}
	}()
	return done, nil
}
