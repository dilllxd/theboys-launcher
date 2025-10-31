package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetCFBundleExecutable tests the Info.plist parsing for CFBundleExecutable
func TestGetCFBundleExecutable(t *testing.T) {
	// Create a temporary test app bundle structure
	tempDir := t.TempDir()
	appBundlePath := filepath.Join(tempDir, "TestApp.app")
	contentsDir := filepath.Join(appBundlePath, "Contents")

	if err := os.MkdirAll(contentsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a sample Info.plist with CFBundleExecutable
	infoPlistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>TestExecutable</string>
	<key>CFBundleIdentifier</key>
	<string>com.example.testapp</string>
	<key>CFBundleName</key>
	<string>TestApp</string>
</dict>
</plist>`

	infoPlistPath := filepath.Join(contentsDir, "Info.plist")
	if err := os.WriteFile(infoPlistPath, []byte(infoPlistContent), 0644); err != nil {
		t.Fatalf("Failed to create Info.plist: %v", err)
	}

	// Test getCFBundleExecutable
	executable, err := getCFBundleExecutable(appBundlePath)
	if err != nil {
		t.Fatalf("getCFBundleExecutable failed: %v", err)
	}

	if executable != "TestExecutable" {
		t.Errorf("Expected executable name 'TestExecutable', got '%s'", executable)
	}
}

// TestGetCFBundleExecutableNotFound tests error handling when CFBundleExecutable is missing
func TestGetCFBundleExecutableNotFound(t *testing.T) {
	tempDir := t.TempDir()
	appBundlePath := filepath.Join(tempDir, "TestApp.app")
	contentsDir := filepath.Join(appBundlePath, "Contents")

	if err := os.MkdirAll(contentsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create an Info.plist without CFBundleExecutable
	infoPlistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.example.testapp</string>
</dict>
</plist>`

	infoPlistPath := filepath.Join(contentsDir, "Info.plist")
	if err := os.WriteFile(infoPlistPath, []byte(infoPlistContent), 0644); err != nil {
		t.Fatalf("Failed to create Info.plist: %v", err)
	}

	// Test getCFBundleExecutable - should return error
	_, err := getCFBundleExecutable(appBundlePath)
	if err == nil {
		t.Error("Expected error when CFBundleExecutable is missing, got nil")
	}
}

// TestGetCFBundleExecutableFileNotFound tests error handling when Info.plist doesn't exist
func TestGetCFBundleExecutableFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	appBundlePath := filepath.Join(tempDir, "NonExistentApp.app")

	// Test getCFBundleExecutable - should return error
	_, err := getCFBundleExecutable(appBundlePath)
	if err == nil {
		t.Error("Expected error when Info.plist doesn't exist, got nil")
	}
}
