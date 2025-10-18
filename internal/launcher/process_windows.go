//go:build windows
// +build windows

package launcher

import (
	"fmt"
)

// KillUnixProcess is not available on Windows
func KillUnixProcess(pid int) error {
	return fmt.Errorf("Unix process termination not available on Windows")
}

// IsUnixProcessRunning is not available on Windows
func IsUnixProcessRunning(pid int) bool {
	return false
}