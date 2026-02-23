package app

import (
	"image/color"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/haua/futu/app/drag"
	"github.com/haua/futu/app/platform"
	"github.com/haua/futu/app/player"
	"github.com/haua/futu/app/utils"
)

var (
	getWindowPosition       = platform.GetWindowPosition
	getCursorPosition       = platform.GetCursorPosition
	moveWindowTo            = platform.MoveWindowTo
	isWindowInVisibleBounds = platform.IsWindowInVisibleBounds
	windowSizeInPixels      = utils.WindowSizeInPixels
)

const (
	windowPosXKey      = "window.pos_x"
	windowPosYKey      = "window.pos_y"
	windowPosSetKey    = "window.pos_set"
	alwaysOnTopKey     = "window.always_on_top"
	mouseFarOpacityKey = "window.mouse_far_opacity"
	startupValueName   = "Futu"
	mouseFadeRange     = float32(200)
	mouseFadeTick      = 50 * time.Millisecond
	modeHintDuration   = 1200 * time.Millisecond
	imageTickInterval  = time.Hour
)

type FloatingWindow struct {
	App    fyne.App
	Window fyne.Window
	Player *player.Player
	// 是否处于编辑模式
	editMode             atomic.Bool
	alwaysOnTop          atomic.Bool
	topMostCtl           *utils.WindowTopMost
	topMostSet           func(enabled bool) bool
	taskbarCtl           *utils.WindowTaskbar
	taskbarSet           func(visible bool) bool
	mouseCtl             *utils.WindowMousePassthrough
	mouseSet             func(enabled bool) bool
	opacityCtl           *utils.WindowOpacity
	opacitySet           func(opacity float64) bool
	fadeLoopMu           sync.Mutex
	fadeLoopStop         chan struct{}
	fadeStateMu          sync.Mutex
	lastCursor           fyne.Position
	hasCursor            bool
	lastOpacity          uint8
	hasOpacity           bool
	mouseFarOpacity      uint8
	launchAtStartup      atomic.Bool
	startupCtl           *utils.LaunchAtStartup
	startupSet           func(enabled bool) bool
	startupGet           func() (bool, bool)
	hotkeyCtl            *utils.GlobalHotkey
	hotkeySupported      func() bool
	hotkeyRegister       func(mod uint32, key uint32, onTrigger func()) bool
	hotkeyUnregister     func()
	hideHotkeyCtl        *utils.GlobalHotkey
	hideHotkeySupported  func() bool
	hideHotkeyRegister   func(mod uint32, key uint32, onTrigger func()) bool
	hideHotkeyUnregister func()
	hotkeyMu             sync.Mutex
	modeHotkey           string
	hideWindowHotkey     string
	windowHidden         atomic.Bool
	hotkeyCapturing      atomic.Bool
	modeHintLabel        *widget.Label
	modeHintBox          fyne.CanvasObject
	modeHintMu           sync.Mutex
	modeHintTimer        *time.Timer
	imageSourceMu        sync.Mutex
	imageSourceMode      string
	fixedImagePath       string
	randomFolderPath     string
	lastRandomImagePath  string
	imageTickerStop      chan struct{}
	imageTickerInterval  time.Duration
	randomIntn           func(int) int
}

type modeHintTheme struct {
	base fyne.Theme
}

func (t modeHintTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameForeground {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return t.base.Color(name, variant)
}

func (t modeHintTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t modeHintTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t modeHintTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}

func NewFloatingWindow(a fyne.App) *FloatingWindow {
	var w fyne.Window
	if d, ok := a.Driver().(desktop.Driver); ok {
		w = d.CreateSplashWindow() // 无边框窗口
	} else {
		// 这个应用只在 desktop 使用，不会进这个分支的，但也写个兜底吧
		w = a.NewWindow("Futu")
	}

	title := a.Metadata().Name
	if title == "" {
		title = "Futu"
	}
	w.SetTitle(title)

	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(200, 200))
	w.SetPadded(false) // 去掉内容内边距
	w.SetMaster()

	player_instance := player.NewPlayer(a, w)

	fw := &FloatingWindow{
		App:                 a,
		Window:              w,
		Player:              player_instance,
		topMostCtl:          utils.NewWindowTopMost(w),
		taskbarCtl:          utils.NewWindowTaskbar(w),
		mouseCtl:            utils.NewWindowMousePassthrough(w),
		opacityCtl:          utils.NewWindowOpacity(w),
		startupCtl:          utils.NewLaunchAtStartup(startupValueName),
		hotkeyCtl:           utils.NewGlobalHotkey(),
		hideHotkeyCtl:       utils.NewGlobalHotkey(),
		imageTickerInterval: imageTickInterval,
		randomIntn:          rand.Intn,
	}
	fw.topMostSet = fw.topMostCtl.Set
	fw.taskbarSet = fw.taskbarCtl.SetVisible
	fw.mouseSet = fw.mouseCtl.SetEnabled
	fw.opacitySet = fw.opacityCtl.Set
	fw.startupSet = func(enabled bool) bool {
		if fw.startupCtl == nil {
			return false
		}
		return fw.startupCtl.SetEnabled(enabled) == nil
	}
	fw.startupGet = func() (bool, bool) {
		if fw.startupCtl == nil {
			return false, false
		}
		enabled, err := fw.startupCtl.IsEnabled()
		return enabled, err == nil
	}
	fw.hotkeySupported = fw.hotkeyCtl.Supported
	fw.hotkeyRegister = fw.hotkeyCtl.Register
	fw.hotkeyUnregister = fw.hotkeyCtl.Unregister
	fw.hideHotkeySupported = fw.hideHotkeyCtl.Supported
	fw.hideHotkeyRegister = fw.hideHotkeyCtl.Register
	fw.hideHotkeyUnregister = fw.hideHotkeyCtl.Unregister
	fw.RefreshLaunchAtStartup()
	fw.restoreModeToggleHotkey()
	fw.restoreHideWindowHotkey()
	fw.restoreImageSource()
	fw.editMode.Store(true)
	fw.mouseFarOpacity = opacityToAlpha(1)

	mainContent := drag.NewWidget(
		w,
		player_instance.Canvas,
		player_instance.SetRenderPaused,
		player_instance.AdjustScaleByScroll,
		fw.SaveWindowPosition,
		fw.IsEditMode,
	)
	fw.initModeHint()
	hintOverlay := container.NewVBox(
		layout.NewSpacer(),
		container.NewPadded(
			container.NewHBox(
				layout.NewSpacer(),
				fw.modeHintBox,
				layout.NewSpacer(),
			),
		),
	)
	w.SetContent(container.NewStack(mainContent, hintOverlay))

	return fw
}

func (f *FloatingWindow) Show() {
	// 播放上一次选的图片
	f.playImageOnStartup()
	f.Window.Show()
	f.restoreWindowPlacement()
	f.restoreAlwaysOnTop()
	f.restoreMouseFarOpacity()
	f.windowHidden.Store(false)
	f.applyModeToggleHotkey()
	f.applyHideWindowHotkey()
	if f.IsEditMode() {
		f.stopMouseFadeLoop()
		if f.Player != nil {
			f.Player.SetFullyTransparentPaused(false)
		}
		f.applyWindowOpacity(1.0)
		return
	}
	f.startMouseFadeLoop()
	f.updateWindowOpacityByCursor()
}

func (f *FloatingWindow) IsEditMode() bool {
	return f.editMode.Load()
}

func (f *FloatingWindow) ToggleEditMode() bool {
	f.EnsureWindowVisible()
	for {
		current := f.editMode.Load()
		next := !current
		if f.editMode.CompareAndSwap(current, next) {
			f.applyTaskbarVisibility(next)
			f.applyMousePassthrough(!next)
			if next {
				f.stopMouseFadeLoop()
				if f.Player != nil {
					f.Player.SetFullyTransparentPaused(false)
				}
				f.applyWindowOpacity(1.0)
			} else {
				f.startMouseFadeLoop()
				f.updateWindowOpacityByCursor()
			}
			f.showModeHint(next)
			return next
		}
	}
}

func (f *FloatingWindow) IsAlwaysOnTop() bool {
	return f.alwaysOnTop.Load()
}

func (f *FloatingWindow) SetAlwaysOnTop(enabled bool) bool {
	if !f.applyAlwaysOnTop(enabled) {
		return false
	}
	f.alwaysOnTop.Store(enabled)
	f.saveAlwaysOnTopPreference(enabled)
	return true
}

func (f *FloatingWindow) ToggleAlwaysOnTop() bool {
	next := !f.IsAlwaysOnTop()
	if f.SetAlwaysOnTop(next) {
		return next
	}
	return f.IsAlwaysOnTop()
}

func (f *FloatingWindow) IsLaunchAtStartup() bool {
	return f.launchAtStartup.Load()
}

func (f *FloatingWindow) SetLaunchAtStartup(enabled bool) bool {
	if !f.applyLaunchAtStartup(enabled) {
		return false
	}
	f.launchAtStartup.Store(enabled)
	return true
}

func (f *FloatingWindow) RefreshLaunchAtStartup() bool {
	enabled, ok := f.queryLaunchAtStartup()
	if !ok {
		return false
	}
	f.launchAtStartup.Store(enabled)
	return true
}

func (f *FloatingWindow) applyLaunchAtStartup(enabled bool) bool {
	if f.startupSet != nil {
		return f.startupSet(enabled)
	}
	if f.startupCtl == nil {
		return false
	}
	return f.startupCtl.SetEnabled(enabled) == nil
}

func (f *FloatingWindow) queryLaunchAtStartup() (bool, bool) {
	if f.startupGet != nil {
		return f.startupGet()
	}
	if f.startupCtl == nil {
		return false, false
	}
	enabled, err := f.startupCtl.IsEnabled()
	return enabled, err == nil
}

func (f *FloatingWindow) applyAlwaysOnTop(enabled bool) bool {
	if f.topMostSet != nil {
		return f.topMostSet(enabled)
	}
	if f.topMostCtl == nil {
		return false
	}
	return f.topMostCtl.Set(enabled)
}

func (f *FloatingWindow) applyTaskbarVisibility(visible bool) bool {
	if f.taskbarSet != nil {
		return f.taskbarSet(visible)
	}
	if f.taskbarCtl == nil {
		return false
	}
	return f.taskbarCtl.SetVisible(visible)
}

func (f *FloatingWindow) applyMousePassthrough(enabled bool) bool {
	if f.mouseSet != nil {
		return f.mouseSet(enabled)
	}
	if f.mouseCtl == nil {
		return false
	}
	return f.mouseCtl.SetEnabled(enabled)
}

func (f *FloatingWindow) applyWindowOpacity(opacity float64) bool {
	alpha := opacityToAlpha(opacity)

	f.fadeStateMu.Lock()
	if f.hasOpacity && f.lastOpacity == alpha {
		f.fadeStateMu.Unlock()
		return true
	}
	f.lastOpacity = alpha
	f.hasOpacity = true
	f.fadeStateMu.Unlock()
	if f.Player != nil {
		shouldPause := !f.IsEditMode() && alpha == 0
		f.Player.SetFullyTransparentPaused(shouldPause)
	}

	if f.opacitySet != nil {
		return f.opacitySet(opacity)
	}
	if f.opacityCtl == nil {
		return false
	}
	return f.opacityCtl.Set(opacity)
}

func (f *FloatingWindow) startMouseFadeLoop() {
	f.fadeLoopMu.Lock()
	if f.fadeLoopStop != nil {
		f.fadeLoopMu.Unlock()
		return
	}
	stop := make(chan struct{})
	f.fadeLoopStop = stop
	f.fadeLoopMu.Unlock()

	go func() {
		ticker := time.NewTicker(mouseFadeTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f.updateWindowOpacityByCursor()
			case <-stop:
				return
			}
		}
	}()
}

func (f *FloatingWindow) stopMouseFadeLoop() {
	f.fadeLoopMu.Lock()
	stop := f.fadeLoopStop
	if stop != nil {
		f.fadeLoopStop = nil
	}
	f.fadeLoopMu.Unlock()
	if stop != nil {
		close(stop)
	}
	f.resetFadeState()
}

func (f *FloatingWindow) updateWindowOpacityByCursor() {
	if f == nil || f.Window == nil {
		return
	}
	if f.IsEditMode() {
		f.applyWindowOpacity(1.0)
		return
	}

	cursorPos, ok := getCursorPosition()
	if !ok {
		return
	}

	f.fadeStateMu.Lock()
	sameCursor := f.hasCursor && cursorPos == f.lastCursor
	if !sameCursor {
		f.lastCursor = cursorPos
		f.hasCursor = true
	}
	f.fadeStateMu.Unlock()
	if sameCursor {
		return
	}

	winPos, ok := getWindowPosition(f.Window)
	if !ok {
		return
	}
	size := windowSizeInPixels(f.Window)
	distance := cursorDistanceToRect(cursorPos, winPos, size)
	f.applyWindowOpacity(opacityByCursorDistance(distance, f.MouseFarOpacity()))
}

func (f *FloatingWindow) resetFadeState() {
	f.fadeStateMu.Lock()
	f.hasCursor = false
	f.hasOpacity = false
	f.fadeStateMu.Unlock()
}

func opacityToAlpha(opacity float64) uint8 {
	opacity = utils.ClampFloat64(opacity, 0, 1)
	return uint8(math.Round(opacity * 255))
}

func opacityByCursorDistance(distance float32, maxOpacity float64) float64 {
	maxOpacity = utils.ClampFloat64(maxOpacity, 0, 1)
	if distance <= 0 {
		return 0
	}
	if distance >= mouseFadeRange {
		return maxOpacity
	}
	return float64(distance/mouseFadeRange) * maxOpacity
}

func (f *FloatingWindow) MouseFarOpacity() float64 {
	if f == nil {
		return 1
	}
	f.fadeStateMu.Lock()
	alpha := f.mouseFarOpacity
	f.fadeStateMu.Unlock()
	return float64(alpha) / 255
}

func (f *FloatingWindow) SetMouseFarOpacity(opacity float64) {
	if f == nil {
		return
	}

	alpha := opacityToAlpha(utils.ClampFloat64(opacity, 0, 1))
	f.fadeStateMu.Lock()
	f.mouseFarOpacity = alpha
	// Force recompute even when cursor did not move.
	f.hasCursor = false
	f.fadeStateMu.Unlock()

	f.saveMouseFarOpacity()
	if !f.IsEditMode() {
		f.updateWindowOpacityByCursor()
	}
}

func (f *FloatingWindow) saveMouseFarOpacity() {
	if f == nil || f.App == nil {
		return
	}
	f.App.Preferences().SetFloat(mouseFarOpacityKey, f.MouseFarOpacity())
}

func (f *FloatingWindow) restoreMouseFarOpacity() {
	if f == nil {
		return
	}
	if f.App == nil {
		f.SetMouseFarOpacity(1)
		return
	}
	prefs := f.App.Preferences()
	value := prefs.Float(mouseFarOpacityKey)
	if value <= 0 {
		f.SetMouseFarOpacity(1)
		return
	}
	f.SetMouseFarOpacity(value)
}

func cursorDistanceToRect(cursor, winPos fyne.Position, size fyne.Size) float32 {
	left := winPos.X
	top := winPos.Y
	right := winPos.X + size.Width
	bottom := winPos.Y + size.Height

	var dx float32
	if cursor.X < left {
		dx = left - cursor.X
	} else if cursor.X > right {
		dx = cursor.X - right
	}

	var dy float32
	if cursor.Y < top {
		dy = top - cursor.Y
	} else if cursor.Y > bottom {
		dy = cursor.Y - bottom
	}

	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
	}
	return float32(math.Hypot(float64(dx), float64(dy)))
}

func (f *FloatingWindow) saveAlwaysOnTopPreference(enabled bool) {
	if f.App == nil {
		return
	}
	f.App.Preferences().SetBool(alwaysOnTopKey, enabled)
}

func (f *FloatingWindow) restoreAlwaysOnTop() {
	if f.App == nil {
		return
	}
	if !f.App.Preferences().Bool(alwaysOnTopKey) {
		f.alwaysOnTop.Store(false)
		return
	}
	if f.SetAlwaysOnTop(true) {
		return
	}
	f.alwaysOnTop.Store(false)
}

func (f *FloatingWindow) SaveWindowPosition(pos fyne.Position) {
	prefs := f.App.Preferences()
	prefs.SetFloat(windowPosXKey, float64(pos.X))
	prefs.SetFloat(windowPosYKey, float64(pos.Y))
	prefs.SetBool(windowPosSetKey, true)
}

func (f *FloatingWindow) restoreWindowPlacement() {
	prefs := f.App.Preferences()
	if !prefs.Bool(windowPosSetKey) {
		f.Window.CenterOnScreen()
		if pos, ok := getWindowPosition(f.Window); ok {
			f.SaveWindowPosition(pos)
		}
		return
	}

	pos := fyne.NewPos(
		float32(prefs.Float(windowPosXKey)),
		float32(prefs.Float(windowPosYKey)),
	)
	size := windowSizeInPixels(f.Window)
	if isWindowInVisibleBounds(pos, size) && moveWindowTo(f.Window, pos.X, pos.Y) {
		return
	}

	f.Window.CenterOnScreen()
	if centeredPos, ok := getWindowPosition(f.Window); ok {
		f.SaveWindowPosition(centeredPos)
	}
}

func (f *FloatingWindow) initModeHint() {
	text := widget.NewLabel("")
	text.Alignment = fyne.TextAlignCenter
	text.TextStyle = fyne.TextStyle{}

	baseTheme := theme.DefaultTheme()
	if f != nil && f.App != nil && f.App.Settings() != nil && f.App.Settings().Theme() != nil {
		baseTheme = f.App.Settings().Theme()
	}
	textObj := container.NewThemeOverride(text, modeHintTheme{base: baseTheme})

	bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 170})
	bubble := container.NewStack(
		bg,
		container.NewPadded(
			container.NewCenter(textObj),
		),
	)
	box := container.NewPadded(bubble)
	box.Hide()

	f.modeHintLabel = text
	f.modeHintBox = box
}

func modeHintText(isEdit bool) string {
	if isEdit {
		return "编辑模式"
	}
	return "常态模式"
}

func (f *FloatingWindow) showModeHint(isEdit bool) {
	if f == nil || f.modeHintLabel == nil || f.modeHintBox == nil {
		return
	}

	f.modeHintMu.Lock()
	if f.modeHintTimer != nil {
		f.modeHintTimer.Stop()
		f.modeHintTimer = nil
	}
	hintText := modeHintText(isEdit)
	fyne.Do(func() {
		f.modeHintLabel.SetText(hintText)
		f.modeHintBox.Show()
		f.modeHintBox.Refresh()
	})

	timer := time.AfterFunc(modeHintDuration, func() {
		f.hideModeHint()
	})
	f.modeHintTimer = timer
	f.modeHintMu.Unlock()
}

func (f *FloatingWindow) hideModeHint() {
	if f == nil || f.modeHintBox == nil {
		return
	}

	f.modeHintMu.Lock()
	f.modeHintTimer = nil
	f.modeHintMu.Unlock()

	fyne.Do(func() {
		f.modeHintBox.Hide()
		if f.Window != nil && f.Window.Canvas() != nil {
			f.Window.Canvas().Refresh(f.modeHintBox)
		}
	})
}
