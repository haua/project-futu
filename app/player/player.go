package player

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type Player struct {
	Canvas *canvas.Image
	window fyne.Window
}

func NewPlayer(w fyne.Window) *Player {
	img := canvas.NewImageFromImage(nil)
	img.Resize(fyne.NewSize(200, 200))
	img.FillMode = canvas.ImageFillContain

	return &Player{
		Canvas: img,
		window: w,
	}
}

func (p *Player) Play(path string) {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".gif") {
		PlayGIF(p, path)
	} else if strings.HasSuffix(lower, ".webp") {
		PlayWebP(p, path)
	}
}

func (p *Player) PlayLast() {
	// 可扩展：从 config 中读取上次路径
}
