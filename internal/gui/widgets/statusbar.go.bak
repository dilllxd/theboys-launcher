// Package widgets provides custom GUI widgets for TheBoys Launcher
package widgets

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/app"
)

// StatusBar represents a status bar widget
type StatusBar struct {
	widget.BaseWidget
	state          *app.State
	statusLabel    *widget.Label
	timeLabel      *widget.Label
	progressBar    *widget.ProgressBar
	statusContainer *container.HBox
	lastUpdate     time.Time
}

// NewStatusBar creates a new status bar
func NewStatusBar(state *app.State) *StatusBar {
	sb := &StatusBar{
		state:       state,
		lastUpdate:  time.Now(),
	}

	sb.ExtendBaseWidget(sb)

	// Create status components
	sb.statusLabel = widget.NewLabel("Ready")
	sb.timeLabel = widget.NewLabel("")
	sb.progressBar = widget.NewProgressBar()

	// Set initial state
	sb.progressBar.SetValue(0)
	sb.updateTime()

	// Create status container
	sb.statusContainer = container.NewHBox(
		sb.statusLabel,
		widget.NewSeparator(),
		sb.progressBar,
		widget.NewSeparator(),
		sb.timeLabel,
	)

	// Start time update goroutine
	go sb.startUpdateTime()

	return sb
}

// CreateRenderer creates the renderer for the status bar
func (sb *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sb.statusContainer)
}

// SetStatus updates the status message
func (sb *StatusBar) SetStatus(message string) {
	sb.statusLabel.SetText(message)
	sb.lastUpdate = time.Now()
}

// SetProgress updates the progress bar
func (sb *StatusBar) SetProgress(progress float64) {
	sb.progressBar.SetValue(progress)
}

// SetProgressWithStatus updates both progress and status
func (sb *StatusBar) SetProgressWithStatus(progress float64, status string) {
	sb.SetProgress(progress)
	sb.SetStatus(status)
}

// updateTime updates the time display
func (sb *StatusBar) updateTime() {
	now := time.Now()
	timeStr := now.Format("3:04 PM")
	sb.timeLabel.SetText(timeStr)
}

// startUpdateTime starts the time update loop
func (sb *StatusBar) startUpdateTime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sb.updateTime()
		}
	}
}

// SetLoadingState sets the status bar to loading state
func (sb *StatusBar) SetLoadingState(message string) {
	sb.SetStatus(message)
	sb.SetProgress(0)
}

// SetCompleteState sets the status bar to complete state
func (sb *StatusBar) SetCompleteState(message string) {
	sb.SetStatus(message)
	sb.SetProgress(1.0)
}

// SetErrorState sets the status bar to error state
func (sb *StatusBar) SetErrorState(message string) {
	sb.SetStatus("Error: " + message)
	sb.SetProgress(0)
}

// GetStatus returns the current status message
func (sb *StatusBar) GetStatus() string {
	return sb.statusLabel.Text
}

// GetProgress returns the current progress value
func (sb *StatusBar) GetProgress() float64 {
	return sb.progressBar.Value
}