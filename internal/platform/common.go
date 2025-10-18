package platform

import (
	"fmt"
	"os"
	"runtime"
	"path/filepath"
	"time"
	"theboys-launcher/pkg/types"
)

// CommonPlatform provides cross-platform implementations
type CommonPlatform struct{}

// Platform type declarations - these are extended in platform-specific files
type WindowsPlatform struct {
	CommonPlatform
}

type DarwinPlatform struct {
	CommonPlatform
}

type LinuxPlatform struct {
	CommonPlatform
}

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
	// Use a simple cross-platform implementation
	// On Windows, we can't easily get disk space without complex syscalls
	// For now, return a reasonable default value
	return 10737418240, nil // 10GB default
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

// NewPlatform creates a new platform-specific implementation
func NewPlatform() Platform {
	switch runtime.GOOS {
	case "windows":
		return &WindowsPlatform{CommonPlatform{}}
	case "darwin":
		return &DarwinPlatform{CommonPlatform{}}
	case "linux":
		return &LinuxPlatform{CommonPlatform{}}
	default:
		return &LinuxPlatform{CommonPlatform{}}
	}
}

// Default interface implementations - these can be overridden by platform-specific code

// DetectJavaInstallations provides a default implementation
func (p *CommonPlatform) DetectJavaInstallations() ([]types.JavaInstallation, error) {
	return []types.JavaInstallation{}, nil
}

// GetDefaultJavaPath provides a default implementation
func (p *CommonPlatform) GetDefaultJavaPath() (string, error) {
	return "", fmt.Errorf("Java not found")
}

// IsJavaCompatible provides a default implementation
func (p *CommonPlatform) IsJavaCompatible(path, requiredVersion string) bool {
	return false
}

// LaunchProcess provides a default implementation
func (p *CommonPlatform) LaunchProcess(cmd string, args []string, workingDir string) error {
	return fmt.Errorf("process launch not implemented")
}

// TerminateProcess provides a default implementation
func (p *CommonPlatform) TerminateProcess(pid int) error {
	return fmt.Errorf("process termination not implemented")
}

// GetFilePermissions provides a default implementation
func (p *CommonPlatform) GetFilePermissions(path string) (uint32, error) {
	return 0644, nil
}

// ValidateFilePermissions provides a default implementation
func (p *CommonPlatform) ValidateFilePermissions(path string) error {
	// Try to create a test file
	testFile := filepath.Join(path, ".theboys-launcher-test")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}

// SetWindowState provides a default implementation
func (p *CommonPlatform) SetWindowState(state string) error {
	return fmt.Errorf("window state management not implemented")
}

// ShowNotification provides a default implementation
func (p *CommonPlatform) ShowNotification(title, message string) error {
	fmt.Printf("Notification: %s - %s\n", title, message)
	return nil
}

// OpenURL provides a default implementation
func (p *CommonPlatform) OpenURL(url string) error {
	fmt.Printf("Please open this URL in your browser: %s\n", url)
	return nil
}

// GetSupportedFileExtensions provides a default implementation
func (p *CommonPlatform) GetSupportedFileExtensions() []string {
	return []string{".jar", ".zip", ".json", ".toml"}
}