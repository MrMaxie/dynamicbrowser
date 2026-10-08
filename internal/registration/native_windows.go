package registration

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/windows"
)

const (
	applicationName = "dynamicbrowser"
	clientKey       = `Software\Clients\StartMenuInternet\dynamicbrowser`
	capabilitiesKey = clientKey + `\Capabilities`
	stateKey        = `Software\dynamicbrowser\BrowserRegistration`
	registeredApps  = `Software\RegisteredApplications`

	assocNoFixups        = 0x100
	assocFixedProgID     = 0x800
	assocProtocol        = 0x1000
	assocVerify          = 0x40
	assocExecutable      = 2
	assocFriendlyAppName = 4
	assocDelegateExecute = 18
	assocProgID          = 20
)

var ownDefaults = defaults{HTTP: "dynamicbrowser.Url.HTTP", HTTPS: "dynamicbrowser.Url.HTTPS"}

var associationQuery = sync.OnceValue(func() func(uint32, uint32, *uint16, *uint16, *uint16, *uint32) int32 {
	var query func(uint32, uint32, *uint16, *uint16, *uint16, *uint32) int32
	purego.RegisterFunc(&query, windows.NewLazySystemDLL("shlwapi.dll").NewProc("AssocQueryStringW").Addr())
	return query
})

func queryAssociation(flags, kind uint32, association string) (string, error) {
	name, err := windows.UTF16PtrFromString(association)
	if err != nil {
		return "", err
	}
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	result := associationQuery()(flags|assocNoFixups, kind, name, nil, &buffer[0], &size)
	runtime.KeepAlive(name)
	runtime.KeepAlive(buffer)
	if result != 0 {
		return "", fmt.Errorf("association query failed (HRESULT 0x%08x)", uint32(result))
	}
	return windows.UTF16ToString(buffer), nil
}

func currentDefaults() (defaults, error) {
	http, err := queryAssociation(assocProtocol, assocProgID, "http")
	if err != nil {
		return defaults{}, fmt.Errorf("read current HTTP browser: %w", err)
	}
	https, err := queryAssociation(assocProtocol, assocProgID, "https")
	if err != nil {
		return defaults{}, fmt.Errorf("read current HTTPS browser: %w", err)
	}
	return defaults{HTTP: http, HTTPS: https}, nil
}

func notifyAssociations() {
	_, _, _ = windows.NewLazySystemDLL("shell32.dll").NewProc("SHChangeNotify").Call(0x08000000, 0, 0, 0)
}

func Register() error { return runSetup(false) }
func Restore() error  { return runSetup(true) }

func runSetup(restore bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED|windows.COINIT_DISABLE_OLE1DDE); err != nil && err != windows.Errno(1) {
		return fmt.Errorf("initialize Windows Shell integration: %w", err)
	}
	defer windows.CoUninitialize()
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(`Global\dynamicbrowser.browser-registration.` + user.User.Sid.String())
	if err != nil {
		return err
	}
	mutex, err := windows.CreateMutex(nil, false, name)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	defer func() { _ = windows.CloseHandle(mutex) }()
	result, err := windows.WaitForSingleObject(mutex, 0)
	if err != nil {
		return err
	}
	if result != windows.WAIT_OBJECT_0 && result != windows.WAIT_ABANDONED {
		return errors.New("default-browser setup is already open; finish or cancel it before starting another setup action")
	}
	defer func() { _ = windows.ReleaseMutex(mutex) }()
	operation := setup{system: windowsPlatform{}, own: ownDefaults}
	if restore {
		return operation.restore()
	}
	return operation.register()
}
