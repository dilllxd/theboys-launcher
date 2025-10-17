package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// Manager handles application configuration
type Manager struct {
	platform   platform.Platform
	logger     logging.Logger
	mu         sync.RWMutex
	settings   *types.LauncherSettings
	configPath string
	settingsLoaded bool
}

// NewManager creates a new configuration manager
func NewManager(platform platform.Platform, logger logging.Logger) *Manager {
	return &Manager{
		platform: platform,
		logger:   logger,
		settings: &types.LauncherSettings{
			MemoryMB:        platform.GetDefaultRAM(),
			AutoUpdate:      true,
			CheckForUpdates: true,
			KeepConsoleOpen: false,
			Theme:           "dark",
			WindowSize: types.Size{
				Width:  1200,
				Height: 800,
			},
		},
	}
}

// Initialize sets up the configuration manager
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get application data directory
	appDataDir, err := m.platform.GetAppDataDir()
	if err != nil {
		return fmt.Errorf("failed to get app data directory: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(appDataDir, "config")
	if err := m.platform.CreateDirectory(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config file path
	m.configPath = filepath.Join(configDir, "settings.json")

	m.logger.Debug("Configuration manager initialized. Config path: %s", m.configPath)
	return nil
}

// LoadSettings loads the settings from file
func (m *Manager) LoadSettings() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.configPath == "" {
		return fmt.Errorf("configuration not initialized")
	}

	// Check if config file exists
	if !m.platform.FileExists(m.configPath) {
		m.logger.Info("Settings file not found, using defaults")
		return nil
	}

	// Read settings file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse JSON
	var settings types.LauncherSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		m.logger.Warn("Failed to parse settings file, using defaults: %v", err)
		return nil
	}

	// Validate settings
	m.validateSettings(&settings)

	m.settings = &settings
	m.settingsLoaded = true

	m.logger.Info("Settings loaded successfully")
	return nil
}

// SaveSettings saves the current settings to file
func (m *Manager) SaveSettings() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.configPath == "" {
		return fmt.Errorf("configuration not initialized")
	}

	// Validate settings before saving
	m.validateSettings(m.settings)

	// Serialize to JSON
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	m.logger.Debug("Settings saved successfully")
	return nil
}

// GetSettings returns the current settings
func (m *Manager) GetSettings() *types.LauncherSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	settingsCopy := *m.settings
	return &settingsCopy
}

// UpdateSettings updates the current settings
func (m *Manager) UpdateSettings(settings *types.LauncherSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate new settings
	m.validateSettings(settings)

	m.settings = settings
	m.logger.Info("Settings updated")

	return nil
}

// SetMemoryMB sets the memory allocation in MB
func (m *Manager) SetMemoryMB(memoryMB int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate memory range
	maxRAM := m.platform.GetMaxRAM()
	if memoryMB < 1024 || memoryMB > maxRAM {
		return fmt.Errorf("memory must be between 1024MB and %dMB", maxRAM)
	}

	m.settings.MemoryMB = memoryMB
	m.logger.Debug("Memory allocation set to %dMB", memoryMB)
	return nil
}

// SetAutoUpdate sets whether auto-update is enabled
func (m *Manager) SetAutoUpdate(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.AutoUpdate = enabled
	m.logger.Debug("Auto-update set to %t", enabled)
}

// SetTheme sets the UI theme
func (m *Manager) SetTheme(theme string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	validThemes := []string{"light", "dark", "auto"}
	for _, validTheme := range validThemes {
		if theme == validTheme {
			m.settings.Theme = theme
			m.logger.Debug("Theme set to %s", theme)
			return nil
		}
	}

	return fmt.Errorf("invalid theme: %s. Valid themes: %v", theme, validThemes)
}

// SetLastModpackID sets the last selected modpack ID
func (m *Manager) SetLastModpackID(modpackID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.LastModpackID = modpackID
	m.logger.Debug("Last modpack ID set to %s", modpackID)
}

// SetWindowSize sets the window dimensions
func (m *Manager) SetWindowSize(width, height int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.WindowSize = types.Size{
		Width:  width,
		Height: height,
	}
	m.logger.Debug("Window size set to %dx%d", width, height)
}

// validateSettings validates and corrects settings values
func (m *Manager) validateSettings(settings *types.LauncherSettings) {
	// Validate memory allocation
	if settings.MemoryMB < 1024 {
		settings.MemoryMB = 1024
		m.logger.Debug("Memory allocation too low, set to minimum 1024MB")
	} else if settings.MemoryMB > m.platform.GetMaxRAM() {
		settings.MemoryMB = m.platform.GetMaxRAM()
		m.logger.Debug("Memory allocation too high, set to maximum %dMB", m.platform.GetMaxRAM())
	}

	// Validate theme
	validThemes := []string{"light", "dark", "auto"}
	themeValid := false
	for _, validTheme := range validThemes {
		if settings.Theme == validTheme {
			themeValid = true
			break
		}
	}
	if !themeValid {
		settings.Theme = "dark"
		m.logger.Debug("Invalid theme, set to default 'dark'")
	}

	// Validate window size
	if settings.WindowSize.Width < 800 {
		settings.WindowSize.Width = 1200
	}
	if settings.WindowSize.Height < 600 {
		settings.WindowSize.Height = 800
	}
}

// IsSettingsLoaded returns whether settings have been loaded
func (m *Manager) IsSettingsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settingsLoaded
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configPath
}