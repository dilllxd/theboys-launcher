// Package app provides core application functionality
package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/gui/windows"
)

// Application represents the main application
type Application struct {
	fyneApp   fyne.App
	state     *State
	mainWin   *windows.MainWindow
	logger    *logging.Logger
	config    *config.Config
}

// NewApplication creates a new application instance
func NewApplication() (*Application, error) {
	// Initialize logging
	logger := logging.NewLogger()
	logger.Info("Initializing TheBoys Launcher v2.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		return nil, err
	}

	// Create Fyne application
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(&CustomTheme{})

	// Set application metadata
	fyneApp.Metadata().Name = "TheBoys Launcher"
	fyneApp.Metadata().ID = "com.theboys.launcher"
	fyneApp.Metadata().Description = "Modern cross-platform Minecraft modpack launcher"

	// Create application state
	state := NewState(cfg, logger)

	application := &Application{
		fyneApp: fyneApp,
		state:   state,
		logger:  logger,
		config:  cfg,
	}

	return application, nil
}

// Run starts the application
func (a *Application) Run() error {
	// Create main window
	a.mainWin = windows.NewMainWindow(a.fyneApp, a.state)

	// Show main window
	a.mainWin.Show()

	a.logger.Info("TheBoys Launcher started successfully")

	// Run the application
	a.fyneApp.Run()

	// Cleanup
	a.logger.Info("TheBoys Launcher shutting down")
	a.state.SetCurrentScreen("shutdown")

	return nil
}

// GetState returns the application state
func (a *Application) GetState() *State {
	return a.state
}

// GetMainWindow returns the main window
func (a *Application) GetMainWindow() *windows.MainWindow {
	return a.mainWin
}

// GetLogger returns the application logger
func (a *Application) GetLogger() *logging.Logger {
	return a.logger
}

// GetConfig returns the application configuration
func (a *Application) GetConfig() *config.Config {
	return a.config
}

// Quit gracefully shuts down the application
func (a *Application) Quit() {
	a.logger.Info("Application quit requested")

	// Save configuration
	if err := a.config.Save(); err != nil {
		a.logger.Error("Failed to save configuration: %v", err)
	}

	// Close logger
	a.logger.Close()

	// Quit Fyne app
	a.fyneApp.Quit()
}

// ShowAboutDialog displays the about dialog
func (a *Application) ShowAboutDialog() {
	content := container.NewVBox(
		widget.NewLabelWithStyle("TheBoys Launcher", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Version 2.0.0"),
		widget.NewSeparator(),
		widget.NewLabel("Modern cross-platform Minecraft modpack launcher"),
		widget.NewLabel("Built with Fyne for native GUI"),
		widget.NewLabel("Cross-platform support: Windows, macOS, Linux"),
		widget.NewSeparator(),
		widget.NewLabel("© 2024 TheBoys Launcher"),
	)

	dialog := widget.NewModalPopUp(content, a.mainWin.Canvas())
	dialog.Resize(fyne.NewSize(400, 300))

	closeBtn := widget.NewButton("Close", func() {
		dialog.Hide()
	})
	content.Add(closeBtn)

	dialog.Show()
}

// ShowErrorDialog displays an error dialog
func (a *Application) ShowErrorDialog(title, message string) {
	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabel(message),
			widget.NewButton("OK", func() { /* Will be overridden */ }),
		),
		a.mainWin.Canvas(),
	)

	dialog.Resize(fyne.NewSize(400, 200))

	// Override close button
	buttons := dialog.(*widget.PopUp).Content.(*container.VBox).Objects
	closeBtn := buttons[len(buttons)-1].(*widget.Button)
	closeBtn.OnTapped = func() { dialog.Hide() }

	dialog.Show()
}