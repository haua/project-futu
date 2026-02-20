package app

import (
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/haua/futu/app/drag"
	"github.com/haua/futu/app/player"
	"github.com/haua/futu/app/utils"
)

type FloatingWindow struct {
	Window fyne.Window
	Player *player.Player
	// 是否处于编辑模式
	editMode    atomic.Bool
	alwaysOnTop atomic.Bool
	topMostCtl  *utils.WindowTopMost
}

func NewFloatingWindow(a fyne.App) *FloatingWindow {
	var w fyne.Window
	if d, ok := a.Driver().(desktop.Driver); ok {
		w = d.CreateSplashWindow() // 无边框窗口
	} else {
		// 这个应用只在 desktop 使用，不会进这个分支的，但也写个兜底吧
		w = a.NewWindow("Futu")
	}

	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(200, 200))
	w.SetPadded(false) // 去掉内容内边距
	w.SetMaster()

	player_instance := player.NewPlayer(a, w)

	fw := &FloatingWindow{
		Window:     w,
		Player:     player_instance,
		topMostCtl: utils.NewWindowTopMost(w),
	}
	fw.editMode.Store(true)

	w.SetContent(drag.NewWidget(
		w,
		player_instance.Canvas,
		player_instance.SetRenderPaused,
		player_instance.AdjustScaleByScroll,
		fw.IsEditMode,
	))
	w.CenterOnScreen()

	return fw
}

func (f *FloatingWindow) Show() {
	// 播放上一次选的图片
	f.Player.PlayLast()
	f.Window.Show()
}

func (f *FloatingWindow) IsEditMode() bool {
	return f.editMode.Load()
}

func (f *FloatingWindow) ToggleEditMode() bool {
	for {
		current := f.editMode.Load()
		next := !current
		if f.editMode.CompareAndSwap(current, next) {
			return next
		}
	}
}

func (f *FloatingWindow) IsAlwaysOnTop() bool {
	return f.alwaysOnTop.Load()
}

func (f *FloatingWindow) SetAlwaysOnTop(enabled bool) bool {
	if f.topMostCtl == nil || !f.topMostCtl.Set(enabled) {
		return false
	}
	f.alwaysOnTop.Store(enabled)
	return true
}

func (f *FloatingWindow) ToggleAlwaysOnTop() bool {
	next := !f.IsAlwaysOnTop()
	if f.SetAlwaysOnTop(next) {
		return next
	}
	return f.IsAlwaysOnTop()
}
