// Package config provides configuration management for the TheBoys Launcher
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"theboys-launcher/pkg/platform"
)

// Config holds the main application configuration
type Config struct {
	// Application settings
	WindowWidth  int    `json:"windowWidth"`
	WindowHeight int    `json:"windowHeight"`
	Theme        string `json:"theme"` // "light", "dark", "system"

	// Launcher settings
	MemoryMB         int    `json:"memoryMB"`
	JavaPath         string `json:"javaPath,omitempty"`
	PrismPath        string `json:"prismPath,omitempty"`
	InstancesPath    string `json:"instancesPath,omitempty"`
	TempPath         string `json:"tempPath,omitempty"`

	// Update settings
	AutoUpdate       bool   `json:"autoUpdate"`
	UpdateChannel    string `json:"updateChannel"` // "stable", "beta", "alpha"
	CheckUpdatesOnStartup bool `json:"checkUpdatesOnStartup"`

	// Logging settings
	LogLevel         string `json:"logLevel"`
	EnableDebugLog   bool   `json:"enableDebugLog"`

	// Network settings
	DownloadTimeout  int    `json:"downloadTimeout"` // seconds
	MaxConcurrentDownloads int `json:"maxConcurrentDownloads"`

	// User preferences
	LastModpackID    string `json:"lastModpackId,omitempty"`
	AutoLaunchOnSelect bool `json:"autoLaunchOnSelect"`
	ShowAdvancedOptions bool `json:"showAdvancedOptions"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WindowWidth:    1024,
		WindowHeight:   768,
		Theme:          "system",
		MemoryMB:       4096,
		AutoUpdate:     true,
		UpdateChannel:  "stable",
		CheckUpdatesOnStartup: true,
		LogLevel:       "info",
		EnableDebugLog: false,
		DownloadTimeout: 300,
		MaxConcurrentDownloads: 3,
		AutoLaunchOnSelect: false,
		ShowAdvancedOptions: false,
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	var configDir string

	if platform.IsWindows() {
		// Use AppData/Local on Windows
		appData, err := os.UserConfigDir()
		if err != nil {
			// Fallback to directory next to executable
			exePath, _ := os.Executable()
			return filepath.Join(filepath.Dir(exePath), "config.json")
		}
		configDir = filepath.Join(appData, "TheBoys Launcher")
	} else {
		// Use .config directory on Unix-like systems
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to user config dir
			configDir, _ = os.UserConfigDir()
		} else {
			configDir = filepath.Join(homeDir, ".config", "theboys")
		}
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		// Fallback to current directory
		return "config.json"
	}

	return filepath.Join(configDir, "config.json")
}

// Load loads configuration from file
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing fields
	config.applyDefaults()

	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Validate config before saving
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MemoryMB < 512 || c.MemoryMB > 32768 {
		return fmt.Errorf("memory must be between 512MB and 32GB")
	}

	if c.WindowWidth < 640 || c.WindowHeight < 480 {
		return fmt.Errorf("window dimensions too small (minimum 640x480)")
	}

	if c.DownloadTimeout < 10 || c.DownloadTimeout > 3600 {
		return fmt.Errorf("download timeout must be between 10 and 3600 seconds")
	}

	if c.MaxConcurrentDownloads < 1 || c.MaxConcurrentDownloads > 10 {
		return fmt.Errorf("max concurrent downloads must be between 1 and 10")
	}

	validThemes := map[string]bool{"light": true, "dark": true, "system": true}
	if !validThemes[c.Theme] {
		return fmt.Errorf("invalid theme: %s", c.Theme)
	}

	validChannels := map[string]bool{"stable": true, "beta": true, "alpha": true}
	if !validChannels[c.UpdateChannel] {
		return fmt.Errorf("invalid update channel: %s", c.UpdateChannel)
	}

	return nil
}

// applyDefaults applies default values to unset fields
func (c *Config) applyDefaults() {
	defaults := DefaultConfig()

	if c.WindowWidth == 0 {
		c.WindowWidth = defaults.WindowWidth
	}
	if c.WindowHeight == 0 {
		c.WindowHeight = defaults.WindowHeight
	}
	if c.Theme == "" {
		c.Theme = defaults.Theme
	}
	if c.MemoryMB == 0 {
		c.MemoryMB = defaults.MemoryMB
	}
	if c.DownloadTimeout == 0 {
		c.DownloadTimeout = defaults.DownloadTimeout
	}
	if c.MaxConcurrentDownloads == 0 {
		c.MaxConcurrentDownloads = defaults.MaxConcurrentDownloads
	}
	if c.LogLevel == "" {
		c.LogLevel = defaults.LogLevel
	}
	if c.UpdateChannel == "" {
		c.UpdateChannel = defaults.UpdateChannel
	}
}

// GetInstancesPath returns the path where Minecraft instances should be stored
func (c *Config) GetInstancesPath() string {
	if c.InstancesPath != "" {
		return c.InstancesPath
	}

	// Default instance path based on platform
	if platform.IsWindows() {
		appData, _ := os.UserConfigDir()
		return filepath.Join(appData, "TheBoys Launcher", "instances")
	} else if platform.IsMacOS() {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "theboys", "instances")
	} else {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".local", "share", "theboys", "instances")
	}
}

// GetTempPath returns the path for temporary files
func (c *Config) GetTempPath() string {
	if c.TempPath != "" {
		return c.TempPath
	}

	tempDir := os.TempDir()
	return filepath.Join(tempDir, "theboys")
}