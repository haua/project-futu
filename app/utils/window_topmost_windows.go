//go:build windows

package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type WindowTopMost struct {
	window fyne.Window
}

func NewWindowTopMost(w fyne.Window) *WindowTopMost {
	return &WindowTopMost{window: w}
}

const (
	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpNoActivate = 0x0010
)

var (
	user32TopMost       = windows.NewLazySystemDLL("user32.dll")
	procSetWindowPosTop = user32TopMost.NewProc("SetWindowPos")
)

func (c *WindowTopMost) Set(enabled bool) bool {
	if c == nil || c.window == nil {
		return false
	}

	nw, ok := c.window.(driver.NativeWindow)
	if !ok {
		return false
	}

	okSet := false
	nw.RunNative(func(context any) {
		winCtx, ok := context.(driver.WindowsWindowContext)
		if !ok {
			return
		}
		hwnd := uintptr(winCtx.HWND)
		if hwnd == 0 {
			return
		}

		after := uintptr(^uintptr(1)) // HWND_NOTOPMOST (-2)
		if enabled {
			after = uintptr(^uintptr(0)) // HWND_TOPMOST (-1)
		}

		r1, _, _ := procSetWindowPosTop.Call(
			hwnd,
			after,
			0,
			0,
			0,
			0,
			swpNoSize|swpNoMove|swpNoActivate,
		)
		okSet = r1 != 0
	})

	return okSet
}
