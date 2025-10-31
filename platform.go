package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// GetPrismExecutablePath returns the full path to the Prism executable
func GetPrismExecutablePath(prismDir string) string {
	// First check for direct executable (flat structure)
	directPath := GetDirectPrismExecutablePath(prismDir)
	if exists(directPath) {
		return directPath
	}

	// Check for versioned subdirectory structure (e.g., PrismLauncher-9.4/)
	if runtime.GOOS != "darwin" { // macOS doesn't use portable builds with subdirectories
		files, err := os.ReadDir(prismDir)
		if err == nil {
			for _, file := range files {
				if file.IsDir() && strings.Contains(file.Name(), "PrismLauncher") {
					nestedPath := GetDirectPrismExecutablePath(filepath.Join(prismDir, file.Name()))
					if exists(nestedPath) {
						return nestedPath
					}
				}
			}
		}
	}

	// Fallback to direct path
	return directPath
}

// GetDirectPrismExecutablePath returns the path to the Prism executable assuming a flat structure
func GetDirectPrismExecutablePath(baseDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(baseDir, "PrismLauncher.exe")
	} else if runtime.GOOS == "darwin" {
		// macOS: executable is inside the app bundle (note the space in the name)
		return filepath.Join(baseDir, "Prism Launcher.app", "Contents", "MacOS", "prismlauncher")
	} else {
		// Linux: executable is directly in the directory
		return filepath.Join(baseDir, "PrismLauncher")
	}
}

// getPrismBaseDir returns the actual base directory where Prism executable is located
// This handles both flat and nested directory structures
func getPrismBaseDir(prismDir string) string {
	prismExe := GetPrismExecutablePath(prismDir)
	baseDir := filepath.Dir(prismExe)

	// For macOS, the base dir should be the parent of the app bundle
	if runtime.GOOS == "darwin" && strings.HasSuffix(baseDir, "MacOS") {
		baseDir = filepath.Dir(filepath.Dir(baseDir)) // Go up two levels: MacOS -> Contents -> Prism Launcher.app
	}

	return baseDir
}

// getPathSeparator returns the platform-specific PATH separator
func getPathSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

// buildPathEnv creates a platform-specific PATH environment variable
func BuildPathEnv(additionalPath string) string {
	separator := getPathSeparator()
	return additionalPath + separator + os.Getenv("PATH")
}

// GetPrismConfigDir returns the Prism configuration directory
func GetPrismConfigDir() string {
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

// GetPrismDataDir returns the Prism data directory
func GetPrismDataDir() string {
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
