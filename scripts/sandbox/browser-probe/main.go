package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func main() {
	if os.Getenv("USERNAME") != "WDAGUtilityAccount" {
		os.Exit(1)
	}
	executable, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	file, err := os.OpenFile(filepath.Join(filepath.Dir(executable), "browser-launches.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		os.Exit(1)
	}
	defer func() { _ = file.Close() }()
	if err := json.NewEncoder(file).Encode(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
