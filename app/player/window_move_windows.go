//go:build windows

package player

import (
	"math"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

type winRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

const (
	swpNoSize     = 0x0001
	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010
)

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	procGetWindowRect = user32.NewProc("GetWindowRect")
	procSetWindowPos  = user32.NewProc("SetWindowPos")
)

func getWindowPosition(w fyne.Window) (fyne.Position, bool) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return fyne.Position{}, false
	}

	got := false
	pos := fyne.Position{}
	nw.RunNative(func(context any) {
		winCtx, ok := context.(driver.WindowsWindowContext)
		if !ok {
			return
		}

		hwnd := uintptr(winCtx.HWND)
		if hwnd == 0 {
			return
		}

		var rect winRect
		r1, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
		if r1 == 0 {
			return
		}

		pos = fyne.NewPos(float32(rect.Left), float32(rect.Top))
		got = true
	})

	return pos, got
}

func moveWindowTo(w fyne.Window, x, y float32) bool {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return false
	}

	moved := false
	nw.RunNative(func(context any) {
		winCtx, ok := context.(driver.WindowsWindowContext)
		if !ok {
			return
		}
		hwnd := uintptr(winCtx.HWND)
		if hwnd == 0 {
			return
		}

		procSetWindowPos.Call(
			hwnd,
			0,
			uintptr(int32(math.Round(float64(x)))),
			uintptr(int32(math.Round(float64(y)))),
			0,
			0,
			swpNoSize|swpNoZOrder|swpNoActivate,
		)
		moved = true
	})

	return moved
}
