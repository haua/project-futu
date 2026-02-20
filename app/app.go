package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/haua/futu/app/drag"
	"github.com/haua/futu/app/player"
)

type FloatingWindow struct {
	Window fyne.Window
	Player *player.Player
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

	p := player.NewPlayer(a, w)

	w.SetContent(drag.NewWidget(w, p.Canvas, p.SetRenderPaused))
	w.CenterOnScreen()

	return &FloatingWindow{
		Window: w,
		Player: p,
	}
}

func (f *FloatingWindow) Show() {
	// 播放上一次选的图片
	f.Player.PlayLast()
	f.Window.Show()
}
