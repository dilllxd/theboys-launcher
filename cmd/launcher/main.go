package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/launcher"
)

//go:embed all:frontend/dist
var assets embed.FS

// CLI configuration
type CLIConfig struct {
	CliMode     bool
	ModpackID   string
	ListModpacks bool
	ShowSettings bool
	ShowHelp    bool
}

func parseCLIArgs() CLIConfig {
	config := CLIConfig{}

	flag.BoolVar(&config.CliMode, "cli", false, "Run in CLI/console mode")
	flag.StringVar(&config.ModpackID, "modpack", "", "Select specific modpack ID to install")
	flag.BoolVar(&config.ListModpacks, "list-modpacks", false, "List available modpacks")
	flag.BoolVar(&config.ShowSettings, "settings", false, "Open settings menu")
	flag.BoolVar(&config.ShowHelp, "help", false, "Show help information")

	flag.Parse()

	return config
}

func showHelp() {
	fmt.Println("TheBoys Launcher - Modern Minecraft Modpack Launcher")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  theboys-launcher [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --cli                 Run in CLI/console mode (GUI disabled)")
	fmt.Println("  --modpack <id>        Select specific modpack ID to install")
	fmt.Println("  --list-modpacks       List available modpacks")
	fmt.Println("  --settings            Open settings menu")
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  theboys-launcher                    # Start GUI mode")
	fmt.Println("  theboys-launcher --cli              # Start CLI mode")
	fmt.Println("  theboys-launcher --list-modpacks    # List available modpacks")
	fmt.Println("  theboys-launcher --modpack 123      # Install modpack with ID 123")
	fmt.Println("  theboys-launcher --settings         # Open settings menu")
}

func runCLIMode(config CLIConfig) error {
	fmt.Println("TheBoys Launcher CLI Mode")
	fmt.Println("========================")

	// Create app instance
	appInstance := app.NewApp()

	// Initialize app without GUI
	appInstance.Startup(nil)
	defer appInstance.Shutdown(nil)

	// Handle CLI commands
	if config.ListModpacks {
		return runListModpacks(appInstance)
	}

	if config.ModpackID != "" {
		return runSelectModpack(appInstance, config.ModpackID)
	}

	if config.ShowSettings {
		return runSettings(appInstance)
	}

	// Default CLI behavior - show interactive menu
	return runInteractiveCLI(appInstance)
}

func runListModpacks(appInstance *app.App) error {
	modpacks := appInstance.GetModpacks()

	fmt.Printf("Available Modpacks (%d):\n", len(modpacks))
	fmt.Println(strings.Repeat("=", 50))

	for i, modpack := range modpacks {
		fmt.Printf("%3d. %s\n", i+1, modpack.DisplayName)
		fmt.Printf("     ID: %s\n", modpack.ID)
		fmt.Printf("     Description: %s\n", modpack.Description)
		if modpack.PackURL != "" {
			fmt.Printf("     URL: %s\n", modpack.PackURL)
		}
		if modpack.Default {
			fmt.Printf("     DEFAULT MODPACK\n")
		}
		fmt.Println()
	}

	return nil
}

func runSelectModpack(appInstance *app.App, modpackID string) error {
	fmt.Printf("Selecting modpack %s...\n", modpackID)

	modpack, err := appInstance.SelectModpack(modpackID)
	if err != nil {
		return fmt.Errorf("failed to select modpack: %w", err)
	}

	fmt.Printf("Selected: %s\n", modpack.DisplayName)

	// Create instance
	instance, err := appInstance.CreateInstance(*modpack)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	fmt.Printf("Instance created: %s\n", instance.Name)

	// Ask if user wants to launch
	fmt.Print("Launch instance? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		fmt.Printf("Launching %s...\n", instance.Name)
		return appInstance.LaunchInstance(instance.ID)
	}

	return nil
}

func runSettings(appInstance *app.App) error {
	settings := appInstance.GetSettings()

	fmt.Println("Current Settings:")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("Memory:         %d MB\n", settings.MemoryMB)
	fmt.Printf("Auto Update:    %t\n", settings.AutoUpdate)
	fmt.Printf("Check Updates:  %t\n", settings.CheckForUpdates)
	fmt.Printf("Keep Console:   %t\n", settings.KeepConsoleOpen)
	if settings.JavaPath != "" {
		fmt.Printf("Java Path:      %s\n", settings.JavaPath)
	} else {
		fmt.Printf("Java Path:      Auto-detect\n")
	}
	if settings.LastModpackID != "" {
		fmt.Printf("Last Modpack:   %s\n", settings.LastModpackID)
	}
	fmt.Printf("Window Size:    %dx%d\n", settings.WindowSize.Width, settings.WindowSize.Height)
	fmt.Printf("Theme:          %s\n", settings.Theme)
	fmt.Println()

	fmt.Println("Settings management is currently read-only in CLI mode.")
	fmt.Println("Use the GUI to modify settings.")

	return nil
}

func runInteractiveCLI(appInstance *app.App) error {
	fmt.Println("Interactive CLI Mode (Type 'help' for commands)")
	fmt.Println(strings.Repeat("=", 50))

	for {
		fmt.Print("theboys> ")
		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "help", "?":
			fmt.Println("Available commands:")
			fmt.Println("  list          - List available modpacks")
			fmt.Println("  settings      - Show current settings")
			fmt.Println("  help, ?       - Show this help")
			fmt.Println("  exit, quit    - Exit CLI mode")

		case "list":
			if err := runListModpacks(appInstance); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "settings":
			if err := runSettings(appInstance); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "exit", "quit":
			fmt.Println("Exiting...")
			return nil

		case "":
			// Empty input, continue

		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", input)
		}
	}
}

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

	// Parse CLI arguments
	config := parseCLIArgs()

	// Show help if requested
	if config.ShowHelp {
		showHelp()
		return
	}

	// Run in CLI mode if requested
	if config.CliMode || config.ListModpacks || config.ShowSettings || config.ModpackID != "" {
		if err := runCLIMode(config); err != nil {
			fmt.Printf("CLI Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create an instance of the app structure
	appInstance := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TheBoys Launcher",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appInstance.Startup,
		OnShutdown:       appInstance.Shutdown,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}