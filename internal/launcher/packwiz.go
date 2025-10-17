package launcher

import (
	"bytes"
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

	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/platform"
)

// PackwizManager handles packwiz bootstrap and mod installation
type PackwizManager struct {
	platform platform.Platform
	logger   logging.Logger
}

// NewPackwizManager creates a new packwiz manager
func NewPackwizManager(platform platform.Platform, logger logging.Logger) *PackwizManager {
	return &PackwizManager{
		platform: platform,
		logger:   logger,
	}
}

// PackwizBootstrapInfo contains information about packwiz bootstrap files
type PackwizBootstrapInfo struct {
	ExePath string
	JarPath string
	Version string
}

// Packwiz-specific types

// EnsurePackwizBootstrap ensures packwiz bootstrap is available
func (pm *PackwizManager) EnsurePackwizBootstrap(utilDir string) (*PackwizBootstrapInfo, error) {
	pm.logger.Info("Ensuring packwiz bootstrap...")

	bootstrapExe := filepath.Join(utilDir, "packwiz-installer-bootstrap.exe")
	bootstrapJar := filepath.Join(utilDir, "packwiz-installer-bootstrap.jar")

	// Check if bootstrap already exists
	if pm.platform.FileExists(bootstrapExe) {
		pm.logger.Debug("Packwiz bootstrap EXE already exists")
		return &PackwizBootstrapInfo{
			ExePath: bootstrapExe,
			JarPath: bootstrapJar,
			Version: "unknown",
		}, nil
	}

	if pm.platform.FileExists(bootstrapJar) {
		pm.logger.Debug("Packwiz bootstrap JAR already exists")
		return &PackwizBootstrapInfo{
			ExePath: bootstrapExe,
			JarPath: bootstrapJar,
			Version: "unknown",
		}, nil
	}

	// Download bootstrap
	return pm.downloadPackwizBootstrap(utilDir)
}

// downloadPackwizBootstrap downloads the packwiz bootstrap
func (pm *PackwizManager) downloadPackwizBootstrap(utilDir string) (*PackwizBootstrapInfo, error) {
	pm.logger.Info("Downloading packwiz bootstrap...")

	// Fetch latest release from GitHub
	release, err := pm.fetchLatestRelease("packwiz/packwiz-installer-bootstrap")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packwiz release: %w", err)
	}

	if len(release.Assets) == 0 {
		return nil, fmt.Errorf("no assets found in packwiz release")
	}

	// Find appropriate asset for current platform
	var targetAsset *GitHubReleaseAsset

	for _, asset := range release.Assets {
		if runtime.GOOS == "windows" && strings.HasSuffix(asset.Name, ".exe") {
			targetAsset = &asset
			break
		} else if strings.HasSuffix(asset.Name, ".jar") {
			targetAsset = &asset
			break
		}
	}

	if targetAsset == nil {
		return nil, fmt.Errorf("no suitable packwiz bootstrap asset found for platform")
	}

	// Download the asset
	resp, err := http.Get(targetAsset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download packwiz bootstrap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("packwiz bootstrap download failed: HTTP %d", resp.StatusCode)
	}

	// Determine file path
	var targetPath string
	if strings.HasSuffix(targetAsset.Name, ".exe") {
		targetPath = filepath.Join(utilDir, "packwiz-installer-bootstrap.exe")
	} else {
		targetPath = filepath.Join(utilDir, "packwiz-installer-bootstrap.jar")
	}

	// Save file
	file, err := os.Create(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create packwiz bootstrap file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save packwiz bootstrap: %w", err)
	}

	pm.logger.Info("Packwiz bootstrap installed: %s", targetAsset.Name)

	return &PackwizBootstrapInfo{
		ExePath: filepath.Join(utilDir, "packwiz-installer-bootstrap.exe"),
		JarPath: filepath.Join(utilDir, "packwiz-installer-bootstrap.jar"),
		Version: release.TagName,
	}, nil
}

// fetchLatestRelease fetches the latest release from a GitHub repository
func (pm *PackwizManager) fetchLatestRelease(repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

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

// InstallModpack installs a modpack using packwiz
func (pm *PackwizManager) InstallModpack(packURL, instanceDir, javaPath string, progressCallback ProgressCallback) error {
	pm.logger.Info("Installing modpack with packwiz: %s", packURL)

	// Get packwiz bootstrap
	utilDir := filepath.Join(instanceDir, "..", "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		return fmt.Errorf("failed to create util directory: %w", err)
	}

	bootstrap, err := pm.EnsurePackwizBootstrap(utilDir)
	if err != nil {
		return fmt.Errorf("failed to ensure packwiz bootstrap: %w", err)
	}

	// Determine command to run
	var cmd *exec.Cmd
	if _, err := os.Stat(bootstrap.ExePath); err == nil {
		cmd = exec.Command(bootstrap.ExePath, "-g", packURL)
	} else if _, err := os.Stat(bootstrap.JarPath); err == nil {
		cmd = exec.Command(javaPath, "-jar", bootstrap.JarPath, "-g", packURL)
	} else {
		return fmt.Errorf("packwiz bootstrap not found after download")
	}

	// Set working directory to minecraft directory
	cmd.Dir = instanceDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start progress tracking
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		progress := 0.0
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				progress += 0.05
				if progress > 0.9 {
					progress = 0.9
				}
				if progressCallback != nil {
					progressCallback(progress)
				}
			}
		}
	}()

	// Run packwiz
	err = cmd.Run()
	close(progressDone)

	// Set final progress
	if progressCallback != nil {
		progressCallback(1.0)
	}

	if err != nil {
		pm.logger.Error("Packwiz failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		return fmt.Errorf("packwiz installation failed: %w", err)
	}

	pm.logger.Info("Packwiz installation completed successfully")
	return nil
}

// ParsePackInfo parses pack.toml from a URL
func (pm *PackwizManager) ParsePackInfo(packURL string) (*PackInfo, error) {
	pm.logger.Debug("Parsing pack info from: %s", packURL)

	// Download pack.toml
	resp, err := http.Get(packURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download pack.toml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pack.toml download failed: HTTP %d", resp.StatusCode)
	}

	// For now, we'll do a simple parse - in a real implementation,
	// you'd use a proper TOML parser
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read pack.toml: %w", err)
	}

	// Simple parsing - in production, use github.com/BurntSushi/toml
	lines := strings.Split(string(content), "\n")
	packInfo := &PackInfo{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name = ") {
			packInfo.Name = strings.Trim(strings.TrimPrefix(line, "name = "), `"`)
		} else if strings.HasPrefix(line, "version = ") {
			packInfo.Version = strings.Trim(strings.TrimPrefix(line, "version = "), `"`)
		}
	}

	return packInfo, nil
}

// GetLWJGLVersionForMinecraft fetches LWJGL version from PrismLauncher meta-launcher
func (pm *PackwizManager) GetLWJGLVersionForMinecraft(mcVersion string) (*LWJGLInfo, error) {
	pm.logger.Debug("Fetching LWJGL version for Minecraft %s", mcVersion)

	cleanVersion := strings.TrimSpace(mcVersion)
	if cleanVersion == "" {
		return &LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}, nil
	}

	// Construct GitHub URL for PrismLauncher meta-launcher data
	url := fmt.Sprintf("https://raw.githubusercontent.com/PrismLauncher/meta-launcher/refs/heads/master/net.minecraft/%s.json", cleanVersion)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		pm.logger.Warn("Failed to fetch LWJGL data for Minecraft %s: %v", cleanVersion, err)
		return &LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		pm.logger.Warn("LWJGL data not found for Minecraft %s (HTTP %d)", cleanVersion, resp.StatusCode)
		return &LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}, nil
	}

	// Parse the JSON response
	var data struct {
		Requires []struct {
			Suggests string `json:"suggests"`
			UID      string `json:"uid"`
		} `json:"requires"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		pm.logger.Warn("Failed to parse LWJGL data for Minecraft %s: %v", cleanVersion, err)
		return &LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}, nil
	}

	// Look for LWJGL requirement
	for _, req := range data.Requires {
		if req.UID == "org.lwjgl" || req.UID == "org.lwjgl3" {
			if req.Suggests != "" {
				var name string
				if req.UID == "org.lwjgl" {
					name = "LWJGL 2"
				} else {
					name = "LWJGL 3"
				}
				return &LWJGLInfo{
					Version: req.Suggests,
					UID:     req.UID,
					Name:    name,
				}, nil
			}
		}
	}

	// Default fallback
	return &LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}, nil
}

// CreateModpackBackup creates a comprehensive backup of the current modpack installation
func (pm *PackwizManager) CreateModpackBackup(instanceDir, modpackID string) (string, error) {
	pm.logger.Info("Creating comprehensive backup for %s", modpackID)

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-backup-%s", modpackID, timestamp)
	backupPath := filepath.Join(instanceDir, "..", "backups", backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Calculate total backup size for progress tracking
	var totalSize int64
	dirsToBackup := []string{"mods", "config", "resourcepacks", "shaderpacks", "scripts", "options.txt"}
	for _, dir := range dirsToBackup {
		if size, err := pm.getDirectorySize(filepath.Join(instanceDir, dir)); err == nil {
			totalSize += size
		}
	}

	var backedUpSize int64 = 0
	totalSizePtr := totalSize
	// Backup important directories and files
	filesToBackup := []string{
		"pack.toml",
		"mods",
		"config",
		"resourcepacks",
		"shaderpacks",
		"scripts",
		"options.txt",
		"servers.dat",
		"realmstoken.dat",
	}

	for _, file := range filesToBackup {
		src := filepath.Join(instanceDir, file)
		dst := filepath.Join(backupPath, file)

		if info, err := os.Stat(src); err == nil {
			if info.IsDir() {
				// Backup directory with progress
				if err := pm.copyDirWithProgress(src, dst, &backedUpSize, totalSizePtr); err != nil {
					pm.logger.Warn("Failed to backup directory %s: %v", file, err)
				}
			} else {
				// Backup single file
				if err := copyFile(src, dst); err != nil {
					pm.logger.Warn("Failed to backup file %s: %v", file, err)
				} else {
					backedUpSize += info.Size()
				}
			}
		}
	}

	// Create comprehensive backup metadata
	metadata := map[string]interface{}{
		"timestamp":       timestamp,
		"modpack_id":      modpackID,
		"instance_path":   instanceDir,
		"backup_path":     backupPath,
		"created_at":      time.Now().Format(time.RFC3339),
		"total_size":      totalSize,
		"backed_up_size":  backedUpSize,
		"files_backed_up": len(filesToBackup),
		"minecraft_version": pm.detectMinecraftVersion(instanceDir),
		"modloader": pm.detectModloader(instanceDir),
		"backup_format": "comprehensive_v1",
	}

	metadataFile := filepath.Join(backupPath, "backup.json")
	metadataBytes, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(metadataFile, metadataBytes, 0644); err != nil {
		pm.logger.Warn("Failed to write backup metadata: %v", err)
	}

	pm.logger.Info("Comprehensive backup created: %s (%s)", backupPath, formatBytes(backedUpSize))
	return backupPath, nil
}

// RestoreModpackBackup restores a modpack from backup
func (pm *PackwizManager) RestoreModpackBackup(instanceDir, backupPath string) error {
	pm.logger.Info("Restoring modpack from backup: %s", backupPath)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup path does not exist: %s", backupPath)
	}

	// Restore files from backup
	filesToRestore := []string{
		"mods",
		"config",
		"resourcepacks",
		"shaderpacks",
	}

	for _, file := range filesToRestore {
		src := filepath.Join(backupPath, file)
		dst := filepath.Join(instanceDir, file)

		if _, err := os.Stat(src); err == nil {
			// Remove existing directory
			if _, err := os.Stat(dst); err == nil {
				os.RemoveAll(dst)
			}

			if err := copyDir(src, dst); err != nil {
				pm.logger.Warn("Failed to restore %s: %v", file, err)
			}
		}
	}

	pm.logger.Info("Modpack restored from backup")
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// getDirectorySize calculates the total size of a directory recursively
func (pm *PackwizManager) getDirectorySize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// copyDirWithProgress copies a directory with progress tracking
func (pm *PackwizManager) copyDirWithProgress(src, dst string, currentSize *int64, totalSize int64) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := pm.copyDirWithProgress(srcPath, dstPath, currentSize, totalSize); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
			// Update progress
			if info, err := entry.Info(); err == nil {
				*currentSize += info.Size()
				if totalSize > 0 {
					progress := float64(*currentSize) / float64(totalSize)
					pm.logger.Debug("Backup progress: %.1f%%", progress*100)
				}
			}
		}
	}

	return nil
}

// detectMinecraftVersion attempts to detect Minecraft version from instance
func (pm *PackwizManager) detectMinecraftVersion(instancePath string) string {
	// Try to read from pack.toml first
	packFile := filepath.Join(instancePath, "pack.toml")
	if content, err := os.ReadFile(packFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "minecraft = ") {
				version := strings.Trim(strings.TrimPrefix(line, "minecraft = "), `"`)
				if version != "" {
					return version
				}
			}
		}
	}

	// Try to read from PrismLauncher instance config
	instanceConfig := filepath.Join(instancePath, "..", "instance.cfg")
	if content, err := os.ReadFile(instanceConfig); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "MinecraftVersion=") {
				version := strings.TrimPrefix(line, "MinecraftVersion=")
				return version
			}
		}
	}

	return "unknown"
}

// detectModloader attempts to detect the modloader type from instance
func (pm *PackwizManager) detectModloader(instancePath string) string {
	// Check for Forge markers
	if _, err := os.Stat(filepath.Join(instancePath, "forge.toml")); err == nil {
		return "forge"
	}

	// Check for Fabric markers
	if _, err := os.Stat(filepath.Join(instancePath, "fabric-loader.json")); err == nil {
		return "fabric"
	}

	// Check for Quilt markers
	if _, err := os.Stat(filepath.Join(instancePath, "quilt-loader.json")); err == nil {
		return "quilt"
	}

	// Check for NeoForge markers
	if _, err := os.Stat(filepath.Join(instancePath, "neoforge.mods.toml")); err == nil {
		return "neoforge"
	}

	// Try to detect from pack.toml
	packFile := filepath.Join(instancePath, "pack.toml")
	if content, err := os.ReadFile(packFile); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "fabric") {
			return "fabric"
		} else if strings.Contains(contentStr, "forge") {
			return "forge"
		} else if strings.Contains(contentStr, "quilt") {
			return "quilt"
		} else if strings.Contains(contentStr, "neoforge") {
			return "neoforge"
		}
	}

	return "unknown"
}

