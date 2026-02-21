package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAssetResource_FromCWDAssets(t *testing.T) {
	tmp := t.TempDir()
	assetsDir := filepath.Join(tmp, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("mkdir assets dir: %v", err)
	}

	fileName := "unit-test-resource.bin"
	want := []byte{1, 2, 3, 4}
	if err := os.WriteFile(filepath.Join(assetsDir, fileName), want, 0o644); err != nil {
		t.Fatalf("write asset file: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	res := LoadAssetResource(fileName)
	if res == nil {
		t.Fatalf("resource should not be nil")
	}
	if res.Name() != fileName {
		t.Fatalf("resource name = %q, want %q", res.Name(), fileName)
	}
	got := res.Content()
	if len(got) != len(want) {
		t.Fatalf("resource length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("resource content mismatch at %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestLoadAssetResource_MissingReturnsNil(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	res := LoadAssetResource("definitely-not-existing.asset")
	if res != nil {
		t.Fatalf("expected nil resource for missing file")
	}
}
