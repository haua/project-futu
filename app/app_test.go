package app

import (
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
