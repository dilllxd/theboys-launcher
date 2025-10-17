package types

import "time"

// JavaInstallation represents a detected Java installation
type JavaInstallation struct {
	Path        string `json:"path"`
	Version     string `json:"version"`
	IsJDK       bool   `json:"is_jdk"`
	Architecture string `json:"architecture"`
}

// Modpack represents a modpack configuration
type Modpack struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	PackURL      string `json:"packUrl"`
	InstanceName string `json:"instanceName"`
	Description  string `json:"description"`
	Default      bool   `json:"default,omitempty"`
}

// LauncherSettings represents user-configurable launcher settings
type LauncherSettings struct {
	MemoryMB          int    `json:"memoryMB"`           // Memory allocation in MB (2-16GB range)
	AutoUpdate        bool   `json:"autoUpdate"`         // Enable automatic updates
	CheckForUpdates   bool   `json:"checkForUpdates"`    // Check for updates on startup
	KeepConsoleOpen   bool   `json:"keepConsoleOpen"`    // Keep console open after launch
	JavaPath          string `json:"javaPath,omitempty"` // Custom Java path (empty for auto-detect)
	LastModpackID     string `json:"lastModpackID"`      // Last selected modpack
	WindowSize        Size   `json:"windowSize"`         // Window dimensions
	Theme             string `json:"theme"`              // UI theme (light/dark)
}

// Size represents window dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Instance represents a Minecraft instance
type Instance struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ModpackID    string            `json:"modpackID"`
	GameVersion  string            `json:"gameVersion"`
	JavaVersion  string            `json:"javaVersion"`
	MemoryMB     int               `json:"memoryMB"`
	LastPlayed   *time.Time        `json:"lastPlayed,omitempty"`
	TotalTime    time.Duration     `json:"totalTime"`
	InstancePath string            `json:"instancePath"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// DownloadProgress represents download progress information
type DownloadProgress struct {
	URL           string  `json:"url"`
	TotalBytes    int64   `json:"totalBytes"`
	DownloadedBytes int64 `json:"downloadedBytes"`
	Percentage    float64 `json:"percentage"`
	Speed         int64   `json:"speed"`        // bytes per second
	ETA           int64   `json:"eta"`          // estimated time remaining in seconds
	Status        string  `json:"status"`       // downloading, completed, error, etc.
	Error         string  `json:"error,omitempty"`
}

// UpdateInfo represents available update information
type UpdateInfo struct {
	Version     string    `json:"version"`
	ReleaseURL  string    `json:"releaseUrl"`
	DownloadURL string    `json:"downloadUrl"`
	ReleaseNotes string   `json:"releaseNotes"`
	ReleasedAt  time.Time `json:"releasedAt"`
	IsRequired  bool      `json:"isRequired"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // debug, info, warn, error
	Message   string    `json:"message"`
	Component string    `json:"component"` // launcher, updater, etc.
}