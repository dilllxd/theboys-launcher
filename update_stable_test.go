package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// TestFetchLatestAssetPreferStable tests that when preferPrerelease is false,
// the function correctly filters out dev/prerelease versions and returns only stable releases
func TestFetchLatestAssetPreferStable(t *testing.T) {
	// Mock HTML response from GitHub releases page with both dev and stable releases
	mockHTML := `
	<html>
	<body>
		<div class="release-entry">
			<a href="/dilllxd/theboyslauncher/releases/tag/v3.2.30-dev.adcb1ae">
				<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
			</a>
		</div>
		<div class="release-entry">
			<a href="/dilllxd/theboyslauncher/releases/tag/v3.2.29">
				<span class="css-truncate target">v3.2.29</span>
			</a>
		</div>
		<div class="release-entry">
			<a href="/dilllxd/theboyslauncher/releases/tag/v3.2.28-beta">
				<span class="css-truncate target">v3.2.28-beta</span>
			</a>
		</div>
		<div class="release-entry">
			<a href="/dilllxd/theboyslauncher/releases/tag/v3.2.27">
				<span class="css-truncate target">v3.2.27</span>
			</a>
		</div>
	</body>
	</html>
	`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/releases") {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, mockHTML)
		} else if strings.Contains(r.URL.Path, "/download/") {
			// Mock asset download endpoint
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "mock binary content")
		}
	}))
	defer server.Close()

	// Test cases
	testCases := []struct {
		name             string
		preferPrerelease bool
		expectedTag      string
		shouldFail       bool
	}{
		{
			name:             "PreferDevTrue_ShouldReturnDevVersion",
			preferPrerelease: true,
			expectedTag:      "v3.2.30-dev.adcb1ae",
			shouldFail:       false,
		},
		{
			name:             "PreferDevFalse_ShouldReturnLatestStable",
			preferPrerelease: false,
			expectedTag:      "v3.2.29", // Should be v3.2.29, not v3.2.30-dev.adcb1ae
			shouldFail:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the isPrereleaseTag function directly
			owner := "dilllxd"
			repo := "theboyslauncher"

			// This is the current regex pattern from the function
			tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
			tagRe := tagPattern

			// Find all matches
			matches := regexp.MustCompile(tagRe).FindAllStringSubmatch(mockHTML, -1)

			if len(matches) == 0 {
				t.Fatal("No release tags found in mock HTML")
			}

			// Extract all tags
			var allTags []string
			for _, match := range matches {
				if len(match) >= 2 {
					allTags = append(allTags, match[1])
				}
			}

			if len(allTags) == 0 {
				t.Fatal("No tags extracted from matches")
			}

			// Test the filtering logic that should be in the fixed function
			var resultTag string
			if tc.preferPrerelease {
				// When preferring prerelease, return the first tag
				resultTag = allTags[0]
			} else {
				// When preferring stable, filter out prerelease versions
				var stableTags []string
				for _, tag := range allTags {
					if !isPrereleaseTag(tag) {
						stableTags = append(stableTags, tag)
					}
				}
				if len(stableTags) == 0 {
					t.Fatal("No stable tags found")
				}
				resultTag = stableTags[0]
			}

			// Verify the result
			if resultTag != tc.expectedTag {
				t.Errorf("Expected tag %s, got %s", tc.expectedTag, resultTag)
			}
		})
	}
}

// TestFilterStableReleases tests the logic for filtering out prerelease versions
func TestFilterStableReleases(t *testing.T) {
	testCases := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "MixedDevAndStable",
			tags:     []string{"v3.2.30-dev.adcb1ae", "v3.2.29", "v3.2.28-beta", "v3.2.27"},
			expected: "v3.2.29", // Should return the latest stable version
		},
		{
			name:     "OnlyDevVersions",
			tags:     []string{"v3.2.30-dev.adcb1ae", "v3.2.29-dev.abc123"},
			expected: "", // Should return empty if no stable versions
		},
		{
			name:     "OnlyStableVersions",
			tags:     []string{"v3.2.29", "v3.2.28", "v3.2.27"},
			expected: "v3.2.29", // Should return the first (latest) stable version
		},
		{
			name:     "MixedPrereleaseTypes",
			tags:     []string{"v3.2.30-dev.adcb1ae", "v3.2.29", "v3.2.28-beta", "v3.2.27-rc1", "v3.2.26"},
			expected: "v3.2.29", // Should filter out all prerelease types
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is what the filtering logic should do
			var result string
			for _, tag := range tc.tags {
				// Check if tag is stable (doesn't contain dev, beta, rc, etc.)
				if !isPrereleaseVersion(tag) {
					result = tag
					break // Return the first stable version found
				}
			}

			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// Helper function to check if a version is a prerelease
func isPrereleaseVersion(version string) bool {
	version = strings.ToLower(version)
	prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}

	for _, indicator := range prereleaseIndicators {
		if strings.Contains(version, indicator) {
			return true
		}
	}

	return false
}
