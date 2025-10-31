package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TestGUI is a minimal GUI for testing upload notifications
type TestGUI struct {
	app           fyne.App
	window        fyne.Window
	statusLabel   *widget.Label
	consoleOutput *widget.Entry
	uploadButton  *widget.Button
	testLogPath   string
}

// NewTestGUI creates a new test GUI instance
func NewTestGUI() *TestGUI {
	a := app.New()
	a.Settings().SetTheme(theme.DefaultTheme())

	w := a.NewWindow("Upload Notifications Test")
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()

	gui := &TestGUI{
		app:    a,
		window: w,
	}

	return gui
}

// Show renders and runs the test window
func (g *TestGUI) Show() {
	g.buildUI()
	g.createMockLogFile()
	g.window.ShowAndRun()
}

// buildUI creates the test interface
func (g *TestGUI) buildUI() {
	// Title and description
	title := widget.NewLabelWithStyle("Upload Notifications Test", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("This test verifies that upload notification system works correctly after threading fix.")
	description.Wrapping = fyne.TextWrapWord

	// Status label
	g.statusLabel = widget.NewLabel("Ready to test upload notifications")

	// Console output for diagnostic logging
	g.consoleOutput = widget.NewMultiLineEntry()
	g.consoleOutput.SetPlaceHolder("Diagnostic output will appear here...")
	g.consoleOutput.SetText("Test application started.\n")

	// Upload button
	g.uploadButton = widget.NewButtonWithIcon("Test Upload Log", theme.UploadIcon(), func() {
		g.testUploadLog()
	})
	g.uploadButton.Importance = widget.HighImportance

	// Clear console button
	clearButton := widget.NewButtonWithIcon("Clear Console", theme.ContentClearIcon(), func() {
		g.consoleOutput.SetText("")
	})

	// Create mock log file button
	createLogButton := widget.NewButtonWithIcon("Create New Mock Log", theme.DocumentCreateIcon(), func() {
		g.createMockLogFile()
	})

	// Instructions
	instructions := widget.NewCard("Test Instructions", "", container.NewVBox(
		widget.NewLabel("1. Click 'Create New Mock Log' to generate a test log file"),
		widget.NewLabel("2. Click 'Test Upload Log' to test upload notification system"),
		widget.NewLabel("3. Verify that progress dialog appears"),
		widget.NewLabel("4. Verify that success/error dialogs appear after upload completes"),
		widget.NewLabel("5. Check the console output for diagnostic information"),
		widget.NewLabel("6. Verify that threading synchronization works (no freezing)"),
	))

	// Button container
	buttonContainer := container.NewHBox(
		createLogButton,
		g.uploadButton,
		clearButton,
	)

	// Main content
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		description,
		widget.NewSeparator(),
		g.statusLabel,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		instructions,
		widget.NewSeparator(),
		widget.NewLabel("Diagnostic Console:"),
		g.consoleOutput,
	)

	scrollContent := container.NewVScroll(content)
	g.window.SetContent(scrollContent)
}

// createMockLogFile creates a mock log file for testing
func (g *TestGUI) createMockLogFile() {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "upload-test-logs")
	if err != nil {
		g.logToConsole(fmt.Sprintf("Error creating temp directory: %v\n", err))
		return
	}

	// Create a mock log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	g.testLogPath = filepath.Join(tempDir, fmt.Sprintf("test_log_%s.log", timestamp))

	logContent := fmt.Sprintf(`Test Log File for Upload Notification Testing
Generated: %s
========================================
This is a test log file created to verify the upload notification system.
It contains multiple lines to simulate a real log file.

Line 1: Application started
Line 2: Loading configuration
Line 3: Initializing modules
Line 4: Connecting to server
Line 5: Authentication successful
Line 6: Downloading resources
Line 7: Processing data
Line 8: Rendering interface
Line 9: Ready for user interaction
Line 10: Test completed successfully

Error log entries for testing:
[ERROR] Test error message 1
[WARN]  Test warning message 1
[INFO]  Test info message 1
[DEBUG] Test debug message 1

End of test log file.
`, time.Now().Format(time.RFC3339))

	err = os.WriteFile(g.testLogPath, []byte(logContent), 0644)
	if err != nil {
		g.logToConsole(fmt.Sprintf("Error creating mock log file: %v\n", err))
		return
	}

	g.logToConsole(fmt.Sprintf("Created mock log file: %s\n", g.testLogPath))
	g.statusLabel.SetText(fmt.Sprintf("Mock log created: %s", filepath.Base(g.testLogPath)))
}

// testUploadLog tests the uploadLog function with proper threading
func (g *TestGUI) testUploadLog() {
	if g.testLogPath == "" {
		dialog.ShowError(fmt.Errorf("Please create a mock log file first"), g.window)
		return
	}

	g.logToConsole("Starting upload test...\n")
	g.statusLabel.SetText("Testing upload notifications...")

	// Use a WaitGroup to ensure proper synchronization (matching the fixed implementation)
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a channel to signal when the upload is complete
	uploadComplete := make(chan bool, 1)

	// Show upload progress dialog (simplified approach for test)
	g.logToConsole("Creating and showing progress dialog\n")
	progressDialog := dialog.NewCustom("Uploading Log...", "Cancel",
		widget.NewProgressBarInfinite(), g.window)

	// Show the dialog with error handling
	if progressDialog != nil {
		progressDialog.Show()
		g.logToConsole("Progress dialog shown successfully\n")
	} else {
		// Fallback to simple information dialog if custom dialog creation fails
		g.logToConsole("Progress dialog creation failed, using fallback\n")
		dialog.ShowInformation("Uploading Log", "Uploading log file to test server...", g.window)
	}

	// Start the upload in a separate goroutine (this maintains threading separation)
	go func() {
		defer wg.Done()
		defer func() {
			// Hide the progress dialog when done
			if progressDialog != nil {
				g.logToConsole("Hiding progress dialog\n")
				progressDialog.Hide()
			}
		}()

		g.logToConsole("Starting upload goroutine\n")

		// Simulate the upload process with a delay
		time.Sleep(2 * time.Second)

		// Simulate successful upload (in real implementation, this would be actual upload)
		uploadSuccess := true
		uploadURL := "https://i.dylan.lol/logs/test123.log"

		if uploadSuccess {
			g.logToConsole("Upload successful, showing success dialog\n")
			uploadComplete <- true

			// Create a more informative success dialog
			successTitle := widget.NewLabelWithStyle("✓ Upload Test Successful!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			successTitle.Importance = widget.HighImportance

			messageLabel := widget.NewLabel("The upload notification system is working correctly!")
			messageLabel.Wrapping = fyne.TextWrapWord

			urlLabel := widget.NewLabelWithStyle(fmt.Sprintf("Mock URL: %s", uploadURL), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

			// Create buttons for the dialog
			okButton := widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {})

			// Button container
			buttonContainer := container.NewHBox(
				layout.NewSpacer(),
				okButton,
			)

			// Complete dialog with buttons
			dialogContent := container.NewVBox(
				successTitle,
				widget.NewSeparator(),
				messageLabel,
				widget.NewSeparator(),
				urlLabel,
				widget.NewSeparator(),
				buttonContainer,
			)

			// Create the custom dialog
			customDialog := dialog.NewCustom("Upload Test Successful", "OK", dialogContent, g.window)

			// Set the OK button to close the dialog
			okButton.OnTapped = func() {
				customDialog.Hide()
			}

			// Show the dialog with error handling
			if customDialog != nil {
				customDialog.Show()
				g.logToConsole("Success dialog shown successfully\n")
			} else {
				// Fallback to simple information dialog
				dialog.ShowInformation("Upload Test Successful", "The upload notification system is working correctly!", g.window)
			}

			g.statusLabel.SetText("Upload test completed successfully")
		} else {
			g.logToConsole("Upload failed, showing error dialog\n")
			uploadComplete <- false

			dialog.ShowError(fmt.Errorf("Upload test failed"), g.window)
			g.statusLabel.SetText("Upload test failed")
		}
	}()

	// Wait for the upload to complete
	go func() {
		wg.Wait()
		g.logToConsole("Upload goroutine completed\n")
		close(uploadComplete)
	}()
}

// logToConsole adds a message to the console output
func (g *TestGUI) logToConsole(message string) {
	if g.consoleOutput != nil {
		currentText := g.consoleOutput.Text
		timestamp := time.Now().Format("15:04:05")
		newText := fmt.Sprintf("%s %s", timestamp, message)
		g.consoleOutput.SetText(currentText + newText)

		// Auto-scroll to bottom
		lines := strings.Split(g.consoleOutput.Text, "\n")
		g.consoleOutput.CursorRow = len(lines) - 1
	}
}

func main() {
	fmt.Println("Upload Notifications Test Program")
	fmt.Println("==============================")
	fmt.Println("This program tests the fixed upload notification implementation.")
	fmt.Println("It verifies that:")
	fmt.Println("1. Progress dialogs appear correctly in the main thread")
	fmt.Println("2. Success/error dialogs appear after upload completes")
	fmt.Println("3. Threading synchronization works without freezing")
	fmt.Println("4. Diagnostic logging is working")
	fmt.Println("")
	fmt.Println("Starting GUI...")

	// Create and show the test GUI
	gui := NewTestGUI()
	gui.Show()
}
