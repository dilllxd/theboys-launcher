package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/gui/screens"
	"theboys-launcher/internal/gui/widgets"
	"theboys-launcher/internal/logging"
)

// Screen represents a screen in the application
type Screen interface {
	GetID() string
	GetTitle() string
	GetContent() fyne.CanvasObject
	OnShow() error
	OnHide() error
}

// NavigationManager handles screen navigation
type NavigationManager struct {
	app        *Application
	screens    map[string]Screen
	current    string
	navigation *container.AppTabs
	logger     *logging.Logger
}

// NewNavigationManager creates a new navigation manager
func NewNavigationManager(app *Application) *NavigationManager {
	var logger *logging.Logger
	if app != nil {
		logger = app.GetLogger()
	} else {
		logger = logging.NewLogger()
	}

	return &NavigationManager{
		app:     app,
		screens: make(map[string]Screen),
		logger:  logger,
	}
}

// RegisterScreen registers a screen with the navigation manager
func (nm *NavigationManager) RegisterScreen(screen Screen) {
	nm.screens[screen.GetID()] = screen
	nm.logger.Info("Registered screen: %s", screen.GetID())
}

// NavigateTo navigates to the specified screen
func (nm *NavigationManager) NavigateTo(screenID string) error {
	screen, exists := nm.screens[screenID]
	if !exists {
		return fyne.NewError("Screen not found: " + screenID)
	}

	// Hide current screen
	if nm.current != "" {
		if currentScreen, exists := nm.screens[nm.current]; exists {
			if err := currentScreen.OnHide(); err != nil {
				nm.logger.Error("Error hiding screen %s: %v", nm.current, err)
			}
		}
	}

	// Show new screen
	if err := screen.OnShow(); err != nil {
		nm.logger.Error("Error showing screen %s: %v", screenID, err)
		return err
	}

	nm.current = screenID
	nm.app.state.SetCurrentScreen(screenID)
	nm.logger.Info("Navigated to screen: %s", screenID)

	return nil
}

// GetCurrentScreen returns the current screen
func (nm *NavigationManager) GetCurrentScreen() (Screen, error) {
	if nm.current == "" {
		return nil, fyne.NewError("No current screen")
	}

	screen, exists := nm.screens[nm.current]
	if !exists {
		return nil, fyne.NewError("Current screen not found: " + nm.current)
	}

	return screen, nil
}

// GetTabContainer returns the tab container for navigation
func (nm *NavigationManager) GetTabContainer() *container.AppTabs {
	if nm.navigation == nil {
		tabs := container.NewAppTabs()
		nm.navigation = tabs

		// Set up tab change handler
		tabs.OnSelected = func(tab *container.TabItem) {
			if tab != nil && tab.Content != nil {
				// Find screen by content (simplified for now)
				for id, screen := range nm.screens {
					if screen.GetContent() == tab.Content {
						nm.NavigateTo(id)
						break
					}
				}
			}
		}
	}
	return nm.navigation
}

// CreateNavigationPanel creates the main navigation panel
func (nm *NavigationManager) CreateNavigationPanel() fyne.CanvasObject {
	tabs := nm.GetTabContainer()

	// Add main screens as tabs
	tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), nm.createHomeScreen()))
	tabs.Append(container.NewTabItemWithIcon("Modpacks", theme.ComputerIcon(), nm.createModpacksScreen()))
	tabs.Append(container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), nm.createSettingsScreen()))

	return tabs
}

// createHomeScreen creates the home screen content
func (nm *NavigationManager) createHomeScreen() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewCard("Welcome", "TheBoys Launcher", container.NewVBox(
			widget.NewLabel("Welcome to TheBoys Launcher!"),
			widget.NewLabel("Select a tab to get started."),
		)),
	)
}

// createModpacksScreen creates the modpacks screen content
func (nm *NavigationManager) createModpacksScreen() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewCard("Modpacks", "Available Modpacks", container.NewVBox(
			widget.NewLabel("Modpack selection will be implemented in Phase 4."),
			widget.NewLabel("This screen will display available modpacks."),
		)),
	)
}

// createSettingsScreen creates the settings screen content
func (nm *NavigationManager) createSettingsScreen() fyne.CanvasObject {
	if nm.app == nil {
		return widget.NewLabel("Settings unavailable - application not initialized")
	}

	// Create the settings screen
	settingsScreen := screens.NewSettingsScreen(nm.app)
	if settingsScreen == nil {
		return widget.NewLabel("Failed to create settings screen")
	}

	// Register the screen
	nm.RegisterScreen(settingsScreen)

	return settingsScreen.GetContent()
}

// CreateStatusBar creates a status bar component
func (nm *NavigationManager) CreateStatusBar() *widgets.StatusBar {
	statusBar := widgets.NewStatusBar(nm.app.state)
	return statusBar
}

// RefreshNavigation refreshes the navigation components
func (nm *NavigationManager) RefreshNavigation() {
	if nm.navigation != nil {
		nm.navigation.Refresh()
	}
}

// SetTabEnabled enables or disables a tab
func (nm *NavigationManager) SetTabEnabled(index int, enabled bool) {
	if nm.navigation != nil && index >= 0 && index < len(nm.navigation.Items) {
		// Note: Fyne doesn't directly support disabling tabs
		// This would require custom implementation
		nm.logger.Info("Set tab %d enabled: %v (not implemented)", index, enabled)
	}
}