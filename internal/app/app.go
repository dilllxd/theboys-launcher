package app

import (
	"context"
	"fmt"
	"path/filepath"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/gui"
	"theboys-launcher/internal/config"
	"theboys-launcher/internal/launcher"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// App represents the main application
type App struct {
	ctx             context.Context
	gui             *gui.GUI
	config          *config.Manager
	modpackManager  *config.ModpackManager
	launcherManager *launcher.ModpackManager
	javaManager     *launcher.JavaManager
	platform        platform.Platform
	logger          *logging.Logger
}

// NewApp creates a new application instance
func NewApp() *App {
	// Initialize logger
	logger := logging.NewLogger()

	// Get platform implementation
	platformImpl := platform.NewPlatform()

	// Initialize configuration manager
	configManager := config.NewManager(platformImpl, logger)

	// Initialize modpack configuration manager
	modpackConfigManager := config.NewModpackManager(platformImpl, logger)

	// Initialize launcher modpack manager
	launcherModpackManager := launcher.NewModpackManager(configManager, modpackConfigManager, platformImpl, logger)

	// Initialize Java manager
	javaManager := launcher.NewJavaManager(platformImpl, logger)

	// Initialize GUI
	gui := gui.NewGUI(configManager, platformImpl, logger)

	return &App{
		gui:             gui,
		config:          configManager,
		modpackManager:  modpackConfigManager,
		launcherManager: launcherModpackManager,
		javaManager:     javaManager,
		platform:        platformImpl,
		logger:          logger,
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("TheBoys Launcher starting up")

	// Initialize configuration
	if err := a.config.Initialize(); err != nil {
		a.logger.Error("Failed to initialize configuration: %v", err)
		return
	}

	// Load settings
	if err := a.config.LoadSettings(); err != nil {
		a.logger.Warn("Could not load settings, using defaults: %v", err)
	}

	// Initialize logging system
	appDataDir, err := a.platform.GetAppDataDir()
	if err != nil {
		a.logger.Error("Failed to get app data directory: %v", err)
		return
	}

	logDir := a.logger.GetLogDir()
	if logDir == "" {
		logDir = filepath.Join(appDataDir, "logs")
	}

	if err := a.logger.Initialize(logDir); err != nil {
		a.logger.Error("Failed to initialize logging: %v", err)
		return
	}

	// Load modpacks
	a.logger.Info("Loading modpacks...")
	_, err = a.launcherManager.LoadModpacks(true) // Fetch remote modpacks
	if err != nil {
		a.logger.Error("Failed to load modpacks: %v", err)
		// Continue with application startup even if modpacks fail
	} else {
		modpacks := a.launcherManager.GetModpacks()
		a.logger.Info("Loaded %d modpack(s)", len(modpacks))
	}

	// Detect Java installations
	a.logger.Info("Detecting Java installations...")
	javaInstallations, err := a.javaManager.DetectJavaInstallations()
	if err != nil {
		a.logger.Warn("Failed to detect Java installations: %v", err)
	} else {
		a.logger.Info("Found %d Java installation(s)", len(javaInstallations))
		for _, java := range javaInstallations {
			a.logger.Debug("Java %s at %s (%s)", java.Version, java.Path, map[bool]string{true: "JDK", false: "JRE"}[java.IsJDK])
		}
	}

	// Initialize GUI components
	a.gui.Initialize(ctx)

	a.logger.Info("TheBoys Launcher started successfully")
}

// Shutdown is called when the app is shutting down
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("TheBoys Launcher shutting down")

	// Save settings
	if err := a.config.SaveSettings(); err != nil {
		a.logger.Error("Failed to save settings: %v", err)
	}

	// Cleanup
	a.gui.Cleanup()
}

// Greet returns a greeting for the given name (placeholder for testing)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to TheBoys Launcher!", name)
}

// GetModpacks returns the available modpacks
func (a *App) GetModpacks() []types.Modpack {
	return a.launcherManager.GetModpacks()
}

// SelectModpack selects a modpack by ID
func (a *App) SelectModpack(modpackID string) (*types.Modpack, error) {
	return a.launcherManager.SelectModpack(modpackID)
}

// RefreshModpacks refreshes the modpack list from remote sources
func (a *App) RefreshModpacks() error {
	return a.launcherManager.RefreshModpacks()
}

// GetSettings returns the current application settings
func (a *App) GetSettings() *types.LauncherSettings {
	return a.config.GetSettings()
}

// GetJavaInstallations returns detected Java installations
func (a *App) GetJavaInstallations() ([]types.JavaInstallation, error) {
	return a.javaManager.DetectJavaInstallations()
}

// GetBestJavaInstallation returns the best Java installation for a Minecraft version
func (a *App) GetBestJavaInstallation(mcVersion string) (*types.JavaInstallation, error) {
	return a.javaManager.GetBestJavaInstallation(mcVersion)
}

// GetJavaVersionForMinecraft returns the recommended Java version for a Minecraft version
func (a *App) GetJavaVersionForMinecraft(mcVersion string) string {
	return a.javaManager.GetJavaVersionForMinecraft(mcVersion)
}

// DownloadJava downloads and installs a Java runtime
func (a *App) DownloadJava(javaVersion string, installDir string) error {
	return a.javaManager.DownloadJava(javaVersion, installDir, nil)
}