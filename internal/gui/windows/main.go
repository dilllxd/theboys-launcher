// Package windows provides window implementations for the TheBoys Launcher
package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/gui/widgets"
)

// MainWindow represents the main application window
type MainWindow struct {
	fyne.Window
	appState     *app.State
	content      fyne.CanvasObject
	navigation   *app.NavigationManager
	statusBar    *widgets.StatusBar
	loading      *widgets.LoadingOverlay
}

// NewMainWindow creates a new main window
func NewMainWindow(fyneApp fyne.App, appState *app.State) *MainWindow {
	window := fyneApp.NewWindow("TheBoys Launcher")

	// Set window properties
	window.SetIcon(fyne.NewStaticResource("app", []byte{})) // Will be replaced with actual icon
	window.Resize(fyne.NewSize(float32(appState.Config.WindowWidth), float32(appState.Config.WindowHeight)))
	window.SetFixedSize(false)
	window.CenterOnScreen()

	mainWindow := &MainWindow{
		Window:   window,
		appState: appState,
	}

	// Create navigation manager
	// For now, we'll create a simple one and enhance it later
	mainWindow.navigation = app.NewNavigationManager(nil)
	mainWindow.appState = appState

	// Create status bar
	mainWindow.statusBar = widgets.NewStatusBar(appState)

	// Create loading overlay
	mainWindow.loading = widgets.CreateLoadingOverlay("Loading...", window.Canvas())

	// Create main content
	mainWindow.createContent()

	// Setup window close handler
	window.SetCloseIntercept(func() {
		mainWindow.onClose()
	})

	return mainWindow
}

// createContent creates the main window content
func (mw *MainWindow) createContent() {
	// Create welcome header
	welcomeLabel := widget.NewLabelWithStyle(
		"Welcome to TheBoys Launcher",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	versionLabel := widget.NewLabelWithStyle(
		"Version 2.0.0 - Modern Cross-Platform Launcher",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Create main content area with tabs
	navigationTabs := mw.navigation.CreateNavigationPanel()

	// Create main layout
	mainContent := container.NewBorder(
		nil, // Top
		mw.statusBar, // Bottom
		nil, // Left
		nil, // Right
		container.NewVBox( // Center
			widget.NewSeparator(),
			welcomeLabel,
			versionLabel,
			widget.NewSeparator(),
			navigationTabs,
		),
	)

	mw.content = mainContent
	mw.SetContent(mainContent)
}

// createModpackSection creates the modpack selection section
func (mw *MainWindow) createModpackSection() fyne.CanvasObject {
	sectionLabel := widget.NewLabelWithStyle(
		"Select Modpack",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Placeholder for modpack selection
	modpackContainer := container.NewVBox(
		widget.NewLabel("Modpack selection will be implemented in Phase 4"),
		widget.NewLabel("This will show available modpacks with cards"),
	)

	return container.NewVBox(
		sectionLabel,
		modpackContainer,
	)
}

// createActionButtons creates the action buttons section
func (mw *MainWindow) createActionButtons() fyne.CanvasObject {
	// Create buttons
	settingsButton := widget.NewButton("Settings", func() {
		mw.showSettings()
	})

	checkUpdatesButton := widget.NewButton("Check for Updates", func() {
		mw.checkForUpdates()
	})

	aboutButton := widget.NewButton("About", func() {
		mw.showAbout()
	})

	quitButton := widget.NewButton("Quit", func() {
		mw.onClose()
	})

	// Create button container
	buttonContainer := container.NewHBox(
		settingsButton,
		checkUpdatesButton,
		aboutButton,
		quitButton,
	)

	return container.NewVBox(
		widget.NewSeparator(),
		buttonContainer,
	)
}

// showSettings displays the settings window
func (mw *MainWindow) showSettings() {
	settingsWindow := NewSettingsWindow(mw.appState)
	settingsWindow.Show()
}

// checkForUpdates checks for application updates
func (mw *MainWindow) checkForUpdates() {
	// Placeholder for update checking
	mw.appState.Logger.Info("Checking for updates (to be implemented in Phase 9)")
}

// showAbout displays the about dialog
func (mw *MainWindow) showAbout() {
	content := container.NewVBox(
		widget.NewLabel("TheBoys Launcher v2.0.0"),
		widget.NewLabel("Modern cross-platform Minecraft modpack launcher"),
		widget.NewSeparator(),
		widget.NewLabel("Built with Fyne for native GUI"),
		widget.NewLabel("Cross-platform support: Windows, macOS, Linux"),
		widget.NewSeparator(),
		widget.NewLabel("© 2024 TheBoys Launcher"),
	)

	dialog := widget.NewModalPopUp(content, mw.Canvas())
	dialog.Resize(fyne.NewSize(400, 300))

	// Add close button
	closeButton := widget.NewButton("Close", func() {
		dialog.Hide()
	})

	content.Add(closeButton)
	dialog.Show()
}

// onClose handles window close event
func (mw *MainWindow) onClose() {
	// Save window size
	mw.appState.Config.WindowWidth = int(mw.Canvas().Size().Width)
	mw.appState.Config.WindowHeight = int(mw.Canvas().Size().Height)

	// Save configuration
	if err := mw.appState.Config.Save(); err != nil {
		mw.appState.Logger.Error("Failed to save configuration: %v", err)
	}

	mw.appState.Logger.Info("Application shutting down")
	mw.Close()
}

// Show displays the main window
func (mw *MainWindow) Show() {
	mw.Window.Show()
}

// ShowLoading shows the loading overlay
func (mw *MainWindow) ShowLoading(message string) {
	mw.loading.SetMessage(message)
	mw.loading.Show()
}

// HideLoading hides the loading overlay
func (mw *MainWindow) HideLoading() {
	mw.loading.Hide()
}

// SetStatus updates the status bar
func (mw *MainWindow) SetStatus(message string) {
	mw.statusBar.SetStatus(message)
}

// SetProgress updates the progress bar
func (mw *MainWindow) SetProgress(progress float64) {
	mw.statusBar.SetProgress(progress)
}

// NavigateTo navigates to a specific screen
func (mw *MainWindow) NavigateTo(screenID string) error {
	return mw.navigation.NavigateTo(screenID)
}