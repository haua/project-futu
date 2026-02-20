// 这是使用 fyne 的版本

package main

import (
	fyneapp "fyne.io/fyne/v2/app"

	futuapp "github.com/haua/futu/app"
)

func main() {
	a := fyneapp.NewWithID("com.futu.desktop")

	win := futuapp.NewFloatingWindow(a)
	a.Lifecycle().SetOnStarted(func() {
		futuapp.SetupTray(a, win)
	})

	win.Show()
	a.Run()
}
