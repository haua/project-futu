package app

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/haua/futu/app/utils"
)

const (
	settingsAuthor  = "haua (https://github.com/haua)"
	settingsLicense = "MIT"
	projectURL      = "https://github.com/haua/project-futu"
)

type plainSelectableTheme struct {
	base fyne.Theme
}

func (t plainSelectableTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground, theme.ColorNameInputBorder:
		return color.Transparent
	default:
		return t.base.Color(name, variant)
	}
}

func (t plainSelectableTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t plainSelectableTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t plainSelectableTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameInputBorder {
		return 0
	}
	return t.base.Size(name)
}

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
		"项目地址：" + projectURL,
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

func newSelectableText(text string, multiline bool, baseTheme fyne.Theme) fyne.CanvasObject {
	var entry *widget.Entry
	if multiline {
		entry = widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		// Let outer dialog scroll handle overflow by expanding entry height to content rows.
		entry.SetMinRowsVisible(strings.Count(text, "\n") + 1)
	} else {
		entry = widget.NewEntry()
		entry.Wrapping = fyne.TextWrapOff
	}

	entry.SetText(text)

	lockedText := text
	resetting := false
	entry.OnChanged = func(current string) {
		if resetting || current == lockedText {
			return
		}
		resetting = true
		entry.SetText(lockedText)
		resetting = false
	}

	return container.NewThemeOverride(entry, plainSelectableTheme{base: baseTheme})
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
			status.SetText("")
			status.Hide()
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

func openSettingsWindow(a fyne.App, win *FloatingWindow) {
	if a == nil {
		return
	}

	settingsWin := a.NewWindow("设置")
	baseTheme := a.Settings().Theme()

	content := container.NewVBox(
		newSelectableText(softwareInfoText(a), true, baseTheme),
		widget.NewSeparator(),
		newSelectableText(operationGuideText(), true, baseTheme),
		widget.NewSeparator(),
		newSelectableText(settingsAppendNoticeText(), false, baseTheme),
		newLaunchAtStartupSetting(win),
		newMouseFarOpacitySetting(win),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(480, 360))

	settingsWin.SetContent(scroll)
	settingsWin.Resize(fyne.NewSize(520, 420))
	settingsWin.Show()
}
