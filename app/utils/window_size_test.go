package utils

import (
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
)

func TestWindowSizeInPixels_UsesCanvasSize(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	w.Resize(fyne.NewSize(321, 123))
	got := WindowSizeInPixels(w)

	if got.Width != 321 || got.Height != 123 {
		t.Fatalf("size = (%v, %v), want (321, 123)", got.Width, got.Height)
	}
}

func TestWindowSizeInPixels_DefaultFallback(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	got := WindowSizeInPixels(w)
	canvasSize := w.Canvas().Size()

	if got.Width != canvasSize.Width || got.Height != canvasSize.Height {
		t.Fatalf(
			"default size = (%v, %v), want canvas size (%v, %v)",
			got.Width, got.Height, canvasSize.Width, canvasSize.Height,
		)
	}
}
