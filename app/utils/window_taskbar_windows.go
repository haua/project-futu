//go:build windows

package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type WindowTaskbar struct {
	window fyne.Window
}

func NewWindowTaskbar(w fyne.Window) *WindowTaskbar {
	return &WindowTaskbar{window: w}
}

const (
	gwlExStyle       = ^uintptr(19) // GWL_EXSTYLE (-20)
	wsExToolWindow   = 0x00000080
	wsExAppWindow    = 0x00040000
	swpNoSizeTaskbar = 0x0001
	swpNoMoveTaskbar = 0x0002
	swpNoZOrder      = 0x0004
	swpNoActivate2   = 0x0010
	swpFrameChanged  = 0x0020
)

var (
	user32Taskbar           = windows.NewLazySystemDLL("user32.dll")
	procGetWindowLongPtrW   = user32Taskbar.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW   = user32Taskbar.NewProc("SetWindowLongPtrW")
	procSetWindowPosTaskbar = user32Taskbar.NewProc("SetWindowPos")
)

func (c *WindowTaskbar) SetVisible(visible bool) bool {
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

		styleRaw, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
		style := styleRaw
		if visible {
			style |= wsExAppWindow
			style &^= wsExToolWindow
		} else {
			style |= wsExToolWindow
			style &^= wsExAppWindow
		}

		procSetWindowLongPtrW.Call(hwnd, gwlExStyle, style)

		r1, _, _ := procSetWindowPosTaskbar.Call(
			hwnd,
			0,
			0,
			0,
			0,
			0,
			swpNoSizeTaskbar|swpNoMoveTaskbar|swpNoZOrder|swpNoActivate2|swpFrameChanged,
		)
		okSet = r1 != 0
	})

	return okSet
}
