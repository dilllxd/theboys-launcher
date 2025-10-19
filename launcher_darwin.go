//go:build darwin
// +build darwin

package main

import (
	"os/exec"
)

// setPackwizProcessAttributes sets macOS-specific process attributes for packwiz
func setPackwizProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setPackwizRetryProcessAttributes sets macOS-specific process attributes for packwiz retry
func setPackwizRetryProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}