//go:build windows

package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type WindowMousePassthrough struct {
	window fyne.Window
}

func NewWindowMousePassthrough(w fyne.Window) *WindowMousePassthrough {
	return &WindowMousePassthrough{window: w}
}

const (
	gwlExStyleMouse      = ^uintptr(19) // GWL_EXSTYLE (-20)
	wsExLayeredMouse     = 0x00080000
	wsExTransparentMouse = 0x00000020
	swpNoSizeMouse       = 0x0001
	swpNoMoveMouse       = 0x0002
	swpNoZOrderMouse     = 0x0004
	swpNoActivateMouse   = 0x0010
	swpFrameChangedMouse = 0x0020
)

var (
	user32Mouse            = windows.NewLazySystemDLL("user32.dll")
	procGetWindowLongMouse = user32Mouse.NewProc("GetWindowLongPtrW")
	procSetWindowLongMouse = user32Mouse.NewProc("SetWindowLongPtrW")
	procSetWindowPosMouse  = user32Mouse.NewProc("SetWindowPos")
)

func (c *WindowMousePassthrough) SetEnabled(enabled bool) bool {
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

		style, _, _ := procGetWindowLongMouse.Call(hwnd, gwlExStyleMouse)
		if enabled {
			style |= wsExLayeredMouse | wsExTransparentMouse
		} else {
			style &^= wsExTransparentMouse
		}

		procSetWindowLongMouse.Call(hwnd, gwlExStyleMouse, style)

		r1, _, _ := procSetWindowPosMouse.Call(
			hwnd,
			0,
			0,
			0,
			0,
			0,
			swpNoSizeMouse|swpNoMoveMouse|swpNoZOrderMouse|swpNoActivateMouse|swpFrameChangedMouse,
		)
		okSet = r1 != 0
	})

	return okSet
}
