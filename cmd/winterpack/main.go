// TheBoys Launcher - Modern cross-platform Minecraft modpack launcher
// Built with Fyne for native GUI across Windows, macOS, and Linux
//
// This is the main entry point for the TheBoys Launcher application.

package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/storage"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/gui/windows"
	"theboys-launcher/pkg/version"
)

func main() {
	// Create application using the enhanced app package
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Set up error handling
	defer func() {
		if r := recover(); r != nil {
			application.GetLogger().Error("Application panic: %v", r)
			application.Quit()
		}
	}()

	// Run the application
	if err := application.Run(); err != nil {
		application.GetLogger().Error("Application error: %v", err)
		os.Exit(1)
	}
}