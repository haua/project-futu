package app

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
		return "\u7f6e\u9876\uff1a\u5f00"
	}
	return "\u7f6e\u9876\uff1a\u5173"
}

func windowVisibilityMenuLabel(visible bool) string {
	if visible {
		return "\u7a97\u53e3\uff1a\u663e\u793a"
	}
	return "\u7a97\u53e3\uff1a\u9690\u85cf"
}

func modeMenuLabel(isEdit bool) string {
	return modeHintText(isEdit)
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

func pickAndPlayImage(win *FloatingWindow) {
	if win == nil {
		return
	}
	filterName, allow := imageFileFilters()
	filename, err := sqweek.File().Filter(filterName, allow...).Load()
	if err != nil {
		return
	}
	_ = win.SetFixedImage(filename)
}

func SetupTray(a fyne.App, win *FloatingWindow) {
	desk, ok := a.(desktop.App)
	if !ok {
		return
	}

	var menu *fyne.Menu
	var topMostItem *fyne.MenuItem
	var modeItem *fyne.MenuItem
	var windowVisibilityItem *fyne.MenuItem

	refreshTrayState := func(isEdit bool) {
		modeItem.Label = modeMenuLabel(isEdit)
		windowVisibilityItem.Label = windowVisibilityMenuLabel(win.IsWindowVisible())
		desk.SetSystemTrayMenu(menu)
		SetTrayIcon(desk, isEdit)
	}

	toggleMode := func() {
		next := win.ToggleEditMode()
		refreshTrayState(next)
	}

	topMostItem = fyne.NewMenuItem(topMostMenuLabel(win.IsAlwaysOnTop()), func() {
		next := !win.IsAlwaysOnTop()
		if !win.SetAlwaysOnTop(next) {
			return
		}
		topMostItem.Label = topMostMenuLabel(next)
		desk.SetSystemTrayMenu(menu)
	})
	modeItem = fyne.NewMenuItem(modeMenuLabel(win.IsEditMode()), func() {
		toggleMode()
	})
	windowVisibilityItem = fyne.NewMenuItem(windowVisibilityMenuLabel(win.IsWindowVisible()), func() {
		visible := win.ToggleWindowVisibility()
		windowVisibilityItem.Label = windowVisibilityMenuLabel(visible)
		desk.SetSystemTrayMenu(menu)
	})

	menu = fyne.NewMenu("Futu",
		modeItem,
		topMostItem,
		windowVisibilityItem,
		fyne.NewMenuItem("\u66f4\u6362\u56fe\u7247", func() {
			// Use native file picker for better UX than Fyne file dialog.
			pickAndPlayImage(win)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("\u8bbe\u7f6e", func() {
			openSettingsWindow(a, win)
		}),
		fyne.NewMenuItem("\u9000\u51fa", func() {
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
			toggleMode()
		}
	})
}

func SetTrayIcon(desk desktop.App, isEdit bool) {
	setTrayIconByState(desk.SetSystemTrayIcon, isEdit)
}
