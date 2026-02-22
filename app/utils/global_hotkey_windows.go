//go:build windows

package utils

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/TheTitanrain/w32"
	"golang.org/x/sys/windows"
)

const (
	wmHotKey       = 0x0312
	wmQuit         = 0x0012
	modNoRepeat    = 0x4000
	globalHotkeyID = 1
)

var (
	user32GlobalHotkey          = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey          = user32GlobalHotkey.NewProc("RegisterHotKey")
	procUnregisterHotKey        = user32GlobalHotkey.NewProc("UnregisterHotKey")
	procGetMessageGlobalHotkey  = user32GlobalHotkey.NewProc("GetMessageW")
	procPostThreadMessageHotkey = user32GlobalHotkey.NewProc("PostThreadMessageW")
)

type GlobalHotkey struct {
	mu       sync.Mutex
	threadID uint32
	done     chan struct{}
}

type hotkeyStartResult struct {
	ok       bool
	threadID uint32
}

func NewGlobalHotkey() *GlobalHotkey {
	return &GlobalHotkey{}
}

func (h *GlobalHotkey) Supported() bool {
	return true
}

func (h *GlobalHotkey) Register(mod uint32, key uint32, onTrigger func()) bool {
	if h == nil || onTrigger == nil || key == 0 {
		return false
	}

	h.Unregister()

	started := make(chan hotkeyStartResult, 1)
	done := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		threadID := windows.GetCurrentThreadId()
		r1, _, _ := procRegisterHotKey.Call(
			0,
			globalHotkeyID,
			uintptr(mod|modNoRepeat),
			uintptr(key),
		)
		if r1 == 0 {
			started <- hotkeyStartResult{ok: false}
			close(done)
			return
		}

		started <- hotkeyStartResult{ok: true, threadID: threadID}

		var msg w32.MSG
		for {
			ret, _, _ := procGetMessageGlobalHotkey.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
			)
			if int32(ret) <= 0 {
				break
			}
			if msg.Message == wmHotKey {
				onTrigger()
			}
		}

		procUnregisterHotKey.Call(0, globalHotkeyID)
		close(done)
	}()

	startResult := <-started
	if !startResult.ok {
		return false
	}

	h.mu.Lock()
	h.threadID = startResult.threadID
	h.done = done
	h.mu.Unlock()
	return true
}

func (h *GlobalHotkey) Unregister() {
	if h == nil {
		return
	}

	h.mu.Lock()
	threadID := h.threadID
	done := h.done
	h.threadID = 0
	h.done = nil
	h.mu.Unlock()
	if done == nil {
		return
	}

	procPostThreadMessageHotkey.Call(uintptr(threadID), wmQuit, 0, 0)
	<-done
}
