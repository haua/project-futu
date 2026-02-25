//go:build !windows

package platform

import "fyne.io/fyne/v2"

func GetWindowPosition(_ fyne.Window) (fyne.Position, bool) {
	return fyne.Position{}, false
}

func MoveWindowTo(_ fyne.Window, _, _ float32) bool {
	return false
}

func IsWindowInVisibleBounds(_ fyne.Position, _ fyne.Size) bool {
	return true
}

func GetCursorPosition() (fyne.Position, bool) {
	return fyne.Position{}, false
}

func GetScreenWidthPixels() (int32, bool) {
	return 1920, true
}
