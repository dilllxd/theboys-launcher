//go:build darwin
// +build darwin

package main

import (
	"os/exec"
)

// setMultiMCProcessAttributes sets macOS-specific process attributes for MultiMC
func setMultiMCProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}