//go:build !windows

package utils

import "fyne.io/fyne/v2"

type WindowTaskbar struct {
	window fyne.Window
}

func NewWindowTaskbar(w fyne.Window) *WindowTaskbar {
	return &WindowTaskbar{window: w}
}

func (c *WindowTaskbar) SetVisible(_ bool) bool {
	return false
}
