package config

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.yaml.in/yaml/v3"
)

// Path returns config.yaml next to the running executable.
func Path() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), "config.yaml"), nil
}

// Store retains the last valid YAML configuration.
type Store struct {
	mu      sync.RWMutex
	values  map[string]any
	config  *Config
	err     error
	changes chan struct{}
}

// Changes returns a single-consumer stream of coalesced reload/error events.
func (c *Store) Changes() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.changes == nil {
		c.changes = make(chan struct{}, 1)
	}
	return c.changes
}

func (c *Store) notifyLocked() {
	select {
	case c.changes <- struct{}{}:
	default:
	}
}

func (c *Store) reload(path string) {
	data, err := os.ReadFile(path)
	var values map[string]any
	var configuration *Config
	if err == nil {
		values, err = parse(data)
	}
	if err == nil {
		err = validateConfiguration(values)
	}
	if err == nil {
		configuration, err = decode(data)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	if err == nil {
		c.values = values
		c.config = configuration
	}
	c.notifyLocked()
}

// Snapshot is immutable after publication, including its slices and patterns.
func (c *Store) Snapshot() (*Config, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config, c.err
}

func Load(path string) (*Config, error) {
	if err := ensureFiles(path); err != nil {
		return nil, err
	}
	store := &Store{}
	store.reload(path)
	return store.Snapshot()
}

func parse(data []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); errors.Is(err, io.EOF) {
		return make(map[string]any), nil
	} else if err != nil {
		return nil, err
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("configuration must be a YAML mapping")
	}
	values := make(map[string]any)
	if err := document.Decode(&values); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("configuration must contain only one YAML document")
	}
	return values, nil
}

// Status describes the current configuration health for the tray UI.
func (c *Store) Status() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return "Configuration: error (last valid configuration retained)"
	}
	return "Configuration: loaded"
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
	if err := ensureFiles(path); err != nil {
		_ = watcher.Close()
		return nil, err
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
				store.notifyLocked()
				store.mu.Unlock()
			}
		}
	}()
	return done, nil
}
