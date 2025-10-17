package launcher

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// CurseForgeMod represents a CurseForge mod file
type CurseForgeMod struct {
	ID           int    `json:"id"`
	DisplayName  string `json:"displayName"`
	FileName     string `json:"fileName"`
	FileDate     string `json:"fileDate"`
	FileLength   int64  `json:"fileLength"`
	DownloadURL  string `json:"downloadUrl"`
	GameVersions []string `json:"gameVersions"`
	ModLoader    string `json:"modLoader"`
}

// CurseForgeProject represents a CurseForge project
type CurseForgeProject struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Summary     string `json:"summary"`
	DownloadCount int64 `json:"downloadCount"`
	WebsiteURL  string `json:"websiteUrl"`
}

// CurseForgeSearchResult represents search results from CurseForge
type CurseForgeSearchResult struct {
	Data []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		Summary      string `json:"summary"`
		DownloadCount int64 `json:"downloadCount"`
		LatestFiles  []struct {
			ID           int    `json:"id"`
			DisplayName  string `json:"displayName"`
			FileName     string `json:"fileName"`
			FileDate     string `json:"fileDate"`
			FileLength   int64  `json:"fileLength"`
			DownloadURL  string `json:"downloadUrl"`
			GameVersions []string `json:"gameVersions"`
			ModLoaders   []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"modLoaders"`
		} `json:"latestFiles"`
	} `json:"data"`
}

// CurseForgeManager handles CurseForge integration
type CurseForgeManager struct {
	platform   platform.Platform
	logger     logging.Logger
	downloader *Downloader
	client     *http.Client
}

// NewCurseForgeManager creates a new CurseForge manager instance
func NewCurseForgeManager(platform platform.Platform, logger logging.Logger) *CurseForgeManager {
	return &CurseForgeManager{
		platform:   platform,
		logger:     logger,
		downloader: NewDownloader(platform, logger),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseCurseForgeURL parses various CurseForge URL formats to extract file information
func (cf *CurseForgeManager) ParseCurseForgeURL(cfURL string) (*CurseForgeMod, error) {
	cf.logger.Info("Parsing CurseForge URL: %s", cfURL)

	// Try different URL patterns
	patterns := []struct {
		pattern *regexp.Regexp
		handler func([]string) (*CurseForgeMod, error)
	}{
		{
			regexp.MustCompile(`curseforge\.com/minecraft/mc-mods/([^/]+)/files/(\d+)`),
			cf.handleFileURL,
		},
		{
			regexp.MustCompile(`curseforge\.com/minecraft/mc-mods/([^/]+)`),
			cf.handleProjectURL,
		},
		{
			regexp.MustCompile(`mediafilez\.forgecdn\.net/files/(\d+)/(\d+)/([^?]+)`),
			cf.handleDirectURL,
		},
		{
			regexp.MustCompile(`edgeforge\.net/files/(\d+)/(\d+)/([^?]+)`),
			cf.handleDirectURL,
		},
	}

	for _, p := range patterns {
		if matches := p.pattern.FindStringSubmatch(cfURL); matches != nil {
			return p.handler(matches)
		}
	}

	return nil, fmt.Errorf("unsupported CurseForge URL format: %s", cfURL)
}

// handleFileURL handles direct file URLs like /mc-mods/modname/files/123456
func (cf *CurseForgeManager) handleFileURL(matches []string) (*CurseForgeMod, error) {
	slug := matches[1]
	fileID := matches[2]

	cf.logger.Info("Found project slug: %s, file ID: %s", slug, fileID)

	// Try to get file info through API (requires API key, so we'll construct basic info)
	return &CurseForgeMod{
		ID:          parseIntSafe(fileID),
		FileName:    fmt.Sprintf("%s-file-%s.jar", slug, fileID),
		DisplayName: fmt.Sprintf("%s File %s", slug, fileID),
		DownloadURL: matches[0], // Use the full matched URL
	}, nil
}

// handleProjectURL handles project URLs like /mc-mods/modname
func (cf *CurseForgeManager) handleProjectURL(matches []string) (*CurseForgeMod, error) {
	slug := matches[1]

	cf.logger.Info("Found project slug: %s", slug)

	// For project URLs, we'd need to search the project and get the latest file
	// For now, return an error indicating we need a direct file URL
	return nil, fmt.Errorf("project URLs not yet supported, please use a direct file URL with /files/123456")
}

// handleDirectURL handles direct download URLs from CDN
func (cf *CurseForgeManager) handleDirectURL(matches []string) (*CurseForgeMod, error) {
	projectID := matches[1]
	fileID := matches[2]
	fileName := matches[3]

	cf.logger.Info("Found direct download: project %s, file %s, name %s", projectID, fileID, fileName)

	return &CurseForgeMod{
		ID:          parseIntSafe(fileID),
		FileName:    fileName,
		DisplayName: fileName,
		DownloadURL: matches[0], // Use the full matched URL
	}, nil
}

// SearchCurseForge searches for mods on CurseForge
func (cf *CurseForgeManager) SearchCurseForge(query string, gameVersion string, limit int) ([]CurseForgeMod, error) {
	cf.logger.Info("Searching CurseForge for: %s (version: %s)", query, gameVersion)

	// Note: CurseForge API requires authentication for full functionality
	// This is a simplified implementation that would need API keys for production

	// For now, return an empty slice with a helpful message
	return []CurseForgeMod{}, fmt.Errorf("CurseForge API search requires authentication - please use direct file URLs")
}

// DownloadMod downloads a mod from CurseForge to the specified directory
func (cf *CurseForgeManager) DownloadMod(mod *CurseForgeMod, downloadDir string, progressCallback func(*DownloadProgress)) (string, error) {
	cf.logger.Info("Downloading mod: %s (%s)", mod.DisplayName, mod.FileName)

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Construct output path
	outputPath := filepath.Join(downloadDir, mod.FileName)

	// Use the downloader to download the file
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	err := cf.downloader.DownloadFile(ctx, mod.DownloadURL, outputPath, progressCallback)
	if err != nil {
		return "", fmt.Errorf("failed to download mod: %w", err)
	}

	cf.logger.Info("Successfully downloaded mod to: %s", outputPath)
	return outputPath, nil
}

// AssistManualDownload assists user with manual download for failed downloads
func (cf *CurseForgeManager) AssistManualDownload(mod *CurseForgeMod, instancePath string) error {
	cf.logger.Info("Assisting manual download for: %s", mod.DisplayName)

	// Create mods directory if it doesn't exist
	modsDir := filepath.Join(instancePath, "mods")
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods directory: %w", err)
	}

	// Instructions for user
	instructions := fmt.Sprintf(`
Manual Download Required for: %s

Due to download restrictions or network issues, please download the mod manually:

1. Open your web browser and go to: %s
2. Download the file: %s
3. Place the downloaded file in: %s
4. The mod will be available when you launch the instance

File Details:
- Name: %s
- Size: %s
- ID: %d

Press Enter to continue once you've noted these instructions...`,
		mod.DisplayName,
		mod.DownloadURL,
		mod.FileName,
		modsDir,
		mod.DisplayName,
		formatBytes(mod.FileLength),
		mod.ID,
	)

	cf.logger.Info(instructions)

	// In a GUI context, this would open a dialog or browser
	// For now, we'll create a help file
	helpFile := filepath.Join(instancePath, "MANUAL_DOWNLOAD_NEEDED.txt")
	if err := os.WriteFile(helpFile, []byte(instructions), 0644); err != nil {
		cf.logger.Warn("Failed to create help file: %v", err)
	}

	return nil
}

// RetryDownloadWithBackoff retries a download with exponential backoff
func (cf *CurseForgeManager) RetryDownloadWithBackoff(mod *CurseForgeMod, downloadDir string, maxRetries int, progressCallback func(*DownloadProgress)) (string, error) {
	cf.logger.Info("Attempting download with retry for: %s", mod.FileName)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(1<<uint(attempt)) * time.Second
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}

			cf.logger.Info("Retrying download in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		filePath, err := cf.DownloadMod(mod, downloadDir, progressCallback)
		if err == nil {
			cf.logger.Info("Download successful on attempt %d", attempt+1)
			return filePath, nil
		}

		lastErr = err
		cf.logger.Warn("Download attempt %d failed: %v", attempt+1, err)
	}

	// All retries failed, assist with manual download
	cf.logger.Error("All download attempts failed for %s: %v", mod.FileName, lastErr)

	// Create placeholder for manual download
	instancePath := filepath.Dir(downloadDir)
	if err := cf.AssistManualDownload(mod, instancePath); err != nil {
		cf.logger.Warn("Failed to create manual download assistance: %v", err)
	}

	return "", fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

// GetModInfoFromURL extracts mod information from a CurseForge URL
func (cf *CurseForgeManager) GetModInfoFromURL(cfURL string) (*CurseForgeMod, error) {
	// Try to parse the URL
	mod, err := cf.ParseCurseForgeURL(cfURL)
	if err != nil {
		return nil, err
	}

	// If we don't have enough info, try to get more details
	if mod.DisplayName == "" || mod.FileName == "" {
		// This would require API calls for full functionality
		cf.logger.Warn("Limited information available for URL: %s", cfURL)
	}

	return mod, nil
}

// ValidateCurseForgeURL checks if a URL is a valid CurseForge URL
func (cf *CurseForgeManager) ValidateCurseForgeURL(cfURL string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`curseforge\.com/minecraft/mc-mods/[^/]+/files/\d+`),
		regexp.MustCompile(`curseforge\.com/minecraft/mc-mods/[^/]+`),
		regexp.MustCompile(`mediafilez\.forgecdn\.net/files/\d+/\d+`),
		regexp.MustCompile(`edgeforge\.net/files/\d+/\d+`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(cfURL) {
			return true
		}
	}

	return false
}

// Helper functions

func parseIntSafe(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}