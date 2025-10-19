//go:build windows
// +build windows

package main

import (
	"golang.org/x/sys/windows"
	"os/exec"
)

// setMultiMCProcessAttributes sets Windows-specific process attributes for MultiMC
func setMultiMCProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}
