package app

import "strings"

const (
	aboutAuthorURL     = "https://maxie.dev"
	aboutLicenseURL    = "https://github.com/MrMaxie/dynamicbrowser/blob/master/LICENSE"
	aboutRepositoryURL = "https://github.com/MrMaxie/dynamicbrowser"
	aboutDescription   = "Routes links to the appropriate browser based on URL and source of the request."
	aboutFooter        = `<a href="` + aboutAuthorURL + `">Author: Maciej Mieńko / Maxie</a> · ` +
		`<a href="` + aboutLicenseURL + `">License: Apache-2.0</a> · ` +
		`<a href="` + aboutRepositoryURL + `">Repository: GitHub</a>`
)

func aboutContent(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		version = "dev"
	}
	return "Version " + version + "\n\n" + aboutDescription
}

func aboutText(version string) string {
	return "DynamicBrowser\n" + aboutContent(version) +
		"\n\nAuthor: Maciej Mieńko / Maxie — " + aboutAuthorURL +
		"\nLicense: Apache-2.0 — " + aboutLicenseURL +
		"\nRepository: GitHub — " + aboutRepositoryURL
}

func isAboutLink(url string) bool {
	return url == aboutAuthorURL || url == aboutLicenseURL || url == aboutRepositoryURL
}
