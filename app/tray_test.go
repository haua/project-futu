package app

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
)

func TestTopMostMenuLabel(t *testing.T) {
	t.Parallel()

	if got := topMostMenuLabel(true); got != "置顶：开" {
		t.Fatalf("topMostMenuLabel(true) = %q", got)
	}
	if got := topMostMenuLabel(false); got != "置顶：关" {
		t.Fatalf("topMostMenuLabel(false) = %q", got)
	}
}

func TestImageFileFilters(t *testing.T) {
	t.Parallel()

	name, allow := imageFileFilters()
	if name != "png,jpeg,jpg,gif,webp" {
		t.Fatalf("filter name = %q", name)
	}
	want := []string{"png", "jpeg", "jpg", "gif", "webp"}
	if len(allow) != len(want) {
		t.Fatalf("allow length = %d, want %d", len(allow), len(want))
	}
	for i := range want {
		if allow[i] != want[i] {
			t.Fatalf("allow[%d] = %q, want %q", i, allow[i], want[i])
		}
	}
}

func TestTrayIconName(t *testing.T) {
	t.Parallel()

	if got := trayIconName(false); got != "icon.png" {
		t.Fatalf("trayIconName(false) = %q", got)
	}
	if got := trayIconName(true); got != "icon-edit.png" {
		t.Fatalf("trayIconName(true) = %q", got)
	}
}

func TestDetectDoubleTap(t *testing.T) {
	t.Parallel()

	now := time.Now()
	delay := 300 * time.Millisecond

	double, next := detectDoubleTap(time.Time{}, now, delay)
	if double {
		t.Fatalf("first tap should not be double tap")
	}
	if !next.Equal(now) {
		t.Fatalf("first tap next time should be now")
	}

	double, next = detectDoubleTap(now.Add(-100*time.Millisecond), now, delay)
	if !double {
		t.Fatalf("tap within delay should be double tap")
	}
	if !next.IsZero() {
		t.Fatalf("double tap should reset last tap time")
	}

	double, next = detectDoubleTap(now.Add(-delay-time.Millisecond), now, delay)
	if double {
		t.Fatalf("tap after delay should not be double tap")
	}
	if !next.Equal(now) {
		t.Fatalf("next last tap should be now")
	}
}

func TestSetTrayIconByState(t *testing.T) {
	oldLoader := loadAssetResource
	t.Cleanup(func() {
		loadAssetResource = oldLoader
	})

	t.Run("sets icon when resource exists", func(t *testing.T) {
		loadAssetResource = func(name string) fyne.Resource {
			return fyne.NewStaticResource(name, []byte{1})
		}

		var gotName string
		setTrayIconByState(func(r fyne.Resource) {
			gotName = r.Name()
		}, true)

		if gotName != "icon-edit.png" {
			t.Fatalf("set icon name = %q, want icon-edit.png", gotName)
		}
	})

	t.Run("does not set icon when resource missing", func(t *testing.T) {
		loadAssetResource = func(string) fyne.Resource { return nil }

		called := false
		setTrayIconByState(func(fyne.Resource) {
			called = true
		}, false)

		if called {
			t.Fatalf("setIcon should not be called when resource is nil")
		}
	})
}
