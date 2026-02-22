package app

import (
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
)

func BenchmarkUpdateWindowOpacityByCursor_StaticCursor(b *testing.B) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("bench")
	defer w.Close()

	oldGetCursor := getCursorPosition
	oldGetWindow := getWindowPosition
	oldWindowSize := windowSizeInPixels
	b.Cleanup(func() {
		getCursorPosition = oldGetCursor
		getWindowPosition = oldGetWindow
		windowSizeInPixels = oldWindowSize
	})

	getCursorPosition = func() (fyne.Position, bool) {
		return fyne.NewPos(500, 500), true
	}
	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		return fyne.NewPos(100, 100), true
	}
	windowSizeInPixels = func(fyne.Window) fyne.Size {
		return fyne.NewSize(200, 200)
	}

	fw := &FloatingWindow{
		Window:     w,
		opacitySet: func(float64) bool { return true },
	}
	fw.editMode.Store(false)
	fw.updateWindowOpacityByCursor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fw.updateWindowOpacityByCursor()
	}
}

func BenchmarkUpdateWindowOpacityByCursor_MovingCursor(b *testing.B) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("bench")
	defer w.Close()

	oldGetCursor := getCursorPosition
	oldGetWindow := getWindowPosition
	oldWindowSize := windowSizeInPixels
	b.Cleanup(func() {
		getCursorPosition = oldGetCursor
		getWindowPosition = oldGetWindow
		windowSizeInPixels = oldWindowSize
	})

	var x float32 = 10
	getCursorPosition = func() (fyne.Position, bool) {
		x++
		if x > 350 {
			x = 10
		}
		return fyne.NewPos(x, 10), true
	}
	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		return fyne.NewPos(100, 100), true
	}
	windowSizeInPixels = func(fyne.Window) fyne.Size {
		return fyne.NewSize(200, 200)
	}

	fw := &FloatingWindow{
		Window:     w,
		opacitySet: func(float64) bool { return true },
	}
	fw.editMode.Store(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fw.updateWindowOpacityByCursor()
	}
}
