//go:build linux
// +build linux

package main

import (
	"os"
	"os/exec"
)

// setPackwizProcessAttributes sets Linux-specific process attributes for packwiz
func setPackwizProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setPackwizRetryProcessAttributes sets Linux-specific process attributes for packwiz retry
func setPackwizRetryProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}