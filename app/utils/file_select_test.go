package utils

import (
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
)

func TestShowFileOpen_UsesHostAndForwardsCallback(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("host")
	defer w.Close()

	oldNewHost := newHostWindow
	oldShowDialog := showFileOpenDialog
	t.Cleanup(func() {
		newHostWindow = oldNewHost
		showFileOpenDialog = oldShowDialog
	})

	newHostWindow = func(fyne.App) fyne.Window {
		return w
	}

	var gotHost fyne.Window
	dialogCalled := false
	showFileOpenDialog = func(callback func(fyne.URIReadCloser, error), parent fyne.Window) {
		dialogCalled = true
		gotHost = parent
		if callback == nil {
			t.Fatalf("callback should be forwarded")
		}
	}

	ShowFileOpen(a, func(fyne.URIReadCloser, error) {})

	if !dialogCalled {
		t.Fatalf("dialog should be called")
	}
	if gotHost != w {
		t.Fatalf("host window should come from newHostWindow")
	}
}
