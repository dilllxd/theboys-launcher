//go:build darwin
// +build darwin

package main

import (
	"os/exec"
)

// setUpdateProcessAttributes sets macOS-specific process attributes for updates
func setUpdateProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setFallbackUpdateProcessAttributes sets macOS-specific process attributes for fallback
func setFallbackUpdateProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setRestartUpdateProcessAttributes sets macOS-specific process attributes for restart
func setRestartUpdateProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}