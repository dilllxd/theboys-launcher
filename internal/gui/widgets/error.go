package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ErrorWidget displays an error message
type ErrorWidget struct {
	widget.BaseWidget
	content *fyne.Container
}

// NewErrorWidget creates a new error widget
func NewErrorWidget(title, message string) *ErrorWidget {
	ew := &ErrorWidget{}
	ew.ExtendBaseWidget(ew)

	// Create error content
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	icon := widget.NewIcon(theme.ErrorIcon())

	content := container.NewVBox(
		container.NewHBox(
			icon,
			titleLabel,
		),
		widget.NewSeparator(),
		messageLabel,
		widget.NewButton("OK", func() {
			// This would typically close the parent window
		}),
	)

	ew.content = container.NewCenter(content)
	return ew
}

// CreateRenderer creates the renderer for the error widget
func (ew *ErrorWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ew.content)
}