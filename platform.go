package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// Platform-specific constants
const (
	LauncherName = "TheBoys Launcher"
)

// Platform-specific executable names
var (
	LauncherExeName = getLauncherExeName()
	PrismExeName    = getPrismExeName()
	JavaBinName     = getJavaBinName()
	JavawBinName    = getJavawBinName()
)

// Platform detection functions
func getLauncherExeName() string {
	if runtime.GOOS == "windows" {
		return "TheBoysLauncher.exe"
	}
	return "TheBoysLauncher"
}

func getPrismExeName() string {
	if runtime.GOOS == "windows" {
		return "PrismLauncher.exe"
	}
	return "PrismLauncher"
}

// getPrismExecutablePath returns the full path to the Prism executable
func getPrismExecutablePath(prismDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(prismDir, "PrismLauncher.exe")
	}
	// macOS: executable is inside the app bundle
	return filepath.Join(prismDir, "PrismLauncher.app", "Contents", "MacOS", "PrismLauncher")
}

// getPathSeparator returns the platform-specific PATH separator
func getPathSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

// buildPathEnv creates a platform-specific PATH environment variable
func buildPathEnv(additionalPath string) string {
	separator := getPathSeparator()
	return additionalPath + separator + os.Getenv("PATH")
}

func getJavaBinName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func getJavawBinName() string {
	if runtime.GOOS == "windows" {
		return "javaw.exe"
	}
	// macOS has no javaw equivalent
	return "java"
}

// Platform validation
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// Platform support check
func IsSupportedPlatform() bool {
	return IsWindows() || IsDarwin()
}
