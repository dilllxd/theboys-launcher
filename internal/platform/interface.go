package platform

import (
	"theboys-launcher/pkg/types"
)

// Platform defines platform-specific operations
type Platform interface {
	// System information
	GetOS() string
	GetArch() string
	GetExecutablePath() (string, error)
	GetAppDataDir() (string, error)
	GetCustomDataDir() (string, error) // Custom installation directory support

	// Java management
	DetectJavaInstallations() ([]types.JavaInstallation, error)
	GetDefaultJavaPath() (string, error)
	IsJavaCompatible(path, requiredVersion string) bool

	// Process management
	LaunchProcess(cmd string, args []string, workingDir string) error
	TerminateProcess(pid int) error

	// File system
	CreateDirectory(path string) error
	FileExists(path string) bool
	GetFilePermissions(path string) (uint32, error)

	// System resources
	GetDefaultRAM() int
	GetMaxRAM() int
	GetAvailableDiskSpace(path string) (int64, error)

	// Platform-specific features
	SupportsAutoUpdate() bool
	GetSupportedFileExtensions() []string
	ValidateFilePermissions(path string) error
	CanCreateShortcut() bool
	CreateShortcut(target, shortcutPath string) error

	// UI integration
	SetWindowState(state string) error // minimize, maximize, restore
	ShowNotification(title, message string) error
	OpenURL(url string) error

	// Installation support
	IsInstalled() bool
	GetInstallationPath() (string, error)
	RegisterInstallation(path string) error
	UnregisterInstallation() error
}