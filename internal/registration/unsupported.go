//go:build !windows

package registration

import "errors"

var errUnsupported = errors.New("default-browser setup is supported on Windows only")

func Register() error { return errUnsupported }
func Restore() error  { return errUnsupported }
