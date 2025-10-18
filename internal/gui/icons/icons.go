// Package icons provides custom icons for TheBoys Launcher
package icons

import (
	"fyne.io/fyne/v2"
)

// Custom application icons
var (
	// Application icon (will be replaced with actual icon file)
	AppIcon fyne.Resource = nil

	// UI icons
	HomeIcon      = homeIcon()
	ModpackIcon   = modpackIcon()
	SettingsIcon  = settingsIcon()
	DownloadIcon  = downloadIcon()
	LaunchIcon    = launchIcon()
	UpdateIcon    = updateIcon()
	InfoIcon      = infoIcon()
	WarningIcon   = warningIcon()
	ErrorIcon     = errorIcon()
	SuccessIcon   = successIcon()
)

// homeIcon creates a home icon
func homeIcon() fyne.Resource {
	// Simple SVG home icon as a resource
	return fyne.NewStaticResource("home", []byte{
		// This would be replaced with actual icon data
		// For now, we'll use a simple placeholder
	})
}

// modpackIcon creates a modpack icon
func modpackIcon() fyne.Resource {
	return fyne.NewStaticResource("modpack", []byte{
		// Placeholder for modpack icon
	})
}

// settingsIcon creates a settings icon
func settingsIcon() fyne.Resource {
	return fyne.NewStaticResource("settings", []byte{
		// Placeholder for settings icon
	})
}

// downloadIcon creates a download icon
func downloadIcon() fyne.Resource {
	return fyne.NewStaticResource("download", []byte{
		// Placeholder for download icon
	})
}

// launchIcon creates a launch icon
func launchIcon() fyne.Resource {
	return fyne.NewStaticResource("launch", []byte{
		// Placeholder for launch icon
	})
}

// updateIcon creates an update icon
func updateIcon() fyne.Resource {
	return fyne.NewStaticResource("update", []byte{
		// Placeholder for update icon
	})
}

// infoIcon creates an info icon
func infoIcon() fyne.Resource {
	return fyne.NewStaticResource("info", []byte{
		// Placeholder for info icon
	})
}

// warningIcon creates a warning icon
func warningIcon() fyne.Resource {
	return fyne.NewStaticResource("warning", []byte{
		// Placeholder for warning icon
	})
}

// errorIcon creates an error icon
func errorIcon() fyne.Resource {
	return fyne.NewStaticResource("error", []byte{
		// Placeholder for error icon
	})
}

// successIcon creates a success icon
func successIcon() fyne.Resource {
	return fyne.NewStaticResource("success", []byte{
		// Placeholder for success icon
	})
}

// LoadIconResources loads icon resources from files
func LoadIconResources() error {
	// This would load actual icon files from assets directory
	// For now, we'll just log that icons are loaded
	// In a real implementation, this would:
	// 1. Load PNG/SVG files from assets/icons/
	// 2. Convert them to fyne.Resource
	// 3. Set the global icon variables

	return nil
}

// GetStatusIcon returns the appropriate status icon
func GetStatusIcon(status string) fyne.Resource {
	switch status {
	case "error":
		return ErrorIcon
	case "warning":
		return WarningIcon
	case "success":
		return SuccessIcon
	default:
		return InfoIcon
	}
}