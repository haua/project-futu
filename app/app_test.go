package app

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
)

func TestToggleEditMode(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{}
	if fw.IsEditMode() {
		t.Fatalf("zero value edit mode should be false")
	}

	if got := fw.ToggleEditMode(); !got {
		t.Fatalf("first toggle should return true")
	}
	if !fw.IsEditMode() {
		t.Fatalf("edit mode should be true after first toggle")
	}

	if got := fw.ToggleEditMode(); got {
		t.Fatalf("second toggle should return false")
	}
	if fw.IsEditMode() {
		t.Fatalf("edit mode should be false after second toggle")
	}
}

func TestToggleEditMode_UpdatesTaskbarVisibility(t *testing.T) {
	t.Parallel()

	var calls []bool
	fw := &FloatingWindow{
		taskbarSet: func(visible bool) bool {
			calls = append(calls, visible)
			return true
		},
	}
	fw.editMode.Store(true)

	if got := fw.ToggleEditMode(); got {
		t.Fatalf("first toggle should return false")
	}
	if got := fw.ToggleEditMode(); !got {
		t.Fatalf("second toggle should return true")
	}

	if len(calls) != 2 {
		t.Fatalf("taskbarSet calls = %d, want 2", len(calls))
	}
	if calls[0] {
		t.Fatalf("first call should hide taskbar (false)")
	}
	if !calls[1] {
		t.Fatalf("second call should show taskbar (true)")
	}
}

func TestToggleEditMode_UpdatesMousePassthrough(t *testing.T) {
	t.Parallel()

	var calls []bool
	fw := &FloatingWindow{
		mouseSet: func(enabled bool) bool {
			calls = append(calls, enabled)
			return true
		},
	}
	fw.editMode.Store(true)

	if got := fw.ToggleEditMode(); got {
		t.Fatalf("first toggle should return false")
	}
	if got := fw.ToggleEditMode(); !got {
		t.Fatalf("second toggle should return true")
	}

	if len(calls) != 2 {
		t.Fatalf("mouseSet calls = %d, want 2", len(calls))
	}
	if !calls[0] {
		t.Fatalf("first call should enable passthrough (true)")
	}
	if calls[1] {
		t.Fatalf("second call should disable passthrough (false)")
	}
}

func TestToggleEditMode_EditModeRestoresOpaque(t *testing.T) {
	t.Parallel()

	var got []float64
	fw := &FloatingWindow{
		opacitySet: func(opacity float64) bool {
			got = append(got, opacity)
			return true
		},
	}
	fw.editMode.Store(false)

	if next := fw.ToggleEditMode(); !next {
		t.Fatalf("toggle from non-edit should return true")
	}
	if len(got) != 1 {
		t.Fatalf("opacitySet calls = %d, want 1", len(got))
	}
	if got[0] != 1 {
		t.Fatalf("opacity value = %v, want 1", got[0])
	}
}

func TestOpacityByCursorDistance(t *testing.T) {
	t.Parallel()

	if got := opacityByCursorDistance(-1, 1); got != 0 {
		t.Fatalf("opacity(-1) = %v, want 0", got)
	}
	if got := opacityByCursorDistance(0, 1); got != 0 {
		t.Fatalf("opacity(0) = %v, want 0", got)
	}
	if got := opacityByCursorDistance(mouseFadeRange/2, 1); math.Abs(got-0.5) > 1e-6 {
		t.Fatalf("opacity(half) = %v, want 0.5", got)
	}
	if got := opacityByCursorDistance(mouseFadeRange, 1); got != 1 {
		t.Fatalf("opacity(range) = %v, want 1", got)
	}
	if got := opacityByCursorDistance(mouseFadeRange+10, 1); got != 1 {
		t.Fatalf("opacity(range+) = %v, want 1", got)
	}
	if got := opacityByCursorDistance(mouseFadeRange, 0.7); math.Abs(got-0.7) > 1e-6 {
		t.Fatalf("opacity(range,max=0.7) = %v, want 0.7", got)
	}
}

func TestCursorDistanceToRect(t *testing.T) {
	t.Parallel()

	pos := fyne.NewPos(100, 200)
	size := fyne.NewSize(300, 150)

	if got := cursorDistanceToRect(fyne.NewPos(150, 220), pos, size); got != 0 {
		t.Fatalf("inside distance = %v, want 0", got)
	}
	if got := cursorDistanceToRect(fyne.NewPos(90, 220), pos, size); got != 10 {
		t.Fatalf("left distance = %v, want 10", got)
	}
	if got := cursorDistanceToRect(fyne.NewPos(150, 180), pos, size); got != 20 {
		t.Fatalf("top distance = %v, want 20", got)
	}
	got := cursorDistanceToRect(fyne.NewPos(90, 190), pos, size)
	if math.Abs(float64(got)-math.Hypot(10, 10)) > 1e-5 {
		t.Fatalf("corner distance = %v, want sqrt(200)", got)
	}
}

func TestSaveWindowPosition(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	fw := &FloatingWindow{App: a}
	pos := fyne.NewPos(12.5, 34.5)
	fw.SaveWindowPosition(pos)

	prefs := a.Preferences()
	if got := float32(prefs.Float(windowPosXKey)); got != pos.X {
		t.Fatalf("saved X = %v, want %v", got, pos.X)
	}
	if got := float32(prefs.Float(windowPosYKey)); got != pos.Y {
		t.Fatalf("saved Y = %v, want %v", got, pos.Y)
	}
	if !prefs.Bool(windowPosSetKey) {
		t.Fatalf("windowPosSetKey should be true")
	}
}

func TestSetMouseFarOpacity_PersistsPreference(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	fw := &FloatingWindow{App: a}
	fw.SetMouseFarOpacity(0.6)

	if got := fw.MouseFarOpacity(); math.Abs(got-0.6) > 0.01 {
		t.Fatalf("MouseFarOpacity() = %v, want about 0.6", got)
	}
	if got := a.Preferences().Float(mouseFarOpacityKey); math.Abs(got-0.6) > 0.01 {
		t.Fatalf("saved mouseFarOpacity = %v, want about 0.6", got)
	}
}

func TestRestoreMouseFarOpacity_DefaultIsOne(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	fw := &FloatingWindow{App: a}
	fw.restoreMouseFarOpacity()

	if got := fw.MouseFarOpacity(); got != 1 {
		t.Fatalf("MouseFarOpacity() = %v, want 1", got)
	}
}

func TestRestoreMouseFarOpacity_FromPreference(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)
	a.Preferences().SetFloat(mouseFarOpacityKey, 0.7)

	fw := &FloatingWindow{App: a}
	fw.restoreMouseFarOpacity()

	if got := fw.MouseFarOpacity(); math.Abs(got-0.7) > 0.01 {
		t.Fatalf("MouseFarOpacity() = %v, want about 0.7", got)
	}
}

func TestSetAlwaysOnTop_NoController(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{}
	if fw.IsAlwaysOnTop() {
		t.Fatalf("zero value always-on-top should be false")
	}

	if ok := fw.SetAlwaysOnTop(true); ok {
		t.Fatalf("SetAlwaysOnTop should fail when topMostCtl is nil")
	}
	if fw.IsAlwaysOnTop() {
		t.Fatalf("state should remain unchanged when SetAlwaysOnTop fails")
	}
}

func TestToggleAlwaysOnTop_NoControllerKeepsState(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{}
	if got := fw.ToggleAlwaysOnTop(); got {
		t.Fatalf("toggle should keep false when topMostCtl is nil")
	}

	fw.alwaysOnTop.Store(true)
	if got := fw.ToggleAlwaysOnTop(); !got {
		t.Fatalf("toggle should keep true when topMostCtl is nil")
	}
}

func TestSetAlwaysOnTop_PersistsPreferenceOnSuccess(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	fw := &FloatingWindow{
		App:        a,
		topMostSet: func(bool) bool { return true },
	}

	if ok := fw.SetAlwaysOnTop(true); !ok {
		t.Fatalf("SetAlwaysOnTop(true) should succeed")
	}
	if !fw.IsAlwaysOnTop() {
		t.Fatalf("always-on-top should be true")
	}
	if !a.Preferences().Bool(alwaysOnTopKey) {
		t.Fatalf("alwaysOnTopKey should be true after success")
	}

	if ok := fw.SetAlwaysOnTop(false); !ok {
		t.Fatalf("SetAlwaysOnTop(false) should succeed")
	}
	if fw.IsAlwaysOnTop() {
		t.Fatalf("always-on-top should be false")
	}
	if a.Preferences().Bool(alwaysOnTopKey) {
		t.Fatalf("alwaysOnTopKey should be false after success")
	}
}

func TestSetAlwaysOnTop_FailureDoesNotOverwritePreference(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)
	a.Preferences().SetBool(alwaysOnTopKey, true)

	fw := &FloatingWindow{
		App:        a,
		topMostSet: func(bool) bool { return false },
	}

	if ok := fw.SetAlwaysOnTop(false); ok {
		t.Fatalf("SetAlwaysOnTop should fail")
	}
	if !a.Preferences().Bool(alwaysOnTopKey) {
		t.Fatalf("failed set should not overwrite saved preference")
	}
}

func TestRestoreAlwaysOnTop_FromPreference(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)
	a.Preferences().SetBool(alwaysOnTopKey, true)

	calls := 0
	fw := &FloatingWindow{
		App: a,
		topMostSet: func(enabled bool) bool {
			calls++
			return enabled
		},
	}

	fw.restoreAlwaysOnTop()

	if calls != 1 {
		t.Fatalf("restore should try to apply once, calls=%d", calls)
	}
	if !fw.IsAlwaysOnTop() {
		t.Fatalf("always-on-top should be restored to true")
	}
}

func TestRestoreAlwaysOnTop_DisabledPreferenceSkipsApply(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)
	a.Preferences().SetBool(alwaysOnTopKey, false)

	called := false
	fw := &FloatingWindow{
		App: a,
		topMostSet: func(bool) bool {
			called = true
			return true
		},
	}
	fw.alwaysOnTop.Store(true)

	fw.restoreAlwaysOnTop()

	if called {
		t.Fatalf("restore should not apply when preference is false")
	}
	if fw.IsAlwaysOnTop() {
		t.Fatalf("always-on-top should be false")
	}
}

func TestSetLaunchAtStartup_NoController(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{}
	if fw.IsLaunchAtStartup() {
		t.Fatalf("zero value launch-at-startup should be false")
	}

	if ok := fw.SetLaunchAtStartup(true); ok {
		t.Fatalf("SetLaunchAtStartup should fail when startupCtl is nil")
	}
	if fw.IsLaunchAtStartup() {
		t.Fatalf("state should remain unchanged when SetLaunchAtStartup fails")
	}
}

func TestSetLaunchAtStartup_SuccessUpdatesState(t *testing.T) {
	t.Parallel()

	var calls []bool
	fw := &FloatingWindow{
		startupSet: func(enabled bool) bool {
			calls = append(calls, enabled)
			return true
		},
	}

	if ok := fw.SetLaunchAtStartup(true); !ok {
		t.Fatalf("SetLaunchAtStartup(true) should succeed")
	}
	if !fw.IsLaunchAtStartup() {
		t.Fatalf("launch-at-startup should be true")
	}
	if len(calls) != 1 || !calls[0] {
		t.Fatalf("startupSet calls = %v, want [true]", calls)
	}

	if ok := fw.SetLaunchAtStartup(false); !ok {
		t.Fatalf("SetLaunchAtStartup(false) should succeed")
	}
	if fw.IsLaunchAtStartup() {
		t.Fatalf("launch-at-startup should be false")
	}
	if len(calls) != 2 || calls[1] {
		t.Fatalf("startupSet second call should be false, calls=%v", calls)
	}
}

func TestSetLaunchAtStartup_FailureKeepsState(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{
		startupSet: func(bool) bool { return false },
	}
	fw.launchAtStartup.Store(true)

	if ok := fw.SetLaunchAtStartup(false); ok {
		t.Fatalf("SetLaunchAtStartup should fail")
	}
	if !fw.IsLaunchAtStartup() {
		t.Fatalf("state should remain true on failure")
	}
}

func TestRefreshLaunchAtStartup(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{
		startupGet: func() (bool, bool) {
			return true, true
		},
	}

	if ok := fw.RefreshLaunchAtStartup(); !ok {
		t.Fatalf("RefreshLaunchAtStartup should succeed")
	}
	if !fw.IsLaunchAtStartup() {
		t.Fatalf("launch-at-startup should be true")
	}
}

func TestRefreshLaunchAtStartup_FailureKeepsState(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{
		startupGet: func() (bool, bool) {
			return false, false
		},
	}
	fw.launchAtStartup.Store(true)

	if ok := fw.RefreshLaunchAtStartup(); ok {
		t.Fatalf("RefreshLaunchAtStartup should fail")
	}
	if !fw.IsLaunchAtStartup() {
		t.Fatalf("state should remain unchanged on refresh failure")
	}
}

func TestParseModeToggleHotkey(t *testing.T) {
	t.Parallel()

	parsed, ok := parseModeToggleHotkey(" ctrl + shift + f2 ")
	if !ok {
		t.Fatalf("parseModeToggleHotkey should succeed")
	}
	if parsed.label != "Ctrl+Shift+F2" {
		t.Fatalf("parsed label = %q, want Ctrl+Shift+F2", parsed.label)
	}
	if parsed.mod != (0x0002 | 0x0004) {
		t.Fatalf("parsed mod = %#x, want %#x", parsed.mod, 0x0002|0x0004)
	}
	if parsed.key != 0x71 {
		t.Fatalf("parsed key = %#x, want %#x", parsed.key, 0x71)
	}
}

func TestParseModeToggleHotkey_Invalid(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"M",
		"Ctrl+Alt",
		"Ctrl+Ctrl+M",
		"Ctrl+Alt+UnknownKey",
	}
	for _, c := range cases {
		if _, ok := parseModeToggleHotkey(c); ok {
			t.Fatalf("parseModeToggleHotkey(%q) should fail", c)
		}
	}
}

func TestModeToggleHotkeyFromKeyEvent(t *testing.T) {
	t.Parallel()

	got, ok := modeToggleHotkeyFromKeyEvent(fyne.KeyF2, fyne.KeyModifierControl|fyne.KeyModifierShift)
	if !ok {
		t.Fatalf("modeToggleHotkeyFromKeyEvent should succeed")
	}
	if got != "Ctrl+Shift+F2" {
		t.Fatalf("captured hotkey = %q, want Ctrl+Shift+F2", got)
	}
}

func TestModeToggleHotkeyFromKeyEvent_Invalid(t *testing.T) {
	t.Parallel()

	if _, ok := modeToggleHotkeyFromKeyEvent(fyne.KeyM, 0); ok {
		t.Fatalf("missing modifier should fail")
	}
	if _, ok := modeToggleHotkeyFromKeyEvent(fyne.KeyUnknown, fyne.KeyModifierControl); ok {
		t.Fatalf("unknown key should fail")
	}
	if _, ok := modeToggleHotkeyFromKeyEvent(fyne.KeyM, fyne.KeyModifierSuper); ok {
		t.Fatalf("super modifier should fail")
	}
}

func TestSetModeToggleHotkey(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	var gotMod uint32
	var gotKey uint32
	calls := 0
	fw := &FloatingWindow{
		App: a,
		hotkeySupported: func() bool {
			return true
		},
		hotkeyRegister: func(mod uint32, key uint32, _ func()) bool {
			calls++
			gotMod = mod
			gotKey = key
			return true
		},
	}
	fw.restoreModeToggleHotkey()

	if ok := fw.SetModeToggleHotkey("Ctrl+Shift+M"); !ok {
		t.Fatalf("SetModeToggleHotkey should succeed")
	}
	if calls != 1 {
		t.Fatalf("hotkeyRegister calls = %d, want 1", calls)
	}
	if gotMod != (0x0002|0x0004) || gotKey != 0x4D {
		t.Fatalf("register args = (%#x,%#x), want (%#x,%#x)", gotMod, gotKey, 0x0002|0x0004, 0x4D)
	}
	if got := fw.ModeToggleHotkey(); got != "Ctrl+Shift+M" {
		t.Fatalf("ModeToggleHotkey() = %q, want Ctrl+Shift+M", got)
	}
	if got := a.Preferences().String(modeToggleHotkeyKey); got != "Ctrl+Shift+M" {
		t.Fatalf("saved hotkey = %q, want Ctrl+Shift+M", got)
	}
}

func TestSetModeToggleHotkey_RegisterFailureRollsBack(t *testing.T) {
	t.Parallel()

	fw := &FloatingWindow{
		hotkeySupported: func() bool {
			return true
		},
		hotkeyRegister: func(mod uint32, key uint32, _ func()) bool {
			return mod == (0x0002|0x0001) && key == 0x4D
		},
	}
	fw.restoreModeToggleHotkey()

	if ok := fw.SetModeToggleHotkey("Ctrl+Shift+M"); ok {
		t.Fatalf("SetModeToggleHotkey should fail when register fails")
	}
	if got := fw.ModeToggleHotkey(); got != defaultModeToggleHotkey {
		t.Fatalf("ModeToggleHotkey() after rollback = %q, want %q", got, defaultModeToggleHotkey)
	}
}

func TestRestoreModeToggleHotkey_DefaultAndSaved(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	fw := &FloatingWindow{App: a}
	fw.restoreModeToggleHotkey()
	if got := fw.ModeToggleHotkey(); got != defaultModeToggleHotkey {
		t.Fatalf("default ModeToggleHotkey() = %q, want %q", got, defaultModeToggleHotkey)
	}

	a.Preferences().SetString(modeToggleHotkeyKey, "Alt+Shift+M")
	fw.restoreModeToggleHotkey()
	if got := fw.ModeToggleHotkey(); got != "Alt+Shift+M" {
		t.Fatalf("saved ModeToggleHotkey() = %q, want Alt+Shift+M", got)
	}

	a.Preferences().SetString(modeToggleHotkeyKey, "invalid")
	fw.restoreModeToggleHotkey()
	if got := fw.ModeToggleHotkey(); got != defaultModeToggleHotkey {
		t.Fatalf("invalid saved value should fallback, got %q", got)
	}
}

func TestRestoreWindowPlacement_NoSavedPos_CentersAndPersists(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	oldGet := getWindowPosition
	oldMove := moveWindowTo
	oldVisible := isWindowInVisibleBounds
	oldSize := windowSizeInPixels
	t.Cleanup(func() {
		getWindowPosition = oldGet
		moveWindowTo = oldMove
		isWindowInVisibleBounds = oldVisible
		windowSizeInPixels = oldSize
	})

	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		return fyne.NewPos(11, 22), true
	}
	moveWindowTo = func(fyne.Window, float32, float32) bool { return false }
	isWindowInVisibleBounds = func(fyne.Position, fyne.Size) bool { return false }
	windowSizeInPixels = func(fyne.Window) fyne.Size { return fyne.NewSize(1, 1) }

	fw := &FloatingWindow{App: a, Window: w}
	fw.restoreWindowPlacement()

	prefs := a.Preferences()
	if !prefs.Bool(windowPosSetKey) {
		t.Fatalf("position should be persisted")
	}
	if got := float32(prefs.Float(windowPosXKey)); got != 11 {
		t.Fatalf("saved X = %v, want 11", got)
	}
	if got := float32(prefs.Float(windowPosYKey)); got != 22 {
		t.Fatalf("saved Y = %v, want 22", got)
	}
}

func TestRestoreWindowPlacement_SavedVisibleAndMoveOK_ReturnsEarly(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	oldGet := getWindowPosition
	oldMove := moveWindowTo
	oldVisible := isWindowInVisibleBounds
	oldSize := windowSizeInPixels
	t.Cleanup(func() {
		getWindowPosition = oldGet
		moveWindowTo = oldMove
		isWindowInVisibleBounds = oldVisible
		windowSizeInPixels = oldSize
	})

	var movedX, movedY float32
	moveCalls := 0
	getCalls := 0
	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		getCalls++
		return fyne.NewPos(99, 99), true
	}
	moveWindowTo = func(_ fyne.Window, x, y float32) bool {
		moveCalls++
		movedX, movedY = x, y
		return true
	}
	isWindowInVisibleBounds = func(pos fyne.Position, size fyne.Size) bool {
		return pos.X == 7 && pos.Y == 9 && size.Width == 300 && size.Height == 200
	}
	windowSizeInPixels = func(fyne.Window) fyne.Size { return fyne.NewSize(300, 200) }

	prefs := a.Preferences()
	prefs.SetBool(windowPosSetKey, true)
	prefs.SetFloat(windowPosXKey, 7)
	prefs.SetFloat(windowPosYKey, 9)

	fw := &FloatingWindow{App: a, Window: w}
	fw.restoreWindowPlacement()

	if moveCalls != 1 {
		t.Fatalf("moveWindowTo calls = %d, want 1", moveCalls)
	}
	if movedX != 7 || movedY != 9 {
		t.Fatalf("moved to (%v, %v), want (7, 9)", movedX, movedY)
	}
	if getCalls != 0 {
		t.Fatalf("getWindowPosition should not be called on early return, got %d", getCalls)
	}
}

func TestRestoreWindowPlacement_MoveFails_FallsBackAndPersistsCentered(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	oldGet := getWindowPosition
	oldMove := moveWindowTo
	oldVisible := isWindowInVisibleBounds
	oldSize := windowSizeInPixels
	t.Cleanup(func() {
		getWindowPosition = oldGet
		moveWindowTo = oldMove
		isWindowInVisibleBounds = oldVisible
		windowSizeInPixels = oldSize
	})

	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		return fyne.NewPos(30, 40), true
	}
	moveWindowTo = func(_ fyne.Window, _, _ float32) bool { return false }
	isWindowInVisibleBounds = func(fyne.Position, fyne.Size) bool { return true }
	windowSizeInPixels = func(fyne.Window) fyne.Size { return fyne.NewSize(300, 200) }

	prefs := a.Preferences()
	prefs.SetBool(windowPosSetKey, true)
	prefs.SetFloat(windowPosXKey, 123)
	prefs.SetFloat(windowPosYKey, 456)

	fw := &FloatingWindow{App: a, Window: w}
	fw.restoreWindowPlacement()

	if got := float32(prefs.Float(windowPosXKey)); got != 30 {
		t.Fatalf("fallback saved X = %v, want 30", got)
	}
	if got := float32(prefs.Float(windowPosYKey)); got != 40 {
		t.Fatalf("fallback saved Y = %v, want 40", got)
	}
}
