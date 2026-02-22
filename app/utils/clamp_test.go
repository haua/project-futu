package utils

import "testing"

func TestClampFloat32(t *testing.T) {
	t.Parallel()

	if got := ClampFloat32(-1, 0, 10); got != 0 {
		t.Fatalf("ClampFloat32(-1,0,10) = %v, want 0", got)
	}
	if got := ClampFloat32(3, 0, 10); got != 3 {
		t.Fatalf("ClampFloat32(3,0,10) = %v, want 3", got)
	}
	if got := ClampFloat32(99, 0, 10); got != 10 {
		t.Fatalf("ClampFloat32(99,0,10) = %v, want 10", got)
	}
}

func TestClampFloat64(t *testing.T) {
	t.Parallel()

	if got := ClampFloat64(-1, 0, 1); got != 0 {
		t.Fatalf("ClampFloat64(-1,0,1) = %v, want 0", got)
	}
	if got := ClampFloat64(0.6, 0, 1); got != 0.6 {
		t.Fatalf("ClampFloat64(0.6,0,1) = %v, want 0.6", got)
	}
	if got := ClampFloat64(2, 0, 1); got != 1 {
		t.Fatalf("ClampFloat64(2,0,1) = %v, want 1", got)
	}
}
