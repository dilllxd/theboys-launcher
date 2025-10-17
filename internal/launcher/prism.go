package launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

const (
	prismGitHubAPIURL      = "https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/latest"
	prismDownloadTimeout    = 10 * time.Minute
	defaultPrismVersion    = "latest"
	prismConfigFile        = "prismlauncher.cfg"
	instanceConfigFile     = "instance.cfg"
	packDotTomlFile       = "pack.toml"
)

// PrismRelease represents a Prism Launcher release from GitHub
type PrismRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// PrismManager handles Prism Launcher operations
type PrismManager struct {
	platform   platform.Platform
	logger     *logging.Logger
	downloader *Downloader
	installer  *JavaInstaller
}

// NewPrismManager creates a new Prism manager instance
func NewPrismManager(platform platform.Platform, logger *logging.Logger) *PrismManager {
	return &PrismManager{
		platform:   platform,
		logger:     logger,
		downloader: NewDownloader(platform, logger),
		installer:  NewJavaInstaller(platform, logger),
	}
}

// EnsurePrismInstallation ensures Prism Launcher is installed and returns whether it was downloaded
func (p *PrismManager) EnsurePrismInstallation(installDir string) (bool, error) {
	p.logger.Info("Checking Prism Launcher installation in %s", installDir)

	// Check if Prism is already installed
	prismExe := p.getPrismExecutable(installDir)
	if p.platform.FileExists(prismExe) {
		p.logger.Debug("Prism Launcher already installed at %s", prismExe)
		return false, nil
	}

	// Download latest Prism
	p.logger.Info("Prism Launcher not found, downloading latest version")
	downloadURL, err := p.fetchLatestPrismURL()
	if err != nil {
		return false, fmt.Errorf("failed to fetch Prism download URL: %w", err)
	}

	p.logger.Info("Downloading Prism from %s", downloadURL)

	// Download Prism
	ctx, cancel := context.WithTimeout(context.Background(), prismDownloadTimeout)
	defer cancel()

	zipPath := filepath.Join(installDir, "prism.zip")
	err = p.downloader.DownloadFile(ctx, downloadURL, zipPath, nil)
	if err != nil {
		return false, fmt.Errorf("failed to download Prism: %w", err)
	}

	// Extract Prism
	if err := p.installer.InstallJavaFromZip(zipPath, installDir); err != nil {
		return false, fmt.Errorf("failed to extract Prism: %w", err)
	}

	// Force portable mode
	p.logger.Debug("Configuring Prism for portable mode")
	configPath := filepath.Join(installDir, prismConfigFile)
	configContent := "Portable=true\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		p.logger.Warn("Failed to write Prism config: %v", err)
	}

	// Clean up zip file
	os.Remove(zipPath)

	p.logger.Info("Prism Launcher installed successfully")
	return true, nil
}

// GetPrismVersion returns the installed Prism version (if detectable)
func (p *PrismManager) GetPrismVersion(installDir string) string {
	prismExe := p.getPrismExecutable(installDir)
	if !p.platform.FileExists(prismExe) {
		return ""
	}

	// Try to get version from executable name or directory
	// This is a simplified approach - could be enhanced
	return "unknown"
}

// IsPrismInstalled checks if Prism Launcher is installed
func (p *PrismManager) IsPrismInstalled(installDir string) bool {
	prismExe := p.getPrismExecutable(installDir)
	return p.platform.FileExists(prismExe)
}

// LaunchPrism launches Prism Launcher with the given instance
func (p *PrismManager) LaunchPrism(installDir, instanceID string, workingDir string) error {
	prismExe := p.getPrismExecutable(installDir)
	if !p.platform.FileExists(prismExe) {
		return fmt.Errorf("Prism Launcher not found at %s", prismExe)
	}

	args := []string{}
	if instanceID != "" {
		args = append(args, "--launch", instanceID)
	}

	p.logger.Info("Launching Prism Launcher for instance %s", instanceID)

	return p.platform.LaunchProcess(prismExe, args, workingDir)
}

// CreateInstance creates a new Prism instance for a modpack
func (p *PrismManager) CreateInstance(modpack types.Modpack, packInfo *PackInfo, instanceDir, javaExe string) error {
	p.logger.Info("Creating Prism instance for %s in %s", modpack.DisplayName, instanceDir)

	// Create instance directory
	if err := p.platform.CreateDirectory(instanceDir); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Calculate optimal memory allocation
	minMB, maxMB := p.calculateOptimalMemory()

	// Create instance.cfg
	instanceConfig := fmt.Sprintf(`InstanceType=OneSix
name=%s
iconKey=default
OverrideMemory=true
MinMemAlloc=%d
MaxMemAlloc=%d
OverrideJava=true
JavaPath=%s
Notes=Managed by TheBoys Launcher
`,
		modpack.InstanceName,
		minMB,
		maxMB,
		javaExe,
	)

	instanceConfigPath := filepath.Join(instanceDir, instanceConfigFile)
	if err := os.WriteFile(instanceConfigPath, []byte(instanceConfig), 0644); err != nil {
		return fmt.Errorf("failed to write instance.cfg: %w", err)
	}

	// Create components.json
	if err := p.createComponentConfig(packInfo, instanceDir); err != nil {
		return fmt.Errorf("failed to create components.json: %w", err)
	}

	// Create minecraft directory
	minecraftDir := filepath.Join(instanceDir, ".minecraft")
	if err := p.platform.CreateDirectory(minecraftDir); err != nil {
		return fmt.Errorf("failed to create .minecraft directory: %w", err)
	}

	p.logger.Info("Prism instance created successfully")
	return nil
}

// GetPrismExecutable returns the path to the Prism Launcher executable
func (p *PrismManager) GetPrismExecutable(installDir string) string {
	return p.getPrismExecutable(installDir)
}

// Helper methods

func (p *PrismManager) getPrismExecutable(installDir string) string {
	os := p.platform.GetOS()
	switch os {
	case "windows":
		return filepath.Join(installDir, "PrismLauncher.exe")
	case "darwin":
		return filepath.Join(installDir, "PrismLauncher")
	case "linux":
		return filepath.Join(installDir, "PrismLauncher")
	default:
		return filepath.Join(installDir, "PrismLauncher")
	}
}

func (p *PrismManager) fetchLatestPrismURL() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", prismGitHubAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", p.getUserAgent("Prism"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release PrismRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Build priority patterns by architecture
	type pattern struct {
		re *regexp.Regexp
	}

	var patterns []pattern

	if runtime.GOARCH == "amd64" {
		// 1) MinGW w64 portable zip
		patterns = append(patterns, pattern{regexp.MustCompile(`(?i)Windows-MinGW-w64-Portable-.*\.zip$`)})
		// 2) MSVC portable zip
		patterns = append(patterns, pattern{regexp.MustCompile(`(?i)Windows-MSVC-Portable-.*\.zip$`)})
	} else if runtime.GOARCH == "arm64" {
		// MSVC arm64 portable zip
		patterns = append(patterns, pattern{regexp.MustCompile(`(?i)Windows-MSVC-arm64-Portable-.*\.zip$`)})
	}

	// Fallbacks for unexpected naming: generic portable zips
	patterns = append(patterns,
		pattern{regexp.MustCompile(`(?i)Windows-.*Portable-.*\.zip$`)},
		pattern{regexp.MustCompile(`(?i)Windows-.*\.zip$`)},
	)

	// Search in priority order
	for _, pat := range patterns {
		for _, asset := range release.Assets {
			if pat.re.MatchString(asset.Name) {
				p.logger.Info("Found Prism asset: %s", asset.Name)
				return asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", fmt.Errorf("no suitable Prism portable asset found in latest release")
}

func (p *PrismManager) createComponentConfig(packInfo *PackInfo, instanceDir string) error {
	// Build components dynamically based on pack info
	lwjglInfo := p.getLWJGLVersionForMinecraft(packInfo.Minecraft)

	components := make([]interface{}, 0)

	// Add LWJGL component
	components = append(components, map[string]interface{}{
		"cachedName":     lwjglInfo.Name,
		"cachedVersion":  lwjglInfo.Version,
		"cachedVolatile": true,
		"dependencyOnly": true,
		"uid":            lwjglInfo.UID,
		"version":        lwjglInfo.Version,
	})

	// Add Minecraft component
	components = append(components, map[string]interface{}{
		"cachedName":    "Minecraft",
		"cachedVersion": packInfo.Minecraft,
		"cachedRequires": []interface{}{
			map[string]interface{}{
				"suggests": lwjglInfo.Version,
				"uid":      lwjglInfo.UID,
			},
		},
		"important": true,
		"uid":       "net.minecraft",
		"version":   packInfo.Minecraft,
	})

	// Add modloader component if present
	if packInfo.ModLoader != "" {
		modloaderComponent := p.createModloaderComponent(packInfo)
		if modloaderComponent != nil {
			components = append(components, modloaderComponent)
		}
	}

	// Write components.json
	componentsPath := filepath.Join(instanceDir, "components.json")
	componentsData, err := json.MarshalIndent(components, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	if err := os.WriteFile(componentsPath, componentsData, 0644); err != nil {
		return fmt.Errorf("failed to write components.json: %w", err)
	}

	return nil
}

func (p *PrismManager) createModloaderComponent(packInfo *PackInfo) map[string]interface{} {
	switch packInfo.ModLoader {
	case "forge":
		return map[string]interface{}{
			"cachedName":    "Forge",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "net.minecraftforge",
			"version": packInfo.LoaderVersion,
		}
	case "fabric":
		return map[string]interface{}{
			"cachedName":    "Fabric Loader",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "net.fabricmc.fabric-loader",
			"version": packInfo.LoaderVersion,
		}
	case "quilt":
		return map[string]interface{}{
			"cachedName":    "Quilt Loader",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "org.quiltmc.quilt-loader",
			"version": packInfo.LoaderVersion,
		}
	}
	return nil
}

func (p *PrismManager) calculateOptimalMemory() (int, int) {
	// Use platform-specific defaults
	defaultRAM := p.platform.GetDefaultRAM()
	maxRAM := p.platform.GetMaxRAM()

	// Use 75% of available RAM as maximum, but respect platform limits
	availableRAM := defaultRAM
	if availableRAM > maxRAM {
		availableRAM = maxRAM
	}

	// Minimum 2GB, maximum availableRAM
	minMB := 2048
	if minMB > availableRAM {
		minMB = availableRAM
	}

	maxMB := availableRAM

	return minMB, maxMB
}

func (p *PrismManager) getUserAgent(component string) string {
	return fmt.Sprintf("TheBoys-%s/dev", component)
}

// PackInfo represents information extracted from a pack.toml file
type PackInfo struct {
	Minecraft     string `json:"minecraft"`
	ModLoader     string `json:"modloader"`
	LoaderVersion string `json:"loader_version"`
	Name          string `json:"name"`
}

// LWJGLInfo holds version and UID information for LWJGL
type LWJGLInfo struct {
	Version string `json:"version"`
	UID     string `json:"uid"`
	Name    string `json:"name"`
}

// getLWJGLVersionForMinecraft fetches LWJGL version from PrismLauncher meta-launcher GitHub
func (p *PrismManager) getLWJGLVersionForMinecraft(mcVersion string) LWJGLInfo {
	cleanVersion := strings.TrimSpace(mcVersion)
	if cleanVersion == "" {
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}

	// For now, return a default LWJGL version
	// In a full implementation, this would fetch from PrismLauncher meta-launcher GitHub
	// similar to how Java version compatibility is determined
	return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"}
}