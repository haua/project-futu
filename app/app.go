package app

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
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
	windowPosXKey   = "window.pos_x"
	windowPosYKey   = "window.pos_y"
	windowPosSetKey = "window.pos_set"
	alwaysOnTopKey  = "window.always_on_top"
	mouseFadeRange  = float32(200)
	mouseFadeTick   = 50 * time.Millisecond
)

type FloatingWindow struct {
	App    fyne.App
	Window fyne.Window
	Player *player.Player
	// 是否处于编辑模式
	editMode     atomic.Bool
	alwaysOnTop  atomic.Bool
	topMostCtl   *utils.WindowTopMost
	topMostSet   func(enabled bool) bool
	taskbarCtl   *utils.WindowTaskbar
	taskbarSet   func(visible bool) bool
	mouseCtl     *utils.WindowMousePassthrough
	mouseSet     func(enabled bool) bool
	opacityCtl   *utils.WindowOpacity
	opacitySet   func(opacity float64) bool
	fadeLoopMu   sync.Mutex
	fadeLoopStop chan struct{}
	fadeStateMu  sync.Mutex
	lastCursor   fyne.Position
	hasCursor    bool
	lastOpacity  uint8
	hasOpacity   bool
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
		App:        a,
		Window:     w,
		Player:     player_instance,
		topMostCtl: utils.NewWindowTopMost(w),
		taskbarCtl: utils.NewWindowTaskbar(w),
		mouseCtl:   utils.NewWindowMousePassthrough(w),
		opacityCtl: utils.NewWindowOpacity(w),
	}
	fw.topMostSet = fw.topMostCtl.Set
	fw.taskbarSet = fw.taskbarCtl.SetVisible
	fw.mouseSet = fw.mouseCtl.SetEnabled
	fw.opacitySet = fw.opacityCtl.Set
	fw.editMode.Store(true)

	w.SetContent(drag.NewWidget(
		w,
		player_instance.Canvas,
		player_instance.SetRenderPaused,
		player_instance.AdjustScaleByScroll,
		fw.SaveWindowPosition,
		fw.IsEditMode,
	))

	return fw
}

func (f *FloatingWindow) Show() {
	// 播放上一次选的图片
	f.Player.PlayLast()
	f.Window.Show()
	f.restoreWindowPlacement()
	f.restoreAlwaysOnTop()
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
	f.applyWindowOpacity(opacityByCursorDistance(distance))
}

func (f *FloatingWindow) resetFadeState() {
	f.fadeStateMu.Lock()
	f.hasCursor = false
	f.hasOpacity = false
	f.fadeStateMu.Unlock()
}

func opacityToAlpha(opacity float64) uint8 {
	if opacity < 0 {
		opacity = 0
	} else if opacity > 1 {
		opacity = 1
	}
	return uint8(math.Round(opacity * 255))
}

func opacityByCursorDistance(distance float32) float64 {
	if distance <= 0 {
		return 0
	}
	if distance >= mouseFadeRange {
		return 1
	}
	return float64(distance / mouseFadeRange)
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
