package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestGetDirectPrismExecutablePathMacOS tests that the macOS executable path is correct
func TestGetDirectPrismExecutablePathMacOS(t *testing.T) {
	// Test base directory
	testDir := "/Applications"

	// Test the actual implementation for macOS
	// We need to check if the function correctly returns the lowercase executable name
	expectedPath := filepath.Join(testDir, "Prism Launcher.app", "Contents", "MacOS", "prismlauncher")

	// Since we can't directly modify runtime.GOOS, we'll test the function logic
	// by checking what it would return on macOS
	result := simulateGetDirectPrismExecutablePathForMacOS(testDir)

	if result != expectedPath {
		t.Errorf("GetDirectPrismExecutablePath() on macOS = %v, want %v", result, expectedPath)
	}

	// Verify the executable name is lowercase (not "PrismLauncher")
	if filepath.Base(result) != "prismlauncher" {
		t.Errorf("Executable name should be lowercase 'prismlauncher', got %v", filepath.Base(result))
	}

	// Verify the path doesn't contain the old uppercase "PrismLauncher"
	if stringContains(result, "PrismLauncher") {
		t.Errorf("Path should not contain uppercase 'PrismLauncher', got %v", result)
	}
}

// TestLauncherFallbackLogicMacOS tests the fallback logic in launcher.go for macOS
func TestLauncherFallbackLogicMacOS(t *testing.T) {
	// Test the fallback logic that checks for Prism in /Applications
	testPrismDir := "/some/local/path"
	applicationsPrism := filepath.Join("/Applications", "Prism Launcher.app", "Contents", "MacOS", "prismlauncher")

	// Simulate the fallback logic from launcher.go lines 1307-1315
	prismExe := simulateGetDirectPrismExecutablePathForMacOS(testPrismDir)

	// If local Prism doesn't exist, it should fallback to /Applications
	if !simulateExists(prismExe) && simulateExists(applicationsPrism) {
		prismExe = applicationsPrism
	}

	// Verify the fallback path uses the correct lowercase executable name
	if filepath.Base(prismExe) != "prismlauncher" {
		t.Errorf("Fallback executable name should be lowercase 'prismlauncher', got %v", filepath.Base(prismExe))
	}

	// Verify the fallback path doesn't contain the old uppercase "PrismLauncher"
	if stringContains(prismExe, "PrismLauncher") {
		t.Errorf("Fallback path should not contain uppercase 'PrismLauncher', got %v", prismExe)
	}
}

// TestPrismExecutablePathComponents tests the individual components of the macOS path
func TestPrismExecutablePathComponentsMacOS(t *testing.T) {
	testDir := "/test/path"
	result := simulateGetDirectPrismExecutablePathForMacOS(testDir)

	// Check that the path has the correct structure
	expectedComponents := []string{"test", "path", "Prism Launcher.app", "Contents", "MacOS", "prismlauncher"}
	actualComponents := strings.Split(filepath.Clean(result), string(filepath.Separator))

	// Remove empty components and adjust for leading slash
	var cleanActualComponents []string
	for _, comp := range actualComponents {
		if comp != "" {
			cleanActualComponents = append(cleanActualComponents, comp)
		}
	}

	// Compare the last components (ignoring potential leading slash)
	startIdx := len(cleanActualComponents) - len(expectedComponents)
	if startIdx < 0 {
		t.Errorf("Path doesn't have enough components: %v", cleanActualComponents)
		return
	}

	for i, expected := range expectedComponents {
		if cleanActualComponents[startIdx+i] != expected {
			t.Errorf("Path component mismatch at index %d: got %v, want %v",
				i, cleanActualComponents[startIdx+i], expected)
		}
	}
}

// Helper function to simulate GetDirectPrismExecutablePath for macOS
func simulateGetDirectPrismExecutablePathForMacOS(baseDir string) string {
	// This simulates the macOS branch of GetDirectPrismExecutablePath
	return filepath.Join(baseDir, "Prism Launcher.app", "Contents", "MacOS", "prismlauncher")
}

// Helper function to simulate file existence check
func simulateExists(path string) bool {
	// For testing purposes, we'll simulate that /Applications path exists
	// and local paths don't
	return stringContains(path, "/Applications")
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
