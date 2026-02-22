package player

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
)

const minGIFFrameDelay = time.Second / 60

func PlayGIF(p *Player, path string, playbackID uint64) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("open gif failed: %v", err)
		return
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		log.Printf("decode gif failed: %v", err)
		return
	}

	frames := composeGIFFrames(g)
	if len(frames) == 0 {
		return
	}
	if !p.isPlaybackActive(playbackID) {
		return
	}

	fyne.Do(func() {
		if !p.isPlaybackActive(playbackID) {
			return
		}
		p.updateBaseSize(g.Config.Width, g.Config.Height)
	})

	go func(currentID uint64) {
		for p.isPlaybackActive(currentID) {
			for i := range frames {
				if !p.isPlaybackActive(currentID) {
					return
				}
				if !p.waitRenderResumed(currentID) {
					return
				}

				frame := frames[i]
				delay := normalizedGIFFrameDelay(g, i)

				fyne.Do(func() {
					if !p.isPlaybackActive(currentID) {
						return
					}
					p.Canvas.Image = frame
					p.Canvas.Refresh()
				})
				time.Sleep(delay)
			}
		}
	}(playbackID)
}

func normalizedGIFFrameDelay(g *gif.GIF, frameIndex int) time.Duration {
	delay := 10 * time.Millisecond
	if g != nil && frameIndex >= 0 && frameIndex < len(g.Delay) && g.Delay[frameIndex] > 0 {
		delay = time.Duration(g.Delay[frameIndex]) * 10 * time.Millisecond
	}
	if delay < minGIFFrameDelay {
		return minGIFFrameDelay
	}
	return delay
}

func composeGIFFrames(g *gif.GIF) []image.Image {
	if g == nil || len(g.Image) == 0 {
		return nil
	}

	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)
	transparent := image.NewUniform(color.Transparent)

	var restore *image.RGBA
	frames := make([]image.Image, 0, len(g.Image))

	for i, src := range g.Image {
		if i > 0 {
			prev := g.Image[i-1]
			switch disposalAt(g, i-1) {
			case gif.DisposalBackground:
				draw.Draw(canvas, prev.Bounds(), transparent, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				if restore != nil {
					draw.Draw(canvas, bounds, restore, bounds.Min, draw.Src)
				}
			}
		}

		if disposalAt(g, i) == gif.DisposalPrevious {
			restore = cloneRGBA(canvas, bounds)
		} else {
			restore = nil
		}

		draw.Draw(canvas, src.Bounds(), src, src.Bounds().Min, draw.Over)
		frames = append(frames, cloneRGBA(canvas, bounds))
	}

	return frames
}

func disposalAt(g *gif.GIF, i int) byte {
	if i >= 0 && i < len(g.Disposal) {
		return g.Disposal[i]
	}
	return gif.DisposalNone
}

func cloneRGBA(src *image.RGBA, bounds image.Rectangle) *image.RGBA {
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}
