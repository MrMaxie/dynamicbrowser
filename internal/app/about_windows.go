package app

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	tdfEnableHyperlinks        = 0x0001
	tdfAllowDialogCancellation = 0x0008
	tdfSizeToContent           = 0x01000000
	tdcbfOKButton              = 0x0001
	tdnHyperlinkClicked        = 3
)

var (
	taskDialogIndirect = windows.NewLazySystemDLL("comctl32.dll").NewProc("TaskDialogIndirect")
	aboutCallback      = windows.NewCallback(aboutDialogCallback)
)

// TASKDIALOGCONFIG uses one-byte packing in commctrl.h. Pointer-sized arrays
// of uint32 preserve its layout without Go's implicit pointer alignment.
// The pointed-to UTF-16 strings are retained separately during the native call.
type taskDialogPointer [unsafe.Sizeof(uintptr(0)) / 4]uint32

func packedTaskDialogPointer(value uintptr) taskDialogPointer {
	var pointer taskDialogPointer
	for i := range pointer {
		pointer[i] = uint32(value >> (32 * i))
	}
	return pointer
}

type taskDialogConfig struct {
	Size                 uint32
	Parent               taskDialogPointer
	Instance             taskDialogPointer
	Flags                uint32
	CommonButtons        uint32
	WindowTitle          taskDialogPointer
	MainIcon             taskDialogPointer
	MainInstruction      taskDialogPointer
	Content              taskDialogPointer
	ButtonCount          uint32
	Buttons              taskDialogPointer
	DefaultButton        int32
	RadioButtonCount     uint32
	RadioButtons         taskDialogPointer
	DefaultRadioButton   int32
	VerificationText     taskDialogPointer
	ExpandedInformation  taskDialogPointer
	ExpandedControlText  taskDialogPointer
	CollapsedControlText taskDialogPointer
	FooterIcon           taskDialogPointer
	Footer               taskDialogPointer
	Callback             taskDialogPointer
	CallbackData         taskDialogPointer
	Width                uint32
}

func aboutTaskDialog(version string, callback uintptr) (*taskDialogConfig, []*uint16, error) {
	cfg := &taskDialogConfig{
		Flags:         tdfEnableHyperlinks | tdfAllowDialogCancellation | tdfSizeToContent,
		CommonButtons: tdcbfOKButton,
		Callback:      packedTaskDialogPointer(callback),
	}
	cfg.Size = uint32(unsafe.Sizeof(*cfg))
	var strings []*uint16
	for _, field := range []struct {
		destination *taskDialogPointer
		text        string
	}{
		{&cfg.WindowTitle, "About DynamicBrowser"},
		{&cfg.MainInstruction, "DynamicBrowser"},
		{&cfg.Content, aboutContent(version)},
		{&cfg.Footer, aboutFooter},
	} {
		text, err := windows.UTF16PtrFromString(field.text)
		if err != nil {
			return nil, nil, err
		}
		strings = append(strings, text)
		*field.destination = packedTaskDialogPointer(uintptr(unsafe.Pointer(text)))
	}
	return cfg, strings, nil
}

func showAbout(version string) error {
	cfg, strings, err := aboutTaskDialog(version, aboutCallback)
	if err != nil {
		return err
	}
	if err := taskDialogIndirect.Find(); err != nil {
		return fmt.Errorf("load TaskDialogIndirect: %w", err)
	}
	result, _, _ := taskDialogIndirect.Call(uintptr(unsafe.Pointer(cfg)), 0, 0, 0)
	runtime.KeepAlive(strings)
	if int32(result) < 0 {
		return fmt.Errorf("TaskDialogIndirect failed (HRESULT 0x%08X)", uint32(result))
	}
	return nil
}

func aboutDialogCallback(hwnd uintptr, notification uint32, _ uintptr, lParam *uint16, _ uintptr) uintptr {
	if notification != tdnHyperlinkClicked || lParam == nil {
		return 0
	}
	url := windows.UTF16PtrToString(lParam)
	if !isAboutLink(url) {
		return 0
	}
	file, err := windows.UTF16PtrFromString(url)
	if err == nil {
		err = windows.ShellExecute(windows.Handle(hwnd), nil, file, nil, nil, windows.SW_SHOWNORMAL)
	}
	if err != nil {
		ShowStartupError(fmt.Errorf("open About link: %w", err))
	}
	return 0
}
