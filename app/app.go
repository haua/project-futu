package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/haua/futu/app/player"
)

type FloatingWindow struct {
	Window fyne.Window
	Player *player.Player
}

func NewFloatingWindow(a fyne.App) *FloatingWindow {
	w := a.NewWindow("Futu")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(200, 200))
	w.SetMaster()

	p := player.NewPlayer(w)

	w.SetContent(container.NewWithoutLayout(p.Canvas))
	w.CenterOnScreen()
	w.SetPadded(false)

	return &FloatingWindow{
		Window: w,
		Player: p,
	}
}

func (f *FloatingWindow) Show() {
	f.Player.PlayLast()
	f.Window.Show()
}
