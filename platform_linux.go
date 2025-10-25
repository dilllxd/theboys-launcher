//go:build linux
// +build linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Linux memory detection using /proc/meminfo
func totalRAMMB() int {
	// Read memory information from /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		// Fallback to 8GB if we can't read meminfo
		return 8192
	}

	// Parse MemTotal from /proc/meminfo
	// Format: MemTotal:       16384000 kB
	var memTotalKB int64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if parsed, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memTotalKB = parsed
				}
			}
			break
		}
	}

	// Convert KB to MB
	totalMB := int(memTotalKB / 1024)
	return validateMemoryResult(totalMB)
}

// Linux-specific directory paths
func getLauncherHome() string {
	// Linux: ~/.theboyslauncher
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// Fallback to current directory if HOME is not set
		if exePath, err := os.Executable(); err == nil {
			return filepath.Dir(exePath)
		}
		return "."
	}
	return filepath.Join(homeDir, ".theboyslauncher")
}

// Linux-specific process creation attributes
func setProcessAttributes(cmd *exec.Cmd) {
	// Linux doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// Linux-specific user environment
func getCurrentUser() string {
	return os.Getenv("USER")
}

// Linux-specific app data directory
func getAppDataDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "/tmp"
	}
	return filepath.Join(homeDir, ".local", "share")
}

// Linux-specific executable extensions (no extension)
func getExecutableExtension() string {
	return ""
}

// Linux-specific file permissions
func setExecutablePermissions(path string) error {
	// Set executable permissions on Unix-like systems
	return os.Chmod(path, 0755)
}

// Linux architecture detection
func getArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64"
	case "arm":
		return "arm"
	default:
		return runtime.GOARCH
	}
}

// Check if path is an app bundle (Linux doesn't use app bundles)
func isAppBundle(path string) bool {
	return false
}
