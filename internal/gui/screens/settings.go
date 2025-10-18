package screens

import (
	"fyne.io/fyne/v2"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/config"
	"theboys-launcher/internal/gui/widgets"
)

// SettingsScreen represents the settings screen
type SettingsScreen struct {
	id           string
	title        string
	app          *app.Application
	settingsWidget *widgets.SettingsWidget
	settingsManager *config.SettingsManager
}

// NewSettingsScreen creates a new settings screen
func NewSettingsScreen(application *app.Application) *SettingsScreen {
	// Create settings manager
	settingsManager, err := config.NewSettingsManager(application.GetLogger())
	if err != nil {
		application.GetLogger().Error("Failed to create settings manager: %v", err)
		return nil
	}

	// Create settings widget
	settingsWidget := widgets.NewSettingsWidget(settingsManager)

	screen := &SettingsScreen{
		id:             "settings",
		title:          "Settings",
		app:            application,
		settingsWidget: settingsWidget,
		settingsManager: settingsManager,
	}

	// Add settings change listener
	settingsManager.AddListener(config.SettingsChangeFunc(func(config *config.Config, changes map[string]interface{}) {
		application.GetLogger().Info("Settings changed: %v", changes)
		// Update application config
		application.GetConfig().MemoryMB = config.MemoryMB
		application.GetConfig().Theme = config.Theme
		// ... update other config fields as needed
	}))

	return screen
}

// GetID returns the screen ID
func (s *SettingsScreen) GetID() string {
	return s.id
}

// GetTitle returns the screen title
func (s *SettingsScreen) GetTitle() string {
	return s.title
}

// GetContent returns the screen content
func (s *SettingsScreen) GetContent() fyne.CanvasObject {
	return s.settingsWidget
}

// OnShow is called when the screen is shown
func (s *SettingsScreen) OnShow() error {
	s.app.GetLogger().Info("Settings screen shown")
	return nil
}

// OnHide is called when the screen is hidden
func (s *SettingsScreen) OnHide() error {
	s.app.GetLogger().Info("Settings screen hidden")
	return nil
}

// GetSettingsManager returns the settings manager
func (s *SettingsScreen) GetSettingsManager() *config.SettingsManager {
	return s.settingsManager
}