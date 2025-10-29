//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
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

// readInstallationPathFromRegistry reads the installation path from the registry
// Returns the installation path if found and valid, otherwise returns empty string
func readInstallationPathFromRegistry() string {
	// Open the registry key for current user
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\TheBoysLauncher`, registry.READ)
	if err != nil {
		// Registry key doesn't exist or access denied
		return ""
	}
	defer key.Close()

	// Read the InstallPath value
	installPath, _, err := key.GetStringValue("InstallPath")
	if err != nil {
		// InstallPath value doesn't exist
		return ""
	}

	// Validate that the path exists and is accessible
	if installPath == "" {
		return ""
	}

	// Check if the path exists
	if _, err := os.Stat(installPath); err != nil {
		// Path doesn't exist or is not accessible
		return ""
	}

	return installPath
}

// isInstalledMode determines if the launcher is running in installed mode
// Returns true if running from a custom installation location (not LocalAppData)
func isInstalledMode(installPath string) bool {
	if installPath == "" {
		return false
	}

	// Get the default LocalAppData path
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return false
	}

	defaultPath := filepath.Join(localAppData, "TheBoysLauncher")

	// Normalize paths for comparison
	installPath = filepath.Clean(installPath)
	defaultPath = filepath.Clean(defaultPath)

	// If the installation path is different from the default LocalAppData path,
	// we're in installed mode
	return installPath != defaultPath
}

// Windows-specific directory paths
func getLauncherHome() string {
	// First, check the registry for custom installation path
	installPath := readInstallationPathFromRegistry()

	if installPath != "" {
		// If we have a custom installation path from registry
		if isInstalledMode(installPath) {
			// In installed mode, store data alongside the executable (portable-style)
			return installPath
		}
		// If it's the default path, continue with normal logic
	}

	// Default behavior for existing installations or when registry is not available
	// Prefer LocalAppData\TheBoysLauncher on Windows for per-user installs
	// Falls back to %USERPROFILE%\.theboyslauncher for compatibility
	appData := os.Getenv("LOCALAPPDATA")
	if appData != "" {
		return filepath.Join(appData, "TheBoysLauncher")
	}

	// Fallback to USERPROFILE dot-folder (legacy)
	homeDir := os.Getenv("USERPROFILE")
	if homeDir == "" {
		if exePath, err := os.Executable(); err == nil {
			return filepath.Dir(exePath)
		}
		return "."
	}
	return filepath.Join(homeDir, ".theboyslauncher")
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
