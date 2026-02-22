//go:build !windows

package utils

import "errors"

var errLaunchAtStartupUnsupported = errors.New("launch at startup is only supported on windows")

type LaunchAtStartup struct {
	valueName string
}

func NewLaunchAtStartup(valueName string) *LaunchAtStartup {
	return &LaunchAtStartup{valueName: valueName}
}

func (c *LaunchAtStartup) IsEnabled() (bool, error) {
	return false, nil
}

func (c *LaunchAtStartup) SetEnabled(_ bool) error {
	return errLaunchAtStartupUnsupported
}
