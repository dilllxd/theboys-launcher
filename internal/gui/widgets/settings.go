package widgets

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"theboys-launcher/internal/config"
)

// SettingsWidget represents a settings widget with real-time updates
type SettingsWidget struct {
	widget.BaseWidget
	content          *container.Scroll
	settingsManager  *config.SettingsManager
	categoryTabs     *container.AppTabs
	generalTab       *container.VBox
	launcherTab      *container.VBox
	advancedTab      *container.VBox
}

// NewSettingsWidget creates a new settings widget
func NewSettingsWidget(settingsManager *config.SettingsManager) *SettingsWidget {
	sw := &SettingsWidget{
		settingsManager: settingsManager,
	}
	sw.ExtendBaseWidget(sw)

	// Create category tabs
	sw.categoryTabs = container.NewAppTabs()

	// Create individual tabs
	sw.createGeneralTab()
	sw.createLauncherTab()
	sw.createAdvancedTab()

	// Add tabs to container
	sw.categoryTabs.Append(container.NewTabItemWithIcon("General", theme.SettingsIcon(), sw.generalTab))
	sw.categoryTabs.Append(container.NewTabItemWithIcon("Launcher", theme.ComputerIcon(), sw.launcherTab))
	sw.categoryTabs.Append(container.NewTabItemWithIcon("Advanced", theme.SettingsIcon(), sw.advancedTab))

	// Create scrollable content
	sw.content = container.NewScroll(sw.categoryTabs)

	return sw
}

// CreateRenderer creates the renderer for the settings widget
func (sw *SettingsWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sw.content)
}

// createGeneralTab creates the general settings tab
func (sw *SettingsWidget) createGeneralTab() {
	currentConfig := sw.settingsManager.GetConfig()

	// Theme selection
	themeSelect := widget.NewSelect([]string{"Light", "Dark", "System"}, func(selected string) {
		sw.updateSetting("Theme", selected)
	})
	themeSelect.SetSelected(currentConfig.Theme)

	// Update settings
	autoUpdateCheck := widget.NewCheck("Automatically check for updates", func(checked bool) {
		sw.updateSetting("CheckUpdatesOnStartup", checked)
	})
	autoUpdateCheck.SetChecked(currentConfig.CheckUpdatesOnStartup)

	updateChannelSelect := widget.NewSelect([]string{"Stable", "Beta", "Alpha"}, func(selected string) {
		sw.updateSetting("UpdateChannel", selected)
	})
	updateChannelSelect.SetSelected(currentConfig.UpdateChannel)

	// Create general settings form
	sw.generalTab = container.NewVBox(
		widget.NewCard("Appearance", "", container.NewVBox(
			widget.NewLabel("Theme:"),
			themeSelect,
		)),
		widget.NewCard("Updates", "", container.NewVBox(
			autoUpdateCheck,
			widget.NewLabel("Update Channel:"),
			updateChannelSelect,
		)),
	)
}

// createLauncherTab creates the launcher settings tab
func (sw *SettingsWidget) createLauncherTab() {
	currentConfig := sw.settingsManager.GetConfig()

	// Memory allocation
	memorySlider := widget.NewSlider(512, 32768)
	memorySlider.SetValue(float64(currentConfig.MemoryMB))
	memoryLabel := widget.NewLabel(fmt.Sprintf("Memory: %d MB (%.1f GB)", currentConfig.MemoryMB, float64(currentConfig.MemoryMB)/1024))

	memorySlider.OnChanged = func(value float64) {
		memoryMB := int(value)
		memoryLabel.SetText(fmt.Sprintf("Memory: %d MB (%.1f GB)", memoryMB, float64(memoryMB)/1024))
		sw.updateSetting("MemoryMB", memoryMB)
	}

	// Behavior settings
	autoLaunchCheck := widget.NewCheck("Auto-launch modpack on selection", func(checked bool) {
		sw.updateSetting("AutoLaunchOnSelect", checked)
	})
	autoLaunchCheck.SetChecked(currentConfig.AutoLaunchOnSelect)

	showAdvancedCheck := widget.NewCheck("Show advanced options", func(checked bool) {
		sw.updateSetting("ShowAdvancedOptions", checked)
	})
	showAdvancedCheck.SetChecked(currentConfig.ShowAdvancedOptions)

	// Create launcher settings form
	sw.launcherTab = container.NewVBox(
		widget.NewCard("Performance", "", container.NewVBox(
			memoryLabel,
			memorySlider,
			widget.NewLabel("Memory allocation affects how much RAM Minecraft can use."),
		)),
		widget.NewCard("Behavior", "", container.NewVBox(
			autoLaunchCheck,
			showAdvancedCheck,
		)),
	)
}

// createAdvancedTab creates the advanced settings tab
func (sw *SettingsWidget) createAdvancedTab() {
	currentConfig := sw.settingsManager.GetConfig()

	// Download timeout
	timeoutSlider := widget.NewSlider(10, 3600)
	timeoutSlider.SetValue(float64(currentConfig.DownloadTimeout))
	timeoutLabel := widget.NewLabel(fmt.Sprintf("Download Timeout: %d seconds", currentConfig.DownloadTimeout))

	timeoutSlider.OnChanged = func(value float64) {
		timeout := int(value)
		timeoutLabel.SetText(fmt.Sprintf("Download Timeout: %d seconds", timeout))
		sw.updateSetting("DownloadTimeout", timeout)
	}

	// Concurrent downloads
	concurrentSlider := widget.NewSlider(1, 10)
	concurrentSlider.SetValue(float64(currentConfig.MaxConcurrentDownloads))
	concurrentLabel := widget.NewLabel(fmt.Sprintf("Max Concurrent Downloads: %d", currentConfig.MaxConcurrentDownloads))

	concurrentSlider.OnChanged = func(value float64) {
		concurrent := int(value)
		concurrentLabel.SetText(fmt.Sprintf("Max Concurrent Downloads: %d", concurrent))
		sw.updateSetting("MaxConcurrentDownloads", concurrent)
	}

	// Debug logging
	debugCheck := widget.NewCheck("Enable debug logging", func(checked bool) {
		sw.updateSetting("EnableDebugLog", checked)
	})
	debugCheck.SetChecked(currentConfig.EnableDebugLog)

	// Log level
	logLevelSelect := widget.NewSelect([]string{"Debug", "Info", "Warning", "Error"}, func(selected string) {
		sw.updateSetting("LogLevel", selected)
	})
	logLevelSelect.SetSelected(currentConfig.LogLevel)

	// Path settings
	javaPathEntry := sw.createPathEntry("Java Path", currentConfig.JavaPath, "JavaPath")
	prismPathEntry := sw.createPathEntry("Prism Launcher Path", currentConfig.PrismPath, "PrismPath")
	instancesPathEntry := sw.createPathEntry("Instances Path", currentConfig.InstancesPath, "InstancesPath")
	tempPathEntry := sw.createPathEntry("Temporary Files Path", currentConfig.TempPath, "TempPath")

	// Create advanced settings form
	sw.advancedTab = container.NewVBox(
		widget.NewCard("Network", "", container.NewVBox(
			timeoutLabel,
			timeoutSlider,
			concurrentLabel,
			concurrentSlider,
		)),
		widget.NewCard("Logging", "", container.NewVBox(
			debugCheck,
			widget.NewLabel("Log Level:"),
			logLevelSelect,
		)),
		widget.NewCard("Paths", "", container.NewVBox(
			javaPathEntry,
			prismPathEntry,
			instancesPathEntry,
			tempPathEntry,
		)),
		widget.NewCard("Actions", "", container.NewVBox(
			widget.NewButton("Reset to Defaults", sw.resetToDefaults),
			widget.NewButton("Export Settings", sw.exportSettings),
			widget.NewButton("Import Settings", sw.importSettings),
		)),
	)
}

// createPathEntry creates a path entry with browse button
func (sw *SettingsWidget) createPathEntry(label, currentValue, settingKey string) *fyne.Container {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Auto-detect")
	entry.SetText(currentValue)

	browseButton := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				path := reader.URI().Path()
				if len(path) > 0 && path[0] == '/' {
					// Remove leading slash on Windows
					path = path[1:]
				}
				entry.SetText(path)
				sw.updateSetting(settingKey, path)
				reader.Close()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
	})

	container := container.NewBorder(
		nil, nil,
		widget.NewLabel(label+":"),
		browseButton,
		entry,
	)

	return container
}

// updateSetting updates a setting value
func (sw *SettingsWidget) updateSetting(key string, value interface{}) {
	updates := map[string]interface{}{key: value}
	if err := sw.settingsManager.UpdateConfig(updates); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
	}
}

// resetToDefaults resets all settings to defaults
func (sw *SettingsWidget) resetToDefaults() {
	dialog.ShowConfirm("Reset Settings",
		"Are you sure you want to reset all settings to their default values? This action cannot be undone.",
		func(confirmed bool) {
			if confirmed {
				if err := sw.settingsManager.ResetToDefaults(); err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				} else {
					dialog.ShowInformation("Settings Reset", "All settings have been reset to their default values.",
						fyne.CurrentApp().Driver().AllWindows()[0])
					// Refresh the UI
					sw.refreshAllTabs()
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// exportSettings exports current settings to a file
func (sw *SettingsWidget) exportSettings() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err == nil && writer != nil {
			if err := sw.settingsManager.ExportSettings(writer.URI().Path()); err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			} else {
				dialog.ShowInformation("Export Complete", "Settings have been exported successfully.",
					fyne.CurrentApp().Driver().AllWindows()[0])
			}
			writer.Close()
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// importSettings imports settings from a file
func (sw *SettingsWidget) importSettings() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			if err := sw.settingsManager.ImportSettings(reader.URI().Path()); err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			} else {
				dialog.ShowInformation("Import Complete", "Settings have been imported successfully.",
					fyne.CurrentApp().Driver().AllWindows()[0])
				// Refresh the UI
				sw.refreshAllTabs()
			}
			reader.Close()
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// refreshAllTabs refreshes all settings tabs
func (sw *SettingsWidget) refreshAllTabs() {
	// Clear and recreate all tabs
	sw.generalTab.Objects = nil
	sw.launcherTab.Objects = nil
	sw.advancedTab.Objects = nil

	sw.createGeneralTab()
	sw.createLauncherTab()
	sw.createAdvancedTab()

	sw.categoryTabs.Refresh()
}

// GetSettingsManager returns the settings manager
func (sw *SettingsWidget) GetSettingsManager() *config.SettingsManager {
	return sw.settingsManager
}