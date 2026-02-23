package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	fynetest "fyne.io/fyne/v2/test"
)

func TestNormalizeImageSourceMode(t *testing.T) {
	t.Parallel()

	if got := normalizeImageSourceMode("folder"); got != imageSourceModeFolder {
		t.Fatalf("normalizeImageSourceMode(folder) = %q", got)
	}
	if got := normalizeImageSourceMode("FOLDER"); got != imageSourceModeFolder {
		t.Fatalf("normalizeImageSourceMode(FOLDER) = %q", got)
	}
	if got := normalizeImageSourceMode("anything"); got != imageSourceModeSingle {
		t.Fatalf("normalizeImageSourceMode(anything) = %q", got)
	}
}

func TestListSupportedImageFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	makeFile := func(name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write file %q: %v", name, err)
		}
	}

	makeFile("a.png")
	makeFile("b.JPG")
	makeFile("c.gif")
	makeFile("d.txt")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	got, err := listSupportedImageFiles(dir)
	if err != nil {
		t.Fatalf("listSupportedImageFiles: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("supported files = %d, want 3", len(got))
	}
}

func TestPickRandomImagePath_AvoidsImmediateRepeat(t *testing.T) {
	t.Parallel()

	candidates := []string{"a.png", "b.png", "c.png"}
	got := pickRandomImagePath(candidates, "b.png", func(_ int) int { return 1 })
	if got == "b.png" {
		t.Fatalf("pick should avoid immediate repeat when possible")
	}
}

func TestSetFixedImage(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	img := filepath.Join(t.TempDir(), "fixed.png")
	if err := os.WriteFile(img, []byte("x"), 0o600); err != nil {
		t.Fatalf("write fixed image: %v", err)
	}

	fw := &FloatingWindow{App: a}
	if ok := fw.SetFixedImage(img); !ok {
		t.Fatalf("SetFixedImage should succeed")
	}
	if got := fw.ImageSourceMode(); got != imageSourceModeSingle {
		t.Fatalf("ImageSourceMode() = %q, want single", got)
	}
	if got := fw.FixedImagePath(); got != img {
		t.Fatalf("FixedImagePath() = %q, want %q", got, img)
	}
	if got := a.Preferences().String(fixedImagePathPrefKey); got != img {
		t.Fatalf("saved fixed image path = %q, want %q", got, img)
	}
}

func TestSetRandomImageFolder(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.png"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write random image: %v", err)
	}

	fw := &FloatingWindow{
		App:                 a,
		imageTickerInterval: 24 * time.Hour,
	}
	t.Cleanup(fw.stopImageTicker)

	if ok := fw.SetRandomImageFolder(dir); !ok {
		t.Fatalf("SetRandomImageFolder should succeed")
	}
	if got := fw.ImageSourceMode(); got != imageSourceModeFolder {
		t.Fatalf("ImageSourceMode() = %q, want folder", got)
	}
	if got := fw.RandomFolderPath(); got != dir {
		t.Fatalf("RandomFolderPath() = %q, want %q", got, dir)
	}
	if got := a.Preferences().String(randomFolderPathPrefKey); got != dir {
		t.Fatalf("saved random folder path = %q, want %q", got, dir)
	}
}

func TestRestoreImageSource_FallbackFromLastImage(t *testing.T) {
	t.Parallel()

	a := fynetest.NewApp()
	t.Cleanup(a.Quit)
	a.Preferences().SetString("player.last_image_path", "C:/tmp/legacy.png")

	fw := &FloatingWindow{App: a}
	fw.restoreImageSource()

	if got := fw.FixedImagePath(); got != "C:/tmp/legacy.png" {
		t.Fatalf("FixedImagePath() = %q, want legacy path", got)
	}
}
