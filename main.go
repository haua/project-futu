// 这是使用 fyne 的版本

package main

import (
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	futuapp "github.com/haua/futu/app"
)

func main() {
	a := fyneapp.NewWithID("cn.haua.futu.desktop")

	win := futuapp.NewFloatingWindow(a)
	a.Lifecycle().SetOnStarted(func() {
		futuapp.SetupTray(a, win)
	})
	a.Lifecycle().SetOnStopped(func() {
		win.Shutdown()
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		fyne.Do(func() {
			_ = win.ReapplyAlwaysOnTop()
		})
	})

	win.Show()
	a.Run()
}
