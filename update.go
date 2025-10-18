package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// -------------------- Self-update (no downgrades) --------------------

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func selfUpdate(root, exePath string) error {
	tag, assetURL, err := fetchLatestAsset(UPDATE_OWNER, UPDATE_REPO, UPDATE_ASSET)
	if err != nil || tag == "" || assetURL == "" {
		return err
	}

	localTag := normalizeTag(version)
	remoteTag := normalizeTag(tag)

	switch compareSemver(localTag, remoteTag) {
	case 0:
		logf("%s up to date (%s).", launcherShortName, version)
		return nil
	case 1:
		// local > remote → don't downgrade
		logf("Local launcher (%s) is newer than latest release (%s). Skipping update.", version, tag)
		return nil
	case -1:
		// remote > local → proceed
	}

	logf("New %s available: %s (current %s).", launcherShortName, tag, version)
	logf("%s", stepLine("Downloading update..."))

	tmpNew := exePath + ".new"
	if err := downloadTo(assetURL, tmpNew, 0755); err != nil {
		return err
	}

	logf("%s", successLine("Update downloaded successfully"))
	logf("%s", stepLine("The launcher will now restart to apply the update"))
	logf("Please wait while the launcher restarts with the new version...")
	logf("")
	logf("Restarting in 10 seconds...")

	// Give users time to read the message before restarting
	time.Sleep(10 * time.Second)

	// Use a pure Go approach to replace the executable and restart
	if err := replaceAndRestart(exePath, tmpNew); err != nil {
		return fmt.Errorf("failed to replace launcher: %w", err)
	}

	// This shouldn't be reached if replaceAndRestart succeeds
	os.Exit(0)
	return nil
}

// replaceAndRestart replaces the current executable with the new one and launches it
func replaceAndRestart(currentExe, newExe string) error {
	// Start the new launcher from the .new file, passing it the paths to clean up
	cmd := exec.Command(newExe, "--cleanup-after-update", currentExe, newExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the new process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new launcher: %w", err)
	}

	// Exit the current process
	os.Exit(0)
	return nil
}

// performUpdateCleanup handles the cleanup after an update
func performUpdateCleanup(oldExe, newExe string) {
	// Wait a moment for the old process to fully exit
	time.Sleep(2 * time.Second)

	// Try to replace the old executable with the new one
	if err := os.Rename(newExe, oldExe); err != nil {
		// If rename fails, try copy+delete approach
		if copyErr := copyFile(newExe, oldExe); copyErr == nil {
			os.Remove(newExe)
		}
	}

	// Use rundll32 to start the launcher (Windows-specific and more reliable)
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", oldExe).Start()

	// Exit this cleanup process
	os.Exit(0)
}

func fetchLatestAsset(owner, repo, wantName string) (tag, url string, err error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", getUserAgent("Updater"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", "", err
	}
	for _, a := range r.Assets {
		if a.Name == wantName {
			return r.TagName, a.BrowserDownloadURL, nil
		}
	}
	return r.TagName, "", errors.New("asset not found in latest release")
}

// --- Version helpers (semver-ish) ---

func normalizeTag(t string) string {
	// Strip leading "v" or "V" and whitespace
	t = strings.TrimSpace(t)
	if len(t) > 0 && (t[0] == 'v' || t[0] == 'V') {
		t = t[1:]
	}
	// Drop any pre-release/build suffix ("-rc.1", "+meta")
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

// returns -1 if a<b, 0 if equal, +1 if a>b
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