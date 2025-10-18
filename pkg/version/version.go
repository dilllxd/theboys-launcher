// Package version manages version information for the Winterpack Launcher
package version

import (
	"fmt"
	"runtime"
)

// Version holds the current application version
const Version = "2.0.0"

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	GoVersion string
	BuildTime string
	Platform  string
}

// GetBuildInfo returns comprehensive build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// String returns a formatted version string
func (b BuildInfo) String() string {
	return fmt.Sprintf("Winterpack Launcher v%s (built with %s for %s)",
		b.Version, b.GoVersion, b.Platform)
}

// IsVersionNewer compares two version strings
func IsVersionNewer(current, new string) bool {
	// Simple version comparison - can be enhanced with semver library
	return current != new
}

// GetVersion returns the current version
func GetVersion() string {
	return Version
}