package player

import (
	"image"
	"image/color"
	"image/gif"
	"testing"
	"time"
)

func TestComposeGIFFrames_DisposalBackground(t *testing.T) {
	t.Parallel()

	palette := color.Palette{
		color.RGBA{0, 0, 0, 0},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	frame0 := image.NewPaletted(image.Rect(0, 0, 2, 1), palette)
	frame0.SetColorIndex(0, 0, 1)
	frame0.SetColorIndex(1, 0, 1)

	frame1 := image.NewPaletted(image.Rect(0, 0, 1, 1), palette)
	frame1.SetColorIndex(0, 0, 2)

	g := &gif.GIF{
		Image:    []*image.Paletted{frame0, frame1},
		Disposal: []byte{gif.DisposalBackground, gif.DisposalNone},
		Config: image.Config{
			Width:  2,
			Height: 1,
		},
	}

	frames := composeGIFFrames(g)
	if len(frames) != 2 {
		t.Fatalf("len(frames) = %d, want 2", len(frames))
	}

	assertRGBA(t, frames[0].At(0, 0), color.RGBA{255, 0, 0, 255})
	assertRGBA(t, frames[0].At(1, 0), color.RGBA{255, 0, 0, 255})
	assertRGBA(t, frames[1].At(0, 0), color.RGBA{0, 0, 255, 255})
	assertRGBA(t, frames[1].At(1, 0), color.RGBA{0, 0, 0, 0})
}

func TestComposeGIFFrames_DisposalPrevious(t *testing.T) {
	t.Parallel()

	palette := color.Palette{
		color.RGBA{0, 0, 0, 0},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 0, 255, 255},
		color.RGBA{0, 255, 0, 255},
	}

	frame0 := image.NewPaletted(image.Rect(0, 0, 2, 1), palette)
	frame0.SetColorIndex(0, 0, 1)
	frame0.SetColorIndex(1, 0, 1)

	frame1 := image.NewPaletted(image.Rect(0, 0, 1, 1), palette)
	frame1.SetColorIndex(0, 0, 2)

	frame2 := image.NewPaletted(image.Rect(1, 0, 2, 1), palette)
	frame2.SetColorIndex(1, 0, 3)

	g := &gif.GIF{
		Image:    []*image.Paletted{frame0, frame1, frame2},
		Disposal: []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalNone},
		Config: image.Config{
			Width:  2,
			Height: 1,
		},
	}

	frames := composeGIFFrames(g)
	if len(frames) != 3 {
		t.Fatalf("len(frames) = %d, want 3", len(frames))
	}

	assertRGBA(t, frames[1].At(0, 0), color.RGBA{0, 0, 255, 255})
	assertRGBA(t, frames[1].At(1, 0), color.RGBA{255, 0, 0, 255})

	assertRGBA(t, frames[2].At(0, 0), color.RGBA{255, 0, 0, 255})
	assertRGBA(t, frames[2].At(1, 0), color.RGBA{0, 255, 0, 255})
}

func TestDisposalAt_OutOfRange(t *testing.T) {
	t.Parallel()

	g := &gif.GIF{Disposal: []byte{gif.DisposalBackground}}
	if got := disposalAt(g, 10); got != gif.DisposalNone {
		t.Fatalf("disposalAt out-of-range = %v, want %v", got, gif.DisposalNone)
	}
}

func TestComposeGIFFrames_NilOrEmpty(t *testing.T) {
	t.Parallel()

	if got := composeGIFFrames(nil); got != nil {
		t.Fatalf("composeGIFFrames(nil) should return nil")
	}
	if got := composeGIFFrames(&gif.GIF{}); got != nil {
		t.Fatalf("composeGIFFrames(empty) should return nil")
	}
}

func TestCloneRGBA_DeepCopy(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})

	dst := cloneRGBA(src, src.Bounds())
	src.SetRGBA(0, 0, color.RGBA{0, 0, 255, 255})

	assertRGBA(t, dst.At(0, 0), color.RGBA{255, 0, 0, 255})
}

func TestNormalizedGIFFrameDelay(t *testing.T) {
	t.Parallel()

	g := &gif.GIF{Delay: []int{0, 1, 3}}
	if got := normalizedGIFFrameDelay(g, 0); got != minGIFFrameDelay {
		t.Fatalf("delay[0] = %v, want %v", got, minGIFFrameDelay)
	}
	if got := normalizedGIFFrameDelay(g, 1); got != minGIFFrameDelay {
		t.Fatalf("delay[1] = %v, want %v", got, minGIFFrameDelay)
	}
	if got := normalizedGIFFrameDelay(g, 2); got != 30*time.Millisecond {
		t.Fatalf("delay[2] = %v, want 30ms", got)
	}
	if got := normalizedGIFFrameDelay(g, 99); got != minGIFFrameDelay {
		t.Fatalf("delay[out-of-range] = %v, want %v", got, minGIFFrameDelay)
	}
}

func assertRGBA(t *testing.T, got color.Color, want color.RGBA) {
	t.Helper()
	r, g, b, a := got.RGBA()
	gotRGBA := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
	if gotRGBA != want {
		t.Fatalf("got RGBA %+v, want %+v", gotRGBA, want)
	}
}
