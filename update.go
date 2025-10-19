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

	"golang.org/x/sys/windows"
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

	tag, assetURL, err := fetchLatestAsset(UPDATE_OWNER, UPDATE_REPO, launcherExeName+getExecutableExtension())
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

// replaceAndRestart replaces the current executable with the new one and launches it
func replaceAndRestart(currentExe, newExe string) error {
	cmd := exec.Command(newExe, "--cleanup-after-update", "--cleanup-old-exe", currentExe, "--cleanup-new-exe", newExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

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
			// Hide console window on Windows
			fallbackCmd.SysProcAttr = &windows.SysProcAttr{
				HideWindow:    true,
				CreationFlags: windows.CREATE_NO_WINDOW,
			}
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
	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	if err := cmd.Start(); err != nil {
		logf("Failed to restart launcher: %v", err)
	}

	os.Exit(0)
}

func fetchLatestAsset(owner, repo, wantName string) (tag, url string, err error) {
	// Use GitHub's releases archive page to find the latest release without API
	// This approach avoids rate limiting entirely
	releasesURL := fmt.Sprintf("https://github.com/%s/%s/releases", owner, repo)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(releasesURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch releases page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub releases page returned status %d", resp.StatusCode)
	}

	// Read HTML content
	htmlBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read releases page HTML: %w", err)
	}
	html := string(htmlBody)

	// Extract the first (latest) release tag from the releases page
	// Look for release links in the format: /owner/repo/releases/tag/version
	tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"]+)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo))
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindStringSubmatch(html)

	if len(tagMatches) < 2 {
		return "", "", fmt.Errorf("could not find any release tags for %s/%s", owner, repo)
	}

	tag = tagMatches[1]
	fmt.Printf("Found latest tag: %s\n", tag)

	// Construct the direct download URL using GitHub's predictable URL pattern
	assetURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, tag, wantName)

	// Verify the asset exists by making a HEAD request
	headReq, err := http.NewRequest("HEAD", assetURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create HEAD request: %w", err)
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

func normalizeTag(t string) string {
	t = strings.TrimSpace(t)
	if len(t) > 0 && (t[0] == 'v' || t[0] == 'V') {
		t = t[1:]
	}
	if i := strings.IndexAny(t, "-+"); i >= 0 {
		t = t[:i]
	}
	return t
}

func parseSemverInts(t string) (major, minor, patch int) {
	parts := strings.Split(t, ".")
	get := func(i int) int {
		if i >= len(parts) || parts[i] == "" {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	return get(0), get(1), get(2)
}

func compareSemver(a, b string) int {
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
	return 0
}
