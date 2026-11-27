//go:build !windows && !linux

package registration

import "errors"

var errUnsupported = errors.New("default-browser setup is supported on Windows and Linux")

func Register() error { return errUnsupported }
func Restore() error  { return errUnsupported }
