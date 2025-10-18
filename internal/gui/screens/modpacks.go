// Package screens provides screen implementations for the TheBoys Launcher
package screens

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/app"
	"theboys-launcher/internal/gui/widgets"
	"theboys-launcher/internal/modpack"
)

// ModpacksScreen represents the modpacks screen
type ModpacksScreen struct {
	app     *app.Application
	window  fyne.Window
	modpackMgr *modpack.Manager
	modpacks   []*modpack.Modpack

	// GUI components
	searchBar     *widgets.SearchBar
	filterButtons *widgets.FilterButtons
	modpackList   *widgets.ModpackList
	statusLabel   *widget.Label
	refreshBtn    *widget.Button

	// State
	currentFilter string
	searchQuery   string
	loading       bool
}

// NewModpacksScreen creates a new modpacks screen
func NewModpacksScreen(application *app.Application) *ModpacksScreen {
	screen := &ModpacksScreen{
		app:          application,
		window:       application.GetMainWindow(),
		modpackMgr:   modpack.NewManager(application.GetConfig(), application.GetLogger()),
		currentFilter: "All",
	}

	// Create GUI components
	screen.createComponents()

	// Load initial data
	screen.loadModpacks()

	return screen
}

// GetID returns the screen ID
func (ms *ModpacksScreen) GetID() string {
	return "modpacks"
}

// GetTitle returns the screen title
func (ms *ModpacksScreen) GetTitle() string {
	return "Modpacks"
}

// GetContent returns the screen content
func (ms *ModpacksScreen) GetContent() fyne.CanvasObject {
	// Create header with search and filters
	header := container.NewVBox(
		container.NewHBox(
			ms.refreshBtn,
			widget.NewLabel("Modpack Management"),
		),
		ms.searchBar,
		widget.NewSeparator(),
		ms.filterButtons,
		widget.NewSeparator(),
	)

	// Create main content with list and status
	mainContent := container.NewBorder(
		header, // Top
		ms.statusLabel, // Bottom
		nil, // Left
		nil, // Right
		ms.modpackList, // Center
	)

	return mainContent
}

// OnShow is called when the screen is shown
func (ms *ModpacksScreen) OnShow() error {
	ms.app.GetLogger().Info("Showing modpacks screen")

	// Refresh modpacks when screen is shown
	ms.loadModpacks()

	return nil
}

// OnHide is called when the screen is hidden
func (ms *ModpacksScreen) OnHide() error {
	ms.app.GetLogger().Info("Hiding modpacks screen")
	return nil
}

// createComponents creates the GUI components
func (ms *ModpacksScreen) createComponents() {
	// Create search bar
	ms.searchBar = widgets.NewSearchBar()
	ms.searchBar.SetOnSearch(func(query string) {
		ms.searchQuery = strings.ToLower(query)
		ms.filterModpacks()
	})

	// Create filter buttons
	ms.filterButtons = widgets.NewFilterButtons()
	ms.filterButtons.SetOnSearch(func(filter string) {
		ms.currentFilter = filter
		ms.filterModpacks()
	})

	// Create modpack list
	ms.modpackList = widgets.NewModpackList(ms.app.GetState())
	ms.modpackList.SetOnSelect(func(modpack *modpack.Modpack) {
		ms.handleModpackAction(modpack)
	})

	// Create status label
	ms.statusLabel = widget.NewLabel("Loading modpacks...")

	// Create refresh button
	ms.refreshBtn = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		ms.loadModpacks()
	})
}

// loadModpacks loads all available modpacks
func (ms *ModpacksScreen) loadModpacks() {
	ms.loading = true
	ms.statusLabel.SetText("Loading modpacks...")
	ms.refreshBtn.Disable()

	// Load modpacks in a goroutine to avoid blocking UI
	go func() {
		modpacks, err := ms.modpackMgr.GetAvailableModpacks()
		if err != nil {
			ms.app.GetLogger().Error("Failed to load modpacks: %v", err)
			ms.statusLabel.SetText(fmt.Sprintf("Error loading modpacks: %v", err))
		} else {
			ms.modpacks = modpacks
			ms.filterModpacks()
			ms.statusLabel.SetText(fmt.Sprintf("Found %d modpack(s)", len(modpacks)))
		}

		ms.loading = false
		ms.refreshBtn.Enable()
	}()
}

// filterModpacks filters modpacks based on current search and filter
func (ms *ModpacksScreen) filterModpacks() {
	if len(ms.modpacks) == 0 {
		ms.modpackList.SetModpacks([]*modpack.Modpack{})
		return
	}

	var filtered []*modpack.Modpack

	for _, modpack := range ms.modpacks {
		// Apply search filter
		if ms.searchQuery != "" {
			if !strings.Contains(strings.ToLower(modpack.Name), ms.searchQuery) &&
			   !strings.Contains(strings.ToLower(modpack.Description), ms.searchQuery) &&
			   !strings.Contains(strings.ToLower(modpack.Author), ms.searchQuery) {
				continue
			}
		}

		// Apply status filter
		switch ms.currentFilter {
		case "Installed":
			if !modpack.IsInstalled() {
				continue
			}
		case "Not Installed":
			if modpack.IsInstalled() {
				continue
			}
		case "Updates Available":
			if !modpack.NeedsUpdate() {
				continue
			}
		}

		filtered = append(filtered, modpack)
	}

	ms.modpackList.SetModpacks(filtered)
	ms.statusLabel.SetText(fmt.Sprintf("Showing %d modpack(s)", len(filtered)))
}

// handleModpackAction handles actions on modpacks
func (ms *ModpacksScreen) handleModpackAction(modpack *modpack.Modpack) {
	switch modpack.Status {
	case modpack.StatusNotInstalled:
		ms.installModpack(modpack)
	case modpack.StatusInstalled:
		ms.showModpackOptions(modpack)
	case modpack.StatusUpdateAvailable:
		ms.updateModpack(modpack)
	case modpack.StatusDownloading, modpack.StatusInstalling:
		ms.showInstallationProgress(modpack)
	case modpack.StatusError:
		ms.showErrorDialog(modpack)
	}
}

// installModpack installs a modpack
func (ms *ModpacksScreen) installModpack(modpack *modpack.Modpack) {
	ms.app.GetLogger().Info("Installing modpack: %s", modpack.Name)

	// Show confirmation dialog
	dialog.ShowConfirm(
		"Install Modpack",
		fmt.Sprintf("Are you sure you want to install %s?\n\nThis will download approximately %s.",
			modpack.Name, modpack.GetFormattedSize()),
		func(confirmed bool) {
			if confirmed {
				ms.performInstall(modpack)
			}
		},
		ms.window,
	)
}

// performInstall performs the actual installation
func (ms *ModpacksScreen) performInstall(modpack *modpack.Modpack) {
	// Update modpack status
	modpack.Status = modpack.StatusDownloading
	ms.modpackList.Refresh()

	// Create progress dialog
	progressDialog := dialog.NewCustom("Installing Modpack", "Cancel", ms.createProgressDialog(modpack), ms.app.GetMainWindow())
	progressDialog.Show()

	// Start installation in background
	go func() {
		err := ms.modpackMgr.InstallModpack(modpack, func(progress *modpack.InstallationProgress) {
			// Update progress dialog (this would need more complex implementation)
			ms.app.GetLogger().Debug("Installation progress: %.1f%%", progress.Progress*100)
		})

		// Close progress dialog
		progressDialog.Hide()

		if err != nil {
			ms.app.GetLogger().Error("Failed to install modpack %s: %v", modpack.Name, err)
			dialog.ShowError(fmt.Errorf("Failed to install modpack: %w", err), ms.app.GetMainWindow())
		} else {
			ms.app.GetLogger().Info("Successfully installed modpack: %s", modpack.Name)
			dialog.ShowInformation("Installation Complete",
				fmt.Sprintf("%s has been successfully installed!", modpack.Name),
				ms.app.GetMainWindow())
		}

		// Refresh modpack list
		ms.loadModpacks()
	}()
}

// showModpackOptions shows options for an installed modpack
func (ms *ModpacksScreen) showModpackOptions(modpack *modpack.Modpack) {
	// Create options dialog
	options := []string{
		"Launch",
		"Uninstall",
		"View Details",
		"Configure",
	}

	dialog.ShowCustom("Modpack Options", "Close",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Select action for %s:", modpack.Name)),
			widget.NewSeparator(),
			widget.NewButton("Launch", func() {
				ms.launchModpack(modpack)
			}),
			widget.NewButton("Uninstall", func() {
				ms.uninstallModpack(modpack)
			}),
			widget.NewButton("View Details", func() {
				ms.showModpackDetails(modpack)
			}),
			widget.NewButton("Configure", func() {
				ms.showModpackConfiguration(modpack)
			}),
		),
		ms.window,
	)
}

// launchModpack launches a modpack
func (ms *ModpacksScreen) launchModpack(modpack *modpack.Modpack) {
	ms.app.GetLogger().Info("Launching modpack: %s", modpack.Name)

	err := ms.modpackMgr.LaunchModpack(modpack)
	if err != nil {
		ms.app.GetLogger().Error("Failed to launch modpack %s: %v", modpack.Name, err)
		dialog.ShowError(fmt.Errorf("Failed to launch modpack: %w", err), ms.app.GetMainWindow())
	} else {
		ms.app.GetLogger().Info("Successfully launched modpack: %s", modpack.Name)
		dialog.ShowInformation("Launch Complete",
			fmt.Sprintf("%s has been launched!", modpack.Name),
			ms.app.GetMainWindow())
	}
}

// uninstallModpack uninstalls a modpack
func (ms *ModpacksScreen) uninstallModpack(modpack *modpack.Modpack) {
	ms.app.GetLogger().Info("Uninstalling modpack: %s", modpack.Name)

	dialog.ShowConfirm(
		"Uninstall Modpack",
		fmt.Sprintf("Are you sure you want to uninstall %s?\n\nThis will remove all files and cannot be undone.", modpack.Name),
		func(confirmed bool) {
			if confirmed {
				ms.performUninstall(modpack)
			}
		},
		ms.window,
	)
}

// performUninstall performs the actual uninstallation
func (ms *ModpacksScreen) performUninstall(modpack *modpack.Modpack) {
	err := ms.modpackMgr.UninstallModpack(modpack)
	if err != nil {
		ms.app.GetLogger().Error("Failed to uninstall modpack %s: %v", modpack.Name, err)
		dialog.ShowError(fmt.Errorf("Failed to uninstall modpack: %w", err), ms.app.GetMainWindow())
	} else {
		ms.app.GetLogger().Info("Successfully uninstalled modpack: %s", modpack.Name)
		dialog.ShowInformation("Uninstall Complete",
			fmt.Sprintf("%s has been successfully uninstalled!", modpack.Name),
			ms.app.GetMainWindow())

		// Refresh modpack list
		ms.loadModpacks()
	}
}

// updateModpack updates a modpack
func (ms *ModpacksScreen) updateModpack(modpack *modpack.Modpack) {
	ms.app.GetLogger().Info("Updating modpack: %s", modpack.Name)

	dialog.ShowConfirm(
		"Update Modpack",
		fmt.Sprintf("Are you sure you want to update %s to the latest version?", modpack.Name),
		func(confirmed bool) {
			if confirmed {
				ms.performUpdate(modpack)
			}
		},
		ms.window,
	)
}

// performUpdate performs the actual update
func (ms *ModpacksScreen) performUpdate(modpack *modpack.Modpack) {
	// Update modpack status
	modpack.Status = modpack.StatusDownloading
	ms.modpackList.Refresh()

	// Create progress dialog
	progressDialog := dialog.NewCustom("Updating Modpack", "Cancel", ms.createProgressDialog(modpack), ms.app.GetMainWindow())
	progressDialog.Show()

	// Start update in background
	go func() {
		err := ms.modpackMgr.UpdateModpack(modpack, func(progress *modpack.InstallationProgress) {
			ms.app.GetLogger().Debug("Update progress: %.1f%%", progress.Progress*100)
		})

		// Close progress dialog
		progressDialog.Hide()

		if err != nil {
			ms.app.GetLogger().Error("Failed to update modpack %s: %v", modpack.Name, err)
			dialog.ShowError(fmt.Errorf("Failed to update modpack: %w", err), ms.app.GetMainWindow())
		} else {
			ms.app.GetLogger().Info("Successfully updated modpack: %s", modpack.Name)
			dialog.ShowInformation("Update Complete",
				fmt.Sprintf("%s has been successfully updated!", modpack.Name),
				ms.app.GetMainWindow())
		}

		// Refresh modpack list
		ms.loadModpacks()
	}()
}

// showModpackDetails shows detailed information about a modpack
func (ms *ModpacksScreen) showModpackDetails(modpack *modpack.Modpack) {
	// Create details content
	details := container.NewVBox(
		widget.NewLabelWithStyle(modpack.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Author: %s", modpack.Author)),
		widget.NewLabel(fmt.Sprintf("Version: %s", modpack.Version)),
		widget.NewLabel(fmt.Sprintf("Minecraft: %s", modpack.MinecraftVersion.ID)),
		widget.NewLabel(fmt.Sprintf("Mod Loader: %s %s", modpack.ModLoader.Type, modpack.ModLoader.Version)),
		widget.NewLabel(fmt.Sprintf("Type: %s", string(modpack.Type))),
		widget.NewLabel(fmt.Sprintf("Size: %s", modpack.GetFormattedSize())),
		widget.NewLabel(fmt.Sprintf("Memory: %d MB", modpack.RequiredMemory)),
		widget.NewSeparator(),
		widget.NewLabel("Description:"),
		widget.NewLabel(modpack.Description),
	)

	if len(modpack.Tags) > 0 {
		details.Add(widget.NewSeparator())
		details.Add(widget.NewLabel("Tags:"))
		details.Add(widget.NewLabel(strings.Join(modpack.Tags, ", ")))
	}

	if len(modpack.Features) > 0 {
		details.Add(widget.NewSeparator())
		details.Add(widget.NewLabel("Features:"))
		for _, feature := range modpack.Features {
			details.Add(widget.NewLabel("• " + feature))
		}
	}

	dialog.ShowCustom("Modpack Details", "Close", details, ms.app.GetMainWindow())
}

// showModpackConfiguration shows configuration options for a modpack
func (ms *ModpacksScreen) showModpackConfiguration(modpack *modpack.Modpack) {
	// This is a placeholder for modpack configuration
	// In a real implementation, this would show modpack-specific settings
	configContent := container.NewVBox(
		widget.NewLabelWithStyle("Modpack Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Configuration for %s", modpack.Name)),
		widget.NewLabel("This feature will be implemented in a future phase."),
	)

	dialog.ShowCustom("Configuration", "Close", configContent, ms.app.GetMainWindow())
}

// showInstallationProgress shows installation progress
func (ms *ModpacksScreen) showInstallationProgress(modpack *modpack.Modpack) {
	progressContent := container.NewVBox(
		widget.NewLabelWithStyle("Installation in Progress", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Installing %s...", modpack.Name)),
		widget.NewProgressBar(),
		widget.NewLabel("Please wait..."),
	)

	dialog.ShowCustom("Installing", "", progressContent, ms.app.GetMainWindow())
}

// showErrorDialog shows error details for a modpack
func (ms *ModpacksScreen) showErrorDialog(modpack *modpack.Modpack) {
	errorContent := container.NewVBox(
		widget.NewLabelWithStyle("Installation Error", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("An error occurred while installing %s", modpack.Name)),
		widget.NewLabel("Please check the logs for more details."),
		widget.NewButton("Retry Installation", func() {
			ms.installModpack(modpack)
		}),
	)

	dialog.ShowCustom("Error", "Close", errorContent, ms.app.GetMainWindow())
}

// createProgressDialog creates a progress dialog content
func (ms *ModpacksScreen) createProgressDialog(modpack *modpack.Modpack) fyne.CanvasObject {
	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("Starting installation...")

	return container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Installing %s", modpack.Name)),
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		widget.NewLabel("This may take several minutes..."),
	)
}