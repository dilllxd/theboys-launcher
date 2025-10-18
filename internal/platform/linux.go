//go:build linux
// +build linux

package platform

import (
	"os/exec"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"theboys-launcher/pkg/types"
)



// DetectJavaInstallations finds Java installations on Linux
func (p *LinuxPlatform) DetectJavaInstallations() ([]types.JavaInstallation, error) {
	var installations []types.JavaInstallation

	// Common Java installation paths on Linux
	javaPaths := []string{
		"/usr/lib/jvm",
		"/usr/java",
		"/opt/java",
		"/usr/lib/jvm/default-java",
		"/usr/lib/jvm/java-11-openjdk-amd64",
		"/usr/lib/jvm/java-17-openjdk-amd64",
		"/usr/lib/jvm/java-21-openjdk-amd64",
	}

	for _, path := range javaPaths {
		if p.FileExists(path) {
			// Look for java executable in subdirectories
			javaExe := filepath.Join(path, "bin", "java")
			if p.FileExists(javaExe) {
				// Get version information
				cmd := exec.Command(javaExe, "-version")
				output, err := cmd.CombinedOutput()
				if err == nil {
					version := p.parseJavaVersion(string(output))
					installations = append(installations, types.JavaInstallation{
						Path:         javaExe,
						Version:      version,
						IsJDK:        p.isJDK(path),
						Architecture: p.GetArch(),
					})
				}
			}
		}
	}

	// Also check PATH
	if path, err := exec.LookPath("java"); err == nil {
		cmd := exec.Command(path, "-version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			version := p.parseJavaVersion(string(output))
			installations = append(installations, types.JavaInstallation{
				Path:         path,
				Version:      version,
				IsJDK:        false, // Unknown if PATH contains JDK or JRE
				Architecture: p.GetArch(),
			})
		}
	}

	return installations, nil
}

// GetDefaultJavaPath returns the default Java installation path
func (p *LinuxPlatform) GetDefaultJavaPath() (string, error) {
	// Check system PATH first
	if path, err := exec.LookPath("java"); err == nil {
		return path, nil
	}

	// Check common installation locations
	javaPaths := []string{
		"/usr/lib/jvm/default-java/bin/java",
		"/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
		"/usr/lib/jvm/java-17-openjdk-amd64/bin/java",
		"/usr/lib/jvm/java-11-openjdk-amd64/bin/java",
	}

	for _, path := range javaPaths {
		if p.FileExists(path) {
			return path, nil
		}
	}

	return "", exec.ErrNotFound
}

// IsJavaCompatible checks if a Java installation is compatible with the required version
func (p *LinuxPlatform) IsJavaCompatible(path, requiredVersion string) bool {
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	version := p.parseJavaVersion(string(output))
	return p.isVersionCompatible(version, requiredVersion)
}

// LaunchProcess launches a process with the given command and arguments
func (p *LinuxPlatform) LaunchProcess(cmd string, args []string, workingDir string) error {
	process := exec.Command(cmd, args...)
	if workingDir != "" {
		process.Dir = workingDir
	}
	process.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return process.Start()
}

// TerminateProcess terminates a process by PID
func (p *LinuxPlatform) TerminateProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// GetFilePermissions returns file permissions
func (p *LinuxPlatform) GetFilePermissions(path string) (uint32, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(path, &stat)
	return stat.Mode, err
}

// ValidateFilePermissions validates that we can write to a directory
func (p *LinuxPlatform) ValidateFilePermissions(path string) error {
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
func (p *LinuxPlatform) SetWindowState(state string) error {
	// Linux implementation would depend on the window manager
	// For now, return not implemented
	return exec.ErrNotFound
}

// ShowNotification shows a desktop notification
func (p *LinuxPlatform) ShowNotification(title, message string) error {
	// Try notify-send first
	if _, err := exec.LookPath("notify-send"); err == nil {
		cmd := exec.Command("notify-send", title, message)
		return cmd.Run()
	}

	// Fallback to echo
	return nil
}

// OpenURL opens a URL in the default browser
func (p *LinuxPlatform) OpenURL(url string) error {
	// Try xdg-open first
	if _, err := exec.LookPath("xdg-open"); err == nil {
		cmd := exec.Command("xdg-open", url)
		return cmd.Run()
	}

	// Fallback commands for different desktop environments
	browsers := []string{"firefox", "chromium", "google-chrome", "opera"}
	for _, browser := range browsers {
		if _, err := exec.LookPath(browser); err == nil {
			cmd := exec.Command(browser, url)
			return cmd.Run()
		}
	}

	return exec.ErrNotFound
}

// GetSupportedFileExtensions returns supported file extensions
func (p *LinuxPlatform) GetSupportedFileExtensions() []string {
	return []string{".jar", ".zip", ".json", ".toml", ".sh"}
}

// Helper functions

func (p *LinuxPlatform) parseJavaVersion(output string) string {
	// Parse java version output to extract version string
	// This is a simplified implementation
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			// Extract version between quotes
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && end > start {
				return line[start+1 : end]
			}
		}
	}
	return "unknown"
}

func (p *LinuxPlatform) isJDK(javaPath string) bool {
	// Check for JDK indicators
	indicators := []string{
		filepath.Join(javaPath, "bin", "javac"),
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

func (p *LinuxPlatform) isVersionCompatible(current, required string) bool {
	// Simplified version compatibility check
	// In a real implementation, this would be more sophisticated
	return current != "unknown"
}