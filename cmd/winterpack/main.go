// Winterpack Launcher - Modern cross-platform Minecraft modpack launcher
// Built with Fyne for native GUI across Windows, macOS, and Linux
//
// This is the main entry point for the Winterpack Launcher application.

package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/storage"

	"winterpack-launcher/internal/app"
	"winterpack-launcher/internal/config"
	"winterpack-launcher/internal/logging"
	"winterpack-launcher/internal/gui/windows"
	"winterpack-launcher/pkg/version"
)

func main() {
	// Initialize logging system
	logger := logging.NewLogger()
	defer logger.Close()

	// Log application startup
	logger.Info("Starting Winterpack Launcher v%s", version.Version)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Create Fyne application
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(&app.CustomTheme{})

	// Set application metadata
	fyneApp.Metadata().Name = "Winterpack Launcher"
	fyneApp.Metadata().ID = "com.winterpack.launcher"
	fyneApp.Metadata().Version = version.Version
	fyneApp.Metadata().Description = "Cross-platform Minecraft modpack launcher"
	fyneApp.Metadata().CustomIcon = storage.NewFileURI("assets/icons/app.png")

	// Initialize application state
	appState := app.NewState(cfg, logger)

	// Create and show main window
	mainWindow := windows.NewMainWindow(fyneApp, appState)
	mainWindow.Show()

	// Run the application
	logger.Info("Winterpack Launcher started successfully")
	fyneApp.Run()

	// Cleanup on exit
	logger.Info("Winterpack Launcher shutting down")
}