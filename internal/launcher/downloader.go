package launcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// Downloader handles file downloads with progress tracking
type Downloader struct {
	platform platform.Platform
	logger   logging.Logger
	client   *http.Client
}

// NewDownloader creates a new downloader instance
func NewDownloader(platform platform.Platform, logger logging.Logger) *Downloader {
	return &Downloader{
		platform: platform,
		logger:   logger,
		client: &http.Client{
			Timeout: 30 * time.Minute, // Long timeout for large downloads
		},
	}
}

// DownloadProgress represents download progress information
type DownloadProgress struct {
	URL             string
	TotalBytes      int64
	DownloadedBytes int64
	Percentage      float64
	Speed           int64 // bytes per second
	ETA             int64 // estimated time remaining in seconds
	Status          string
	Error           string
}

// DownloadFile downloads a file with progress tracking
func (d *Downloader) DownloadFile(ctx context.Context, url, outputPath string, progressCallback func(*DownloadProgress)) error {
	return d.DownloadFileWithRetry(ctx, url, outputPath, progressCallback, 3)
}

// DownloadFileWithRetry downloads a file with retry logic
func (d *Downloader) DownloadFileWithRetry(ctx context.Context, url, outputPath string, progressCallback func(*DownloadProgress), maxRetries int) error {
	d.logger.Info("Starting download: %s (max retries: %d)", url, maxRetries)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(1<<uint(attempt)) * time.Second
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}

			d.logger.Info("Retrying download in %v (attempt %d/%d)", delay, attempt+1, maxRetries)

			// Update progress for retry
			if progressCallback != nil {
				progress := &DownloadProgress{
					URL:    url,
					Status: fmt.Sprintf("retrying in %v", delay),
					Error:  fmt.Sprintf("Attempt %d failed, retrying...", attempt+1),
				}
				progressCallback(progress)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		err := d.downloadFileAttempt(ctx, url, outputPath, progressCallback, attempt+1, maxRetries)
		if err == nil {
			d.logger.Info("Download successful on attempt %d", attempt+1)
			return nil
		}

		lastErr = err
		d.logger.Warn("Download attempt %d failed: %v", attempt+1, err)

		// Check if error is retryable
		if !d.isRetryableError(err) {
			d.logger.Info("Error is not retryable, giving up: %v", err)
			break
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadFileAttempt performs a single download attempt
func (d *Downloader) downloadFileAttempt(ctx context.Context, url, outputPath string, progressCallback func(*DownloadProgress), attempt, maxAttempts int) error {
	d.logger.Info("Download attempt %d/%d: %s", attempt, maxAttempts, url)

	// Create progress tracker
	progress := &DownloadProgress{
		URL:    url,
		Status: "starting",
	}

	if progressCallback != nil {
		progressCallback(progress)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		progress.Status = "error"
		progress.Error = fmt.Sprintf("Failed to create request: %v", err)
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", d.getUserAgent("Downloader"))

	// Start download
	resp, err := d.client.Do(req)
	if err != nil {
		progress.Status = "error"
		progress.Error = fmt.Sprintf("Download failed: %v", err)
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress.Status = "error"
		progress.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Get content length
	progress.TotalBytes = resp.ContentLength
	if progress.TotalBytes <= 0 {
		d.logger.Warn("Unknown content length for %s", url)
		progress.TotalBytes = 0
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := d.platform.CreateDirectory(dir); err != nil {
		progress.Status = "error"
		progress.Error = fmt.Sprintf("Failed to create output directory: %v", err)
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		progress.Status = "error"
		progress.Error = fmt.Sprintf("Failed to create output file: %v", err)
		if progressCallback != nil {
			progressCallback(progress)
		}
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Start progress tracking
	progress.Status = "downloading"
	startTime := time.Now()
	lastUpdate := startTime

	if progressCallback != nil {
		progressCallback(progress)
	}

	// Download with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64

	for {
		select {
		case <-ctx.Done():
			progress.Status = "cancelled"
			progress.Error = ctx.Err().Error()
			if progressCallback != nil {
				progressCallback(progress)
			}
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, err := file.Write(buf[:n])
			if err != nil {
				progress.Status = "error"
				progress.Error = fmt.Sprintf("Write error: %v", err)
				if progressCallback != nil {
					progressCallback(progress)
				}
				return fmt.Errorf("write error: %w", err)
			}

			downloaded += int64(written)
			progress.DownloadedBytes = downloaded

			// Update progress calculations
			if progress.TotalBytes > 0 {
				progress.Percentage = float64(downloaded) / float64(progress.TotalBytes) * 100
			}

			// Update speed and ETA (throttled to avoid too frequent updates)
			now := time.Now()
			if now.Sub(lastUpdate) >= 100*time.Millisecond {
				elapsed := now.Sub(startTime).Seconds()
				if elapsed > 0 {
					progress.Speed = int64(float64(downloaded) / elapsed)
				}

				if progress.Speed > 0 && progress.TotalBytes > 0 {
					remaining := progress.TotalBytes - downloaded
					progress.ETA = remaining / progress.Speed
				}

				if progressCallback != nil {
					progressCallback(progress)
				}

				lastUpdate = now
			}
		}

		if err != nil {
			if err == io.EOF {
				break // Download complete
			}
			progress.Status = "error"
			progress.Error = fmt.Sprintf("Read error: %v", err)
			if progressCallback != nil {
				progressCallback(progress)
			}
			return fmt.Errorf("read error: %w", err)
		}
	}

	// Final progress update
	progress.Status = "completed"
	progress.Percentage = 100
	progress.ETA = 0
	if progressCallback != nil {
		progressCallback(progress)
	}

	d.logger.Info("Download completed: %s (%s)", outputPath, formatBytes(downloaded))
	return nil
}

// DownloadTo downloads a file to the specified path with executable permissions
func (d *Downloader) DownloadTo(url, outputPath string, mode os.FileMode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	err := d.DownloadFile(ctx, url, outputPath, nil)
	if err != nil {
		return err
	}

	// Set file permissions
	if mode != 0 {
		return os.Chmod(outputPath, mode)
	}

	return nil
}

// DownloadToMemory downloads a file into memory
func (d *Downloader) DownloadToMemory(ctx context.Context, url string) ([]byte, error) {
	d.logger.Info("Downloading to memory: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.getUserAgent("Downloader"))

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	d.logger.Info("Downloaded %d bytes to memory", len(data))
	return data, nil
}

// IsURLValid checks if a URL is accessible
func (d *Downloader) IsURLValid(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", d.getUserAgent("Downloader"))

	resp, err := d.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetURLSize returns the size of a file at a URL
func (d *Downloader) GetURLSize(url string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.getUserAgent("Downloader"))

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp.ContentLength, nil
}

// isRetryableError determines if an error is worth retrying
func (d *Downloader) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network-related errors that are typically retryable
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"server returned 502",
		"server returned 503",
		"server returned 504",
		"server returned 429", // Too Many Requests
		"read: connection reset by peer",
		"write: connection reset by peer",
		"broken pipe",
		"ssl handshake timeout",
		"tls handshake timeout",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	// HTTP status codes that are retryable
	if strings.Contains(errStr, "HTTP 502") ||
	   strings.Contains(errStr, "HTTP 503") ||
	   strings.Contains(errStr, "HTTP 504") ||
	   strings.Contains(errStr, "HTTP 429") {
		return true
	}

	// Don't retry on client errors (4xx range) except 429
	if strings.Contains(errStr, "HTTP 4") && !strings.Contains(errStr, "HTTP 429") {
		return false
	}

	// Don't retry on file system errors
	if strings.Contains(errStr, "permission denied") ||
	   strings.Contains(errStr, "no such file") ||
	   strings.Contains(errStr, "disk full") {
		return false
	}

	// Default to retrying for unknown errors
	return true
}

// getUserAgent returns a user agent string
func (d *Downloader) getUserAgent(component string) string {
	return fmt.Sprintf("TheBoys-%s/dev", component)
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// ExtractFilenameFromURL extracts a filename from a URL
func ExtractFilenameFromURL(url string) string {
	// Remove query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Extract filename
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if filename != "" {
			return filename
		}
	}

	return "download"
}