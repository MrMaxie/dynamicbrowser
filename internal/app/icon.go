package app

import (
	"embed"
	"fmt"
)

//go:embed assets
var trayIcons embed.FS

func trayIconForSize(size int) ([]byte, error) {
	for _, available := range []int{16, 20, 22, 24, 28, 32, 36, 40, 44, 48, 56, 64, 128, 256} {
		if available >= size {
			return trayIcons.ReadFile(fmt.Sprintf("assets/icon-%d.png", available))
		}
	}
	return nil, fmt.Errorf("unsupported tray icon size: %d", size)
}
