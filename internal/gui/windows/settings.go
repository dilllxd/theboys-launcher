package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"winterpack-launcher/internal/app"
)

// SettingsWindow represents the settings window
type SettingsWindow struct {
	fyne.Window
	appState *app.State
	content  fyne.CanvasObject
}

// NewSettingsWindow creates a new settings window
func NewSettingsWindow(appState *app.State) *SettingsWindow {
	window := fyne.CurrentApp().NewWindow("Settings")
	window.Resize(fyne.NewSize(600, 500))
	window.SetFixedSize(false)

	settingsWindow := &SettingsWindow{
		Window:   window,
		appState: appState,
	}

	settingsWindow.createContent()

	// Setup window close handler
	window.SetCloseIntercept(func() {
		settingsWindow.onClose()
	})

	return settingsWindow
}

// createContent creates the settings window content
func (sw *SettingsWindow) createContent() {
	// Create tabs for different setting categories
	tabs := container.NewAppTabs(
		container.NewTabItem("General", sw.createGeneralSettings()),
		container.NewTabItem("Launcher", sw.createLauncherSettings()),
		container.NewTabItem("Advanced", sw.createAdvancedSettings()),
	)

	sw.content = tabs
	sw.SetContent(tabs)
}

// createGeneralSettings creates the general settings section
func (sw *SettingsWindow) createGeneralSettings() fyne.CanvasObject {
	// Theme selection
	themeSelect := widget.NewSelect([]string{"Light", "Dark", "System"}, func(selected string) {
		sw.appState.Config.Theme = selected
	})
	themeSelect.SetSelected(sw.appState.Config.Theme)

	// Auto-update setting
	autoUpdateCheck := widget.NewCheck("Check for updates on startup", func(checked bool) {
		sw.appState.Config.CheckUpdatesOnStartup = checked
	})
	autoUpdateCheck.SetChecked(sw.appState.Config.CheckUpdatesOnStartup)

	// Update channel
	updateChannelSelect := widget.NewSelect([]string{"Stable", "Beta", "Alpha"}, func(selected string) {
		sw.appState.Config.UpdateChannel = selected
	})
	updateChannelSelect.SetSelected(sw.appState.Config.UpdateChannel)

	// Create layout
	return container.NewVBox(
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

// createLauncherSettings creates the launcher settings section
func (sw *SettingsWindow) createLauncherSettings() fyne.CanvasObject {
	// Memory allocation
	memorySlider := widget.NewSlider(512, 32768)
	memorySlider.SetValue(float64(sw.appState.Config.MemoryMB))
	memoryLabel := widget.NewLabel("Memory: 4096 MB")

	memorySlider.OnChanged = func(value float64) {
		sw.appState.Config.MemoryMB = int(value)
		memoryLabel.SetText("Memory: " + sw.formatMemory(int(value)))
	}
	memoryLabel.SetText("Memory: " + sw.formatMemory(sw.appState.Config.MemoryMB))

	// Auto-launch setting
	autoLaunchCheck := widget.NewCheck("Auto-launch modpack on selection", func(checked bool) {
		sw.appState.Config.AutoLaunchOnSelect = checked
	})
	autoLaunchCheck.SetChecked(sw.appState.Config.AutoLaunchOnSelect)

	// Show advanced options
	advancedCheck := widget.NewCheck("Show advanced options", func(checked bool) {
		sw.appState.Config.ShowAdvancedOptions = checked
	})
	advancedCheck.SetChecked(sw.appState.Config.ShowAdvancedOptions)

	return container.NewVBox(
		widget.NewCard("Performance", "", container.NewVBox(
			memoryLabel,
			memorySlider,
		)),
		widget.NewCard("Behavior", "", container.NewVBox(
			autoLaunchCheck,
			advancedCheck,
		)),
	)
}

// createAdvancedSettings creates the advanced settings section
func (sw *SettingsWindow) createAdvancedSettings() fyne.CanvasObject {
	// Download timeout
	timeoutSlider := widget.NewSlider(10, 3600)
	timeoutSlider.SetValue(float64(sw.appState.Config.DownloadTimeout))
	timeoutLabel := widget.NewLabel("Download Timeout: 300 seconds")

	timeoutSlider.OnChanged = func(value float64) {
		sw.appState.Config.DownloadTimeout = int(value)
		timeoutLabel.SetText("Download Timeout: " + sw.formatTimeout(int(value)))
	}
	timeoutLabel.SetText("Download Timeout: " + sw.formatTimeout(sw.appState.Config.DownloadTimeout))

	// Max concurrent downloads
	concurrentSlider := widget.NewSlider(1, 10)
	concurrentSlider.SetValue(float64(sw.appState.Config.MaxConcurrentDownloads))
	concurrentLabel := widget.NewLabel("Max Concurrent Downloads: 3")

	concurrentSlider.OnChanged = func(value float64) {
		sw.appState.Config.MaxConcurrentDownloads = int(value)
		concurrentLabel.SetText("Max Concurrent Downloads: " + sw.formatInt(int(value)))
	}
	concurrentLabel.SetText("Max Concurrent Downloads: " + sw.formatInt(sw.appState.Config.MaxConcurrentDownloads))

	// Debug logging
	debugCheck := widget.NewCheck("Enable debug logging", func(checked bool) {
		sw.appState.Config.EnableDebugLog = checked
	})
	debugCheck.SetChecked(sw.appState.Config.EnableDebugLog)

	// Paths section
	javaPathEntry := widget.NewEntry()
	javaPathEntry.SetPlaceHolder("Auto-detect")
	javaPathEntry.SetText(sw.appState.Config.JavaPath)

	prismPathEntry := widget.NewEntry()
	prismPathEntry.SetPlaceHolder("Auto-download")
	prismPathEntry.SetText(sw.appState.Config.PrismPath)

	return container.NewVBox(
		widget.NewCard("Network", "", container.NewVBox(
			timeoutLabel,
			timeoutSlider,
			concurrentLabel,
			concurrentSlider,
		)),
		widget.NewCard("Logging", "", container.NewVBox(
			debugCheck,
		)),
		widget.NewCard("Paths", "", container.NewVBox(
			widget.NewLabel("Java Path:"),
			javaPathEntry,
			widget.NewLabel("Prism Path:"),
			prismPathEntry,
		)),
	)
}

// onClose handles settings window close event
func (sw *SettingsWindow) onClose() {
	// Save configuration
	if err := sw.appState.Config.Save(); err != nil {
		dialog.ShowError(err, sw)
		return
	}

	sw.Close()
}

// formatMemory formats memory size for display
func (sw *SettingsWindow) formatMemory(mb int) string {
	if mb >= 1024 {
		return sw.formatInt(mb/1024) + " GB"
	}
	return sw.formatInt(mb) + " MB"
}

// formatTimeout formats timeout for display
func (sw *SettingsWindow) formatTimeout(seconds int) string {
	if seconds >= 60 {
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		if remainingSeconds > 0 {
			return sw.formatInt(minutes) + " min " + sw.formatInt(remainingSeconds) + " sec"
		}
		return sw.formatInt(minutes) + " min"
	}
	return sw.formatInt(seconds) + " sec"
}

// formatInt formats an integer for display
func (sw *SettingsWindow) formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

// Show displays the settings window
func (sw *SettingsWindow) Show() {
	sw.Window.Show()
}