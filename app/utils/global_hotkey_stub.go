//go:build !windows

package utils

type GlobalHotkey struct{}

func NewGlobalHotkey() *GlobalHotkey {
	return &GlobalHotkey{}
}

func (h *GlobalHotkey) Supported() bool {
	return false
}

func (h *GlobalHotkey) Register(_ uint32, _ uint32, _ func()) bool {
	return false
}

func (h *GlobalHotkey) Unregister() {}
