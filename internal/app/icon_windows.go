package app

func trayIcon() ([]byte, error) {
	return trayIconForSize(nativeTrayIconSize())
}
