package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	launcherName      = "TheBoys Launcher"
	launcherShortName = "TheBoys"
	launcherExeName   = "TheBoysLauncher.exe"

	// Self-update source (GitHub Releases of this EXE)
	UPDATE_OWNER      = "dilllxd"
	UPDATE_REPO       = "theboys-launcher"
	UPDATE_ASSET      = launcherExeName
	remoteModpacksURL = "https://raw.githubusercontent.com/dilllxd/theboys-launcher/refs/heads/main/modpacks.json"

	envCacheBust = "THEBOYS_CACHEBUST"
	envNoPause   = "THEBOYS_NOPAUSE"
)

type Modpack struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	PackURL      string `json:"packUrl"`
	InstanceName string `json:"instanceName"`
	Description  string `json:"description"`
	Default      bool   `json:"default,omitempty"`
}

// LauncherSettings holds user-configurable launcher settings
type LauncherSettings struct {
	MemoryMB int `json:"memoryMB"` // Memory allocation in MB (2-16GB range)
}

var defaultModpackID string
var settings LauncherSettings

// Use TUI interface by default
var interactive = false

// Populated at build time via -X main.version=vX.Y.Z
var version = "dev"

// getUserAgent returns a user agent string with the launcher version and component name
func getUserAgent(component string) string {
	if version == "dev" {
		return fmt.Sprintf("TheBoys-%s/dev", component)
	}
	return fmt.Sprintf("TheBoys-%s/%s", component, version)
}

// loadSettings loads launcher settings from settings.json, creates defaults if needed
func loadSettings(root string) error {
	settingsPath := filepath.Join(root, "settings.json")

	// Default settings - use half of system RAM, within 2-16GB range
	totalRAM := totalRAMMB()
	if totalRAM <= 0 {
		totalRAM = 65536 // fallback 64GB
	}

	// Debug: Show detected RAM
	logf("%s", infoLine(fmt.Sprintf("Detected system RAM: %d GB", totalRAM/1024)))

	defaultMemory := totalRAM / 2 // Use half of system RAM
	if defaultMemory > 16384 {
		defaultMemory = 16384 // Cap at 16GB
	}
	if defaultMemory < 2048 {
		defaultMemory = 2048 // Minimum 2GB
	}

	defaultSettings := LauncherSettings{
		MemoryMB: defaultMemory,
	}

	// Try to load existing settings
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err == nil {
			// Validate loaded settings are within bounds
			if settings.MemoryMB < 2048 {
				settings.MemoryMB = 2048
			}
			if settings.MemoryMB > 16384 {
				settings.MemoryMB = 16384
			}
			return nil
		}
	}

	// Use defaults if loading failed
	settings = defaultSettings
	return saveSettings(root)
}

// saveSettings saves current settings to settings.json
func saveSettings(root string) error {
	settingsPath := filepath.Join(root, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

// resetToAutoSettings resets memory to auto-detected values
func resetToAutoSettings(root string) {
	totalRAM := totalRAMMB()
	if totalRAM <= 0 {
		totalRAM = 65536 // fallback 64GB
	}

	autoMemory := totalRAM / 2 // Use half of system RAM
	if autoMemory > 16384 {
		autoMemory = 16384 // Cap at 16GB
	}
	if autoMemory < 2048 {
		autoMemory = 2048 // Minimum 2GB
	}

	settings.MemoryMB = autoMemory

	fmt.Printf("\n%s", dividerLine())
	fmt.Printf("%s", successLine("Memory settings reset to auto:"))
	fmt.Printf("  ■ Memory: %d GB\n", autoMemory/1024)
	fmt.Printf("  %s\n", infoLine(fmt.Sprintf("Based on half of your %d GB total RAM", totalRAM/1024)))
	fmt.Printf("%s", dividerLine())
}

func autoRAM() (minMB, maxMB int) {
	// Use configured settings as both min and max for simplicity
	return settings.MemoryMB, settings.MemoryMB
}

// totalRAMMB returns total system memory in MB
func totalRAMMB() int {
	// Use Windows GlobalMemoryStatusEx API to get total physical memory
	// Define the structure ourselves since it's not in the basic windows package
	type memoryStatusEx struct {
		DwLength                uint32
		DwMemoryLoad            uint32
		UllTotalPhys            uint64
		UllAvailPhys            uint64
		UllTotalPageFile        uint64
		UllAvailPageFile        uint64
		UllTotalVirtual         uint64
		UllAvailVirtual         uint64
		UllAvailExtendedVirtual uint64
	}

	var memStatus memoryStatusEx
	memStatus.DwLength = uint32(unsafe.Sizeof(memStatus))

	// Call the Windows API directly
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		// Fallback to default if API call fails
		return 16384 // 16GB default
	}

	// Convert bytes to megabytes
	totalMB := int(memStatus.UllTotalPhys / (1024 * 1024))

	// Validate the result seems reasonable
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		return 16384 // Use default if result seems invalid
	}

	return totalMB
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}