package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/platform"
)

// MigrationManager handles migration from portable to user directory mode
type MigrationManager struct {
	platform platform.Platform
	logger   logging.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(platform platform.Platform, logger logging.Logger) *MigrationManager {
	return &MigrationManager{
		platform: platform,
		logger:   logger,
	}
}

// PortableInfo represents information about a portable installation
type PortableInfo struct {
	HasInstances      bool     `json:"has_instances"`
	HasConfig         bool     `json:"has_config"`
	HasPrism          bool     `json:"has_prism"`
	HasUtil           bool     `json:"has_util"`
	HasLogs           bool     `json:"has_logs"`
	InstanceCount     int      `json:"instance_count"`
	Instances         []string `json:"instances"`
	ConfigFiles       []string `json:"config_files"`
	TotalSize         int64    `json:"total_size"`
	LastModified      string   `json:"last_modified"`
	PortableVersion   string   `json:"portable_version"`
	MigrationRequired bool     `json:"migration_required"`
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Success      bool     `json:"success"`
	MigratedDirs []string `json:"migrated_dirs"`
	MigratedFiles int     `json:"migrated_files"`
	SkippedFiles  int     `json:"skipped_files"`
	Errors        []string `json:"errors"`
	Duration      string   `json:"duration"`
	BackupPath    string   `json:"backup_path"`
}

// DetectPortableInstallation detects if running from a portable installation
func (mm *MigrationManager) DetectPortableInstallation() (*PortableInfo, error) {
	exePath, err := mm.platform.GetExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	info := &PortableInfo{
		MigrationRequired: false,
	}

	// Check for portable indicators
	portableIndicators := []string{
		"prism",
		"util",
		"instances",
		"config",
		"logs",
	}

	var totalSize int64
	var lastModified time.Time

	for _, indicator := range portableIndicators {
		indicatorPath := filepath.Join(exeDir, indicator)
		if mm.platform.FileExists(indicatorPath) {
			switch indicator {
			case "instances":
				info.HasInstances = true
				instances, err := mm.detectInstances(indicatorPath)
				if err == nil {
					info.Instances = instances
					info.InstanceCount = len(instances)
				}
			case "config":
				info.HasConfig = true
				configFiles, err := mm.detectConfigFiles(indicatorPath)
				if err == nil {
					info.ConfigFiles = configFiles
				}
			case "prism":
				info.HasPrism = true
			case "util":
				info.HasUtil = true
			case "logs":
				info.HasLogs = true
			}

			// Calculate directory size
			size, err := mm.calculateDirSize(indicatorPath)
			if err == nil {
				totalSize += size
			}

			// Get last modified time
			if modTime, err := mm.getLastModified(indicatorPath); err == nil {
				if modTime.After(lastModified) {
					lastModified = modTime
				}
			}
		}
	}

	info.TotalSize = totalSize
	info.LastModified = lastModified.Format(time.RFC3339)

	// Check if migration is required
	// Only require migration if we have significant data and we're on Windows
	if mm.platform.GetOS() == "windows" &&
		(info.HasInstances || info.HasConfig || info.HasPrism) {
		info.MigrationRequired = true

		// Try to detect legacy version
		if legacyVersion := mm.detectLegacyVersion(exeDir); legacyVersion != "" {
			info.PortableVersion = legacyVersion
		} else {
			info.PortableVersion = "unknown"
		}
	}

	return info, nil
}

// MigratePortableInstallation migrates data from portable to user directory
func (mm *MigrationManager) MigratePortableInstallation() (*MigrationResult, error) {
	startTime := time.Now()

	result := &MigrationResult{
		MigratedDirs: []string{},
		Errors:       []string{},
	}

	// Get paths
	exePath, err := mm.platform.GetExecutablePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	portableDir := filepath.Dir(exePath)
	userDataDir, err := mm.platform.GetAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user data directory: %w", err)
	}

	// Create backup
	backupPath, err := mm.createBackup(portableDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create backup: %v", err))
	} else {
		result.BackupPath = backupPath
	}

	mm.logger.Info("Starting migration from portable mode to user directory")
	mm.logger.Info("Source: %s", portableDir)
	mm.logger.Info("Destination: %s", userDataDir)

	// Create user data directory
	if err := mm.platform.CreateDirectory(userDataDir); err != nil {
		return nil, fmt.Errorf("failed to create user data directory: %w", err)
	}

	// Directories to migrate
	dirsToMigrate := []string{
		"instances",
		"config",
		"prism",
		"util",
		"logs",
	}

	// Migrate each directory
	for _, dir := range dirsToMigrate {
		sourceDir := filepath.Join(portableDir, dir)
		destDir := filepath.Join(userDataDir, dir)

		if mm.platform.FileExists(sourceDir) {
			migratedFiles, err := mm.migrateDirectory(sourceDir, destDir)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to migrate %s: %v", dir, err))
				continue
			}

			result.MigratedDirs = append(result.MigratedDirs, dir)
			result.MigratedFiles += migratedFiles

			mm.logger.Info("Migrated directory: %s (%d files)", dir, migratedFiles)
		}
	}

	// Migrate specific files
	filesToMigrate := []string{
		"modpacks.json",
		"settings.json",
	}

	for _, file := range filesToMigrate {
		sourceFile := filepath.Join(portableDir, file)
		destFile := filepath.Join(userDataDir, "config", file)

		if mm.platform.FileExists(sourceFile) {
			if err := mm.migrateFile(sourceFile, destFile); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to migrate %s: %v", file, err))
			} else {
				result.MigratedFiles++
				mm.logger.Info("Migrated file: %s", file)
			}
		}
	}

	// Create migration marker
	migrationMarker := filepath.Join(userDataDir, ".migration-completed")
	mm.createMigrationMarker(migrationMarker, portableDir, result)

	// Calculate duration
	duration := time.Since(startTime)
	result.Duration = duration.String()

	// Determine success
	result.Success = len(result.Errors) == 0 && len(result.MigratedDirs) > 0

	if result.Success {
		mm.logger.Info("Migration completed successfully in %s", duration.String())
		mm.logger.Info("Migrated %d directories and %d files", len(result.MigratedDirs), result.MigratedFiles)
		if result.BackupPath != "" {
			mm.logger.Info("Backup created at: %s", result.BackupPath)
		}
	} else {
		mm.logger.Error("Migration completed with %d errors", len(result.Errors))
		for _, err := range result.Errors {
			mm.logger.Error("Migration error: %s", err)
		}
	}

	return result, nil
}

// ShouldPromptMigration checks if we should prompt the user for migration
func (mm *MigrationManager) ShouldPromptMigration() (bool, *PortableInfo) {
	info, err := mm.DetectPortableInstallation()
	if err != nil {
		mm.logger.Warn("Failed to detect portable installation: %v", err)
		return false, nil
	}

	// Don't prompt if already migrated
	userDataDir, err := mm.platform.GetAppDataDir()
	if err != nil {
		return false, info
	}

	migrationMarker := filepath.Join(userDataDir, ".migration-completed")
	if mm.platform.FileExists(migrationMarker) {
		return false, info
	}

	return info.MigrationRequired, info
}

// Helper methods

func (mm *MigrationManager) detectInstances(instancesDir string) ([]string, error) {
	var instances []string

	entries, err := os.ReadDir(instancesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			instances = append(instances, entry.Name())
		}
	}

	return instances, nil
}

func (mm *MigrationManager) detectConfigFiles(configDir string) ([]string, error) {
	var configFiles []string

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			configFiles = append(configFiles, entry.Name())
		}
	}

	return configFiles, nil
}

func (mm *MigrationManager) calculateDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func (mm *MigrationManager) getLastModified(path string) (time.Time, error) {
	var lastMod time.Time

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(lastMod) {
			lastMod = info.ModTime()
		}
		return nil
	})

	return lastMod, err
}

func (mm *MigrationManager) detectLegacyVersion(exeDir string) string {
	// Check for legacy version file
	versionFile := filepath.Join(exeDir, "version.txt")
	if mm.platform.FileExists(versionFile) {
		data, err := os.ReadFile(versionFile)
		if err == nil {
			return string(data)
		}
	}

	// Check legacy settings file for version
	settingsFile := filepath.Join(exeDir, "settings.json")
	if mm.platform.FileExists(settingsFile) {
		var settings map[string]interface{}
		data, err := os.ReadFile(settingsFile)
		if err == nil {
			if json.Unmarshal(data, &settings) == nil {
				if version, ok := settings["version"].(string); ok {
					return version
				}
			}
		}
	}

	return ""
}

func (mm *MigrationManager) createBackup(portableDir string) (string, error) {
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupName := fmt.Sprintf("portable-backup-%s", timestamp)

	// Create backup in user's temp directory
	tempDir := os.TempDir()
	backupDir := filepath.Join(tempDir, backupName)

	mm.logger.Info("Creating backup at: %s", backupDir)

	// Copy portable directory to backup location
	if _, err := mm.copyDirectory(portableDir, backupDir); err != nil {
		return "", err
	}

	return backupDir, nil
}

func (mm *MigrationManager) migrateDirectory(sourceDir, destDir string) (int, error) {
	if err := mm.platform.CreateDirectory(destDir); err != nil {
		return 0, err
	}

	return mm.copyDirectory(sourceDir, destDir)
}

func (mm *MigrationManager) migrateFile(sourceFile, destFile string) error {
	// Ensure destination directory exists
	destDir := filepath.Dir(destFile)
	if err := mm.platform.CreateDirectory(destDir); err != nil {
		return err
	}

	// Copy file
	return mm.copyFile(sourceFile, destFile)
}

func (mm *MigrationManager) copyDirectory(sourceDir, destDir string) (int, error) {
	fileCount := 0

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return mm.platform.CreateDirectory(destPath)
		}

		if err := mm.copyFile(path, destPath); err != nil {
			return err
		}

		fileCount++
		return nil
	})

	return fileCount, err
}

func (mm *MigrationManager) copyFile(sourceFile, destFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

func (mm *MigrationManager) createMigrationMarker(markerPath, portableDir string, result *MigrationResult) {
	markerData := map[string]interface{}{
		"migration_date":    time.Now().Format(time.RFC3339),
		"portable_dir":      portableDir,
		"migrated_dirs":     result.MigratedDirs,
		"migrated_files":    result.MigratedFiles,
		"backup_path":       result.BackupPath,
		"migration_version": "1.0.0",
	}

	data, err := json.MarshalIndent(markerData, "", "  ")
	if err != nil {
		mm.logger.Warn("Failed to create migration marker: %v", err)
		return
	}

	if err := os.WriteFile(markerPath, data, 0644); err != nil {
		mm.logger.Warn("Failed to write migration marker: %v", err)
	}
}