package registration

import (
	"errors"
	"strings"
	"testing"
)

var testOwn = defaults{HTTP: "dynamicbrowser.Url.HTTP", HTTPS: "dynamicbrowser.Url.HTTPS"}
var testPrevious = defaults{HTTP: "first.HTTP", HTTPS: "second.HTTPS"}

type fakePlatform struct {
	selected, saved defaults
	hasSaved        bool
	installed       bool
	removed         bool
	confirm         bool
	next            defaults
	assistCalls     int
	saveCalls       int
	failSave        bool
	failInstall     bool
	failRemove      bool
	unavailable     string
	message         string
	afterDescribe   *defaults
	afterInstall    *defaults
	queryCalls      int
	failQueryAt     int
	failPrevious    bool
	failAssist      bool
}

func (f *fakePlatform) current() (defaults, error) {
	f.queryCalls++
	if f.queryCalls == f.failQueryAt {
		return defaults{}, errors.New("query failed")
	}
	return f.selected, nil
}
func (f *fakePlatform) previous() (defaults, bool, error) {
	if f.failPrevious {
		return defaults{}, false, errors.New("snapshot read failed")
	}
	return f.saved, f.hasSaved, nil
}
func (f *fakePlatform) savePrevious(previous defaults) error {
	if f.failSave {
		return errors.New("save failed")
	}
	f.saved, f.hasSaved = previous, true
	f.saveCalls++
	return nil
}
func (f *fakePlatform) install() error {
	if f.failInstall {
		return errors.New("install failed")
	}
	f.installed = true
	if f.afterInstall != nil {
		f.selected = *f.afterInstall
	}
	return nil
}
func (f *fakePlatform) remove() error {
	if f.failRemove {
		return errors.New("remove failed")
	}
	f.removed, f.hasSaved = true, false
	return nil
}
func (f *fakePlatform) describe(id string) (string, error) {
	if f.afterDescribe != nil {
		f.selected = *f.afterDescribe
		f.afterDescribe = nil
	}
	if id == "" || id == f.unavailable {
		return "", errors.New("browser unavailable")
	}
	return id, nil
}
func (f *fakePlatform) assist(message string, registration bool) (bool, error) {
	f.assistCalls++
	f.message = message
	if f.failAssist {
		return false, errors.New("settings failed")
	}
	if f.confirm {
		f.selected = f.next
	}
	return f.confirm, nil
}

func TestRegisterRequiresVerifiedUserChoice(t *testing.T) {
	for _, tt := range []struct {
		name      string
		confirmed bool
		next      defaults
		want      error
	}{
		{"cancelled", false, defaults{}, ErrCancelled},
		{"confirmed both", true, testOwn, nil},
		{"only HTTP", true, defaults{HTTP: testOwn.HTTP, HTTPS: testPrevious.HTTPS}, errors.New("not verified")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			platform := &fakePlatform{selected: testPrevious, confirm: tt.confirmed, next: tt.next}
			err := (setup{system: platform, own: testOwn}).register()
			if (err == nil) != (tt.want == nil) || (tt.want == ErrCancelled && !errors.Is(err, ErrCancelled)) {
				t.Fatalf("registration: %v", err)
			}
			if !platform.installed || !platform.hasSaved || platform.saved != testPrevious {
				t.Fatal("previous handlers were not retained before registration")
			}
			platform.confirm, platform.next = true, testOwn
			if err := (setup{system: platform, own: testOwn}).register(); err != nil {
				t.Fatal(err)
			}
			if platform.saveCalls != 1 || platform.saved != testPrevious {
				t.Fatal("repeat registration replaced the previous handlers")
			}
		})
	}
}

func TestRegisterRechecksAfterPublication(t *testing.T) {
	platform := &fakePlatform{selected: testOwn, saved: testPrevious, hasSaved: true, afterInstall: &testPrevious}
	if err := (setup{system: platform, own: testOwn}).register(); !errors.Is(err, ErrCancelled) || platform.assistCalls != 1 {
		t.Fatalf("reported success using a stale selection: %v", err)
	}
}

func TestRegisterNeverSnapshotsOwnAssociations(t *testing.T) {
	for _, current := range []defaults{testOwn, {HTTP: "dynamicbrowser.url.https", HTTPS: testPrevious.HTTPS}} {
		platform := &fakePlatform{selected: current}
		if err := (setup{system: platform, own: testOwn}).register(); err == nil || platform.installed || platform.hasSaved {
			t.Fatal("registration captured its own associations")
		}
	}
}

func TestRestoreVerifiesSeparateTargets(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		current, next          defaults
		confirmed, wantSuccess bool
	}{
		{"both protocols", testOwn, testPrevious, true, true},
		{"mixed protocols", defaults{HTTP: testOwn.HTTP, HTTPS: "new.HTTPS"}, defaults{HTTP: testPrevious.HTTP, HTTPS: "new.HTTPS"}, true, true},
		{"wrong saved HTTPS", testOwn, defaults{HTTP: testPrevious.HTTP, HTTPS: "wrong.HTTPS"}, true, false},
		{"independent HTTPS changed in Settings", defaults{HTTP: testOwn.HTTP, HTTPS: "new.HTTPS"}, testPrevious, true, false},
		{"independent HTTP changed in Settings", defaults{HTTP: "new.HTTP", HTTPS: testOwn.HTTPS}, testPrevious, true, false},
		{"cancelled", testOwn, testPrevious, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			platform := &fakePlatform{selected: tt.current, saved: testPrevious, hasSaved: true, installed: true, confirm: tt.confirmed, next: tt.next}
			err := (setup{system: platform, own: testOwn}).restore()
			if (err == nil) != tt.wantSuccess || platform.removed != tt.wantSuccess || platform.hasSaved == tt.wantSuccess {
				t.Fatalf("restore: removed=%v, saved=%v, err=%v", platform.removed, platform.hasSaved, err)
			}
			if tt.name == "mixed protocols" && strings.Contains(platform.message, "HTTPS:") {
				t.Fatal("asked to replace the independently selected HTTPS handler")
			}
			if !tt.confirmed && !errors.Is(err, ErrCancelled) {
				t.Fatal("cancel was reported as another failure")
			}
		})
	}
}

func TestRestorePreservesIndependentChoices(t *testing.T) {
	independent := defaults{HTTP: "new.HTTP", HTTPS: "new.HTTPS"}
	platform := &fakePlatform{selected: independent, saved: testPrevious, hasSaved: true, installed: true}
	if err := (setup{system: platform, own: testOwn}).restore(); err != nil || !platform.removed || platform.assistCalls != 0 || platform.selected != independent {
		t.Fatalf("independent choices changed: %v", err)
	}
}

func TestRestoreRechecksBeforeGuidance(t *testing.T) {
	mixed := defaults{HTTP: testOwn.HTTP, HTTPS: "new.HTTPS"}
	platform := &fakePlatform{selected: testOwn, saved: testPrevious, hasSaved: true, installed: true, confirm: true, next: defaults{HTTP: testPrevious.HTTP, HTTPS: mixed.HTTPS}, afterDescribe: &mixed}
	if err := (setup{system: platform, own: testOwn}).restore(); err != nil || strings.Contains(platform.message, "HTTPS:") {
		t.Fatalf("stale restoration guidance: %v", err)
	}
}

func TestRestoreFailsClosed(t *testing.T) {
	for _, platform := range []*fakePlatform{
		{selected: testOwn, installed: true},
		{selected: testOwn, saved: testOwn, hasSaved: true, installed: true},
		{selected: testOwn, saved: testPrevious, hasSaved: true, installed: true, unavailable: testPrevious.HTTP},
		{selected: testPrevious, saved: testPrevious, hasSaved: true, installed: true, failRemove: true},
	} {
		if err := (setup{system: platform, own: testOwn}).restore(); err == nil || platform.removed {
			t.Fatal("unsafe restoration removed registration")
		}
		if platform.installed && platform.failRemove && !platform.hasSaved {
			t.Fatal("failed cleanup cleared the snapshot")
		}
	}
}

func TestSetupReadAndSettingsFailures(t *testing.T) {
	for _, tt := range []struct {
		name     string
		restore  bool
		platform fakePlatform
	}{
		{"registration initial query", false, fakePlatform{selected: testPrevious, failQueryAt: 1}},
		{"registration snapshot", false, fakePlatform{selected: testPrevious, failPrevious: true}},
		{"registration published query", false, fakePlatform{selected: testPrevious, failQueryAt: 2}},
		{"registration settings", false, fakePlatform{selected: testPrevious, failAssist: true}},
		{"registration verification", false, fakePlatform{selected: testPrevious, confirm: true, next: testOwn, failQueryAt: 3}},
		{"restore initial query", true, fakePlatform{selected: testOwn, failQueryAt: 1}},
		{"restore snapshot", true, fakePlatform{selected: testOwn, failPrevious: true}},
		{"restore guidance query", true, fakePlatform{selected: testOwn, failQueryAt: 2}},
		{"restore settings", true, fakePlatform{selected: testOwn, failAssist: true}},
		{"restore verification", true, fakePlatform{selected: testOwn, confirm: true, next: testPrevious, failQueryAt: 3}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			platform := &tt.platform
			platform.saved, platform.hasSaved = testPrevious, true
			operation := setup{system: platform, own: testOwn}
			var err error
			if tt.restore {
				err = operation.restore()
			} else {
				err = operation.register()
			}
			if err == nil || platform.removed || !platform.hasSaved || platform.saved != testPrevious {
				t.Fatalf("failure did not retain registration state: %v", err)
			}
		})
	}
}

func TestRegisterFailureKeepsStateSafe(t *testing.T) {
	platform := &fakePlatform{selected: testPrevious, failSave: true}
	if err := (setup{system: platform, own: testOwn}).register(); err == nil || platform.installed {
		t.Fatal("registered without a saved snapshot")
	}
	platform.failSave, platform.failInstall = false, true
	if err := (setup{system: platform, own: testOwn}).register(); err == nil || !platform.hasSaved || platform.saved != testPrevious {
		t.Fatal("failed install lost the previous handlers")
	}
}
