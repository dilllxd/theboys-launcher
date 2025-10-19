package main

import (
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