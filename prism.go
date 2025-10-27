package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// -------------------- Prism + Instance --------------------

func ensurePrism(dir string) (bool, error) {
	if exists(GetPrismExecutablePath(dir)) {
		return false, nil
	}

	var url string
	var err error

	// Handle different platforms - macOS doesn't have portable builds
	if runtime.GOOS == "darwin" {
		// macOS: download universal ZIP to Applications folder
		applicationsDir := "/Applications"
		prismAppPath := filepath.Join(applicationsDir, "PrismLauncher.app")

		// Check if Prism is already installed in Applications
		if exists(prismAppPath) {
			logf("%s", successLine("Prism Launcher found in Applications folder"))
			return false, nil
		}

		// Download to temp directory first
		tempDir := filepath.Join(os.TempDir(), "prism-download")
		os.MkdirAll(tempDir, 0755)
		defer os.RemoveAll(tempDir)

		url, err = fetchLatestPrismPortableURL()
		if err != nil {
			return false, err
		}

		logf("%s", stepLine(fmt.Sprintf("Downloading Prism universal build: %s", url)))
		if err := downloadAndUnzipTo(url, tempDir); err != nil {
			return false, err
		}

		// Debug: Show what was actually extracted
		logf("DEBUG: Contents of extracted archive:")
		filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, _ := filepath.Rel(tempDir, path)
				logf("  %s", relPath)
			}
			return nil
		})

		// Look for Prism Launcher.app in various locations
		var tempAppPath string
		possiblePaths := []string{
			filepath.Join(tempDir, "Prism Launcher.app"), // Correct name with space
			filepath.Join(tempDir, "PrismLauncher.app"),  // Fallback for no space
			filepath.Join(tempDir, "PrismLauncher"),      // Maybe it's just the app contents
		}

		for _, path := range possiblePaths {
			if exists(path) {
				tempAppPath = path
				logf("DEBUG: Found Prism at: %s", path)
				break
			}
		}

		if tempAppPath == "" {
			return false, fmt.Errorf("PrismLauncher.app not found in downloaded archive")
		}

		// Try to copy to Applications folder
		if err := copyDir(tempAppPath, prismAppPath); err != nil {
			return false, fmt.Errorf("failed to copy PrismLauncher to Applications folder: %w", err)
		}

		logf("%s", successLine("Prism Launcher installed in Applications folder"))

		// Create local config directory for our customizations
		configDir := getPrismConfigDir()
		os.MkdirAll(configDir, 0755)

		// macOS configuration (disable auto Java management)
		cfg := filepath.Join(configDir, "prismlauncher.cfg")
		prismConfig := `JavaDir=java
IgnoreJavaWizard=true
AutomaticJavaDownload=false
AutomaticJavaSwitch=false
UserAskedAboutAutomaticJavaDownload=true
`
		_ = os.WriteFile(cfg, []byte(prismConfig), 0644)
	} else {
		// Windows/Linux: download portable builds
		url, err = fetchLatestPrismPortableURL()
		if err != nil {
			return false, err
		}
		logf("%s", stepLine(fmt.Sprintf("Downloading Prism portable build: %s", url)))
		if err := downloadAndUnzipTo(url, dir); err != nil {
			return false, err
		}

		// Fix Qt plugin RPATH settings on Linux to ensure plugins can find bundled libraries
		if runtime.GOOS == "linux" {
			if err := fixQtPluginRPATH(dir); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to fix Qt plugin RPATH: %v", err)))
				// Don't fail the entire operation, just log the warning
			}
		}

		// Force portable mode and disable automatic Java management
		cfg := filepath.Join(dir, "prismlauncher.cfg")
		prismConfig := `Portable=true
JavaDir=java
IgnoreJavaWizard=true
AutomaticJavaDownload=false
AutomaticJavaSwitch=false
UserAskedAboutAutomaticJavaDownload=true
`
		_ = os.WriteFile(cfg, []byte(prismConfig), 0644)
	}

	return true, nil
}

// updatePrismJavaPath updates the JavaPath in prismlauncher.cfg
func updatePrismJavaPath(prismDir, javaPath string) error {
	var cfgPath string
	if runtime.GOOS == "darwin" {
		// macOS: use our custom config directory
		cfgPath = filepath.Join(getPrismConfigDir(), "prismlauncher.cfg")
	} else {
		// Windows/Linux: use portable config directory
		cfgPath = filepath.Join(prismDir, "prismlauncher.cfg")
	}

	// Read current config
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to read prismlauncher.cfg: %w", err)
	}

	// Parse and update JavaPath
	lines := strings.Split(string(content), "\n")
	var updatedLines []string
	javaPathUpdated := false

	for _, line := range lines {
		if strings.HasPrefix(line, "JavaPath=") {
			updatedLines = append(updatedLines, "JavaPath="+filepath.ToSlash(javaPath))
			javaPathUpdated = true
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	// Add JavaPath if it wasn't present
	if !javaPathUpdated {
		updatedLines = append(updatedLines, "JavaPath="+filepath.ToSlash(javaPath))
	}

	// Write updated config
	updatedContent := strings.Join(updatedLines, "\n")
	return os.WriteFile(cfgPath, []byte(updatedContent), 0644)
}

type prismRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Cross-platform Prism download with platform-specific patterns:
// - Windows: MinGW w64 portable (amd64), MSVC portable (arm64)
// - macOS: tar.gz archives with architecture-specific builds
// - Linux: tar.gz archives as fallback
func fetchLatestPrismPortableURL() (string, error) {
	// Use GitHub's releases page to find the latest Prism Launcher without API
	releasesURL := "https://github.com/PrismLauncher/PrismLauncher/releases"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(releasesURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Prism releases page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Prism releases page returned status %d", resp.StatusCode)
	}

	// Read HTML content
	htmlBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Prism releases page HTML: %w", err)
	}
	html := string(htmlBody)

	// Extract the first (latest) release tag from the releases page
	tagPattern := `/PrismLauncher/PrismLauncher/releases/tag/([^"]+)`
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindStringSubmatch(html)

	if len(tagMatches) < 2 {
		return "", errors.New("could not find any Prism Launcher release tags")
	}

	latestTag := tagMatches[1]

	// Build priority patterns by platform and arch
	var patterns []string

	if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" {
			// 1) MinGW w64 portable zip
			patterns = append(patterns, fmt.Sprintf("PrismLauncher-Windows-MinGW-w64-Portable-%s.zip", latestTag))
			// 2) MSVC portable zip
			patterns = append(patterns, fmt.Sprintf("PrismLauncher-Windows-MSVC-Portable-%s.zip", latestTag))
		} else if runtime.GOARCH == "arm64" {
			// MSVC arm64 portable zip
			patterns = append(patterns, fmt.Sprintf("PrismLauncher-Windows-MSVC-arm64-Portable-%s.zip", latestTag))
		}
		// Fallbacks for unexpected naming: generic portable zips
		patterns = append(patterns,
			fmt.Sprintf("PrismLauncher-Windows-Portable-%s.zip", latestTag),
			fmt.Sprintf("PrismLauncher-Windows-%s.zip", latestTag),
		)
	} else if runtime.GOOS == "darwin" {
		// macOS has no portable builds, only universal ZIP files
		// Priority order: main universal build first, then legacy
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-macOS-%s.zip", latestTag))

		// Fallback: legacy version for older macOS (High Sierra to Catalina)
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-macOS-Legacy-%s.zip", latestTag))

		// Last resort fallbacks
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-macos-%s.zip", latestTag))
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-darwin-%s.zip", latestTag))
	} else {
		// Linux: prioritize Qt6 Portable for better compatibility, fallback to Qt5
		// Priority 1: Qt6 Portable (newer, more compatible)
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt6-Portable-%s.tar.gz", latestTag))
		// Priority 2: Qt5 Portable (fallback for older systems)
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt5-Portable-%s.tar.gz", latestTag))
		// Fallbacks for older naming conventions
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-%s.tar.gz", latestTag))
		patterns = append(patterns, fmt.Sprintf("PrismLauncher-linux-%s.tar.gz", latestTag))
	}

	// Try each pattern to find a working download URL
	for _, assetName := range patterns {
		assetURL := fmt.Sprintf("https://github.com/PrismLauncher/PrismLauncher/releases/download/%s/%s", latestTag, assetName)

		// Verify the asset exists by making a HEAD request
		headReq, err := http.NewRequest("HEAD", assetURL, nil)
		if err != nil {
			continue
		}
		headReq.Header.Set("User-Agent", getUserAgent("General"))

		headResp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			continue
		}
		headResp.Body.Close()

		if headResp.StatusCode == 200 {
			return assetURL, nil
		}
	}

	return "", errors.New("no suitable Prism portable asset found in latest release")
}
