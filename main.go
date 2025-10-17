package main

import (
	"embed"
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"theboys-launcher/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

// App is a wrapper around the internal app for Wails bindings
type App struct {
	internalApp *app.App
}

// NewApp creates a new app wrapper
func NewApp() *App {
	return &App{
		internalApp: app.NewApp(),
	}
}

// Startup passes through to internal app
func (a *App) Startup(ctx context.Context) {
	a.internalApp.Startup(ctx)
}

// Shutdown passes through to internal app
func (a *App) Shutdown(ctx context.Context) {
	a.internalApp.Shutdown(ctx)
}

// Wrapper methods for frontend binding
func (a *App) GetModpacks() interface{} {
	return a.internalApp.GetModpacks()
}

func (a *App) GetInstances() interface{} {
	instances, _ := a.internalApp.GetInstances()
	return instances
}

func (a *App) GetSettings() interface{} {
	return a.internalApp.GetSettings()
}

func (a *App) GetJavaInstallations() (interface{}, error) {
	return a.internalApp.GetJavaInstallations()
}

func (a *App) SelectModpack(modpackID string) (interface{}, error) {
	return a.internalApp.SelectModpack(modpackID)
}

func (a *App) RefreshModpacks() error {
	return a.internalApp.RefreshModpacks()
}

func (a *App) CreateInstance(modpack interface{}) (interface{}, error) {
	// This would need proper type conversion
	return nil, nil
}

func (a *App) GetInstance(instanceID string) (interface{}, error) {
	return a.internalApp.GetInstance(instanceID)
}

func (a *App) LaunchInstance(instanceID string) error {
	return a.internalApp.LaunchInstance(instanceID)
}

func (a *App) DeleteInstance(instanceID string) error {
	return a.internalApp.DeleteInstance(instanceID)
}

func (a *App) GetBestJavaInstallation(mcVersion string) (interface{}, error) {
	return a.internalApp.GetBestJavaInstallation(mcVersion)
}

func (a *App) GetJavaVersionForMinecraft(mcVersion string) string {
	return a.internalApp.GetJavaVersionForMinecraft(mcVersion)
}

func (a *App) DownloadJava(javaVersion string, installDir string) error {
	return a.internalApp.DownloadJava(javaVersion, installDir)
}

func (a *App) IsPrismInstalled() bool {
	return a.internalApp.IsPrismInstalled()
}

func (a *App) InstallModpackWithPackwiz(instanceID string, progressCallback func(float64)) error {
	return a.internalApp.InstallModpackWithPackwiz(instanceID, progressCallback)
}

func (a *App) CheckModpackUpdate(instanceID string) (bool, string, string, error) {
	return a.internalApp.CheckModpackUpdate(instanceID)
}

func (a *App) UpdateModpack(instanceID string, progressCallback func(float64)) error {
	return a.internalApp.UpdateModpack(instanceID, progressCallback)
}

func (a *App) GetLWJGLVersionForMinecraft(mcVersion string) (interface{}, error) {
	return a.internalApp.GetLWJGLVersionForMinecraft(mcVersion)
}

func (a *App) CheckForUpdates() (interface{}, error) {
	return a.internalApp.CheckForUpdates()
}

func (a *App) DownloadUpdate(downloadURL string, progressCallback func(float64)) (string, error) {
	return a.internalApp.DownloadUpdate(downloadURL, progressCallback)
}

func (a *App) InstallUpdate(updatePath string) error {
	return a.internalApp.InstallUpdate(updatePath)
}

func (a *App) GetVersion() string {
	return a.internalApp.GetVersion()
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

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
	})

	if err != nil {
		panic("Error during application startup: " + err.Error())
	}
}