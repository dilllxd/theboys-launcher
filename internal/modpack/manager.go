// Package modpack provides modpack management functionality
package modpack

import (
	"archive/zip"
	"crypto/md5"
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

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
)

// Manager handles modpack operations
type Manager struct {
	repository     *Repository
	downloader     *DownloadManager
	configMgr      *ConfigManager
	config         *config.Config
	logger         *logging.Logger
}

// NewManager creates a new modpack manager
func NewManager(cfg *config.Config, logger *logging.Logger) *Manager {
	downloadFolder := filepath.Join(cfg.CacheDir, "downloads")

	return &Manager{
		repository: NewRepository(cfg, logger),
		downloader: NewDownloadManager(logger, 3, downloadFolder), // Max 3 concurrent downloads
		configMgr:  NewConfigManager(logger),
		config:     cfg,
		logger:     logger,
	}
}

// GetRepository returns the modpack repository
func (m *Manager) GetRepository() *Repository {
	return m.repository
}

// GetConfigManager returns the configuration manager
func (m *Manager) GetConfigManager() *ConfigManager {
	return m.configMgr
}

// GetAvailableModpacks returns all available modpacks
func (m *Manager) GetAvailableModpacks() ([]*Modpack, error) {
	return m.repository.GetAllModpacks()
}

// GetInstalledModpacks returns only installed modpacks
func (m *Manager) GetInstalledModpacks() ([]*Modpack, error) {
	allModpacks, err := m.GetAvailableModpacks()
	if err != nil {
		return nil, err
	}

	var installed []*Modpack
	for _, modpack := range allModpacks {
		if modpack.IsInstalled() {
			installed = append(installed, modpack)
		}
	}

	return installed, nil
}

// InstallModpack installs a modpack
func (m *Manager) InstallModpack(modpack *Modpack, progressCallback func(*InstallationProgress)) error {
	m.logger.Info("Starting installation of modpack: %s", modpack.Name)

	// Update status
	modpack.Status = StatusDownloading
	modpack.DateModified = time.Now()

	// Create instance directory
	instanceDir := filepath.Join(m.config.LauncherDir, "instances", modpack.ID)
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Download and extract modpack
	if err := m.downloadAndExtractModpack(modpack, instanceDir, progressCallback); err != nil {
		modpack.Status = StatusError
		return fmt.Errorf("failed to download/extract modpack: %w", err)
	}

	// Update modpack info
	modpack.Status = StatusInstalled
	modpack.InstallPath = instanceDir
	modpack.DateModified = time.Now()

	// Save modpack configuration
	if err := m.saveModpackConfig(modpack, instanceDir); err != nil {
		m.logger.Error("Failed to save modpack config: %v", err)
	}

	m.logger.Info("Successfully installed modpack: %s", modpack.Name)
	return nil
}

// UninstallModpack uninstalls a modpack
func (m *Manager) UninstallModpack(modpack *Modpack) error {
	m.logger.Info("Uninstalling modpack: %s", modpack.Name)

	if modpack.InstallPath == "" {
		return fmt.Errorf("modpack has no installation path")
	}

	// Remove instance directory
	if err := os.RemoveAll(modpack.InstallPath); err != nil {
		return fmt.Errorf("failed to remove instance directory: %w", err)
	}

	// Update status
	modpack.Status = StatusNotInstalled
	modpack.InstallPath = ""

	m.logger.Info("Successfully uninstalled modpack: %s", modpack.Name)
	return nil
}

// UpdateModpack updates an installed modpack
func (m *Manager) UpdateModpack(modpack *Modpack, progressCallback func(*InstallationProgress)) error {
	m.logger.Info("Updating modpack: %s", modpack.Name)

	if !modpack.IsInstalled() {
		return fmt.Errorf("modpack is not installed")
	}

	// For simplicity, we'll reinstall the modpack
	// In a real implementation, you'd want to do incremental updates
	modpack.Status = StatusDownloading

	if err := m.InstallModpack(modpack, progressCallback); err != nil {
		modpack.Status = StatusError
		return err
	}

	m.logger.Info("Successfully updated modpack: %s", modpack.Name)
	return nil
}

// LaunchModpack launches a modpack
func (m *Manager) LaunchModpack(modpack *Modpack) error {
	m.logger.Info("Launching modpack: %s", modpack.Name)

	if !modpack.IsInstalled() {
		return fmt.Errorf("modpack is not installed")
	}

	// Check Java installation
	javaPath := m.config.JavaPath
	if javaPath == "" {
		// Try to find Java automatically
		javaPath = m.findJavaInstallation()
		if javaPath == "" {
			return fmt.Errorf("Java not found. Please configure Java path in settings")
		}
	}

	// Prepare launch command
	args := []string{
		"-Xmx" + fmt.Sprintf("%dM", m.config.MaxMemory),
		"-Xms" + fmt.Sprintf("%dM", m.config.MinMemory),
	}

	// Add custom JVM args
	args = append(args, modpack.CustomJvmArgs...)

	// Add main class and classpath
	mcDir := filepath.Join(modpack.InstallPath, "minecraft")
	jarPath := filepath.Join(mcDir, "client.jar")

	// This is a simplified launcher - in reality, you'd need to:
	// 1. Set up the proper classpath with all libraries
	// 2. Handle authentication with Minecraft/Mojang servers
	// 3. Download missing assets/libraries
	// 4. Handle different mod loaders (Forge, Fabric, etc.)

	args = append(args, "-jar", jarPath)

	// Add custom launch args
	args = append(args, modpack.CustomLaunchArgs...)

	cmd := exec.Command(javaPath, args...)
	cmd.Dir = modpack.InstallPath

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch Minecraft: %w", err)
	}

	// Update last played time
	now := time.Now()
	modpack.LastPlayed = &now

	m.logger.Info("Successfully launched modpack: %s", modpack.Name)
	return nil
}

// SearchModpacks searches for modpacks
func (m *Manager) SearchModpacks(query string) ([]*Modpack, error) {
	allModpacks, err := m.GetAvailableModpacks()
	if err != nil {
		return nil, err
	}

	var results []*Modpack
	query = strings.ToLower(query)

	for _, modpack := range allModpacks {
		if strings.Contains(strings.ToLower(modpack.Name), query) ||
		   strings.Contains(strings.ToLower(modpack.Description), query) ||
		   strings.Contains(strings.ToLower(modpack.Author), query) {
			results = append(results, modpack)
		}
	}

	return results, nil
}

// GetModpackByID retrieves a modpack by ID
func (m *Manager) GetModpackByID(id string) (*Modpack, error) {
	allModpacks, err := m.GetAvailableModpacks()
	if err != nil {
		return nil, err
	}

	for _, modpack := range allModpacks {
		if modpack.ID == id {
			return modpack, nil
		}
	}

	return nil, fmt.Errorf("modpack not found: %s", id)
}

// ValidateModpack validates a modpack configuration
func (m *Manager) ValidateModpack(modpack *Modpack) error {
	if modpack.ID == "" {
		return fmt.Errorf("modpack ID is required")
	}

	if modpack.Name == "" {
		return fmt.Errorf("modpack name is required")
	}

	if modpack.MinecraftVersion.ID == "" {
		return fmt.Errorf("Minecraft version is required")
	}

	if modpack.ModLoader.Type == "" {
		return fmt.Errorf("mod loader type is required")
	}

	if modpack.Type == "" {
		return fmt.Errorf("modpack type is required")
	}

	return nil
}

// downloadAndExtractModpack downloads and extracts a modpack
func (m *Manager) downloadAndExtractModpack(modpack *Modpack, instanceDir string, progressCallback func(*InstallationProgress)) error {
	if modpack.DownloadURL == "" {
		return fmt.Errorf("modpack has no download URL")
	}

	// Download modpack using download manager
	if err := m.downloader.DownloadModpack(modpack, progressCallback); err != nil {
		return fmt.Errorf("failed to download modpack: %w", err)
	}

	// Get downloaded file path
	tempDir := filepath.Join(m.config.CacheDir, "downloads", modpack.ID)
	zipPath := filepath.Join(tempDir, "modpack.zip")

	// Update progress for extraction stage
	if progressCallback != nil {
		progressCallback(&InstallationProgress{
			ModpackID:      modpack.ID,
			Stage:          "extracting",
			Progress:       0.0,
			CurrentFile:    "modpack.zip",
			TotalFiles:     1,
			CompletedFiles: 0,
		})
	}

	// Extract archive
	if err := m.extractArchive(zipPath, instanceDir, progressCallback, modpack.ID); err != nil {
		return fmt.Errorf("failed to extract modpack: %w", err)
	}

	// Clean up temporary files
	if err := os.RemoveAll(tempDir); err != nil {
		m.logger.Error("Failed to clean up temporary files: %v", err)
	}

	return nil
}

// downloadFile downloads a file with progress tracking
func (m *Manager) downloadFile(url string, dest *os.File, progress func(int64, int64, int64)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	total := resp.ContentLength
	var written int64

	startTime := time.Now()
	lastUpdate := startTime

	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		nr, err := resp.Body.Read(buf)
		if nr > 0 {
			nw, err := dest.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if err != nil {
				return err
			}

			// Calculate speed and update progress
			now := time.Now()
			if now.Sub(lastUpdate) >= time.Second {
				elapsed := now.Sub(startTime).Seconds()
				var speed int64
				if elapsed > 0 {
					speed = int64(float64(written) / elapsed)
				}

				if progress != nil {
					progress(written, total, speed)
				}
				lastUpdate = now
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}

// extractArchive extracts a zip archive
func (m *Manager) extractArchive(src, dest string, progressCallback func(*InstallationProgress), modpackID string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	var totalFiles int
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			totalFiles++
		}
	}

	var extractedFiles int

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		// Ensure path is within destination directory (security check)
		if !strings.HasPrefix(path, dest+string(filepath.Separator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}

		// Create directory for file
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return err
		}

		if _, err := io.Copy(destFile, fileReader); err != nil {
			fileReader.Close()
			destFile.Close()
			return err
		}

		fileReader.Close()
		destFile.Close()

		extractedFiles++

		// Update progress
		if progressCallback != nil {
			progress := float64(extractedFiles) / float64(totalFiles)
			progressCallback(&InstallationProgress{
				ModpackID:      modpackID,
				Stage:          "extracting",
				Progress:       progress,
				CurrentFile:    file.Name,
				TotalFiles:     totalFiles,
				CompletedFiles: extractedFiles,
			})
		}
	}

	return nil
}

// saveModpackConfig saves the modpack configuration to disk
func (m *Manager) saveModpackConfig(modpack *Modpack, instanceDir string) error {
	configPath := filepath.Join(instanceDir, "modpack.json")

	data, err := json.MarshalIndent(modpack, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// findJavaInstallation tries to find Java installation
func (m *Manager) findJavaInstallation() string {
	// Common Java installation paths
	var javaPaths []string

	switch runtime.GOOS {
	case "windows":
		javaPaths = []string{
			"C:\\Program Files\\Java\\jdk-21\\bin\\java.exe",
			"C:\\Program Files\\Java\\jdk-17\\bin\\java.exe",
			"C:\\Program Files\\Eclipse Adoptium\\jdk-21.0.2.13-hotspot\\bin\\java.exe",
			"C:\\Program Files\\Eclipse Adoptium\\jdk-17.0.9.9-hotspot\\bin\\java.exe",
		}
	case "darwin":
		javaPaths = []string{
			"/usr/local/opt/openjdk/bin/java",
			"/usr/local/opt/openjdk@11/bin/java",
			"/Library/Java/JavaVirtualMachines/openjdk-21.jdk/Contents/Home/bin/java",
			"/Library/Java/JavaVirtualMachines/openjdk-17.jdk/Contents/Home/bin/java",
			"/System/Library/Frameworks/JavaVM.framework/Versions/Current/Commands/java",
		}
	case "linux":
		javaPaths = []string{
			"/usr/bin/java",
			"/usr/local/bin/java",
			"/opt/java/bin/java",
			"/usr/lib/jvm/java-21-openjdk/bin/java",
			"/usr/lib/jvm/java-17-openjdk/bin/java",
		}
	}

	// Check if Java exists in PATH
	if path, err := exec.LookPath("java"); err == nil {
		return path
	}

	// Check common installation paths
	for _, path := range javaPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetModpackIcon retrieves the icon for a modpack
func (m *Manager) GetModpackIcon(modpack *Modpack) (string, error) {
	// Check if icon exists locally
	if modpack.InstallPath != "" {
		localIcon := filepath.Join(modpack.InstallPath, "icon.png")
		if _, err := os.Stat(localIcon); err == nil {
			return localIcon, nil
		}
	}

	// Download icon if URL is available
	if modpack.IconURL != "" {
		iconPath := filepath.Join(m.config.CacheDir, "icons", modpack.ID+".png")

		// Create icons directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(iconPath), 0755); err != nil {
			return "", err
		}

		// Check if already cached
		if _, err := os.Stat(iconPath); err == nil {
			return iconPath, nil
		}

		// Download icon
		resp, err := http.Get(modpack.IconURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to download icon: %s", resp.Status)
		}

		file, err := os.Create(iconPath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		if _, err := io.Copy(file, resp.Body); err != nil {
			os.Remove(iconPath)
			return "", err
		}

		return iconPath, nil
	}

	return "", fmt.Errorf("no icon available")
}

// CalculateModpackHash calculates MD5 hash of a modpack file
func (m *Manager) CalculateModpackHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}