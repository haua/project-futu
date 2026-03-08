//go:build windows

package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type WindowDisplayAffinity struct {
	window fyne.Window
}

func NewWindowDisplayAffinity(w fyne.Window) *WindowDisplayAffinity {
	return &WindowDisplayAffinity{window: w}
}

const (
	wdaNone               = 0x0
	wdaExcludeFromCapture = 0x11
)

var (
	user32DisplayAffinity        = windows.NewLazySystemDLL("user32.dll")
	procSetWindowDisplayAffinity = user32DisplayAffinity.NewProc("SetWindowDisplayAffinity")
)

func (c *WindowDisplayAffinity) SetExcludeFromCapture(exclude bool) bool {
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

		affinity := uintptr(wdaNone)
		if exclude {
			affinity = uintptr(wdaExcludeFromCapture)
		}
		r1, _, _ := procSetWindowDisplayAffinity.Call(hwnd, affinity)
		okSet = r1 != 0
	})

	return okSet
}
