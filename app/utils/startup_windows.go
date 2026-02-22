//go:build windows

package utils

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const startupRunKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

type LaunchAtStartup struct {
	valueName      string
	executablePath func() (string, error)
}

func NewLaunchAtStartup(valueName string) *LaunchAtStartup {
	return &LaunchAtStartup{
		valueName:      valueName,
		executablePath: os.Executable,
	}
}

func (c *LaunchAtStartup) IsEnabled() (bool, error) {
	if c == nil || strings.TrimSpace(c.valueName) == "" {
		return false, nil
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, startupRunKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	value, _, err := key.GetStringValue(c.valueName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(value) != "", nil
}

func (c *LaunchAtStartup) SetEnabled(enabled bool) error {
	if c == nil || strings.TrimSpace(c.valueName) == "" {
		return nil
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, startupRunKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if !enabled {
		err = key.DeleteValue(c.valueName)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}

	exePath := ""
	if c.executablePath != nil {
		exePath, err = c.executablePath()
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(exePath) == "" {
		return nil
	}

	return key.SetStringValue(c.valueName, quoteWindowsCommand(exePath))
}

func quoteWindowsCommand(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), `"`)
	if trimmed == "" {
		return ""
	}
	return `"` + trimmed + `"`
}
