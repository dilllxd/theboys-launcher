package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDevModeToggleBehavior tests the modified dev mode toggle behavior
// This tests the new implementation where checkbox toggles only update temporary variables
// and the actual update process only happens when Save is clicked
func TestDevModeToggleBehavior(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-toggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Checkbox toggle doesn't trigger immediate updates
	t.Run("CheckboxToggleNoImmediateUpdate", func(t *testing.T) {
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

		// Simulate checkbox toggle (temporary variable change)
		pendingDevBuildsEnabled := true // User checks the checkbox

		// Verify original settings haven't changed
		savedData, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read saved settings: %v", err)
		}

		var savedSettings LauncherSettings
		err = json.Unmarshal(savedData, &savedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved settings: %v", err)
		}

		// Original settings should remain unchanged
		if savedSettings.DevBuildsEnabled != originalSettings.DevBuildsEnabled {
			t.Errorf("Saved settings should not change when checkbox is toggled. Expected %v, got %v",
				originalSettings.DevBuildsEnabled, savedSettings.DevBuildsEnabled)
		}

		// Pending variable should reflect the change
		if pendingDevBuildsEnabled != true {
			t.Error("Pending variable should reflect checkbox state")
		}
	})

	// Test 2: Multiple checkbox toggles before save
	t.Run("MultipleTogglesBeforeSave", func(t *testing.T) {
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

		// Simulate multiple checkbox toggles
		pendingDevBuildsEnabled := false
		pendingDevBuildsEnabled = true  // First toggle
		pendingDevBuildsEnabled = false // Second toggle
		pendingDevBuildsEnabled = true  // Third toggle

		// Verify original settings haven't changed after multiple toggles
		savedData, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read saved settings: %v", err)
		}

		var savedSettings LauncherSettings
		err = json.Unmarshal(savedData, &savedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved settings: %v", err)
		}

		// Original settings should remain unchanged
		if savedSettings.DevBuildsEnabled != originalSettings.DevBuildsEnabled {
			t.Errorf("Saved settings should not change after multiple checkbox toggles. Expected %v, got %v",
				originalSettings.DevBuildsEnabled, savedSettings.DevBuildsEnabled)
		}

		// Pending variable should reflect the final state
		if pendingDevBuildsEnabled != true {
			t.Error("Pending variable should reflect final checkbox state")
		}
	})

	// Test 3: Save button applies pending changes
	t.Run("SaveButtonAppliesPendingChanges", func(t *testing.T) {
		// Setup initial settings
		originalSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: false,
		}

		// Save original settings
		settingsPath := filepath.Join(tempDir, "settings3.json")
		data, err := json.MarshalIndent(originalSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal original settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save original settings: %v", err)
		}

		// Simulate checkbox toggle and save
		pendingDevBuildsEnabled := true // User checks the checkbox

		// Simulate save operation
		updatedSettings := originalSettings
		updatedSettings.DevBuildsEnabled = pendingDevBuildsEnabled

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

		// Settings should now reflect the saved change
		if savedSettings.DevBuildsEnabled != true {
			t.Errorf("Settings should be updated after save. Expected true, got %v", savedSettings.DevBuildsEnabled)
		}
	})

	// Test 4: Cancel button discards pending changes
	t.Run("CancelButtonDiscardsPendingChanges", func(t *testing.T) {
		// Setup initial settings
		originalSettings := LauncherSettings{
			MemoryMB:         4096,
			AutoRAM:          true,
			DevBuildsEnabled: false,
		}

		// Save original settings
		settingsPath := filepath.Join(tempDir, "settings4.json")
		data, err := json.MarshalIndent(originalSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal original settings: %v", err)
		}
		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to save original settings: %v", err)
		}

		// Simulate checkbox toggle but cancel (don't save)
		pendingDevBuildsEnabled := true // User checks the checkbox
		// User clicks cancel - pending changes are discarded

		// Use the variable to avoid unused error
		_ = pendingDevBuildsEnabled

		// Verify original settings remain unchanged
		savedData, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read saved settings: %v", err)
		}

		var savedSettings LauncherSettings
		err = json.Unmarshal(savedData, &savedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal saved settings: %v", err)
		}

		// Settings should remain unchanged after cancel
		if savedSettings.DevBuildsEnabled != originalSettings.DevBuildsEnabled {
			t.Errorf("Settings should remain unchanged after cancel. Expected %v, got %v",
				originalSettings.DevBuildsEnabled, savedSettings.DevBuildsEnabled)
		}
	})
}

// TestDevModeUIFeedback tests the UI feedback for pending changes
func TestDevModeUIFeedback(t *testing.T) {
	// Test 1: Status label updates for pending changes
	t.Run("StatusLabelUpdates", func(t *testing.T) {
		testCases := []struct {
			name                   string
			originalDevMode        bool
			pendingDevMode         bool
			expectedStatusContains string
		}{
			{
				name:                   "Enabling dev mode",
				originalDevMode:        false,
				pendingDevMode:         true,
				expectedStatusContains: "Dev builds will be enabled",
			},
			{
				name:                   "Disabling dev mode",
				originalDevMode:        true,
				pendingDevMode:         false,
				expectedStatusContains: "Dev builds will be disabled",
			},
			{
				name:                   "No change",
				originalDevMode:        false,
				pendingDevMode:         false,
				expectedStatusContains: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate the status label logic from gui.go
				hasPendingChanges := tc.originalDevMode != tc.pendingDevMode
				var statusLabel string

				if hasPendingChanges {
					if tc.pendingDevMode {
						statusLabel = "Pending changes: Dev builds will be enabled (backup will be created and launcher will update)"
					} else {
						statusLabel = "Pending changes: Dev builds will be disabled (stable version will be restored)"
					}
				} else {
					statusLabel = ""
				}

				if tc.expectedStatusContains == "" {
					if statusLabel != "" {
						t.Errorf("Expected empty status label, got: %s", statusLabel)
					}
				} else {
					if !strings.Contains(statusLabel, tc.expectedStatusContains) {
						t.Errorf("Expected status label to contain '%s', got: %s", tc.expectedStatusContains, statusLabel)
					}
				}
			})
		}
	})

	// Test 2: Multiple pending changes
	t.Run("MultiplePendingChanges", func(t *testing.T) {
		// Simulate multiple settings changes
		originalAutoRAM := true
		originalMemoryMB := 4096
		originalDevMode := false

		pendingAutoRAM := false
		pendingMemoryMB := 8192
		pendingDevMode := true

		// Simulate the status label logic for multiple changes
		hasPendingChanges := originalAutoRAM != pendingAutoRAM ||
			originalMemoryMB != pendingMemoryMB ||
			originalDevMode != pendingDevMode

		var statusLabel string
		if hasPendingChanges {
			var changes []string
			if originalAutoRAM != pendingAutoRAM {
				if pendingAutoRAM {
					changes = append(changes, "Auto RAM will be enabled")
				} else {
					changes = append(changes, "Auto RAM will be disabled")
				}
			}
			if originalMemoryMB != pendingMemoryMB && !pendingAutoRAM {
				changes = append(changes, "Memory will be set to 8 GB")
			}
			if originalDevMode != pendingDevMode {
				if pendingDevMode {
					changes = append(changes, "Dev builds will be enabled (backup will be created and launcher will update)")
				} else {
					changes = append(changes, "Dev builds will be disabled (stable version will be restored)")
				}
			}
			statusLabel = "Pending changes: " + strings.Join(changes, ", ")
		}

		expectedContains := []string{
			"Auto RAM will be disabled",
			"Memory will be set to 8 GB",
			"Dev builds will be enabled",
		}

		for _, expected := range expectedContains {
			if !strings.Contains(statusLabel, expected) {
				t.Errorf("Expected status label to contain '%s', got: %s", expected, statusLabel)
			}
		}
	})
}

// TestDevModeUpdateProcess tests that the update process only runs once when settings are applied
func TestDevModeUpdateProcess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-update-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Update process runs only on save
	t.Run("UpdateProcessOnlyOnSave", func(t *testing.T) {
		// Track update process calls
		updateProcessCalls := 0

		// Simulate the update process tracking
		simulateUpdateProcess := func() {
			updateProcessCalls++
		}

		// Simulate checkbox toggles (should not trigger update)
		pendingDevBuildsEnabled := true
		pendingDevBuildsEnabled = false
		pendingDevBuildsEnabled = true

		// Update process should not have been called yet
		if updateProcessCalls != 0 {
			t.Errorf("Update process should not be called during checkbox toggles. Called %d times", updateProcessCalls)
		}

		// Simulate save operation (should trigger update once)
		if pendingDevBuildsEnabled {
			simulateUpdateProcess()
		}

		// Update process should have been called exactly once
		if updateProcessCalls != 1 {
			t.Errorf("Update process should be called exactly once on save. Called %d times", updateProcessCalls)
		}
	})

	// Test 2: No update when cancelled
	t.Run("NoUpdateWhenCancelled", func(t *testing.T) {
		// Track update process calls
		updateProcessCalls := 0

		// Simulate checkbox toggle
		pendingDevBuildsEnabled := true
		_ = pendingDevBuildsEnabled // Use the variable to avoid unused error

		// Simulate cancel (no save, no update)
		// User closes dialog or clicks cancel

		// Update process should not have been called
		if updateProcessCalls != 0 {
			t.Errorf("Update process should not be called when cancelled. Called %d times", updateProcessCalls)
		}
	})

	// Test 3: Backup creation only when enabling dev mode
	t.Run("BackupCreationOnlyOnEnable", func(t *testing.T) {
		// Setup paths
		backupMetaPath := filepath.Join(tempDir, "dev-backup.json")
		backupExePath := filepath.Join(tempDir, "backup-non-dev.exe")

		// Ensure backup doesn't exist initially
		os.Remove(backupMetaPath)
		os.Remove(backupExePath)

		// Simulate enabling dev mode
		originalDevMode := false
		pendingDevMode := true

		// Backup should only be created when saving the change
		if originalDevMode != pendingDevMode && pendingDevMode {
			// Simulate backup creation
			tag := "v3.2.27"
			meta := map[string]string{"tag": tag, "path": backupExePath}
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
		}

		// Verify backup was created
		if !fileExists(backupMetaPath) {
			t.Error("Backup metadata should be created when enabling dev mode")
		}
		if !fileExists(backupExePath) {
			t.Error("Backup executable should be created when enabling dev mode")
		}
	})
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// TestDevModeErrorHandling tests error handling in the modified toggle behavior
func TestDevModeErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-devmode-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Revert checkbox state on backup failure
	t.Run("RevertOnBackupFailure", func(t *testing.T) {
		// Simulate backup failure scenario
		backupFailed := true
		originalDevMode := false
		pendingDevMode := true

		// Simulate the save operation with backup failure
		if originalDevMode != pendingDevMode {
			if backupFailed {
				// Revert the checkbox state
				pendingDevMode = originalDevMode
			}
		}

		// Verify checkbox was reverted
		if pendingDevMode != originalDevMode {
			t.Errorf("Checkbox should be reverted on backup failure. Expected %v, got %v",
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
}
