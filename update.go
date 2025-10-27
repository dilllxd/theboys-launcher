package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// -------------------- Self-update (no downgrades) --------------------

func selfUpdate(root, exePath string, report func(string)) error {
	_ = root

	notify := func(msg string) {
		if report != nil {
			report(msg)
		}
	}

	notify("Checking for launcher updates...")

	// Prefer prerelease/dev builds if the user has enabled them
	preferDev := settings.DevBuildsEnabled
	tag, assetURL, err := FetchLatestAssetPreferPrerelease(UPDATE_OWNER, UPDATE_REPO, LauncherAssetName, preferDev)
	if err != nil || tag == "" || assetURL == "" {
		if err == nil {
			err = errors.New("update metadata missing")
		}
		notify(fmt.Sprintf("Update check failed: %v", err))
		return err
	}

	localTag := normalizeTag(version)
	remoteTag := normalizeTag(tag)

	switch compareSemver(localTag, remoteTag) {
	case 0:
		msg := fmt.Sprintf("%s is up to date (%s)", launcherShortName, version)
		notify(msg)
		return nil
	case 1:
		msg := fmt.Sprintf("Local launcher (%s) is newer than latest release (%s). Skipping update.", version, tag)
		notify(msg)
		return nil
	case -1:
		// remote > local -> proceed with update
	}

	logf("New %s available: %s (current %s).", launcherShortName, tag, version)
	notify(fmt.Sprintf("Downloading update %s...", tag))
	logf("%s", stepLine("Downloading update..."))

	tmpNew := exePath + ".new"
	if err := downloadTo(assetURL, tmpNew, 0755); err != nil {
		notify(fmt.Sprintf("Update download failed: %v", err))
		return err
	}

	// Remove quarantine attribute on macOS (no-op on Windows)
	if err := removeQuarantineAttribute(tmpNew); err != nil {
		notify(fmt.Sprintf("Warning: Failed to remove quarantine attribute: %v", err))
		// Don't fail the update, just warn the user
	}

	notify("Update downloaded successfully")
	logf("%s", successLine("Update downloaded successfully"))
	notify("Preparing to restart with the new version...")
	logf("%s", stepLine("The launcher will now restart to apply the update"))
	logf("Please wait while the launcher restarts with the new version...")
	logf("")
	logf("Restarting in 10 seconds...")

	time.Sleep(10 * time.Second)
	notify("Restarting to apply update...")

	if err := replaceAndRestart(exePath, tmpNew); err != nil {
		notify(fmt.Sprintf("Failed to restart launcher: %v", err))
		return fmt.Errorf("failed to replace launcher: %w", err)
	}

	os.Exit(0)
	return nil
}

// forceUpdate forces an update to the latest version regardless of current version
func forceUpdate(root, exePath string, preferDev bool, report func(string)) error {
	_ = root

	notify := func(msg string) {
		if report != nil {
			report(msg)
		}
	}

	notify("Checking for latest launcher version...")

	// Fetch the latest asset based on preference (dev or stable)
	tag, assetURL, err := FetchLatestAssetPreferPrerelease(UPDATE_OWNER, UPDATE_REPO, LauncherAssetName, preferDev)
	if err != nil || tag == "" || assetURL == "" {
		if err == nil {
			err = errors.New("update metadata missing")
		}
		notify(fmt.Sprintf("Failed to fetch latest version: %v", err))
		return err
	}

	channel := "stable"
	if preferDev {
		channel = "dev"
	}

	logf("Force updating to latest %s version: %s", channel, tag)
	notify(fmt.Sprintf("Downloading %s version %s...", channel, tag))
	logf("%s", stepLine("Downloading update..."))

	tmpNew := exePath + ".new"
	if err := downloadTo(assetURL, tmpNew, 0755); err != nil {
		notify(fmt.Sprintf("Update download failed: %v", err))
		return err
	}

	// Remove quarantine attribute on macOS (no-op on Windows)
	if err := removeQuarantineAttribute(tmpNew); err != nil {
		notify(fmt.Sprintf("Warning: Failed to remove quarantine attribute: %v", err))
		// Don't fail the update, just warn the user
	}

	notify("Update downloaded successfully")
	logf("%s", successLine("Update downloaded successfully"))
	notify("Preparing to restart with the new version...")
	logf("%s", stepLine("The launcher will now restart to apply the update"))
	logf("Please wait while the launcher restarts with the new version...")
	logf("")
	logf("Restarting in 10 seconds...")

	time.Sleep(10 * time.Second)
	notify("Restarting to apply update...")

	if err := replaceAndRestart(exePath, tmpNew); err != nil {
		notify(fmt.Sprintf("Failed to restart launcher: %v", err))
		return fmt.Errorf("failed to replace launcher: %w", err)
	}

	os.Exit(0)
	return nil
}

// FetchLatestAssetPreferPrerelease fetches the latest asset URL for the desired binary.
// If preferPrerelease is true it will attempt to find a prerelease tag (containing "dev") first,
// otherwise it falls back to the latest normal release.
func FetchLatestAssetPreferPrerelease(owner, repo, wantName string, preferPrerelease bool) (tag, url string, err error) {
	const maxPages = 10 // Limit pagination to avoid infinite loops

	// If preferPrerelease, we only need to check the first page since dev builds are recent
	if preferPrerelease {
		return fetchFromPage(owner, repo, wantName, 1, preferPrerelease)
	}

	// For stable releases, we need to check multiple pages since stable releases might be on older pages
	for page := 1; page <= maxPages; page++ {
		logf("Checking page %d for stable releases...", page)
		tag, url, err := fetchFromPage(owner, repo, wantName, page, preferPrerelease)
		if err != nil {
			// If we get an error that indicates no more releases, stop pagination
			if strings.Contains(err.Error(), "could not find any release tags") {
				logf("No more releases found on page %d, stopping pagination", page)
				break
			}
			// For other errors, continue to next page
			logf("Error checking page %d: %v", page, err)
			continue
		}

		// If we found a stable release, return it
		if tag != "" && url != "" {
			logf("Found stable release %s on page %d", tag, page)
			return tag, url, nil
		}

		// Check if there are more pages by looking for pagination indicators
		hasMore, err := hasMorePages(owner, repo, page)
		if err != nil {
			logf("Error checking for more pages: %v", err)
			break
		}
		if !hasMore {
			logf("No more pages available, stopping pagination at page %d", page)
			break
		}
	}

	return "", "", fmt.Errorf("no stable releases found for %s/%s after checking %d pages", owner, repo, maxPages)
}

// fetchFromPage fetches releases from a specific page and returns the appropriate tag/URL
func fetchFromPage(owner, repo, wantName string, page int, preferPrerelease bool) (tag, url string, err error) {
	var releasesURL string
	if page == 1 {
		releasesURL = fmt.Sprintf("https://github.com/%s/%s/releases", owner, repo)
	} else {
		releasesURL = fmt.Sprintf("https://github.com/%s/%s/releases?page=%d", owner, repo, page)
	}

	resp, err := http.Get(releasesURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch releases page %d: %w", page, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub releases page %d returned status %d", page, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read releases page %d HTML: %w", page, err)
	}
	html := string(body)

	// If preferPrerelease, try to locate a tag that contains 'dev' (our autobump uses dev.<sha>)
	if preferPrerelease {
		// Match tags like v1.2.3-dev.<sha> (look for '-dev.' to reduce false positives)
		prereleaseRe := regexp.MustCompile(fmt.Sprintf(`/%s/%s/releases/tag/([^"']*-dev[.\-][^"']*)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo)))
		if m := prereleaseRe.FindStringSubmatch(html); len(m) >= 2 {
			tag = m[1]
			assetURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, tag, wantName)
			// verify
			headReq, _ := http.NewRequest("HEAD", assetURL, nil)
			headReq.Header.Set("User-Agent", getUserAgent("General"))
			headResp, err := http.DefaultClient.Do(headReq)
			if err == nil && headResp != nil {
				headResp.Body.Close()
				if headResp.StatusCode == 200 {
					return tag, assetURL, nil
				}
			}
		}
		// fall through to normal latest release
	}

	// Fallback: find release tags on the releases page
	tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo))
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindAllStringSubmatch(html, -1)

	if len(tagMatches) == 0 {
		return "", "", fmt.Errorf("could not find any release tags for %s/%s on page %d", owner, repo, page)
	}

	// Extract all tags
	var allTags []string
	for _, match := range tagMatches {
		if len(match) >= 2 {
			allTags = append(allTags, match[1])
		}
	}

	// If preferPrerelease is false, filter out dev/prerelease versions
	if !preferPrerelease {
		var stableTags []string
		for _, tag := range allTags {
			if !isPrereleaseTag(tag) {
				stableTags = append(stableTags, tag)
			}
		}
		if len(stableTags) == 0 {
			// Return empty results but no error - let the caller decide to continue pagination
			return "", "", nil
		}
		tag = stableTags[0] // Return the first stable tag found
	} else {
		// When preferring prerelease, return the first tag found
		tag = allTags[0]
	}

	assetURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, tag, wantName)

	// Verify the asset exists by making a HEAD request
	headReq, err := http.NewRequest("HEAD", assetURL, nil)
	if err != nil {
		return tag, "", fmt.Errorf("failed to create HEAD request: %w", err)
	}
	headReq.Header.Set("User-Agent", getUserAgent("General"))

	headResp, err := http.DefaultClient.Do(headReq)
	if err != nil {
		return tag, "", fmt.Errorf("failed to verify asset exists: %w", err)
	}
	defer headResp.Body.Close()

	if headResp.StatusCode != 200 {
		return tag, "", fmt.Errorf("asset %s not found for release %s (HTTP %d)", wantName, tag, headResp.StatusCode)
	}

	return tag, assetURL, nil
}

// hasMorePages checks if there are more pages of releases by looking for pagination indicators
func hasMorePages(owner, repo string, currentPage int) (bool, error) {
	// For this implementation, we'll check if the current page has any releases
	// If it has releases, we assume there might be more pages
	// This is a simple approach that works well for our use case

	var releasesURL string
	if currentPage == 1 {
		releasesURL = fmt.Sprintf("https://github.com/%s/%s/releases", owner, repo)
	} else {
		releasesURL = fmt.Sprintf("https://github.com/%s/%s/releases?page=%d", owner, repo, currentPage)
	}

	resp, err := http.Get(releasesURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("GitHub releases page %d returned status %d", currentPage, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	html := string(body)

	// Check if there are any release tags on this page
	tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo))
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindAllStringSubmatch(html, -1)

	// If we found releases on this page, there might be more pages
	// This is a conservative approach that ensures we don't miss stable releases
	return len(tagMatches) > 0, nil
}

// isPrereleaseTag checks if a version tag represents a prerelease/dev version
func isPrereleaseTag(tag string) bool {
	tag = strings.ToLower(tag)
	prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}

	for _, indicator := range prereleaseIndicators {
		if strings.Contains(tag, indicator) {
			return true
		}
	}

	return false
}

// replaceAndRestart replaces the current executable with the new one and launches it
func replaceAndRestart(currentExe, newExe string) error {
	cmd := exec.Command(newExe, "--cleanup-after-update", "--cleanup-old-exe", currentExe, "--cleanup-new-exe", newExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set platform-specific process attributes
	setUpdateProcessAttributes(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new launcher: %w", err)
	}

	os.Exit(0)
	return nil
}

// performUpdateCleanup handles the cleanup after an update
func performUpdateCleanup(oldExe, newExe string) {
	time.Sleep(2 * time.Second)

	// Try to rename the new executable to replace the old one
	if err := os.Rename(newExe, oldExe); err != nil {
		// If rename fails (common on Windows with running executables), try copy+remove
		logf("Rename failed, attempting copy operation: %v", err)
		if copyErr := copyFile(newExe, oldExe); copyErr != nil {
			logf("Failed to copy new executable: %v", copyErr)
			// Try to remove the new exe and restart with old version
			os.Remove(newExe)
			fallbackCmd := exec.Command(oldExe)
			// Set platform-specific process attributes for fallback
			setFallbackUpdateProcessAttributes(fallbackCmd)
			_ = fallbackCmd.Start()
		} else {
			logf("Successfully copied new executable")
			os.Remove(newExe)
		}
	} else {
		logf("Successfully renamed new executable")
	}

	// Start the (now updated) launcher
	cmd := exec.Command(oldExe)
	// Set platform-specific process attributes for restart
	setRestartUpdateProcessAttributes(cmd)
	if err := cmd.Start(); err != nil {
		logf("Failed to restart launcher: %v", err)
	}

	os.Exit(0)
}

func fetchLatestAsset(owner, repo, wantName string) (tag, url string, err error) {
	// Delegate to the prefer-prerelease fetcher so callers automatically respect the
	// global DevBuildsEnabled setting when present.
	return FetchLatestAssetPreferPrerelease(owner, repo, wantName, settings.DevBuildsEnabled)
}

func normalizeTag(t string) string {
	t = strings.TrimSpace(t)
	if len(t) > 0 && (t[0] == 'v' || t[0] == 'V') {
		t = t[1:]
	}
	// Don't strip prerelease information - keep it for proper comparison
	return t
}

func parseSemverInts(t string) (major, minor, patch int) {
	// Split version to separate core version from prerelease
	coreVersion := t
	if i := strings.Index(t, "-"); i >= 0 {
		coreVersion = t[:i]
	}

	parts := strings.Split(coreVersion, ".")
	get := func(i int) int {
		if i >= len(parts) || parts[i] == "" {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	return get(0), get(1), get(2)
}

func getPrerelease(t string) string {
	if i := strings.Index(t, "-"); i >= 0 {
		return t[i+1:]
	}
	return ""
}

func comparePrerelease(a, b string) int {
	// According to semver:
	// 1. A version without prerelease is considered higher than one with prerelease
	// 2. Prerelease is compared dot-separated identifiers
	// 3. Numeric identifiers are compared numerically
	// 4. Alphanumeric identifiers are compared lexically
	// 5. Numeric identifiers are lower than alphanumeric identifiers

	if a == "" && b != "" {
		return 1 // a is stable, b is prerelease -> a is newer
	}
	if a != "" && b == "" {
		return -1 // a is prerelease, b is stable -> b is newer
	}
	if a == "" && b == "" {
		return 0 // both are stable
	}

	// Both have prerelease, compare them
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aPart, bPart string
		if i < len(aParts) {
			aPart = aParts[i]
		}
		if i < len(bParts) {
			bPart = bParts[i]
		}

		// If one has fewer parts, the one with more parts is considered newer
		if aPart == "" && bPart != "" {
			return -1
		}
		if aPart != "" && bPart == "" {
			return 1
		}

		// Try to compare as numbers
		aNum, aErr := strconv.Atoi(aPart)
		bNum, bErr := strconv.Atoi(bPart)

		if aErr == nil && bErr == nil {
			// Both numeric, compare numerically
			if aNum < bNum {
				return -1
			}
			if aNum > bNum {
				return 1
			}
		} else if aErr == nil {
			// a is numeric, b is alphanumeric -> a is lower
			return -1
		} else if bErr == nil {
			// a is alphanumeric, b is numeric -> a is higher
			return 1
		} else {
			// Both alphanumeric, compare lexically
			if aPart < bPart {
				return -1
			}
			if aPart > bPart {
				return 1
			}
		}
	}

	return 0 // prereleases are equal
}

func compareSemver(a, b string) int {
	// Compare core version (major.minor.patch)
	amaj, amin, apat := parseSemverInts(a)
	bmaj, bmin, bpat := parseSemverInts(b)

	if amaj != bmaj {
		if amaj < bmaj {
			return -1
		}
		return 1
	}
	if amin != bmin {
		if amin < bmin {
			return -1
		}
		return 1
	}
	if apat != bpat {
		if apat < bpat {
			return -1
		}
		return 1
	}

	// Core versions are equal, compare prerelease
	aPre := getPrerelease(a)
	bPre := getPrerelease(b)
	return comparePrerelease(aPre, bPre)
}
