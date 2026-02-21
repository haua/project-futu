package utils

import "testing"

func TestWindowTopMostSet_NilReceiver(t *testing.T) {
	t.Parallel()

	var ctl *WindowTopMost
	if ctl.Set(true) {
		t.Fatalf("nil controller Set should return false")
	}
}

func TestWindowTopMostSet_NilWindow(t *testing.T) {
	t.Parallel()

	ctl := &WindowTopMost{}
	if ctl.Set(true) {
		t.Fatalf("controller with nil window should return false")
	}
}
