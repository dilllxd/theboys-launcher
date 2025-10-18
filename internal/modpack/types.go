// Package modpack provides modpack management functionality
package modpack

import (
	"fmt"
	"time"
)

// ModpackType represents the type of modpack
type ModpackType string

const (
	ModpackTypeCurseForge ModpackType = "curseforge"
	ModpackTypeModrinth   ModpackType = "modrinth"
	ModpackTypeCustom     ModpackType = "custom"
	ModpackTypeLocal      ModpackType = "local"
)

// ModpackStatus represents the installation status of a modpack
type ModpackStatus string

const (
	StatusNotInstalled    ModpackStatus = "not_installed"
	StatusInstalled       ModpackStatus = "installed"
	StatusPartial         ModpackStatus = "partial"
	StatusUpdateAvailable ModpackStatus = "update_available"
	StatusDownloading     ModpackStatus = "downloading"
	StatusInstalling      ModpackStatus = "installing"
	StatusError           ModpackStatus = "error"
)

// MinecraftVersion represents a Minecraft version
type MinecraftVersion struct {
	ID          string `json:"id"`
	ReleaseType string `json:"release_type"` // "release", "snapshot", "beta", "alpha"
	Date        string `json:"date"`
}

// ModLoader represents a mod loader type
type ModLoader struct {
	Type    string `json:"type"`    // "forge", "fabric", "quilt", "neoforge"
	Version string `json:"version"` // e.g., "47.2.0" for Forge 1.20.2
}

// Modpack represents a modpack configuration
type Modpack struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	Author               string         `json:"author"`
	Version              string         `json:"version"`
	Type                 ModpackType    `json:"type"`
	Status               ModpackStatus  `json:"status"`
	InstallPath          string         `json:"install_path"`
	MinecraftVersion     MinecraftVersion `json:"minecraft_version"`
	ModLoader            ModLoader      `json:"mod_loader"`
	IconURL              string         `json:"icon_url"`
	ScreenshotURLs       []string       `json:"screenshot_urls"`
	DownloadURL          string         `json:"download_url"`
	SourceURL            string         `json:"source_url"`
	RequiredMemory       int            `json:"required_memory_mb"` // Required RAM in MB
	TotalSize            int64          `json:"total_size"`         // Total size in bytes
	InstalledSize        int64          `json:"installed_size"`     // Installed size in bytes
	DateCreated          time.Time      `json:"date_created"`
	DateModified         time.Time      `json:"date_modified"`
	LastPlayed           *time.Time     `json:"last_played,omitempty"`
	PlayTime             int64          `json:"play_time_seconds"`   // Total play time in seconds
	Dependencies         []Dependency   `json:"dependencies"`
	Tags                 []string       `json:"tags"`
	Features             []string       `json:"features"`
	IsFavorite           bool           `json:"is_favorite"`
	AutoUpdate           bool           `json:"auto_update"`
	CustomLaunchArgs     []string       `json:"custom_launch_args,omitempty"`
	CustomJvmArgs        []string       `json:"custom_jvm_args,omitempty"`
}

// Dependency represents a modpack dependency
type Dependency struct {
	ID      string `json:"id"`
	Type    string `json:"type"`    // "modpack", "mod", "library"
	Version string `json:"version"`
	Required bool   `json:"required"`
}

// ModpackManifest represents the manifest file for a modpack
type ModpackManifest struct {
	FormatVersion int          `json:"format_version"`
	Game          string       `json:"game"`           // Should be "minecraft"
	VersionID     string       `json:"version_id"`     // Minecraft version
	Name          string       `json:"name"`
	Summary       string       `json:"summary,omitempty"`
	Files         []ManifestFile `json:"files"`
	Dependencies  []ManifestDependency `json:"dependencies"`
}

// ManifestFile represents a file in the modpack manifest
type ManifestFile struct {
	ProjectID   int    `json:"projectID"`
	FileID      int    `json:"fileID"`
	Required    bool   `json:"required"`
	DownloadURL string `json:"download_url,omitempty"`
	Filename    string `json:"filename,omitempty"`
}

// ManifestDependency represents a dependency in the modpack manifest
type ManifestDependency struct {
	ModID       string `json:"modid"`       // For Forge
	Side        string `json:"side"`        // "client", "server", "both"
	Type        string `json:"type"`        // "forge", "fabric", etc.
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// InstallationProgress represents the progress of a modpack installation
type InstallationProgress struct {
	ModpackID      string  `json:"modpack_id"`
	Stage          string  `json:"stage"`          // "downloading", "extracting", "installing", "finishing"
	Progress       float64 `json:"progress"`       // 0.0 to 1.0
	CurrentFile    string  `json:"current_file"`
	TotalFiles     int     `json:"total_files"`
	CompletedFiles int     `json:"completed_files"`
	DownloadSpeed  int64   `json:"download_speed"`  // bytes per second
	ETA            int     `json:"eta_seconds"`    // estimated time remaining
	Error          string  `json:"error,omitempty"`
}

// ModpackInstance represents a locally installed modpack instance
type ModpackInstance struct {
	ModpackID      string    `json:"modpack_id"`
	InstanceID     string    `json:"instance_id"`
	Name           string    `json:"name"`
	InstancePath   string    `json:"instance_path"`
	DateInstalled  time.Time `json:"date_installed"`
	LastModified   time.Time `json:"last_modified"`
	LastPlayed     *time.Time `json:"last_played,omitempty"`
	PlayTime       int64     `json:"play_time_seconds"`
	IsRunning      bool      `json:"is_running"`
	CurrentVersion string    `json:"current_version"`
	UpdatesAvailable int     `json:"updates_available"`
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`
}

// IsValid checks if the modpack data is valid
func (m *Modpack) IsValid() bool {
	return m.ID != "" &&
		   m.Name != "" &&
		   m.MinecraftVersion.ID != "" &&
		   m.ModLoader.Type != "" &&
		   m.Type != ""
}

// IsInstalled checks if the modpack is installed
func (m *Modpack) IsInstalled() bool {
	return m.Status == StatusInstalled ||
		   m.Status == StatusPartial ||
		   m.Status == StatusUpdateAvailable
}

// NeedsUpdate checks if the modpack has updates available
func (m *Modpack) NeedsUpdate() bool {
	return m.Status == StatusUpdateAvailable
}

// GetFormattedSize returns a human-readable size string
func (m *Modpack) GetFormattedSize() string {
	if m.TotalSize == 0 {
		return "Unknown"
	}
	return formatBytes(m.TotalSize)
}

// GetFormattedInstalledSize returns a human-readable installed size string
func (m *Modpack) GetFormattedInstalledSize() string {
	if m.InstalledSize == 0 {
		return "Unknown"
	}
	return formatBytes(m.InstalledSize)
}

// GetFormattedPlayTime returns a human-readable play time string
func (m *Modpack) GetFormattedPlayTime() string {
	hours := m.PlayTime / 3600
	minutes := (m.PlayTime % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}