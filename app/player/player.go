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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/haua/futu/app/platform"
)

const lastImagePathKey = "player.last_image_path"
const lastCanvasWidthKey = "player.last_canvas_width"

const (
	defaultImageSize = 200
	minZoom          = 0.2
	maxZoom          = 2.0
	zoomStep         = 0.1
)

const (
	pauseReasonDrag uint32 = 1 << iota
	pauseReasonFullyTransparent
)

type Player struct {
	app          fyne.App
	Canvas       *canvas.Image
	window       fyne.Window
	pauseReasons atomic.Uint32
	playbackID   atomic.Uint64
	pauseSignal  chan struct{}
	baseSize     fyne.Size
	zoom         float32
}

func NewPlayer(a fyne.App, w fyne.Window) *Player {
	img := canvas.NewImageFromImage(nil)
	img.Resize(fyne.NewSize(defaultImageSize, defaultImageSize))
	img.FillMode = canvas.ImageFillContain

	savedWidth := float32(a.Preferences().Float(lastCanvasWidthKey))
	initialZoom := float32(1.0)
	if savedWidth > 0 {
		initialZoom = clampFloat32(savedWidth/defaultImageSize, minZoom, maxZoom)
	}

	p := &Player{
		app:    a,
		Canvas: img,
		window: w,
		baseSize: fyne.NewSize(
			defaultImageSize,
			defaultImageSize,
		),
		zoom:        initialZoom,
		pauseSignal: make(chan struct{}, 1),
	}
	p.applyScaledSize()
	return p
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
	playbackID := p.beginPlayback()
	if strings.HasSuffix(lower, ".gif") {
		PlayGIF(p, path, playbackID)
	} else if strings.HasSuffix(lower, ".webp") {
		PlayWebP(p, path, playbackID)
	} else if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		PlayImage(p, path, playbackID)
	}

	// 记录本次播放的图，下次打开app自动用
	p.app.Preferences().SetString(lastImagePathKey, path)
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
	p.setPauseReason(pauseReasonDrag, paused)
}

func (p *Player) SetFullyTransparentPaused(paused bool) {
	p.setPauseReason(pauseReasonFullyTransparent, paused)
}

func (p *Player) RenderPaused() bool {
	return p.pauseReasons.Load() != 0
}

func (p *Player) waitRenderResumed(playbackID uint64) bool {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for p.RenderPaused() {
		if !p.isPlaybackActive(playbackID) {
			return false
		}
		select {
		case <-p.pauseSignal:
		case <-ticker.C:
		}
	}
	return p.isPlaybackActive(playbackID)
}

func (p *Player) setPauseReason(reason uint32, paused bool) {
	for {
		current := p.pauseReasons.Load()
		next := current
		if paused {
			next |= reason
		} else {
			next &^= reason
		}
		if current == next {
			return
		}
		if p.pauseReasons.CompareAndSwap(current, next) {
			if current != 0 && next == 0 {
				p.notifyRenderResumed()
			}
			return
		}
	}
}

func (p *Player) notifyRenderResumed() {
	if p.pauseSignal == nil {
		return
	}
	select {
	case p.pauseSignal <- struct{}{}:
	default:
	}
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

	winPos, canMove := platform.GetWindowPosition(p.window)
	nextWinPos := fyne.NewPos(
		winPos.X+float32(oldAnchorPX-newAnchorPX),
		winPos.Y+float32(oldAnchorPY-newAnchorPY),
	)

	p.zoom = target
	p.applyScaledSize()
	p.app.Preferences().SetFloat(lastCanvasWidthKey, float64(p.Canvas.Size().Width))
	if canMove {
		platform.MoveWindowTo(p.window, nextWinPos.X, nextWinPos.Y)
	}
}

func (p *Player) updateBaseSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	oldSize := p.Canvas.Size()
	oldPos, canMove := platform.GetWindowPosition(p.window)

	targetWidth := float32(p.app.Preferences().Float(lastCanvasWidthKey))
	if targetWidth <= 0 {
		targetWidth = oldSize.Width
	}
	if targetWidth <= 0 {
		targetWidth = p.baseSize.Width
	}

	p.baseSize = fyne.NewSize(float32(width), float32(height))
	if p.baseSize.Width > 0 && targetWidth > 0 {
		p.zoom = targetWidth / p.baseSize.Width
	}
	p.zoom = clampFloat32(p.zoom, minZoom, maxZoom)

	newSize := p.scaledSizeForZoom(p.zoom)
	p.applyScaledSize()
	p.app.Preferences().SetFloat(lastCanvasWidthKey, float64(newSize.Width))
	if canMove {
		nextX := oldPos.X + (oldSize.Width-newSize.Width)/2
		nextY := oldPos.Y + (oldSize.Height-newSize.Height)/2
		platform.MoveWindowTo(p.window, nextX, nextY)
	}
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

func (p *Player) beginPlayback() uint64 {
	return p.playbackID.Add(1)
}

func (p *Player) isPlaybackActive(id uint64) bool {
	return p.playbackID.Load() == id
}

func PlayImage(p *Player, path string, playbackID uint64) {
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
	if !p.isPlaybackActive(playbackID) {
		return
	}

	fyne.Do(func() {
		if !p.isPlaybackActive(playbackID) {
			return
		}
		b := img.Bounds()
		p.updateBaseSize(b.Dx(), b.Dy())
		p.Canvas.Image = img
		p.Canvas.Refresh()
	})
}
