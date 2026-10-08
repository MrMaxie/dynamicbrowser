package config

import (
	"os"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestPatterns(t *testing.T) {
	for _, tt := range []struct {
		pattern string
		value   string
		want    bool
	}{
		{"Editor*", "Editor/App", true},
		{"Editor?", "Editoré", true},
		{"Editor?", "Editorab", false},
		{`Editor\*`, "Editor*", true},
		{`Editor\*`, "EditorApp", false},
		{`a\?b`, "a?b", true},
		{"regex:^Viewer[0-9]+$", "Viewer2", true},
		{"regex:^Viewer[0-9]+$", "OtherViewer2", false},
		{"regex:Viewer", "OtherViewer2", true},
		{"[abc]", "[abc]", true},
		{"[abc]", "a", false},
		{"*", "first\nsecond", true},
	} {
		t.Run(tt.pattern+tt.value, func(t *testing.T) {
			pattern, err := compilePattern(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if pattern.MatchString(tt.value) != tt.want {
				t.Fatalf("pattern %q on %q, want %v", tt.pattern, tt.value, tt.want)
			}
		})
	}
	for _, invalid := range []string{"regex:[", `glob\`} {
		if _, err := compilePattern(invalid); err == nil {
			t.Fatalf("accepted invalid pattern %q", invalid)
		}
	}
}

func TestPatternAlternatives(t *testing.T) {
	var patterns Patterns
	if err := yaml.Unmarshal([]byte("[Editor*, 'regex:^Viewer[0-9]+$']"), &patterns); err != nil {
		t.Fatal(err)
	}
	if !patterns.Match("EditorApp") || !patterns.Match("Viewer2") || patterns.Match("Other") {
		t.Fatal("alternatives did not use OR")
	}
	if err := yaml.Unmarshal([]byte("[]"), &patterns); err != nil {
		t.Fatal(err)
	}
	if patterns.Match("EditorApp") {
		t.Fatal("empty alternatives matched")
	}
	for _, invalid := range []string{"42", "[ok, 42]", "{glob: '*'}"} {
		if err := yaml.Unmarshal([]byte(invalid), &Patterns{}); err == nil {
			t.Fatalf("accepted invalid pattern shape %q", invalid)
		}
	}
}

func TestBrowserOrderAndArgs(t *testing.T) {
	configuration, err := decode([]byte("browsers:\n  z:\n    exe: browser.exe\n    args: ['--profile=Example Profile']\n  a:\n    exe: second.exe\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(configuration.Browsers) != 2 || configuration.Browsers[0].Name != "z" || configuration.Browsers[1].Name != "a" || configuration.Browsers[0].Args[0] != "--profile=Example Profile" {
		t.Fatal("browser declaration order or argument boundaries changed")
	}
	if _, err := decode([]byte("browsers:\n  test:\n    exe: browser.exe\n    args: {profile: value}\n")); err == nil {
		t.Fatal("argument map accepted")
	}
}

func TestInvalidPatternRetainsSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("default: example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	store := &Store{}
	store.reload(path)
	previous, err := store.Snapshot()
	if err != nil || previous.Default != "example" {
		t.Fatal("valid snapshot not loaded")
	}
	if err := os.WriteFile(path, []byte("rules:\n  - browser: example\n    target:\n      url: 'regex:['\n"), 0600); err != nil {
		t.Fatal(err)
	}
	store.reload(path)
	retained, err := store.Snapshot()
	if err == nil || retained != previous {
		t.Fatal("invalid pattern replaced the last valid snapshot")
	}
}
