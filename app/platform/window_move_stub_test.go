//go:build !windows

package platform

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestGetWindowPositionStub(t *testing.T) {
	t.Parallel()

	pos, ok := GetWindowPosition(nil)
	if ok {
		t.Fatalf("stub GetWindowPosition should return ok=false")
	}
	if pos != (fyne.Position{}) {
		t.Fatalf("stub position should be zero value")
	}
}

func TestMoveWindowToStub(t *testing.T) {
	t.Parallel()

	if MoveWindowTo(nil, 1, 2) {
		t.Fatalf("stub MoveWindowTo should return false")
	}
}

func TestIsWindowInVisibleBoundsStub(t *testing.T) {
	t.Parallel()

	if !IsWindowInVisibleBounds(fyne.NewPos(-999, -999), fyne.NewSize(1, 1)) {
		t.Fatalf("stub IsWindowInVisibleBounds should return true")
	}
}

func TestGetCursorPositionStub(t *testing.T) {
	t.Parallel()

	pos, ok := GetCursorPosition()
	if ok {
		t.Fatalf("stub GetCursorPosition should return ok=false")
	}
	if pos != (fyne.Position{}) {
		t.Fatalf("stub cursor position should be zero value")
	}
}

func TestGetScreenWidthPixelsStub(t *testing.T) {
	t.Parallel()

	width, ok := GetScreenWidthPixels()
	if !ok {
		t.Fatalf("stub GetScreenWidthPixels should return ok=true")
	}
	if width != 1920 {
		t.Fatalf("stub screen width should be 1920")
	}
}
