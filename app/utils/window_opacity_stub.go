//go:build !windows

package utils

import "fyne.io/fyne/v2"

type WindowOpacity struct {
	window fyne.Window
}

func NewWindowOpacity(w fyne.Window) *WindowOpacity {
	return &WindowOpacity{window: w}
}

func (c *WindowOpacity) Set(_ float64) bool {
	return false
}
