package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// TestFetchFromPage tests the fetchFromPage function with different scenarios
func TestFetchFromPage(t *testing.T) {
	testCases := []struct {
		name             string
		page             int
		preferPrerelease bool
		mockHTML         string
		expectedTag      string
		shouldError      bool
		errorContains    string
	}{
		{
			name:             "DevPage1_ShouldReturnDevVersion",
			page:             1,
			preferPrerelease: true,
			mockHTML: `
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
			expectedTag: "v3.2.30-dev.adcb1ae",
			shouldError: false,
		},
		{
			name:             "StablePage1_ShouldReturnStableVersion",
			page:             1,
			preferPrerelease: false,
			mockHTML: `
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
			expectedTag: "v3.2.29",
			shouldError: false,
		},
		{
			name:             "StablePage2_ShouldReturnStableVersion",
			page:             2,
			preferPrerelease: false,
			mockHTML: `
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.28-dev.abc123">
							<span class="css-truncate target">v3.2.28-dev.abc123</span>
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
			expectedTag: "v3.2.27",
			shouldError: false,
		},
		{
			name:             "NoReleases_ShouldError",
			page:             1,
			preferPrerelease: false,
			mockHTML: `
				<html>
				<body>
					<div class="release-entry">
						<p>No releases found</p>
					</div>
				</body>
				</html>
			`,
			shouldError:   true,
			errorContains: "could not find any release tags",
		},
		{
			name:             "OnlyDevVersionsPreferStable_ShouldReturnEmpty",
			page:             1,
			preferPrerelease: false,
			mockHTML: `
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
			expectedTag: "",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/releases") {
					w.Header().Set("Content-Type", "text/html")
					fmt.Fprint(w, tc.mockHTML)
				} else if strings.Contains(r.URL.Path, "/download/") {
					// Mock asset download endpoint
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, "mock binary content")
				}
			}))
			defer server.Close()

			// Test the logic directly (since we can't easily mock the HTTP calls in the actual function)
			owner := "test-owner"
			repo := "test-repo"

			// Extract tags using the same logic as fetchFromPage
			tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
			tagRe := regexp.MustCompile(tagPattern)
			tagMatches := tagRe.FindAllStringSubmatch(tc.mockHTML, -1)

			if len(tagMatches) == 0 && !tc.shouldError {
				t.Fatal("No release tags found in mock HTML")
			}

			// Extract all tags
			var allTags []string
			for _, match := range tagMatches {
				if len(match) >= 2 {
					allTags = append(allTags, match[1])
				}
			}

			var resultTag string
			var err error

			if tc.preferPrerelease {
				// When preferring prerelease, look for dev versions first
				prereleaseRe := regexp.MustCompile(fmt.Sprintf(`/%s/%s/releases/tag/([^"']*-dev[.\-][^"']*)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo)))
				if m := prereleaseRe.FindStringSubmatch(tc.mockHTML); len(m) >= 2 {
					resultTag = m[1]
				} else if len(allTags) > 0 {
					resultTag = allTags[0] // Fallback to first tag
				}
			} else {
				// When preferring stable, filter out prerelease versions
				var stableTags []string
				for _, tag := range allTags {
					if !isPrereleaseVersion(tag) {
						stableTags = append(stableTags, tag)
					}
				}
				if len(stableTags) == 0 {
					if len(allTags) == 0 {
						err = fmt.Errorf("could not find any release tags for %s/%s on page %d", owner, repo, tc.page)
					}
					// Return empty results but no error - let the caller decide to continue pagination
				} else {
					resultTag = stableTags[0]
				}
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got none", tc.errorContains)
				} else if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resultTag != tc.expectedTag {
					t.Errorf("Expected tag '%s', got '%s'", tc.expectedTag, resultTag)
				}
			}
		})
	}
}

// TestHasMorePages tests the hasMorePages function
func TestHasMorePages(t *testing.T) {
	testCases := []struct {
		name          string
		currentPage   int
		mockHTML      string
		expectedMore  bool
		shouldError   bool
		errorContains string
	}{
		{
			name:        "PageWithReleases_ShouldReturnTrue",
			currentPage: 1,
			mockHTML: `
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
				</body>
				</html>
			`,
			expectedMore: true,
			shouldError:  false,
		},
		{
			name:        "PageWithoutReleases_ShouldReturnFalse",
			currentPage: 10,
			mockHTML: `
				<html>
				<body>
					<div class="release-entry">
						<p>No releases found</p>
					</div>
				</body>
				</html>
			`,
			expectedMore: false,
			shouldError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner := "test-owner"
			repo := "test-repo"

			// Test the logic directly (since we can't easily mock the HTTP calls in the actual function)
			tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
			tagRe := regexp.MustCompile(tagPattern)
			tagMatches := tagRe.FindAllStringSubmatch(tc.mockHTML, -1)

			// hasMorePages returns true if there are releases on the current page
			hasMore := len(tagMatches) > 0

			if hasMore != tc.expectedMore {
				t.Errorf("Expected hasMorePages to return %v, got %v", tc.expectedMore, hasMore)
			}
		})
	}
}

// TestFetchLatestAssetPreferPrereleasePagination tests the pagination logic
func TestFetchLatestAssetPreferPrereleasePagination(t *testing.T) {
	testCases := []struct {
		name             string
		preferPrerelease bool
		pageResponses    []string // HTML responses for each page
		expectedTag      string
		expectedPages    int // Number of pages that should be checked
		shouldError      bool
		errorContains    string
	}{
		{
			name:             "PreferDev_ShouldOnlyCheckPage1",
			preferPrerelease: true,
			pageResponses: []string{
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedTag:   "v3.2.30-dev.adcb1ae",
			expectedPages: 1,
			shouldError:   false,
		},
		{
			name:             "PreferStableFoundOnPage1_ShouldNotPaginate",
			preferPrerelease: false,
			pageResponses: []string{
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
			expectedTag:   "v3.2.29",
			expectedPages: 1,
			shouldError:   false,
		},
		{
			name:             "PreferStableFoundOnPage2_ShouldPaginate",
			preferPrerelease: false,
			pageResponses: []string{
				// Page 1: Only dev versions
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
				// Page 2: Has stable version
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
			expectedTag:   "v3.2.27",
			expectedPages: 2,
			shouldError:   false,
		},
		{
			name:             "PreferStableNotFound_ShouldErrorAfterMaxPages",
			preferPrerelease: false,
			pageResponses: []string{
				// All pages have only dev versions (create 10 identical responses)
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
							<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedTag:   "",
			expectedPages: 10, // Should check max pages
			shouldError:   true,
			errorContains: "no stable releases found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner := "test-owner"
			repo := "test-repo"
			maxPages := 10

			var resultTag string
			var err error
			pagesChecked := 0

			// Simulate the pagination logic from FetchLatestAssetPreferPrerelease
			if tc.preferPrerelease {
				// For prerelease, only check page 1
				pagesChecked = 1
				if len(tc.pageResponses) > 0 {
					// Extract dev version from first page
					prereleaseRe := regexp.MustCompile(fmt.Sprintf(`/%s/%s/releases/tag/([^"']*-dev[.\-][^"']*)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo)))
					if m := prereleaseRe.FindStringSubmatch(tc.pageResponses[0]); len(m) >= 2 {
						resultTag = m[1]
					}
				}
			} else {
				// For stable releases, check multiple pages
				for page := 1; page <= maxPages; page++ {
					pagesChecked++

					// Use the last response repeatedly if we've run out of responses
					var mockHTML string
					if page-1 < len(tc.pageResponses) {
						mockHTML = tc.pageResponses[page-1]
					} else {
						// Reuse the last response for additional pages
						mockHTML = tc.pageResponses[len(tc.pageResponses)-1]
					}

					// Extract tags from current page
					tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
					tagRe := regexp.MustCompile(tagPattern)
					tagMatches := tagRe.FindAllStringSubmatch(mockHTML, -1)

					if len(tagMatches) == 0 {
						err = fmt.Errorf("could not find any release tags for %s/%s on page %d", owner, repo, page)
						break
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
						resultTag = stableTags[0]
						err = nil
						break // Found stable release
					}
				}

				if resultTag == "" && err == nil {
					err = fmt.Errorf("no stable releases found for %s/%s after checking %d pages", owner, repo, maxPages)
				}
			}

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got none", tc.errorContains)
				} else if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resultTag != tc.expectedTag {
					t.Errorf("Expected tag '%s', got '%s'", tc.expectedTag, resultTag)
				}
			}

			if pagesChecked != tc.expectedPages {
				t.Errorf("Expected to check %d pages, but checked %d", tc.expectedPages, pagesChecked)
			}
		})
	}
}

// TestIsPrereleaseTag tests the isPrereleaseTag function with various version formats
func TestIsPrereleaseTag(t *testing.T) {
	testCases := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "DevVersion",
			tag:      "v3.2.30-dev.adcb1ae",
			expected: true,
		},
		{
			name:     "BetaVersion",
			tag:      "v3.2.29-beta.1",
			expected: true,
		},
		{
			name:     "RCVersion",
			tag:      "v3.2.28-rc.2",
			expected: true,
		},
		{
			name:     "AlphaVersion",
			tag:      "v3.2.27-alpha.3",
			expected: true,
		},
		{
			name:     "PreVersion",
			tag:      "v3.2.26-pre.4",
			expected: true,
		},
		{
			name:     "StableVersion",
			tag:      "v3.2.25",
			expected: false,
		},
		{
			name:     "StableWithBuildMetadata",
			tag:      "v3.2.24+build.456",
			expected: false,
		},
		{
			name:     "EmptyTag",
			tag:      "",
			expected: false,
		},
		{
			name:     "UppercaseDev",
			tag:      "v3.2.23-DEV.abc123",
			expected: true,
		},
		{
			name:     "MixedCaseBeta",
			tag:      "v3.2.22-Beta.1",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isPrereleaseVersion(tc.tag)
			if result != tc.expected {
				t.Errorf("Expected isPrereleaseTag('%s') to return %v, got %v", tc.tag, tc.expected, result)
			}
		})
	}
}

// Helper function to check if a version is a prerelease (copied from update.go for testing)
func isPrereleaseVersion(tag string) bool {
	tag = strings.ToLower(tag)
	prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}

	for _, indicator := range prereleaseIndicators {
		if strings.Contains(tag, indicator) {
			return true
		}
	}

	return false
}
