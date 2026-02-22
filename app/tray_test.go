package app

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestTopMostMenuLabel(t *testing.T) {
	t.Parallel()

	if got := topMostMenuLabel(true); got != "\u7f6e\u9876\uff1a\u5f00" {
		t.Fatalf("topMostMenuLabel(true) = %q", got)
	}
	if got := topMostMenuLabel(false); got != "\u7f6e\u9876\uff1a\u5173" {
		t.Fatalf("topMostMenuLabel(false) = %q", got)
	}
}

func TestWindowVisibilityMenuLabel(t *testing.T) {
	t.Parallel()

	if got := windowVisibilityMenuLabel(true); got != "\u7a97\u53e3\uff1a\u663e\u793a" {
		t.Fatalf("windowVisibilityMenuLabel(true) = %q", got)
	}
	if got := windowVisibilityMenuLabel(false); got != "\u7a97\u53e3\uff1a\u9690\u85cf" {
		t.Fatalf("windowVisibilityMenuLabel(false) = %q", got)
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

func TestAppVersionText(t *testing.T) {
	t.Parallel()

	if got := appVersionText(nil); got != "\u7248\u672c\uff1aunknown" {
		t.Fatalf("appVersionText(nil) = %q", got)
	}

	a := test.NewApp()
	t.Cleanup(a.Quit)

	if got := appVersionText(a); got == "" {
		t.Fatalf("appVersionText(a) should not be empty")
	}
}

func TestOperationGuideText(t *testing.T) {
	t.Parallel()

	guide := operationGuideText()
	if guide == "" {
		t.Fatalf("operation guide should not be empty")
	}
	if !strings.HasPrefix(guide, "\u64cd\u4f5c\u6307\u5357") {
		t.Fatalf("operation guide should start with 操作指南, got %q", guide)
	}
}
