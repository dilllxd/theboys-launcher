package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func debugReleases() {
	owner := "dilllxd"
	repo := "theboyslauncher"

	fmt.Printf("Debugging releases for %s/%s\n\n", owner, repo)

	// Fetch releases page
	releasesURL := fmt.Sprintf("https://github.com/%s/%s/releases", owner, repo)
	fmt.Printf("Fetching: %s\n", releasesURL)

	resp, err := http.Get(releasesURL)
	if err != nil {
		fmt.Printf("Failed to fetch releases page: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("GitHub releases page returned status %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read releases page HTML: %v\n", err)
		return
	}
	html := string(body)

	// Find all release tags
	tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo))
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindAllStringSubmatch(html, -1)

	if len(tagMatches) == 0 {
		fmt.Printf("No release tags found!\n")
		return
	}

	// Extract all tags
	var allTags []string
	for _, match := range tagMatches {
		if len(match) >= 2 {
			allTags = append(allTags, match[1])
		}
	}

	fmt.Printf("Found %d release tags:\n", len(allTags))
	for i, tag := range allTags {
		fmt.Printf("%d. %s\n", i+1, tag)
	}

	// Test isPrereleaseTag function
	fmt.Printf("\nPrerelease analysis:\n")
	var stableTags []string
	var prereleaseTags []string

	for _, tag := range allTags {
		if debugIsPrereleaseTag(tag) {
			prereleaseTags = append(prereleaseTags, tag)
		} else {
			stableTags = append(stableTags, tag)
		}
	}

	fmt.Printf("Stable releases (%d):\n", len(stableTags))
	for _, tag := range stableTags {
		fmt.Printf("  - %s\n", tag)
	}

	fmt.Printf("\nPrerelease releases (%d):\n", len(prereleaseTags))
	for _, tag := range prereleaseTags {
		fmt.Printf("  - %s\n", tag)
	}

	if len(stableTags) == 0 {
		fmt.Printf("\n*** ERROR: No stable releases found! ***\n")
		fmt.Printf("This explains the error: 'no stable releases found for %s/%s'\n", owner, repo)
	} else {
		fmt.Printf("\nFirst stable release that would be selected: %s\n", stableTags[0])
	}
}

// debugIsPrereleaseTag checks if a version tag represents a prerelease/dev version
func debugIsPrereleaseTag(tag string) bool {
	tag = strings.ToLower(tag)
	prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}

	for _, indicator := range prereleaseIndicators {
		if strings.Contains(tag, indicator) {
			return true
		}
	}

	return false
}
