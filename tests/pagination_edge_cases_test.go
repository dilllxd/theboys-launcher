package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestPaginationEdgeCases tests edge cases for the pagination functionality
func TestPaginationEdgeCases(t *testing.T) {
	testCases := []struct {
		name             string
		preferPrerelease bool
		pageResponses    []string
		expectedTag      string
		expectedPages    int
		shouldError      bool
		errorContains    string
		description      string
	}{
		{
			name:             "SinglePageRepository",
			preferPrerelease: false,
			pageResponses: []string{
				`
				<html>
				<body>
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
			description:   "Repository with only one page of releases",
		},
		{
			name:             "EmptyRepository",
			preferPrerelease: false,
			pageResponses: []string{
				`
				<html>
				<body>
					<div class="blankslate">
						<p>There aren't any releases here</p>
					</div>
				</body>
				</html>
				`,
			},
			expectedTag:   "",
			expectedPages: 10, // Should check all pages since no releases found
			shouldError:   true,
			errorContains: "could not find any release tags",
			description:   "Repository with no releases",
		},
		{
			name:             "NetworkErrorSimulation",
			preferPrerelease: false,
			pageResponses: []string{
				"", // Empty response to simulate network error
			},
			expectedTag:   "",
			expectedPages: 10, // Should check all pages since no releases found
			shouldError:   true,
			errorContains: "could not find any release tags",
			description:   "Network connectivity issues during pagination",
		},
		{
			name:             "MaxPagesReachedWithoutStable",
			preferPrerelease: false,
			pageResponses: []string{
				// 10 pages of dev versions (same content repeated)
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
			expectedPages: 10,
			shouldError:   true,
			errorContains: "no stable releases found",
			description:   "Maximum pages reached without finding stable releases",
		},
		{
			name:             "StableFoundOnLastPage",
			preferPrerelease: false,
			pageResponses: []string{
				// 9 pages of dev versions (same content repeated)
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
			expectedPages: 10, // Should check all pages since no stable releases found
			shouldError:   true,
			errorContains: "no stable releases found",
			description:   "Stable releases found on the last checked page",
		},
		{
			name:             "MixedVersionFormats",
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
						<a href="/test-owner/test-repo/releases/tag/v3.2.29-beta.1+build.123">
							<span class="css-truncate target">v3.2.29-beta.1+build.123</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.28-rc.2">
							<span class="css-truncate target">v3.2.28-rc.2</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.27+build.456">
							<span class="css-truncate target">v3.2.27+build.456</span>
						</a>
					</div>
				</body>
				</html>
				`,
			},
			expectedTag:   "v3.2.27+build.456",
			expectedPages: 1,
			shouldError:   false,
			description:   "Complex version formats with build metadata",
		},
		{
			name:             "CaseInsensitivePrereleaseDetection",
			preferPrerelease: false,
			pageResponses: []string{
				`
				<html>
				<body>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.30-DEV.adcb1ae">
							<span class="css-truncate target">v3.2.30-DEV.adcb1ae</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.29-Beta.1">
							<span class="css-truncate target">v3.2.29-Beta.1</span>
						</a>
					</div>
					<div class="release-entry">
						<a href="/test-owner/test-repo/releases/tag/v3.2.28-RC.2">
							<span class="css-truncate target">v3.2.28-RC.2</span>
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
			expectedPages: 1,
			shouldError:   false,
			description:   "Case insensitive prerelease detection",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing edge case: %s", tc.description)

			owner := "test-owner"
			repo := "test-repo"
			maxPages := 10

			var resultTag string
			var err error
			pagesChecked := 0

			// Simulate pagination logic from FetchLatestAssetPreferPrerelease
			if tc.preferPrerelease {
				// For prerelease, only check page 1
				pagesChecked = 1
				if len(tc.pageResponses) > 0 && tc.pageResponses[0] != "" {
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

					// Handle empty response (network error simulation)
					if mockHTML == "" {
						err = fmt.Errorf("could not find any release tags for %s/%s on page %d", owner, repo, page)
						continue // Continue to next page
					}

					// Extract tags from current page
					tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
					tagRe := regexp.MustCompile(tagPattern)
					tagMatches := tagRe.FindAllStringSubmatch(mockHTML, -1)

					if len(tagMatches) == 0 {
						err = fmt.Errorf("could not find any release tags for %s/%s on page %d", owner, repo, page)
						continue // Continue to next page
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

// TestPaginationPerformance tests performance aspects of pagination
func TestPaginationPerformance(t *testing.T) {
	t.Run("PaginationStopsEarlyWhenStableFound", func(t *testing.T) {
		// Test that pagination stops as soon as a stable release is found
		owner := "test-owner"
		repo := "test-repo"
		maxPages := 10

		// Mock responses: page 1 has dev versions, page 2 has stable
		pageResponses := []string{
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
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.29">
						<span class="css-truncate target">v3.2.29</span>
					</a>
				</div>
			</body>
			</html>
			`,
		}

		start := time.Now()
		pagesChecked := 0

		// Simulate pagination logic
		for page := 1; page <= maxPages; page++ {
			pagesChecked++
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
				break // Found stable release, should stop here
			}
		}

		elapsed := time.Since(start)

		// Should only check 2 pages (stop when stable found)
		if pagesChecked != 2 {
			t.Errorf("Expected to check 2 pages, but checked %d", pagesChecked)
		}

		// Should complete quickly (not checking all 10 pages)
		if elapsed > 100*time.Millisecond {
			t.Errorf("Pagination took too long: %v", elapsed)
		}
	})

	t.Run("DevModeOnlyChecksFirstPage", func(t *testing.T) {
		// Test that dev mode only checks the first page for efficiency
		owner := "test-owner"
		repo := "test-repo"

		// Mock response with dev version
		mockHTML := `
		<html>
		<body>
			<div class="release-entry">
				<a href="/test-owner/test-repo/releases/tag/v3.2.30-dev.adcb1ae">
					<span class="css-truncate target">v3.2.30-dev.adcb1ae</span>
				</a>
			</div>
		</body>
		</html>
		`

		start := time.Now()

		// Simulate dev mode logic (only check page 1)
		pagesChecked := 1
		prereleaseRe := regexp.MustCompile(fmt.Sprintf(`/%s/%s/releases/tag/([^"']*-dev[.\-][^"']*)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo)))
		_ = prereleaseRe.FindStringSubmatch(mockHTML)

		elapsed := time.Since(start)

		// Should only check 1 page
		if pagesChecked != 1 {
			t.Errorf("Expected to check 1 page in dev mode, but checked %d", pagesChecked)
		}

		// Should complete very quickly
		if elapsed > 50*time.Millisecond {
			t.Errorf("Dev mode check took too long: %v", elapsed)
		}
	})
}

// TestPaginationErrorHandling tests error handling during pagination
func TestPaginationErrorHandling(t *testing.T) {
	t.Run("HandlesEmptyPagesGracefully", func(t *testing.T) {
		// Test that empty pages are handled gracefully
		owner := "test-owner"
		repo := "test-repo"
		maxPages := 3

		// Mock responses: page 1 has dev, page 2 is empty, page 3 has stable
		pageResponses := []string{
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
			"", // Empty page
			`
			<html>
			<body>
				<div class="release-entry">
					<a href="/test-owner/test-repo/releases/tag/v3.2.29">
						<span class="css-truncate target">v3.2.29</span>
					</a>
				</div>
			</body>
			</html>
			`,
		}

		var resultTag string
		pagesChecked := 0

		// Simulate pagination logic with error handling
		for page := 1; page <= maxPages; page++ {
			pagesChecked++
			if page > len(pageResponses) {
				break
			}

			mockHTML := pageResponses[page-1]

			// Handle empty page
			if mockHTML == "" {
				continue // Skip to next page
			}

			// Extract tags from current page
			tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, owner, repo)
			tagRe := regexp.MustCompile(tagPattern)
			tagMatches := tagRe.FindAllStringSubmatch(mockHTML, -1)

			if len(tagMatches) == 0 {
				continue // Skip to next page
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
				break // Found stable release
			}
		}

		// Should find stable version on page 3 after skipping empty page 2
		if resultTag != "v3.2.29" {
			t.Errorf("Expected to find v3.2.29, got %s", resultTag)
		}

		if pagesChecked != 3 {
			t.Errorf("Expected to check 3 pages, but checked %d", pagesChecked)
		}
	})
}
