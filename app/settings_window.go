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
	sqweek "github.com/sqweek/dialog"
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
		"一张浮图，半刻治愈",
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

func hotkeyButtonText(label string) string {
	if strings.TrimSpace(label) == "" {
		return "未设置"
	}
	return label
}

func newHotkeyRecordRow(
	win *FloatingWindow,
	settingsWin fyne.Window,
	rowLabel string,
	getCurrent func() string,
	setHotkey func(string) bool,
) fyne.CanvasObject {
	if win == nil {
		return widget.NewLabel("无法加载快捷键设置")
	}
	if !win.IsGlobalHotkeySupported() {
		return widget.NewLabel("当前平台不支持全局快捷键")
	}

	recording := false
	originHotkey := getCurrent()
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
		recordBtn.SetText(hotkeyButtonText(originHotkey))
		clearHooks()
	}

	stopRecordWithCurrent := func() {
		if !recording {
			return
		}
		recording = false
		win.EndModeToggleHotkeyCapture()
		recordBtn.SetText(hotkeyButtonText(getCurrent()))
		clearHooks()
	}

	recordBtn = newFocusAwareButton(hotkeyButtonText(getCurrent()), nil, func() {
		cancelRecord()
	})

	clearBtn := widget.NewButton("清空", func() {
		if recording {
			cancelRecord()
		}
		if !setHotkey("") {
			return
		}
		recordBtn.SetText(hotkeyButtonText(getCurrent()))
	})

	recordBtn.OnTapped = func() {
		if recording {
			return
		}
		if settingsWin == nil || settingsWin.Canvas() == nil {
			return
		}

		originHotkey = getCurrent()
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
			if !setHotkey(label) {
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

	return container.NewBorder(nil, nil, widget.NewLabel(rowLabel), clearBtn, recordBtn)
}

func newModeToggleHotkeySetting(win *FloatingWindow, settingsWin fyne.Window) fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("全局快捷键（设置时必须带修饰键Ctrl/Alt/Shift，如遇冲突会设置失败）"),
		newHotkeyRecordRow(win, settingsWin, "切换模式", win.ModeToggleHotkey, win.SetModeToggleHotkey),
		newHotkeyRecordRow(win, settingsWin, "隐藏窗口", win.HideWindowHotkey, win.SetHideWindowHotkey),
	)
}

func imageSourceModeLabel(mode string) string {
	if mode == imageSourceModeFolder {
		return "文件夹随机（每小时）"
	}
	return "固定图片"
}

func imageSourceModeFromLabel(label string) string {
	if strings.Contains(label, "文件夹") {
		return imageSourceModeFolder
	}
	return imageSourceModeSingle
}

func sourcePathText(prefix, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return prefix + "未设置"
	}
	return prefix + path
}

func newImageSourceSetting(win *FloatingWindow) fyne.CanvasObject {
	if win == nil {
		return widget.NewLabel("无法加载播放来源设置")
	}

	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord
	status.Hide()

	fixedPathLabel := widget.NewLabel("")
	fixedPathLabel.Wrapping = fyne.TextWrapWord
	folderPathLabel := widget.NewLabel("")
	folderPathLabel.Wrapping = fyne.TextWrapWord

	modeRadio := widget.NewRadioGroup([]string{
		imageSourceModeLabel(imageSourceModeSingle),
		imageSourceModeLabel(imageSourceModeFolder),
	}, nil)

	refreshView := func() {
		fixedPathLabel.SetText(sourcePathText("固定图片：", win.FixedImagePath()))
		folderPathLabel.SetText(sourcePathText("随机文件夹：", win.RandomFolderPath()))
		modeRadio.SetSelected(imageSourceModeLabel(win.ImageSourceMode()))
	}
	refreshView()

	showError := func(text string) {
		status.SetText(text)
		status.Show()
	}
	hideError := func() {
		status.Hide()
		status.SetText("")
	}

	modeRadio.OnChanged = func(selected string) {
		if selected == "" {
			return
		}
		if win.SetImageSourceMode(imageSourceModeFromLabel(selected)) {
			hideError()
			refreshView()
			return
		}
		refreshView()
		showError("切换失败：请先选择有效的图片或文件夹（文件夹需包含支持格式）")
	}

	selectFixedBtn := widget.NewButton("选择图片", func() {
		filterName, allow := imageFileFilters()
		filename, err := sqweek.File().Filter(filterName, allow...).Load()
		if err != nil {
			return
		}
		if !win.SetFixedImage(filename) {
			showError("设置失败：请选择有效图片文件")
			return
		}
		hideError()
		refreshView()
	})

	selectFolderBtn := widget.NewButton("选择文件夹", func() {
		folder, err := sqweek.Directory().Browse()
		if err != nil {
			return
		}
		if !win.SetRandomImageFolder(folder) {
			showError("设置失败：文件夹不存在或没有支持的图片")
			return
		}
		hideError()
		refreshView()
	})

	randomNowBtn := widget.NewButton("立即随机一张", func() {
		if win.PlayRandomImageNow() {
			hideError()
			return
		}
		showError("随机失败：请先设置有效的图片文件夹")
	})

	return container.NewVBox(
		widget.NewLabel("播放来源"),
		modeRadio,
		container.NewHBox(selectFixedBtn, selectFolderBtn, randomNowBtn),
		fixedPathLabel,
		folderPathLabel,
		status,
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
		newImageSourceSetting(win),
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
