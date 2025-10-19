//go:build windows
// +build windows

package main

import (
	"os/exec"
	"golang.org/x/sys/windows"
)

// setUpdateProcessAttributes sets Windows-specific process attributes for updates
func setUpdateProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}

// setFallbackUpdateProcessAttributes sets Windows-specific process attributes for fallback
func setFallbackUpdateProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}

// setRestartUpdateProcessAttributes sets Windows-specific process attributes for restart
func setRestartUpdateProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}