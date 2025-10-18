package windows

import (
	"fyne.io/fyne/v2"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/config"
	"theboys-launcher/internal/gui/widgets"
)

// SettingsWindow represents the settings window
type SettingsWindow struct {
	fyne.Window
	appState       *app.State
	settingsWidget *widgets.SettingsWidget
	settingsManager *config.SettingsManager
}

// NewSettingsWindow creates a new settings window
func NewSettingsWindow(appState *app.State) *SettingsWindow {
	window := fyne.CurrentApp().NewWindow("Settings")
	window.Resize(fyne.NewSize(700, 600))
	window.SetFixedSize(false)

	// Create settings manager
	settingsManager, err := config.NewSettingsManager(appState.Logger)
	if err != nil {
		appState.Logger.Error("Failed to create settings manager: %v", err)
		// Fallback to simple error window
		window.SetContent(widgets.NewErrorWidget("Settings Error", "Failed to initialize settings manager"))
		return &SettingsWindow{
			Window: window,
			appState: appState,
		}
	}

	// Create settings widget
	settingsWidget := widgets.NewSettingsWidget(settingsManager)

	settingsWindow := &SettingsWindow{
		Window:         window,
		appState:       appState,
		settingsWidget: settingsWidget,
		settingsManager: settingsManager,
	}

	// Set content
	window.SetContent(settingsWidget)

	// Setup window close handler
	window.SetCloseIntercept(func() {
		settingsWindow.onClose()
	})

	// Add settings change listener
	settingsManager.AddListener(config.SettingsChangeFunc(func(config *config.Config, changes map[string]interface{}) {
		appState.Logger.Info("Settings changed: %v", changes)
		// Update app state with new config
		appState.Config = config
	}))

	return settingsWindow
}

// onClose handles settings window close event
func (sw *SettingsWindow) onClose() {
	// Settings are automatically saved by the settings manager
	sw.appState.Logger.Info("Settings window closed")
	sw.Close()
}

// Show displays the settings window
func (sw *SettingsWindow) Show() {
	sw.Window.Show()
}

// GetSettingsManager returns the settings manager
func (sw *SettingsWindow) GetSettingsManager() *config.SettingsManager {
	return sw.settingsManager
}