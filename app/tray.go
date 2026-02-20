package app

// 系统托盘

import (
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"
	"github.com/haua/futu/app/utils"
	sqweek "github.com/sqweek/dialog"
)

func SetupTray(a fyne.App, win *FloatingWindow) {
	desk, ok := a.(desktop.App)
	if !ok {
		return
	}

	menu := fyne.NewMenu("Futu",
		fyne.NewMenuItem("更换图片", func() {
			// fyne用的是自己绘制的文件选择器，不好看
			// 这个 sqweek.File 才是系统原生的，好用
			allow := [...]string{"png", "jpeg", "jpg", "gif", "webp"}
			filename, err := sqweek.File().Filter(strings.Join(allow[:], ","), allow[:]...).Load()
			if err != nil {
				return
			}
			win.Player.Play(filename)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() {
			a.Quit()
		}),
	)

	desk.SetSystemTrayMenu(menu)

	doubleTapDelay := a.Driver().DoubleTapDelay()
	if doubleTapDelay <= 0 {
		doubleTapDelay = 300 * time.Millisecond
	}

	SetTrayIcon(desk, win.IsEditMode())

	var (
		lastTap time.Time
		tapMu   sync.Mutex
	)
	systray.SetOnTapped(func() {
		now := time.Now()
		isDoubleTap := false

		tapMu.Lock()
		if !lastTap.IsZero() && now.Sub(lastTap) <= doubleTapDelay {
			isDoubleTap = true
			lastTap = time.Time{}
		} else {
			lastTap = now
		}
		tapMu.Unlock()

		if isDoubleTap {
			win.ToggleEditMode()
			SetTrayIcon(desk, win.IsEditMode())
		}
	})
}

// 设置托盘图标
func SetTrayIcon(desk desktop.App, isEdit bool) {
	icon := "icon.png"
	if isEdit {
		icon = "icon-edit.png"
	}
	if trayIcon := utils.LoadAssetResource(icon); trayIcon != nil {
		desk.SetSystemTrayIcon(trayIcon)
	}
}
