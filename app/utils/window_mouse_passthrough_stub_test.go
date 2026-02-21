//go:build !windows

package utils

import "testing"

func TestNewWindowMousePassthroughStub(t *testing.T) {
	t.Parallel()

	c := NewWindowMousePassthrough(nil)
	if c == nil {
		t.Fatalf("NewWindowMousePassthrough should return non-nil controller")
	}
	if c.window != nil {
		t.Fatalf("stub controller should store provided window value")
	}
}

func TestWindowMousePassthroughSetEnabledStub(t *testing.T) {
	t.Parallel()

	c := NewWindowMousePassthrough(nil)
	if c.SetEnabled(true) {
		t.Fatalf("stub SetEnabled(true) should return false")
	}
	if c.SetEnabled(false) {
		t.Fatalf("stub SetEnabled(false) should return false")
	}
}
