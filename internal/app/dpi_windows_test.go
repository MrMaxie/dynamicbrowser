package app

import (
	"bytes"
	"context"
	"image/png"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestTrayIconNativeSize(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	restore, err := setPerMonitorThreadDPI()
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	data, err := trayIcon()
	if err != nil {
		t.Fatal(err)
	}
	icon, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	size := nativeTrayIconSize()
	if icon.Bounds().Dx() != size || icon.Bounds().Dy() != size {
		t.Fatalf("source PNG %v differs from native HICON size %dx%d", icon.Bounds().Size(), size, size)
	}
}

func TestTrayWindowDPIAndLeftClickMenu(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "dynamicbrowser.exe")
	flags := "-X github.com/MrMaxie/dynamicbrowser/internal/app.instanceName=" + instanceName
	if output, err := exec.Command("go", "build", "-ldflags", flags, "-o", binary, "../../cmd/dynamicbrowser").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var hwnd windows.HWND
	defer func() {
		if hwnd != 0 {
			_ = postNativeMessage(hwnd, wmCancelMode, 0, 0)
			_ = postNativeMessage(hwnd, wmClose, 0, 0)
		} else {
			_ = cmd.Process.Kill()
		}
		if err := cmd.Wait(); err != nil {
			t.Errorf("tray shutdown: %v", err)
		}
	}()
	hwnd = waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(cmd.Process.Pid))
	if awareness := nativeWindowAwareness(hwnd); awareness != 2 {
		t.Fatalf("tray DPI awareness = %d; want per-monitor aware (2)", awareness)
	}
	if err := postNativeMessage(hwnd, systrayCallback, 0, wmLeftButtonUp); err != nil {
		t.Fatal(err)
	}
	waitForProcessWindow(t, 0, "#32768", uint32(cmd.Process.Pid))
}

func waitForProcessWindow(t *testing.T, parent windows.HWND, class string, pid uint32) windows.HWND {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var after windows.HWND
		for {
			hwnd, err := findNativeWindow(parent, after, class)
			if err != nil {
				t.Fatal(err)
			}
			if hwnd == 0 {
				break
			}
			var owner uint32
			_, err = windows.GetWindowThreadProcessId(hwnd, &owner)
			if err == nil && owner == pid {
				return hwnd
			}
			after = hwnd
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("process %d did not create window %q", pid, class)
	return 0
}
