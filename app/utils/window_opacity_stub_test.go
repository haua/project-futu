//go:build !windows

package utils

import "testing"

func TestNewWindowOpacityStub(t *testing.T) {
	t.Parallel()

	c := NewWindowOpacity(nil)
	if c == nil {
		t.Fatalf("NewWindowOpacity should return non-nil controller")
	}
	if c.window != nil {
		t.Fatalf("stub controller should store provided window value")
	}
}

func TestWindowOpacitySetStub(t *testing.T) {
	t.Parallel()

	c := NewWindowOpacity(nil)
	if c.Set(1) {
		t.Fatalf("stub Set(1) should return false")
	}
	if c.Set(0.5) {
		t.Fatalf("stub Set(0.5) should return false")
	}
	if c.Set(0) {
		t.Fatalf("stub Set(0) should return false")
	}
}
