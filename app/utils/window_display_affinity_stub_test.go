//go:build !windows

package utils

import "testing"

func TestNewWindowDisplayAffinityStub(t *testing.T) {
	t.Parallel()

	c := NewWindowDisplayAffinity(nil)
	if c == nil {
		t.Fatalf("NewWindowDisplayAffinity should return non-nil controller")
	}
	if c.window != nil {
		t.Fatalf("stub controller should store provided window value")
	}
}

func TestWindowDisplayAffinitySetExcludeFromCaptureStub(t *testing.T) {
	t.Parallel()

	c := NewWindowDisplayAffinity(nil)
	if c.SetExcludeFromCapture(true) {
		t.Fatalf("stub SetExcludeFromCapture(true) should return false")
	}
	if c.SetExcludeFromCapture(false) {
		t.Fatalf("stub SetExcludeFromCapture(false) should return false")
	}
}
