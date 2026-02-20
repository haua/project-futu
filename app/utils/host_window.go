package utils

// 创建一个临时窗口，用来承载设置界面

import "fyne.io/fyne/v2"

func NewHostWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("FutuHost")
	w.Resize(fyne.NewSize(800, 600))
	w.SetPadded(false)
	w.Show()
	w.SetOnClosed(func() {}) // 防止误关，但是估计可以搞个回调
	return w
}
