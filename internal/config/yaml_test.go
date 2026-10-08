package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseYAML(t *testing.T) {
	for _, tt := range []struct {
		name    string
		content string
		want    map[string]any
		wantErr bool
	}{
		{name: "empty", want: map[string]any{}},
		{name: "comments", content: "# yaml-language-server: $schema=./config.schema.json\n", want: map[string]any{}},
		{name: "mapping", content: "custom: true\nanything:\n  list: [1, 2]\n", want: map[string]any{"custom": true, "anything": map[string]any{"list": []any{1, 2}}}},
		{name: "multiline", content: "custom: |\n  first\n  second\n", want: map[string]any{"custom": "first\nsecond\n"}},
		{name: "aliases", content: "original: &value [1, 2]\ncustom: *value\n", want: map[string]any{"original": []any{1, 2}, "custom": []any{1, 2}}},
		{name: "syntax", content: "custom: [", wantErr: true},
		{name: "duplicates", content: "custom: first\ncustom: second\n", wantErr: true},
		{name: "sequence root", content: "- first\n", wantErr: true},
		{name: "scalar root", content: "first\n", wantErr: true},
		{name: "null root", content: "null\n", wantErr: true},
		{name: "second document", content: "custom: first\n---\ncustom: second\n", wantErr: true},
		{name: "invalid trailing document", content: "custom: first\n---\ncustom: [\n", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parse error = %v, want error %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("values = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestInvalidYAMLRetainsLastValidConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	store := &Store{}
	for _, content := range []string{"custom: first\n", "custom: second\n---\nignored: true\n", "custom: second\ncustom: third\n"} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		store.reload(path)
		if store.values["custom"] != "first" {
			t.Fatal("invalid YAML replaced the last valid configuration")
		}
		if content != "custom: first\n" && store.err == nil {
			t.Fatal("invalid YAML did not report an error")
		}
	}
}
