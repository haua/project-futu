//go:build windows

package utils

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type WindowOpacity struct {
	window fyne.Window
}

func NewWindowOpacity(w fyne.Window) *WindowOpacity {
	return &WindowOpacity{window: w}
}

const (
	gwlExStyleOpacity      = ^uintptr(19) // GWL_EXSTYLE (-20)
	wsExLayeredOpacity     = 0x00080000
	lwaAlpha               = 0x00000002
	swpNoSizeOpacity       = 0x0001
	swpNoMoveOpacity       = 0x0002
	swpNoZOrderOpacity     = 0x0004
	swpNoActivateOpacity   = 0x0010
	swpFrameChangedOpacity = 0x0020
)

var (
	user32Opacity                  = windows.NewLazySystemDLL("user32.dll")
	procGetWindowLongPtrWOpacity   = user32Opacity.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrWOpacity   = user32Opacity.NewProc("SetWindowLongPtrW")
	procSetLayeredWindowAttributes = user32Opacity.NewProc("SetLayeredWindowAttributes")
	procSetWindowPosOpacity        = user32Opacity.NewProc("SetWindowPos")
)

func (c *WindowOpacity) Set(opacity float64) bool {
	if c == nil || c.window == nil {
		return false
	}

	nw, ok := c.window.(driver.NativeWindow)
	if !ok {
		return false
	}

	opacity = ClampFloat64(opacity, 0, 1)
	alpha := byte(math.Round(opacity * 255))

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

		style, _, _ := procGetWindowLongPtrWOpacity.Call(hwnd, gwlExStyleOpacity)
		style |= wsExLayeredOpacity
		procSetWindowLongPtrWOpacity.Call(hwnd, gwlExStyleOpacity, style)

		r1, _, _ := procSetLayeredWindowAttributes.Call(
			hwnd,
			0,
			uintptr(alpha),
			lwaAlpha,
		)
		if r1 == 0 {
			return
		}

		r2, _, _ := procSetWindowPosOpacity.Call(
			hwnd,
			0,
			0,
			0,
			0,
			0,
			swpNoSizeOpacity|swpNoMoveOpacity|swpNoZOrderOpacity|swpNoActivateOpacity|swpFrameChangedOpacity,
		)
		okSet = r2 != 0
	})

	return okSet
}
