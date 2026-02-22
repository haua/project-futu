package app

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

const (
	modeToggleHotkeyKey     = "hotkey.mode_toggle"
	defaultModeToggleHotkey = "Ctrl+Alt+M"
)

type parsedHotkey struct {
	label string
	mod   uint32
	key   uint32
}

func normalizeToken(token string) string {
	return strings.ToUpper(strings.TrimSpace(token))
}

func keyTokenToVK(token string) (uint32, bool) {
	t := normalizeToken(token)
	if len(t) == 1 {
		ch := t[0]
		if ch >= 'A' && ch <= 'Z' {
			return uint32(ch), true
		}
		if ch >= '0' && ch <= '9' {
			return uint32(ch), true
		}
	}

	switch t {
	case "TAB":
		return 0x09, true
	case "ENTER", "RETURN":
		return 0x0D, true
	case "ESC", "ESCAPE":
		return 0x1B, true
	case "SPACE":
		return 0x20, true
	case "LEFT":
		return 0x25, true
	case "UP":
		return 0x26, true
	case "RIGHT":
		return 0x27, true
	case "DOWN":
		return 0x28, true
	case "HOME":
		return 0x24, true
	case "END":
		return 0x23, true
	case "PAGEUP", "PGUP":
		return 0x21, true
	case "PAGEDOWN", "PGDN":
		return 0x22, true
	case "INSERT", "INS":
		return 0x2D, true
	case "DELETE", "DEL":
		return 0x2E, true
	}

	if strings.HasPrefix(t, "F") && len(t) <= 3 {
		var n int
		_, err := fmt.Sscanf(t, "F%d", &n)
		if err == nil && n >= 1 && n <= 24 {
			return uint32(0x70 + n - 1), true
		}
	}

	return 0, false
}

func parseModeToggleHotkey(label string) (parsedHotkey, bool) {
	parts := strings.Split(label, "+")
	if len(parts) < 2 {
		return parsedHotkey{}, false
	}

	mod := uint32(0)
	hasCtrl := false
	hasAlt := false
	hasShift := false
	keyToken := ""

	for _, raw := range parts {
		token := normalizeToken(raw)
		if token == "" {
			return parsedHotkey{}, false
		}

		switch token {
		case "CTRL", "CONTROL":
			if hasCtrl {
				return parsedHotkey{}, false
			}
			hasCtrl = true
			mod |= 0x0002
		case "ALT":
			if hasAlt {
				return parsedHotkey{}, false
			}
			hasAlt = true
			mod |= 0x0001
		case "SHIFT":
			if hasShift {
				return parsedHotkey{}, false
			}
			hasShift = true
			mod |= 0x0004
		default:
			if keyToken != "" {
				return parsedHotkey{}, false
			}
			keyToken = token
		}
	}

	if mod == 0 || keyToken == "" {
		return parsedHotkey{}, false
	}

	key, ok := keyTokenToVK(keyToken)
	if !ok {
		return parsedHotkey{}, false
	}

	mods := make([]string, 0, 3)
	if hasCtrl {
		mods = append(mods, "Ctrl")
	}
	if hasAlt {
		mods = append(mods, "Alt")
	}
	if hasShift {
		mods = append(mods, "Shift")
	}
	sort.SliceStable(mods, func(i, j int) bool {
		order := map[string]int{"Ctrl": 0, "Alt": 1, "Shift": 2}
		return order[mods[i]] < order[mods[j]]
	})

	canonicalKey := strings.ToUpper(keyToken)
	labelOut := strings.Join(append(mods, canonicalKey), "+")
	return parsedHotkey{label: labelOut, mod: mod, key: key}, true
}

func modeToggleHotkeyHint() string {
	return "点击“开始录制”后按下组合键（需包含 Ctrl/Alt/Shift）"
}

func isModeToggleHotkeyValid(label string) bool {
	_, ok := parseModeToggleHotkey(label)
	return ok
}

func normalizedModeToggleHotkeyOrDefault(label string) string {
	parsed, ok := parseModeToggleHotkey(label)
	if !ok {
		return defaultModeToggleHotkey
	}
	return parsed.label
}

func keyNameToHotkeyToken(name fyne.KeyName) (string, bool) {
	switch name {
	case fyne.KeyTab:
		return "Tab", true
	case fyne.KeyReturn, fyne.KeyEnter:
		return "Enter", true
	case fyne.KeyEscape:
		return "Esc", true
	case fyne.KeySpace:
		return "Space", true
	case fyne.KeyLeft:
		return "Left", true
	case fyne.KeyUp:
		return "Up", true
	case fyne.KeyRight:
		return "Right", true
	case fyne.KeyDown:
		return "Down", true
	case fyne.KeyHome:
		return "Home", true
	case fyne.KeyEnd:
		return "End", true
	case fyne.KeyPageUp:
		return "PageUp", true
	case fyne.KeyPageDown:
		return "PageDown", true
	case fyne.KeyInsert:
		return "Insert", true
	case fyne.KeyDelete:
		return "Delete", true
	}

	key := string(name)
	if key == "" {
		return "", false
	}

	upper := strings.ToUpper(key)
	if len(upper) == 1 {
		ch := upper[0]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			return upper, true
		}
	}
	if strings.HasPrefix(upper, "F") {
		if _, ok := keyTokenToVK(upper); ok {
			return upper, true
		}
	}
	return "", false
}

func modeToggleHotkeyFromKeyEvent(name fyne.KeyName, mods fyne.KeyModifier) (string, bool) {
	if (mods & fyne.KeyModifierSuper) != 0 {
		return "", false
	}

	token, ok := keyNameToHotkeyToken(name)
	if !ok {
		return "", false
	}

	parts := make([]string, 0, 4)
	if (mods & fyne.KeyModifierControl) != 0 {
		parts = append(parts, "Ctrl")
	}
	if (mods & fyne.KeyModifierAlt) != 0 {
		parts = append(parts, "Alt")
	}
	if (mods & fyne.KeyModifierShift) != 0 {
		parts = append(parts, "Shift")
	}
	if len(parts) == 0 {
		return "", false
	}
	parts = append(parts, token)
	return normalizedModeToggleHotkeyOrDefault(strings.Join(parts, "+")), true
}

func supportedModeToggleHotkeyKeys() []fyne.KeyName {
	keys := []fyne.KeyName{
		fyne.KeyTab,
		fyne.KeyReturn,
		fyne.KeyEnter,
		fyne.KeyEscape,
		fyne.KeySpace,
		fyne.KeyLeft,
		fyne.KeyUp,
		fyne.KeyRight,
		fyne.KeyDown,
		fyne.KeyHome,
		fyne.KeyEnd,
		fyne.KeyPageUp,
		fyne.KeyPageDown,
		fyne.KeyInsert,
		fyne.KeyDelete,
	}

	for c := '0'; c <= '9'; c++ {
		keys = append(keys, fyne.KeyName(string(c)))
	}
	for c := 'A'; c <= 'Z'; c++ {
		keys = append(keys, fyne.KeyName(string(c)))
	}
	for i := 1; i <= 24; i++ {
		keys = append(keys, fyne.KeyName(fmt.Sprintf("F%d", i)))
	}
	return keys
}

func (f *FloatingWindow) IsGlobalHotkeySupported() bool {
	if f == nil || f.hotkeySupported == nil {
		return false
	}
	return f.hotkeySupported()
}

func (f *FloatingWindow) BeginModeToggleHotkeyCapture() {
	if f == nil {
		return
	}
	f.hotkeyCapturing.Store(true)
	if f.hotkeyUnregister != nil {
		f.hotkeyUnregister()
	}
}

func (f *FloatingWindow) EndModeToggleHotkeyCapture() {
	if f == nil {
		return
	}
	f.hotkeyCapturing.Store(false)
	_ = f.applyModeToggleHotkey()
}

func (f *FloatingWindow) ModeToggleHotkey() string {
	if f == nil {
		return defaultModeToggleHotkey
	}

	f.hotkeyMu.Lock()
	defer f.hotkeyMu.Unlock()
	if f.modeHotkey == "" {
		return defaultModeToggleHotkey
	}
	return f.modeHotkey
}

func (f *FloatingWindow) SetModeToggleHotkey(label string) bool {
	if f == nil {
		return false
	}

	parsed, ok := parseModeToggleHotkey(label)
	if !ok {
		return false
	}

	f.hotkeyMu.Lock()
	old := f.modeHotkey
	f.modeHotkey = parsed.label
	f.hotkeyMu.Unlock()

	if !f.applyModeToggleHotkey() {
		f.hotkeyMu.Lock()
		f.modeHotkey = old
		f.hotkeyMu.Unlock()
		_ = f.applyModeToggleHotkey()
		return false
	}

	f.saveModeToggleHotkey()
	return true
}

func (f *FloatingWindow) saveModeToggleHotkey() {
	if f == nil || f.App == nil {
		return
	}
	f.App.Preferences().SetString(modeToggleHotkeyKey, f.ModeToggleHotkey())
}

func (f *FloatingWindow) restoreModeToggleHotkey() {
	if f == nil {
		return
	}

	hotkey := defaultModeToggleHotkey
	if f.App != nil {
		if saved := strings.TrimSpace(f.App.Preferences().String(modeToggleHotkeyKey)); saved != "" {
			if parsed, ok := parseModeToggleHotkey(saved); ok {
				hotkey = parsed.label
			}
		}
	}

	f.hotkeyMu.Lock()
	f.modeHotkey = hotkey
	f.hotkeyMu.Unlock()
}

func (f *FloatingWindow) applyModeToggleHotkey() bool {
	if f == nil {
		return false
	}
	if !f.IsGlobalHotkeySupported() {
		return true
	}
	if f.hotkeyRegister == nil {
		return false
	}

	choice, ok := parseModeToggleHotkey(f.ModeToggleHotkey())
	if !ok {
		choice, _ = parseModeToggleHotkey(defaultModeToggleHotkey)
	}
	return f.hotkeyRegister(choice.mod, choice.key, f.onModeToggleHotkeyTriggered)
}

func (f *FloatingWindow) onModeToggleHotkeyTriggered() {
	if f == nil {
		return
	}
	if f.hotkeyCapturing.Load() {
		return
	}
	fyne.Do(func() {
		next := f.ToggleEditMode()
		desk, ok := f.App.(desktop.App)
		if !ok {
			return
		}
		SetTrayIcon(desk, next)
	})
}

func (f *FloatingWindow) Shutdown() {
	if f == nil {
		return
	}
	if f.hotkeyUnregister != nil {
		f.hotkeyUnregister()
	}
}
