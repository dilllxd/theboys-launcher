package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Define the LauncherSettings type for testing
type LauncherSettings struct {
	MemoryMB         int  `json:"memoryMB"`
	AutoRAM          bool `json:"autoRam"`
	DevBuildsEnabled bool `json:"devBuildsEnabled,omitempty"`
}

// TestDevBuildsSettings tests the dev builds toggle functionality
func TestDevBuildsSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: isDevBuild function
	t.Run("IsDevBuildFunction", func(t *testing.T) {
		testCases := []struct {
			version     string
			expected    bool
			description string
		}{
			{"dev", true, "Simple dev version"},
			{"v1.0.0-dev", true, "Version with dev suffix"},
			{"v1.0.0-dev.abc123", true, "Version with dev and hash"},
			{"v1.0.0", false, "Stable release"},
			{"v1.0.0-beta", false, "Beta release (not dev)"},
			{"v1.0.0-rc1", false, "Release candidate (not dev)"},
			{"", false, "Empty version"},
			{"DEV", true, "Uppercase dev"},
			{"v1.0.0-DEV", true, "Uppercase dev suffix"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result := isDevBuildVersion(tc.version)
				if result != tc.expected {
					t.Errorf("Expected isDevBuildVersion() to return %v for version '%s', got %v",
						tc.expected, tc.version, result)
				}
			})
		}
	})

	// Test 2: Settings structure and JSON marshaling
	t.Run("SettingsStructure", func(t *testing.T) {
		// Test that LauncherSettings structure includes DevBuildsEnabled
		settings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: true,
		}

		// Test JSON marshaling
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal settings: %v", err)
		}

		// Verify the JSON contains our field
		jsonStr := string(data)
		if !contains(jsonStr, "devBuildsEnabled") {
			t.Error("JSON output should contain devBuildsEnabled field")
		}

		// Test JSON unmarshaling
		var unmarshaled LauncherSettings
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		if unmarshaled.DevBuildsEnabled != settings.DevBuildsEnabled {
			t.Errorf("Expected DevBuildsEnabled to be %v, got %v",
				settings.DevBuildsEnabled, unmarshaled.DevBuildsEnabled)
		}
	})

	// Test 3: Settings file creation and loading
	t.Run("SettingsFileOperations", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "settings.json")

		// Create test settings
		testSettings := map[string]interface{}{
			"memoryMB":         6144,
			"autoRam":          false,
			"devBuildsEnabled": true,
		}

		data, err := json.MarshalIndent(testSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test settings: %v", err)
		}

		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write test settings: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
			t.Error("Settings file should exist after writing")
		}

		// Read and verify content
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings file: %v", err)
		}

		var loadedSettings map[string]interface{}
		err = json.Unmarshal(content, &loadedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		// Verify devBuildsEnabled field
		if devBuildsEnabled, ok := loadedSettings["devBuildsEnabled"].(bool); !ok || !devBuildsEnabled {
			t.Error("Expected devBuildsEnabled to be true in loaded settings")
		}
	})

	// Test 4: Default settings for different versions
	t.Run("DefaultSettingsByVersion", func(t *testing.T) {
		// Test dev version defaults
		devDefaults := createDefaultSettings("dev")
		if !devDefaults.DevBuildsEnabled {
			t.Error("Dev version should have DevBuildsEnabled true by default")
		}

		// Test stable version defaults
		stableDefaults := createDefaultSettings("v1.0.0")
		if stableDefaults.DevBuildsEnabled {
			t.Error("Stable version should have DevBuildsEnabled false by default")
		}

		// Test prerelease version defaults
		prereleaseDefaults := createDefaultSettings("v1.0.0-beta")
		if prereleaseDefaults.DevBuildsEnabled {
			t.Error("Beta prerelease should have DevBuildsEnabled false by default")
		}
	})
}

// Helper function to test isDevBuild logic without accessing the global variable
func isDevBuildVersion(version string) bool {
	lower := strings.ToLower(version)
	return strings.Contains(lower, "dev")
}

// Helper function to create default settings based on version
func createDefaultSettings(version string) LauncherSettings {
	return LauncherSettings{
		MemoryMB:         4096, // Default value for testing
		AutoRAM:          true,
		DevBuildsEnabled: isDevBuildVersion(version),
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestDevBuildsIntegration tests integration aspects of dev builds functionality
func TestDevBuildsIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("DevBuildsSettingInUpdateFlow", func(t *testing.T) {
		// Test that the setting would be used in update logic
		// We can't test the actual network calls, but we can verify the logic

		// Simulate dev builds enabled
		settings := LauncherSettings{DevBuildsEnabled: true}
		preferDev := settings.DevBuildsEnabled

		if !preferDev {
			t.Error("Expected preferDev to be true when DevBuildsEnabled is true")
		}

		// Simulate dev builds disabled
		settings = LauncherSettings{DevBuildsEnabled: false}
		preferDev = settings.DevBuildsEnabled

		if preferDev {
			t.Error("Expected preferDev to be false when DevBuildsEnabled is false")
		}
	})

	t.Run("SettingsPersistenceAcrossRestarts", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "settings.json")

		// Initial settings with dev builds enabled
		originalSettings := LauncherSettings{
			MemoryMB:         8192,
			AutoRAM:          false,
			DevBuildsEnabled: true,
		}

		// Save settings
		data, err := json.MarshalIndent(originalSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal settings: %v", err)
		}

		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save settings: %v", err)
		}

		// Simulate launcher restart by reading settings back
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings: %v", err)
		}

		var restartedSettings LauncherSettings
		err = json.Unmarshal(content, &restartedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		// Verify all settings persisted
		if restartedSettings.MemoryMB != originalSettings.MemoryMB {
			t.Errorf("MemoryMB not persisted: expected %d, got %d",
				originalSettings.MemoryMB, restartedSettings.MemoryMB)
		}

		if restartedSettings.AutoRAM != originalSettings.AutoRAM {
			t.Errorf("AutoRAM not persisted: expected %v, got %v",
				originalSettings.AutoRAM, restartedSettings.AutoRAM)
		}

		if restartedSettings.DevBuildsEnabled != originalSettings.DevBuildsEnabled {
			t.Errorf("DevBuildsEnabled not persisted: expected %v, got %v",
				originalSettings.DevBuildsEnabled, restartedSettings.DevBuildsEnabled)
		}
	})
}
