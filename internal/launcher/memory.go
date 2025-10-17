package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// MemoryInfo represents system memory information
type MemoryInfo struct {
	TotalMB   int
	AvailableMB int
	UsageMB   int
}

// MemoryManager handles memory detection and configuration
type MemoryManager struct {
	platform platform.Platform
	logger   logging.Logger
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(platform platform.Platform, logger logging.Logger) *MemoryManager {
	return &MemoryManager{
		platform: platform,
		logger:   logger,
	}
}

// GetSystemMemory returns system memory information
func (mm *MemoryManager) GetSystemMemory() (*MemoryInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return mm.getWindowsMemory()
	case "linux":
		return mm.getLinuxMemory()
	case "darwin":
		return mm.getDarwinMemory()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getWindowsMemory gets memory information on Windows
func (mm *MemoryManager) getWindowsMemory() (*MemoryInfo, error) {
	// Use PowerShell to get memory information
	cmd := exec.Command("powershell", "-Command",
		"Get-CimInstance -ClassName Win32_OperatingSystem | Select-Object TotalVisibleMemorySize, FreePhysicalMemory | ConvertTo-Json")

	output, err := cmd.Output()
	if err != nil {
		mm.logger.Warn("Failed to get Windows memory info via PowerShell: %v", err)
		return mm.getWindowsMemoryFallback()
	}

	// Parse PowerShell output
	outputStr := strings.TrimSpace(string(output))
	if !strings.HasPrefix(outputStr, "{") {
		mm.logger.Warn("Unexpected PowerShell output format: %s", outputStr)
		return mm.getWindowsMemoryFallback()
	}

	// Simple parsing for JSON output (without importing encoding/json to avoid dependencies)
	lines := strings.Split(outputStr, "\n")
	var totalMB, freeMB int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "TotalVisibleMemorySize") {
			// Extract number from "TotalVisibleMemorySize": 16384
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				totalStr := strings.TrimSpace(strings.Trim(parts[1], ", "))
				if total, err := strconv.Atoi(totalStr); err == nil {
					// Convert from KB to MB
					totalMB = total / 1024
				}
			}
		} else if strings.Contains(line, "FreePhysicalMemory") {
			// Extract number from "FreePhysicalMemory": 8192
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				freeStr := strings.TrimSpace(strings.Trim(parts[1], ", "))
				if free, err := strconv.Atoi(freeStr); err == nil {
					// Convert from KB to MB
					freeMB = free / 1024
				}
			}
		}
	}

	if totalMB > 0 {
		return &MemoryInfo{
			TotalMB:     totalMB,
			AvailableMB: freeMB,
			UsageMB:     totalMB - freeMB,
		}, nil
	}

	return mm.getWindowsMemoryFallback()
}

// getWindowsMemoryFallback provides a fallback method for Windows memory detection
func (mm *MemoryManager) getWindowsMemoryFallback() (*MemoryInfo, error) {
	// Use WMIC as fallback
	cmd := exec.Command("wmic", "computersystem", "get", "totalphysicalmemory")
	output, err := cmd.Output()
	if err != nil {
		mm.logger.Warn("Failed to get Windows memory info via WMIC: %v", err)
		return mm.getDefaultMemoryInfo(), nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "TotalPhysicalMemory") {
			// Skip header line
			continue
		}
		if line != "" {
			// Parse memory value
			if totalBytes, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64); err == nil {
				totalMB := int(totalBytes / (1024 * 1024))
				// Estimate available memory (75% of total as rough estimate)
				availableMB := int(float64(totalMB) * 0.75)
				return &MemoryInfo{
					TotalMB:     totalMB,
					AvailableMB: availableMB,
					UsageMB:     totalMB - availableMB,
				}, nil
			}
		}
	}

	return mm.getDefaultMemoryInfo(), nil
}

// getLinuxMemory gets memory information on Linux
func (mm *MemoryManager) getLinuxMemory() (*MemoryInfo, error) {
	// Read from /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		mm.logger.Warn("Failed to read /proc/meminfo: %v", err)
		return mm.getDefaultMemoryInfo(), nil
	}

	lines := strings.Split(string(data), "\n")
	var totalKB, availableKB int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if total, err := strconv.Atoi(parts[1]); err == nil {
					totalKB = total
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if available, err := strconv.Atoi(parts[1]); err == nil {
					availableKB = available
				}
			}
		}
	}

	if totalKB > 0 {
		if availableKB == 0 {
			// If MemAvailable is not available (older kernels), estimate it
			availableKB = int(float64(totalKB) * 0.7)
		}
		return &MemoryInfo{
			TotalMB:     totalKB / 1024,
			AvailableMB: availableKB / 1024,
			UsageMB:     (totalKB - availableKB) / 1024,
		}, nil
	}

	return mm.getDefaultMemoryInfo(), nil
}

// getDarwinMemory gets memory information on macOS
func (mm *MemoryManager) getDarwinMemory() (*MemoryInfo, error) {
	// Use sysctl to get memory info
	cmd := exec.Command("sysctl", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		mm.logger.Warn("Failed to get macOS memory info via sysctl: %v", err)
		return mm.getDefaultMemoryInfo(), nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "hw.memsize:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if totalBytes, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					totalMB := int(totalBytes / (1024 * 1024))
					// Estimate available memory (70% of total as rough estimate)
					availableMB := int(float64(totalMB) * 0.7)
					return &MemoryInfo{
						TotalMB:     totalMB,
						AvailableMB: availableMB,
						UsageMB:     totalMB - availableMB,
					}, nil
				}
			}
		}
	}

	return mm.getDefaultMemoryInfo(), nil
}

// getDefaultMemoryInfo returns default memory information when detection fails
func (mm *MemoryManager) getDefaultMemoryInfo() *MemoryInfo {
	mm.logger.Info("Using default memory configuration")
	return &MemoryInfo{
		TotalMB:     16384, // 16GB default
		AvailableMB: 12288, // 12GB available (75%)
		UsageMB:     4096,  // 4GB used
	}
}

// GetOptimalMemorySettings calculates optimal memory settings for Minecraft
func (mm *MemoryManager) GetOptimalMemorySettings() (minMB, maxMB int) {
	memInfo, err := mm.GetSystemMemory()
	if err != nil {
		mm.logger.Warn("Failed to get system memory info, using defaults: %v", err)
		return 2048, 4096 // 2GB min, 4GB max as safe defaults
	}

	mm.logger.Debug("System memory: %d MB total, %d MB available", memInfo.TotalMB, memInfo.AvailableMB)

	totalMB := memInfo.TotalMB
	availableMB := memInfo.AvailableMB

	// Calculate optimal memory settings based on available memory
	switch {
	case totalMB >= 32768: // 32GB+ system
		minMB = 4096  // 4GB min
		maxMB = 12288 // 12GB max

	case totalMB >= 16384: // 16GB+ system
		minMB = 3072  // 3GB min
		maxMB = 8192  // 8GB max

	case totalMB >= 8192: // 8GB+ system
		minMB = 2048  // 2GB min
		maxMB = 6144  // 6GB max

	case totalMB >= 4096: // 4GB+ system
		minMB = 1024  // 1GB min
		maxMB = 3072  // 3GB max

	default: // Less than 4GB
		minMB = 512   // 512MB min
		maxMB = 2048  // 2GB max
	}

	// Ensure we don't exceed available memory (leave some for system)
	availableForMinecraft := int(float64(availableMB) * 0.8) // Use 80% of available memory

	if maxMB > availableForMinecraft {
		maxMB = availableForMinecraft
		// Adjust min if it's too close to max
		if minMB > maxMB-512 {
			minMB = maxMB - 512
		}
	}

	// Ensure minimum requirements
	if minMB < 512 {
		minMB = 512
	}
	if maxMB < minMB {
		maxMB = minMB
	}

	mm.logger.Info("Optimal memory settings: %d MB min, %d MB max", minMB, maxMB)
	return minMB, maxMB
}

// GetMemoryPresets returns available memory presets
func (mm *MemoryManager) GetMemoryPresets() []MemoryPreset {
	presets := []MemoryPreset{
		{Name: "Low", MinMB: 1024, MaxMB: 2048, Description: "For systems with 4-8GB RAM"},
		{Name: "Medium", MinMB: 2048, MaxMB: 4096, Description: "For systems with 8-16GB RAM"},
		{Name: "High", MinMB: 4096, MaxMB: 8192, Description: "For systems with 16-32GB RAM"},
		{Name: "Ultra", MinMB: 6144, MaxMB: 12288, Description: "For systems with 32GB+ RAM"},
	}

	// Filter presets based on system memory
	memInfo, err := mm.GetSystemMemory()
	if err == nil {
		filteredPresets := make([]MemoryPreset, 0)
		for _, preset := range presets {
			if preset.MaxMB <= memInfo.TotalMB {
				filteredPresets = append(filteredPresets, preset)
			}
		}
		if len(filteredPresets) > 0 {
			return filteredPresets
		}
	}

	return presets
}

// MemoryPreset represents a memory configuration preset
type MemoryPreset struct {
	Name        string
	MinMB       int
	MaxMB       int
	Description string
}

// ValidateMemorySettings validates memory settings
func (mm *MemoryManager) ValidateMemorySettings(minMB, maxMB int) error {
	if minMB < 512 {
		return fmt.Errorf("minimum memory must be at least 512MB")
	}
	if maxMB < minMB {
		return fmt.Errorf("maximum memory cannot be less than minimum memory")
	}

	memInfo, err := mm.GetSystemMemory()
	if err == nil {
		if maxMB > memInfo.TotalMB {
			return fmt.Errorf("maximum memory (%d MB) exceeds total system memory (%d MB)", maxMB, memInfo.TotalMB)
		}
		if maxMB > memInfo.AvailableMB {
			mm.logger.Warn("Maximum memory (%d MB) exceeds available memory (%d MB)", maxMB, memInfo.AvailableMB)
		}
	}

	return nil
}