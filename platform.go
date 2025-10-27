package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// Platform-specific constants
const (
	LauncherName = "TheBoysLauncher"
)

// Platform-specific executable names
var (
	LauncherExeName   = GetLauncherExeName()
	LauncherAssetName = GetLauncherAssetName()
	PrismExeName      = GetPrismExeName()
	JavaBinName       = GetJavaBinName()
	JavawBinName      = GetJavawBinName()
)

// Platform detection functions
func GetLauncherExeName() string {
	if runtime.GOOS == "windows" {
		return "TheBoysLauncher.exe"
	}
	return "TheBoysLauncher"
}

// getLauncherAssetName returns the platform-specific asset name for updates
func GetLauncherAssetName() string {
	if runtime.GOOS == "windows" {
		return "TheBoysLauncher.exe"
	} else if runtime.GOOS == "darwin" {
		// macOS uses the universal binary name
		return "TheBoysLauncher-mac-universal"
	} else if runtime.GOOS == "linux" {
		// Linux uses the platform-specific binary name
		return "TheBoysLauncher-linux"
	}
	// Fallback for other platforms
	return "TheBoysLauncher"
}

func GetPrismExeName() string {
	if runtime.GOOS == "windows" {
		return "PrismLauncher.exe"
	}
	return "PrismLauncher"
}

// getPrismExecutablePath returns the full path to the Prism executable
func GetPrismExecutablePath(prismDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(prismDir, "PrismLauncher.exe")
	}
	// macOS: executable is inside the app bundle (note the space in the name)
	return filepath.Join(prismDir, "Prism Launcher.app", "Contents", "MacOS", "PrismLauncher")
}

// getPathSeparator returns the platform-specific PATH separator
func GetPathSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

// buildPathEnv creates a platform-specific PATH environment variable
func BuildPathEnv(additionalPath string) string {
	separator := GetPathSeparator()
	return additionalPath + separator + os.Getenv("PATH")
}

// getPrismConfigDir returns the Prism configuration directory
func getPrismConfigDir() string {
	if runtime.GOOS == "windows" {
		// Use our launcher directory for Windows portable
		return getLauncherHome()
	}
	// macOS/Linux: use Prism's standard config directory
	if runtime.GOOS == "darwin" {
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "PrismLauncher")
	}
	// Linux (assuming XDG_CONFIG_HOME or fallback)
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configHome, "PrismLauncher")
}

// getPrismDataDir returns the Prism data directory
func getPrismDataDir() string {
	if runtime.GOOS == "windows" {
		// Use our launcher directory for Windows portable
		return getLauncherHome()
	}
	// macOS/Linux: use Prism's standard data directory
	if runtime.GOOS == "darwin" {
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "PrismLauncher")
	}
	// Linux (assuming XDG_DATA_HOME or fallback)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(dataHome, "PrismLauncher")
}

func GetJavaBinName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func GetJavawBinName() string {
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
	return IsWindows() || IsDarwin() || IsLinux()
}
