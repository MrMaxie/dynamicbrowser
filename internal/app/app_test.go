package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLI(t *testing.T) {
	version, err := os.ReadFile(filepath.Join("..", "..", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "dynamicbrowser.exe")
	flags := "-X main.version=" + strings.TrimSpace(string(version))
	if output, err := exec.Command("go", "build", "-ldflags", flags, "-o", binary, "../../cmd/dynamicbrowser").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}

	t.Run("duplicate exits silently before touching config", func(t *testing.T) {
		release, acquired, err := acquireInstance()
		if err != nil || !acquired {
			t.Fatalf("acquire singleton: %v, %v", acquired, err)
		}
		defer release()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary).CombinedOutput()
		if err != nil || len(output) != 0 {
			t.Fatalf("duplicate: %v, output=%q", err, output)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(binary), "config.toml")); !os.IsNotExist(err) {
			t.Fatalf("duplicate touched config: %v", err)
		}
	})

	for _, tt := range []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "version", args: []string{"--version"}, want: strings.TrimSpace(string(version)) + "\n"},
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
