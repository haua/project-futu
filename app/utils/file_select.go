package utils

// 文件选择器

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

var (
	newHostWindow      = NewHostWindow
	showFileOpenDialog = dialog.ShowFileOpen
)

func ShowFileOpen(a fyne.App, callback func(fyne.URIReadCloser, error)) {
	// filter := dialog.NewExtensionFileFilter([]string{".txt", ".pdf"})

	host := newHostWindow(a)
	showFileOpenDialog(callback, host)
}
