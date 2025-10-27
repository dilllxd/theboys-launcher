package main

import (
	"runtime"
	"testing"
)

func TestGetLauncherExeName(t *testing.T) {
	expected := "TheBoysLauncher"
	if runtime.GOOS == "windows" {
		expected = "TheBoysLauncher.exe"
	}

	result := GetLauncherExeName()
	if result != expected {
		t.Errorf("getLauncherExeName() = %v, want %v", result, expected)
	}
}

func TestGetLauncherAssetName(t *testing.T) {
	var expected string

	switch runtime.GOOS {
	case "windows":
		expected = "TheBoysLauncher.exe"
	case "darwin":
		expected = "TheBoysLauncher-mac-universal"
	case "linux":
		expected = "TheBoysLauncher-linux"
	default:
		expected = "TheBoysLauncher"
	}

	result := GetLauncherAssetName()
	if result != expected {
		t.Errorf("getLauncherAssetName() = %v, want %v", result, expected)
	}
}

func TestGetPrismExeName(t *testing.T) {
	expected := "PrismLauncher"
	if runtime.GOOS == "windows" {
		expected = "PrismLauncher.exe"
	}

	result := GetPrismExeName()
	if result != expected {
		t.Errorf("getPrismExeName() = %v, want %v", result, expected)
	}
}

func TestGetJavaBinName(t *testing.T) {
	expected := "java"
	if runtime.GOOS == "windows" {
		expected = "java.exe"
	}

	result := GetJavaBinName()
	if result != expected {
		t.Errorf("getJavaBinName() = %v, want %v", result, expected)
	}
}

func TestGetJavawBinName(t *testing.T) {
	expected := "java"
	if runtime.GOOS == "windows" {
		expected = "javaw.exe"
	}

	result := GetJavawBinName()
	if result != expected {
		t.Errorf("getJavawBinName() = %v, want %v", result, expected)
	}
}

func TestGetPathSeparator(t *testing.T) {
	expected := ":"
	if runtime.GOOS == "windows" {
		expected = ";"
	}

	result := GetPathSeparator()
	if result != expected {
		t.Errorf("getPathSeparator() = %v, want %v", result, expected)
	}
}

func TestPlatformDetection(t *testing.T) {
	// Test platform detection functions
	isWindows := IsWindows()
	isDarwin := IsDarwin()
	isLinux := IsLinux()
	isSupported := IsSupportedPlatform()

	// Only one of these should be true
	trueCount := 0
	if isWindows {
		trueCount++
	}
	if isDarwin {
		trueCount++
	}
	if isLinux {
		trueCount++
	}

	if trueCount != 1 {
		t.Errorf("Exactly one platform should be detected, got %d true values", trueCount)
	}

	// Supported platform should match the current OS
	expectedSupported := runtime.GOOS == "windows" || runtime.GOOS == "darwin" || runtime.GOOS == "linux"
	if isSupported != expectedSupported {
		t.Errorf("IsSupportedPlatform() = %v, want %v", isSupported, expectedSupported)
	}
}

func TestGetPrismExecutablePath(t *testing.T) {
	prismDir := "/test/path"
	result := GetPrismExecutablePath(prismDir)

	var expected string
	if runtime.GOOS == "windows" {
		expected = "/test/path/PrismLauncher.exe"
	} else if runtime.GOOS == "darwin" {
		expected = "/test/path/Prism Launcher.app/Contents/MacOS/PrismLauncher"
	} else {
		// Linux and other platforms
		expected = "/test/path/PrismLauncher"
	}

	if result != expected {
		t.Errorf("getPrismExecutablePath() = %v, want %v", result, expected)
	}
}

func TestBuildPathEnv(t *testing.T) {
	additionalPath := "/additional/path"
	result := BuildPathEnv(additionalPath)

	separator := GetPathSeparator()
	expected := additionalPath + separator + "PATH"

	if result != expected {
		t.Errorf("buildPathEnv() = %v, want %v", result, expected)
	}
}
