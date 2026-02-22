//go:build !windows

package utils

import "testing"

func TestNewLaunchAtStartupStub(t *testing.T) {
	t.Parallel()

	c := NewLaunchAtStartup("Futu")
	if c == nil {
		t.Fatalf("NewLaunchAtStartup should return non-nil controller")
	}
	if c.valueName != "Futu" {
		t.Fatalf("valueName = %q, want %q", c.valueName, "Futu")
	}
}

func TestLaunchAtStartupStubIsEnabled(t *testing.T) {
	t.Parallel()

	c := NewLaunchAtStartup("Futu")
	enabled, err := c.IsEnabled()
	if err != nil {
		t.Fatalf("IsEnabled error = %v, want nil", err)
	}
	if enabled {
		t.Fatalf("IsEnabled = true, want false")
	}
}

func TestLaunchAtStartupStubSetEnabled(t *testing.T) {
	t.Parallel()

	c := NewLaunchAtStartup("Futu")
	if err := c.SetEnabled(true); err == nil {
		t.Fatalf("SetEnabled(true) error = nil, want non-nil")
	}
}
