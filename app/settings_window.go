package app

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	return "后续设置项会追加到本页。"
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

func openSettingsWindow(a fyne.App) {
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
		newSelectableText(settingsAppendNoticeText(), true, baseTheme),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(480, 360))

	settingsWin.SetContent(scroll)
	settingsWin.Resize(fyne.NewSize(520, 420))
	settingsWin.Show()
}
