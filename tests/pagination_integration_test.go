package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestPaginationIntegration tests the complete flow from dev to stable switching
func TestPaginationIntegration(t *testing.T) {
	testCases := []struct {
		name            string
		initialDevMode  bool
		finalDevMode    bool
		pageResponses   []string
		expectedInitial string
		expectedFinal   string
		description     string
	}{
		{
			name:           "DevToStableSwitch",
			initialDevMode: true,
			finalDevMode:   false,
			pageResponses: []string{
				// Page 1: Latest dev version
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.29-dev.abc123">
							<span class="css-truncate target">v3.2.29-dev.abc123</span>
						</a>
					</div>
				</body>
				</html>
				`,
				// Page 2: Stable version
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.28-dev.def456">
							<span class="css-truncate target">v3.2.28-dev.def456</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.27">
							<span class="css-truncate target">v3.2.27</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedInitial: "v3.2.30-dev.adcb1ae",
			expectedFinal:   "v3.2.27",
			description:     "Switch from dev mode to stable mode",
		},
		{
			name:           "StableToDevSwitch",
			initialDevMode: false,
			finalDevMode:   true,
			pageResponses: []string{
				// Page 1: Mixed versions
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.29">
							<span class="css-truncate target">v3.2.29</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedInitial: "v3.2.29",
			expectedFinal:   "v3.2.30-dev.adcb1ae",
			description:     "Switch from stable mode to dev mode",
		},
		{
			name:           "StableToStableNoChange",
			initialDevMode: false,
			finalDevMode:   false,
			pageResponses: []string{
				// Page 1: Stable version available
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.29">
							<span class="css-truncate target">v3.2.29</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedInitial: "v3.2.29",
			expectedFinal:   "v3.2.29",
			description:     "Stay in stable mode",
		},
		{
			name:           "DevToDevNoChange",
			initialDevMode: true,
			finalDevMode:   true,
			pageResponses: []string{
				// Page 1: Dev version available
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.29">
							<span class="css-truncate target">v3.2.29</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedInitial: "v3.2.30-dev.adcb1ae",
			expectedFinal:   "v3.2.30-dev.adcb1ae",
			description:     "Stay in dev mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing integration scenario: %s", tc.description)

			owner := "test-owner"
			repo := "test-repo"
			wantName := "test-binary"

			// Test initial mode
			initialTag := simulateFetchLatestAsset(owner, repo, wantName, tc.initialDevMode, tc.pageResponses)
			if initialTag != tc.expectedInitial {
				t.Errorf("Expected initial tag '%s', got '%s'", tc.expectedInitial, initialTag)
			}

			// Test final mode
			finalTag := simulateFetchLatestAsset(owner, repo, wantName, tc.finalDevMode, tc.pageResponses)
			if finalTag != tc.expectedFinal {
				t.Errorf("Expected final tag '%s', got '%s'", tc.expectedFinal, finalTag)
			}

			// Verify the switch worked correctly
			if tc.initialDevMode != tc.finalDevMode {
				if initialTag == finalTag {
					t.Errorf("Expected different versions when switching modes, but got same version: %s", initialTag)
				}
			}
		})
	}
}

// TestPaginationWithSettings tests integration with settings persistence
func TestPaginationWithSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-pagination-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("SettingsPersistenceWithPagination", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "settings.json")

		// Test 1: Start with dev mode enabled
		initialSettings := map[string]interface{}{
			"memoryMB":         4096,
			"autoRam":          true,
			"devBuildsEnabled": true,
		}

		data, err := json.MarshalIndent(initialSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal initial settings: %v", err)
		}

		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write initial settings: %v", err)
		}

		// Read settings back
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings: %v", err)
		}

		var loadedSettings map[string]interface{}
		err = json.Unmarshal(content, &loadedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}

		// Verify dev mode is enabled
		if devBuildsEnabled, ok := loadedSettings["devBuildsEnabled"].(bool); !ok || !devBuildsEnabled {
			t.Error("Expected devBuildsEnabled to be true in initial settings")
		}

		// Test 2: Switch to stable mode
		loadedSettings["devBuildsEnabled"] = false

		data, err = json.MarshalIndent(loadedSettings, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal updated settings: %v", err)
		}

		err = os.WriteFile(settingsPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write updated settings: %v", err)
		}

		// Read settings back again
		content, err = os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read updated settings: %v", err)
		}

		var reloadedSettings map[string]interface{}
		err = json.Unmarshal(content, &reloadedSettings)
		if err != nil {
			t.Fatalf("Failed to unmarshal updated settings: %v", err)
		}

		// Verify dev mode is now disabled
		if devBuildsEnabled, ok := reloadedSettings["devBuildsEnabled"].(bool); !ok || devBuildsEnabled {
			t.Error("Expected devBuildsEnabled to be false in updated settings")
		}

		// Test 3: Simulate version selection based on settings
		pageResponses := []string{
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
						<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
					</a>
				</div>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.29">
						<span class="css-truncate target">v3.2.29</span>
					</a>
				</div>
			</body>
			</html>
			`,
		}

		// With dev mode disabled, should select stable version
		stableTag := simulateFetchLatestAsset("test-owner", "test-repo", "test-binary", false, pageResponses)
		if stableTag != "v3.2.29" {
			t.Errorf("Expected stable version v3.2.29 when dev mode is disabled, got %s", stableTag)
		}

		// With dev mode enabled, should select dev version
		devTag := simulateFetchLatestAsset("test-owner", "test-repo", "test-binary", true, pageResponses)
		if devTag != "v3.2.30-dev.adcb1ae" {
			t.Errorf("Expected dev version v3.2.30-dev.adcb1ae when dev mode is enabled, got %s", devTag)
		}
	})
}

// TestPaginationWithRealWorldScenarios tests real-world scenarios
func TestPaginationWithRealWorldScenarios(t *testing.T) {
	t.Run("ManyDevReleasesFewStable", func(t *testing.T) {
		// Simulate a repository with many dev releases and few stable releases
		pageResponses := make([]string, 5)

		// Pages 1-4: Only dev releases
		for i := 0; i < 4; i++ {
			pageResponses[i] = fmt.Sprintf(`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.%d-dev.abc123">
						<span class="css-truncate target">v3.2.%d-dev.abc123</span>
					</a>
				</div>
			</body>
			</html>
			`, 30-i, 30-i)
		}

		// Page 5: Finally a stable release
		pageResponses[4] = `
		<html>
		<body>
			<div class="release-entry">
				<a href="/test-owner/test-repo/releases/tag/v3.2.26-dev.def456">
					<span class="css-truncate target">v3.2.26-dev.def456</span>
				</a>
			</div>
			<div class="release-entry">
				<a href="/test-owner/test-repo/releases/tag/v3.2.25">
					<span class="css-truncate target">v3.2.25</span>
				</a>
			</div>
		</body>
		</html>
		`

		// Should find stable release on page 5
		stableTag := simulateFetchLatestAsset("test-owner", "test-repo", "test-binary", false, pageResponses)
		if stableTag != "v3.2.25" {
			t.Errorf("Expected to find stable version v3.2.25 on page 5, got %s", stableTag)
		}

		// Dev mode should still only check page 1
		devTag := simulateFetchLatestAsset("test-owner", "test-repo", "test-binary", true, pageResponses)
		if devTag != "v3.2.30-dev.abc123" {
			t.Errorf("Expected dev version v3.2.30-dev.abc123 from page 1, got %s", devTag)
		}
	})

	t.Run("StableReleasesOnMultiplePages", func(t *testing.T) {
		// Simulate stable releases spread across multiple pages
		pageResponses := []string{
			// Page 1: Only dev releases
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
						<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
					</a>
				</div>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.29-dev.abc123">
						<span class="css-truncate target">v3.2.29-dev.abc123</span>
					</a>
				</div>
			</body>
			</html>
			`,
			// Page 2: Mixed, but stable is not first
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.28-dev.def456">
						<span class="css-truncate target">v3.2.28-dev.def456</span>
					</a>
				</div>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.27-beta.1">
						<span class="css-truncate target">v3.2.27-beta.1</span>
					</a>
				</div>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.26">
						<span class="css-truncate target">v3.2.26</span>
					</a>
				</div>
			</body>
			</html>
			`,
			// Page 3: Stable release is first
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.25">
						<span class="css-truncate target">v3.2.25</span>
					</a>
				</div>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.24-dev.ghi789">
						<span class="css-truncate target">v3.2.24-dev.ghi789</span>
					</a>
				</div>
			</body>
			</html>
			`,
		}

		// Should find the first stable release (v3.2.26 on page 2)
		stableTag := simulateFetchLatestAsset("test-owner", "test-repo", "test-binary", false, pageResponses)
		if stableTag != "v3.2.26" {
			t.Errorf("Expected to find first stable version v3.2.26 on page 2, got %s", stableTag)
		}
	})
}

// Helper function to simulate FetchLatestAssetPreferPrerelease logic
func simulateFetchLatestAsset(owner, repo, wantName string, preferPrerelease bool, pageResponses []string) string {
	maxPages := 10

	if preferPrerelease {
		// For prerelease, only check page 1
		if len(pageResponses) > 0 {
			mockHTML := pageResponses[0]
			prereleaseRe := regexp.MustCompile(fmt.Sprintf(`/%s/%s/releases/tag/([^"']*-dev[.\-][^"']*)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo)))
			if m := prereleaseRe.FindStringSubmatch(mockHTML); len(m) >= 2 {
				return m[1]
			}
		}
	} else {
		// For stable releases, check multiple pages
		for page := 1; page <= maxPages; page++ {
			if page > len(pageResponses) {
				break
			}

			mockHTML := pageResponses[page-1]

			// Extract tags from current page
			tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
			tagRe := regexp.MustCompile(tagPattern)
			tagMatches := tagRe.FindAllStringSubmatch(mockHTML, -1)

			if len(tagMatches) == 0 {
				continue
			}

			// Extract all tags
			var allTags []string
			for _, match := range tagMatches {
				if len(match) >= 2 {
					allTags = append(allTags, match[1])
				}
			}

			// Filter for stable versions
			var stableTags []string
			for _, tag := range allTags {
				if !isPrereleaseVersion(tag) {
					stableTags = append(stableTags, tag)
				}
			}

			if len(stableTags) > 0 {
				return stableTags[0] // Return first stable tag found
			}
		}
	}

	return ""
}
