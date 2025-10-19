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
	if exists(filepath.Join(dir, "PrismLauncher.exe")) {
		return false, nil
	}
	url, err := fetchLatestPrismPortableURL()
	if err != nil {
		return false, err
	}
	logf("%s", stepLine(fmt.Sprintf("Downloading Prism portable build: %s", url)))
	if err := downloadAndUnzipTo(url, dir); err != nil {
		return false, err
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
	return true, nil
}

// updatePrismJavaPath updates the JavaPath in prismlauncher.cfg
func updatePrismJavaPath(prismDir, javaPath string) error {
	cfgPath := filepath.Join(prismDir, "prismlauncher.cfg")

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

// Prefer MinGW w64 portable on amd64; fall back to MSVC portable.
// On arm64, use MSVC arm64 portable.
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

	// Build priority patterns by arch
	var patterns []string

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
