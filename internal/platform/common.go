package platform

import (
	"os"
	"runtime"
	"path/filepath"
	"syscall"
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
	if p.GetOS() == "windows" {
		// For Windows, use the directory containing the executable
		exePath, err := p.GetExecutablePath()
		if err != nil {
			return "", err
		}
		return filepath.Dir(exePath), nil
	}

	// For Unix-like systems, use user's home directory
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