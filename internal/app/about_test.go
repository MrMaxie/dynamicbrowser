package app

import (
	"strings"
	"testing"
)

func TestAboutText(t *testing.T) {
	for _, version := range []string{"0.9.1", " 0.9.1\n", "", "dev"} {
		wantVersion := strings.TrimSpace(version)
		if wantVersion == "" {
			wantVersion = "dev"
		}
		text := aboutText(version)
		for _, want := range []string{"DynamicBrowser", "Version " + wantVersion + "\n", aboutDescription, "Maciej Mieńko / Maxie", aboutAuthorURL, aboutLicenseURL, aboutRepositoryURL, "Apache-2.0"} {
			if !strings.Contains(text, want) {
				t.Errorf("aboutText(%q) missing %q", version, want)
			}
		}
	}
}

func TestAboutFooterLinks(t *testing.T) {
	want := `<a href="https://maxie.dev">Author: Maciej Mieńko / Maxie</a> · ` +
		`<a href="https://github.com/MrMaxie/dynamicbrowser/blob/master/LICENSE">License: Apache-2.0</a> · ` +
		`<a href="https://github.com/MrMaxie/dynamicbrowser">Repository: GitHub</a>`
	if aboutFooter != want {
		t.Fatalf("unexpected About footer: %s", aboutFooter)
	}
	for _, url := range []string{aboutAuthorURL, aboutLicenseURL, aboutRepositoryURL} {
		if !isAboutLink(url) {
			t.Errorf("About link rejected: %s", url)
		}
	}
	for _, url := range []string{"", "file:///C:/Windows/system32/cmd.exe", "https://example.com", aboutAuthorURL + "/unexpected"} {
		if isAboutLink(url) {
			t.Errorf("unexpected About link allowed: %s", url)
		}
	}
}
