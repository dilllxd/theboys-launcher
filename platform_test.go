package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetLauncherExeName(t *testing.T) {
	expected := "TheBoysLauncher"
	if runtime.GOOS == "windows" {
		expected = "TheBoysLauncher.exe"
	}

	result := GetLauncherExeName()
	if result != expected {
		t.Errorf("GetLauncherExeName() = %v, want %v", result, expected)
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
		t.Errorf("GetLauncherAssetName() = %v, want %v", result, expected)
	}
}

func TestGetPrismExeName(t *testing.T) {
	expected := "PrismLauncher"
	if runtime.GOOS == "windows" {
		expected = "PrismLauncher.exe"
	}

	result := GetPrismExeName()
	if result != expected {
		t.Errorf("GetPrismExeName() = %v, want %v", result, expected)
	}
}

func TestGetJavaBinName(t *testing.T) {
	expected := "java"
	if runtime.GOOS == "windows" {
		expected = "java.exe"
	}

	result := GetJavaBinName()
	if result != expected {
		t.Errorf("GetJavaBinName() = %v, want %v", result, expected)
	}
}

func TestGetJavawBinName(t *testing.T) {
	expected := "java"
	if runtime.GOOS == "windows" {
		expected = "javaw.exe"
	}

	result := GetJavawBinName()
	if result != expected {
		t.Errorf("GetJavawBinName() = %v, want %v", result, expected)
	}
}

func TestGetPathSeparator(t *testing.T) {
	expected := ":"
	if runtime.GOOS == "windows" {
		expected = ";"
	}

	result := GetPathSeparator()
	if result != expected {
		t.Errorf("GetPathSeparator() = %v, want %v", result, expected)
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
		expected = filepath.Join(prismDir, "PrismLauncher.exe")
	} else {
		expected = filepath.Join(prismDir, "Prism Launcher.app", "Contents", "MacOS", "PrismLauncher")
	}

	if result != expected {
		t.Errorf("GetPrismExecutablePath() = %v, want %v", result, expected)
	}
}

func TestBuildPathEnv(t *testing.T) {
	additionalPath := "/additional/path"
	result := BuildPathEnv(additionalPath)

	separator := GetPathSeparator()
	expected := additionalPath + separator + "PATH"

	// Just check that the result contains our expected pattern
	// The actual PATH will be much longer on Windows due to system PATH
	if !strings.Contains(result, expected) {
		t.Errorf("BuildPathEnv() = %v, want to contain %v", result, expected)
	}
}
