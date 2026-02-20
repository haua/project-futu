//go:build !windows

package player

import "fyne.io/fyne/v2"

func getWindowPosition(_ fyne.Window) (fyne.Position, bool) {
	return fyne.Position{}, false
}

func moveWindowTo(_ fyne.Window, _, _ float32) bool {
	return false
}
