package drag

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fynetest "fyne.io/fyne/v2/test"
)

func TestScrolled_RespectsEditMode(t *testing.T) {
	t.Parallel()

	ev := &fyne.ScrollEvent{
		Scrolled: fyne.Delta{DX: 1, DY: -1},
	}

	t.Run("blocked when not edit mode", func(t *testing.T) {
		t.Parallel()

		called := false
		w := &Widget{
			isEditMode: func() bool { return false },
			onScrolled: func(*fyne.ScrollEvent) { called = true },
		}

		w.Scrolled(ev)
		if called {
			t.Fatalf("onScrolled should not be called when edit mode is disabled")
		}
	})

	t.Run("called when edit mode", func(t *testing.T) {
		t.Parallel()

		var got *fyne.ScrollEvent
		w := &Widget{
			isEditMode: func() bool { return true },
			onScrolled: func(e *fyne.ScrollEvent) { got = e },
		}

		w.Scrolled(ev)
		if got != ev {
			t.Fatalf("onScrolled should receive original event pointer")
		}
	})
}

func TestDragEnd_OnlyEmitsWhenDragging(t *testing.T) {
	t.Parallel()

	t.Run("emit false and reset when dragging", func(t *testing.T) {
		t.Parallel()

		var calledWith []bool
		w := &Widget{
			dragging: true,
			onDragChanged: func(v bool) {
				calledWith = append(calledWith, v)
			},
		}

		w.DragEnd()

		if w.dragging {
			t.Fatalf("dragging should be reset to false")
		}
		if len(calledWith) != 1 || calledWith[0] != false {
			t.Fatalf("onDragChanged should be called once with false, got %v", calledWith)
		}
	})

	t.Run("no emit when already idle", func(t *testing.T) {
		t.Parallel()

		called := false
		w := &Widget{
			dragging: false,
			onDragChanged: func(bool) {
				called = true
			},
		}

		w.DragEnd()

		if called {
			t.Fatalf("onDragChanged should not be called when not dragging")
		}
	})
}

func TestNewWidget_InitializesFields(t *testing.T) {
	t.Parallel()

	content := canvas.NewRectangle(nil)
	obj := NewWidget(nil, content, nil, nil, nil, nil)
	w, ok := obj.(*Widget)
	if !ok {
		t.Fatalf("NewWidget should return *Widget")
	}
	if w.content != content {
		t.Fatalf("widget content should be assigned")
	}
	if w.window != nil {
		t.Fatalf("window should be nil as provided")
	}
}

func TestDragged_StartsAndMoves(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	win := a.NewWindow("test")
	defer win.Close()

	oldGetWin := getWindowPosition
	oldGetCursor := getCursorPosition
	oldMove := moveWindowTo
	t.Cleanup(func() {
		getWindowPosition = oldGetWin
		getCursorPosition = oldGetCursor
		moveWindowTo = oldMove
	})

	getWindowPosition = func(fyne.Window) (fyne.Position, bool) {
		return fyne.NewPos(100, 200), true
	}
	cursorCalls := 0
	getCursorPosition = func() (fyne.Position, bool) {
		cursorCalls++
		if cursorCalls == 1 {
			return fyne.NewPos(10, 20), true
		}
		return fyne.NewPos(25, 45), true
	}
	var moved fyne.Position
	moveWindowTo = func(_ fyne.Window, x, y float32) bool {
		moved = fyne.NewPos(x, y)
		return true
	}

	var dragChanged []bool
	var movedCb fyne.Position
	w := &Widget{
		window:        win,
		isEditMode:    func() bool { return true },
		onDragChanged: func(v bool) { dragChanged = append(dragChanged, v) },
		onMoved:       func(p fyne.Position) { movedCb = p },
	}

	w.Dragged(nil)

	if !w.dragging {
		t.Fatalf("dragging should be true after first drag")
	}
	if len(dragChanged) != 1 || !dragChanged[0] {
		t.Fatalf("onDragChanged should be called once with true, got %v", dragChanged)
	}
	if moved.X != 115 || moved.Y != 225 {
		t.Fatalf("moveWindowTo target = (%v,%v), want (115,225)", moved.X, moved.Y)
	}
	if movedCb.X != 115 || movedCb.Y != 225 {
		t.Fatalf("onMoved target = (%v,%v), want (115,225)", movedCb.X, movedCb.Y)
	}
}

func TestDragged_NonEditModeEndsDrag(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	win := a.NewWindow("test")
	defer win.Close()

	var changed []bool
	w := &Widget{
		window:        win,
		dragging:      true,
		isEditMode:    func() bool { return false },
		onDragChanged: func(v bool) { changed = append(changed, v) },
	}

	w.Dragged(nil)

	if w.dragging {
		t.Fatalf("dragging should be reset")
	}
	if len(changed) != 1 || changed[0] != false {
		t.Fatalf("onDragChanged should emit false once, got %v", changed)
	}
}

func TestDragged_NilWindowNoop(t *testing.T) {
	t.Parallel()

	w := &Widget{
		window: nil,
	}
	w.Dragged(nil)
}
