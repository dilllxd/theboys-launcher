package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"theboys-launcher/internal/logging"
)

// SettingsManager manages configuration with real-time updates and persistence
type SettingsManager struct {
	config     *Config
	validator  *Validator
	logger     *logging.Logger
	listeners  []SettingsChangeListener
	autoSave   bool
	lastSave   time.Time
}

// SettingsChangeListener is called when settings change
type SettingsChangeListener interface {
	OnSettingsChanged(config *Config, changes map[string]interface{})
}

// SettingsChangeFunc is a function adapter for SettingsChangeListener
type SettingsChangeFunc func(config *Config, changes map[string]interface{})

func (f SettingsChangeFunc) OnSettingsChanged(config *Config, changes map[string]interface{}) {
	f(config, changes)
}

// NewSettingsManager creates a new settings manager
func NewSettingsManager(logger *logging.Logger) (*SettingsManager, error) {
	// Load configuration
	config, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	manager := &SettingsManager{
		config:    config,
		validator: NewValidator(),
		logger:    logger,
		listeners: []SettingsChangeListener{},
		autoSave:  true,
		lastSave:  time.Now(),
	}

	return manager, nil
}

// GetConfig returns the current configuration
func (sm *SettingsManager) GetConfig() *Config {
	return sm.config
}

// UpdateConfig updates the configuration with validation
func (sm *SettingsManager) UpdateConfig(updates map[string]interface{}) error {
	// Create a copy of current config for rollback
	oldConfig := sm.cloneConfig()

	// Apply updates
	changes := make(map[string]interface{})
	for key, value := range updates {
		if err := sm.applyUpdate(sm.config, key, value); err != nil {
			// Rollback on error
			sm.config = oldConfig
			return fmt.Errorf("failed to apply update %s: %w", key, err)
		}
		changes[key] = value
	}

	// Validate updated configuration
	if validation := sm.validator.ValidateConfig(sm.config); !validation.Valid {
		// Rollback on validation failure
		sm.config = oldConfig
		return fmt.Errorf("validation failed: %s", validation.Errors[0])
	}

	// Notify listeners
	sm.notifyListeners(changes)

	// Auto-save if enabled
	if sm.autoSave {
		if err := sm.Save(); err != nil {
			sm.logger.Error("Failed to auto-save configuration: %v", err)
		}
	}

	sm.logger.Info("Configuration updated: %v", changes)
	return nil
}

// Save saves the configuration to disk
func (sm *SettingsManager) Save() error {
	if err := sm.config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	sm.lastSave = time.Now()
	sm.logger.Debug("Configuration saved to disk")
	return nil
}

// ResetToDefaults resets all settings to defaults
func (sm *SettingsManager) ResetToDefaults() error {
	defaultConfig := DefaultConfig()

	// Preserve some settings that shouldn't be reset
	preserveFields := []string{"WindowWidth", "WindowHeight", "LastModpackID"}
	for _, field := range preserveFields {
		switch field {
		case "WindowWidth":
			defaultConfig.WindowWidth = sm.config.WindowWidth
		case "WindowHeight":
			defaultConfig.WindowHeight = sm.config.WindowHeight
		case "LastModpackID":
			defaultConfig.LastModpackID = sm.config.LastModpackID
		}
	}

	sm.config = defaultConfig

	// Notify listeners of reset
	sm.notifyListeners(map[string]interface{}{
		"reset": true,
	})

	// Save the reset configuration
	if err := sm.Save(); err != nil {
		return fmt.Errorf("failed to save reset configuration: %w", err)
	}

	sm.logger.Info("Configuration reset to defaults")
	return nil
}

// ExportSettings exports settings to a file
func (sm *SettingsManager) ExportSettings(filePath string) error {
	data, err := json.MarshalIndent(sm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	sm.logger.Info("Settings exported to: %s", filePath)
	return nil
}

// ImportSettings imports settings from a file
func (sm *SettingsManager) ImportSettings(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var importedConfig Config
	if err := json.Unmarshal(data, &importedConfig); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	// Validate imported configuration
	if validation := sm.validator.ValidateConfig(&importedConfig); !validation.Valid {
		return fmt.Errorf("imported configuration is invalid: %s", validation.Errors[0])
	}

	// Apply imported settings
	sm.config = &importedConfig

	// Notify listeners
	sm.notifyListeners(map[string]interface{}{
		"imported": true,
	})

	// Save imported configuration
	if err := sm.Save(); err != nil {
		return fmt.Errorf("failed to save imported configuration: %w", err)
	}

	sm.logger.Info("Settings imported from: %s", filePath)
	return nil
}

// AddListener adds a settings change listener
func (sm *SettingsManager) AddListener(listener SettingsChangeListener) {
	sm.listeners = append(sm.listeners, listener)
}

// RemoveListener removes a settings change listener
func (sm *SettingsManager) RemoveListener(listener SettingsChangeListener) {
	for i, l := range sm.listeners {
		if l == listener {
			sm.listeners = append(sm.listeners[:i], sm.listeners[i+1:]...)
			break
		}
	}
}

// SetAutoSave enables or disables auto-save
func (sm *SettingsManager) SetAutoSave(autoSave bool) {
	sm.autoSave = autoSave
	sm.logger.Debug("Auto-save set to: %v", autoSave)
}

// GetLastSaveTime returns the last save time
func (sm *SettingsManager) GetLastSaveTime() time.Time {
	return sm.lastSave
}

// ValidateCurrentSettings validates the current configuration
func (sm *SettingsManager) ValidateCurrentSettings() *ValidationResult {
	return sm.validator.ValidateConfig(sm.config)
}

// applyUpdate applies a single update to the configuration
func (sm *SettingsManager) applyUpdate(config *Config, key string, value interface{}) error {
	switch key {
	case "WindowWidth":
		if v, ok := value.(float64); ok {
			config.WindowWidth = int(v)
		} else if v, ok := value.(int); ok {
			config.WindowWidth = v
		} else {
			return fmt.Errorf("invalid type for WindowWidth")
		}
	case "WindowHeight":
		if v, ok := value.(float64); ok {
			config.WindowHeight = int(v)
		} else if v, ok := value.(int); ok {
			config.WindowHeight = v
		} else {
			return fmt.Errorf("invalid type for WindowHeight")
		}
	case "Theme":
		if v, ok := value.(string); ok {
			config.Theme = v
		} else {
			return fmt.Errorf("invalid type for Theme")
		}
	case "MemoryMB":
		if v, ok := value.(float64); ok {
			config.MemoryMB = int(v)
		} else if v, ok := value.(int); ok {
			config.MemoryMB = v
		} else {
			return fmt.Errorf("invalid type for MemoryMB")
		}
	case "JavaPath":
		if v, ok := value.(string); ok {
			config.JavaPath = v
		} else {
			return fmt.Errorf("invalid type for JavaPath")
		}
	case "PrismPath":
		if v, ok := value.(string); ok {
			config.PrismPath = v
		} else {
			return fmt.Errorf("invalid type for PrismPath")
		}
	case "InstancesPath":
		if v, ok := value.(string); ok {
			config.InstancesPath = v
		} else {
			return fmt.Errorf("invalid type for InstancesPath")
		}
	case "TempPath":
		if v, ok := value.(string); ok {
			config.TempPath = v
		} else {
			return fmt.Errorf("invalid type for TempPath")
		}
	case "AutoUpdate":
		if v, ok := value.(bool); ok {
			config.AutoUpdate = v
		} else {
			return fmt.Errorf("invalid type for AutoUpdate")
		}
	case "UpdateChannel":
		if v, ok := value.(string); ok {
			config.UpdateChannel = v
		} else {
			return fmt.Errorf("invalid type for UpdateChannel")
		}
	case "CheckUpdatesOnStartup":
		if v, ok := value.(bool); ok {
			config.CheckUpdatesOnStartup = v
		} else {
			return fmt.Errorf("invalid type for CheckUpdatesOnStartup")
		}
	case "LogLevel":
		if v, ok := value.(string); ok {
			config.LogLevel = v
		} else {
			return fmt.Errorf("invalid type for LogLevel")
		}
	case "EnableDebugLog":
		if v, ok := value.(bool); ok {
			config.EnableDebugLog = v
		} else {
			return fmt.Errorf("invalid type for EnableDebugLog")
		}
	case "DownloadTimeout":
		if v, ok := value.(float64); ok {
			config.DownloadTimeout = int(v)
		} else if v, ok := value.(int); ok {
			config.DownloadTimeout = v
		} else {
			return fmt.Errorf("invalid type for DownloadTimeout")
		}
	case "MaxConcurrentDownloads":
		if v, ok := value.(float64); ok {
			config.MaxConcurrentDownloads = int(v)
		} else if v, ok := value.(int); ok {
			config.MaxConcurrentDownloads = v
		} else {
			return fmt.Errorf("invalid type for MaxConcurrentDownloads")
		}
	case "AutoLaunchOnSelect":
		if v, ok := value.(bool); ok {
			config.AutoLaunchOnSelect = v
		} else {
			return fmt.Errorf("invalid type for AutoLaunchOnSelect")
		}
	case "ShowAdvancedOptions":
		if v, ok := value.(bool); ok {
			config.ShowAdvancedOptions = v
		} else {
			return fmt.Errorf("invalid type for ShowAdvancedOptions")
		}
	default:
		return fmt.Errorf("unknown setting key: %s", key)
	}

	return nil
}

// notifyListeners notifies all listeners of settings changes
func (sm *SettingsManager) notifyListeners(changes map[string]interface{}) {
	for _, listener := range sm.listeners {
		listener.OnSettingsChanged(sm.config, changes)
	}
}

// cloneConfig creates a deep copy of the configuration
func (sm *SettingsManager) cloneConfig() *Config {
	data, _ := json.Marshal(sm.config)
	var clone Config
	json.Unmarshal(data, &clone)
	return &clone
}