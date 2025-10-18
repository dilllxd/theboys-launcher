package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	// Force portable mode
	cfg := filepath.Join(dir, "prismlauncher.cfg")
	_ = os.WriteFile(cfg, []byte("Portable=true\n"), 0644)
	return true, nil
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
	api := "https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/latest"
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", getUserAgent("Prism"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var rel prismRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}

	// Build priority patterns by arch
	type pat struct{ re *regexp.Regexp }
	var patterns []pat

	if runtime.GOARCH == "amd64" {
		// 1) MinGW w64 portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MinGW-w64-Portable-.*\.zip$`)})
		// 2) MSVC portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MSVC-Portable-.*\.zip$`)})
	} else if runtime.GOARCH == "arm64" {
		// MSVC arm64 portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MSVC-arm64-Portable-.*\.zip$`)})
	}

	// Fallbacks for unexpected naming: generic portable zips
	patterns = append(patterns,
		pat{regexp.MustCompile(`(?i)Windows-.*Portable-.*\.zip$`)},
		pat{regexp.MustCompile(`(?i)Windows-.*\.zip$`)},
	)

	// Search in priority order
	for _, p := range patterns {
		for _, a := range rel.Assets {
			if p.re.MatchString(a.Name) {
				return a.URL, nil
			}
		}
	}
	return "", errors.New("no suitable Prism portable asset found in latest release")
}