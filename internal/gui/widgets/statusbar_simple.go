// Package widgets provides custom GUI widgets for TheBoys Launcher
package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SimpleStatusBar represents a simple status bar widget
type SimpleStatusBar struct {
	widget.BaseWidget
	statusLabel    *widget.Label
	timeLabel      *widget.Label
	progressBar    *widget.ProgressBar
	statusContainer *container.HBox
	lastUpdate     time.Time
}

// NewSimpleStatusBar creates a new simple status bar
func NewSimpleStatusBar() *SimpleStatusBar {
	sb := &SimpleStatusBar{
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
func (sb *SimpleStatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sb.statusContainer)
}

// SetStatus updates the status message
func (sb *SimpleStatusBar) SetStatus(message string) {
	sb.statusLabel.SetText(message)
	sb.lastUpdate = time.Now()
}

// SetProgress updates the progress bar
func (sb *SimpleStatusBar) SetProgress(progress float64) {
	sb.progressBar.SetValue(progress)
}

// updateTime updates the time display
func (sb *SimpleStatusBar) updateTime() {
	now := time.Now()
	timeStr := now.Format("3:04 PM")
	sb.timeLabel.SetText(timeStr)
}

// startUpdateTime starts the time update loop
func (sb *SimpleStatusBar) startUpdateTime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sb.updateTime()
		}
	}
}