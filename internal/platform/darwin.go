package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"theboys-launcher/pkg/types"
)

// DarwinPlatform provides macOS-specific implementations
type DarwinPlatform struct {
	CommonPlatform
}

// DetectJavaInstallations finds Java installations on macOS
func (p *DarwinPlatform) DetectJavaInstallations() ([]types.JavaInstallation, error) {
	var installations []types.JavaInstallation

	// Check common Java installation paths on macOS
	javaPaths := []string{
		"/Library/Java/JavaVirtualMachines",
		"/System/Library/Java/JavaVirtualMachines",
		"/usr/local/opt/openjdk",
		"/opt/homebrew/opt/openjdk",
	}

	for _, basePath := range javaPaths {
		if p.FileExists(basePath) {
			// List subdirectories
			entries, err := filepath.Glob(filepath.Join(basePath, "*"))
			if err != nil {
				continue
			}

			for _, entry := range entries {
				javaExe := filepath.Join(entry, "Contents", "Home", "bin", "java")
				if !p.FileExists(javaExe) {
					javaExe = filepath.Join(entry, "bin", "java")
				}

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

	// Also check /usr/libexec/java_home
	if cmd := exec.Command("/usr/libexec/java_home", "-V"); cmd != nil {
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Parse java_home output for additional installations
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "/Library/Java/JavaVirtualMachines") {
					fields := strings.Fields(line)
					for _, field := range fields {
						if strings.Contains(field, "/Library/Java/JavaVirtualMachines") {
							javaExe := filepath.Join(field, "Contents", "Home", "bin", "java")
							if p.FileExists(javaExe) {
								versionCmd := exec.Command(javaExe, "-version")
								versionOutput, versionErr := versionCmd.CombinedOutput()
								if versionErr == nil {
									version := p.parseJavaVersion(string(versionOutput))
									installations = append(installations, types.JavaInstallation{
										Path:         javaExe,
										Version:      version,
										IsJDK:        p.isJDK(field),
										Architecture: p.GetArch(),
									})
								}
							}
						}
					}
				}
			}
		}
	}

	// Check PATH
	if path, err := exec.LookPath("java"); err == nil {
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
func (p *DarwinPlatform) GetDefaultJavaPath() (string, error) {
	// Use java_home to find the default Java
	if cmd := exec.Command("/usr/libexec/java_home"); cmd != nil {
		output, err := cmd.CombinedOutput()
		if err == nil {
			javaHome := strings.TrimSpace(string(output))
			javaExe := filepath.Join(javaHome, "bin", "java")
			if p.FileExists(javaExe) {
				return javaExe, nil
			}
		}
	}

	// Check system PATH
	if path, err := exec.LookPath("java"); err == nil {
		return path, nil
	}

	// Check common installation locations
	javaPaths := []string{
		"/Library/Java/JavaVirtualMachines/openjdk.jdk/Contents/Home/bin/java",
		"/Library/Java/JavaVirtualMachines/zulu-21.jdk/Contents/Home/bin/java",
		"/Library/Java/JavaVirtualMachines/zulu-17.jdk/Contents/Home/bin/java",
		"/Library/Java/JavaVirtualMachines/zulu-11.jdk/Contents/Home/bin/java",
		"/usr/local/opt/openjdk/bin/java",
		"/opt/homebrew/opt/openjdk/bin/java",
	}

	for _, path := range javaPaths {
		if p.FileExists(path) {
			return path, nil
		}
	}

	return "", exec.ErrNotFound
}

// IsJavaCompatible checks if a Java installation is compatible with the required version
func (p *DarwinPlatform) IsJavaCompatible(path, requiredVersion string) bool {
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	version := p.parseJavaVersion(string(output))
	return p.isVersionCompatible(version, requiredVersion)
}

// LaunchProcess launches a process with the given command and arguments
func (p *DarwinPlatform) LaunchProcess(cmd string, args []string, workingDir string) error {
	process := exec.Command(cmd, args...)
	if workingDir != "" {
		process.Dir = workingDir
	}
	return process.Start()
}

// TerminateProcess terminates a process by PID
func (p *DarwinPlatform) TerminateProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// GetFilePermissions returns file permissions
func (p *DarwinPlatform) GetFilePermissions(path string) (uint32, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(path, &stat)
	return stat.Mode, err
}

// ValidateFilePermissions validates that we can write to a directory
func (p *DarwinPlatform) ValidateFilePermissions(path string) error {
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
func (p *DarwinPlatform) SetWindowState(state string) error {
	// macOS implementation would use AppleScript or NSWindow APIs
	return exec.ErrNotFound
}

// ShowNotification shows a desktop notification
func (p *DarwinPlatform) ShowNotification(title, message string) error {
	// Use osascript to show notification
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
	return cmd.Run()
}

// OpenURL opens a URL in the default browser
func (p *DarwinPlatform) OpenURL(url string) error {
	return exec.Command("open", url).Start()
}

// GetSupportedFileExtensions returns supported file extensions
func (p *DarwinPlatform) GetSupportedFileExtensions() []string {
	return []string{".jar", ".zip", ".json", ".toml", ".app", ".command", ".sh"}
}

// Helper functions

func (p *DarwinPlatform) parseJavaVersion(output string) string {
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

func (p *DarwinPlatform) isJDK(javaPath string) bool {
	indicators := []string{
		filepath.Join(javaPath, "Contents", "Home", "bin", "javac"),
		filepath.Join(javaPath, "Contents", "Home", "include"),
		filepath.Join(javaPath, "Contents", "Home", "lib"),
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

func (p *DarwinPlatform) isVersionCompatible(current, required string) bool {
	return current != "unknown"
}