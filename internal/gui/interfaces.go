// Package gui provides interfaces for GUI components
package gui

import (
	"fyne.io/fyne/v2"

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
)

// ApplicationContext provides access to application services without creating import cycles
type ApplicationContext interface {
	GetConfig() *config.Config
	GetLogger() *logging.Logger
	GetMainWindow() fyne.Window
}

// Simple context implementation for screens
type ScreenContext struct {
	cfg     *config.Config
	logger  *logging.Logger
	window  fyne.Window
}

// NewScreenContext creates a new screen context
func NewScreenContext(cfg *config.Config, logger *logging.Logger, window fyne.Window) *ScreenContext {
	return &ScreenContext{
		cfg:    cfg,
		logger: logger,
		window: window,
	}
}

// GetConfig returns the configuration
func (sc *ScreenContext) GetConfig() *config.Config {
	return sc.cfg
}

// GetLogger returns the logger
func (sc *ScreenContext) GetLogger() *logging.Logger {
	return sc.logger
}

// GetMainWindow returns the main window
func (sc *ScreenContext) GetMainWindow() fyne.Window {
	return sc.window
}