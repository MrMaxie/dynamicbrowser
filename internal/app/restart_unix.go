//go:build linux || darwin

package app

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

func restartSocketPath() (string, error) {
	dir, err := instanceDirectory()
	if err != nil {
		return "", err
	}
	identity := sha256.Sum256([]byte(instanceName))
	return filepath.Join(dir, fmt.Sprintf("%x.sock", identity[:8])), nil
}

func startInstanceControl(quit func()) (func(), error) {
	path, err := restartSocketPath()
	if err != nil {
		return nil, err
	}
	// Only the lock owner removes a socket left behind by a terminated instance.
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
	done := make(chan struct{})
	go func() {
		defer close(done)
		request, stop := restartAction(quit)
		defer stop()
		for {
			conn, err := listener.AcceptUnix()
			if err != nil {
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			var command [8]byte
			_, err = io.ReadFull(conn, command[:])
			_ = conn.Close()
			if err == nil && string(command[:]) == "restart\n" {
				request()
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			_ = listener.Close()
			<-done
		})
	}, nil
}

func requestInstanceRestart() error {
	path, err := restartSocketPath()
	if err != nil {
		return err
	}
	conn, err := net.DialTimeout("unix", path, 500*time.Millisecond)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ECONNREFUSED) {
		return errRestartNotReady
	}
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if err := conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		return err
	}
	_, err = io.WriteString(conn, "restart\n")
	return err
}
