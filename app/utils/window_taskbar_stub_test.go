//go:build !windows

package utils

import "testing"

func TestNewWindowTaskbarStub(t *testing.T) {
	t.Parallel()

	c := NewWindowTaskbar(nil)
	if c == nil {
		t.Fatalf("NewWindowTaskbar should return non-nil controller")
	}
	if c.window != nil {
		t.Fatalf("stub controller should store provided window value")
	}
}

func TestWindowTaskbarSetVisibleStub(t *testing.T) {
	t.Parallel()

	c := NewWindowTaskbar(nil)
	if c.SetVisible(true) {
		t.Fatalf("stub SetVisible(true) should return false")
	}
	if c.SetVisible(false) {
		t.Fatalf("stub SetVisible(false) should return false")
	}
}
