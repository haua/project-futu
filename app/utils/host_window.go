package utils

// 创建一个隐藏窗口，文件选择器绑定到这个窗口中

import "fyne.io/fyne/v2"

func NewHostWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("FutuHost")
	w.Resize(fyne.NewSize(800, 600))
	w.SetPadded(false)
	w.Show()
	w.SetOnClosed(func() {}) // 防止误关，但是估计可以搞个回调
	return w
}
