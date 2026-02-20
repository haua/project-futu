//go:build !windows

package utils

import "fyne.io/fyne/v2"

type WindowTopMost struct {
	window fyne.Window
}

func NewWindowTopMost(w fyne.Window) *WindowTopMost {
	return &WindowTopMost{window: w}
}

func (c *WindowTopMost) Set(_ bool) bool {
	return false
}
