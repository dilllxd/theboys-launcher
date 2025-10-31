package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	launcherName      = "TheBoysLauncher"
	launcherShortName = "TheBoysLauncher"
	launcherExeName   = "TheBoysLauncher" // Base name without extension

	// Self-update source (GitHub Releases of this EXE)
	UPDATE_OWNER      = "dilllxd"
	UPDATE_REPO       = "theboyslauncher"
	remoteModpacksURL = "https://modpacks.dylan.lol/modpacks.json"

	envCacheBust = "THEBOYS_CACHEBUST"
	envNoPause   = "THEBOYS_NOPAUSE"
)

type Modpack struct {
	ID             string   `json:"id"`
	DisplayName    string   `json:"displayName"`
	PackURL        string   `json:"packUrl"`
	InstanceName   string   `json:"instanceName"`
	Description    string   `json:"description"`
	Author         string   `json:"author"`
	Tags           []string `json:"tags"`
	LastUpdated    string   `json:"lastUpdated"`
	Category       string   `json:"category"`
	MinRam         int      `json:"minRam"`
	RecommendedRam int      `json:"recommendedRam"`
	Changelog      string   `json:"changelog"`
	// Legacy support
	Default bool `json:"default,omitempty"`
}

// LauncherSettings holds user-configurable launcher settings
type LauncherSettings struct {
	MemoryMB int  `json:"memoryMB"` // Memory allocation in MB (2-16GB range)
	AutoRAM  bool `json:"autoRam"`  // Whether to auto-manage RAM per modpack
	// If true, the launcher will check and install prerelease/dev builds from releases
	DevBuildsEnabled bool `json:"devBuildsEnabled,omitempty"`
	// If true, enables debug logging for troubleshooting
	DebugEnabled bool `json:"debugEnabled,omitempty"`
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

	defaultSettings := LauncherSettings{
		MemoryMB:         clampMemoryMB(DefaultAutoMemoryMB()),
		AutoRAM:          true,
		DevBuildsEnabled: isDevBuild(),
		DebugEnabled:     false, // Debug disabled by default for better user experience
	}

	// Try to load existing settings
	if data, err := os.ReadFile(settingsPath); err == nil {
		type storedSettings struct {
			MemoryMB         int   `json:"memoryMB"`
			AutoRAM          *bool `json:"autoRam"`
			DevBuildsEnabled *bool `json:"devBuildsEnabled"`
			DebugEnabled     *bool `json:"debugEnabled,omitempty"`
		}
		var stored storedSettings
		if err := json.Unmarshal(data, &stored); err == nil {
			settings.MemoryMB = clampMemoryMB(stored.MemoryMB)
			if settings.MemoryMB == 0 {
				settings.MemoryMB = defaultSettings.MemoryMB
			}
			if stored.AutoRAM == nil {
				settings.AutoRAM = true
			} else {
				settings.AutoRAM = *stored.AutoRAM
			}
			if stored.DevBuildsEnabled == nil {
				settings.DevBuildsEnabled = false
			} else {
				settings.DevBuildsEnabled = *stored.DevBuildsEnabled
			}
			if stored.DebugEnabled == nil {
				settings.DebugEnabled = defaultSettings.DebugEnabled
			} else {
				settings.DebugEnabled = *stored.DebugEnabled
			}
			if !settings.AutoRAM {
				settings.MemoryMB = clampMemoryMB(settings.MemoryMB)
			}
			// Only log dev build status without overriding user preference
			if isDevBuild() {
				if settings.DevBuildsEnabled {
					logf("%s", infoLine(fmt.Sprintf("Dev build detected (version: %s), dev builds already enabled by user preference", version)))
				} else {
					logf("%s", infoLine(fmt.Sprintf("Dev build detected (version: %s), dev builds disabled by user preference", version)))
				}
			}
			return nil
		}
	}

	// Use defaults if loading failed
	settings = defaultSettings
	// Log when using default dev builds setting
	if isDevBuild() && settings.DevBuildsEnabled {
		logf("%s", infoLine(fmt.Sprintf("New installation detected with dev build (version: %s), dev builds enabled by default", version)))
	}
	return saveSettings(root)
}

// saveSettings saves current settings to settings.json
func saveSettings(root string) error {
	settingsPath := filepath.Join(root, "settings.json")
	logf("%s", infoLine(fmt.Sprintf("Saving settings: DevBuildsEnabled=%t, AutoRAM=%t, MemoryMB=%d, DebugEnabled=%t",
		settings.DevBuildsEnabled, settings.AutoRAM, settings.MemoryMB, settings.DebugEnabled)))
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to write settings file: %v", err)))
	} else {
		logf("%s", successLine("Settings saved successfully"))
	}
	return err
}

// resetToAutoSettings resets memory to auto-detected values
func resetToAutoSettings(root string) {
	settings.AutoRAM = true
	settings.MemoryMB = clampMemoryMB(DefaultAutoMemoryMB())

	fmt.Printf("\n%s", dividerLine())
	fmt.Printf("%s", successLine("Memory settings reset to auto"))
	fmt.Printf("  ■ Auto RAM enabled\n")
	fmt.Printf("  ■ Baseline memory: %d GB\n", settings.MemoryMB/1024)
	fmt.Printf("%s", dividerLine())
}

func clampMemoryMB(mb int) int {
	if mb < 2048 {
		return 2048
	}
	if mb > 16384 {
		return 16384
	}
	return mb
}

// DefaultAutoMemoryMB returns the baseline auto RAM target (half total RAM capped 2-16GB)
func DefaultAutoMemoryMB() int {
	total := totalRAMMB()
	
	// If total memory detection fails, fall back to 32GB
	if total <= 0 {
		total = 32768 // fallback 32GB
	}
	
	// Calculate half of total memory and clamp to 2-16GB range
	auto := clampMemoryMB(total / 2)
	return auto
}

func computeAutoRAMForModpack(modpack Modpack) int {
	auto := DefaultAutoMemoryMB()
	total := totalRAMMB()
	
	// Ensure we don't allocate more than total memory
	if total > 0 && auto > total {
		auto = clampMemoryMB(total)
	}

	// If modpack has a recommended RAM value, use the minimum of auto and recommended
	if modpack.RecommendedRam > 0 && modpack.RecommendedRam <= 16384 {
		recommended := clampMemoryMB(modpack.RecommendedRam)
		// Ensure recommended RAM doesn't exceed total memory
		if total > 0 && recommended > total {
			recommended = clampMemoryMB(total)
		}
		// Return the minimum of auto-calculated RAM and recommended RAM
		return min(auto, recommended)
	}
	return auto
}

// MemoryForModpack returns the memory allocation that should be applied for the given modpack
func MemoryForModpack(modpack Modpack) int {
	if settings.AutoRAM {
		mem := clampMemoryMB(computeAutoRAMForModpack(modpack))
		settings.MemoryMB = mem
		return mem
	}
	settings.MemoryMB = clampMemoryMB(settings.MemoryMB)
	return settings.MemoryMB
}

// totalRAMMB is now implemented in platform-specific files
// This function is handled by platform_windows.go and platform_darwin.go

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

func isDevBuild() bool {
	lower := strings.ToLower(version)
	return strings.Contains(lower, "dev")
}
