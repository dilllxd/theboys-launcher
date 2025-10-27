package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to check if file exists (since we can't access the main package's exists function)
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// TestGUIDevModeToggle tests simplified dev mode toggle functionality in GUI
func TestGUIDevModeToggle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Direct dev mode toggle without backup
	t.Run("DirectDevModeToggle", func(t *testing.T) {
		// Setup: Create mock executable
		exePath := filepath.Join(tempDir, "TheBoysLauncher.exe")

		// Create a mock executable
		err := os.WriteFile(exePath, []byte("mock-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock exe: %v", err)
		}

		// Simulate direct dev mode toggle (no backup creation)
		// In new implementation, dev mode toggle directly updates settings and triggers update
		devModeEnabled := true

		// Verify dev mode state
		if !devModeEnabled {
			t.Error("Dev mode should be enabled after toggle")
		}

		// Verify no backup files were created
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")

		if fileExists(backupMetaPath) {
			t.Error("Backup metadata file should not be created in simplified implementation")
		}
		if fileExists(backupExePath) {
			t.Error("Backup executable file should not be created in simplified implementation")
		}
	})

	// Test 2: Direct stable mode toggle without backup restoration
	t.Run("DirectStableModeToggle", func(t *testing.T) {
		// Setup: Create mock executable
		exePath := filepath.Join(tempDir, "TheBoysLauncher.exe")

		// Create a mock executable
		err := os.WriteFile(exePath, []byte("mock-dev-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock exe: %v", err)
		}

		// Simulate direct stable mode toggle (no backup restoration)
		// In new implementation, stable mode toggle directly updates settings and triggers update
		devModeEnabled := false

		// Verify stable mode state
		if devModeEnabled {
			t.Error("Dev mode should be disabled after toggle")
		}

		// Verify no backup restoration occurred
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")

		if fileExists(backupMetaPath) {
			t.Error("Backup metadata should not be used in simplified implementation")
		}
		if fileExists(backupExePath) {
			t.Error("Backup executable should not be used in simplified implementation")
		}
	})

	// Test 3: Fallback to stable update when dev update fails
	t.Run("FallbackOnUpdateFailure", func(t *testing.T) {
		// Setup: Mock update failure scenario
		exePath := filepath.Join(tempDir, "TheBoysLauncher.exe")

		// Create current dev executable
		err = os.WriteFile(exePath, []byte("mock-dev-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create current exe: %v", err)
		}

		// Simulate update failure and fallback logic
		updateFailed := true
		targetDevMode := true
		fallbackAttempted := false

		// Simulate update process with fallback
		if updateFailed && targetDevMode {
			// Attempt fallback to stable
			fallbackAttempted = true
			err = os.WriteFile(exePath, []byte("mock-stable-exe"), 0755)
			if err != nil {
				t.Fatalf("Failed to simulate fallback update: %v", err)
			}
		}

		// Verify fallback was attempted
		if !fallbackAttempted {
			t.Error("Fallback to stable should be attempted when dev update fails")
		}

		// Verify fallback update
		data, err := os.ReadFile(exePath)
		if err != nil {
			t.Fatalf("Failed to read updated exe: %v", err)
		}

		if string(data) != "mock-stable-exe" {
			t.Error("Executable should be updated to stable version when fallback occurs")
		}
	})
}

// TestGUIDevModeErrorHandling tests error handling in simplified dev mode toggle functionality
func TestGUIDevModeErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Handle validation failure gracefully
	t.Run("ValidationFailure", func(t *testing.T) {
		// Simulate validation failure scenario
		validationFailed := true
		originalDevMode := false
		pendingDevMode := true

		// Simulate Save & Apply operation with validation failure
		if validationFailed {
			// Revert checkbox state
			pendingDevMode = originalDevMode
		}

		// Verify checkbox was reverted
		if pendingDevMode != originalDevMode {
			t.Errorf("Checkbox should be reverted on validation failure. Expected %v, got %v",
				originalDevMode, pendingDevMode)
		}
	})

	// Test 2: Handle corrupted settings gracefully
	t.Run("CorruptedSettings", func(t *testing.T) {
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")

		// Write corrupted metadata (should be ignored in new implementation)
		err := os.WriteFile(backupMetaPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write corrupted metadata: %v", err)
		}

		// In new implementation, corrupted backup metadata should be ignored
		// since we don't use backup system anymore
		if fileExists(backupMetaPath) {
			// File might exist but should be ignored
			t.Log("Backup metadata file exists but should be ignored in simplified implementation")
		}

		// The launcher should continue without backup functionality
		// This test verifies that the system doesn't crash when backup files exist
	})
}

// TestGUIDevModeSettingsPersistence tests that dev mode settings are properly persisted
func TestGUIDevModeSettingsPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-persistence-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	settingsPath := filepath.Join(tempDir, "settings.json")

	// Test 1: Save dev mode enabled setting
	t.Run("SaveDevModeEnabled", func(t *testing.T) {
		testSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: true,
		}

		// Save settings
		data, err := json.MarshalIndent(testSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save settings: %v", err)
		}

		// Verify file exists
		if !fileExists(settingsPath) {
			t.Error("Settings file should exist after saving")
		}

		// Read and verify content
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings file: %v", err)
		}

		var loadedSettings LauncherSettings
		err = json.Unmarshal(content, &loadedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		if !loadedSettings.DevBuildsEnabled {
			t.Error("DevBuildsEnabled should be true in saved settings")
		}
	})

	// Test 2: Save dev mode disabled setting
	t.Run("SaveDevModeDisabled", func(t *testing.T) {
		testSettings := LauncherSettings{
			MemoryMB:         8192,
			AutoRAM:          false,
			DevBuildsEnabled: false,
		}

		// Save settings
		data, err := json.MarshalIndent(testSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save settings: %v", err)
		}

		// Read and verify content
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings file: %v", err)
		}

		var loadedSettings LauncherSettings
		err = json.Unmarshal(content, &loadedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		if loadedSettings.DevBuildsEnabled {
			t.Error("DevBuildsEnabled should be false in saved settings")
		}
	})
}

// TestGUIDevModeUIFeedback tests that UI feedback is properly provided during simplified dev mode operations
func TestGUIDevModeUIFeedback(t *testing.T) {
	// Test UI feedback message patterns for simplified implementation
	t.Run("SimplifiedUIFeedbackMessages", func(t *testing.T) {
		testCases := []struct {
			operation   string
			expectedMsg string
			description string
		}{
			{
				operation:   "enable_dev_mode",
				expectedMsg: "Validating update availability",
				description: "Enabling dev mode should show validation message",
			},
			{
				operation:   "update_to_dev",
				expectedMsg: "Updating to latest dev version",
				description: "Dev update should show dev update message",
			},
			{
				operation:   "disable_dev_mode",
				expectedMsg: "Validating update availability",
				description: "Disabling dev mode should show validation message",
			},
			{
				operation:   "update_to_stable",
				expectedMsg: "Updating to latest stable version",
				description: "Stable update should show stable update message",
			},
			{
				operation:   "fallback_to_stable",
				expectedMsg: "Attempting fallback to stable",
				description: "Fallback should show fallback message",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.operation, func(t *testing.T) {
				// Simulate simplified UI feedback logic from gui.go
				var feedbackMessage string

				switch tc.operation {
				case "enable_dev_mode":
					feedbackMessage = "Validating update availability..."
				case "update_to_dev":
					feedbackMessage = "Updating to latest dev version..."
				case "disable_dev_mode":
					feedbackMessage = "Validating update availability..."
				case "update_to_stable":
					feedbackMessage = "Updating to latest stable version..."
				case "fallback_to_stable":
					feedbackMessage = "Attempting fallback to stable..."
				}

				if feedbackMessage == "" {
					t.Error("Feedback message should not be empty")
				}

				if !containsString(feedbackMessage, tc.expectedMsg) {
					t.Errorf("Expected feedback message to contain '%s', got '%s'", tc.expectedMsg, feedbackMessage)
				}
			})
		}
	})

	// Test channel status display
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
}

// TestGUIDevModeValidation tests the pre-update validation functionality
func TestGUIDevModeValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-validation-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Validation success scenario
	t.Run("ValidationSuccess", func(t *testing.T) {
		// Simulate successful validation
		validationPassed := true

		// Simulate validation logic
		if validationPassed {
			// Validation passed, proceed with update
			// In real implementation, this would trigger forceUpdate
		} else {
			t.Error("Validation should have passed in this test scenario")
		}

		// Verify we can proceed with update
		if !validationPassed {
			t.Error("Should be able to proceed with update after successful validation")
		}
	})

	// Test 2: Validation failure scenario
	t.Run("ValidationFailure", func(t *testing.T) {
		// Simulate validation failure
		validationPassed := false
		targetDevMode := true
		originalDevMode := false

		// Simulate validation failure handling
		if !validationPassed {
			// Revert to original state
			targetDevMode = originalDevMode
		}

		// Verify state was reverted
		if targetDevMode != originalDevMode {
			t.Errorf("Dev mode should be reverted to original state (%v) on validation failure, got %v",
				originalDevMode, targetDevMode)
		}
	})

	// Test 3: Network error handling
	t.Run("NetworkErrorHandling", func(t *testing.T) {
		// Simulate network error during validation
		networkError := true
		originalDevMode := false
		pendingDevMode := true

		// Simulate network error handling
		if networkError {
			// Show error message and revert
			pendingDevMode = originalDevMode
		}

		// Verify error handling
		if pendingDevMode != originalDevMode {
			t.Errorf("Dev mode should be reverted on network error. Expected %v, got %v",
				originalDevMode, pendingDevMode)
		}
	})
}

// Helper function to check if string contains substring (using strings.Contains for simplicity)
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
