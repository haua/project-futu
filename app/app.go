package app

import (
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/haua/futu/app/drag"
	"github.com/haua/futu/app/platform"
	"github.com/haua/futu/app/player"
	"github.com/haua/futu/app/utils"
)

var (
	getWindowPosition       = platform.GetWindowPosition
	moveWindowTo            = platform.MoveWindowTo
	isWindowInVisibleBounds = platform.IsWindowInVisibleBounds
	windowSizeInPixels      = utils.WindowSizeInPixels
)

const (
	windowPosXKey   = "window.pos_x"
	windowPosYKey   = "window.pos_y"
	windowPosSetKey = "window.pos_set"
	alwaysOnTopKey  = "window.always_on_top"
)

type FloatingWindow struct {
	App    fyne.App
	Window fyne.Window
	Player *player.Player
	// 是否处于编辑模式
	editMode    atomic.Bool
	alwaysOnTop atomic.Bool
	topMostCtl  *utils.WindowTopMost
	topMostSet  func(enabled bool) bool
	taskbarCtl  *utils.WindowTaskbar
	taskbarSet  func(visible bool) bool
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
	}
	fw.topMostSet = fw.topMostCtl.Set
	fw.taskbarSet = fw.taskbarCtl.SetVisible
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
