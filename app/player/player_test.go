package player

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
)

func TestClampFloat32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v    float32
		min  float32
		max  float32
		want float32
	}{
		{name: "below min", v: -1, min: 0, max: 10, want: 0},
		{name: "within range", v: 3, min: 0, max: 10, want: 3},
		{name: "above max", v: 99, min: 0, max: 10, want: 10},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := clampFloat32(tc.v, tc.min, tc.max)
			if got != tc.want {
				t.Fatalf("clampFloat32(%v, %v, %v) = %v, want %v", tc.v, tc.min, tc.max, got, tc.want)
			}
		})
	}
}

func TestClampPosition(t *testing.T) {
	t.Parallel()

	bounds := fyne.NewSize(100, 50)
	got := clampPosition(fyne.NewPos(-5, 80), bounds)

	if got.X != 0 || got.Y != 50 {
		t.Fatalf("clampPosition result = (%v, %v), want (0, 50)", got.X, got.Y)
	}
}

func TestToScreenPixels(t *testing.T) {
	t.Parallel()

	if got := toScreenPixels(10.2, 1.5); got != 15 {
		t.Fatalf("toScreenPixels(10.2, 1.5) = %d, want 15", got)
	}
}

func TestMaxFloat32(t *testing.T) {
	t.Parallel()

	if got := maxFloat32(2, 3); got != 3 {
		t.Fatalf("maxFloat32(2,3) = %v, want 3", got)
	}
	if got := maxFloat32(4, 1); got != 4 {
		t.Fatalf("maxFloat32(4,1) = %v, want 4", got)
	}
}

func TestScaledSizeForZoom(t *testing.T) {
	t.Parallel()

	p := &Player{baseSize: fyne.NewSize(3, 5)}
	got := p.scaledSizeForZoom(0.2)

	if got.Width != 1 || got.Height != 1 {
		t.Fatalf("scaledSizeForZoom(0.2) = (%v, %v), want (1, 1)", got.Width, got.Height)
	}
}

func TestPlaybackState(t *testing.T) {
	t.Parallel()

	p := &Player{}
	id1 := p.beginPlayback()
	id2 := p.beginPlayback()

	if id1 != 1 || id2 != 2 {
		t.Fatalf("unexpected playback ids: id1=%d id2=%d", id1, id2)
	}
	if p.isPlaybackActive(id1) {
		t.Fatalf("old playback id should be inactive")
	}
	if !p.isPlaybackActive(id2) {
		t.Fatalf("latest playback id should be active")
	}
}

func TestRenderPausedState(t *testing.T) {
	t.Parallel()

	p := &Player{pauseSignal: make(chan struct{}, 1)}
	if p.RenderPaused() {
		t.Fatalf("zero value should be not paused")
	}
	p.SetRenderPaused(true)
	if !p.RenderPaused() {
		t.Fatalf("expected paused state")
	}
	p.SetRenderPaused(false)
	if p.RenderPaused() {
		t.Fatalf("expected resumed state")
	}
}

func TestWaitRenderResumed_UnblocksOnResume(t *testing.T) {
	t.Parallel()

	p := &Player{pauseSignal: make(chan struct{}, 1)}
	id := p.beginPlayback()
	p.SetRenderPaused(true)

	done := make(chan bool, 1)
	go func() {
		done <- p.waitRenderResumed(id)
	}()

	select {
	case <-done:
		t.Fatalf("waitRenderResumed should block while paused")
	case <-time.After(30 * time.Millisecond):
	}

	p.SetRenderPaused(false)

	select {
	case ok := <-done:
		if !ok {
			t.Fatalf("waitRenderResumed should return true when resumed and playback active")
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("waitRenderResumed did not unblock after resume")
	}
}

func TestWaitRenderResumed_ReturnsFalseWhenPlaybackInactive(t *testing.T) {
	t.Parallel()

	p := &Player{pauseSignal: make(chan struct{}, 1)}
	id := p.beginPlayback()
	p.SetRenderPaused(true)

	go func() {
		time.Sleep(20 * time.Millisecond)
		p.beginPlayback()
	}()

	if ok := p.waitRenderResumed(id); ok {
		t.Fatalf("waitRenderResumed should return false when playback becomes inactive")
	}
}

func TestRenderPause_MultipleReasons(t *testing.T) {
	t.Parallel()

	p := &Player{pauseSignal: make(chan struct{}, 1)}
	if p.RenderPaused() {
		t.Fatalf("zero value should be not paused")
	}

	p.SetRenderPaused(true)
	if !p.RenderPaused() {
		t.Fatalf("drag pause should pause rendering")
	}

	p.SetFullyTransparentPaused(true)
	p.SetRenderPaused(false)
	if !p.RenderPaused() {
		t.Fatalf("transparent pause should keep rendering paused")
	}

	p.SetFullyTransparentPaused(false)
	if p.RenderPaused() {
		t.Fatalf("rendering should resume after all pause reasons are cleared")
	}
}

func TestPlay_EmptyOrMissingPathDoesNotPersist(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()

	p := &Player{app: a}

	p.Play("   ")
	if got := a.Preferences().String(lastImagePathKey); got != "" {
		t.Fatalf("empty path should not be persisted, got %q", got)
	}

	missing := filepath.Join(t.TempDir(), "no-such-file.png")
	p.Play(missing)
	if got := a.Preferences().String(lastImagePathKey); got != "" {
		t.Fatalf("missing path should not be persisted, got %q", got)
	}
}

func TestPlay_ExistingUnsupportedFileStillPersistsPath(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()

	p := &Player{app: a}
	path := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	p.Play(path)

	if got := a.Preferences().String(lastImagePathKey); got != path {
		t.Fatalf("persisted path = %q, want %q", got, path)
	}
	if !p.isPlaybackActive(1) {
		t.Fatalf("playback id should have advanced to 1")
	}
}

func TestPlayLast_MissingFileClearsPreference(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()

	p := &Player{app: a}
	path := filepath.Join(t.TempDir(), "gone.png")
	a.Preferences().SetString(lastImagePathKey, path)

	p.PlayLast()

	if got := a.Preferences().String(lastImagePathKey); got != "" {
		t.Fatalf("missing remembered file should be cleared, got %q", got)
	}
}

func TestPlayLast_ExistingFileKeepsPreference(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()

	p := &Player{app: a}
	path := filepath.Join(t.TempDir(), "remember.txt")
	if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	a.Preferences().SetString(lastImagePathKey, path)

	p.PlayLast()

	if got := a.Preferences().String(lastImagePathKey); got != path {
		t.Fatalf("existing remembered file should remain, got %q want %q", got, path)
	}
}

func TestNewPlayer_UsesSavedCanvasWidthForInitialZoom(t *testing.T) {
	tests := []struct {
		name       string
		savedWidth float64
		wantWidth  float32
	}{
		{name: "normal saved width", savedWidth: 100, wantWidth: 100},
		{name: "clamped to min zoom", savedWidth: 1, wantWidth: 40},
		{name: "clamped to max zoom", savedWidth: 5000, wantWidth: 400},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a := fynetest.NewApp()
			defer a.Quit()
			w := a.NewWindow("test")
			defer w.Close()

			a.Preferences().SetFloat(lastCanvasWidthKey, tc.savedWidth)
			p := NewPlayer(a, w)

			if got := p.Canvas.Size().Width; got != tc.wantWidth {
				t.Fatalf("initial canvas width = %v, want %v", got, tc.wantWidth)
			}
		})
	}
}

func TestAdjustScaleByScroll_ZoomAndClamp(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	originWidth := p.Canvas.Size().Width

	p.AdjustScaleByScroll(&fyne.ScrollEvent{
		PointEvent: fyne.PointEvent{
			AbsolutePosition: fyne.NewPos(100, 100),
		},
		Scrolled: fyne.Delta{DY: 1},
	})
	if got := p.Canvas.Size().Width; got <= originWidth {
		t.Fatalf("expected zoom in width > %v, got %v", originWidth, got)
	}

	for i := 0; i < 200; i++ {
		p.AdjustScaleByScroll(&fyne.ScrollEvent{
			PointEvent: fyne.PointEvent{
				AbsolutePosition: fyne.NewPos(100, 100),
			},
			Scrolled: fyne.Delta{DY: -1},
		})
	}
	if got := p.Canvas.Size().Width; got != 40 {
		t.Fatalf("zoom out should clamp to min width 40, got %v", got)
	}
}

func TestAdjustScaleByScroll_NilOrZeroNoop(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	before := p.Canvas.Size()

	p.AdjustScaleByScroll(nil)
	p.AdjustScaleByScroll(&fyne.ScrollEvent{
		Scrolled: fyne.Delta{},
	})
	after := p.Canvas.Size()

	if before != after {
		t.Fatalf("size should not change on nil/zero scroll, before=%v after=%v", before, after)
	}
}

func TestAdjustScaleByScroll_UsesDXWhenDYZero(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	before := p.Canvas.Size().Width

	p.AdjustScaleByScroll(&fyne.ScrollEvent{
		PointEvent: fyne.PointEvent{
			AbsolutePosition: fyne.NewPos(100, 100),
		},
		Scrolled: fyne.Delta{DX: 1, DY: 0},
	})

	after := p.Canvas.Size().Width
	if after <= before {
		t.Fatalf("DX scroll should zoom in, before=%v after=%v", before, after)
	}
}

func TestPlayImage_LoadsAndAppliesImage(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	imgPath := filepath.Join(t.TempDir(), "sample.png")
	file, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("create image file: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 3, 2))
	src.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(file, src); err != nil {
		_ = file.Close()
		t.Fatalf("encode png: %v", err)
	}
	_ = file.Close()

	id := p.beginPlayback()
	PlayImage(p, imgPath, id)
	fyne.DoAndWait(func() {})

	if p.Canvas.Image == nil {
		t.Fatalf("canvas image should be set")
	}
	if p.baseSize.Width != 3 || p.baseSize.Height != 2 {
		t.Fatalf("base size = (%v,%v), want (3,2)", p.baseSize.Width, p.baseSize.Height)
	}
}

func TestPlayImage_IgnoresInactivePlayback(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	imgPath := filepath.Join(t.TempDir(), "sample.png")
	file, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("create image file: %v", err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		_ = file.Close()
		t.Fatalf("encode png: %v", err)
	}
	_ = file.Close()

	currentID := p.beginPlayback()
	PlayImage(p, imgPath, currentID-1)
	fyne.DoAndWait(func() {})

	if p.Canvas.Image != nil {
		t.Fatalf("canvas image should stay nil for inactive playback")
	}
}

func TestUpdateBaseSize_InvalidInputNoop(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	oldBase := p.baseSize
	oldCanvas := p.Canvas.Size()

	p.updateBaseSize(0, 10)
	p.updateBaseSize(10, 0)
	p.updateBaseSize(-1, -1)

	if p.baseSize != oldBase {
		t.Fatalf("base size should not change for invalid input")
	}
	if p.Canvas.Size() != oldCanvas {
		t.Fatalf("canvas size should not change for invalid input")
	}
}

func TestUpdateBaseSize_UsesRememberedCanvasWidth(t *testing.T) {
	a := fynetest.NewApp()
	defer a.Quit()
	w := a.NewWindow("test")
	defer w.Close()

	p := NewPlayer(a, w)
	a.Preferences().SetFloat(lastCanvasWidthKey, 300)

	p.updateBaseSize(100, 50)

	if got := p.Canvas.Size().Width; got != 200 {
		t.Fatalf("canvas width = %v, want 200", got)
	}
	if got := p.Canvas.Size().Height; got != 100 {
		t.Fatalf("canvas height = %v, want 100", got)
	}
}
