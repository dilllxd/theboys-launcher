//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// Windows-specific memory detection is now in memory_windows.go

// Windows memory detection using GlobalMemoryStatusEx API
func totalRAMMB() int {
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		// Fallback to 8GB if API call fails
		return 8192
	}

	// Convert bytes to megabytes and validate
	totalMB := int(memInfo.ullTotalPhys / (1024 * 1024))
	return validateMemoryResult(totalMB)
}

// Windows-specific directory paths
func getLauncherHome() string {
	// Windows: %USERPROFILE%\.theboys-launcher
	homeDir := os.Getenv("USERPROFILE")
	if homeDir == "" {
		// Fallback to current directory if USERPROFILE is not set
		if exePath, err := os.Executable(); err == nil {
			return filepath.Dir(exePath)
		}
		return "."
	}
	return filepath.Join(homeDir, ".theboys-launcher")
}

// Windows-specific process management is now in process_windows.go

// Windows-specific process creation attributes
func setProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}

// Windows-specific user environment
func getCurrentUser() string {
	return os.Getenv("USERNAME")
}

// Windows-specific app data directory
func getAppDataDir() string {
	return os.Getenv("APPDATA")
}

// Windows-specific executable extensions
func getExecutableExtension() string {
	return ".exe"
}

// Windows-specific file permissions (no-op on Windows)
func setExecutablePermissions(path string) error {
	// Windows doesn't have executable permissions in the same way as Unix
	return nil
}