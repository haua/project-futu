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

var (
	loadAssetResource = utils.LoadAssetResource
	setOnTrayTapped   = systray.SetOnTapped
)

func topMostMenuLabel(enabled bool) string {
	if enabled {
		return "置顶：开"
	}
	return "置顶：关"
}

func imageFileFilters() (string, []string) {
	allow := []string{"png", "jpeg", "jpg", "gif", "webp"}
	return strings.Join(allow, ","), allow
}

func trayIconName(isEdit bool) string {
	if isEdit {
		return "icon-edit.png"
	}
	return "icon.png"
}

func detectDoubleTap(lastTap, now time.Time, delay time.Duration) (bool, time.Time) {
	if !lastTap.IsZero() && now.Sub(lastTap) <= delay {
		return true, time.Time{}
	}
	return false, now
}

func setTrayIconByState(setIcon func(fyne.Resource), isEdit bool) {
	icon := trayIconName(isEdit)
	if trayIcon := loadAssetResource(icon); trayIcon != nil {
		setIcon(trayIcon)
	}
}

func SetupTray(a fyne.App, win *FloatingWindow) {
	desk, ok := a.(desktop.App)
	if !ok {
		return
	}

	var menu *fyne.Menu
	var topMostItem *fyne.MenuItem
	topMostItem = fyne.NewMenuItem(topMostMenuLabel(win.IsAlwaysOnTop()), func() {
		next := !win.IsAlwaysOnTop()
		if !win.SetAlwaysOnTop(next) {
			return
		}
		topMostItem.Label = topMostMenuLabel(next)
		desk.SetSystemTrayMenu(menu)
	})

	menu = fyne.NewMenu("Futu",
		topMostItem,
		fyne.NewMenuItem("更换图片", func() {
			// fyne 的文件选择器不是系统原生，这里用 sqweek 的系统文件选择器。
			filterName, allow := imageFileFilters()
			filename, err := sqweek.File().Filter(filterName, allow...).Load()
			if err != nil {
				return
			}
			win.Player.Play(filename)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("设置", func() {
			openSettingsWindow(a)
		}),
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
	setOnTrayTapped(func() {
		now := time.Now()

		tapMu.Lock()
		isDoubleTap, nextLastTap := detectDoubleTap(lastTap, now, doubleTapDelay)
		lastTap = nextLastTap
		tapMu.Unlock()

		if isDoubleTap {
			win.ToggleEditMode()
			SetTrayIcon(desk, win.IsEditMode())
		}
	})
}

// 设置托盘图标
func SetTrayIcon(desk desktop.App, isEdit bool) {
	setTrayIconByState(desk.SetSystemTrayIcon, isEdit)
}
