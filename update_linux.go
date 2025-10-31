//go:build linux
// +build linux

package main

import (
	"os"
	"os/exec"
)

// setUpdateProcessAttributes sets Linux-specific process attributes for updates
func setUpdateProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setFallbackUpdateProcessAttributes sets Linux-specific process attributes for fallback
func setFallbackUpdateProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// setRestartUpdateProcessAttributes sets Linux-specific process attributes for restart
func setRestartUpdateProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// removeQuarantineAttribute removes quarantine attribute from downloaded files
// Linux doesn't have quarantine attributes, so this is a no-op
func removeQuarantineAttribute(filePath string) error {
	return nil
}
