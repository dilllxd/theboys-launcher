// Package platform provides cross-platform detection and utilities
package platform

import (
	"runtime"
)

// Platform represents the current operating system
type Platform int

const (
	Windows Platform = iota
	macOS
	Linux
	Unknown
)

// String returns the string representation of the platform
func (p Platform) String() string {
	switch p {
	case Windows:
		return "windows"
	case macOS:
		return "darwin"
	case Linux:
		return "linux"
	default:
		return "unknown"
	}
}

// GetCurrentPlatform returns the current platform
func GetCurrentPlatform() Platform {
	switch runtime.GOOS {
	case "windows":
		return Windows
	case "darwin":
		return macOS
	case "linux":
		return Linux
	default:
		return Unknown
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return GetCurrentPlatform() == Windows
}

// IsMacOS returns true if running on macOS
func IsMacOS() bool {
	return GetCurrentPlatform() == macOS
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return GetCurrentPlatform() == Linux
}

// IsUnix returns true if running on a Unix-like system
func IsUnix() bool {
	p := GetCurrentPlatform()
	return p == macOS || p == Linux
}

// GetExecutableExtension returns the executable extension for the current platform
func GetExecutableExtension() string {
	switch GetCurrentPlatform() {
	case Windows:
		return ".exe"
	default:
		return ""
	}
}

// GetArchiveExtension returns the preferred archive extension for the current platform
func GetArchiveExtension() string {
	switch GetCurrentPlatform() {
	case Windows:
		return ".zip"
	default:
		return ".tar.gz"
	}
}