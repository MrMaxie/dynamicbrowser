package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestBrowserRoutingHelper(t *testing.T) {
	path := os.Getenv("DYNAMICBROWSER_TEST_BROWSER_LOG")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if err := json.NewEncoder(file).Encode(os.Args); err != nil {
		t.Fatal(err)
	}
}

func TestRoutingCLIAndTray(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "dynamicbrowser.exe")
	flags := "-H=windowsgui -X github.com/MrMaxie/dynamicbrowser/internal/app.instanceName=" + instanceName
	if output, err := exec.Command("go", "build", "-ldflags", flags, "-o", binary, "../../cmd/dynamicbrowser").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	helper, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "browser.jsonl")
	t.Setenv("DYNAMICBROWSER_TEST_BROWSER_LOG", log)
	path := filepath.Join(dir, "config.yaml")
	writeConfig := func(second string) {
		t.Helper()
		name, _ := json.Marshal(helper)
		data := "default: first\nbrowsers:\n  first:\n    exe: " + string(name) + "\n    args: ['-test.run=^TestBrowserRoutingHelper$', '--', 'first profile']\n  second:\n    exe: " + string(name) + "\n    args: ['-test.run=^TestBrowserRoutingHelper$', '--', '" + second + "']\n"
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeConfig("second profile")
	invoke := func(url string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if output, err := exec.CommandContext(ctx, binary, url).CombinedOutput(); err != nil {
			t.Fatalf("route: %v, %s", err, output)
		}
	}
	expectLaunch := func(count int, profile, url string) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			data, _ := os.ReadFile(log)
			var records [][]string
			decoder := json.NewDecoder(bytes.NewReader(data))
			for {
				var args []string
				if decoder.Decode(&args) != nil {
					break
				}
				records = append(records, args)
			}
			if len(records) == count {
				args := records[count-1]
				if !reflect.DeepEqual(args[len(args)-2:], []string{profile, url}) {
					t.Fatalf("browser arguments: %q", args)
				}
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatal("browser helper did not receive expected arguments")
	}
	invoke("https://example.com/on-demand")
	expectLaunch(1, "first profile", "https://example.com/on-demand")
	release, acquired, err := acquireInstance()
	if err != nil || !acquired {
		t.Fatalf("on-demand retained a resident singleton: %v, %v", acquired, err)
	}
	release()
	resident := startTestApp(t, binary)
	hwnd := waitForProcessWindow(t, messageOnlyWindow, systrayClass, uint32(resident.cmd.Process.Pid))
	invoke("https://example.com/resident")
	expectLaunch(2, "first profile", "https://example.com/resident")
	if err := postNativeMessage(hwnd, systrayCallback, 0, wmLeftButtonUp); err != nil {
		t.Fatal(err)
	}
	waitForProcessWindow(t, 0, "#32768", uint32(resident.cmd.Process.Pid))
	writeConfig("reloaded profile")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte("default: first"), []byte("default: second"), 1)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	_ = postNativeMessage(hwnd, wmCancelMode, 0, 0)
	invoke("https://example.com/reloaded")
	expectLaunch(3, "reloaded profile", "https://example.com/reloaded")
	_ = postNativeMessage(hwnd, wmClose, 0, 0)
	select {
	case <-resident.done:
		if resident.err != nil {
			t.Fatal(resident.err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("tray did not exit")
	}
	invoke("https://example.com/after-close")
	expectLaunch(4, "reloaded profile", "https://example.com/after-close")
	if err := os.WriteFile(path, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	failed := startTestApp(t, binary, "https://example.com/no-default")
	dialog := waitForProcessWindow(t, 0, "#32770", uint32(failed.cmd.Process.Pid))
	button := waitForProcessWindow(t, dialog, "Button", uint32(failed.cmd.Process.Pid))
	if err := postNativeMessage(button, 0x00f5, 0, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-failed.done:
		if failed.err == nil {
			t.Fatal("missing default reported success")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("error dialog did not close")
	}
}
