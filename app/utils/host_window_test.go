package utils

import (
	"testing"

	fynetest "fyne.io/fyne/v2/test"
)

func TestNewHostWindow_BasicConfig(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	defer a.Quit()

	w := NewHostWindow(a)
	defer w.Close()

	if w.Title() != "FutuHost" {
		t.Fatalf("title = %q, want FutuHost", w.Title())
	}
	size := w.Canvas().Size()
	if size.Width != 800 || size.Height != 600 {
		t.Fatalf("size = (%v,%v), want (800,600)", size.Width, size.Height)
	}
}
