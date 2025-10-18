package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func loadModpacks(root string) []Modpack {
	remote, err := fetchRemoteModpacks(remoteModpacksURL, 30*time.Second)
	if err != nil {
		fail(fmt.Errorf("failed to fetch remote modpacks.json: %w", err))
	}

	if len(remote) == 0 {
		fail(errors.New("remote modpacks.json returned no modpacks"))
	}

	normalized := normalizeModpacks(remote)
	if len(normalized) == 0 {
		fail(errors.New("remote modpacks.json did not contain any valid modpacks"))
	}

	logf("%s", successLine(fmt.Sprintf("Loaded %d modpack(s) from remote catalog", len(normalized))))
	updateDefaultModpackID(normalized)
	return normalized
}

func selectModpack(modpacks []Modpack, requestedID string) (Modpack, error) {
	if len(modpacks) == 0 {
		return Modpack{}, errors.New("no modpacks available")
	}

	if strings.TrimSpace(requestedID) == "" {
		for _, mp := range modpacks {
			if strings.EqualFold(mp.ID, defaultModpackID) {
				return mp, nil
			}
		}
		return modpacks[0], nil
	}

	id := strings.ToLower(strings.TrimSpace(requestedID))
	for _, mp := range modpacks {
		if strings.ToLower(mp.ID) == id {
			return mp, nil
		}
	}

	return Modpack{}, fmt.Errorf("unknown modpack %q. Use --list-modpacks to view available options.", requestedID)
}

func printModpackList(modpacks []Modpack) {
	fmt.Fprintln(os.Stdout, "Available modpacks:")
	currentDefault := strings.ToLower(defaultModpackID)
	for _, mp := range modpacks {
		label := mp.DisplayName
		if strings.ToLower(mp.ID) == currentDefault {
			label += " [default]"
		}
		desc := strings.TrimSpace(mp.Description)
		if desc == "" {
			desc = "(no description provided)"
		}
		fmt.Fprintf(os.Stdout, " - %s (%s)\n   %s\n", label, mp.ID, desc)
	}
}

func updateDefaultModpackID(modpacks []Modpack) {
	if len(modpacks) == 0 {
		return
	}
	for _, mp := range modpacks {
		if mp.Default {
			defaultModpackID = mp.ID
			return
		}
	}
	defaultModpackID = modpacks[0].ID
}

func fetchRemoteModpacks(url string, timeout time.Duration) ([]Modpack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", getUserAgent("Launcher"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mods []Modpack
	if err := json.Unmarshal(body, &mods); err != nil {
		return nil, err
	}

	return normalizeModpacks(mods), nil
}

func normalizeModpacks(mods []Modpack) []Modpack {
	if len(mods) == 0 {
		return nil
	}

	normalized := make([]Modpack, 0, len(mods))
	index := make(map[string]int, len(mods))

	for _, raw := range mods {
		id := strings.TrimSpace(raw.ID)
		packURL := strings.TrimSpace(raw.PackURL)
		instance := strings.TrimSpace(raw.InstanceName)

		if id == "" || packURL == "" || instance == "" {
			continue
		}

		display := strings.TrimSpace(raw.DisplayName)
		if display == "" {
			display = id
		}

		entry := Modpack{
			ID:           id,
			DisplayName:  display,
			PackURL:      packURL,
			InstanceName: instance,
			Description:  strings.TrimSpace(raw.Description),
			Default:      raw.Default,
		}

		key := strings.ToLower(id)
		if idx, ok := index[key]; ok {
			normalized[idx] = entry
		} else {
			index[key] = len(normalized)
			normalized = append(normalized, entry)
		}
	}

	return normalized
}

// -------------------- TUI for Modpack Selection --------------------

func runLauncherTUI(modpacks []Modpack, initial Modpack) (Modpack, bool, error) {
	if len(modpacks) == 0 {
		return Modpack{}, false, errors.New("no modpacks available")
	}

	if len(modpacks) == 1 {
		return modpacks[0], true, nil
	}

	defaultIndex := 0
	for i, mp := range modpacks {
		if strings.EqualFold(mp.ID, initial.ID) {
			defaultIndex = i
			break
		}
	}

	model := newTUIModel(modpacks, defaultIndex)
	prog := tea.NewProgram(model, tea.WithAltScreen())
	res, err := prog.Run()
	if err != nil {
		return Modpack{}, false, err
	}

	finalModel := res.(tuiModel)

	// Check if user selected "Back" option
	if finalModel.isBack {
		return Modpack{}, false, nil // Return empty modpack and false for back navigation
	}

	return finalModel.selected, finalModel.confirmed, nil
}

type tuiModel struct {
	list      list.Model
	selected  Modpack
	confirmed bool
	isBack    bool
}

func newTUIModel(modpacks []Modpack, defaultIndex int) tuiModel {
	// Create items: back button + modpacks
	items := make([]list.Item, len(modpacks)+1)
	items[0] = backMenuItem{title: "← Back to Main Menu", description: "Return to the main menu"}

	for i, mp := range modpacks {
		items[i+1] = modpackListItem{modpack: mp}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bfc7ff"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#14141f")).Background(lipgloss.Color("#8be9fd")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2f9"))

	l := list.New(items, delegate, 0, 0)
	l.Title = fmt.Sprintf("%s Modpacks", launcherShortName)
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true).PaddingLeft(1)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.InfiniteScrolling = false
	l.Select(defaultIndex + 1) // +1 to account for back button

	selectedModpack := Modpack{}
	if len(modpacks) > 0 {
		selectedModpack = modpacks[defaultIndex]
	}

	return tuiModel{
		list:     l,
		selected: selectedModpack,
		isBack:   false,
	}
}

func (m tuiModel) Init() tea.Cmd { return nil }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			if _, ok := m.list.SelectedItem().(backMenuItem); ok {
				// Back button selected
				m.isBack = true
				m.confirmed = false
			} else if item, ok := m.list.SelectedItem().(modpackListItem); ok {
				// Modpack selected
				m.selected = item.modpack
				m.confirmed = true
				m.isBack = false
			}
			return m, tea.Quit
		case "b":
			// Quick back shortcut
			m.isBack = true
			m.confirmed = false
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		height := msg.Height - 5
		if height < 5 {
			height = 5
		}
		width := msg.Width - 6
		if width < 40 {
			width = 40
		}
		m.list.SetSize(width, height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Update selected item tracking
	if _, ok := m.list.SelectedItem().(backMenuItem); ok {
		m.isBack = false
		m.selected = Modpack{}
	} else if item, ok := m.list.SelectedItem().(modpackListItem); ok {
		m.selected = item.modpack
		m.isBack = false
	}

	return m, cmd
}

func (m tuiModel) View() string {
	frame := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475a")).
		Render(m.list.View())

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8be9fd")).
		MarginTop(1).
		Render("↑/↓ navigate   •   Enter select   •   b back   •   q quit")

	return frame + "\n" + help
}

type modpackListItem struct {
	modpack Modpack
}

func (i modpackListItem) Title() string {
	title := strings.TrimSpace(i.modpack.DisplayName)
	if title == "" {
		title = i.modpack.ID
	}
	return title
}

func (i modpackListItem) Description() string {
	desc := strings.TrimSpace(i.modpack.Description)
	if desc == "" {
		desc = "ID: " + i.modpack.ID
	} else {
		desc = fmt.Sprintf("%s — ID: %s", desc, i.modpack.ID)
	}
	return desc
}

func (i modpackListItem) FilterValue() string {
	return strings.ToLower(i.modpack.ID + " " + i.modpack.DisplayName + " " + i.modpack.Description)
}

type backMenuItem struct {
	title       string
	description string
}

func (i backMenuItem) Title() string        { return i.title }
func (i backMenuItem) Description() string { return i.description }
func (i backMenuItem) FilterValue() string { return strings.ToLower(i.title + " " + i.description) }