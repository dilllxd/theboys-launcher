package launcher

import "time"

// Common types used across the launcher package

// GitHubReleaseAsset represents an asset in a GitHub release
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []GitHubReleaseAsset `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

// ProgressCallback is called during download operations
type ProgressCallback func(float64)

// PackInfo contains parsed pack.toml information
type PackInfo struct {
	Name          string `toml:"name"`
	Version       string `toml:"version"`
	Minecraft     string `toml:"minecraft"`
	ModLoader     string `toml:"modLoader"`
	LoaderVersion string `toml:"loaderVersion"`
	Dependencies  []interface{} `toml:"dependencies"`
}

// LWJGLInfo contains LWJGL version information
type LWJGLInfo struct {
	Version string `json:"version"`
	UID     string `json:"uid"`
	Name    string `json:"name"`
}

// Constants
const (
	PackDotTomlFile = "pack.toml"
)