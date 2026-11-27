package app

import (
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestAboutTaskDialogLayout(t *testing.T) {
	cfg, texts, err := aboutTaskDialog("0.9.1", aboutCallback)
	if err != nil {
		t.Fatal(err)
	}
	pointerSize := unsafe.Sizeof(uintptr(0))
	if want := uint32(32 + 16*pointerSize); cfg.Size != want {
		t.Fatalf("TASKDIALOGCONFIG size = %d, want %d", cfg.Size, want)
	}
	if unsafe.Offsetof(cfg.Parent) != 4 || unsafe.Offsetof(cfg.Content) != 12+5*pointerSize || unsafe.Offsetof(cfg.Callback) != 28+14*pointerSize {
		t.Fatal("TASKDIALOGCONFIG does not match the packed Windows ABI")
	}
	if cfg.Flags&tdfEnableHyperlinks == 0 || cfg.CommonButtons != tdcbfOKButton {
		t.Fatal("missing hyperlink support or OK button")
	}
	want := []string{"About DynamicBrowser", "DynamicBrowser", "Version 0.9.1\n\n" + aboutDescription, aboutFooter}
	for i, text := range texts {
		if got := windows.UTF16PtrToString(text); got != want[i] {
			t.Errorf("dialog field %d = %q, want %q", i, got, want[i])
		}
	}
}

// Exercise the real API with Common Controls v6, closing the dialog as soon as
// it is created. No tray, browser, or Windows association is changed.
func TestAboutNativeTaskDialog(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	manifest, err := filepath.Abs(filepath.Join("..", "..", "assets", "app.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := windows.UTF16PtrFromString(manifest)
	if err != nil {
		t.Fatal(err)
	}
	// ACTCTXW uses the standard ABI (unlike TASKDIALOGCONFIG).
	activation := struct {
		Size                  uint32
		Flags                 uint32
		Source                *uint16
		ProcessorArchitecture uint16
		Language              uint16
		AssemblyDirectory     *uint16
		ResourceName          *uint16
		ApplicationName       *uint16
		Module                windows.Handle
	}{Source: source}
	activation.Size = uint32(unsafe.Sizeof(activation))
	kernel := windows.NewLazySystemDLL("kernel32.dll")
	handle, _, err := kernel.NewProc("CreateActCtxW").Call(uintptr(unsafe.Pointer(&activation)))
	runtime.KeepAlive(source)
	if handle == ^uintptr(0) {
		t.Fatalf("CreateActCtxW: %v", err)
	}
	defer func() { _, _, _ = kernel.NewProc("ReleaseActCtx").Call(handle) }()
	var cookie uintptr
	ok, _, err := kernel.NewProc("ActivateActCtx").Call(handle, uintptr(unsafe.Pointer(&cookie)))
	if ok == 0 {
		t.Fatalf("ActivateActCtx: %v", err)
	}
	defer func() { _, _, _ = kernel.NewProc("DeactivateActCtx").Call(0, cookie) }()

	created := false
	var callbackErr error
	callback := windows.NewCallback(func(hwnd uintptr, notification uint32, _, _, _ uintptr) uintptr {
		if notification == 0 { // TDN_CREATED
			created = true
			callbackErr = postNativeMessage(windows.HWND(hwnd), 0x0400+102, 1, 0) // TDM_CLICK_BUTTON, IDOK
		}
		return 0
	})
	cfg, texts, err := aboutTaskDialog("0.9.1", callback)
	if err != nil {
		t.Fatal(err)
	}
	if err := taskDialogIndirect.Find(); err != nil {
		t.Fatal(err)
	}
	var button int32
	result, _, _ := taskDialogIndirect.Call(uintptr(unsafe.Pointer(cfg)), uintptr(unsafe.Pointer(&button)), 0, 0)
	runtime.KeepAlive(texts)
	if int32(result) < 0 {
		t.Fatalf("TaskDialogIndirect: HRESULT 0x%08X", uint32(result))
	}
	if callbackErr != nil || !created || button != 1 {
		t.Fatalf("created=%v, button=%d, callback error=%v", created, button, callbackErr)
	}
}
