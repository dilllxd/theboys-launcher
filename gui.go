package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"image/color"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GUI drives the launcher UI experience.
type GUI struct {
	app            fyne.App
	window         fyne.Window
	modpacks       []Modpack
	filtered       []Modpack
	searchQuery    string
	activeCategory string
	root           string
	exePath        string
	prismProcess   **os.Process

	// UI elements we mutate
	searchEntry   *widget.Entry
	statusLabel   *widget.Label
	progressBar   *widget.ProgressBar
	consoleOutput *widget.Entry
	tabs          *container.AppTabs
	browseGrid    *fyne.Container
	featuredGrid  *fyne.Container

	// Log file monitoring
	logWatcherActive   bool
	logStopChan        chan struct{}
	logMutex           sync.RWMutex
	loadingOverlay     fyne.CanvasObject
	loadingLabel       *widget.Label
	memorySummaryLabel *widget.Label

	// Modpack status tracking
	modpackStates    map[string]*ModpackState
	cardBindings     map[string][]*modpackCardBinding
	stateMu          sync.RWMutex
	bindingsMu       sync.RWMutex
	runningModpackID string
	runningMu        sync.RWMutex
	processMu        sync.Mutex
}

// modernTheme tweaks the default Fyne look.
type modernTheme struct {
	fyne.Theme
}

func (m *modernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.RGBA{R: 99, G: 102, B: 241, A: 255} // indigo
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.RGBA{R: 19, G: 24, B: 38, A: 255}
		}
		return color.RGBA{R: 245, G: 246, B: 250, A: 255}
	case theme.ColorNameHover:
		return color.RGBA{R: 67, G: 56, B: 202, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 99, G: 102, B: 241, A: 255}
	}
	return m.Theme.Color(name, variant)
}

type PrimaryAction int

const (
	ActionNone PrimaryAction = iota
	ActionInstall
	ActionLaunch
	ActionUpdate
	ActionKill
)

type ModpackState struct {
	ID              string
	Installed       bool
	UpdateAvailable bool
	Running         bool
	Busy            bool
	CurrentAction   PrimaryAction
	LocalVersion    string
	RemoteVersion   string
	RunningPID      int
	LastChecked     time.Time
	Error           error
}

func (s *ModpackState) PrimaryAction() PrimaryAction {
	if s == nil {
		return ActionNone
	}
	if s.Running {
		return ActionKill
	}
	if s.Busy {
		switch s.CurrentAction {
		case ActionInstall, ActionUpdate, ActionLaunch:
			return s.CurrentAction
		default:
			return ActionNone
		}
	}
	if !s.Installed {
		return ActionInstall
	}
	if s.UpdateAvailable {
		return ActionUpdate
	}
	return ActionLaunch
}

func (s *ModpackState) PrimaryLabel() string {
	if s == nil {
		return "Checking..."
	}
	if s.Running {
		return "Kill"
	}
	if s.Busy {
		switch s.CurrentAction {
		case ActionInstall:
			return "Installing..."
		case ActionUpdate:
			return "Updating..."
		case ActionLaunch:
			return "Launching..."
		default:
			return "Working..."
		}
	}
	if !s.Installed {
		return "Install"
	}
	if s.UpdateAvailable {
		return "Update"
	}
	return "Launch"
}

func (s *ModpackState) PrimaryIcon() fyne.Resource {
	if s == nil {
		return theme.ViewRefreshIcon()
	}
	if s.Running {
		return theme.CancelIcon()
	}
	if s.Busy {
		return theme.ViewRefreshIcon()
	}
	if !s.Installed {
		return theme.DownloadIcon()
	}
	if s.UpdateAvailable {
		return theme.ViewRefreshIcon()
	}
	return theme.MediaPlayIcon()
}

func (s *ModpackState) StatusSummary() string {
	if s == nil {
		return "Determining status..."
	}
	if s.Error != nil {
		return fmt.Sprintf("Status error: %v", s.Error)
	}
	if s.Running {
		if s.RunningPID > 0 {
			return fmt.Sprintf("Running (PID %d)", s.RunningPID)
		}
		return "Running"
	}
	if s.Busy {
		switch s.CurrentAction {
		case ActionInstall:
			return "Installing..."
		case ActionUpdate:
			return "Updating..."
		case ActionLaunch:
			return "Launching..."
		default:
			return "Working..."
		}
	}
	if !s.Installed {
		if s.RemoteVersion != "" {
			return fmt.Sprintf("Not installed (latest %s)", s.RemoteVersion)
		}
		return "Not installed"
	}
	if s.UpdateAvailable && s.LocalVersion != "" && s.RemoteVersion != "" {
		return fmt.Sprintf("Update available: %s -> %s", s.LocalVersion, s.RemoteVersion)
	}
	if s.LocalVersion != "" {
		return fmt.Sprintf("Up to date (%s)", s.LocalVersion)
	}
	return "Up to date"
}

type modpackCardBinding struct {
	modpack      Modpack
	view         string
	card         *widget.Card
	statusLabel  *widget.Label
	primaryBtn   *widget.Button
	deleteBtn    *widget.Button
	reinstallBtn *widget.Button
}

const (
	viewBrowse   = "browse"
	viewFeatured = "featured"
)

// NewGUI spins up the modern application shell.
func NewGUI(modpacks []Modpack, root string) *GUI {
	a := app.New()
	a.Settings().SetTheme(&modernTheme{Theme: theme.DefaultTheme()})

	w := a.NewWindow(fmt.Sprintf("%s %s", launcherName, version))
	w.Resize(fyne.NewSize(1280, 820))
	w.CenterOnScreen()
	w.SetFixedSize(false)

	gui := &GUI{
		app:           a,
		window:        w,
		modpacks:      modpacks,
		filtered:      append([]Modpack(nil), modpacks...),
		root:          root,
		modpackStates: make(map[string]*ModpackState),
		cardBindings:  make(map[string][]*modpackCardBinding),
	}

	return gui
}

// Show renders and runs the window loop.
func (g *GUI) Show() {
	g.buildUI()
	g.startUpdateCheck()

	// Set up window close callback to clean up resources
	g.window.SetCloseIntercept(func() {
		g.cleanup()
		g.window.Close()
	})

	g.window.ShowAndRun()
}

// cleanup stops background tasks and releases resources
func (g *GUI) cleanup() {
	g.stopLogFileWatcher()
}

func (g *GUI) launchWithCallback(prismProcess **os.Process, root, exePath string) {
	g.prismProcess = prismProcess
	g.root = root
	g.exePath = exePath
	g.Show()
}

func (g *GUI) buildUI() {
	header := g.buildHeader()
	sidebar := g.buildSidebar()
	content := g.buildContent()
	status := g.buildStatusBar()
	overlay := g.buildLoadingOverlay()

	body := container.NewBorder(
		header,
		status,
		sidebar,
		nil,
		content,
	)

	g.loadingOverlay = overlay
	root := container.NewStack(body, overlay)
	g.window.SetContent(root)
	g.refreshAllModpackStates()
}

func (g *GUI) buildHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(launcherName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel(fmt.Sprintf("Version %s - %d modpacks", version, len(g.modpacks)))
	titleBox := container.NewVBox(title, subtitle)

	g.searchEntry = widget.NewEntry()
	g.searchEntry.SetPlaceHolder("Search modpacks...")
	g.searchEntry.OnChanged = func(q string) {
		g.searchQuery = strings.TrimSpace(q)
		g.applyFilters()
	}

	searchWrap := container.New(layout.NewGridWrapLayout(fyne.NewSize(360, 40)), g.searchEntry)
	headerRow := container.NewHBox(
		titleBox,
		layout.NewSpacer(),
		searchWrap,
	)

	return container.NewVBox(headerRow, widget.NewSeparator())
}

func (g *GUI) buildSidebar() fyne.CanvasObject {
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		g.refreshModpacks()
	})
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		g.showSettings()
	})
	consoleBtn := widget.NewButtonWithIcon("Console", theme.ComputerIcon(), func() {
		g.showConsole()
	})

	quickActions := widget.NewCard("Actions", "", container.NewVBox(
		refreshBtn,
		settingsBtn,
		consoleBtn,
	))

	categoryButtons := []fyne.CanvasObject{}
	for _, cat := range []struct {
		label string
		value string
	}{
		{"All", ""},
		{"Featured", "featured"},
		{"Performance", "performance"},
		{"Visuals", "visuals"},
		{"Adventure", "adventure"},
	} {
		value := cat.value
		btn := widget.NewButton(cat.label, func() {
			g.filterByCategory(value)
		})
		categoryButtons = append(categoryButtons, btn)
	}

	categories := widget.NewCard("Categories", "", container.NewVBox(categoryButtons...))

	g.memorySummaryLabel = widget.NewLabel("")
	g.updateMemorySummaryLabel()
	info := widget.NewCard("Status", "", container.NewVBox(
		g.memorySummaryLabel,
		widget.NewLabel(fmt.Sprintf("Signed in as: %s", getCurrentUser())),
	))

	content := container.NewVBox(
		quickActions,
		categories,
		info,
		layout.NewSpacer(),
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(260, 0))
	return scroll
}

func (g *GUI) buildContent() fyne.CanvasObject {
	g.browseGrid = container.New(layout.NewGridWrapLayout(fyne.NewSize(320, 220)))
	g.featuredGrid = container.New(layout.NewGridWrapLayout(fyne.NewSize(320, 220)))
	g.populateBrowseGrid()
	g.populateFeaturedGrid()

	browse := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		container.NewVBox(g.browseGrid),
	)

	featured := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		container.NewVBox(g.featuredGrid),
	)

	console := g.buildConsoleView()

	g.tabs = container.NewAppTabs(
		container.NewTabItem("Browse", container.NewVScroll(browse)),
		container.NewTabItem("Featured", container.NewVScroll(featured)),
		container.NewTabItem("Console", console),
	)
	g.tabs.SetTabLocation(container.TabLocationTop)
	return g.tabs
}

func (g *GUI) buildConsoleView() fyne.CanvasObject {
	g.consoleOutput = widget.NewMultiLineEntry()
	g.consoleOutput.SetPlaceHolder("Launcher output appears here...")
	g.consoleOutput.SetText("Waiting for log file content...")

	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		g.consoleOutput.SetText("")
	})
	copyBtn := widget.NewButtonWithIcon("Copy All", theme.ContentCopyIcon(), func() {
		g.window.Clipboard().SetContent(g.consoleOutput.Text)
	})
	uploadBtn := widget.NewButtonWithIcon("Upload to i.dylan.lol", theme.UploadIcon(), func() {
		g.uploadLog()
	})

	toolbar := container.NewHBox(clearBtn, copyBtn, uploadBtn, layout.NewSpacer())

	// Start log file monitoring when console view is created
	g.startLogFileWatcher()

	return container.NewBorder(toolbar, nil, nil, nil, g.consoleOutput)
}

func (g *GUI) buildStatusBar() fyne.CanvasObject {
	g.statusLabel = widget.NewLabel("Ready")
	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	bar := container.NewBorder(
		nil,
		nil,
		g.statusLabel,
		container.NewHBox(layout.NewSpacer(), g.progressBar),
	)

	return container.NewVBox(widget.NewSeparator(), container.NewPadded(bar))
}

func (g *GUI) buildLoadingOverlay() fyne.CanvasObject {
	background := canvas.NewRectangle(color.NRGBA{R: 15, G: 23, B: 42, A: 160})
	background.Show()

	spinner := widget.NewProgressBarInfinite()
	g.loadingLabel = widget.NewLabel("Working...")

	card := widget.NewCard("", "", container.NewVBox(
		spinner,
		g.loadingLabel,
	))

	overlay := container.NewMax(background, container.NewCenter(card))
	overlay.Hide()
	return overlay
}

func (g *GUI) populateBrowseGrid() {
	g.clearBindings(viewBrowse)
	g.browseGrid.Objects = g.browseGrid.Objects[:0]
	if len(g.filtered) == 0 {
		g.browseGrid.Add(widget.NewCard("", "", widget.NewLabel("No modpacks match your filters yet.")))
	} else {
		for _, mod := range g.filtered {
			g.browseGrid.Add(g.modpackCard(mod, viewBrowse))
		}
	}
	g.browseGrid.Refresh()
}

func (g *GUI) populateFeaturedGrid() {
	g.clearBindings(viewFeatured)
	g.featuredGrid.Objects = g.featuredGrid.Objects[:0]

	for _, mod := range g.modpacks {
		if mod.Default || strings.EqualFold(mod.Category, "featured") {
			g.featuredGrid.Add(g.modpackCard(mod, viewFeatured))
		}
	}

	if len(g.featuredGrid.Objects) == 0 {
		g.featuredGrid.Add(widget.NewCard("", "", widget.NewLabel("No featured modpacks yet.")))
	}

	g.featuredGrid.Refresh()
}

func (g *GUI) modpackCard(mod Modpack, view string) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(mod.DisplayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	meta := widget.NewLabel(fmt.Sprintf("by %s - %s", mod.Author, mod.LastUpdated))
	meta.Wrapping = fyne.TextWrapWord

	description := widget.NewLabel(mod.Description)
	description.Wrapping = fyne.TextWrapWord

	ram := widget.NewLabel(fmt.Sprintf("Minimum RAM: %d GB - Recommended: %d GB", mod.MinRam/1024, mod.RecommendedRam/1024))

	tagObjects := make([]fyne.CanvasObject, 0, len(mod.Tags))
	for _, tag := range mod.Tags {
		if tag == "" {
			continue
		}
		tagLabel := widget.NewLabel(fmt.Sprintf("#%s", strings.ToLower(tag)))
		tagLabel.Alignment = fyne.TextAlignCenter
		tagObjects = append(tagObjects, tagLabel)
	}
	tagLayout := container.New(layout.NewGridWrapLayout(fyne.NewSize(90, 24)), tagObjects...)
	if len(tagObjects) == 0 {
		tagLayout = container.NewHBox(widget.NewLabel("No tags yet"))
	}

	primaryBtn := widget.NewButtonWithIcon("Launch", theme.MediaPlayIcon(), func() {
		g.handlePrimaryAction(mod)
	})
	primaryBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		g.deleteModpack(mod)
	})
	reinstallBtn := widget.NewButtonWithIcon("Reinstall", theme.ViewRefreshIcon(), func() {
		g.reinstallModpack(mod)
	})

	statusLabel := widget.NewLabel("Checking status...")
	statusLabel.Wrapping = fyne.TextWrapWord

	buttonRow := container.NewHBox(primaryBtn, layout.NewSpacer())
	secondaryRow := container.NewHBox(deleteBtn, reinstallBtn)

	card := widget.NewCard("", "", container.NewVBox(
		title,
		meta,
		description,
		tagLayout,
		ram,
		statusLabel,
		buttonRow,
		secondaryRow,
	))

	card.SetSubTitle(" ")

	binding := &modpackCardBinding{
		modpack:      mod,
		view:         view,
		card:         card,
		statusLabel:  statusLabel,
		primaryBtn:   primaryBtn,
		deleteBtn:    deleteBtn,
		reinstallBtn: reinstallBtn,
	}
	g.registerCardBinding(binding)

	return card
}

func (g *GUI) applyFilters() {
	query := strings.ToLower(g.searchQuery)

	g.filtered = g.filtered[:0]
	for _, mod := range g.modpacks {
		if g.activeCategory != "" && !modMatchesCategory(mod, g.activeCategory) {
			continue
		}
		if query != "" && !modMatchesQuery(mod, query) {
			continue
		}
		g.filtered = append(g.filtered, mod)
	}

	g.populateBrowseGrid()
}

func (g *GUI) clearBindings(view string) {
	g.bindingsMu.Lock()
	defer g.bindingsMu.Unlock()

	for id, list := range g.cardBindings {
		filtered := list[:0]
		for _, binding := range list {
			if binding.view != view {
				filtered = append(filtered, binding)
			}
		}
		if len(filtered) == 0 {
			delete(g.cardBindings, id)
		} else {
			g.cardBindings[id] = filtered
		}
	}
}

func (g *GUI) registerCardBinding(binding *modpackCardBinding) {
	g.bindingsMu.Lock()
	g.cardBindings[binding.modpack.ID] = append(g.cardBindings[binding.modpack.ID], binding)
	g.bindingsMu.Unlock()
	g.applyStateToBinding(binding)
}

func (g *GUI) applyStateToBinding(binding *modpackCardBinding) {
	state := g.getModpackState(binding.modpack.ID)
	fyne.Do(func() {
		g.updateBindingUI(binding, state)
	})
}

func (g *GUI) updateBindingUI(binding *modpackCardBinding, state *ModpackState) {
	if binding == nil {
		return
	}

	if binding.card != nil {
		binding.card.SetSubTitle("")
	}

	summary := "Checking status..."
	if state != nil {
		summary = state.StatusSummary()
	}
	if binding.statusLabel != nil {
		binding.statusLabel.SetText(summary)
	}

	if binding.primaryBtn != nil {
		if state != nil {
			binding.primaryBtn.SetText(state.PrimaryLabel())
			binding.primaryBtn.SetIcon(state.PrimaryIcon())
		} else {
			binding.primaryBtn.SetText("Checking...")
			binding.primaryBtn.SetIcon(theme.ViewRefreshIcon())
		}

		enabled := true
		if state == nil {
			enabled = false
		} else if state.Busy && !state.Running {
			enabled = false
		} else if state.PrimaryAction() == ActionNone && !state.Running {
			enabled = false
		}

		if enabled {
			binding.primaryBtn.Enable()
		} else {
			binding.primaryBtn.Disable()
		}
	}

	canModify := state != nil && state.Installed && !state.Busy && !state.Running
	if binding.deleteBtn != nil {
		if canModify {
			binding.deleteBtn.Enable()
		} else {
			binding.deleteBtn.Disable()
		}
	}
	if binding.reinstallBtn != nil {
		if canModify {
			binding.reinstallBtn.Enable()
		} else {
			binding.reinstallBtn.Disable()
		}
	}
}

func modMatchesCategory(mod Modpack, category string) bool {
	if strings.EqualFold(mod.Category, category) {
		return true
	}
	for _, tag := range mod.Tags {
		if strings.EqualFold(tag, category) {
			return true
		}
	}
	return false
}

func modMatchesQuery(mod Modpack, query string) bool {
	if strings.Contains(strings.ToLower(mod.DisplayName), query) {
		return true
	}
	if strings.Contains(strings.ToLower(mod.Description), query) {
		return true
	}
	if strings.Contains(strings.ToLower(mod.Author), query) {
		return true
	}
	return false
}

func (g *GUI) getModpackState(id string) *ModpackState {
	g.stateMu.RLock()
	defer g.stateMu.RUnlock()
	state, ok := g.modpackStates[id]
	if !ok {
		return nil
	}
	copy := *state
	return &copy
}

func (g *GUI) setModpackState(id string, update func(*ModpackState)) {
	g.stateMu.Lock()
	state, ok := g.modpackStates[id]
	if !ok {
		state = &ModpackState{ID: id}
		g.modpackStates[id] = state
	}
	update(state)
	stateCopy := *state
	g.stateMu.Unlock()

	g.updateUIForState(id, &stateCopy)
}

func (g *GUI) updateUIForState(id string, state *ModpackState) {
	g.bindingsMu.RLock()
	bindings := append([]*modpackCardBinding(nil), g.cardBindings[id]...)
	g.bindingsMu.RUnlock()

	if len(bindings) > 0 {
		fyne.Do(func() {
			for _, binding := range bindings {
				g.updateBindingUI(binding, state)
			}
		})
	}
}

func (g *GUI) refreshAllModpackStates() {
	for _, mod := range g.modpacks {
		modCopy := mod
		go g.refreshModpackState(modCopy)
	}
}

func (g *GUI) refreshModpackState(mod Modpack) {
	instDir := g.modpackInstanceDir(mod)
	installed := g.isModpackInstalled(mod)

	var (
		updateAvailable bool
		localVersion    string
		remoteVersion   string
		err             error
	)

	if installed {
		updateAvailable, localVersion, remoteVersion, err = checkModpackUpdate(mod, instDir)
		if err == nil && localVersion == "" {
			installed = false
			updateAvailable = false
		}
	} else {
		remoteVersion, err = fetchRemotePackVersion(mod.PackURL)
	}

	errCopy := err
	g.setModpackState(mod.ID, func(state *ModpackState) {
		state.Installed = installed
		state.UpdateAvailable = updateAvailable && localVersion != ""
		if installed {
			state.LocalVersion = localVersion
		} else {
			state.LocalVersion = ""
		}
		if remoteVersion != "" {
			state.RemoteVersion = remoteVersion
		}
		state.LastChecked = time.Now()
		if errCopy != nil {
			state.Error = errCopy
		} else {
			state.Error = nil
		}
		if !installed {
			state.Running = false
			state.RunningPID = 0
		}
	})
}

func (g *GUI) modpackInstanceDir(mod Modpack) string {
	return filepath.Join(g.root, "prism", "instances", mod.InstanceName)
}

func (g *GUI) isModpackInstalled(mod Modpack) bool {
	instDir := g.modpackInstanceDir(mod)
	instanceCfg := filepath.Join(instDir, "instance.cfg")
	mmcPack := filepath.Join(instDir, "mmc-pack.json")
	return exists(instanceCfg) && exists(mmcPack)
}

func (g *GUI) handlePrimaryAction(mod Modpack) {
	state := g.getModpackState(mod.ID)
	if state == nil {
		g.updateStatus("Checking modpack status...")
		g.refreshModpackState(mod)
		return
	}

	if state.Busy && !state.Running {
		return
	}

	switch state.PrimaryAction() {
	case ActionInstall:
		g.runModpackOperation(mod, ActionInstall)
	case ActionUpdate:
		g.runModpackOperation(mod, ActionUpdate)
	case ActionLaunch:
		g.runModpackOperation(mod, ActionLaunch)
	case ActionKill:
		g.killRunningInstance(mod)
	default:
		// No action available
	}
}

func (g *GUI) handlePrimaryForSelected() {
	if len(g.modpacks) == 0 {
		return
	}
	g.handlePrimaryAction(g.modpacks[0])
}

func (g *GUI) startUpdateCheck() {
	if g.exePath == "" {
		return
	}
	go func() {
		startMsg := "Checking for launcher updates..."
		g.showLoading(true, startMsg)
		err := selfUpdate(g.root, g.exePath, func(msg string) {
			logf("%s", infoLine(msg))
			g.showLoading(true, msg)
		})
		if err != nil {
			g.updateStatus("Update check failed; continuing")
			g.showLoading(false, "")
			return
		}
		g.updateStatus("Launcher ready")
		g.showLoading(false, "")
	}()
}

func (g *GUI) configureRuntimeForModpack(mod Modpack) int {
	memoryMB := MemoryForModpack(mod)
	mode := "manual"
	if settings.AutoRAM {
		mode = "auto"
	}
	modeLabel := strings.Title(mode)
	logf("%s", infoLine(fmt.Sprintf("%s: using %d GB RAM (%s)", mod.DisplayName, memoryMB/1024, modeLabel)))

	if err := updateInstanceMemory(g.modpackInstanceDir(mod), memoryMB); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Warning: failed to update instance memory for %s: %v", mod.DisplayName, err)))
	}

	fyne.Do(func() {
		if g.statusLabel != nil {
			g.statusLabel.SetText(fmt.Sprintf("RAM mode: %s - %d GB", modeLabel, memoryMB/1024))
		}
	})

	g.updateMemorySummaryLabel()

	return memoryMB
}

func (g *GUI) makeProgressCallback(mod Modpack) func(stage string, step, total int) {
	return func(stage string, step, total int) {
		if total <= 0 {
			total = 1
		}
		if step < 0 {
			step = 0
		}
		if step > total {
			step = total
		}

		logf("%s", infoLine(fmt.Sprintf("%s: %s (%d/%d)", mod.DisplayName, stage, step, total)))

		progress := float64(step) / float64(total)

		fyne.Do(func() {
			if g.progressBar != nil {
				g.progressBar.SetValue(progress)
				g.progressBar.Show()
			}
			if g.statusLabel != nil {
				g.statusLabel.SetText(fmt.Sprintf("%s - %s (%d/%d)", mod.DisplayName, stage, step, total))
			}
		})
	}
}

func (g *GUI) memorySummary() string {
	if settings.AutoRAM {
		auto := clampMemoryMB(DefaultAutoMemoryMB())
		return fmt.Sprintf("RAM Mode: Auto (%d GB)", auto/1024)
	}
	return fmt.Sprintf("RAM Mode: Manual (%d GB)", clampMemoryMB(settings.MemoryMB)/1024)
}

func (g *GUI) updateMemorySummaryLabel() {
	if g.memorySummaryLabel == nil {
		return
	}
	fyne.Do(func() {
		g.memorySummaryLabel.SetText(g.memorySummary())
	})
}

func (g *GUI) runModpackOperation(mod Modpack, action PrimaryAction) {
	if action == ActionInstall || action == ActionUpdate || action == ActionLaunch {
		g.configureRuntimeForModpack(mod)
	}

	var statusMsg string
	var logMsg string

	switch action {
	case ActionInstall:
		statusMsg = fmt.Sprintf("Installing %s...", mod.DisplayName)
		logMsg = fmt.Sprintf("Installing modpack: %s", mod.DisplayName)
	case ActionUpdate:
		statusMsg = fmt.Sprintf("Updating %s...", mod.DisplayName)
		logMsg = fmt.Sprintf("Updating modpack: %s", mod.DisplayName)
	case ActionLaunch:
		statusMsg = fmt.Sprintf("Launching %s...", mod.DisplayName)
		logMsg = fmt.Sprintf("Launching modpack: %s", mod.DisplayName)
	default:
		statusMsg = fmt.Sprintf("Working on %s...", mod.DisplayName)
		logMsg = fmt.Sprintf("Working on modpack: %s", mod.DisplayName)
	}

	g.updateStatus(statusMsg)
	logf("%s", infoLine(logMsg))

	g.setModpackState(mod.ID, func(state *ModpackState) {
		state.Busy = true
		state.Running = false
		state.RunningPID = 0
		state.CurrentAction = action
		state.Error = nil
	})

	fyne.Do(func() {
		if g.progressBar != nil {
			g.progressBar.SetValue(0)
			g.progressBar.Show()
		}
	})

	progressCb := g.makeProgressCallback(mod)

	go func(mod Modpack, action PrimaryAction) {
		g.setRunningModpackID(mod.ID)
		go g.monitorProcessStart(mod)

		runLauncherLogic(g.root, g.exePath, mod, g.prismProcess, progressCb)

		g.setRunningModpackID("")

		g.processMu.Lock()
		if g.prismProcess != nil {
			*g.prismProcess = nil
		}
		g.processMu.Unlock()

		g.setModpackState(mod.ID, func(state *ModpackState) {
			state.Running = false
			state.Busy = false
			state.RunningPID = 0
			if state.CurrentAction == action {
				state.CurrentAction = ActionNone
			}
		})

		fyne.Do(func() {
			if g.progressBar != nil {
				g.progressBar.Hide()
				g.progressBar.SetValue(0)
			}
		})

		g.updateStatus("Operation complete")
		g.refreshModpackState(mod)
	}(mod, action)
}

func (g *GUI) monitorProcessStart(mod Modpack) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if g.getRunningModpackID() != mod.ID {
			return
		}

		proc := g.getPrismProcess()
		if proc == nil {
			continue
		}

		g.setModpackState(mod.ID, func(state *ModpackState) {
			state.Running = true
			state.Busy = false
			state.RunningPID = proc.Pid
			if state.CurrentAction == ActionInstall || state.CurrentAction == ActionLaunch || state.CurrentAction == ActionUpdate {
				state.CurrentAction = ActionNone
			}
		})

		g.updateStatus(fmt.Sprintf("Running %s (PID %d)", mod.DisplayName, proc.Pid))
		logf("%s", infoLine(fmt.Sprintf("%s running (PID %d)", mod.DisplayName, proc.Pid)))
		return
	}
}

func (g *GUI) setRunningModpackID(id string) {
	g.runningMu.Lock()
	g.runningModpackID = id
	g.runningMu.Unlock()
}

func (g *GUI) getRunningModpackID() string {
	g.runningMu.RLock()
	defer g.runningMu.RUnlock()
	return g.runningModpackID
}

func (g *GUI) getPrismProcess() *os.Process {
	g.processMu.Lock()
	defer g.processMu.Unlock()
	if g.prismProcess == nil {
		return nil
	}
	return *g.prismProcess
}

func (g *GUI) killRunningInstance(mod Modpack) {
	if g.getRunningModpackID() != mod.ID {
		g.updateStatus("No running process to kill for this modpack")
		return
	}

	proc := g.getPrismProcess()
	if proc == nil {
		g.updateStatus("No running process detected")
		return
	}

	pid := proc.Pid
	logf("%s", infoLine(fmt.Sprintf("Attempting to kill %s (PID %d)", mod.DisplayName, pid)))

	if err := killProcessByPID(pid); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to kill %s: %v", mod.DisplayName, err)))
		g.updateStatus(fmt.Sprintf("Failed to kill %s", mod.DisplayName))
		return
	}

	g.updateStatus(fmt.Sprintf("Kill signal sent to %s", mod.DisplayName))
	logf("%s", successLine(fmt.Sprintf("Kill signal sent to %s (PID %d)", mod.DisplayName, pid)))
	g.setRunningModpackID("")
	g.setModpackState(mod.ID, func(state *ModpackState) {
		state.Running = false
		state.Busy = false
		state.RunningPID = 0
	})
	g.processMu.Lock()
	if g.prismProcess != nil {
		*g.prismProcess = nil
	}
	g.processMu.Unlock()
}

func (g *GUI) deleteModpack(mod Modpack) {
	state := g.getModpackState(mod.ID)
	if state != nil && (state.Busy || state.Running) {
		g.updateStatus("Cannot delete while modpack is busy or running")
		return
	}

	logf("%s", infoLine(fmt.Sprintf("Deleting modpack data: %s", mod.DisplayName)))

	g.setModpackState(mod.ID, func(state *ModpackState) {
		state.Busy = true
		state.CurrentAction = ActionNone
	})

	go func() {
		if err := g.removeModpackData(mod); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to delete %s: %v", mod.DisplayName, err)))
			g.updateStatus(fmt.Sprintf("Delete failed: %v", err))
			g.setModpackState(mod.ID, func(state *ModpackState) {
				state.Busy = false
				state.Error = err
			})
			return
		}

		g.setModpackState(mod.ID, func(state *ModpackState) {
			state.Busy = false
			state.Installed = false
			state.UpdateAvailable = false
			state.LocalVersion = ""
			state.Running = false
			state.RunningPID = 0
			state.Error = nil
		})

		g.updateStatus(fmt.Sprintf("Deleted %s", mod.DisplayName))
		logf("%s", successLine(fmt.Sprintf("Deleted modpack data: %s", mod.DisplayName)))
		g.refreshModpackState(mod)
	}()
}

func (g *GUI) removeModpackData(mod Modpack) error {
	instDir := g.modpackInstanceDir(mod)
	if !exists(instDir) {
		return nil
	}
	return os.RemoveAll(instDir)
}

func (g *GUI) reinstallModpack(mod Modpack) {
	state := g.getModpackState(mod.ID)
	if state != nil && (state.Busy || state.Running) {
		g.updateStatus("Cannot reinstall while modpack is busy or running")
		return
	}

	logf("%s", infoLine(fmt.Sprintf("Reinstalling modpack: %s", mod.DisplayName)))

	g.setModpackState(mod.ID, func(s *ModpackState) {
		s.Busy = true
		s.CurrentAction = ActionInstall
		s.Error = nil
	})

	go func() {
		if err := g.removeModpackData(mod); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to prepare reinstall for %s: %v", mod.DisplayName, err)))
			g.updateStatus(fmt.Sprintf("Reinstall failed: %v", err))
			g.setModpackState(mod.ID, func(s *ModpackState) {
				s.Busy = false
				s.Error = err
			})
			return
		}

		g.setModpackState(mod.ID, func(s *ModpackState) {
			s.Installed = false
			s.UpdateAvailable = false
			s.LocalVersion = ""
		})

		g.runModpackOperation(mod, ActionInstall)
	}()
}

func (g *GUI) filterByCategory(category string) {
	g.activeCategory = category
	switch category {
	case "":
		g.updateStatus("Showing all modpacks")
	case "featured":
		g.updateStatus("Filtering by featured modpacks")
	default:
		g.updateStatus(fmt.Sprintf("Filtering by %s modpacks", category))
	}
	g.applyFilters()
}

func (g *GUI) refreshModpacks() {
	g.updateStatus("Refreshing modpack list...")
	g.showLoading(true, "Refreshing modpacks...")

	go func() {
		time.Sleep(2 * time.Second)
		g.showLoading(false, "")
		g.updateStatus("Modpack list refreshed")
		g.refreshAllModpackStates()
	}()
}

func (g *GUI) updateStatus(text string) {
	if g.statusLabel == nil {
		return
	}
	fyne.Do(func() {
		g.statusLabel.SetText(text)
	})
}

func (g *GUI) showLoading(show bool, message string) {
	if g.loadingOverlay == nil {
		return
	}

	fyne.Do(func() {
		if message != "" && g.loadingLabel != nil {
			g.loadingLabel.SetText(message)
		}
		if show {
			g.loadingOverlay.Show()
		} else {
			g.loadingOverlay.Hide()
		}
	})
}

func (g *GUI) showConsole() {
	if g.tabs != nil {
		for index, tab := range g.tabs.Items {
			if tab.Text == "Console" {
				g.tabs.SelectIndex(index)
				break
			}
		}
	}
}
func (g *GUI) launchSelectedModpackWithFeedback() {
	if len(g.modpacks) == 0 {
		return
	}
	g.handlePrimaryAction(g.modpacks[0])
}

// startLogFileWatcher begins monitoring the latest.log file and piping it to the GUI console
func (g *GUI) startLogFileWatcher() {
	g.logMutex.Lock()
	defer g.logMutex.Unlock()

	if g.logWatcherActive {
		return
	}

	g.logWatcherActive = true
	g.logStopChan = make(chan struct{})

	logPath := filepath.Join(g.root, "logs", "latest.log")

	// Start combined loading and monitoring
	go g.loadAndWatchLogFile(logPath)
}

// stopLogFileWatcher stops the log file monitoring
func (g *GUI) stopLogFileWatcher() {
	g.logMutex.Lock()
	defer g.logMutex.Unlock()

	if g.logWatcherActive && g.logStopChan != nil {
		close(g.logStopChan)
		g.logWatcherActive = false
	}
}

// loadAndWatchLogFile loads existing log content and monitors for new content
func (g *GUI) loadAndWatchLogFile(logPath string) {
	ticker := time.NewTicker(500 * time.Millisecond) // Check every 500ms
	defer ticker.Stop()

	var initialLoadDone = false
	var lastSize int64 = 0

	for {
		select {
		case <-g.logStopChan:
			return
		case <-ticker.C:
			// Check if file exists
			info, err := os.Stat(logPath)
			if err != nil {
				continue
			}

			if !initialLoadDone {
				// Initial load - read entire file once
				file, err := os.Open(logPath)
				if err != nil {
					continue
				}

				content, err := io.ReadAll(file)
				file.Close()

				if err == nil && len(content) > 0 {
					contentStr := string(content)
					fyne.Do(func() {
						if g.consoleOutput != nil {
							// Replace placeholder with actual log content
							g.consoleOutput.SetText(contentStr)
							// Scroll to bottom
							lines := strings.Split(contentStr, "\n")
							g.consoleOutput.CursorRow = len(lines) - 1
						}
					})
				}

				lastSize = info.Size()
				initialLoadDone = true
			} else {
				// Monitoring mode - only read new content when file grows significantly
				// Use a small threshold to avoid reading incomplete writes
				if info.Size() > lastSize+10 { // Only read if file grew by at least 10 bytes
					file, err := os.Open(logPath)
					if err != nil {
						continue
					}

					// Read the entire file to avoid seek issues
					content, err := io.ReadAll(file)
					file.Close()

					if err == nil && len(content) > 0 {
						contentStr := string(content)

						// Only update if content actually changed and is longer
						fyne.Do(func() {
							if g.consoleOutput != nil && len(contentStr) > len(g.consoleOutput.Text) {
								g.consoleOutput.SetText(contentStr)
								// Scroll to bottom
								lines := strings.Split(contentStr, "\n")
								g.consoleOutput.CursorRow = len(lines) - 1
							}
						})
					}

					lastSize = info.Size()
				}
			}
		}
	}
}

// generateRandomID generates a random 8-character hexadecimal string
func generateRandomID() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3]), nil
}

// uploadLog uploads the latest.log content to i.dylan.lol/logs/
func (g *GUI) uploadLog() {
	logPath := filepath.Join(g.root, "logs", "latest.log")

	// Show upload progress dialog
	progressDialog := dialog.NewCustomWithoutButtons("Uploading Log...",
		widget.NewProgressBarInfinite(), g.window)
	progressDialog.Show()

	// Upload in background goroutine
	go func() {
		defer progressDialog.Hide()

		// Generate a random 8-character ID for the filename
		randomID, err := generateRandomID()
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to generate random ID: %v", err), g.window)
			})
			return
		}
		filename := fmt.Sprintf("%s.log", randomID)

		// Create multipart form with file upload using CreateFormFile to match curl -F format
		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// Add the required "act" field with value "bput" as required by the endpoint
		err = writer.WriteField("act", "bput")
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to add act field: %v", err), g.window)
			})
			return
		}

		// Open the log file for reading
		file, err := os.Open(logPath)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to open log file for upload: %v", err), g.window)
			})
			return
		}
		defer file.Close()

		// Create form file part using CreateFormFile to match curl -F "file=@filename;type=application/octet-stream"
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to create form file: %v", err), g.window)
			})
			return
		}

		// Copy file content to the form part
		_, err = io.Copy(part, file)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to copy file content: %v", err), g.window)
			})
			return
		}

		writer.Close()

		// Create a new HTTP request with the form data
		req, err := http.NewRequest("POST", "https://i.dylan.lol/logs/", &requestBody)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to create request: %v", err), g.window)
			})
			return
		}

		// Set the content type header for form data
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

		// Send the request with TLS 1.2 and timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to upload log: %v", err), g.window)
			})
			return
		}
		defer resp.Body.Close()

		// Read the full response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Failed to read response body: %v", err), g.window)
			})
			return
		}

		// Log the raw response for debugging
		logf("Upload response status: %s", resp.Status)
		logf("Upload response content-type: %s", resp.Header.Get("Content-Type"))
		bodyStr := string(body)
		logf("Upload response body (first 200 chars): %s", bodyStr[:min(200, len(bodyStr))])

		// Log that we're parsing HTML instead of JSON
		logf("Parsing HTML response to extract log file URL")

		// Check if the upload was successful (status code 200-299)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Parse the HTML response to extract the file URL
			bodyStr := string(body)
			var logURL string

			// Try to extract the filename from the HTML response using regex
			// Pattern to match href="/logs/filename.log"
			logPattern := `href="/logs/([^"]+\.log)"`
			re := regexp.MustCompile(logPattern)
			matches := re.FindStringSubmatch(bodyStr)

			if len(matches) > 1 {
				// Extract the filename from the match
				filename := matches[1]
				// Construct the full URL
				logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", filename)
				logf("Successfully extracted filename from HTML: %s", filename)
			} else {
				// If regex fails, fall back to using our random ID
				logf("Failed to extract filename from HTML, falling back to random ID: %s", randomID)
				logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s.log", randomID)
			}

			logf("Final log URL: %s", logURL)

			fyne.Do(func() {
				// Extract filename from the URL for display
				filename := randomID + ".log"
				if matches := regexp.MustCompile(`https://i\.dylan\.lol/logs/([^\.]+\.log)`).FindStringSubmatch(logURL); len(matches) > 1 {
					filename = matches[1]
				}

				// Create a more informative and visually appealing success dialog
				successTitle := widget.NewLabelWithStyle("✓ Log Successfully Uploaded!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
				successTitle.Importance = widget.HighImportance

				// Main message with clipboard information
				messageLabel := widget.NewLabel("The log has been uploaded and the URL has been copied to your clipboard.")
				messageLabel.Wrapping = fyne.TextWrapWord

				// File information
				fileInfoLabel := widget.NewLabelWithStyle(fmt.Sprintf("File: %s", filename), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

				// Parse URL for hyperlink
				parsedURL, err := url.Parse(logURL)
				var successContent fyne.CanvasObject

				if err != nil {
					// If URL parsing fails, just show the text
					urlLabel := widget.NewLabelWithStyle(fmt.Sprintf("URL: %s", logURL), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
					urlLabel.Wrapping = fyne.TextWrapWord

					successContent = container.NewVBox(
						successTitle,
						widget.NewSeparator(),
						messageLabel,
						widget.NewSeparator(),
						fileInfoLabel,
						urlLabel,
					)
				} else {
					// Create a clickable link to the uploaded log
					linkLabel := widget.NewLabelWithStyle("Click the link below to view your uploaded log:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
					hyperlink := widget.NewHyperlink("View Uploaded Log", parsedURL)

					successContent = container.NewVBox(
						successTitle,
						widget.NewSeparator(),
						messageLabel,
						widget.NewSeparator(),
						fileInfoLabel,
						linkLabel,
						hyperlink,
					)
				}

				// Create buttons for the dialog
				okButton := widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {})
				copyButton := widget.NewButtonWithIcon("Copy URL Again", theme.ContentCopyIcon(), func() {
					g.window.Clipboard().SetContent(logURL)
					g.updateStatus("URL copied to clipboard")
				})

				// Button container
				buttonContainer := container.NewHBox(
					layout.NewSpacer(),
					copyButton,
					okButton,
				)

				// Complete dialog with buttons
				dialogContent := container.NewVBox(
					successContent,
					widget.NewSeparator(),
					buttonContainer,
				)

				// Show the custom dialog
				dialog.ShowCustom("Upload Successful", "", dialogContent, g.window)

				// Auto-copy to clipboard (existing functionality)
				g.window.Clipboard().SetContent(logURL)
			})
		} else {
			// Upload failed
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Upload failed with status: %s\nResponse: %s", resp.Status, string(body)[:min(200, len(body))]), g.window)
			})
		}
	}()
}

// extractFileIDFromHTML extracts the file ID from MicroBin's HTML response
func (g *GUI) extractFileIDFromHTML(html string) string {
	// Look for the pattern in the HTML that contains the file ID
	// The file ID appears in URLs like https://logs.dylan.lol/upload/mouse-tiger-fly or https://logs.dylan.lol/file/mouse-tiger-fly

	// First try to find the upload URL pattern (absolute URLs)
	uploadPattern := `href="https://logs\.dylan\.lol/upload/([^"]+)"`
	re := regexp.MustCompile(uploadPattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the file URL pattern (absolute URLs)
	filePattern := `href="https://logs\.dylan\.lol/file/([^"]+)"`
	re = regexp.MustCompile(filePattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the edit URL pattern (absolute URLs)
	editPattern := `href="https://logs\.dylan\.lol/edit/([^"]+)"`
	re = regexp.MustCompile(editPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the JavaScript URL pattern
	jsPattern := `const url = .*https://logs\.dylan\.lol/upload/([^"]+)`
	re = regexp.MustCompile(jsPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: try relative patterns (in case the format changes)
	relativeUploadPattern := `href="/upload/([^"]+)"`
	re = regexp.MustCompile(relativeUploadPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	relativeFilePattern := `href="/file/([^"]+)"`
	re = regexp.MustCompile(relativeFilePattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (g *GUI) showSettings() {
	memLabel := widget.NewLabel("")

	// Current settings values
	autoCheck := widget.NewCheck("Enable Auto RAM", nil)
	autoCheck.SetChecked(settings.AutoRAM)

	memSlider := widget.NewSlider(2, 16)
	memSlider.Step = 1
	memSlider.SetValue(float64(clampMemoryMB(settings.MemoryMB) / 1024))

	// Dev builds checkbox
	devCheck := widget.NewCheck("Enable dev builds (pre-release)", nil)
	devCheck.SetChecked(settings.DevBuildsEnabled)

	// Current channel status label
	channelLabel := widget.NewLabel("")
	if settings.DevBuildsEnabled {
		channelLabel.SetText("Channel: Dev")
	} else {
		channelLabel.SetText("Channel: Stable")
	}

	refreshUI := func() {
		if settings.AutoRAM {
			memLabel.SetText(fmt.Sprintf("Auto RAM baseline: %d GB", DefaultAutoMemoryMB()/1024))
			memSlider.Hide()
		} else {
			memSlider.Show()
			memSlider.SetValue(float64(clampMemoryMB(settings.MemoryMB) / 1024))
			memLabel.SetText(fmt.Sprintf("Manual RAM: %d GB", settings.MemoryMB/1024))
		}
	}

	autoCheck.OnChanged = func(on bool) {
		settings.AutoRAM = on
		if on {
			settings.MemoryMB = clampMemoryMB(DefaultAutoMemoryMB())
		} else {
			settings.MemoryMB = clampMemoryMB(int(memSlider.Value) * 1024)
		}
		refreshUI()
	}

	memSlider.OnChanged = func(v float64) {
		if settings.AutoRAM {
			return
		}
		settings.MemoryMB = clampMemoryMB(int(v) * 1024)
		memLabel.SetText(fmt.Sprintf("Manual RAM: %.0f GB", v))
	}

	// Update channel label when dev mode checkbox is toggled
	devCheck.OnChanged = func(on bool) {
		if on {
			channelLabel.SetText("Channel: Dev")
		} else {
			channelLabel.SetText("Channel: Stable")
		}
	}

	refreshUI()

	dialogContent := container.NewVBox(
		widget.NewLabel("Launcher Settings"),
		autoCheck,
		devCheck,
		container.NewHBox(channelLabel, layout.NewSpacer()),
		memLabel,
		memSlider,
	)

	pop := widget.NewModalPopUp(container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		dialogContent,
	), g.window.Canvas())

	// Save & Apply button that handles all changes
	saveApplyBtn := widget.NewButtonWithIcon("Save & Apply", theme.DocumentSaveIcon(), func() {
		// Close the settings dialog immediately
		pop.Hide()

		// Show loading in main UI instead of dialog
		g.showLoading(true, "Applying settings...")
		g.updateStatus("Applying settings...")

		go func() {
			defer g.showLoading(false, "")

			// Handle dev mode changes with validation
			if devCheck.Checked != settings.DevBuildsEnabled {
				g.updateStatus("Validating update availability...")

				// Pre-update validation: check if the target version is available
				targetDevMode := devCheck.Checked
				var validationErr error

				if targetDevMode {
					// Check if dev builds are available
					_, _, validationErr = FetchLatestAssetPreferPrerelease(UPDATE_OWNER, UPDATE_REPO, LauncherAssetName, true)
				} else {
					// Check if stable builds are available
					_, _, validationErr = FetchLatestAssetPreferPrerelease(UPDATE_OWNER, UPDATE_REPO, LauncherAssetName, false)
				}

				if validationErr != nil {
					logf("%s", warnLine(fmt.Sprintf("Update validation failed: %v", validationErr)))
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("Failed to validate update availability: %v\n\nPlease check your internet connection and try again.", validationErr), g.window)
						// Revert checkbox to current state
						devCheck.SetChecked(settings.DevBuildsEnabled)
					})
					return
				}

				// Apply dev mode change
				settings.DevBuildsEnabled = targetDevMode
				logf("%s", infoLine(fmt.Sprintf("GUI: User %s dev builds", map[bool]string{true: "enabled", false: "disabled"}[targetDevMode])))

				// Save settings before update
				if err := saveSettings(g.root); err != nil {
					logf("%s", warnLine(fmt.Sprintf("Failed to save settings: %v", err)))
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("Failed to save settings: %v", err), g.window)
						// Revert changes
						settings.DevBuildsEnabled = !targetDevMode
						devCheck.SetChecked(settings.DevBuildsEnabled)
					})
					return
				}

				// Force update to the target channel
				g.updateStatus(fmt.Sprintf("Updating to latest %s version...", map[bool]string{true: "dev", false: "stable"}[targetDevMode]))
				updateErr := forceUpdate(g.root, g.exePath, targetDevMode, func(msg string) {
					logf("%s", infoLine(msg))
					fyne.Do(func() {
						g.updateStatus(msg)
					})
				})

				if updateErr != nil {
					logf("%s", warnLine(fmt.Sprintf("Failed to update to %s version: %v", map[bool]string{true: "dev", false: "stable"}[targetDevMode], updateErr)))

					// Fallback: if dev update failed, try to fallback to stable
					if targetDevMode {
						logf("%s", infoLine("Attempting fallback to stable channel..."))
						fyne.Do(func() {
							g.updateStatus("Attempting fallback to stable...")
						})
						fallbackErr := forceUpdate(g.root, g.exePath, false, func(msg string) {
							logf("%s", infoLine(fmt.Sprintf("Fallback: %s", msg)))
							fyne.Do(func() {
								g.updateStatus(msg)
							})
						})

						if fallbackErr != nil {
							logf("%s", warnLine(fmt.Sprintf("Fallback to stable also failed: %v", fallbackErr)))
							fyne.Do(func() {
								dialog.ShowError(fmt.Errorf("Failed to update to dev version and fallback to stable also failed.\n\nDev error: %v\nFallback error: %v\n\nPlease check your internet connection and try again.", updateErr, fallbackErr), g.window)
								// Revert to original state
								settings.DevBuildsEnabled = !targetDevMode
								devCheck.SetChecked(settings.DevBuildsEnabled)
								saveSettings(g.root)
							})
						} else {
							logf("%s", successLine("Successfully fell back to stable channel"))
							fyne.Do(func() {
								dialog.ShowInformation("Update Fallback", "Failed to update to dev version, but successfully fell back to stable channel.\n\nDev builds have been disabled.", g.window)
								settings.DevBuildsEnabled = false
								devCheck.SetChecked(false)
								saveSettings(g.root)
							})
						}
					} else {
						// Stable update failed - no fallback needed
						fyne.Do(func() {
							dialog.ShowError(fmt.Errorf("Failed to update to stable version: %v\n\nPlease check your internet connection and try again.", updateErr), g.window)
							// Revert to original state
							settings.DevBuildsEnabled = !targetDevMode
							devCheck.SetChecked(settings.DevBuildsEnabled)
							saveSettings(g.root)
						})
					}
					return
				}

				logf("%s", successLine(fmt.Sprintf("Successfully updated to %s channel", map[bool]string{true: "dev", false: "stable"}[targetDevMode])))
			}

			// Save RAM settings
			if err := saveSettings(g.root); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to save settings: %v", err)))
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Failed to save settings: %v", err), g.window)
				})
				return
			}

			g.updateMemorySummaryLabel()

			fyne.Do(func() {
				g.updateStatus("Settings applied successfully")
			})
		}()
	})

	dialogContent.Add(container.NewHBox(layout.NewSpacer(), saveApplyBtn))
	pop.Show()
}

// Legacy compatibility helpers ------------------------------------------------

func (g *GUI) createMainContent() {
	g.buildUI()
}

func (g *GUI) createModpackSelection() fyne.CanvasObject {
	return g.buildContent()
}

func (g *GUI) createConsoleView() fyne.CanvasObject {
	return g.buildConsoleView()
}
