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

// ModpackManager handles modpack configuration
type ModpackManager struct {
	platform    platform.Platform
	logger      logging.Logger
	mu          sync.RWMutex
	modpacks    []types.Modpack
	configPath  string
	configURL   string
	loaded      bool
}

const (
	defaultModpacksURL = "https://raw.githubusercontent.com/dilllxd/theboys-launcher/refs/heads/main/modpacks.json"
	localConfigFile   = "modpacks.json"
)

// NewModpackManager creates a new modpack manager
func NewModpackManager(platform platform.Platform, logger logging.Logger) *ModpackManager {
	return &ModpackManager{
		platform:   platform,
		logger:     logger,
		configURL:  defaultModpacksURL,
		modpacks:   []types.Modpack{},
	}
}

// LoadModpacks loads modpack configuration from local file
func (m *ModpackManager) LoadModpacks() ([]types.Modpack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First try to load from application config directory
	appDataDir, err := m.platform.GetAppDataDir()
	if err != nil {
		m.logger.Warn("Could not get app data directory: %v", err)
	} else {
		configPath := filepath.Join(appDataDir, localConfigFile)
		if m.platform.FileExists(configPath) {
			if modpacks, err := m.loadFromFile(configPath); err == nil {
				m.modpacks = modpacks
				m.loaded = true
				m.configPath = configPath
				m.logger.Info("Loaded %d modpacks from config directory", len(modpacks))
				return modpacks, nil
			} else {
				m.logger.Warn("Failed to load modpacks from config directory: %v", err)
			}
		}
	}

	// Try to load from working directory (for development/portability)
	if m.platform.FileExists(localConfigFile) {
		if modpacks, err := m.loadFromFile(localConfigFile); err == nil {
			m.modpacks = modpacks
			m.loaded = true
			m.configPath = localConfigFile
			m.logger.Info("Loaded %d modpacks from working directory", len(modpacks))
			return modpacks, nil
		} else {
			m.logger.Warn("Failed to load modpacks from working directory: %v", err)
		}
	}

	// If no local config found, create a default one
	m.logger.Info("No modpack configuration found, creating default")
	return m.createDefaultModpacks()
}

// SaveModpacks saves the current modpack configuration
func (m *ModpackManager) SaveModpacks() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded || len(m.modpacks) == 0 {
		return fmt.Errorf("no modpacks to save")
	}

	// Determine where to save the file
	var savePath string
	if m.configPath != "" {
		savePath = m.configPath
	} else {
		// Save to application data directory
		appDataDir, err := m.platform.GetAppDataDir()
		if err != nil {
			return fmt.Errorf("failed to get app data directory: %w", err)
		}
		savePath = filepath.Join(appDataDir, localConfigFile)
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(m.modpacks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize modpacks: %w", err)
	}

	// Write to file
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write modpacks file: %w", err)
	}

	m.configPath = savePath
	m.logger.Info("Saved %d modpacks to %s", len(m.modpacks), savePath)
	return nil
}

// GetModpacks returns the loaded modpacks
func (m *ModpackManager) GetModpacks() []types.Modpack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	modpacksCopy := make([]types.Modpack, len(m.modpacks))
	copy(modpacksCopy, m.modpacks)
	return modpacksCopy
}

// GetModpackByID returns a modpack by its ID
func (m *ModpackManager) GetModpackByID(id string) (*types.Modpack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, modpack := range m.modpacks {
		if modpack.ID == id {
			// Return a copy
			modpackCopy := modpack
			return &modpackCopy, nil
		}
	}

	return nil, fmt.Errorf("modpack with ID '%s' not found", id)
}

// GetDefaultModpack returns the default modpack
func (m *ModpackManager) GetDefaultModpack() (*types.Modpack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// First try to find a modpack marked as default
	for _, modpack := range m.modpacks {
		if modpack.Default {
			modpackCopy := modpack
			return &modpackCopy, nil
		}
	}

	// If no default is set, return the first modpack
	if len(m.modpacks) > 0 {
		modpackCopy := m.modpacks[0]
		return &modpackCopy, nil
	}

	return nil, fmt.Errorf("no modpacks available")
}

// AddModpack adds a new modpack to the configuration
func (m *ModpackManager) AddModpack(modpack types.Modpack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if modpack with same ID already exists
	for _, existing := range m.modpacks {
		if existing.ID == modpack.ID {
			return fmt.Errorf("modpack with ID '%s' already exists", modpack.ID)
		}
	}

	// Validate modpack
	if err := m.validateModpack(&modpack); err != nil {
		return err
	}

	m.modpacks = append(m.modpacks, modpack)
	m.logger.Info("Added modpack: %s", modpack.DisplayName)

	return nil
}

// RemoveModpack removes a modpack by ID
func (m *ModpackManager) RemoveModpack(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, modpack := range m.modpacks {
		if modpack.ID == id {
			m.modpacks = append(m.modpacks[:i], m.modpacks[i+1:]...)
			m.logger.Info("Removed modpack: %s", modpack.DisplayName)
			return nil
		}
	}

	return fmt.Errorf("modpack with ID '%s' not found", id)
}

// UpdateModpack updates an existing modpack
func (m *ModpackManager) UpdateModpack(modpack types.Modpack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate modpack
	if err := m.validateModpack(&modpack); err != nil {
		return err
	}

	for i, existing := range m.modpacks {
		if existing.ID == modpack.ID {
			m.modpacks[i] = modpack
			m.logger.Info("Updated modpack: %s", modpack.DisplayName)
			return nil
		}
	}

	return fmt.Errorf("modpack with ID '%s' not found", modpack.ID)
}

// loadFromFile loads modpacks from a JSON file
func (m *ModpackManager) loadFromFile(filePath string) ([]types.Modpack, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read modpacks file: %w", err)
	}

	var modpacks []types.Modpack
	if err := json.Unmarshal(data, &modpacks); err != nil {
		return nil, fmt.Errorf("failed to parse modpacks file: %w", err)
	}

	// Validate loaded modpacks
	validModpacks := []types.Modpack{}
	for i, modpack := range modpacks {
		if err := m.validateModpack(&modpack); err != nil {
			m.logger.Warn("Skipping invalid modpack at index %d: %v", i, err)
			continue
		}
		validModpacks = append(validModpacks, modpack)
	}

	if len(validModpacks) == 0 {
		return nil, fmt.Errorf("no valid modpacks found in file")
	}

	return validModpacks, nil
}

// createDefaultModpacks creates a default modpack configuration
func (m *ModpackManager) createDefaultModpacks() ([]types.Modpack, error) {
	defaultModpacks := []types.Modpack{
		{
			ID:           "winterpack",
			DisplayName:  "WinterPack",
			PackURL:      "https://raw.githubusercontent.com/dilllxd/winterpack-modpack/main/pack.toml",
			InstanceName: "WinterPack",
			Description:  "Official WinterPack release.",
			Default:      true,
		},
		{
			ID:           "examplepack",
			DisplayName:  "examplepack (Dev Preview)",
			PackURL:      "https://raw.githubusercontent.com/dilllxd/examplepack-modpack/main/pack.toml",
			InstanceName: "examplepack",
			Description:  "Development preview slot for staging new modpacks.",
		},
	}

	m.modpacks = defaultModpacks
	m.loaded = true

	// Save the default configuration
	if err := m.SaveModpacks(); err != nil {
		m.logger.Warn("Failed to save default modpack configuration: %v", err)
	}

	m.logger.Info("Created default modpack configuration with %d modpacks", len(defaultModpacks))
	return defaultModpacks, nil
}

// validateModpack validates a modpack configuration
func (m *ModpackManager) validateModpack(modpack *types.Modpack) error {
	if modpack.ID == "" {
		return fmt.Errorf("modpack ID cannot be empty")
	}
	if modpack.DisplayName == "" {
		return fmt.Errorf("modpack display name cannot be empty")
	}
	if modpack.PackURL == "" {
		return fmt.Errorf("modpack pack URL cannot be empty")
	}
	if modpack.InstanceName == "" {
		return fmt.Errorf("modpack instance name cannot be empty")
	}

	return nil
}

// IsLoaded returns whether modpacks have been loaded
func (m *ModpackManager) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loaded
}

// GetCount returns the number of loaded modpacks
func (m *ModpackManager) GetCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.modpacks)
}