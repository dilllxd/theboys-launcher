package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/launcher"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Handle cleanup after update (internal flag, not shown in help)
	if len(os.Args) >= 2 && os.Args[1] == "--cleanup-after-update" && len(os.Args) >= 4 {
		oldExe := os.Args[2]
		newExe := os.Args[3]

		updater := launcher.NewUpdater(nil, nil) // Platform and logger not needed for cleanup
		if err := updater.PerformUpdateCleanup(oldExe, newExe); err != nil {
			fmt.Printf("Failed to cleanup after update: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Update cleanup completed successfully")
		os.Exit(0)
	}

	// Create an instance of the app structure
	app := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TheBoys Launcher",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}