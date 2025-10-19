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
	launcherName      = "TheBoysLauncher"
	launcherShortName = "TheBoysLauncher"
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
		MemoryMB: clampMemoryMB(DefaultAutoMemoryMB()),
		AutoRAM:  true,
	}

	// Try to load existing settings
	if data, err := os.ReadFile(settingsPath); err == nil {
		type storedSettings struct {
			MemoryMB int   `json:"memoryMB"`
			AutoRAM  *bool `json:"autoRam"`
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
			if !settings.AutoRAM {
				settings.MemoryMB = clampMemoryMB(settings.MemoryMB)
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

// DefaultAutoMemoryMB returns the baseline auto RAM target (half system RAM capped 2-16GB)
func DefaultAutoMemoryMB() int {
	total := totalRAMMB()
	if total <= 0 {
		total = 65536 // fallback 64GB
	}
	auto := clampMemoryMB(total / 2)
	if auto > total {
		auto = clampMemoryMB(total)
	}
	return auto
}

func computeAutoRAMForModpack(modpack Modpack) int {
	auto := DefaultAutoMemoryMB()
	total := totalRAMMB()
	if total > 0 && auto > total {
		auto = clampMemoryMB(total)
	}

	if modpack.RecommendedRam > 0 && modpack.RecommendedRam <= 16384 {
		desired := clampMemoryMB(modpack.RecommendedRam)
		if total > 0 && desired > total {
			desired = clampMemoryMB(total)
		}
		return desired
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
