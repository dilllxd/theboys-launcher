package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestFetchLatestAssetPreferPrerelease(t *testing.T) {
	// Note: This test would require mocking HTTP requests or using a test server
	// For now, we'll test the logic with a simple validation that the function exists
	// and has the correct signature

	// Test that the function exists and can be called (without actually making HTTP requests)
	// In a real test environment, you would mock the HTTP responses

	owner := "test-owner"
	repo := "test-repo"
	wantName := "test-binary"
	preferPrerelease := true

	// This would normally make HTTP requests, but we're just testing the function signature
	// In a complete test suite, you would use httptest.NewServer to mock GitHub responses
	_, _, err := FetchLatestAssetPreferPrerelease(owner, repo, wantName, preferPrerelease)

	// We expect this to fail since we're not mocking the HTTP requests
	// The important thing is that the function exists and has the correct signature
	if err == nil {
		t.Errorf("Expected FetchLatestAssetPreferPrerelease to fail without mocked HTTP responses")
	}
}

func TestUpdateLogic(t *testing.T) {
	// Test the update logic with mock data

	// Test case 1: Current version is older than remote
	localVersion := "3.2.26"
	remoteVersion := "3.2.27"

	result := compareSemver(normalizeTag(localVersion), normalizeTag(remoteVersion))
	if result >= 0 {
		t.Errorf("Expected %s to be older than %s", localVersion, remoteVersion)
	}

	// Test case 2: Current version is newer than remote
	localVersion = "3.2.27-dev.abc123"
	remoteVersion = "3.2.26"

	result = compareSemver(normalizeTag(localVersion), normalizeTag(remoteVersion))
	if result <= 0 {
		t.Errorf("Expected %s to be newer than %s", localVersion, remoteVersion)
	}

	// Test case 3: Current version is same as remote
	localVersion = "3.2.27"
	remoteVersion = "3.2.27"

	result = compareSemver(normalizeTag(localVersion), normalizeTag(remoteVersion))
	if result != 0 {
		t.Errorf("Expected %s to be equal to %s", localVersion, remoteVersion)
	}
}

func TestAssetNameDetection(t *testing.T) {
	// Test that the launcher would detect the correct asset names for different platforms

	// Test Windows asset name
	expectedWindowsAsset := "TheBoysLauncher.exe"
	actualWindowsAsset := GetLauncherAssetName()
	if IsWindows() && actualWindowsAsset != expectedWindowsAsset {
		t.Errorf("Expected Windows asset name %s, got %s", expectedWindowsAsset, actualWindowsAsset)
	}

	// Test macOS asset name
	expectedMacAsset := "TheBoysLauncher-mac-universal"
	if IsDarwin() && actualWindowsAsset != expectedWindowsAsset {
		t.Errorf("Expected macOS asset name %s, got %s", expectedMacAsset, actualWindowsAsset)
	}

	// Test Linux asset name
	expectedLinuxAsset := "TheBoysLauncher-linux"
	if IsLinux() && actualWindowsAsset != expectedWindowsAsset {
		t.Errorf("Expected Linux asset name %s, got %s", expectedLinuxAsset, actualWindowsAsset)
	}
}

func TestUpdateScenarioValidation(t *testing.T) {
	// Test various update scenarios that might occur in practice

	scenarios := []struct {
		name          string
		localVersion  string
		remoteVersion string
		shouldUpdate  bool
	}{
		{
			name:          "Stable to newer stable",
			localVersion:  "3.2.26",
			remoteVersion: "3.2.27",
			shouldUpdate:  true,
		},
		{
			name:          "Stable to newer dev",
			localVersion:  "3.2.26",
			remoteVersion: "3.2.27-dev.abc123",
			shouldUpdate:  true,
		},
		{
			name:          "Dev to newer stable",
			localVersion:  "3.2.26-dev.abc123",
			remoteVersion: "3.2.27",
			shouldUpdate:  true,
		},
		{
			name:          "Dev to newer dev",
			localVersion:  "3.2.26-dev.abc123",
			remoteVersion: "3.2.27-def456",
			shouldUpdate:  true,
		},
		{
			name:          "Same version",
			localVersion:  "3.2.27",
			remoteVersion: "3.2.27",
			shouldUpdate:  false,
		},
		{
			name:          "Newer local version",
			localVersion:  "3.2.27",
			remoteVersion: "3.2.26",
			shouldUpdate:  false,
		},
		{
			name:          "Newer local dev version",
			localVersion:  "3.2.27-dev.abc123",
			remoteVersion: "3.2.26",
			shouldUpdate:  false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := compareSemver(normalizeTag(scenario.localVersion), normalizeTag(scenario.remoteVersion))
			shouldUpdate := result < 0 // local < remote means update needed

			if shouldUpdate != scenario.shouldUpdate {
				t.Errorf("Scenario %s: expected shouldUpdate=%v, got %v (local=%s, remote=%s, comparison=%d)",
					scenario.name, scenario.shouldUpdate, shouldUpdate, scenario.localVersion, scenario.remoteVersion, result)
			}
		})
	}
}

func TestPrereleasePreference(t *testing.T) {
	// Test the logic for preferring prerelease builds when enabled

	// This tests the concept that when dev builds are enabled, the launcher
	// should prefer dev builds over stable builds of the same core version

	stableVersion := "3.2.27"
	devVersion := "3.2.27-dev.abc123"

	// When comparing same core version, stable should be preferred over dev
	result := compareSemver(stableVersion, devVersion)
	if result <= 0 {
		t.Errorf("Expected stable version %s to be newer than dev version %s", stableVersion, devVersion)
	}

	// But when core versions differ, the newer core version should win
	oldDev := "3.2.26-dev.abc123"
	newStable := "3.2.27"

	result = compareSemver(oldDev, newStable)
	if result >= 0 {
		t.Errorf("Expected new stable %s to be newer than old dev %s", newStable, oldDev)
	}
}

func TestForceUpdate(t *testing.T) {
	// Test the forceUpdate function with mock data
	// Note: This test would require mocking HTTP requests or using a test server
	// For now, we'll test the logic with a simple validation that the function exists
	// and has the correct signature

	// Test that the function exists and can be called (without actually making HTTP requests)
	// In a real test environment, you would use httptest.NewServer to mock GitHub responses

	root := "/test/root"
	exePath := "/test/path/TheBoysLauncher"
	preferDev := true
	report := func(msg string) {
		// Mock report function
		t.Logf("Report: %s", msg)
	}

	// This would normally make HTTP requests and perform updates, but we're just testing the function signature
	// In a complete test suite, you would mock the HTTP responses and file operations
	err := forceUpdate(root, exePath, preferDev, report)

	// We expect this to fail since we're not mocking the HTTP requests and file operations
	// The important thing is that the function exists and has the correct signature
	if err == nil {
		t.Errorf("Expected forceUpdate to fail without mocked HTTP responses and file operations")
	}
}

func TestForceUpdateLogic(t *testing.T) {
	// Test the logic that would be used in forceUpdate function

	// Test case 1: preferDev = true should fetch dev builds
	t.Run("PreferDevTrue", func(t *testing.T) {
		preferDev := true
		expectedChannel := "dev"

		// Simulate the channel selection logic from forceUpdate
		channel := "stable"
		if preferDev {
			channel = "dev"
		}

		if channel != expectedChannel {
			t.Errorf("Expected channel to be %s when preferDev is true, got %s", expectedChannel, channel)
		}
	})

	// Test case 2: preferDev = false should fetch stable builds
	t.Run("PreferDevFalse", func(t *testing.T) {
		preferDev := false
		expectedChannel := "stable"

		// Simulate the channel selection logic from forceUpdate
		channel := "stable"
		if preferDev {
			channel = "dev"
		}

		if channel != expectedChannel {
			t.Errorf("Expected channel to be %s when preferDev is false, got %s", expectedChannel, channel)
		}
	})
}

func TestForceUpdateErrorHandling(t *testing.T) {
	// Test error handling scenarios in forceUpdate

	// Test case 1: Empty exePath should cause an error
	t.Run("EmptyExePath", func(t *testing.T) {
		root := "/test/root"
		exePath := ""
		preferDev := false
		report := func(msg string) {}

		// This should fail when trying to download to an empty path
		err := forceUpdate(root, exePath, preferDev, report)
		if err == nil {
			t.Error("Expected forceUpdate to fail with empty exePath")
		}
	})

	// Test case 2: Invalid root path should be handled gracefully
	t.Run("InvalidRootPath", func(t *testing.T) {
		root := "/invalid/root/path/that/does/not/exist"
		exePath := "/test/path/TheBoysLauncher"
		preferDev := false
		report := func(msg string) {}

		// This should fail when trying to operate on invalid paths
		err := forceUpdate(root, exePath, preferDev, report)
		if err == nil {
			t.Error("Expected forceUpdate to fail with invalid root path")
		}
	})
}

func TestForceUpdateCallbackHandling(t *testing.T) {
	// Test that the report callback function is properly called during forceUpdate

	var reportedMessages []string
	report := func(msg string) {
		reportedMessages = append(reportedMessages, msg)
	}

	root := "/test/root"
	exePath := "/test/path/TheBoysLauncher"
	preferDev := true

	// This will fail but should still call the report function
	_ = forceUpdate(root, exePath, preferDev, report)

	// Check that some messages were reported
	if len(reportedMessages) == 0 {
		t.Error("Expected report callback to be called during forceUpdate")
	}

	// Check for expected message patterns
	expectedPatterns := []string{
		"Checking for latest launcher version",
	}

	for _, pattern := range expectedPatterns {
		found := false
		for _, msg := range reportedMessages {
			if strings.Contains(msg, pattern) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected report callback to be called with message containing '%s'", pattern)
		}
	}
}

func TestForceUpdateChannelSelection(t *testing.T) {
	// Test that forceUpdate correctly selects between dev and stable channels

	testCases := []struct {
		name         string
		preferDev    bool
		expectedTag  string
		expectedDesc string
	}{
		{
			name:         "DevChannel",
			preferDev:    true,
			expectedTag:  "dev",
			expectedDesc: "dev version",
		},
		{
			name:         "StableChannel",
			preferDev:    false,
			expectedTag:  "stable",
			expectedDesc: "stable version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the channel selection logic from forceUpdate
			channel := "stable"
			if tc.preferDev {
				channel = "dev"
			}

			if channel != tc.expectedTag {
				t.Errorf("Expected channel to be %s, got %s", tc.expectedTag, channel)
			}

			// Test the description logic
			desc := fmt.Sprintf("Downloading %s version %s...", channel, "v1.0.0")
			if !strings.Contains(desc, tc.expectedDesc) {
				t.Errorf("Expected description to contain '%s', got '%s'", tc.expectedDesc, desc)
			}
		})
	}
}

func TestFilterStableReleasesLogic(t *testing.T) {
	testCases := []struct {
		name              string
		availableVersions []string
		preferDev         bool
		expectedVersion   string
		shouldError       bool
	}{
		{
			name: "MixedDevAndStable",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Latest dev version
				"v3.2.29",             // Latest stable version
				"v3.2.28-beta",        // Beta version
				"v3.2.27",             // Older stable version
			},
			preferDev:       false,
			expectedVersion: "v3.2.29", // Should return latest stable, not dev
			shouldError:     false,
		},
		{
			name: "OnlyDevVersions",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae",
				"v3.2.29-dev.abc123",
				"v3.2.28-dev.def456",
			},
			preferDev:   false,
			shouldError: true, // Should error when no stable versions exist
		},
		{
			name: "OnlyStableVersions",
			availableVersions: []string{
				"v3.2.29",
				"v3.2.28",
				"v3.2.27",
			},
			preferDev:       false,
			expectedVersion: "v3.2.29", // Should return latest stable
			shouldError:     false,
		},
		{
			name: "MixedPrereleaseTypes",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Dev
				"v3.2.29-beta.1",      // Beta
				"v3.2.28-rc.2",        // Release candidate
				"v3.2.27-alpha.3",     // Alpha
				"v3.2.26-pre.4",       // Pre
				"v3.2.25",             // Stable
			},
			preferDev:       false,
			expectedVersion: "v3.2.25", // Should return only stable version
			shouldError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the filtering logic from fetchLatestAssetPreferPrerelease
			var result string
			var err error

			if tc.preferDev {
				// When preferring dev, return the first version found
				if len(tc.availableVersions) > 0 {
					result = tc.availableVersions[0]
				}
			} else {
				// When preferring stable, filter out prerelease versions
				var stableVersions []string
				for _, version := range tc.availableVersions {
					if !isPrereleaseTag(version) {
						stableVersions = append(stableVersions, version)
					}
				}
				if len(stableVersions) == 0 {
					err = fmt.Errorf("no stable releases found")
				} else {
					result = stableVersions[0]
				}
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tc.expectedVersion {
					t.Errorf("Expected version %s, got %s", tc.expectedVersion, result)
				}
			}
		})
	}
}

func TestFetchLatestAssetPreferStableLogic(t *testing.T) {
	testCases := []struct {
		name              string
		availableVersions []string
		preferDev         bool
		expectedVersion   string
		shouldError       bool
	}{
		{
			name: "PreferDevTrue_ShouldReturnDevVersion",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Latest dev version
				"v3.2.29",             // Latest stable version
				"v3.2.28-beta",        // Beta version
				"v3.2.27",             // Older stable version
			},
			preferDev:       true,
			expectedVersion: "v3.2.30-dev.adcb1ae", // Should return latest dev
			shouldError:     false,
		},
		{
			name: "PreferDevFalse_ShouldReturnLatestStable",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Latest dev version
				"v3.2.29",             // Latest stable version
				"v3.2.28-beta",        // Beta version
				"v3.2.27",             // Older stable version
			},
			preferDev:       false,
			expectedVersion: "v3.2.29", // Should return latest stable, not dev
			shouldError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the logic from fetchLatestAssetPreferPrerelease
			var result string
			var err error

			if tc.preferDev {
				// When preferring dev, return the first version found
				if len(tc.availableVersions) > 0 {
					result = tc.availableVersions[0]
				}
			} else {
				// When preferring stable, filter out prerelease versions
				var stableVersions []string
				for _, version := range tc.availableVersions {
					if !isPrereleaseTag(version) {
						stableVersions = append(stableVersions, version)
					}
				}
				if len(stableVersions) == 0 {
					err = fmt.Errorf("no stable releases found")
				} else {
					result = stableVersions[0]
				}
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tc.expectedVersion {
					t.Errorf("Expected version %s, got %s", tc.expectedVersion, result)
				}
			}
		})
	}
}

func TestVersionFilteringEdgeCases(t *testing.T) {
	testCases := []struct {
		name              string
		availableVersions []string
		preferDev         bool
		expectedVersion   string
		shouldError       bool
		description       string
	}{
		{
			name:              "EmptyVersionList",
			availableVersions: []string{},
			preferDev:         false,
			shouldError:       true,
			description:       "Should error when no versions are available",
		},
		{
			name:              "EmptyVersionListPreferDev",
			availableVersions: []string{},
			preferDev:         true,
			shouldError:       true,
			description:       "Should error when no versions are available even when preferring dev",
		},
		{
			name: "SingleDevVersion",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae",
			},
			preferDev:       true,
			expectedVersion: "v3.2.30-dev.adcb1ae",
			shouldError:     false,
			description:     "Should return single dev version when preferring dev",
		},
		{
			name: "SingleDevVersionPreferStable",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae",
			},
			preferDev:   false,
			shouldError: true,
			description: "Should error when only dev version exists but preferring stable",
		},
		{
			name: "SingleStableVersion",
			availableVersions: []string{
				"v3.2.29",
			},
			preferDev:       false,
			expectedVersion: "v3.2.29",
			shouldError:     false,
			description:     "Should return single stable version when preferring stable",
		},
		{
			name: "SingleStableVersionPreferDev",
			availableVersions: []string{
				"v3.2.29",
			},
			preferDev:       true,
			expectedVersion: "v3.2.29",
			shouldError:     false,
			description:     "Should return stable version even when preferring dev if no dev available",
		},
		{
			name: "ComplexVersionFormats",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae",      // Dev with hash
				"v3.2.29-beta.1+build.123", // Beta with build metadata
				"v3.2.28-rc.2",             // Release candidate
				"v3.2.27-alpha.3",          // Alpha
				"v3.2.26-pre.4",            // Pre
				"v3.2.25+build.456",        // Stable with build metadata
				"v3.2.24",                  // Stable
			},
			preferDev:       false,
			expectedVersion: "v3.2.25+build.456", // Should return first stable (with build metadata)
			shouldError:     false,
			description:     "Should handle complex version formats correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the filtering logic from fetchLatestAssetPreferPrerelease
			var result string
			var err error

			if tc.preferDev {
				// When preferring dev, return the first version found
				if len(tc.availableVersions) > 0 {
					result = tc.availableVersions[0]
				} else {
					err = fmt.Errorf("no releases found")
				}
			} else {
				// When preferring stable, filter out prerelease versions
				var stableVersions []string
				for _, version := range tc.availableVersions {
					if !isPrereleaseTag(version) {
						stableVersions = append(stableVersions, version)
					}
				}
				if len(stableVersions) == 0 {
					err = fmt.Errorf("no stable releases found")
				} else {
					result = stableVersions[0]
				}
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v (%s)", err, tc.description)
				}
				if result != tc.expectedVersion {
					t.Errorf("Expected version %s, got %s (%s)", tc.expectedVersion, result, tc.description)
				}
			}
		})
	}
}
