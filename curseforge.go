package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// -------------------- CurseForge Direct Download --------------------

// downloadFromCurseForge attempts to download JAR files directly from CurseForge URLs
func downloadFromCurseForge(url, destPath string) error {
	// Handle CurseForge URLs with retry logic
	if strings.Contains(url, "curseforge.com") {
		return downloadCurseForgeFileWithRetry(url, destPath, 3)
	}

	// For non-CurseForge URLs, fall back to regular download
	return downloadTo(url, destPath, 0644)
}

// downloadCurseForgeFileWithRetry attempts to download from CurseForge with multiple retry attempts
func downloadCurseForgeFileWithRetry(pageURL, destPath string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logf("  Attempt %d/%d...", attempt, maxRetries)

		// Remove any existing partial download
		if exists(destPath) {
			os.Remove(destPath)
		}

		err := downloadCurseForgeFile(pageURL, destPath)
		if err == nil {
			return nil // Success!
		}

		lastErr = err
		logf("  Failed: %v", err)

		// Don't wait on the last attempt
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 3 * time.Second
			logf("  Retrying in %d seconds...", waitTime/time.Second)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadCurseForgeFile converts CurseForge file URLs to direct download URLs and downloads the file
func downloadCurseForgeFile(pageURL, destPath string) error {
	// Extract project info from the URL
	game, category, projectSlug, fileID, err := parseCurseForgeFileURL(pageURL)
	if err != nil {
		return fmt.Errorf("failed to parse CurseForge URL: %w", err)
	}

	// Method 1: Try the simple download URL format first
	downloadURL := fmt.Sprintf("https://www.curseforge.com/%s/%s/%s/download/%s", game, category, projectSlug, fileID)
	if err := tryDirectDownload(downloadURL, destPath); err == nil {
		return nil
	}

	// Method 2: Scrape project ID from the file page itself and use API
	projectID, err := getProjectIDFromFilePage(pageURL)
	if err == nil && projectID != "" {
		apiURL := fmt.Sprintf("https://www.curseforge.com/api/v1/mods/%s/files/%s/download", projectID, fileID)
		if err := tryDirectDownload(apiURL, destPath); err == nil {
			return nil
		}
	}

	// Method 3: Fallback to parsing the file page for download links
	return downloadCurseForgeFromPage(pageURL, destPath)
}

// getProjectIDFromFilePage scrapes the project ID from the CurseForge file page itself
func getProjectIDFromFilePage(filePageURL string) (string, error) {
	req, err := http.NewRequest("GET", filePageURL, nil)
	if err != nil {
		return "", err
	}

	// Add realistic browser headers to avoid 403 errors
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body

	// Check if content is gzipped and decompress if needed
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	html := string(body)

	// Look for the specific pattern: <dt>Project ID</dt><dd><div class="project-id-container"><span class="project-id">433760</span>
	projectIDPattern := regexp.MustCompile(`<dt>Project ID</dt>\s*<dd>\s*<div[^>]*class="project-id-container"[^>]*>\s*<span[^>]*class="project-id"[^>]*>(\d+)</span>`)
	matches := projectIDPattern.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// More flexible patterns - try many different ways the project ID might appear
	patterns := []string{
		// Various HTML patterns for project ID
		`<span[^>]*class="project-id"[^>]*>(\d+)</span>`,
		`<dt>Project ID</dt>\s*<dd>(\d+)</dd>`,
		`<dt>Project ID</dt>\s*<dd>\s*(\d+)\s*</dd>`,
		`<div[^>]*project-id[^>]*>(\d+)</div>`,
		`data-project-id="(\d+)"`,
		`project-id="(\d+)"`,

		// JSON patterns in embedded data
		`"project_id":\s*(\d+)`,
		`"projectId":\s*(\d+)`,
		`"project":\s*\{[^}]*"id":\s*(\d+)`,
		`"project":\{[^}]*"id":(\d+)`,
		`globalThis\.project[^=]*=\s*\{[^}]*"id":\s*(\d+)`,
		`window\.project[^=]*=\s*\{[^}]*"id":\s*(\d+)`,
		`"eagerProject":\{[^}]*"id":\s*(\d+)`,
		`"projectData":\{[^}]*"id":\s*(\d+)`,
		`"data":[^}]*"id":\s*(\d+)`,

		// More general numeric patterns in project context
		`"file_id":\d+.*?"project_id":\s*(\d+)`,
		`"fileId":\d+.*?"projectId":\s*(\d+)`,
		`"slug":"[^"]*".*?"id":(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			projectID := matches[1]
			if len(projectID) > 0 && len(projectID) < 10 {
				return projectID, nil
			}
		}
	}

	return "", fmt.Errorf("project ID not found in file page")
}

// parseCurseForgeFileURL extracts game, category, project slug, and file ID from a CurseForge file URL
func parseCurseForgeFileURL(url string) (game, category, projectSlug, fileID string, err error) {
	// Pattern: https://www.curseforge.com/minecraft/mc-mods/mod-name/files/1234567
	// Pattern: https://www.curseforge.com/minecraft/texture-packs/pack-name/files/1234567
	re := regexp.MustCompile(`https://www\.curseforge\.com/([^/]+)/([^/]+)/([^/]+)/files/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 5 {
		return "", "", "", "", fmt.Errorf("invalid CurseForge URL format")
	}

	return matches[1], matches[2], matches[3], matches[4], nil
}

// tryDirectDownload attempts to download from a URL that should be a direct download link
func tryDirectDownload(downloadURL, destPath string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", getUserAgent("General"))

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirects for direct downloads
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Check if we got a file (check content type and disposition)
	contentType := resp.Header.Get("Content-Type")
	contentDisp := resp.Header.Get("Content-Disposition")

	isFile := strings.Contains(contentType, "application/java-archive") ||
		strings.Contains(contentType, "application/zip") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentDisp, ".jar") ||
		strings.Contains(contentDisp, ".zip")

	if !isFile {
		return fmt.Errorf("response doesn't appear to be a file (Content-Type: %s)", contentType)
	}

	// Download the file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return os.WriteFile(destPath, body, 0644)
}

// downloadCurseForgeFromPage fallback method that parses the page HTML
func downloadCurseForgeFromPage(pageURL, destPath string) error {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", getUserAgent("General"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch CurseForge page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body

	// Check if content is gzipped and decompress if needed
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read page content: %w", err)
	}

	// Look for download link patterns in the HTML
	downloadURL := extractCurseForgeDownloadLink(string(body))
	if downloadURL != "" {
		return downloadTo(downloadURL, destPath, 0644)
	}

	return fmt.Errorf("could not extract direct download link from CurseForge page")
}

// extractCurseForgeDownloadLink attempts to find the direct download URL from CurseForge page HTML
func extractCurseForgeDownloadLink(html string) string {
	// Look for various CurseForge download patterns
	patterns := []string{
		`"downloadUrl":"([^"]+\.jar)"`,
		`"url":"([^"]+\.jar)"`,
		`href="([^"]+\.jar)"`,
		`data-download="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			// Unescape JSON-encoded URL if needed
			url := strings.ReplaceAll(matches[1], "\\u0026", "&")
			url = strings.ReplaceAll(url, "\\", "")
			return url
		}
	}

	return ""
}

// -------------------- Packwiz manual download parsing & assist --------------------

type manualItem struct {
	Name string // optional; may be empty
	URL  string
	Path string // absolute path from packwiz message
}

// Parse lines like:
// "Mod Name: ... Please go to https://... and save this file to C:\...\mods\file.jar"
// The actual packwiz output format: "Mod Name: java.lang.Exception: Please go to URL and save this file to PATH"
func parsePackwizManuals(s string) []manualItem {
	// Pattern to match the exact packwiz output format:
	// "Aquaculture Delight (A Farmer's Delight Add-on): java.lang.Exception: This mod is excluded from the CurseForge API and must be downloaded manually."
	// "Please go to https://www.curseforge.com/minecraft/mc-mods/aquaculture-delight/files/6259758 and save this file to C:\path\to\file.jar"

	// Use a more flexible approach to find mod errors and URLs
	lines := strings.Split(s, "\n")
	seen := map[string]bool{}
	var items []manualItem

	for i, line := range lines {
		// Look for lines with mod errors
		if strings.Contains(line, "java.lang.Exception: This mod is excluded from the CurseForge API and must be downloaded manually.") {
			// Extract mod name from the beginning of the line
			if colonIndex := strings.Index(line, ":"); colonIndex > 0 {
				name := strings.TrimSpace(line[:colonIndex])

				// Filter out non-mod entries
				if strings.Contains(name, "Current version") ||
					strings.Contains(name, "at link.infra") ||
					strings.Contains(name, "java.base") ||
					len(name) == 0 {
					continue
				}

				// Look for the corresponding download URL in the next few lines
				for j := i + 1; j < i+10 && j < len(lines); j++ {
					nextLine := strings.TrimSpace(lines[j])
					if strings.Contains(nextLine, "Please go to ") && strings.Contains(nextLine, "curseforge.com") {
						// Extract URL and path
						if strings.Contains(nextLine, " and save this file to ") {
							parts := strings.Split(nextLine, " and save this file to ")
							if len(parts) == 2 {
								url := strings.TrimSpace(strings.TrimPrefix(parts[0], "Please go to "))
								path := strings.TrimSpace(parts[1])

								// Validate URL
								if strings.Contains(url, "curseforge.com") && url != "" && path != "" {
									key := url + "|" + strings.ToLower(path)
									if !seen[key] {
										seen[key] = true
										items = append(items, manualItem{Name: name, URL: url, Path: path})
										break // Found the URL for this mod, move to next mod
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return items
}

func assistManualFromPackwiz(items []manualItem) {
	if len(items) == 0 {
		return
	}

	logf("Downloading %d manual mod(s) directly from CurseForge...", len(items))

	// Ensure destination folders exist
	for _, it := range items {
		_ = os.MkdirAll(filepath.Dir(it.Path), 0755)
	}

	// Download all files directly
	var failedItems []manualItem
	for _, it := range items {
		logf("Downloading %s...", it.Name)
		logf("  From: %s", it.URL)
		logf("  To:   %s", it.Path)

		if err := downloadFromCurseForge(it.URL, it.Path); err != nil {
			logf("  Failed: %v", err)
			failedItems = append(failedItems, it)
		} else {
			logf("  ✓ Downloaded successfully")
		}
	}

	// Handle any failed downloads
	if len(failedItems) > 0 {
		logf("\n%d download(s) failed. These may require manual download:", len(failedItems))
		for _, it := range failedItems {
			logf(" - %s\n   %s\n   Save as: %s", it.Name, it.URL, it.Path)
		}

		if yesNoBox("Some downloads failed. Open remaining pages in browser?", launcherName+" - Download Failed") {
			for _, it := range failedItems {
				_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
			}

			// Wait for manual downloads
			for {
				logf("\nPress Enter after saving the files to re-check…")
				waitEnter()

				still := failedItems[:0]
				for _, it := range failedItems {
					if !exists(it.Path) {
						still = append(still, it)
					}
				}
				failedItems = still
				if len(failedItems) == 0 {
					logf("All manual items found. Continuing…")
					return
				}

				logf("Still missing:")
				for _, it := range failedItems {
					logf(" - %s -> %s", it.Name, it.Path)
				}
				if yesNoBox("Open the pages again?", launcherName) {
					for _, it := range failedItems {
						_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
					}
				}
			}
		}
	} else {
		logf("All manual downloads completed successfully!")
	}
}
