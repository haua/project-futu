//go:build !windows

package utils

import "testing"

func TestGlobalHotkeyStub(t *testing.T) {
	t.Parallel()

	h := NewGlobalHotkey()
	if h == nil {
		t.Fatalf("NewGlobalHotkey should return non-nil controller")
	}
	if h.Supported() {
		t.Fatalf("stub Supported should be false")
	}
	if h.Register(0x2, 0x4D, func() {}) {
		t.Fatalf("stub Register should return false")
	}
	h.Unregister()
}
