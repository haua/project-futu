package player

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const lastImagePathKey = "player.last_image_path"

const (
	defaultImageSize = 200
	minZoom          = 0.2
	maxZoom          = 5.0
	zoomStep         = 0.1
)

type Player struct {
	app          fyne.App
	Canvas       *canvas.Image
	window       fyne.Window
	renderPaused atomic.Bool
	baseSize     fyne.Size
	zoom         float32
}

func NewPlayer(a fyne.App, w fyne.Window) *Player {
	img := canvas.NewImageFromImage(nil)
	img.Resize(fyne.NewSize(defaultImageSize, defaultImageSize))
	img.FillMode = canvas.ImageFillContain

	return &Player{
		app:    a,
		Canvas: img,
		window: w,
		baseSize: fyne.NewSize(
			defaultImageSize,
			defaultImageSize,
		),
		zoom: 1.0,
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

// 鼠标滚轮滚动时会调这个函数，
// 是由widget里面的Scrolled调的。
func (p *Player) AdjustScaleByScroll(ev *fyne.ScrollEvent) {
	if ev == nil {
		return
	}

	delta := ev.Scrolled.DY
	if delta == 0 {
		delta = ev.Scrolled.DX
	}
	if delta == 0 {
		return
	}

	if delta > 0 {
		p.adjustScaleAt(zoomStep, ev.AbsolutePosition)
		return
	}
	p.adjustScaleAt(-zoomStep, ev.AbsolutePosition)
}

func (p *Player) adjustScale(delta float32) {
	p.adjustScaleAt(delta, fyne.Position{})
}

func (p *Player) adjustScaleAt(delta float32, anchor fyne.Position) {
	target := clampFloat32(p.zoom+delta, minZoom, maxZoom)
	if target == p.zoom {
		return
	}

	oldZoom := p.zoom
	oldSize := p.scaledSizeForZoom(oldZoom)
	scale := float32(1.0)
	if c := p.window.Canvas(); c != nil && c.Scale() > 0 {
		scale = c.Scale()
	}

	anchor = clampPosition(anchor, oldSize)
	oldAnchorPX := toScreenPixels(anchor.X, scale)
	oldAnchorPY := toScreenPixels(anchor.Y, scale)

	zoomRatio := target / oldZoom
	newAnchorPX := toScreenPixels(anchor.X*zoomRatio, scale)
	newAnchorPY := toScreenPixels(anchor.Y*zoomRatio, scale)

	winPos, canMove := getWindowPosition(p.window)
	nextWinPos := fyne.NewPos(
		winPos.X+float32(oldAnchorPX-newAnchorPX),
		winPos.Y+float32(oldAnchorPY-newAnchorPY),
	)

	p.zoom = target
	p.applyScaledSize()
	if canMove {
		moveWindowTo(p.window, nextWinPos.X, nextWinPos.Y)
	}
}

func (p *Player) updateBaseSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	p.baseSize = fyne.NewSize(float32(width), float32(height))
	p.zoom = 1.0
	p.applyScaledSize()
}

func (p *Player) applyScaledSize() {
	size := p.scaledSizeForZoom(p.zoom)
	p.Canvas.Resize(size)
	p.window.Resize(size)
}

func (p *Player) scaledSizeForZoom(zoom float32) fyne.Size {
	width := float32(math.Round(float64(p.baseSize.Width * zoom)))
	height := float32(math.Round(float64(p.baseSize.Height * zoom)))
	return fyne.NewSize(maxFloat32(width, 1), maxFloat32(height, 1))
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clampFloat32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampPosition(pos fyne.Position, bounds fyne.Size) fyne.Position {
	x := pos.X
	y := pos.Y

	if x < 0 {
		x = 0
	} else if x > bounds.Width {
		x = bounds.Width
	}
	if y < 0 {
		y = 0
	} else if y > bounds.Height {
		y = bounds.Height
	}

	return fyne.NewPos(x, y)
}

func toScreenPixels(v, scale float32) int {
	return int(math.Round(float64(v * scale)))
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
		b := img.Bounds()
		p.updateBaseSize(b.Dx(), b.Dy())
		p.Canvas.Image = img
		p.Canvas.Refresh()
	})
}
