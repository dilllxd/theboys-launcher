//go:build unix
// +build unix

package launcher

import (
	"syscall"
)

// KillUnixProcess forcefully kills a Unix process using SIGKILL
func KillUnixProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

// IsUnixProcessRunning checks if a Unix process is running using signal 0
func IsUnixProcessRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}