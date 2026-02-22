package app

import (
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	desktopdrv "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/haua/futu/app/utils"
)

const (
	settingsAuthor  = "haua (https://github.com/haua)"
	settingsLicense = "MIT"
	projectURL      = "https://github.com/haua/project-futu"
)

func appVersionText(a fyne.App) string {
	version := "unknown"
	if a != nil && a.Metadata().Version != "" {
		version = a.Metadata().Version
	}
	return fmt.Sprintf("版本：%s", version)
}

func softwareInfoText(a fyne.App) string {
	return strings.Join([]string{
		"软件信息",
		appVersionText(a),
		"作者：" + settingsAuthor,
		"许可证：" + settingsLicense,
	}, "\n")
}

func operationGuideText() string {
	return strings.Join([]string{
		"操作指南：",
		"1. 每次启动应用都会进入编辑模式",
		"2. 双击托盘图标可切换编辑模式与常态模式",
		"3. 编辑模式支持拖拽窗口、滚轮缩放",
		"4. 常态模式会在鼠标靠近时隐藏窗口，不影响你的操作",
	}, "\n")
}

func settingsAppendNoticeText() string {
	return "设置："
}

func newReadonlyText(text string) fyne.CanvasObject {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func newSoftwareInfoSection(a fyne.App) fyne.CanvasObject {
	items := []fyne.CanvasObject{newReadonlyText(softwareInfoText(a))}

	parsedURL, err := url.Parse(projectURL)
	if err != nil {
		items = append(items, widget.NewLabel("项目地址："+projectURL))
		return container.NewVBox(items...)
	}

	link := widget.NewHyperlink("项目地址："+projectURL, parsedURL)
	items = append(items, link)
	return container.NewVBox(items...)
}

func newMouseFarOpacitySetting(win *FloatingWindow) fyne.CanvasObject {
	if win == nil {
		return widget.NewLabel("无法加载透明度设置")
	}

	label := widget.NewLabel("")
	slider := widget.NewSlider(1, 100)
	slider.Step = 1

	updateLabel := func(v float64) {
		label.SetText(fmt.Sprintf("常态模式下最高不透明度：%d%%", int(v+0.5)))
	}

	current := utils.ClampFloat64(win.MouseFarOpacity()*100, 1, 100)
	slider.SetValue(current)
	updateLabel(current)
	slider.OnChanged = func(v float64) {
		updateLabel(v)
		win.SetMouseFarOpacity(v / 100)
	}

	return container.NewVBox(label, slider)
}

func newLaunchAtStartupSetting(win *FloatingWindow) fyne.CanvasObject {
	if win == nil {
		return widget.NewLabel("无法加载开机自启设置")
	}

	status := widget.NewLabel("")
	status.Hide()
	check := widget.NewCheck("开机自启", nil)

	resetting := false
	check.OnChanged = func(enabled bool) {
		if resetting {
			return
		}
		if win.SetLaunchAtStartup(enabled) {
			status.Hide()
			status.SetText("")
			return
		}

		resetting = true
		check.SetChecked(!enabled)
		resetting = false
		status.SetText("设置失败：请检查系统权限")
		status.Show()
	}

	if !win.RefreshLaunchAtStartup() {
		status.SetText("提示：无法读取当前开机自启状态")
		status.Show()
	}
	check.SetChecked(win.IsLaunchAtStartup())

	return container.NewVBox(check, status)
}

type focusAwareButton struct {
	widget.Button
	onFocusLost func()
}

func newFocusAwareButton(label string, tapped func(), onFocusLost func()) *focusAwareButton {
	b := &focusAwareButton{onFocusLost: onFocusLost}
	b.Text = label
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *focusAwareButton) FocusLost() {
	b.Button.FocusLost()
	if b.onFocusLost != nil {
		b.onFocusLost()
	}
}

func newModeToggleHotkeySetting(win *FloatingWindow, settingsWin fyne.Window) fyne.CanvasObject {
	if win == nil {
		return widget.NewLabel("无法加载快捷键设置")
	}
	if !win.IsGlobalHotkeySupported() {
		return widget.NewLabel("当前平台不支持全局快捷键")
	}

	recording := false
	originHotkey := win.ModeToggleHotkey()
	var oldOnTypedKey func(*fyne.KeyEvent)
	var registeredShortcuts []*desktopdrv.CustomShortcut

	var recordBtn *focusAwareButton

	clearHooks := func() {
		if settingsWin == nil || settingsWin.Canvas() == nil {
			return
		}
		canvas := settingsWin.Canvas()
		canvas.SetOnTypedKey(oldOnTypedKey)
		for _, s := range registeredShortcuts {
			canvas.RemoveShortcut(s)
		}
		registeredShortcuts = nil
	}

	cancelRecord := func() {
		if !recording {
			return
		}
		recording = false
		win.EndModeToggleHotkeyCapture()
		recordBtn.SetText(originHotkey)
		clearHooks()
	}

	stopRecordWithCurrent := func() {
		if !recording {
			return
		}
		recording = false
		win.EndModeToggleHotkeyCapture()
		recordBtn.SetText(win.ModeToggleHotkey())
		clearHooks()
	}

	recordBtn = newFocusAwareButton(win.ModeToggleHotkey(), nil, func() {
		cancelRecord()
	})

	recordBtn.OnTapped = func() {
		if recording {
			return
		}
		if settingsWin == nil || settingsWin.Canvas() == nil {
			return
		}

		originHotkey = win.ModeToggleHotkey()
		recording = true
		win.BeginModeToggleHotkeyCapture()
		recordBtn.SetText("录制中...按下组合键")

		canvas := settingsWin.Canvas()
		oldOnTypedKey = canvas.OnTypedKey()
		registeredShortcuts = nil

		capture := func(key fyne.KeyName, mod fyne.KeyModifier) {
			if !recording {
				return
			}
			label, ok := modeToggleHotkeyFromKeyEvent(key, mod)
			if !ok {
				return
			}
			if label == originHotkey {
				cancelRecord()
				return
			}
			if !win.SetModeToggleHotkey(label) {
				cancelRecord()
				return
			}
			stopRecordWithCurrent()
		}

		mods := []fyne.KeyModifier{
			fyne.KeyModifierControl,
			fyne.KeyModifierAlt,
			fyne.KeyModifierShift,
			fyne.KeyModifierControl | fyne.KeyModifierAlt,
			fyne.KeyModifierControl | fyne.KeyModifierShift,
			fyne.KeyModifierAlt | fyne.KeyModifierShift,
			fyne.KeyModifierControl | fyne.KeyModifierAlt | fyne.KeyModifierShift,
		}
		for _, key := range supportedModeToggleHotkeyKeys() {
			for _, mod := range mods {
				sc := &desktopdrv.CustomShortcut{KeyName: key, Modifier: mod}
				registeredShortcuts = append(registeredShortcuts, sc)
				k := key
				m := mod
				canvas.AddShortcut(sc, func(_ fyne.Shortcut) {
					capture(k, m)
				})
			}
		}

		canvas.SetOnTypedKey(func(ev *fyne.KeyEvent) {
			if !recording || ev == nil {
				if oldOnTypedKey != nil {
					oldOnTypedKey(ev)
				}
				return
			}
			if ev.Name == fyne.KeyEscape {
				cancelRecord()
			}
		})

		settingsWin.Canvas().Focus(recordBtn)
	}

	if settingsWin != nil {
		settingsWin.SetOnClosed(func() {
			cancelRecord()
		})
	}

	return container.NewVBox(
		widget.NewLabel("全局快捷键（设置时必须带修饰键Ctrl/Alt/Shift，如遇冲突会设置失败）"),
		container.NewBorder(nil, nil, widget.NewLabel("切换模式"), nil, recordBtn),
	)
}

func openSettingsWindow(a fyne.App, win *FloatingWindow) {
	if a == nil {
		return
	}

	settingsWin := a.NewWindow("设置")

	content := container.NewVBox(
		newSoftwareInfoSection(a),
		widget.NewSeparator(),
		newReadonlyText(operationGuideText()),
		widget.NewSeparator(),
		newReadonlyText(settingsAppendNoticeText()),
		newLaunchAtStartupSetting(win),
		newMouseFarOpacitySetting(win),
		widget.NewSeparator(),
		newModeToggleHotkeySetting(win, settingsWin),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(480, 360))

	settingsWin.SetContent(scroll)
	settingsWin.Resize(fyne.NewSize(520, 420))
	settingsWin.Show()
}
