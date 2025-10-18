// Package widgets provides custom GUI widgets for the TheBoys Launcher
package widgets

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"theboys-launcher/internal/modpack"
)

// ModpackCard represents a card widget for displaying modpack information
type ModpackCard struct {
	widget.BaseWidget
	modpack     *modpack.Modpack
	onInstall   func(*modpack.Modpack)
	onLaunch    func(*modpack.Modpack)
	onUninstall func(*modpack.Modpack)
	onUpdate    func(*modpack.Modpack)
}

// NewModpackCard creates a new modpack card widget
func NewModpackCard(modpack *modpack.Modpack) *ModpackCard {
	card := &ModpackCard{
		modpack: modpack,
	}
	card.ExtendBaseWidget(card)
	return card
}

// CreateRenderer creates the renderer for the modpack card
func (mc *ModpackCard) CreateRenderer() fyne.WidgetRenderer {
	// Create icon placeholder
	icon := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	icon.SetMinSize(fyne.NewSize(64, 64))

	// Create labels
	nameLabel := widget.NewLabelWithStyle(mc.modpack.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	authorLabel := widget.NewLabelWithStyle("By "+mc.modpack.Author, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	// Truncate description if too long
	description := mc.modpack.Description
	if len(description) > 100 {
		description = description[:97] + "..."
	}
	descLabel := widget.NewLabel(description)

	// Create version and status info
	versionLabel := widget.NewLabel(fmt.Sprintf("Version: %s", mc.modpack.Version))
	mcVersionLabel := widget.NewLabel(fmt.Sprintf("Minecraft: %s", mc.modpack.MinecraftVersion.ID))
	modLoaderLabel := widget.NewLabel(fmt.Sprintf("Loader: %s %s", mc.modpack.ModLoader.Type, mc.modpack.ModLoader.Version))

	// Create status label with color
	var statusLabel *widget.Label
	var statusColor color.NRGBA
	switch mc.modpack.Status {
	case modpack.StatusInstalled:
		statusLabel = widget.NewLabelWithStyle("✓ Installed", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		statusColor = color.NRGBA{R: 0, G: 128, B: 0, A: 255} // Green
	case modpack.StatusNotInstalled:
		statusLabel = widget.NewLabel("Not Installed")
		statusColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Gray
	case modpack.StatusUpdateAvailable:
		statusLabel = widget.NewLabelWithStyle("↑ Update Available", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		statusColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255} // Blue
	case modpack.StatusDownloading:
		statusLabel = widget.NewLabelWithStyle("⬇ Downloading...", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		statusColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	case modpack.StatusInstalling:
		statusLabel = widget.NewLabelWithStyle("⚙ Installing...", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		statusColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	case modpack.StatusError:
		statusLabel = widget.NewLabelWithStyle("✗ Error", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		statusColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // Red
	default:
		statusLabel = widget.NewLabel(string(mc.modpack.Status))
		statusColor = theme.Color(theme.ColorNameForeground)
	}

	// Create size label
	sizeLabel := widget.NewLabel(fmt.Sprintf("Size: %s", mc.modpack.GetFormattedSize()))

	// Create memory requirement label
	memLabel := widget.NewLabel(fmt.Sprintf("Memory: %d MB", mc.modpack.RequiredMemory))

	// Create action buttons based on status
	var actionButtons *container.HBox
	switch mc.modpack.Status {
	case modpack.StatusNotInstalled:
		installBtn := widget.NewButton("Install", func() {
			if mc.onInstall != nil {
				mc.onInstall(mc.modpack)
			}
		})
		actionButtons = container.NewHBox(installBtn)
	case modpack.StatusInstalled:
		launchBtn := widget.NewButton("Launch", func() {
			if mc.onLaunch != nil {
				mc.onLaunch(mc.modpack)
			}
		})
		uninstallBtn := widget.NewButton("Uninstall", func() {
			if mc.onUninstall != nil {
				mc.onUninstall(mc.modpack)
			}
		})
		actionButtons = container.NewHBox(launchBtn, uninstallBtn)
	case modpack.StatusUpdateAvailable:
		updateBtn := widget.NewButton("Update", func() {
			if mc.onUpdate != nil {
				mc.onUpdate(mc.modpack)
			}
		})
		launchBtn := widget.NewButton("Launch", func() {
			if mc.onLaunch != nil {
				mc.onLaunch(mc.modpack)
			}
		})
		actionButtons = container.NewHBox(updateBtn, launchBtn)
	default:
		// For downloading/installing/error states, show no action buttons
		actionButtons = container.NewHBox()
	}

	// Create info container
	infoContainer := container.NewVBox(
		nameLabel,
		authorLabel,
		widget.NewSeparator(),
		descLabel,
		widget.NewSeparator(),
		container.NewHBox(versionLabel, mcVersionLabel),
		container.NewHBox(modLoaderLabel, sizeLabel),
		container.NewHBox(statusLabel, memLabel),
		actionButtons,
	)

	// Create main container
	mainContainer := container.NewHBox(
		icon,
		widget.NewScrollContainer(infoContainer),
	)

	return &modpackCardRenderer{
		card:     mc,
		container: mainContainer,
		objects:  []fyne.CanvasObject{mainContainer},
	}
}

// SetOnInstall sets the callback for install button
func (mc *ModpackCard) SetOnInstall(callback func(*modpack.Modpack)) {
	mc.onInstall = callback
}

// SetOnLaunch sets the callback for launch button
func (mc *ModpackCard) SetOnLaunch(callback func(*modpack.Modpack)) {
	mc.onLaunch = callback
}

// SetOnUninstall sets the callback for uninstall button
func (mc *ModpackCard) SetOnUninstall(callback func(*modpack.Modpack)) {
	mc.onUninstall = callback
}

// SetOnUpdate sets the callback for update button
func (mc *ModpackCard) SetOnUpdate(callback func(*modpack.Modpack)) {
	mc.onUpdate = callback
}

// Refresh refreshes the widget
func (mc *ModpackCard) Refresh() {
	mc.BaseWidget.Refresh()
}

// modpackCardRenderer implements the widget renderer
type modpackCardRenderer struct {
	card      *ModpackCard
	container *container.Container
	objects   []fyne.CanvasObject
}

// Layout implements the renderer layout
func (r *modpackCardRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

// MinSize implements the renderer min size
func (r *modpackCardRenderer) MinSize() fyne.Size {
	return r.container.MinSize()
}

// Refresh implements the renderer refresh
func (r *modpackCardRenderer) Refresh() {
	canvas.Refresh(r.container)
}

// Objects implements the renderer objects
func (r *modpackCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy implements the renderer destroy
func (r *modpackCardRenderer) Destroy() {}

// ModpackList represents a list of modpack cards
type ModpackList struct {
	widget.BaseWidget
	modpacks  []*modpack.Modpack
	onSelect  func(*modpack.Modpack)
	container *container.VBox
	scroll    *container.Scroll
}

// NewModpackList creates a new modpack list widget
func NewModpackList() *ModpackList {
	list := &ModpackList{}
	list.ExtendBaseWidget(list)

	// Create container
	vbox := container.NewVBox()
	list.container = vbox
	list.scroll = container.NewScroll(vbox)

	return list
}

// CreateRenderer creates the renderer for the modpack list
func (ml *ModpackList) CreateRenderer() fyne.WidgetRenderer {
	return &modpackListRenderer{
		list:     ml,
		scroll:   ml.scroll,
		objects:  []fyne.CanvasObject{ml.scroll},
	}
}

// SetModpacks sets the modpacks to display
func (ml *ModpackList) SetModpacks(modpacks []*modpack.Modpack) {
	ml.modpacks = modpacks
	ml.refreshList()
}

// AddModpack adds a modpack to the list
func (ml *ModpackList) AddModpack(modpack *modpack.Modpack) {
	ml.modpacks = append(ml.modpacks, modpack)
	ml.refreshList()
}

// RemoveModpack removes a modpack from the list
func (ml *ModpackList) RemoveModpack(modpackID string) {
	for i, modpack := range ml.modpacks {
		if modpack.ID == modpackID {
			ml.modpacks = append(ml.modpacks[:i], ml.modpacks[i+1:]...)
			break
		}
	}
	ml.refreshList()
}

// SetOnSelect sets the callback for modpack selection
func (ml *ModpackList) SetOnSelect(callback func(*modpack.Modpack)) {
	ml.onSelect = callback
}

// refreshList refreshes the displayed list
func (ml *ModpackList) refreshList() {
	// Clear existing items
	ml.container.Objects = nil

	// Add modpack cards
	for _, modpack := range ml.modpacks {
		card := NewModpackCard(modpack)

		// Set callbacks
		card.SetOnInstall(func(m *modpack.Modpack) {
			if ml.onSelect != nil {
				ml.onSelect(m)
			}
		})
		card.SetOnLaunch(func(m *modpack.Modpack) {
			if ml.onSelect != nil {
				ml.onSelect(m)
			}
		})
		card.SetOnUninstall(func(m *modpack.Modpack) {
			if ml.onSelect != nil {
				ml.onSelect(m)
			}
		})
		card.SetOnUpdate(func(m *modpack.Modpack) {
			if ml.onSelect != nil {
				ml.onSelect(m)
			}
		})

		ml.container.Add(card)
		ml.container.Add(widget.NewSeparator())
	}

	ml.container.Refresh()
}

// Refresh refreshes the widget
func (ml *ModpackList) Refresh() {
	ml.BaseWidget.Refresh()
}

// modpackListRenderer implements the widget renderer
type modpackListRenderer struct {
	list     *ModpackList
	scroll   *container.Scroll
	objects  []fyne.CanvasObject
}

// Layout implements the renderer layout
func (r *modpackListRenderer) Layout(size fyne.Size) {
	r.scroll.Resize(size)
}

// MinSize implements the renderer min size
func (r *modpackListRenderer) MinSize() fyne.Size {
	return r.scroll.MinSize()
}

// Refresh implements the renderer refresh
func (r *modpackListRenderer) Refresh() {
	canvas.Refresh(r.scroll)
}

// Objects implements the renderer objects
func (r *modpackListRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy implements the renderer destroy
func (r *modpackListRenderer) Destroy() {}

// SearchBar represents a search bar for modpacks
type SearchBar struct {
	widget.BaseWidget
	entry    *widget.Entry
	onSearch func(string)
}

// NewSearchBar creates a new search bar widget
func NewSearchBar() *SearchBar {
	searchBar := &SearchBar{}
	searchBar.ExtendBaseWidget(searchBar)

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Search modpacks...")
	entry.OnChanged = func(query string) {
		if searchBar.onSearch != nil {
			searchBar.onSearch(query)
		}
	}

	searchBar.entry = entry

	return searchBar
}

// CreateRenderer creates the renderer for the search bar
func (sb *SearchBar) CreateRenderer() fyne.WidgetRenderer {
	searchIcon := widget.NewIcon(theme.SearchIcon())
	container := container.NewBorder(nil, nil, searchIcon, nil, sb.entry)

	return &searchBarRenderer{
		bar:       sb,
		container: container,
		objects:   []fyne.CanvasObject{container},
	}
}

// SetOnSearch sets the callback for search
func (sb *SearchBar) SetOnSearch(callback func(string)) {
	sb.onSearch = callback
}

// GetText returns the current search text
func (sb *SearchBar) GetText() string {
	return sb.entry.Text
}

// SetText sets the search text
func (sb *SearchBar) SetText(text string) {
	sb.entry.SetText(text)
}

// Clear clears the search text
func (sb *SearchBar) Clear() {
	sb.entry.SetText("")
}

// searchBarRenderer implements the widget renderer
type searchBarRenderer struct {
	bar       *SearchBar
	container *container.Container
	objects   []fyne.CanvasObject
}

// Layout implements the renderer layout
func (r *searchBarRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

// MinSize implements the renderer min size
func (r *searchBarRenderer) MinSize() fyne.Size {
	return r.container.MinSize()
}

// Refresh implements the renderer refresh
func (r *searchBarRenderer) Refresh() {
	canvas.Refresh(r.container)
}

// Objects implements the renderer objects
func (r *searchBarRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy implements the renderer destroy
func (r *searchBarRenderer) Destroy() {}

// FilterButtons represents filter buttons for modpack categories
type FilterButtons struct {
	widget.BaseWidget
	buttons   []*widget.Button
	container *container.HBox
	onFilter  func(string)
}

// NewFilterButtons creates a new filter buttons widget
func NewFilterButtons() *FilterButtons {
	filterButtons := &FilterButtons{}
	filterButtons.ExtendBaseWidget(filterButtons)

	// Create filter buttons
	filters := []string{"All", "Installed", "Not Installed", "Updates Available"}
	buttons := make([]*widget.Button, len(filters))

	for i, filter := range filters {
		btn := widget.NewButton(filter, func() {
			// Handle filter selection
			for j, b := range buttons {
				if b == btn {
					b.Importance = widget.HighImportance
				} else {
					b.Importance = widget.MediumImportance
				}
				b.Refresh()
			}

			if filterButtons.onFilter != nil {
				filterButtons.onFilter(filter)
			}
		})

		if i == 0 {
			btn.Importance = widget.HighImportance // Select "All" by default
		}

		buttons[i] = btn
	}

	filterButtons.buttons = buttons
	filterButtons.container = container.NewHBox(buttons...)

	return filterButtons
}

// CreateRenderer creates the renderer for the filter buttons
func (fb *FilterButtons) CreateRenderer() fyne.WidgetRenderer {
	return &filterButtonsRenderer{
		buttons:   fb,
		container: fb.container,
		objects:   []fyne.CanvasObject{fb.container},
	}
}

// SetOnFilter sets the callback for filter selection
func (fb *FilterButtons) SetOnFilter(callback func(string)) {
	fb.onFilter = callback
}

// filterButtonsRenderer implements the widget renderer
type filterButtonsRenderer struct {
	buttons   *FilterButtons
	container *container.Container
	objects   []fyne.CanvasObject
}

// Layout implements the renderer layout
func (r *filterButtonsRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

// MinSize implements the renderer min size
func (r *filterButtonsRenderer) MinSize() fyne.Size {
	return r.container.MinSize()
}

// Refresh implements the renderer refresh
func (r *filterButtonsRenderer) Refresh() {
	canvas.Refresh(r.container)
}

// Objects implements the renderer objects
func (r *filterButtonsRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy implements the renderer destroy
func (r *filterButtonsRenderer) Destroy() {}