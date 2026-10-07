package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "dynamicbrowser.exe")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}

	for _, tt := range []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "greeting", want: "Hello, World!\n"},
		{name: "version", args: []string{"--version"}, want: strings.TrimSpace(version) + "\n"},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := exec.Command(binary, tt.args...).CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("exit error: %v; want error: %v\n%s", err, tt.wantErr, output)
			}
			if !tt.wantErr && string(output) != tt.want {
				t.Errorf("output = %q; want %q", output, tt.want)
			}
		})
	}
}
