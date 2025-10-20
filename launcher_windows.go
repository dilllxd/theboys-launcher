//go:build windows
// +build windows

package main

import (
	"golang.org/x/sys/windows"
	"os/exec"
)

// setPackwizProcessAttributes sets Windows-specific process attributes for packwiz
func setPackwizProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}

// setPackwizRetryProcessAttributes sets Windows-specific process attributes for packwiz retry
func setPackwizRetryProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}
