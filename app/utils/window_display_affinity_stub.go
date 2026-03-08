//go:build !windows

package utils

import "fyne.io/fyne/v2"

type WindowDisplayAffinity struct {
	window fyne.Window
}

func NewWindowDisplayAffinity(w fyne.Window) *WindowDisplayAffinity {
	return &WindowDisplayAffinity{window: w}
}

func (c *WindowDisplayAffinity) SetExcludeFromCapture(_ bool) bool {
	return false
}
