// Package modpack provides modpack management functionality
package modpack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"theboys-launcher/internal/logging"
)

// DownloadTask represents a download task
type DownloadTask struct {
	ID           string
	URL          string
	FilePath     string
	ExpectedSize int64
	Checksum     string // SHA256 checksum for verification
	Progress     float64
	Speed        int64 // bytes per second
	StartTime    time.Time
	Completed    bool
	Error        error
	Cancel       chan struct{}
	Done         chan struct{}
}

// DownloadManager manages download tasks
type DownloadManager struct {
	logger         *logging.Logger
	maxConcurrent  int
	activeTasks    map[string]*DownloadTask
	taskMutex      sync.RWMutex
	semaphore      chan struct{}
	downloadFolder string
}

// NewDownloadManager creates a new download manager
func NewDownloadManager(logger *logging.Logger, maxConcurrent int, downloadFolder string) *DownloadManager {
	dm := &DownloadManager{
		logger:         logger,
		maxConcurrent:  maxConcurrent,
		activeTasks:    make(map[string]*DownloadTask),
		semaphore:      make(chan struct{}, maxConcurrent),
		downloadFolder: downloadFolder,
	}

	// Ensure download folder exists
	if err := os.MkdirAll(downloadFolder, 0755); err != nil {
		logger.Error("Failed to create download folder: %v", err)
	}

	return dm
}

// DownloadFile downloads a file asynchronously
func (dm *DownloadManager) DownloadFile(url, filePath string, expectedSize int64, checksum string) (*DownloadTask, error) {
	// Generate unique task ID
	taskID := generateTaskID(url)

	// Check if already downloading
	dm.taskMutex.RLock()
	if _, exists := dm.activeTasks[taskID]; exists {
		dm.taskMutex.RUnlock()
		return nil, fmt.Errorf("download already in progress: %s", url)
	}
	dm.taskMutex.RUnlock()

	// Create task
	task := &DownloadTask{
		ID:           taskID,
		URL:          url,
		FilePath:     filePath,
		ExpectedSize: expectedSize,
		Checksum:     checksum,
		Progress:     0.0,
		Speed:        0,
		StartTime:    time.Now(),
		Completed:    false,
		Cancel:       make(chan struct{}),
		Done:         make(chan struct{}),
	}

	// Register task
	dm.taskMutex.Lock()
	dm.activeTasks[taskID] = task
	dm.taskMutex.Unlock()

	// Start download
	go dm.performDownload(task)

	return task, nil
}

// DownloadModpack downloads a complete modpack
func (dm *DownloadManager) DownloadModpack(modpack *Modpack, progressCallback func(*InstallationProgress)) error {
	dm.logger.Info("Starting modpack download: %s", modpack.Name)

	// Create progress tracker
	progress := &InstallationProgress{
		ModpackID: modpack.ID,
		Stage:     "downloading",
		Progress:  0.0,
	}

	// Download main modpack file
	tempDir := filepath.Join(dm.downloadFolder, modpack.ID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	zipPath := filepath.Join(tempDir, "modpack.zip")

	task, err := dm.DownloadFile(modpack.DownloadURL, zipPath, modpack.TotalSize, "")
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Monitor download progress
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-task.Done:
				if task.Error != nil {
					progress.Error = task.Error.Error()
				} else {
					progress.Progress = 1.0
				}
				if progressCallback != nil {
					progressCallback(progress)
				}
				return
			case <-ticker:
				progress.Progress = task.Progress
				progress.DownloadSpeed = task.Speed
				if progressCallback != nil {
					progressCallback(progress)
				}
			}
		}
	}()

	// Wait for download to complete
	<-task.Done

	if task.Error != nil {
		return fmt.Errorf("download failed: %w", task.Error)
	}

	dm.logger.Info("Successfully downloaded modpack: %s", modpack.Name)
	return nil
}

// CancelDownload cancels a download task
func (dm *DownloadManager) CancelDownload(taskID string) error {
	dm.taskMutex.RLock()
	task, exists := dm.activeTasks[taskID]
	dm.taskMutex.RUnlock()

	if !exists {
		return fmt.Errorf("download task not found: %s", taskID)
	}

	close(task.Cancel)
	return nil
}

// GetTaskStatus returns the status of a download task
func (dm *DownloadManager) GetTaskStatus(taskID string) (*DownloadTask, error) {
	dm.taskMutex.RLock()
	defer dm.taskMutex.RUnlock()

	task, exists := dm.activeTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("download task not found: %s", taskID)
	}

	return task, nil
}

// GetActiveTasks returns all active download tasks
func (dm *DownloadManager) GetActiveTasks() []*DownloadTask {
	dm.taskMutex.RLock()
	defer dm.taskMutex.RUnlock()

	tasks := make([]*DownloadTask, 0, len(dm.activeTasks))
	for _, task := range dm.activeTasks {
		if !task.Completed {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// performDownload performs the actual download
func (dm *DownloadManager) performDownload(task *DownloadTask) {
	defer close(task.Done)
	defer func() {
		// Clean up task when done
		dm.taskMutex.Lock()
		delete(dm.activeTasks, task.ID)
		dm.taskMutex.Unlock()
	}()

	// Acquire semaphore slot
	dm.semaphore <- struct{}{}
	defer func() { <-dm.semaphore }()

	dm.logger.Info("Starting download: %s", task.URL)

	// Create file
	if err := os.MkdirAll(filepath.Dir(task.FilePath), 0755); err != nil {
		task.Error = fmt.Errorf("failed to create directory: %w", err)
		return
	}

	file, err := os.Create(task.FilePath)
	if err != nil {
		task.Error = fmt.Errorf("failed to create file: %w", err)
		return
	}
	defer file.Close()

	// Make HTTP request
	req, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		task.Error = fmt.Errorf("failed to create request: %w", err)
		return
	}

	client := &http.Client{
		Timeout: 0, // No timeout for large downloads
	}

	resp, err := client.Do(req)
	if err != nil {
		task.Error = fmt.Errorf("failed to download: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		task.Error = fmt.Errorf("download failed with status: %s", resp.Status)
		return
	}

	// Get content length if not provided
	if task.ExpectedSize == 0 {
		task.ExpectedSize = resp.ContentLength
	}

	// Create progress tracker
	progress := &downloadProgress{
		task:    task,
		started: time.Now(),
		lastUpdate: time.Now(),
	}

	// Download with progress tracking
	_, err = io.CopyBuffer(file, &progressReader{
		reader:  resp.Body,
		progress: progress,
		cancel:  task.Cancel,
	}, make([]byte, 32*1024))

	if err != nil {
		// Remove partial file on error
		os.Remove(task.FilePath)
		task.Error = fmt.Errorf("download interrupted: %w", err)
		return
	}

	// Verify file size if expected size is set
	if task.ExpectedSize > 0 {
		info, err := os.Stat(task.FilePath)
		if err != nil {
			task.Error = fmt.Errorf("failed to verify file size: %w", err)
			return
		}

		if info.Size() != task.ExpectedSize {
			task.Error = fmt.Errorf("file size mismatch: expected %d, got %d", task.ExpectedSize, info.Size())
			os.Remove(task.FilePath)
			return
		}
	}

	// Verify checksum if provided
	if task.Checksum != "" {
		if err := dm.verifyChecksum(task.FilePath, task.Checksum); err != nil {
			task.Error = fmt.Errorf("checksum verification failed: %w", err)
			os.Remove(task.FilePath)
			return
		}
	}

	task.Completed = true
	task.Progress = 1.0
	dm.logger.Info("Successfully downloaded: %s", task.URL)
}

// verifyChecksum verifies the SHA256 checksum of a file
func (dm *DownloadManager) verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// downloadProgress tracks download progress
type downloadProgress struct {
	task       *DownloadTask
	started    time.Time
	lastUpdate time.Time
	lastBytes  int64
}

// progressReader wraps an io.Reader to track progress
type progressReader struct {
	reader   io.Reader
	progress *downloadProgress
	cancel   chan struct{}
}

// Read implements io.Reader
func (pr *progressReader) Read(p []byte) (n int, err error) {
	select {
	case <-pr.cancel:
		return 0, fmt.Errorf("download cancelled")
	default:
	}

	n, err = pr.reader.Read(p)
	if n > 0 {
		// Update progress
		now := time.Now()
		elapsed := now.Sub(pr.progress.started).Seconds()

		if elapsed > 0 {
			pr.progress.task.Progress = float64(pr.progress.lastBytes+int64(n)) / float64(pr.progress.task.ExpectedSize)

			// Calculate speed (bytes per second)
			if now.Sub(pr.progress.lastUpdate) >= time.Second {
				timeDiff := now.Sub(pr.progress.lastUpdate).Seconds()
				if timeDiff > 0 {
					pr.progress.task.Speed = int64(float64(pr.progress.lastBytes+int64(n)) / timeDiff)
				}
				pr.progress.lastUpdate = now
				pr.progress.lastBytes += int64(n)
			}
		}
	}

	return n, err
}

// generateTaskID generates a unique task ID from a URL
func generateTaskID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:16]
}

// CleanupTempFiles cleans up temporary download files
func (dm *DownloadManager) CleanupTempFiles() error {
	dm.logger.Info("Cleaning up temporary download files")

	entries, err := os.ReadDir(dm.downloadFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if directory is old (older than 24 hours)
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if time.Since(info.ModTime()) > 24*time.Hour {
				dirPath := filepath.Join(dm.downloadFolder, entry.Name())
				if err := os.RemoveAll(dirPath); err != nil {
					dm.logger.Error("Failed to remove temp directory %s: %v", dirPath, err)
				} else {
					dm.logger.Info("Removed old temp directory: %s", dirPath)
				}
			}
		}
	}

	return nil
}