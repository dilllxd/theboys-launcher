package platform

import (
	"fmt"
	"os"
	"runtime"
	"path/filepath"
	"syscall"
	"time"
)

// CommonPlatform provides cross-platform implementations
type CommonPlatform struct{}

// GetOS returns the operating system name
func (p *CommonPlatform) GetOS() string {
	return runtime.GOOS
}

// GetArch returns the system architecture
func (p *CommonPlatform) GetArch() string {
	return runtime.GOARCH
}

// GetExecutablePath returns the path to the current executable
func (p *CommonPlatform) GetExecutablePath() (string, error) {
	return os.Executable()
}

// GetAppDataDir returns the application data directory
func (p *CommonPlatform) GetAppDataDir() (string, error) {
	// For all platforms, use user's home directory for consistency
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".theboys-launcher"), nil
}

// FileExists checks if a file exists
func (p *CommonPlatform) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirectory creates a directory and all necessary parents
func (p *CommonPlatform) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// SupportsAutoUpdate returns whether the platform supports auto-updates
func (p *CommonPlatform) SupportsAutoUpdate() bool {
	return true
}

// GetDefaultRAM returns the default RAM allocation in MB
func (p *CommonPlatform) GetDefaultRAM() int {
	return 4096 // 4GB default
}

// GetMaxRAM returns the maximum safe RAM allocation in MB
func (p *CommonPlatform) GetMaxRAM() int {
	return 16384 // 16GB max
}

// GetAvailableDiskSpace returns available disk space for a path
func (p *CommonPlatform) GetAvailableDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	if err := syscall.Statfs(wd, &stat); err != nil {
		return 0, err
	}

	// Available blocks * block size
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

// GetCustomDataDir returns a custom data directory if set, falls back to default
func (p *CommonPlatform) GetCustomDataDir() (string, error) {
	// Check if custom directory is set in environment or config
	if customDir := os.Getenv("THEBOYS_DATA_DIR"); customDir != "" {
		return customDir, nil
	}

	// Fall back to default app data directory
	return p.GetAppDataDir()
}

// CanCreateShortcut returns whether the platform supports creating shortcuts
func (p *CommonPlatform) CanCreateShortcut() bool {
	return false // Base implementation, overridden by platform-specific implementations
}

// CreateShortcut creates a desktop shortcut (base implementation)
func (p *CommonPlatform) CreateShortcut(target, shortcutPath string) error {
	return fmt.Errorf("shortcut creation not implemented for this platform")
}

// IsInstalled returns whether the launcher is properly installed
func (p *CommonPlatform) IsInstalled() bool {
	// Check if there's an installation registry/config file
	installPath, err := p.GetInstallationPath()
	if err != nil {
		return false
	}
	return p.FileExists(installPath)
}

// GetInstallationPath returns the installation path
func (p *CommonPlatform) GetInstallationPath() (string, error) {
	// Default implementation: use executable path
	exePath, err := p.GetExecutablePath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// RegisterInstallation registers the installation (base implementation)
func (p *CommonPlatform) RegisterInstallation(path string) error {
	// Create an installation marker file
	installFile := filepath.Join(path, ".theboys-installed")
	file, err := os.Create(installFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write installation info
	fmt.Fprintf(file, "installed_at=%s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "version=%s\n", os.Getenv("THEBOYS_VERSION"))
	return nil
}

// UnregisterInstallation removes the installation registration
func (p *CommonPlatform) UnregisterInstallation() error {
	installPath, err := p.GetInstallationPath()
	if err != nil {
		return err
	}

	installFile := filepath.Join(installPath, ".theboys-installed")
	return os.Remove(installFile)
}