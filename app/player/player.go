package player

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const lastImagePathKey = "player.last_image_path"

type Player struct {
	app          fyne.App
	Canvas       *canvas.Image
	window       fyne.Window
	renderPaused atomic.Bool
}

func NewPlayer(a fyne.App, w fyne.Window) *Player {
	img := canvas.NewImageFromImage(nil)
	img.Resize(fyne.NewSize(200, 200))
	img.FillMode = canvas.ImageFillContain

	return &Player{
		app:    a,
		Canvas: img,
		window: w,
	}
}

func (p *Player) Play(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		log.Printf("image file not found: %q (%v)", path, err)
		return
	}

	lower := strings.ToLower(path)
	// 记录本次播放的图，下次打开app自动用
	p.app.Preferences().SetString(lastImagePathKey, path)
	if strings.HasSuffix(lower, ".gif") {
		PlayGIF(p, path)
	} else if strings.HasSuffix(lower, ".webp") {
		PlayWebP(p, path)
	} else if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		PlayImage(p, path)
	}
}

func (p *Player) PlayLast() {
	path := strings.TrimSpace(p.app.Preferences().String(lastImagePathKey))
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		p.app.Preferences().SetString(lastImagePathKey, "")
		return
	}
	p.Play(path)
}

func (p *Player) SetRenderPaused(paused bool) {
	p.renderPaused.Store(paused)
}

func (p *Player) RenderPaused() bool {
	return p.renderPaused.Load()
}

func PlayImage(p *Player, path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("open image failed: %v", err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("decode image failed: %v", err)
		return
	}

	fyne.Do(func() {
		p.Canvas.Image = img
		p.Canvas.Refresh()
	})
}
