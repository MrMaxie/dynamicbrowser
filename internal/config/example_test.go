package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExampleConfiguration(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "config.example.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	values, err := parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateConfiguration(values); err != nil {
		t.Fatal(err)
	}
	if _, err := decode(data); err != nil {
		t.Fatal(err)
	}
}
