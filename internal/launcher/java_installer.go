package launcher

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// JavaInstaller handles Java installation and extraction
type JavaInstaller struct {
	platform platform.Platform
	logger   logging.Logger
}

// NewJavaInstaller creates a new Java installer
func NewJavaInstaller(platform platform.Platform, logger logging.Logger) *JavaInstaller {
	return &JavaInstaller{
		platform: platform,
		logger:   logger,
	}
}

// InstallJavaFromZip extracts Java from a zip file to the installation directory
func (ji *JavaInstaller) InstallJavaFromZip(zipPath, installDir string) error {
	ji.logger.Info("Installing Java from %s to %s", zipPath, installDir)

	// Open the zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open Java zip file: %w", err)
	}
	defer reader.Close()

	// Find the root directory in the zip (usually something like "jdk-17.0.2+8")
	var rootDir string
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			parts := strings.Split(file.Name, "/")
			if len(parts) == 1 && parts[0] != "" {
				rootDir = parts[0]
				break
			}
		}
	}

	// Extract files
	for _, file := range reader.File {
		// Skip macOS metadata files
		if strings.HasPrefix(file.Name, "__MACOSX/") {
			continue
		}

		// Determine destination path
		var destPath string
		if rootDir != "" && strings.HasPrefix(file.Name, rootDir+"/") {
			// Remove root directory prefix
			relativePath := strings.TrimPrefix(file.Name, rootDir+"/")
			destPath = filepath.Join(installDir, relativePath)
		} else {
			// No root directory or file is not in root
			destPath = filepath.Join(installDir, file.Name)
		}

		// Create directory for file
		if file.FileInfo().IsDir() {
			if err := ji.platform.CreateDirectory(destPath); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			continue
		}

		// Extract file
		if err := ji.extractFile(file, destPath); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	// Verify installation
	javaExe := ji.getJavaExecutable(installDir)
	if !ji.platform.FileExists(javaExe) {
		return fmt.Errorf("Java installation verification failed: %s not found", javaExe)
	}

	// Set executable permissions on Unix-like systems
	if ji.platform.GetOS() != "windows" {
		if err := os.Chmod(javaExe, 0755); err != nil {
			ji.logger.Warn("Failed to set executable permissions on %s: %v", javaExe, err)
		}
	}

	ji.logger.Info("Java installed successfully to %s", installDir)
	return nil
}

// extractFile extracts a single file from the zip archive
func (ji *JavaInstaller) extractFile(file *zip.File, destPath string) error {
	// Create directory for file
	if err := ji.platform.CreateDirectory(filepath.Dir(destPath)); err != nil {
		return err
	}

	// Open file in zip
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, rc)
	return err
}

// getJavaExecutable returns the path to the Java executable in the installation directory
func (ji *JavaInstaller) getJavaExecutable(installDir string) string {
	os := ji.platform.GetOS()
	switch os {
	case "windows":
		return filepath.Join(installDir, "bin", "java.exe")
	case "darwin", "linux":
		return filepath.Join(installDir, "bin", "java")
	default:
		return filepath.Join(installDir, "bin", "java")
	}
}

// GetJavaHome returns the JAVA_HOME directory for a Java installation
func (ji *JavaInstaller) GetJavaHome(installDir string) string {
	// For Windows, JAVA_HOME should be the installation directory
	// For Unix-like systems, it should be the parent of bin directory
	if ji.platform.GetOS() == "windows" {
		return installDir
	}

	// Check if installDir already contains bin/java
	javaExe := ji.getJavaExecutable(installDir)
	if ji.platform.FileExists(javaExe) {
		return installDir
	}

	// Otherwise, assume installDir is inside JAVA_HOME
	return filepath.Dir(installDir)
}

// VerifyJavaInstallation verifies that a Java installation is working
func (ji *JavaInstaller) VerifyJavaInstallation(installDir string) error {
	javaExe := ji.getJavaExecutable(installDir)

	if !ji.platform.FileExists(javaExe) {
		return fmt.Errorf("Java executable not found: %s", javaExe)
	}

	// Try to run java -version
	// This verification should be done by the JavaManager instead
	ji.logger.Debug("Java installation structure verified: %s", javaExe)
	return nil
}

// CleanupInstallation removes a Java installation
func (ji *JavaInstaller) CleanupInstallation(installDir string) error {
	ji.logger.Info("Cleaning up Java installation: %s", installDir)

	if !ji.platform.FileExists(installDir) {
		ji.logger.Debug("Installation directory does not exist: %s", installDir)
		return nil
	}

	// Remove the entire directory
	return os.RemoveAll(installDir)
}

// GetInstallationSize returns the size of a Java installation in bytes
func (ji *JavaInstaller) GetInstallationSize(installDir string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
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

// IsInstallationComplete checks if a Java installation appears to be complete
func (ji *JavaInstaller) IsInstallationComplete(installDir string) bool {
	javaExe := ji.getJavaExecutable(installDir)
	if !ji.platform.FileExists(javaExe) {
		return false
	}

	// Check for key directories/files
	binDir := filepath.Join(installDir, "bin")
	if !ji.platform.FileExists(binDir) {
		return false
	}

	// Check for some common Java components
	keyFiles := []string{
		filepath.Join(binDir, "javac" + ji.getExecutableSuffix()),
		filepath.Join(installDir, "lib"),
	}

	for _, file := range keyFiles {
		if !ji.platform.FileExists(file) {
			ji.logger.Debug("Missing Java component: %s", file)
			return false
		}
	}

	return true
}

// getExecutableSuffix returns the executable suffix for the current platform
func (ji *JavaInstaller) getExecutableSuffix() string {
	if ji.platform.GetOS() == "windows" {
		return ".exe"
	}
	return ""
}