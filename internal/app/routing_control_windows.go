package app

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func routingPipe() (string, string, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", "", err
	}
	var session uint32
	if err := windows.ProcessIdToSessionId(windows.GetCurrentProcessId(), &session); err != nil {
		return "", "", err
	}
	sid := user.User.Sid.String()
	identity := sha256.Sum256(fmt.Appendf(nil, "%s/%s/%d", instanceName, sid, session))
	return fmt.Sprintf(`\\.\pipe\dynamicbrowser-%x`, identity[:16]), "D:P(A;;GA;;;" + sid + ")", nil
}

func listenRouting() (net.Listener, error) {
	name, security, err := routingPipe()
	if err != nil {
		return nil, err
	}
	return winio.ListenPipe(name, &winio.PipeConfig{SecurityDescriptor: security, InputBufferSize: 4096, OutputBufferSize: 4096})
}

func dialRouting() (net.Conn, error) {
	name, _, err := routingPipe()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	conn, err := winio.DialPipeContext(ctx, name)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, context.DeadlineExceeded) {
		return nil, errRestartNotReady
	}
	return conn, err
}
