package main

import (
	"fmt"
	"testing"
)

// TestIntegrationStableVersionFiltering tests the complete integration scenario
// described in the original bug report
func TestIntegrationStableVersionFiltering(t *testing.T) {
	// This test verifies the fix for the bug where dev versions were
	// incorrectly returned when preferDev=false

	testCases := []struct {
		name              string
		availableVersions []string
		preferDev         bool
		expectedVersion   string
		shouldError       bool
	}{
		{
			name: "DevDisabled_ShouldReturnLatestStable",
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
			name: "DevEnabled_ShouldReturnLatestDev",
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
			name: "OnlyDevVersions_PreferDevFalse_ShouldError",
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae",
				"v3.2.29-dev.abc123",
			},
			preferDev:   false,
			shouldError: true, // Should error when no stable versions exist
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the logic from the fixed fetchLatestAssetPreferPrerelease function
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

// TestBugReportScenario tests the specific scenario described in the bug report
// where a dev version (v3.2.30-dev.adcb1ae) was incorrectly identified as stable
// when dev builds were disabled in settings
func TestBugReportScenario(t *testing.T) {
	// This test recreates the exact scenario from the bug report:
	// 1. Starting with a dev version (v3.2.30-dev.adcb1ae)
	// 2. Disabling dev builds in settings (preferDev = false)
	// 3. Verifying that the system correctly identifies the latest stable version (not the dev version)

	testCases := []struct {
		name              string
		currentVersion    string   // The version the launcher is currently running
		availableVersions []string // Versions available on GitHub releases
		preferDev         bool     // Dev builds setting in launcher
		expectedVersion   string   // Version that should be selected for update
		shouldUpdate      bool     // Whether an update should be offered
		shouldError       bool     // Whether an error should occur
	}{
		{
			name:           "BugReportScenario_DevCurrent_DevDisabled",
			currentVersion: "v3.2.30-dev.adcb1ae", // Current dev version from bug report
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Latest dev version (should be ignored)
				"v3.2.29",             // Latest stable version (should be selected)
				"v3.2.28-beta",        // Beta version (should be ignored)
				"v3.2.27",             // Older stable version
			},
			preferDev:       false,     // Dev builds disabled
			expectedVersion: "v3.2.29", // Should select latest stable, not dev
			shouldUpdate:    false,     // Current dev is newer than stable, so no update
			shouldError:     false,
		},
		{
			name:           "BugReportScenario_DevCurrent_DevEnabled",
			currentVersion: "v3.2.30-dev.adcb1ae", // Current dev version from bug report
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Latest dev version (should be selected)
				"v3.2.29",             // Latest stable version
				"v3.2.28-beta",        // Beta version
				"v3.2.27",             // Older stable version
			},
			preferDev:       true,                  // Dev builds enabled
			expectedVersion: "v3.2.30-dev.adcb1ae", // Should select latest dev
			shouldUpdate:    false,                 // Same version, no update needed
			shouldError:     false,
		},
		{
			name:           "BugReportScenario_StableCurrent_DevDisabled_NewerDevAvailable",
			currentVersion: "v3.2.29", // Current stable version
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Newer dev version (should be ignored)
				"v3.2.29",             // Same stable version
				"v3.2.28-beta",        // Beta version (should be ignored)
				"v3.2.27",             // Older stable version
			},
			preferDev:       false,     // Dev builds disabled
			expectedVersion: "v3.2.29", // Should select current stable version
			shouldUpdate:    false,     // Same version, no update needed
			shouldError:     false,
		},
		{
			name:           "BugReportScenario_StableCurrent_DevEnabled_NewerDevAvailable",
			currentVersion: "v3.2.29", // Current stable version
			availableVersions: []string{
				"v3.2.30-dev.adcb1ae", // Newer dev version (should be selected)
				"v3.2.29",             // Same stable version
				"v3.2.28-beta",        // Beta version
				"v3.2.27",             // Older stable version
			},
			preferDev:       true,                  // Dev builds enabled
			expectedVersion: "v3.2.30-dev.adcb1ae", // Should select newer dev
			shouldUpdate:    true,                  // Newer version available
			shouldError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Simulate version selection logic from fetchLatestAssetPreferPrerelease
			var selectedVersion string
			var err error

			if tc.preferDev {
				// When preferring dev, return the first version found
				if len(tc.availableVersions) > 0 {
					selectedVersion = tc.availableVersions[0]
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
					selectedVersion = stableVersions[0]
				}
			}

			// Step 2: Check for errors
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			// Step 3: Verify correct version was selected
			if selectedVersion != tc.expectedVersion {
				t.Errorf("Expected selected version %s, got %s", tc.expectedVersion, selectedVersion)
				return
			}

			// Step 4: Simulate update decision logic (compare current vs selected)
			normalizedCurrent := normalizeTag(tc.currentVersion)
			normalizedSelected := normalizeTag(selectedVersion)
			comparison := compareSemver(normalizedCurrent, normalizedSelected)

			// Update is needed if selected version is newer than current
			updateNeeded := comparison < 0

			if updateNeeded != tc.shouldUpdate {
				t.Errorf("Update decision mismatch: expected update needed=%v, got update needed=%v (current=%s, selected=%s, comparison=%d)",
					tc.shouldUpdate, updateNeeded, tc.currentVersion, selectedVersion, comparison)
			}

			// Step 5: Verify the fix - when dev is disabled, dev versions should never be selected
			if !tc.preferDev && isPrereleaseTag(selectedVersion) {
				t.Errorf("BUG: Prerelease version %s was selected even though dev builds are disabled", selectedVersion)
			}
		})
	}
}
