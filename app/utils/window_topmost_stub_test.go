//go:build !windows

package utils

import "testing"

func TestNewWindowTopMostStub(t *testing.T) {
	t.Parallel()

	c := NewWindowTopMost(nil)
	if c == nil {
		t.Fatalf("NewWindowTopMost should return non-nil controller")
	}
	if c.window != nil {
		t.Fatalf("stub controller should store provided window value")
	}
}

func TestWindowTopMostSetStub(t *testing.T) {
	t.Parallel()

	c := NewWindowTopMost(nil)
	if c.Set(true) {
		t.Fatalf("stub Set(true) should return false")
	}
	if c.Set(false) {
		t.Fatalf("stub Set(false) should return false")
	}
}
