package app

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTrayIcon(t *testing.T) {
	for _, tt := range []struct{ requested, want int }{
		{16, 16}, {20, 20}, {24, 24}, {28, 28}, {32, 32},
		{36, 36}, {40, 40}, {44, 44}, {48, 48}, {56, 56}, {64, 64},
		{128, 128}, {256, 256}, {22, 22}, {23, 24}, {72, 128},
	} {
		data, err := trayIconForSize(tt.requested)
		if err != nil {
			t.Fatal(err)
		}
		icon, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		bounds := icon.Bounds()
		if bounds.Dx() != tt.want || bounds.Dy() != tt.want {
			t.Fatalf("requested %d: icon size = %v; want %dx%d", tt.requested, bounds.Size(), tt.want, tt.want)
		}
		var transparent, visible bool
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, alpha := icon.At(x, y).RGBA()
				transparent = transparent || alpha == 0
				visible = visible || alpha > 0
			}
		}
		if !transparent || !visible {
			t.Fatalf("icon %d must have visible artwork and a transparent background", tt.want)
		}
	}
	if _, err := trayIconForSize(257); err == nil {
		t.Fatal("expected error for unsupported icon size")
	}
}

func TestTrayQuitDrainsUpdatesWithoutBlockingUI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updatesDone := make(chan struct{})
	removed := make(chan struct{})
	var removals atomic.Int32
	quit := trayQuit(cancel, updatesDone, func() {
		if removals.Add(1) == 1 {
			close(removed)
		}
	})
	callbackDone := make(chan struct{})
	go func() {
		quit()
		quit()
		close(callbackDone)
	}()
	select {
	case <-callbackDone:
	case <-time.After(time.Second):
		close(updatesDone)
		t.Fatal("Quit blocked the UI callback while a native update was in flight")
	}
	if ctx.Err() == nil {
		t.Fatal("Quit did not cancel background updates")
	}
	select {
	case <-removed:
		t.Fatal("tray removed before pending updates finished")
	default:
	}
	close(updatesDone)
	select {
	case <-removed:
	case <-time.After(time.Second):
		t.Fatal("tray was not removed after updates finished")
	}
	if removals.Load() != 1 {
		t.Fatal("tray removed more than once")
	}
}

func TestCLI(t *testing.T) {
	version, err := os.ReadFile(filepath.Join("..", "..", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "dynamicbrowser.exe")
	flags := "-X main.version=" + strings.TrimSpace(string(version)) + " -X github.com/MrMaxie/dynamicbrowser/internal/app.instanceName=" + instanceName
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
		for _, args := range [][]string{nil, {"--force=false"}} {
			output, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
			if err != nil || len(output) != 0 {
				t.Fatalf("duplicate %v: %v, output=%q", args, err, output)
			}
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(binary), "config.yaml")); !os.IsNotExist(err) {
			t.Fatalf("duplicate touched config: %v", err)
		}
	})

	for _, tt := range []struct {
		name         string
		args         []string
		want         string
		wantContains string
		wantErr      bool
	}{
		{name: "version", args: []string{"--version"}, want: strings.TrimSpace(string(version)) + "\n"},
		{name: "version with force", args: []string{"--force", "--version"}, want: strings.TrimSpace(string(version)) + "\n"},
		{name: "version with registration", args: []string{"--register", "--version"}, want: strings.TrimSpace(string(version)) + "\n"},
		{name: "version with restore", args: []string{"--restore", "--version"}, want: strings.TrimSpace(string(version)) + "\n"},
		{name: "exclusive setup actions", args: []string{"--register", "--restore"}, wantErr: true},
		{name: "setup excludes force", args: []string{"--register", "--force"}, wantErr: true},
		{name: "help", args: []string{"--help"}, wantContains: "--force"},
		{name: "help with force", args: []string{"--force", "--help"}, wantContains: "Usage: dynamicbrowser"},
		{name: "invalid force", args: []string{"--force=invalid"}, wantErr: true},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: true},
		{name: "version with URL", args: []string{"https://example.com/", "--version"}, want: strings.TrimSpace(string(version)) + "\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, binary, tt.args...).CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("exit error: %v; want error: %v\n%s", err, tt.wantErr, output)
			}
			if !tt.wantErr {
				if tt.wantContains != "" {
					if !strings.Contains(string(output), tt.wantContains) {
						t.Errorf("output = %q; want substring %q", output, tt.wantContains)
					}
				} else if string(output) != tt.want {
					t.Errorf("output = %q; want %q", output, tt.want)
				}
			}
		})
	}
}
