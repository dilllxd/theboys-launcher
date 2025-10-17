package launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// Instance represents a Prism Launcher instance
type Instance struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ModpackID    string                 `json:"modpackId"`
	Minecraft    string                 `json:"minecraft"`
	ModLoader    string                 `json:"modloader"`
	JavaVersion  string                 `json:"javaVersion"`
	JavaPath     string                 `json:"javaPath"`
	MemoryMin    int                    `json:"memoryMin"`
	MemoryMax    int                    `json:"memoryMax"`
	InstancePath string                 `json:"instancePath"`
	PrismPath    string                 `json:"prismPath"`
	LastPlayed   *time.Time             `json:"lastPlayed,omitempty"`
	TotalTime    time.Duration          `json:"totalTime"`
	Properties   map[string]string      `json:"properties,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// InstanceManager handles instance operations
type InstanceManager struct {
	platform   platform.Platform
	logger     *logging.Logger
	prismManager *PrismManager
	javaManager  *JavaManager
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(platform platform.Platform, logger *logging.Logger, prismManager *PrismManager, javaManager *JavaManager) *InstanceManager {
	return &InstanceManager{
		platform:     platform,
		logger:       logger,
		prismManager: prismManager,
		javaManager:  javaManager,
	}
}

// CreateInstance creates a new instance for a modpack
func (im *InstanceManager) CreateInstance(modpack types.Modpack, prismDir, instancesDir string) (*Instance, error) {
	im.logger.Info("Creating instance for modpack %s", modpack.DisplayName)

	// Generate instance ID
	instanceID := im.generateInstanceID(modpack.ID)
	instancePath := filepath.Join(instancesDir, instanceID)

	// Check if instance already exists
	if im.platform.FileExists(instancePath) {
		return nil, fmt.Errorf("instance %s already exists", instanceID)
	}

	// Download and parse pack.toml
	packInfo, err := im.downloadPackInfo(modpack.PackURL, instancePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download pack info: %w", err)
	}

	// Determine Java requirements
	javaVersion := im.javaManager.GetJavaVersionForMinecraft(packInfo.Minecraft)
	javaInstallation, err := im.javaManager.GetBestJavaInstallation(packInfo.Minecraft)
	if err != nil {
		return nil, fmt.Errorf("failed to get Java installation: %w", err)
	}

	var javaPath string
	if javaInstallation != nil {
		javaPath = javaInstallation.Path
	} else {
		// Download Java if needed
		javaDir := filepath.Join(instancesDir, "java", javaVersion)
		if err := im.javaManager.DownloadJava(javaVersion, javaDir, nil); err != nil {
			return nil, fmt.Errorf("failed to download Java: %w", err)
		}
		javaPath = filepath.Join(javaDir, "bin", "java")
	}

	// Create Prism instance
	if err := im.prismManager.CreateInstance(modpack, packInfo, instancePath, javaPath); err != nil {
		return nil, fmt.Errorf("failed to create Prism instance: %w", err)
	}

	// Create instance record
	instance := &Instance{
		ID:           instanceID,
		Name:         modpack.InstanceName,
		ModpackID:    modpack.ID,
		Minecraft:    packInfo.Minecraft,
		ModLoader:    packInfo.ModLoader,
		JavaVersion:  javaVersion,
		JavaPath:     javaPath,
		InstancePath: instancePath,
		PrismPath:    prismDir,
		Properties:   make(map[string]string),
		Config:       make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save instance metadata
	if err := im.saveInstanceMetadata(instance); err != nil {
		im.logger.Warn("Failed to save instance metadata: %v", err)
	}

	im.logger.Info("Instance %s created successfully", instanceID)
	return instance, nil
}

// GetInstance retrieves an instance by ID
func (im *InstanceManager) GetInstance(instanceID string) (*Instance, error) {
	metadataPath := im.getInstanceMetadataPath(instanceID)
	if !im.platform.FileExists(metadataPath) {
		return nil, fmt.Errorf("instance %s not found", instanceID)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read instance metadata: %w", err)
	}

	var instance Instance
	if err := json.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("failed to parse instance metadata: %w", err)
	}

	return &instance, nil
}

// ListInstances returns all instances
func (im *InstanceManager) ListInstances() ([]*Instance, error) {
	instancesDir := im.getInstancesDir()
	if !im.platform.FileExists(instancesDir) {
		return []*Instance{}, nil
	}

	entries, err := os.ReadDir(instancesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read instances directory: %w", err)
	}

	var instances []*Instance
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		instance, err := im.GetInstance(entry.Name())
		if err != nil {
			im.logger.Warn("Failed to load instance %s: %v", entry.Name(), err)
			continue
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// DeleteInstance removes an instance
func (im *InstanceManager) DeleteInstance(instanceID string) error {
	im.logger.Info("Deleting instance %s", instanceID)

	instance, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	// Remove instance directory
	if err := os.RemoveAll(instance.InstancePath); err != nil {
		return fmt.Errorf("failed to remove instance directory: %w", err)
	}

	// Remove metadata
	metadataPath := im.getInstanceMetadataPath(instanceID)
	if err := os.Remove(metadataPath); err != nil {
		im.logger.Warn("Failed to remove instance metadata: %v", err)
	}

	im.logger.Info("Instance %s deleted successfully", instanceID)
	return nil
}

// LaunchInstance launches an instance using Prism Launcher
func (im *InstanceManager) LaunchInstance(instanceID string) error {
	instance, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	// Update last played time
	now := time.Now()
	instance.LastPlayed = &now
	instance.UpdatedAt = now
	im.saveInstanceMetadata(instance)

	im.logger.Info("Launching instance %s (%s)", instance.Name, instance.ID)

	// Launch Prism
	return im.prismManager.LaunchPrism(instance.PrismPath, instance.ID, instance.InstancePath)
}

// UpdateInstance updates instance configuration
func (im *InstanceManager) UpdateInstance(instance *Instance) error {
	instance.UpdatedAt = time.Now()
	return im.saveInstanceMetadata(instance)
}

// GetInstanceForModpack returns the instance for a specific modpack
func (im *InstanceManager) GetInstanceForModpack(modpackID string) (*Instance, error) {
	instances, err := im.ListInstances()
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		if instance.ModpackID == modpackID {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("no instance found for modpack %s", modpackID)
}

// Helper methods

func (im *InstanceManager) generateInstanceID(modpackID string) string {
	// Create a slug from modpack ID and add timestamp
	slug := strings.ToLower(strings.ReplaceAll(modpackID, " ", "-"))
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", slug, timestamp)
}

func (im *InstanceManager) downloadPackInfo(packURL, instancePath string) (*PackInfo, error) {
	im.logger.Info("Downloading pack info from %s", packURL)

	// Download pack.toml
	packTomlPath := filepath.Join(instancePath, packDotTomlFile)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", packURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", im.getUserAgent("PackInfo"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download pack.toml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Save pack.toml
	file, err := os.Create(packTomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pack.toml file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write pack.toml file: %w", err)
	}

	// Parse pack.toml
	return im.parsePackToml(packTomlPath)
}

func (im *InstanceManager) parsePackToml(packTomlPath string) (*PackInfo, error) {
	data, err := os.ReadFile(packTomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pack.toml: %w", err)
	}

	var packInfo struct {
		Name     string `toml:"name"`
		Versions []struct {
			Minecraft string `toml:"minecraft"`
			Loaders   []struct {
				Loader     string `toml:"loader"`
				LoaderVersion string `toml:"loader-version"`
			} `toml:"loaders"`
		} `toml:"versions"`
	}

	if err := toml.Unmarshal(data, &packInfo); err != nil {
		return nil, fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	if len(packInfo.Versions) == 0 {
		return nil, fmt.Errorf("no versions found in pack.toml")
	}

	// Use the latest version
	latestVersion := packInfo.Versions[len(packInfo.Versions)-1]
	var modloader, loaderVersion string

	if len(latestVersion.Loaders) > 0 {
		loaderInfo := latestVersion.Loaders[0]
		modloader = loaderInfo.Loader
		loaderVersion = loaderInfo.LoaderVersion
	}

	return &PackInfo{
		Name:          packInfo.Name,
		Minecraft:     latestVersion.Minecraft,
		ModLoader:     modloader,
		LoaderVersion: loaderVersion,
	}, nil
}

func (im *InstanceManager) saveInstanceMetadata(instance *Instance) error {
	metadataPath := im.getInstanceMetadataPath(instance.ID)

	// Create metadata directory if it doesn't exist
	metadataDir := filepath.Dir(metadataPath)
	if err := im.platform.CreateDirectory(metadataDir); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal instance metadata: %w", err)
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (im *InstanceManager) getInstanceMetadataPath(instanceID string) string {
	return filepath.Join(im.getInstancesDir(), ".metadata", instanceID+".json")
}

func (im *InstanceManager) getInstancesDir() string {
	// Get app data directory and add instances subdirectory
	appDataDir, _ := im.platform.GetAppDataDir()
	return filepath.Join(appDataDir, "instances")
}

func (im *InstanceManager) getUserAgent(component string) string {
	return fmt.Sprintf("TheBoys-%s/dev", component)
}