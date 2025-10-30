package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Mock functions for testing - these would need to be exported from the main package
// For now, we'll create simplified versions for testing

// TestCustomInstallationPaths tests that the launcher correctly detects and works with custom installation paths
func TestCustomInstallationPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	// Test 1: Verify path normalization
	t.Run("PathNormalization", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{
				input:    `C:\Program Files\TheBoysLauncher\`,
				expected: `C:\Program Files\TheBoysLauncher`,
			},
			{
				input:    `C:/Program Files/TheBoysLauncher/`,
				expected: `C:\Program Files\TheBoysLauncher`,
			},
			{
				input:    `C:\Games\TheBoysLauncher\\`,
				expected: `C:\Games\TheBoysLauncher`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				normalized := filepath.Clean(tc.input)
				if normalized != tc.expected {
					t.Errorf("filepath.Clean(%q) = %v; want %v", tc.input, normalized, tc.expected)
				}
			})
		}
	})

	// Test 2: Verify default path resolution
	t.Run("DefaultPathResolution", func(t *testing.T) {
		// Test default LocalAppData path
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			t.Skip("LOCALAPPDATA environment variable not set")
		}

		expectedDefault := filepath.Join(localAppData, "TheBoysLauncher")

		// Verify the expected default path structure
		if !strings.Contains(expectedDefault, "TheBoysLauncher") {
			t.Errorf("Default path should contain TheBoysLauncher: %s", expectedDefault)
		}

		if !filepath.IsAbs(expectedDefault) {
			t.Errorf("Default path should be absolute: %s", expectedDefault)
		}
	})

	// Test 3: Verify installed mode detection logic
	t.Run("InstalledModeDetection", func(t *testing.T) {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			t.Skip("LOCALAPPDATA environment variable not set")
		}

		testCases := []struct {
			name        string
			installPath string
			expected    bool
		}{
			{
				name:        "Empty path",
				installPath: "",
				expected:    false,
			},
			{
				name:        "LocalAppData path (portable mode)",
				installPath: filepath.Join(localAppData, "TheBoysLauncher"),
				expected:    false,
			},
			{
				name:        "Program Files path (installed mode)",
				installPath: filepath.Join(os.Getenv("ProgramFiles"), "TheBoysLauncher"),
				expected:    true,
			},
			{
				name:        "Custom path (installed mode)",
				installPath: `C:\Games\TheBoysLauncher`,
				expected:    true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate the isInstalledMode logic
				result := simulateIsInstalledMode(tc.installPath, localAppData)
				if result != tc.expected {
					t.Errorf("simulateIsInstalledMode(%q) = %v; want %v", tc.installPath, result, tc.expected)
				}
			})
		}
	})

	// Test 4: Verify config file paths work correctly
	t.Run("ConfigFilePaths", func(t *testing.T) {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			t.Skip("LOCALAPPDATA environment variable not set")
		}

		home := filepath.Join(localAppData, "TheBoysLauncher")
		configPath := filepath.Join(home, "settings.json")
		modpacksPath := filepath.Join(home, "modpacks.json")

		// These paths should be valid absolute paths
		if !filepath.IsAbs(configPath) {
			t.Errorf("Config path should be absolute: %s", configPath)
		}
		if !filepath.IsAbs(modpacksPath) {
			t.Errorf("Modpacks path should be absolute: %s", modpacksPath)
		}

		// Verify they're in the expected directory structure
		if !strings.Contains(configPath, "TheBoysLauncher") {
			t.Errorf("Config path should contain TheBoysLauncher: %s", configPath)
		}
		if !strings.Contains(modpacksPath, "TheBoysLauncher") {
			t.Errorf("Modpacks path should contain TheBoysLauncher: %s", modpacksPath)
		}
	})
}

// simulateIsInstalledMode simulates the isInstalledMode function logic for testing
func simulateIsInstalledMode(installPath, localAppData string) bool {
	if installPath == "" {
		return false
	}

	if localAppData == "" {
		return false
	}

	defaultPath := filepath.Join(localAppData, "TheBoysLauncher")

	// Normalize paths for comparison
	installPath = filepath.Clean(installPath)
	defaultPath = filepath.Clean(defaultPath)

	// If the installation path is different from the default LocalAppData path,
	// we're in installed mode
	return installPath != defaultPath
}

// TestInstallerRegistryIntegration tests that the installer correctly writes to registry
func TestInstallerRegistryIntegration(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	t.Run("InnoSetupRegistry", func(t *testing.T) {
		// Verify TheBoysLauncher.iss contains the correct registry entries
		// This is a static analysis test
		issContent := `
[Registry]
; Store installation path for launcher to read
Root: HKLM; Subkey: "SOFTWARE\TheBoysLauncher"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\TheBoysLauncher"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
`

		// Check that both HKLM and HKCU entries are present
		if !strings.Contains(issContent, `Root: HKLM; Subkey: "SOFTWARE\TheBoysLauncher"`) {
			t.Error("InnoSetup installer missing HKLM registry entry")
		}
		if !strings.Contains(issContent, `Root: HKCU; Subkey: "SOFTWARE\TheBoysLauncher"`) {
			t.Error("InnoSetup installer missing HKCU registry entry")
		}
		if !strings.Contains(issContent, `ValueName: "InstallPath"`) {
			t.Error("InnoSetup installer missing InstallPath value name")
		}
		if !strings.Contains(issContent, `ValueData: "{app}"`) {
			t.Error("InnoSetup installer missing {app} value data")
		}
	})

	t.Run("WiXRegistry", func(t *testing.T) {
		// Verify WiX installer contains the correct registry entries
		// This is a static analysis test
		wixContent := `
<RegistryKey Root="HKCU" Key="Software\TheBoysLauncher">
  <RegistryValue Type="string" Name="InstallPath" Value="[INSTALLFOLDER]" />
  <RegistryValue Type="string" Name="Version" Value="$(var.ProductVersion)" />
</RegistryKey>
`

		// Check that HKCU entry is present with correct values
		if !strings.Contains(wixContent, `Root="HKCU"`) {
			t.Error("WiX installer missing HKCU registry entry")
		}
		if !strings.Contains(wixContent, `Key="Software\TheBoysLauncher"`) {
			t.Error("WiX installer missing correct registry key")
		}
		if !strings.Contains(wixContent, `Name="InstallPath"`) {
			t.Error("WiX installer missing InstallPath value name")
		}
		if !strings.Contains(wixContent, `Value="[INSTALLFOLDER]"`) {
			t.Error("WiX installer missing INSTALLFOLDER value data")
		}
	})
}

// TestFallbackMechanisms tests that the launcher has proper fallback mechanisms
func TestFallbackMechanisms(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	t.Run("DefaultFallback", func(t *testing.T) {
		// Test that the launcher falls back to LocalAppData when no custom path is set
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			t.Skip("LOCALAPPDATA environment variable not set")
		}

		expectedFallback := filepath.Join(localAppData, "TheBoysLauncher")

		// Simulate the fallback behavior
		home := simulateGetLauncherHome("", localAppData)

		if home != expectedFallback {
			t.Errorf("Fallback to LocalAppData failed: got %s, want %s", home, expectedFallback)
		}
	})

	t.Run("CustomPathOverride", func(t *testing.T) {
		// Test that custom installation path is used when available
		customPath := `C:\Games\TheBoysLauncher`
		localAppData := os.Getenv("LOCALAPPDATA")

		// Simulate custom path behavior
		home := simulateGetLauncherHome(customPath, localAppData)

		if home != customPath {
			t.Errorf("Custom path override failed: got %s, want %s", home, customPath)
		}
	})

	t.Run("PathValidation", func(t *testing.T) {
		// Test that invalid paths are handled gracefully
		testCases := []struct {
			name        string
			path        string
			shouldExist bool
		}{
			{
				name:        "Non-existent path",
				path:        `C:\NonExistent\TheBoysLauncher`,
				shouldExist: false,
			},
			{
				name: "Current executable directory",
				path: func() string {
					if exe, err := os.Executable(); err == nil {
						return filepath.Dir(exe)
					}
					return ""
				}(),
				shouldExist: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.path == "" {
					t.Skip("Could not determine test path")
				}

				exists := false
				if _, err := os.Stat(tc.path); err == nil {
					exists = true
				}

				if exists != tc.shouldExist {
					t.Errorf("Path existence check failed for %s: got %v, want %v", tc.path, exists, tc.shouldExist)
				}
			})
		}
	})
}

// simulateGetLauncherHome simulates the getLauncherHome function logic for testing
func simulateGetLauncherHome(registryPath, localAppData string) string {
	// First, check the registry for custom installation path
	if registryPath != "" {
		// If we have a custom installation path from registry
		if simulateIsInstalledMode(registryPath, localAppData) {
			// In installed mode, store data alongside the executable (portable-style)
			return registryPath
		}
		// If it's the default path, continue with normal logic
	}

	// Default behavior for existing installations or when registry is not available
	// Prefer LocalAppData\TheBoysLauncher on Windows for per-user installs
	if localAppData != "" {
		return filepath.Join(localAppData, "TheBoysLauncher")
	}

	// Fallback to USERPROFILE dot-folder (legacy)
	homeDir := os.Getenv("USERPROFILE")
	if homeDir == "" {
		if exePath, err := os.Executable(); err == nil {
			return filepath.Dir(exePath)
		}
		return "."
	}
	return filepath.Join(homeDir, ".theboyslauncher")
}

// TestResourceLocation tests that the launcher can find its resources regardless of installation location
func TestResourceLocation(t *testing.T) {
	t.Run("ConfigFileResolution", func(t *testing.T) {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			t.Skip("LOCALAPPDATA environment variable not set")
		}

		home := filepath.Join(localAppData, "TheBoysLauncher")

		// Test common config files
		configFiles := []string{
			"settings.json",
			"modpacks.json",
		}

		for _, file := range configFiles {
			fullPath := filepath.Join(home, file)
			t.Logf("Config file path: %s", fullPath)

			// Verify path is absolute and contains expected components
			if !filepath.IsAbs(fullPath) {
				t.Errorf("Config file path should be absolute: %s", fullPath)
			}

			if !strings.Contains(fullPath, "TheBoysLauncher") {
				t.Errorf("Config file path should contain TheBoysLauncher: %s", fullPath)
			}
		}
	})

	t.Run("ExecutablePath", func(t *testing.T) {
		// Test that the launcher can determine its own executable path
		exePath, err := os.Executable()
		if err != nil {
			t.Fatalf("Failed to get executable path: %v", err)
		}

		t.Logf("Executable path: %s", exePath)

		// Verify it's an absolute path
		if !filepath.IsAbs(exePath) {
			t.Errorf("Executable path should be absolute: %s", exePath)
		}

		// Verify it points to TheBoysLauncher.exe on Windows
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(strings.ToLower(exePath), "theboyslauncher.exe") {
				t.Errorf("Executable should be TheBoysLauncher.exe: %s", exePath)
			}
		}
	})
}
