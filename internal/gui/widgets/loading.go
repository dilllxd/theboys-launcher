package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// LoadingWidget represents a loading animation widget
type LoadingWidget struct {
	widget.BaseWidget
	label      *widget.Label
	progress   *widget.ProgressBarInfinite
	container   *container.VBox
	visible    bool
}

// NewLoadingWidget creates a new loading widget
func NewLoadingWidget(message string) *LoadingWidget {
	lw := &LoadingWidget{}
	lw.ExtendBaseWidget(lw)

	lw.label = widget.NewLabel(message)
	lw.label.Alignment = fyne.TextAlignCenter
	lw.progress = widget.NewProgressBarInfinite()
	lw.container = container.NewVBox(
		lw.progress,
		lw.label,
	)

	return lw
}

// CreateRenderer creates the renderer for the loading widget
func (lw *LoadingWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(lw.container)
}

// SetMessage updates the loading message
func (lw *LoadingWidget) SetMessage(message string) {
	lw.label.SetText(message)
}

// Show makes the loading widget visible
func (lw *LoadingWidget) Show() {
	lw.visible = true
	lw.progress.Start()
	lw.Refresh()
}

// Hide makes the loading widget invisible
func (lw *LoadingWidget) Hide() {
	lw.visible = false
	lw.progress.Stop()
	lw.Refresh()
}

// IsVisible returns whether the loading widget is visible
func (lw *LoadingWidget) IsVisible() bool {
	return lw.visible
}

// CreateLoadingOverlay creates a full-screen loading overlay
func CreateLoadingOverlay(message string, parent fyne.Canvas) *LoadingOverlay {
	return &LoadingOverlay{
		message: message,
		parent:  parent,
		visible: false,
	}
}

// LoadingOverlay represents a full-screen loading overlay
type LoadingOverlay struct {
	message string
	parent  fyne.Canvas
	overlay *widget.PopUp
	loading *LoadingWidget
	visible bool
}

// Show shows the loading overlay
func (lo *LoadingOverlay) Show() {
	if lo.visible {
		return
	}

	// Create loading widget
	lo.loading = NewLoadingWidget(lo.message)

	// Create overlay content
	content := container.NewCenter(
		container.NewVBox(
			widget.NewCard("", "", lo.loading),
		),
	)

	// Create pop-up overlay
	lo.overlay = widget.NewModalPopUp(content, lo.parent)
	lo.overlay.Resize(lo.parent.Size())

	// Show loading animation
	lo.loading.Show()
	lo.visible = true
}

// Hide hides the loading overlay
func (lo *LoadingOverlay) Hide() {
	if !lo.visible {
		return
	}

	if lo.loading != nil {
		lo.loading.Hide()
	}

	if lo.overlay != nil {
		lo.overlay.Hide()
	}

	lo.visible = false
}

// SetMessage updates the overlay message
func (lo *LoadingOverlay) SetMessage(message string) {
	lo.message = message
	if lo.loading != nil {
		lo.loading.SetMessage(message)
	}
}

// IsVisible returns whether the overlay is visible
func (lo *LoadingOverlay) IsVisible() bool {
	return lo.visible
}