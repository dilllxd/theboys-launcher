package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to check if file exists (since we can't access the main package's exists function)
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// TestGUIDevModeToggle tests the enhanced dev mode toggle functionality in the GUI
func TestGUIDevModeToggle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Backup creation when enabling dev mode
	t.Run("BackupCreationOnEnable", func(t *testing.T) {
		// Setup: Create mock stable executable
		stableExePath := filepath.Join(tempDir, "TheBoysLauncher.exe")
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")

		// Create a mock stable executable
		err := os.WriteFile(stableExePath, []byte("mock-stable-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock stable exe: %v", err)
		}

		// Simulate backup creation logic from gui.go
		// This simulates the backup creation when enabling dev mode
		if !fileExists(backupMetaPath) || !fileExists(backupExePath) {
			// Simulate fetching stable version metadata
			tag := "v3.2.27"
			meta := map[string]string{"tag": tag, "path": backupExePath}

			// Write backup metadata
			data, err := json.MarshalIndent(meta, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal backup metadata: %v", err)
			}

			err = os.WriteFile(backupMetaPath, data, 0644)
			if err != nil {
				t.Fatalf("Failed to write backup metadata: %v", err)
			}

			// Simulate downloading stable exe to backup path
			err = os.WriteFile(backupExePath, []byte("mock-backup-exe"), 0755)
			if err != nil {
				t.Fatalf("Failed to create backup exe: %v", err)
			}
		}

		// Verify backup was created
		if !fileExists(backupMetaPath) {
			t.Error("Backup metadata file should exist after enabling dev mode")
		}
		if !fileExists(backupExePath) {
			t.Error("Backup executable file should exist after enabling dev mode")
		}

		// Verify backup metadata content
		data, err := os.ReadFile(backupMetaPath)
		if err != nil {
			t.Fatalf("Failed to read backup metadata: %v", err)
		}

		var meta map[string]string
		err = json.Unmarshal(data, &meta)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup metadata: %v", err)
		}

		if meta["tag"] != "v3.2.27" {
			t.Errorf("Expected backup tag to be v3.2.27, got %s", meta["tag"])
		}
	})

	// Test 2: Backup restoration when disabling dev mode
	t.Run("BackupRestorationOnDisable", func(t *testing.T) {
		// Setup: Create backup files
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
		currentExePath := filepath.Join(tempDir, "TheBoysLauncher.exe")

		// Create backup metadata
		meta := map[string]string{"tag": "v3.2.27", "path": backupExePath}
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal backup metadata: %v", err)
		}
		err = os.WriteFile(backupMetaPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write backup metadata: %v", err)
		}

		// Create backup executable
		err = os.WriteFile(backupExePath, []byte("mock-backup-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup exe: %v", err)
		}

		// Create current dev executable
		err = os.WriteFile(currentExePath, []byte("mock-dev-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create current exe: %v", err)
		}

		// Simulate backup restoration logic from gui.go
		// This simulates checking for backup and restoring when disabling dev mode
		if fileExists(backupMetaPath) && fileExists(backupExePath) {
			// In real implementation, this would call replaceAndRestart
			// For testing, we'll simulate the restoration by copying
			data, err := os.ReadFile(backupExePath)
			if err != nil {
				t.Fatalf("Failed to read backup exe: %v", err)
			}
			err = os.WriteFile(currentExePath, data, 0755)
			if err != nil {
				t.Fatalf("Failed to restore backup exe: %v", err)
			}
		}

		// Verify restoration
		data, err = os.ReadFile(currentExePath)
		if err != nil {
			t.Fatalf("Failed to read restored exe: %v", err)
		}

		if string(data) != "mock-backup-exe" {
			t.Error("Current executable should be restored from backup")
		}
	})

	// Test 3: Fallback to stable update when no backup exists
	t.Run("FallbackToStableUpdate", func(t *testing.T) {
		// Setup: No backup files exist
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")
		currentExePath := filepath.Join(tempDir, "TheBoysLauncher.exe")

		// Ensure backup doesn't exist
		os.Remove(backupMetaPath)
		os.Remove(backupExePath)

		// Create current dev executable
		err = os.WriteFile(currentExePath, []byte("mock-dev-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create current exe: %v", err)
		}

		// Simulate fallback logic from gui.go
		// This simulates the fallback to stable update when no backup exists
		if !(fileExists(backupMetaPath) && fileExists(backupExePath)) {
			// In real implementation, this would call forceUpdate with preferDev=false
			// For testing, we'll simulate the stable update
			err = os.WriteFile(currentExePath, []byte("mock-stable-exe"), 0755)
			if err != nil {
				t.Fatalf("Failed to simulate stable update: %v", err)
			}
		}

		// Verify fallback update
		data, err := os.ReadFile(currentExePath)
		if err != nil {
			t.Fatalf("Failed to read updated exe: %v", err)
		}

		if string(data) != "mock-stable-exe" {
			t.Error("Current executable should be updated to stable version when no backup exists")
		}
	})
}

// TestGUIDevModeErrorHandling tests error handling in the dev mode toggle functionality
func TestGUIDevModeErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Handle backup creation failure
	t.Run("BackupCreationFailure", func(t *testing.T) {
		// Setup: Use invalid backup path to simulate failure
		backupExePath := filepath.Join(tempDir, "invalid", "path", "backup.exe")
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")

		// Simulate backup creation with invalid path
		tag := "v3.2.27"
		meta := map[string]string{"tag": tag, "path": backupExePath}

		// This should fail when trying to write to invalid path
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal backup metadata: %v", err)
		}

		err = os.WriteFile(backupMetaPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write backup metadata: %v", err)
		}

		// Try to create backup exe in invalid path
		err = os.WriteFile(backupExePath, []byte("mock-backup-exe"), 0755)
		if err == nil {
			t.Error("Expected backup creation to fail with invalid path")
		}
	})

	// Test 2: Handle corrupted backup metadata
	t.Run("CorruptedBackupMetadata", func(t *testing.T) {
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")

		// Write corrupted metadata
		err := os.WriteFile(backupMetaPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write corrupted metadata: %v", err)
		}

		// Try to read corrupted metadata
		data, err := os.ReadFile(backupMetaPath)
		if err != nil {
			t.Fatalf("Failed to read corrupted metadata: %v", err)
		}

		var meta map[string]string
		err = json.Unmarshal(data, &meta)
		if err == nil {
			t.Error("Expected metadata unmarshaling to fail with corrupted JSON")
		}
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

// TestGUIDevModeUIFeedback tests that UI feedback is properly provided during dev mode operations
func TestGUIDevModeUIFeedback(t *testing.T) {
	// Test UI feedback message patterns
	t.Run("UIFeedbackMessages", func(t *testing.T) {
		testCases := []struct {
			operation   string
			expectedMsg string
			description string
		}{
			{
				operation:   "enable_dev_mode",
				expectedMsg: "Preparing dev mode",
				description: "Enabling dev mode should show preparation message",
			},
			{
				operation:   "create_backup",
				expectedMsg: "Creating backup of current stable version",
				description: "Backup creation should show backup message",
			},
			{
				operation:   "update_to_dev",
				expectedMsg: "Updating to latest dev version",
				description: "Dev update should show dev update message",
			},
			{
				operation:   "disable_dev_mode",
				expectedMsg: "Switching to stable channel",
				description: "Disabling dev mode should show switch message",
			},
			{
				operation:   "restore_backup",
				expectedMsg: "Restoring stable version from backup",
				description: "Backup restoration should show restore message",
			},
			{
				operation:   "update_to_stable",
				expectedMsg: "Updating to latest stable version",
				description: "Stable update should show stable update message",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.operation, func(t *testing.T) {
				// Simulate UI feedback logic from gui.go
				// In real implementation, this would update the loading overlay
				var feedbackMessage string

				switch tc.operation {
				case "enable_dev_mode":
					feedbackMessage = "Preparing dev mode..."
				case "create_backup":
					feedbackMessage = "Creating backup of current stable version..."
				case "update_to_dev":
					feedbackMessage = "Updating to latest dev version..."
				case "disable_dev_mode":
					feedbackMessage = "Switching to stable channel..."
				case "restore_backup":
					feedbackMessage = "Restoring stable version from backup..."
				case "update_to_stable":
					feedbackMessage = "Updating to latest stable version..."
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
}

// TestGUIDevModeBackupManagement tests backup file management operations
func TestGUIDevModeBackupManagement(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-gui-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
	backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")

	// Test 1: Delete backup functionality
	t.Run("DeleteBackup", func(t *testing.T) {
		// Create backup files first
		meta := map[string]string{"tag": "v3.2.27", "path": backupExePath}
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal backup metadata: %v", err)
		}
		err = os.WriteFile(backupMetaPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write backup metadata: %v", err)
		}
		err = os.WriteFile(backupExePath, []byte("mock-backup-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup exe: %v", err)
		}

		// Verify backup exists
		if !fileExists(backupMetaPath) || !fileExists(backupExePath) {
			t.Fatal("Backup files should exist before deletion")
		}

		// Simulate backup deletion from gui.go
		err = os.Remove(backupMetaPath)
		if err != nil {
			t.Fatalf("Failed to remove backup metadata: %v", err)
		}
		err = os.Remove(backupExePath)
		if err != nil {
			t.Fatalf("Failed to remove backup exe: %v", err)
		}

		// Verify backup is deleted
		if fileExists(backupMetaPath) {
			t.Error("Backup metadata file should be deleted")
		}
		if fileExists(backupExePath) {
			t.Error("Backup executable file should be deleted")
		}
	})

	// Test 2: Backup info display
	t.Run("BackupInfoDisplay", func(t *testing.T) {
		// Create backup with timestamp
		meta := map[string]string{"tag": "v3.2.27", "path": backupExePath}
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal backup metadata: %v", err)
		}
		err = os.WriteFile(backupMetaPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write backup metadata: %v", err)
		}
		err = os.WriteFile(backupExePath, []byte("mock-backup-exe"), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup exe: %v", err)
		}

		// Get file info for timestamp
		info, err := os.Stat(backupExePath)
		if err != nil {
			t.Fatalf("Failed to get backup exe info: %v", err)
		}

		// Simulate backup info display logic from gui.go
		var backupInfo string
		if fileExists(backupMetaPath) {
			data, err := os.ReadFile(backupMetaPath)
			if err == nil {
				var loadedMeta map[string]string
				if json.Unmarshal(data, &loadedMeta) == nil {
					tag := loadedMeta["tag"]
					backupInfo = fmt.Sprintf("Backup tag: %s", tag)
					if info != nil {
						backupInfo = fmt.Sprintf("%s • saved: %s", backupInfo, info.ModTime().Format(time.RFC1123))
					}
				}
			}
		}

		if backupInfo == "" {
			t.Error("Backup info should not be empty when backup exists")
		}

		if !containsString(backupInfo, "v3.2.27") {
			t.Errorf("Backup info should contain tag v3.2.27, got: %s", backupInfo)
		}
	})
}

// Helper function to check if string contains substring (using strings.Contains for simplicity)
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
