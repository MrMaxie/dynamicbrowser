package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInvalidShapeRetainsConfiguration(t *testing.T) {
	for _, invalid := range []string{
		"rules: [{browser: example, target: {domain: null}}]",
		"rules: [{browser: example, source: null}]",
		"rules: [{browser: example, target: {domain: {glob: '*'}}}]",
		"rules: [{browser: example, target: {path_prefix: '/example'}}]",
		"browsers: {example: {exe: browser.exe, executable: old.exe}}",
		"browsers: {example: {exe: browser.exe, args: null}}",
		"browsers: {example: {exe: 123}}",
		"browsers: {example: {exe: browser.exe, args: [123]}}",
		"default: 123",
		"autostart: null",
	} {
		t.Run(invalid, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte("default: example\n"), 0600); err != nil {
				t.Fatal(err)
			}
			store := &Store{}
			store.reload(path)
			previous, err := store.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(invalid), 0600); err != nil {
				t.Fatal(err)
			}
			store.reload(path)
			retained, err := store.Snapshot()
			if err == nil || retained != previous {
				t.Fatal("invalid shape replaced last valid configuration")
			}
		})
	}
}

func TestBrowserAliasesAndMerges(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  Browsers
	}{
		{"alias key", "name: &name example\nbrowsers:\n  *name: {exe: browser.exe}\n", Browsers{{Name: "example", Exe: "browser.exe"}}},
		{"merge mapping", "base: &base {z: {exe: z.exe}, a: {exe: a.exe}}\nbrowsers:\n  <<: *base\n  last: {exe: last.exe}\n", Browsers{{Name: "z", Exe: "z.exe"}, {Name: "a", Exe: "a.exe"}, {Name: "last", Exe: "last.exe"}}},
		{"merge sequence and override", "base: &base {z: {exe: inherited.exe}, a: {exe: a.exe}}\nother: &other {z: {exe: ignored.exe}, b: {exe: b.exe}}\nbrowsers:\n  <<: [*base, *other]\n  z: {exe: explicit.exe}\n", Browsers{{Name: "z", Exe: "explicit.exe"}, {Name: "a", Exe: "a.exe"}, {Name: "b", Exe: "b.exe"}}},
		{"explicit before merge", "base: &base {z: {exe: inherited.exe}, a: {exe: a.exe}}\nbrowsers:\n  z: {exe: explicit.exe}\n  <<: *base\n", Browsers{{Name: "z", Exe: "explicit.exe"}, {Name: "a", Exe: "a.exe"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.input), 0600); err != nil {
				t.Fatal(err)
			}
			store := &Store{}
			store.reload(path)
			cfg, err := store.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cfg.Browsers, tt.want) {
				t.Fatalf("browsers = %#v, want %#v", cfg.Browsers, tt.want)
			}
			if err := os.WriteFile(path, []byte("browsers: {<<: invalid}\n"), 0600); err != nil {
				t.Fatal(err)
			}
			store.reload(path)
			if retained, err := store.Snapshot(); err == nil || retained != cfg {
				t.Fatal("invalid merge replaced the last valid configuration")
			}
		})
	}
}

func TestAliasesInPatternsAndArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	input := "example: &value 'example*'\nbrowsers:\n  example:\n    exe: browser.exe\n    args: [*value]\nrules:\n  - browser: example\n    target:\n      domain: [*value]\n"
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Browsers[0].Args[0] != "example*" || !cfg.Rules[0].Target.Domain.Match("example.com") {
		t.Fatal("YAML string alias did not decode")
	}
}
