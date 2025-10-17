package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"theboys-launcher/pkg/types"
)

// WindowsPlatform provides Windows-specific implementations
type WindowsPlatform struct {
	CommonPlatform
}

// DetectJavaInstallations finds Java installations on Windows
func (p *WindowsPlatform) DetectJavaInstallations() ([]types.JavaInstallation, error) {
	var installations []types.JavaInstallation

	// Check common Java installation paths
	javaPaths := []string{
		`C:\Program Files\Java`,
		`C:\Program Files (x86)\Java`,
	}

	for _, basePath := range javaPaths {
		if p.FileExists(basePath) {
			// List subdirectories
			entries, err := filepath.Glob(filepath.Join(basePath, "*"))
			if err != nil {
				continue
			}

			for _, entry := range entries {
				javaExe := filepath.Join(entry, "bin", "java.exe")
				if p.FileExists(javaExe) {
					// Get version information
					cmd := exec.Command(javaExe, "-version")
					output, err := cmd.CombinedOutput()
					if err == nil {
						version := p.parseJavaVersion(string(output))
						installations = append(installations, types.JavaInstallation{
							Path:         javaExe,
							Version:      version,
							IsJDK:        p.isJDK(entry),
							Architecture: p.GetArch(),
						})
					}
				}
			}
		}
	}

	// Also check PATH
	if path, err := exec.LookPath("java.exe"); err == nil {
		cmd := exec.Command(path, "-version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			version := p.parseJavaVersion(string(output))
			installations = append(installations, types.JavaInstallation{
				Path:         path,
				Version:      version,
				IsJDK:        false,
				Architecture: p.GetArch(),
			})
		}
	}

	return installations, nil
}

// GetDefaultJavaPath returns the default Java installation path
func (p *WindowsPlatform) GetDefaultJavaPath() (string, error) {
	// Check system PATH first
	if path, err := exec.LookPath("java.exe"); err == nil {
		return path, nil
	}

	// Check common installation locations
	javaPaths := []string{
		`C:\Program Files\Java\jdk-21\bin\java.exe`,
		`C:\Program Files\Java\jdk-17\bin\java.exe`,
		`C:\Program Files\Java\jdk-11\bin\java.exe`,
		`C:\Program Files (x86)\Java\jdk-21\bin\java.exe`,
		`C:\Program Files (x86)\Java\jdk-17\bin\java.exe`,
	}

	for _, path := range javaPaths {
		if p.FileExists(path) {
			return path, nil
		}
	}

	return "", exec.ErrNotFound
}

// IsJavaCompatible checks if a Java installation is compatible with the required version
func (p *WindowsPlatform) IsJavaCompatible(path, requiredVersion string) bool {
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	version := p.parseJavaVersion(string(output))
	return p.isVersionCompatible(version, requiredVersion)
}

// LaunchProcess launches a process with the given command and arguments
func (p *WindowsPlatform) LaunchProcess(cmd string, args []string, workingDir string) error {
	process := exec.Command(cmd, args...)
	if workingDir != "" {
		process.Dir = workingDir
	}
	return process.Start()
}

// TerminateProcess terminates a process by PID
func (p *WindowsPlatform) TerminateProcess(pid int) error {
	// Windows-specific process termination
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// GetFilePermissions returns file permissions
func (p *WindowsPlatform) GetFilePermissions(path string) (uint32, error) {
	// Windows uses different permission model
	return 0644, nil
}

// ValidateFilePermissions validates that we can write to a directory
func (p *WindowsPlatform) ValidateFilePermissions(path string) error {
	// Try to create a test file
	testFile := filepath.Join(path, ".winterpack_test")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}

// SetWindowState sets the window state (minimize, maximize, restore)
func (p *WindowsPlatform) SetWindowState(state string) error {
	// Windows-specific window state management
	// Implementation would use Windows API calls
	return nil
}

// ShowNotification shows a desktop notification
func (p *WindowsPlatform) ShowNotification(title, message string) error {
	// Windows toast notification
	// Implementation would use Windows 10+ notification APIs
	return nil
}

// OpenURL opens a URL in the default browser
func (p *WindowsPlatform) OpenURL(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// GetSupportedFileExtensions returns supported file extensions
func (p *WindowsPlatform) GetSupportedFileExtensions() []string {
	return []string{".jar", ".zip", ".json", ".toml", ".exe", ".bat"}
}

// Helper functions

func (p *WindowsPlatform) parseJavaVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && end > start {
				return line[start+1 : end]
			}
		}
	}
	return "unknown"
}

func (p *WindowsPlatform) isJDK(javaPath string) bool {
	indicators := []string{
		filepath.Join(javaPath, "bin", "javac.exe"),
		filepath.Join(javaPath, "include"),
		filepath.Join(javaPath, "lib"),
	}

	for _, indicator := range indicators {
		if p.FileExists(indicator) {
			return true
		}
	}

	return false
}

func (p *WindowsPlatform) isVersionCompatible(current, required string) bool {
	return current != "unknown"
}