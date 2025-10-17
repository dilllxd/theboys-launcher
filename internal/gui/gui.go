package gui

import (
	"context"

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

// GUI represents the user interface manager
type GUI struct {
	config     *config.Manager
	platform   platform.Platform
	logger     logging.Logger
	ctx        context.Context
	initialized bool
}

// NewGUI creates a new GUI instance
func NewGUI(config *config.Manager, platform platform.Platform, logger logging.Logger) *GUI {
	return &GUI{
		config: config,
		platform: platform,
		logger: logger,
	}
}

// Initialize sets up the GUI components
func (g *GUI) Initialize(ctx context.Context) {
	g.ctx = ctx
	g.initialized = true
	g.logger.Info("GUI initialized")
}

// Cleanup performs GUI cleanup operations
func (g *GUI) Cleanup() {
	if g.initialized {
		g.logger.Info("GUI cleanup completed")
		g.initialized = false
	}
}

// IsInitialized returns whether the GUI has been initialized
func (g *GUI) IsInitialized() bool {
	return g.initialized
}