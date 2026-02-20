package drag

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Widget wraps a canvas object and forwards drag motion to the host window.
type Widget struct {
	widget.BaseWidget
	content       fyne.CanvasObject
	window        fyne.Window
	onDragChanged func(bool)
	dragging      bool
	startCursor   fyne.Position
	startWin      fyne.Position
}

func NewWidget(w fyne.Window, content fyne.CanvasObject, onDragChanged func(bool)) fyne.CanvasObject {
	d := &Widget{
		content:       content,
		window:        w,
		onDragChanged: onDragChanged,
	}
	d.ExtendBaseWidget(d)
	return d
}

func (d *Widget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.content)
}

// ev没有使用，是因为fyne拖拽事件传入的鼠标坐标不准，一个是会抖动，一个是拖拽距离不跟手，改成调用系统函数GetCursorPos获取的坐标了
func (d *Widget) Dragged(ev *fyne.DragEvent) {
	if d.window == nil {
		return
	}

	if !d.dragging {
		winPos, ok := getWindowPosition(d.window)
		if !ok {
			return
		}
		cursorPos, ok := getCursorPosition()
		if !ok {
			return
		}

		d.dragging = true
		if d.onDragChanged != nil {
			d.onDragChanged(true)
		}
		d.startWin = winPos
		d.startCursor = cursorPos
	}

	cursorPos, ok := getCursorPosition()
	if !ok {
		return
	}

	dx := cursorPos.X - d.startCursor.X
	dy := cursorPos.Y - d.startCursor.Y
	moveWindowTo(d.window, d.startWin.X+dx, d.startWin.Y+dy)
}

func (d *Widget) DragEnd() {
	if d.dragging && d.onDragChanged != nil {
		d.onDragChanged(false)
	}
	d.dragging = false
}
