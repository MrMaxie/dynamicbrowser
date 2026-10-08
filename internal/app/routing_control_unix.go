//go:build linux || darwin

package app

import (
	"errors"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func routingSocket() (string, error) {
	path, err := restartSocketPath()
	return path + ".routing", err
}

func listenRouting() (net.Listener, error) {
	path, err := routingSocket()
	if err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		return nil, err
	}
	return listener, nil
}

func dialRouting() (net.Conn, error) {
	path, err := routingSocket()
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ECONNREFUSED) {
		return nil, errRestartNotReady
	}
	return conn, err
}
