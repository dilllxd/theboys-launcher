//go:build linux
// +build linux

package main

import (
	"os"
	"os/exec"
)

// setMultiMCProcessAttributes sets Linux-specific process attributes for MultiMC
func setMultiMCProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}