package app

// 系统托盘

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/haua/futu/app/utils"
	sqweek "github.com/sqweek/dialog"
)

func SetupTray(a fyne.App, win *FloatingWindow) {
	desk, ok := a.(desktop.App)
	if !ok {
		// 非桌面平台没有系统托盘
		return
	}

	menu := fyne.NewMenu("Futu",
		fyne.NewMenuItem("更换图片", func() {
			// fyne用的是自己绘制的文件选择器，不好看
			// 这个才是系统原生的，好用
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

	// 设置托盘图标，优先使用 assets/icon-tray.png
	if trayIcon := utils.LoadAssetResource("icon-tray.png"); trayIcon != nil {
		desk.SetSystemTrayIcon(trayIcon)
	}
}
