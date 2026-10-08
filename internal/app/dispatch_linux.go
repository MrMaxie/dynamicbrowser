package app

import (
	"context"

	"github.com/gogpu/systray"
)

func newTrayDispatch(_ *systray.SystemTray, ctx context.Context) (func(func()) error, func(), error) {
	return func(job func()) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		job()
		return nil
	}, func() {}, nil
}
