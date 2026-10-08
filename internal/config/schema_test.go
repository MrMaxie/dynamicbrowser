package config

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigHasLocalSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	ctx, cancel := context.WithCancel(context.Background())
	store := &Store{}
	done, err := Watch(ctx, path, store)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()
	<-done
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != initialConfiguration || len(store.values) != 0 || store.err != nil {
		t.Fatal("new configuration is not an empty YAML mapping with a schema modeline")
	}
	data, err = os.ReadFile(filepath.Join(filepath.Dir(path), "config.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, schema) || !json.Valid(data) {
		t.Fatal("local schema differs from the embedded JSON schema")
	}
}

func TestExistingSchemaIsPreserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	schemaPath := filepath.Join(filepath.Dir(path), "config.schema.json")
	const content = `{"type":"object","additionalProperties":true}`
	if err := os.WriteFile(schemaPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done, err := Watch(ctx, path, &Store{})
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()
	<-done
	data, err := os.ReadFile(schemaPath)
	if err != nil || string(data) != content {
		t.Fatalf("existing schema was modified: %v", err)
	}
}
