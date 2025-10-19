//go:build darwin
// +build darwin

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// macOS memory detection using sysctl
func totalRAMMB() int {
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		// Fallback to 8GB if sysctl fails
		return 8192
	}

	// Convert bytes to megabytes and validate
	totalMB := int(memInfo.TotalMemory / (1024 * 1024))
	return validateMemoryResult(totalMB)
}

// macOS-specific directory paths
func getLauncherHome() string {
	// macOS: ~/Library/Application Support/TheBoysLauncher
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// Fallback to current directory if HOME is not set
		if exePath, err := os.Executable(); err == nil {
			return filepath.Dir(exePath)
		}
		return "."
	}
	return filepath.Join(homeDir, "Library", "Application Support", "TheBoysLauncher")
}

// macOS-specific process management is now in process_darwin.go

// macOS-specific process creation attributes
func setProcessAttributes(cmd *exec.Cmd) {
	// macOS doesn't need special attributes for GUI apps
	// Ensure proper environment for GUI execution
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
}

// macOS-specific user environment
func getCurrentUser() string {
	return os.Getenv("USER")
}

// macOS-specific app data directory
func getAppDataDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "/tmp"
	}
	return filepath.Join(homeDir, "Library", "Application Support")
}

// macOS-specific executable extensions (no extension)
func getExecutableExtension() string {
	return ""
}

// macOS-specific file permissions
func setExecutablePermissions(path string) error {
	// Set executable permissions on Unix-like systems
	return os.Chmod(path, 0755)
}

// macOS architecture detection
func getArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}

// macOS app bundle structure
func getPrismExecutablePath(installDir string) string {
	// macOS Prism Launcher is in an app bundle
	return filepath.Join(installDir, "PrismLauncher.app", "Contents", "MacOS", "PrismLauncher")
}

// Check if path is an app bundle
func isAppBundle(path string) bool {
	return filepath.Ext(path) == ".app"
}