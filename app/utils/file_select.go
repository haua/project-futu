package utils

// 文件选择器

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func ShowFileOpen(a fyne.App, callback func(fyne.URIReadCloser, error)) {
	// filter := dialog.NewExtensionFileFilter([]string{".txt", ".pdf"})

	host := NewHostWindow(a)
	dialog.ShowFileOpen(callback, host)
}
