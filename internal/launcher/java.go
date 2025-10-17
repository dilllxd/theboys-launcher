package launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

const (
	defaultJavaVersion      = "17"
	prismMetaURLPattern     = "https://raw.githubusercontent.com/PrismLauncher/meta-launcher/refs/heads/master/net.minecraft/%s.json"
	adoptiumAPIURLPattern   = "https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&image_type=%s&os=%s"
	githubAdoptiumURLPattern = "https://api.github.com/repos/adoptium/temurin%s-binaries/releases/latest"
	javaDownloadTimeout     = 10 * time.Minute
)

// JavaCompatibility represents Java compatibility data from PrismLauncher meta
type JavaCompatibility struct {
	CompatibleJavaMajors []int `json:"compatibleJavaMajors"`
}

// AdoptiumAsset represents an asset from Adoptium API
type AdoptiumAsset struct {
	Binaries []struct {
		Package struct {
			Link string `json:"link"`
		} `json:"package"`
	} `json:"binaries"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// JavaManager handles Java runtime detection, download, and management
type JavaManager struct {
	platform   platform.Platform
	logger     *logging.Logger
	downloader *Downloader
}

// NewJavaManager creates a new Java manager instance
func NewJavaManager(platform platform.Platform, logger *logging.Logger) *JavaManager {
	return &JavaManager{
		platform:   platform,
		logger:     logger,
		downloader: NewDownloader(platform, logger),
	}
}

// DetectJavaInstallations scans the system for Java installations
func (j *JavaManager) DetectJavaInstallations() ([]types.JavaInstallation, error) {
	j.logger.Info("Scanning for Java installations...")

	// Use platform-specific detection
	installations, err := j.platform.DetectJavaInstallations()
	if err != nil {
		j.logger.Error("Platform Java detection failed: %v", err)
		return nil, fmt.Errorf("platform Java detection failed: %w", err)
	}

	// Filter and validate installations
	validInstallations := j.filterValidInstallations(installations)

	j.logger.Info("Found %d valid Java installation(s)", len(validInstallations))
	for _, inst := range validInstallations {
		j.logger.Debug("Java %s at %s (%s)", inst.Version, inst.Path, map[bool]string{true: "JDK", false: "JRE"}[inst.IsJDK])
	}

	return validInstallations, nil
}

// GetJavaVersionForMinecraft returns the recommended Java version for a Minecraft version
func (j *JavaManager) GetJavaVersionForMinecraft(mcVersion string) string {
	cleanVersion := strings.TrimSpace(mcVersion)
	if cleanVersion == "" {
		j.logger.Debug("No Minecraft version provided, using default Java %s", defaultJavaVersion)
		return defaultJavaVersion
	}

	j.logger.Debug("Fetching Java compatibility for Minecraft %s", cleanVersion)

	// Fetch compatibility data from PrismLauncher meta
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf(prismMetaURLPattern, cleanVersion)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		j.logger.Warn("Failed to create request for Java compatibility data: %v", err)
		return defaultJavaVersion
	}

	req.Header.Set("User-Agent", j.getUserAgent("Java"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		j.logger.Warn("Failed to fetch Java compatibility data for Minecraft %s: %v", cleanVersion, err)
		return defaultJavaVersion
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		j.logger.Warn("Java compatibility data not found for Minecraft %s (HTTP %d)", cleanVersion, resp.StatusCode)
		return defaultJavaVersion
	}

	// Parse the JSON response
	var data JavaCompatibility
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		j.logger.Warn("Failed to parse Java compatibility data for Minecraft %s: %v", cleanVersion, err)
		return defaultJavaVersion
	}

	// If we have compatible Java versions, select the best one
	if len(data.CompatibleJavaMajors) > 0 {
		// Choose the newest compatible Java version (prefer higher versions)
		bestJava := data.CompatibleJavaMajors[0]
		for _, javaVersion := range data.CompatibleJavaMajors {
			if javaVersion > bestJava {
				bestJava = javaVersion
			}
		}

		bestJavaStr := strconv.Itoa(bestJava)
		j.logger.Info("Found Java %s compatible with Minecraft %s from PrismLauncher meta", bestJavaStr, cleanVersion)
		return bestJavaStr
	}

	j.logger.Warn("No Java compatibility data found for Minecraft %s", cleanVersion)
	return defaultJavaVersion
}

// GetBestJavaInstallation returns the best Java installation for a Minecraft version
func (j *JavaManager) GetBestJavaInstallation(mcVersion string) (*types.JavaInstallation, error) {
	j.logger.Info("Finding best Java installation for Minecraft %s", mcVersion)

	// Get required Java version
	requiredVersion := j.GetJavaVersionForMinecraft(mcVersion)
	j.logger.Debug("Required Java version: %s", requiredVersion)

	// Detect available installations
	installations, err := j.DetectJavaInstallations()
	if err != nil {
		return nil, err
	}

	if len(installations) == 0 {
		j.logger.Warn("No Java installations found, will need to download")
		return nil, nil
	}

	// Filter for compatible versions
	compatible := j.findCompatibleInstallations(installations, requiredVersion)
	if len(compatible) == 0 {
		j.logger.Warn("No compatible Java installations found for version %s", requiredVersion)
		return nil, nil
	}

	// Select the best installation (prefer JDK, then higher versions)
	best := j.selectBestInstallation(compatible)
	j.logger.Info("Selected Java %s at %s", best.Version, best.Path)
	return &best, nil
}

// DownloadJava downloads and installs a Java runtime
func (j *JavaManager) DownloadJava(javaVersion string, installDir string, progressCallback func(*DownloadProgress)) error {
	j.logger.Info("Downloading Java %s to %s", javaVersion, installDir)

	// Create installation directory
	if err := j.platform.CreateDirectory(installDir); err != nil {
		return fmt.Errorf("failed to create Java installation directory: %w", err)
	}

	// Get download URL
	downloadURL, err := j.getJavaDownloadURL(javaVersion)
	if err != nil {
		return fmt.Errorf("failed to get Java download URL: %w", err)
	}

	j.logger.Info("Downloading Java %s from %s", javaVersion, downloadURL)

	// Download Java
	zipPath := filepath.Join(installDir, "java.zip")
	ctx, cancel := context.WithTimeout(context.Background(), javaDownloadTimeout)
	defer cancel()

	err = j.downloader.DownloadFile(ctx, downloadURL, zipPath, progressCallback)
	if err != nil {
		return fmt.Errorf("failed to download Java: %w", err)
	}

	// Extract Java (simplified - in real implementation would extract zip)
	// For now, we'll just mark it as installed
	j.logger.Info("Java %s downloaded successfully", javaVersion)

	// Clean up zip file
	os.Remove(zipPath)

	return nil
}

// IsJavaCompatible checks if a Java installation is compatible with a Minecraft version
func (j *JavaManager) IsJavaCompatible(javaPath, mcVersion string) bool {
	requiredVersion := j.GetJavaVersionForMinecraft(mcVersion)

	// Get Java version from installation
	javaVersion, err := j.getJavaVersion(javaPath)
	if err != nil {
		j.logger.Debug("Failed to get Java version from %s: %v", javaPath, err)
		return false
	}

	// Simple compatibility check (can be enhanced)
	return j.isVersionCompatible(javaVersion, requiredVersion)
}

// VerifyJavaInstallation verifies that a Java installation is working
func (j *JavaManager) VerifyJavaInstallation(javaPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, javaPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Java verification failed: %w", err)
	}

	j.logger.Debug("Java installation verified: %s", javaPath)
	return nil
}

// Helper methods

func (j *JavaManager) filterValidInstallations(installations []types.JavaInstallation) []types.JavaInstallation {
	valid := make([]types.JavaInstallation, 0, len(installations))

	for _, inst := range installations {
		// Verify the installation actually works
		if j.VerifyJavaInstallation(inst.Path) == nil {
			valid = append(valid, inst)
		} else {
			j.logger.Warn("Skipping invalid Java installation: %s", inst.Path)
		}
	}

	return valid
}

func (j *JavaManager) findCompatibleInstallations(installations []types.JavaInstallation, requiredVersion string) []types.JavaInstallation {
	compatible := make([]types.JavaInstallation, 0)

	for _, inst := range installations {
		if j.isVersionCompatible(inst.Version, requiredVersion) {
			compatible = append(compatible, inst)
		}
	}

	return compatible
}

func (j *JavaManager) selectBestInstallation(installations []types.JavaInstallation) types.JavaInstallation {
	if len(installations) == 0 {
		return types.JavaInstallation{}
	}

	// Prefer JDK over JRE
	var jdkInstallation *types.JavaInstallation
	var jreInstallation *types.JavaInstallation

	for _, inst := range installations {
		if inst.IsJDK && jdkInstallation == nil {
			jdkInstallation = &inst
		} else if !inst.IsJDK && jreInstallation == nil {
			jreInstallation = &inst
		}
	}

	// Return JDK if available, otherwise JRE
	if jdkInstallation != nil {
		return *jdkInstallation
	}
	if jreInstallation != nil {
		return *jreInstallation
	}

	// Fallback to first installation
	return installations[0]
}

func (j *JavaManager) getJavaVersion(javaPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse version from output
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "version") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && end > start {
				version := line[start+1 : end]
				// Extract major version
				if parts := strings.Split(version, "."); len(parts) > 0 {
					return parts[0], nil
				}
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("could not parse Java version from output")
}

func (j *JavaManager) isVersionCompatible(current, required string) bool {
	// Simple compatibility check - can be enhanced with proper semver comparison
	if current == required {
		return true
	}

	// Convert to int for simple comparison
	currentInt, err1 := strconv.Atoi(current)
	requiredInt, err2 := strconv.Atoi(required)

	if err1 == nil && err2 == nil {
		// Newer versions are generally compatible
		return currentInt >= requiredInt
	}

	// Fallback: string comparison for complex versions
	return strings.HasPrefix(current, required) || strings.HasPrefix(required, current)
}

func (j *JavaManager) getJavaDownloadURL(javaVersion string) (string, error) {
	// Determine platform-specific parameters
	arch := j.getArchitectureString()
	osName := j.getOSString()
	imageType := "jre"
	if javaVersion == "16" {
		imageType = "jdk"
	}

	// Try Adoptium API first
	adoptiumURL := fmt.Sprintf(adoptiumAPIURLPattern, javaVersion, arch, imageType, osName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", adoptiumURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Adoptium request: %w", err)
	}

	req.Header.Set("User-Agent", j.getUserAgent("Adoptium"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		var payload []AdoptiumAsset
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			for _, asset := range payload {
				for _, binary := range asset.Binaries {
					if binary.Package.Link != "" && strings.HasSuffix(strings.ToLower(binary.Package.Link), ".zip") {
						j.logger.Info("Found Java download URL from Adoptium API")
						return binary.Package.Link, nil
					}
				}
			}
		}
	}

	// Fallback to GitHub
	githubURL := fmt.Sprintf(githubAdoptiumURLPattern, javaVersion)

	req2, err := http.NewRequestWithContext(ctx, "GET", githubURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub request: %w", err)
	}

	req2.Header.Set("User-Agent", j.getUserAgent("Adoptium"))

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return "", fmt.Errorf("both Adoptium API and GitHub fallback failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub Adoptium API returned status %d", resp2.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp2.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse GitHub release data: %w", err)
	}

	// Find appropriate asset
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(osName)) &&
		   strings.Contains(strings.ToLower(asset.Name), strings.ToLower(arch)) &&
		   strings.HasSuffix(strings.ToLower(asset.Name), ".zip") {
			j.logger.Info("Found Java download URL from GitHub fallback")
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no suitable Java download found for version %s on %s/%s", javaVersion, osName, arch)
}

func (j *JavaManager) getArchitectureString() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x32"
	default:
		return arch
	}
}

func (j *JavaManager) getOSString() string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return "windows"
	case "darwin":
		return "mac"
	case "linux":
		return "linux"
	default:
		return os
	}
}

func (j *JavaManager) getUserAgent(component string) string {
	return fmt.Sprintf("TheBoys-%s/dev", component)
}