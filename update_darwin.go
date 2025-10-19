//go:build darwin
// +build darwin

package main

import (
	"os"
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

// removeQuarantineAttribute removes macOS quarantine attribute from downloaded files
func removeQuarantineAttribute(filePath string) error {
	// Use xattr to remove the quarantine attribute on macOS
	cmd := exec.Command("xattr", "-d", "com.apple.quarantine", filePath)
	if err := cmd.Run(); err != nil {
		// Don't fail if xattr is not available or attribute doesn't exist
		// This is common on systems without xattr or if file has no quarantine
		return nil
	}
	return nil
}
