// Package modpack provides modpack management functionality
package modpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"theboys-launcher/internal/logging"
)

// ConfigManager manages modpack configurations
type ConfigManager struct {
	logger *logging.Logger
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(logger *logging.Logger) *ConfigManager {
	return &ConfigManager{
		logger: logger,
	}
}

// ModpackConfig represents configuration options for a modpack
type ModpackConfig struct {
	ModpackID          string            `json:"modpack_id"`
	InstanceName       string            `json:"instance_name"`
	MemorySettings     MemoryConfig      `json:"memory_settings"`
	JavaSettings       JavaConfig        `json:"java_settings"`
	WindowSettings     WindowConfig      `json:"window_settings"`
	CustomArguments    ArgumentsConfig   `json:"custom_arguments"`
	ModSettings        map[string]string `json:"mod_settings,omitempty"`
	ResourcePacks      []string          `json:"resource_packs,omitempty"`
	ShaderPacks        []string          `json:"shader_packs,omitempty"`
	Optimizations      OptimizationConfig `json:"optimizations"`
	BackupSettings     BackupConfig      `json:"backup_settings"`
	AutoStart          bool              `json:"auto_start"`
	CustomDirectory    string            `json:"custom_directory,omitempty"`
}

// MemoryConfig represents memory configuration
type MemoryConfig struct {
	MinMemoryMB   int  `json:"min_memory_mb"`
	MaxMemoryMB   int  `json:"max_memory_mb"`
	AutoDetect    bool `json:"auto_detect"`
	PermGenMB     int  `json:"perm_gen_mb,omitempty"`
	UseG1GC       bool `json:"use_g1_gc"`
	UseZGC        bool `json:"use_z_gc,omitempty"`
}

// JavaConfig represents Java runtime configuration
type JavaConfig struct {
	JavaPath          string   `json:"java_path"`
	DetectJava        bool     `json:"detect_java"`
	PreferredVersion  string   `json:"preferred_version"`
	JVMArguments      []string `json:"jvm_arguments"`
	UseSystemJava     bool     `json:"use_system_java"`
	JavaArchitecture  string   `json:"java_architecture"` // "x64", "x86", "arm64"
}

// WindowConfig represents window/display configuration
type WindowConfig struct {
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Fullscreen    bool   `json:"fullscreen"`
	Borderless    bool   `json:"borderless"`
	DisplayNumber int    `json:"display_number"`
	VSync         bool   `json:"vsync"`
	FPSLimit      int    `json:"fps_limit"`
}

// ArgumentsConfig represents custom launch arguments
type ArgumentsConfig struct {
	GameArguments   []string `json:"game_arguments"`
	JVMArguments    []string `json:"jvm_arguments"`
	EnvironmentVars map[string]string `json:"environment_vars,omitempty"`
}

// OptimizationConfig represents performance optimization settings
type OptimizationConfig struct {
	EnableFOVFix          bool     `json:"enable_fov_fix"`
	EnableMemoryFix       bool     `json:"enable_memory_fix"`
	EnableThreadOptimization bool   `json:"enable_thread_optimization"`
	DisableTelemetry      bool     `json:"disable_telemetry"`
	OptimizedTextureLoading bool   `json:"optimized_texture_loading"`
	PreloadChunks        int      `json:"preload_chunks"`
	AdditionalJavaOpts   []string `json:"additional_java_opts"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	AutoBackup       bool   `json:"auto_backup"`
	BackupInterval   string `json:"backup_interval"` // "daily", "weekly", "monthly"
	MaxBackups       int    `json:"max_backups"`
	BackupLocation   string `json:"backup_location,omitempty"`
	IncludeMods      bool   `json:"include_mods"`
	IncludeConfigs   bool   `json:"include_configs"`
	IncludeSaves     bool   `json:"include_saves"`
	IncludeResourcePacks bool `json:"include_resource_packs"`
}

// LoadModpackConfig loads configuration for a modpack instance
func (cm *ConfigManager) LoadModpackConfig(instancePath string) (*ModpackConfig, error) {
	configPath := filepath.Join(instancePath, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return cm.createDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ModpackConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and fix config
	cm.validateAndFixConfig(&config)

	return &config, nil
}

// SaveModpackConfig saves configuration for a modpack instance
func (cm *ConfigManager) SaveModpackConfig(instancePath string, config *ModpackConfig) error {
	configPath := filepath.Join(instancePath, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(instancePath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Validate config
	cm.validateAndFixConfig(config)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.Info("Saved modpack config to: %s", configPath)
	return nil
}

// createDefaultConfig creates a default modpack configuration
func (cm *ConfigManager) createDefaultConfig() *ModpackConfig {
	return &ModpackConfig{
		InstanceName: "New Instance",
		MemorySettings: MemoryConfig{
			MinMemoryMB: 1024,
			MaxMemoryMB: 4096,
			AutoDetect:  true,
			UseG1GC:     true,
		},
		JavaSettings: JavaConfig{
			DetectJava:      true,
			PreferredVersion: "17",
			UseSystemJava:   true,
			JVMArguments: []string{
				"-XX:+UseG1GC",
				"-XX:+ParallelRefProcEnabled",
				"-XX:MaxGCPauseMillis=200",
				"-XX:+UnlockExperimentalVMOptions",
				"-XX:+DisableExplicitGC",
				"-XX:+AlwaysPreTouch",
				"-XX:G1NewSizePercent=30",
				"-XX:G1MaxNewSizePercent=40",
				"-XX:G1HeapRegionSize=8M",
				"-XX:G1ReservePercent=20",
				"-XX:G1HeapWastePercent=5",
				"-XX:G1MixedGCCountTarget=4",
				"-XX:InitiatingHeapOccupancyPercent=15",
				"-XX:G1MixedGCLiveThresholdPercent=90",
				"-XX:G1RSetUpdatingPauseTimePercent=5",
				"-XX:SurvivorRatio=32",
				"-XX:+PerfDisableSharedMem",
				"-XX:MaxTenuringThreshold=1",
			},
		},
		WindowSettings: WindowSettings{
			Width:      854,
			Height:     480,
			Fullscreen: false,
			DisplayNumber: 0,
			VSync:      true,
			FPSLimit:   60,
		},
		CustomArguments: ArgumentsConfig{
			GameArguments:   []string{},
			JVMArguments:    []string{},
			EnvironmentVars: make(map[string]string),
		},
		Optimizations: OptimizationConfig{
			EnableFOVFix:              true,
			EnableMemoryFix:           true,
			EnableThreadOptimization:  true,
			DisableTelemetry:          true,
			OptimizedTextureLoading:   true,
			PreloadChunks:             2,
			AdditionalJavaOpts:        []string{},
		},
		BackupSettings: BackupConfig{
			AutoBackup:              true,
			BackupInterval:          "weekly",
			MaxBackups:              5,
			IncludeMods:             true,
			IncludeConfigs:          true,
			IncludeSaves:            true,
			IncludeResourcePacks:    true,
		},
		AutoStart:       false,
		ResourcePacks:   []string{},
		ShaderPacks:     []string{},
		ModSettings:     make(map[string]string),
	}
}

// validateAndFixConfig validates and fixes configuration values
func (cm *ConfigManager) validateAndFixConfig(config *ModpackConfig) {
	// Validate memory settings
	if config.MemorySettings.MinMemoryMB < 512 {
		config.MemorySettings.MinMemoryMB = 512
		cm.logger.Warn("Minimum memory adjusted to 512MB for stability")
	}

	if config.MemorySettings.MaxMemoryMB < config.MemorySettings.MinMemoryMB {
		config.MemorySettings.MaxMemoryMB = config.MemorySettings.MinMemoryMB * 2
		cm.logger.Warn("Maximum memory adjusted to be 2x minimum memory")
	}

	if config.MemorySettings.MaxMemoryMB > 16384 { // 16GB limit
		config.MemorySettings.MaxMemoryMB = 16384
		cm.logger.Warn("Maximum memory limited to 16GB")
	}

	// Validate window settings
	if config.WindowSettings.Width < 640 {
		config.WindowSettings.Width = 640
	}

	if config.WindowSettings.Height < 480 {
		config.WindowSettings.Height = 480
	}

	// Validate FPS limit
	if config.WindowSettings.FPSLimit < 0 || config.WindowSettings.FPSLimit > 1000 {
		config.WindowSettings.FPSLimit = 60
	}

	// Validate backup settings
	if config.BackupSettings.MaxBackups < 1 {
		config.BackupSettings.MaxBackups = 1
	} else if config.BackupSettings.MaxBackups > 50 {
		config.BackupSettings.MaxBackups = 50
	}

	// Validate backup interval
	validIntervals := []string{"daily", "weekly", "monthly"}
	validInterval := false
	for _, interval := range validIntervals {
		if config.BackupSettings.BackupInterval == interval {
			validInterval = true
			break
		}
	}
	if !validInterval {
		config.BackupSettings.BackupInterval = "weekly"
	}
}

// GetRecommendedMemory gets recommended memory settings based on system
func (cm *ConfigManager) GetRecommendedMemory() (min, max int) {
	// This is a simplified implementation
	// In a real application, you would detect system memory
	// For now, assume a system with 8GB RAM
	totalMemoryMB := 8192

	// Recommended minimum: 1GB or 10% of total memory, whichever is higher
	min = 1024
	if totalMemoryMB/10 > min {
		min = totalMemoryMB / 10
	}

	// Recommended maximum: 50% of total memory, not exceeding 8GB
	max = totalMemoryMB / 2
	if max > 8192 {
		max = 8192
	}

	return min, max
}

// DetectJavaVersions detects available Java installations
func (cm *ConfigManager) DetectJavaVersions() ([]JavaVersion, error) {
	// This is a simplified implementation
	// In a real application, you would scan common Java installation paths
	// and check version information

	return []JavaVersion{
		{
			Path:    "java",
			Version: "17.0.2",
			Arch:    "x64",
			Is64Bit: true,
		},
		{
			Path:    "java",
			Version: "21.0.1",
			Arch:    "x64",
			Is64Bit: true,
		},
	}, nil
}

// JavaVersion represents a Java installation
type JavaVersion struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Is64Bit bool   `json:"is_64bit"`
}

// GenerateLaunchArguments generates launch arguments based on configuration
func (cm *ConfigManager) GenerateLaunchArguments(config *ModpackConfig, modpack *Modpack) ([]string, []string, error) {
	var jvmArgs, gameArgs []string

	// Java executable
	javaPath := config.JavaSettings.JavaPath
	if javaPath == "" {
		javaPath = "java"
	}

	// Memory arguments
	jvmArgs = append(jvmArgs, fmt.Sprintf("-Xms%dM", config.MemorySettings.MinMemoryMB))
	jvmArgs = append(jvmArgs, fmt.Sprintf("-Xmx%dM", config.MemorySettings.MaxMemoryMB))

	// PermGen/Metaspace (for older Java versions)
	if config.MemorySettings.PermGenMB > 0 {
		jvmArgs = append(jvmArgs, fmt.Sprintf("-XX:PermSize=%dM", config.MemorySettings.PermGenMB))
		jvmArgs = append(jvmArgs, fmt.Sprintf("-XX:MaxPermSize=%dM", config.MemorySettings.PermGenMB))
	}

	// GC settings
	if config.JavaSettings.UseSystemJava {
		// Add system-specific optimizations
		if config.MemorySettings.UseG1GC {
			jvmArgs = append(jvmArgs, "-XX:+UseG1GC")
		} else if config.MemorySettings.UseZGC {
			jvmArgs = append(jvmArgs, "-XX:+UseZGC")
		}
	}

	// Custom JVM arguments
	jvmArgs = append(jvmArgs, config.JavaSettings.JVMArguments...)
	jvmArgs = append(jvmArgs, config.CustomArguments.JVMArguments...)

	// Optimization arguments
	if config.Optimizations.EnableFOVFix {
		jvmArgs = append(jvmArgs, "-Dfml.ignoreInvalidMinecraftCertificates=true")
	}

	if config.Optimizations.EnableMemoryFix {
		jvmArgs = append(jvmArgs, "-XX:+UseCompressedOops")
		jvmArgs = append(jvmArgs, "-XX:+UseCompressedClassPointers")
	}

	if config.Optimizations.EnableThreadOptimization {
		jvmArgs = append(jvmArgs, "-XX:+UseConcMarkSweepGC")
		jvmArgs = append(jvmArgs, "-XX:+CMSIncrementalMode")
	}

	// Additional optimization options
	jvmArgs = append(jvmArgs, config.Optimizations.AdditionalJavaOpts...)

	// Game arguments
	gameArgs = append(gameArgs, config.CustomArguments.GameArguments...)

	// Window settings
	if config.WindowSettings.Fullscreen {
		gameArgs = append(gameArgs, "--fullscreen")
	} else {
		gameArgs = append(gameArgs, fmt.Sprintf("--width=%d", config.WindowSettings.Width))
		gameArgs = append(gameArgs, fmt.Sprintf("--height=%d", config.WindowSettings.Height))
	}

	// User and authentication (simplified - real implementation would handle Minecraft auth)
	gameArgs = append(gameArgs, "--username=Player")
	gameArgs = append(gameArgs, "--uuid=fake-uuid-for-demo")
	gameArgs = append(gameArgs, "--accessToken=fake-token-for-demo")

	// Server settings (if connecting to a server)
	// This would be configured per instance

	return jvmArgs, gameArgs, nil
}

// ExportConfig exports configuration to a file
func (cm *ConfigManager) ExportConfig(config *ModpackConfig, filePath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportConfig imports configuration from a file
func (cm *ConfigManager) ImportConfig(filePath string) (*ModpackConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ModpackConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.validateAndFixConfig(&config)
	return &config, nil
}

// ResetToDefaults resets configuration to default values
func (cm *ConfigManager) ResetToDefaults(config *ModpackConfig) {
	defaultConfig := cm.createDefaultConfig()

	// Preserve certain fields that shouldn't be reset
	modpackID := config.ModpackID
	instanceName := config.InstanceName
	customDirectory := config.CustomDirectory

	*config = *defaultConfig

	// Restore preserved fields
	config.ModpackID = modpackID
	config.InstanceName = instanceName
	config.CustomDirectory = customDirectory
}

// GetConfigSummary returns a human-readable summary of the configuration
func (cm *ConfigManager) GetConfigSummary(config *ModpackConfig) string {
	summary := fmt.Sprintf("Instance: %s\n", config.InstanceName)
	summary += fmt.Sprintf("Memory: %dMB - %dMB\n", config.MemorySettings.MinMemoryMB, config.MemorySettings.MaxMemoryMB)
	summary += fmt.Sprintf("Window: %dx%d", config.WindowSettings.Width, config.WindowSettings.Height)

	if config.WindowSettings.Fullscreen {
		summary += " (Fullscreen)"
	}

	summary += fmt.Sprintf("\nJava: %s", config.JavaSettings.PreferredVersion)

	if len(config.CustomArguments.JVMArguments) > 0 {
		summary += fmt.Sprintf("\nCustom JVM Args: %d", len(config.CustomArguments.JVMArguments))
	}

	if len(config.ResourcePacks) > 0 {
		summary += fmt.Sprintf("\nResource Packs: %d", len(config.ResourcePacks))
	}

	if len(config.ShaderPacks) > 0 {
		summary += fmt.Sprintf("\nShader Packs: %d", len(config.ShaderPacks))
	}

	return summary
}

// ValidateConfiguration validates a configuration and returns any issues
func (cm *ConfigManager) ValidateConfiguration(config *ModpackConfig) []string {
	var issues []string

	// Check memory settings
	if config.MemorySettings.MinMemoryMB < 512 {
		issues = append(issues, "Minimum memory should be at least 512MB")
	}

	if config.MemorySettings.MaxMemoryMB < config.MemorySettings.MinMemoryMB {
		issues = append(issues, "Maximum memory should be greater than minimum memory")
	}

	// Check window settings
	if config.WindowSettings.Width < 640 || config.WindowSettings.Height < 480 {
		issues = append(issues, "Window size should be at least 640x480")
	}

	// Check Java settings
	if config.JavaSettings.JavaPath != "" {
		if _, err := os.Stat(config.JavaSettings.JavaPath); os.IsNotExist(err) {
			issues = append(issues, "Java executable not found at specified path")
		}
	}

	// Check backup settings
	if config.BackupSettings.BackupLocation != "" {
		if _, err := os.Stat(config.BackupSettings.BackupLocation); os.IsNotExist(err) {
			issues = append(issues, "Backup directory does not exist")
		}
	}

	return issues
}