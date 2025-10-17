package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// Version information (populated at build time)
var Version = "dev"

// Updater handles application self-updates
type Updater struct {
	platform     platform.Platform
	logger       logging.Logger
	repoOwner    string
	repoName     string
	currentVersion string
}

// NewUpdater creates a new updater instance
func NewUpdater(platform platform.Platform, logger logging.Logger) *Updater {
	return &Updater{
		platform:       platform,
		logger:         logger,
		repoOwner:      "theboys-launcher", // This should be the actual repo
		repoName:       "theboys-launcher",
		currentVersion: Version,
	}
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available     bool      `json:"available"`
	CurrentVersion string    `json:"currentVersion"`
	LatestVersion string    `json:"latestVersion"`
	ReleaseName   string    `json:"releaseName"`
	ReleaseNotes  string    `json:"releaseNotes"`
	DownloadURL   string    `json:"downloadURL"`
	FileSize      int64     `json:"fileSize"`
	PublishedAt   time.Time `json:"publishedAt"`
}

// CheckForUpdates checks if there are updates available
func (u *Updater) CheckForUpdates() (*UpdateInfo, error) {
	u.logger.Info("Checking for updates...")

	if u.currentVersion == "dev" {
		u.logger.Info("Development version detected, skipping update check")
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: u.currentVersion,
		}, nil
	}

	// Fetch latest release from GitHub
	release, err := u.fetchLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	// Compare versions
	updateAvailable := u.isNewerVersion(release.TagName, u.currentVersion)

	updateInfo := &UpdateInfo{
		Available:      updateAvailable,
		CurrentVersion: u.currentVersion,
		LatestVersion:  release.TagName,
		ReleaseName:    release.Name,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	if updateAvailable {
		// Find appropriate download asset
		asset := u.findDownloadAsset(release.Assets)
		if asset != nil {
			updateInfo.DownloadURL = asset.BrowserDownloadURL
			updateInfo.FileSize = asset.Size
		}
	}

	u.logger.Info("Update check complete: %s -> %s (available: %v)",
		u.currentVersion, release.TagName, updateAvailable)

	return updateInfo, nil
}

// fetchLatestRelease fetches the latest release from GitHub
func (u *Updater) fetchLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.repoOwner, u.repoName)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub release: %w", err)
	}

	return &release, nil
}

// findDownloadAsset finds the appropriate download asset for the current platform
func (u *Updater) findDownloadAsset(assets []GitHubReleaseAsset) *GitHubReleaseAsset {
	// Determine expected filename based on platform
	var expectedName string
	switch runtime.GOOS {
	case "windows":
		expectedName = "theboys-launcher.exe"
	case "darwin":
		expectedName = "theboys-launcher-darwin"
	case "linux":
		if runtime.GOARCH == "arm64" {
			expectedName = "theboys-launcher-linux-arm64"
		} else {
			expectedName = "theboys-launcher-linux"
		}
	default:
		expectedName = "theboys-launcher"
	}

	// Look for matching asset
	for _, asset := range assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(expectedName)) {
			return &asset
		}
	}

	// Fallback: look for any executable file
	for _, asset := range assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".exe") ||
		   strings.HasSuffix(strings.ToLower(asset.Name), "-darwin") ||
		   strings.HasSuffix(strings.ToLower(asset.Name), "-linux") {
			return &asset
		}
	}

	return nil
}

// isNewerVersion checks if the new version is newer than the current version
func (u *Updater) isNewerVersion(newVersion, currentVersion string) bool {
	// Remove 'v' prefix if present
	newVer := strings.TrimPrefix(newVersion, "v")
	curVer := strings.TrimPrefix(currentVersion, "v")

	// Simple semantic version comparison
	newParts := strings.Split(newVer, ".")
	curParts := strings.Split(curVer, ".")

	for i := 0; i < 3; i++ {
		if i >= len(newParts) || i >= len(curParts) {
			break
		}

		var newNum, curNum int
		_, err1 := fmt.Sscanf(newParts[i], "%d", &newNum)
		_, err2 := fmt.Sscanf(curParts[i], "%d", &curNum)

		if err1 != nil || err2 != nil {
			// Fallback to string comparison
			return newVersion != currentVersion
		}

		if newNum > curNum {
			return true
		}
		if newNum < curNum {
			return false
		}
	}

	return false
}

// DownloadUpdate downloads the update to a temporary file
func (u *Updater) DownloadUpdate(downloadURL string, progressCallback ProgressCallback) (string, error) {
	u.logger.Info("Downloading update from %s", downloadURL)

	// Create temporary file
	tempFile, err := os.CreateTemp("", "theboys-launcher-update-*.exe")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()

	// Download the update
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempPath)
		return "", fmt.Errorf("update download failed: HTTP %d", resp.StatusCode)
	}

	// Get content length for progress tracking
	contentLength := resp.ContentLength

	// Copy file with progress tracking
	if progressCallback != nil {
		progressCallback(0.0)
	}

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, err := tempFile.Write(buf[:n])
			if err != nil {
				os.Remove(tempPath)
				return "", fmt.Errorf("failed to write update file: %w", err)
			}
			downloaded += int64(written)

			if progressCallback != nil && contentLength > 0 {
				progress := float64(downloaded) / float64(contentLength)
				if progress > 1.0 {
					progress = 1.0
				}
				progressCallback(progress)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempPath)
			return "", fmt.Errorf("download interrupted: %w", err)
		}
	}

	if progressCallback != nil {
		progressCallback(1.0)
	}

	u.logger.Info("Update downloaded to %s", tempPath)
	return tempPath, nil
}

// InstallUpdate performs the update installation
func (u *Updater) InstallUpdate(updatePath string) error {
	u.logger.Info("Installing update from %s", updatePath)

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Check if we're running from an installed location
	isInstalled := u.platform.IsInstalled()
	installPath, _ := u.platform.GetInstallationPath()

	u.logger.Info("Current executable: %s", currentExe)
	u.logger.Info("Running from installed location: %v", isInstalled)
	if installPath != "" {
		u.logger.Info("Installation path: %s", installPath)
	}

	// Create update script with installation awareness
	var scriptPath string
	switch runtime.GOOS {
	case "windows":
		scriptPath = u.createWindowsUpdateScript(currentExe, updatePath, isInstalled, installPath)
	case "darwin", "linux":
		scriptPath = u.createUnixUpdateScript(currentExe, updatePath, isInstalled, installPath)
	default:
		return fmt.Errorf("unsupported platform for update")
	}

	if scriptPath == "" {
		return fmt.Errorf("failed to create update script")
	}

	// Execute update script
	u.logger.Info("Starting update process...")
	cmd := exec.Command(scriptPath)

	// For Windows, we need to start the process and exit immediately
	if runtime.GOOS == "windows" {
		cmd.Start()
		os.Exit(0)
		return nil
	}

	// For Unix systems, we can wait for the script to complete
	return cmd.Run()
}

// createWindowsUpdateScript creates a batch script for Windows updates
func (u *Updater) createWindowsUpdateScript(currentExe, updatePath string, isInstalled bool, installPath string) string {
	var scriptContent string

	if isInstalled && installPath != "" {
		// Update installed application
		scriptContent = fmt.Sprintf(`@echo off
echo Updating TheBoys Launcher (Installed Version)...
echo Current: %s
echo Update: %s
echo Install Path: %s
timeout /t 2 /nobreak >nul

:: Check if we have admin rights for system installation
net session >nul 2>&1
if %%errorlevel%% equ 0 (
    echo Running with administrator privileges
    move /Y "%s" "%s.old.bak" >nul 2>&1
    move /Y "%s" "%s" >nul
    if %%errorlevel%% neq 0 (
        echo Failed to update installed version
        move /Y "%s.old.bak" "%s" >nul 2>&1
        pause
        exit /b 1
    )
    del "%s.old.bak" >nul 2>&1
    echo Successfully updated installed version
    start "" "%s"
) else (
    echo No administrator privileges detected
    echo Please run TheBoys Launcher as administrator to update
    pause
    exit /b 1
)

del "%%~f0"
`,
			currentExe, updatePath, installPath,
			currentExe, currentExe+".old.bak",
			updatePath, currentExe,
			currentExe+".old.bak", currentExe,
			currentExe)
	} else {
		// Update portable/non-installed application
		scriptContent = fmt.Sprintf(`@echo off
echo Updating TheBoys Launcher (Portable Version)...
echo Current: %s
echo Update: %s
timeout /t 2 /nobreak >nul

move /Y "%s" "%s.old" >nul
move /Y "%s" "%s" >nul
if %%errorlevel%% neq 0 (
    echo Update failed, restoring backup
    move /Y "%s.old" "%s" >nul
    pause
    exit /b 1
)

del "%s.old" >nul 2>&1
echo Successfully updated portable version
start "" "%s"

del "%%~f0"
`,
			currentExe, updatePath,
			currentExe, currentExe+".old",
			updatePath, currentExe,
			currentExe+".old", currentExe,
			currentExe+".old",
			currentExe)
	}

	scriptPath := filepath.Join(os.TempDir(), "theboys-update.bat")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		u.logger.Error("Failed to create update script: %v", err)
		return ""
	}

	return scriptPath
}

// createUnixUpdateScript creates a shell script for Unix updates
func (u *Updater) createUnixUpdateScript(currentExe, updatePath string, isInstalled bool, installPath string) string {
	var scriptContent string

	if isInstalled && installPath != "" {
		// Update installed application
		scriptContent = fmt.Sprintf(`#!/bin/bash
echo "Updating TheBoys Launcher (Installed Version)..."
echo "Current: %s"
echo "Update: %s"
echo "Install Path: %s"

# Check write permissions
if [ ! -w "$(dirname "%s")" ]; then
    echo "Error: No write permissions to installation directory"
    echo "Please run with appropriate privileges (sudo on system installations)"
    echo "Or reinstall using your package manager"
    read -p "Press Enter to continue..."
    exit 1
fi

sleep 2

# Create backup
mv "%s" "%s.old.bak" || {
    echo "Failed to create backup"
    exit 1
}

# Install update
mv "%s" "%s" || {
    echo "Failed to install update, restoring backup"
    mv "%s.old.bak" "%s"
    exit 1
}

# Set permissions
chmod +x "%s"

# Clean up
rm -f "%s.old.bak" 2>/dev/null

echo "Successfully updated installed version"
exec "%s" --cleanup-after-update "%s.old.bak" "%s"

rm -f "$0"
`,
			currentExe, updatePath, installPath, currentExe,
			currentExe, currentExe+".old.bak",
			updatePath, currentExe,
			currentExe+".old.bak", currentExe,
			currentExe,
			currentExe+".old.bak", currentExe)
	} else {
		// Update portable/non-installed application
		scriptContent = fmt.Sprintf(`#!/bin/bash
echo "Updating TheBoys Launcher (Portable Version)..."
echo "Current: %s"
echo "Update: %s"

sleep 2

# Create backup
mv "%s" "%s.old" || {
    echo "Failed to create backup"
    exit 1
}

# Install update
mv "%s" "%s" || {
    echo "Failed to install update, restoring backup"
    mv "%s.old" "%s"
    exit 1
}

# Set permissions
chmod +x "%s"

# Clean up
rm -f "%s.old" 2>/dev/null

echo "Successfully updated portable version"
exec "%s" --cleanup-after-update "%s.old" "%s"

rm -f "$0"
`,
			currentExe, updatePath,
			currentExe, currentExe+".old",
			updatePath, currentExe,
			currentExe+".old", currentExe,
			currentExe,
			currentExe+".old", currentExe)
	}

	scriptPath := filepath.Join(os.TempDir(), "theboys-update.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		u.logger.Error("Failed to create update script: %v", err)
		return ""
	}

	return scriptPath
}

// PerformUpdateCleanup cleans up after an update
func (u *Updater) PerformUpdateCleanup(oldExe, newExe string) error {
	u.logger.Info("Performing post-update cleanup...")

	// Remove old executable
	if err := os.Remove(oldExe); err != nil {
		u.logger.Warn("Failed to remove old executable: %v", err)
	}

	u.logger.Info("Update cleanup completed")
	return nil
}