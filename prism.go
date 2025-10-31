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

		// Check both naming conventions for existing installation
		prismAppPathWithoutSpace := filepath.Join(applicationsDir, "PrismLauncher.app")
		prismAppPathWithSpace := filepath.Join(applicationsDir, "Prism Launcher.app")

		// Check if Prism is already installed in Applications (try both naming conventions)
		if exists(prismAppPathWithoutSpace) {
			logf("%s", successLine("Prism Launcher found in Applications folder"))
			return false, nil
		} else if exists(prismAppPathWithSpace) {
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
				break
			}
		}

		if tempAppPath == "" {
			return false, fmt.Errorf("PrismLauncher.app not found in downloaded archive")
		}

		// Use the naming convention that matches the downloaded archive
		var targetAppPath string
		if strings.Contains(tempAppPath, "PrismLauncher.app") {
			targetAppPath = prismAppPathWithoutSpace
		} else {
			targetAppPath = prismAppPathWithSpace
		}

		// Try to copy to Applications folder
		if err := copyDir(tempAppPath, targetAppPath); err != nil {
			return false, fmt.Errorf("failed to copy PrismLauncher to Applications folder: %w", err)
		}

		// Fix executable permissions on macOS
		prismExecutable := filepath.Join(targetAppPath, "Contents", "MacOS", "prismlauncher")
		if err := setExecutablePermissions(prismExecutable); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to set executable permissions: %v", err)))
			// Don't fail the entire operation, but warn the user
		} else {
			logf("%s", successLine("Fixed executable permissions for Prism Launcher"))
		}

		// Also fix permissions for other executables in the app bundle
		macOSDir := filepath.Join(targetAppPath, "Contents", "MacOS")
		if err := fixMacOSExecutablePermissions(macOSDir); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to fix all executable permissions: %v", err)))
			// Don't fail the entire operation, but warn the user
		}

		logf("%s", successLine("Prism Launcher installed in Applications folder"))

		// Create local config directory for our customizations
		configDir := GetPrismConfigDir()
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
			// Use the actual base directory where Prism executable is located
			actualPrismDir := getPrismBaseDir(dir)
			if err := fixQtPluginRPATH(actualPrismDir); err != nil {
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
		cfgPath = filepath.Join(GetPrismConfigDir(), "prismlauncher.cfg")
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

// fixMacOSExecutablePermissions fixes permissions for all executable files in a macOS app bundle
func fixMacOSExecutablePermissions(macOSDir string) error {
	if runtime.GOOS != "darwin" {
		return nil // Only apply on macOS
	}

	if !exists(macOSDir) {
		return fmt.Errorf("macOS directory not found: %s", macOSDir)
	}

	// Walk through all files in the MacOS directory and fix executable permissions
	return filepath.Walk(macOSDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is executable (has execute bit already or is a known executable type)
		isExecutable := (info.Mode().Perm() & 0111) != 0 // Has any execute bit
		isKnownExecutable := strings.HasSuffix(info.Name(), "Updater") ||
			strings.HasSuffix(info.Name(), "Autoupdate") ||
			info.Name() == "prismlauncher"

		if isExecutable || isKnownExecutable {
			// Set executable permissions (755)
			newMode := info.Mode() | 0111 // Add execute bit for owner
			newMode = newMode | 0110      // Add execute bit for group
			newMode = newMode | 0001      // Add execute bit for others

			if newMode != info.Mode() {
				if err := os.Chmod(path, newMode); err != nil {
					logf("Failed to set permissions for %s: %v", filepath.Base(path), err)
					return err
				}
				logf("Fixed permissions for %s", filepath.Base(path))
			}
		}

		return nil
	})
}
