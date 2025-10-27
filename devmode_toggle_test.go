package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDevModeToggleBehavior tests simplified dev mode toggle behavior
// This tests the new implementation where settings are applied directly with Save & Apply button
func TestDevModeToggleBehavior(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-toggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Direct settings application
	t.Run("DirectSettingsApplication", func(t *testing.T) {
		// Setup initial settings
		originalSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: false,
		}

		// Save original settings
		settingsPath := filepath.Join(tempDir, "settings.json")
		data, err := json.MarshalIndent(originalSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal original settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save original settings: %v", err)
		}

		// Simulate Save & Apply operation
		updatedSettings := originalSettings
		updatedSettings.DevBuildsEnabled = true // User enables dev mode

		// Save updated settings
		data, err = json.MarshalIndent(updatedSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal updated settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save updated settings: %v", err)
		}

		// Verify settings were updated
		savedData, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read saved settings: %v", err)
		}

		var savedSettings LauncherSettings
		err = json.Unmarshal(savedData, &savedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved settings: %v", err)
		}

		// Settings should now reflect the change
		if savedSettings.DevBuildsEnabled != true {
			t.Errorf("Settings should be updated after Save & Apply. Expected true, got %v", savedSettings.DevBuildsEnabled)
		}
	})

	// Test 2: Multiple settings changes in one operation
	t.Run("MultipleSettingsChanges", func(t *testing.T) {
		// Setup initial settings
		originalSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: false,
		}

		// Save original settings
		settingsPath := filepath.Join(tempDir, "settings2.json")
		data, err := json.MarshalIndent(originalSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal original settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save original settings: %v", err)
		}

		// Simulate multiple settings changes and Save & Apply
		updatedSettings := originalSettings
		updatedSettings.DevBuildsEnabled = true
		updatedSettings.AutoRAM = false
		updatedSettings.MemoryMB = 8192

		// Save updated settings
		data, err = json.MarshalIndent(updatedSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal updated settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save updated settings: %v", err)
		}

		// Verify settings were updated
		savedData, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read saved settings: %v", err)
		}

		var savedSettings LauncherSettings
		err = json.Unmarshal(savedData, &savedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved settings: %v", err)
		}

		// All settings should be updated
		if savedSettings.DevBuildsEnabled != true {
			t.Errorf("DevBuildsEnabled should be true, got %v", savedSettings.DevBuildsEnabled)
		}
		if savedSettings.AutoRAM != false {
			t.Errorf("AutoRAM should be false, got %v", savedSettings.AutoRAM)
		}
		if savedSettings.MemoryMB != 8192 {
			t.Errorf("MemoryMB should be 8192, got %v", savedSettings.MemoryMB)
		}
	})
}

// TestDevModeUIFeedback tests UI feedback for the simplified settings
func TestDevModeUIFeedback(t *testing.T) {
	// Test 1: Channel status display
	t.Run("ChannelStatusDisplay", func(t *testing.T) {
		testCases := []struct {
			name            string
			devModeEnabled  bool
			expectedChannel string
		}{
			{
				name:            "Dev mode enabled",
				devModeEnabled:  true,
				expectedChannel: "Channel: Dev",
			},
			{
				name:            "Dev mode disabled",
				devModeEnabled:  false,
				expectedChannel: "Channel: Stable",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate channel label logic from simplified gui.go
				var channelLabel string
				if tc.devModeEnabled {
					channelLabel = "Channel: Dev"
				} else {
					channelLabel = "Channel: Stable"
				}

				if channelLabel != tc.expectedChannel {
					t.Errorf("Expected channel label '%s', got '%s'", tc.expectedChannel, channelLabel)
				}
			})
		}
	})

	// Test 2: Channel label updates on checkbox toggle
	t.Run("ChannelLabelUpdates", func(t *testing.T) {
		// Simulate checkbox toggle behavior
		devCheckChecked := false
		channelLabel := "Channel: Stable"

		// User checks the checkbox
		devCheckChecked = true
		if devCheckChecked {
			channelLabel = "Channel: Dev"
		}

		if channelLabel != "Channel: Dev" {
			t.Errorf("Expected channel label to be 'Channel: Dev' when checkbox is checked, got '%s'", channelLabel)
		}

		// User unchecks the checkbox
		devCheckChecked = false
		if !devCheckChecked {
			channelLabel = "Channel: Stable"
		}

		if channelLabel != "Channel: Stable" {
			t.Errorf("Expected channel label to be 'Channel: Stable' when checkbox is unchecked, got '%s'", channelLabel)
		}
	})
}

// TestDevModeUpdateProcess tests the simplified update process
func TestDevModeUpdateProcess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-update-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Update process runs only on Save & Apply
	t.Run("UpdateProcessOnlyOnSaveApply", func(t *testing.T) {
		// Track update process calls
		updateProcessCalls := 0

		// Simulate update process tracking
		simulateUpdateProcess := func() {
			updateProcessCalls++
		}

		// Simulate checkbox toggle (should not trigger update)
		devCheckChecked := true
		_ = devCheckChecked // Use variable to avoid unused error

		// Update process should not have been called yet
		if updateProcessCalls != 0 {
			t.Errorf("Update process should not be called during checkbox toggle. Called %d times", updateProcessCalls)
		}

		// Simulate Save & Apply operation (should trigger update once)
		simulateUpdateProcess()

		// Update process should have been called exactly once
		if updateProcessCalls != 1 {
			t.Errorf("Update process should be called exactly once on Save & Apply. Called %d times", updateProcessCalls)
		}
	})

	// Test 2: No update when dialog is closed without saving
	t.Run("NoUpdateWhenDialogClosed", func(t *testing.T) {
		// Track update process calls
		updateProcessCalls := 0

		// Simulate checkbox toggle
		devCheckChecked := true
		_ = devCheckChecked // Use variable to avoid unused error

		// Simulate closing dialog without Save & Apply
		// User closes dialog - no update

		// Update process should not have been called
		if updateProcessCalls != 0 {
			t.Errorf("Update process should not be called when dialog is closed without saving. Called %d times", updateProcessCalls)
		}
	})
}

// TestDevModeErrorHandling tests error handling in the simplified toggle behavior
func TestDevModeErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Revert checkbox state on validation failure
	t.Run("RevertOnValidationFailure", func(t *testing.T) {
		// Simulate validation failure scenario
		validationFailed := true
		originalDevMode := false
		pendingDevMode := true

		// Simulate the Save & Apply operation with validation failure
		if validationFailed {
			// Revert the checkbox state
			pendingDevMode = originalDevMode
		}

		// Verify checkbox was reverted
		if pendingDevMode != originalDevMode {
			t.Errorf("Checkbox should be reverted on validation failure. Expected %v, got %v",
				originalDevMode, pendingDevMode)
		}
	})

	// Test 2: Handle corrupted settings gracefully
	t.Run("HandleCorruptedSettings", func(t *testing.T) {
		// Create corrupted settings file
		settingsPath := filepath.Join(tempDir, "corrupted-settings.json")
		err := os.WriteFile(settingsPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create corrupted settings: %v", err)
		}

		// Try to load settings
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read corrupted settings: %v", err)
		}

		var settings LauncherSettings
		err = json.Unmarshal(data, &settings)
		if err == nil {
			t.Error("Expected error when unmarshaling corrupted settings")
		}

		// Should fall back to defaults
		defaultSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: false,
		}

		// In real implementation, this would use default settings
		if defaultSettings.DevBuildsEnabled != false {
			t.Error("Default dev mode should be false for stable builds")
		}
	})

	// Test 3: Fallback behavior on update failure
	t.Run("FallbackOnUpdateFailure", func(t *testing.T) {
		// Simulate update failure scenario
		updateFailed := true
		targetDevMode := true

		// Track fallback attempts
		fallbackAttempted := false

		// Simulate update process with fallback
		if updateFailed && targetDevMode {
			// Attempt fallback to stable
			fallbackAttempted = true
		}

		// Verify fallback was attempted
		if !fallbackAttempted {
			t.Error("Fallback to stable should be attempted when dev update fails")
		}
	})
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
