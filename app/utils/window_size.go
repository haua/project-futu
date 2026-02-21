package utils

import (
	"math"

	"fyne.io/fyne/v2"
)

func WindowSizeInPixels(w fyne.Window) fyne.Size {
	size := fyne.NewSize(200, 200)
	if c := w.Canvas(); c != nil && c.Size().Width > 0 && c.Size().Height > 0 {
		size = c.Size()
	}
	scale := float32(1.0)
	if c := w.Canvas(); c != nil && c.Scale() > 0 {
		scale = c.Scale()
	}
	return fyne.NewSize(
		float32(math.Round(float64(size.Width*scale))),
		float32(math.Round(float64(size.Height*scale))),
	)
}
