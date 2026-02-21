//go:build !windows

package utils

import "fyne.io/fyne/v2"

type WindowMousePassthrough struct {
	window fyne.Window
}

func NewWindowMousePassthrough(w fyne.Window) *WindowMousePassthrough {
	return &WindowMousePassthrough{window: w}
}

func (c *WindowMousePassthrough) SetEnabled(_ bool) bool {
	return false
}
