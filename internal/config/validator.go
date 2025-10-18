package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"theboys-launcher/pkg/platform"
)

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// Validator handles configuration validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConfig performs comprehensive configuration validation
func (v *Validator) ValidateConfig(config *Config) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Validate window settings
	v.validateWindowSettings(config, result)

	// Validate theme settings
	v.validateThemeSettings(config, result)

	// Validate launcher settings
	v.validateLauncherSettings(config, result)

	// Validate update settings
	v.validateUpdateSettings(config, result)

	// Validate logging settings
	v.validateLoggingSettings(config, result)

	// Validate network settings
	v.validateNetworkSettings(config, result)

	// Validate paths
	v.validatePaths(config, result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// validateWindowSettings validates window-related settings
func (v *Validator) validateWindowSettings(config *Config, result *ValidationResult) {
	// Window dimensions
	if config.WindowWidth < 640 {
		result.Errors = append(result.Errors, "Window width must be at least 640 pixels")
		result.Valid = false
	}
	if config.WindowWidth > 7680 {
		result.Errors = append(result.Errors, "Window width cannot exceed 7680 pixels")
		result.Valid = false
	}

	if config.WindowHeight < 480 {
		result.Errors = append(result.Errors, "Window height must be at least 480 pixels")
		result.Valid = false
	}
	if config.WindowHeight > 4320 {
		result.Errors = append(result.Errors, "Window height cannot exceed 4320 pixels")
		result.Valid = false
	}
}

// validateThemeSettings validates theme-related settings
func (v *Validator) validateThemeSettings(config *Config, result *ValidationResult) {
	validThemes := map[string]bool{
		"light":  true,
		"dark":   true,
		"system": true,
	}

	if !validThemes[config.Theme] {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid theme: %s. Must be one of: light, dark, system", config.Theme))
		result.Valid = false
	}
}

// validateLauncherSettings validates launcher-related settings
func (v *Validator) validateLauncherSettings(config *Config, result *ValidationResult) {
	// Memory allocation
	if config.MemoryMB < 512 {
		result.Errors = append(result.Errors, "Memory allocation must be at least 512MB")
		result.Valid = false
	}
	if config.MemoryMB > 32768 {
		result.Errors = append(result.Errors, "Memory allocation cannot exceed 32GB (32768MB)")
		result.Valid = false
	}

	// Java path validation
	if config.JavaPath != "" {
		if err := v.validateJavaPath(config.JavaPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Java path validation failed: %v", err))
			result.Valid = false
		}
	}

	// Prism path validation
	if config.PrismPath != "" {
		if err := v.validatePrismPath(config.PrismPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Prism path validation failed: %v", err))
			result.Valid = false
		}
	}
}

// validateUpdateSettings validates update-related settings
func (v *Validator) validateUpdateSettings(config *Config, result *ValidationResult) {
	validChannels := map[string]bool{
		"stable": true,
		"beta":   true,
		"alpha":  true,
	}

	if !validChannels[config.UpdateChannel] {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid update channel: %s. Must be one of: stable, beta, alpha", config.UpdateChannel))
		result.Valid = false
	}
}

// validateLoggingSettings validates logging-related settings
func (v *Validator) validateLoggingSettings(config *Config, result *ValidationResult) {
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.LogLevel] {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid log level: %s. Must be one of: debug, info, warn, error", config.LogLevel))
		result.Valid = false
	}
}

// validateNetworkSettings validates network-related settings
func (v *Validator) validateNetworkSettings(config *Config, result *ValidationResult) {
	// Download timeout
	if config.DownloadTimeout < 10 {
		result.Errors = append(result.Errors, "Download timeout must be at least 10 seconds")
		result.Valid = false
	}
	if config.DownloadTimeout > 3600 {
		result.Errors = append(result.Errors, "Download timeout cannot exceed 3600 seconds (1 hour)")
		result.Valid = false
	}

	// Concurrent downloads
	if config.MaxConcurrentDownloads < 1 {
		result.Errors = append(result.Errors, "Maximum concurrent downloads must be at least 1")
		result.Valid = false
	}
	if config.MaxConcurrentDownloads > 10 {
		result.Errors = append(result.Errors, "Maximum concurrent downloads cannot exceed 10")
		result.Valid = false
	}
}

// validatePaths validates path-related settings
func (v *Validator) validatePaths(config *Config, result *ValidationResult) {
	// Instances path
	if config.InstancesPath != "" {
		if err := v.validateDirectoryPath(config.InstancesPath, "instances", true); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Instances path validation failed: %v", err))
			result.Valid = false
		}
	}

	// Temp path
	if config.TempPath != "" {
		if err := v.validateDirectoryPath(config.TempPath, "temp", true); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Temp path validation failed: %v", err))
			result.Valid = false
		}
	}
}

// validateJavaPath validates the Java executable path
func (v *Validator) validateJavaPath(javaPath string) error {
	// Check if file exists
	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return fmt.Errorf("Java executable not found at: %s", javaPath)
	}

	// Check if it's executable
	if !v.isExecutable(javaPath) {
		return fmt.Errorf("File is not executable: %s", javaPath)
	}

	// TODO: Add more specific Java validation if needed
	return nil
}

// validatePrismPath validates the Prism launcher path
func (v *Validator) validatePrismPath(prismPath string) error {
	// Check if file exists
	if _, err := os.Stat(prismPath); os.IsNotExist(err) {
		return fmt.Errorf("Prism launcher not found at: %s", prismPath)
	}

	// Check if it's executable
	if !v.isExecutable(prismPath) {
		return fmt.Errorf("File is not executable: %s", prismPath)
	}

	return nil
}

// validateDirectoryPath validates a directory path
func (v *Validator) validateDirectoryPath(path, pathType string, createIfMissing bool) error {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if createIfMissing {
			// Try to create the directory
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("Failed to create %s directory: %v", pathType, err)
			}
			return nil
		}
		return fmt.Errorf("%s directory does not exist: %s", pathType, path)
	}

	if err != nil {
		return fmt.Errorf("Failed to access %s directory: %v", pathType, err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("Path is not a directory: %s", path)
	}

	// Check if directory is writable
	testFile := filepath.Join(path, ".theboys_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("%s directory is not writable: %v", pathType, err)
	}
	os.Remove(testFile)

	return nil
}

// isExecutable checks if a file is executable
func (v *Validator) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check executable bits based on platform
	if platform.IsWindows() {
		// On Windows, check file extension
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".exe"
	}

	// On Unix-like systems, check executable permission
	return info.Mode().Perm()&0111 != 0
}

// ValidateMemoryMB validates a memory value in MB
func (v *Validator) ValidateMemoryMB(memoryMB int) error {
	if memoryMB < 512 {
		return fmt.Errorf("memory must be at least 512MB")
	}
	if memoryMB > 32768 {
		return fmt.Errorf("memory cannot exceed 32GB (32768MB)")
	}

	// Check if it's a reasonable increment (e.g., multiple of 256MB)
	if memoryMB%256 != 0 {
		return fmt.Errorf("memory should be in increments of 256MB for optimal performance")
	}

	return nil
}

// ValidateTimeout validates a timeout value in seconds
func (v *Validator) ValidateTimeout(timeout int) error {
	if timeout < 10 {
		return fmt.Errorf("timeout must be at least 10 seconds")
	}
	if timeout > 3600 {
		return fmt.Errorf("timeout cannot exceed 3600 seconds (1 hour)")
	}
	return nil
}

// ParseMemoryString parses a memory string (e.g., "4GB", "2048MB") into MB
func (v *Validator) ParseMemoryString(memoryStr string) (int, error) {
	memoryStr = strings.ToUpper(strings.TrimSpace(memoryStr))

	if strings.HasSuffix(memoryStr, "GB") {
		gbStr := strings.TrimSuffix(memoryStr, "GB")
		gb, err := strconv.Atoi(gbStr)
		if err != nil {
			return 0, fmt.Errorf("invalid GB value: %s", gbStr)
		}
		return gb * 1024, nil
	}

	if strings.HasSuffix(memoryStr, "MB") {
		mbStr := strings.TrimSuffix(memoryStr, "MB")
		mb, err := strconv.Atoi(mbStr)
		if err != nil {
			return 0, fmt.Errorf("invalid MB value: %s", mbStr)
		}
		return mb, nil
	}

	// Assume MB if no unit specified
	mb, err := strconv.Atoi(memoryStr)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value: %s", memoryStr)
	}
	return mb, nil
}