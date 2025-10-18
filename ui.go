package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// global writer used by log/fail and for piping subprocess output
var out io.Writer = os.Stdout

var (
	// Modern, friendly colors - less intimidating than terminal colors
	headerStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#007ACC")).Bold(true)
	sectionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#0066CC")).Bold(true).MarginTop(1).MarginBottom(1)
	stepBulletStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Bold(true)
	stepTextStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	successBulletStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#28A745")).Bold(true)
	successTextStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#28A745"))
	warnBulletStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC107")).Bold(true)
	warnTextStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#856404"))
	infoStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#17A2B8")).Italic(true)
)

func headerLine(title string) string {
	return headerStyle.Render(fmt.Sprintf("╔═ %s ═╗", title))
}

func sectionLine(title string) string {
	border := "═"
	padding := strings.Repeat(border, len(title)+4)
	return fmt.Sprintf("%s\n║ %s ║\n%s",
		sectionStyle.Render("╔"+padding+"╗"),
		sectionStyle.Render("║  "+title+"  ║"),
		sectionStyle.Render("╚"+padding+"╝"))
}

func stepLine(msg string) string {
	return fmt.Sprintf("  %s %s", stepBulletStyle.Render("●"), stepTextStyle.Render(msg))
}

func successLine(msg string) string {
	return fmt.Sprintf("  %s %s", successBulletStyle.Render("✓"), successTextStyle.Render(msg))
}

func warnLine(msg string) string {
	return fmt.Sprintf("  %s %s", warnBulletStyle.Render("⚠"), warnTextStyle.Render(msg))
}

func infoLine(msg string) string {
	return fmt.Sprintf("  ℹ %s", infoStyle.Render(msg))
}

func dividerLine() string {
	return stepTextStyle.Render("────────────────────────────────────────")
}

// -------------------- Main TUI --------------------

// runMainTUI displays the main TUI with modpack selection and settings
func runMainTUI(modpacks []Modpack) (Modpack, bool) {
	for {
		// Create only menu items - no modpacks in main menu
		items := []list.Item{
			mainMenuItem{title: "Select Modpack", description: "Choose a modpack to launch"},
			mainMenuItem{title: "Settings", description: "Configure launcher settings"},
		}

		defaultIndex := 0 // Select "Select Modpack" by default

		model := newMainTUIModel(items, defaultIndex)
		prog := tea.NewProgram(model, tea.WithAltScreen())
		res, err := prog.Run()
		if err != nil {
			return Modpack{}, false
		}

		finalModel := res.(mainTUIModel)

		// Check if user selected settings
		if finalModel.selectedIndex == 1 {
			return Modpack{}, true // Settings chosen
		}

		// If user selected "Select Modpack" (index 0)
		if finalModel.selectedIndex == 0 {
			// Open modpack selection TUI
			selectedModpack, confirmed, err := runLauncherTUI(modpacks, Modpack{})
			if err != nil {
				return Modpack{}, false // Error occurred
			}

			// Check if user pressed back or cancelled
			if !confirmed || selectedModpack.ID == "" {
				// User pressed back or cancelled - show main menu again
				continue
			}

			// User selected a modpack
			return selectedModpack, false
		}

		// User quit or cancelled from main menu
		return Modpack{}, false
	}
}

type mainTUIModel struct {
	list         list.Model
	selectedIndex int
}

type mainMenuItem struct {
	title       string
	description string
}

func (i mainMenuItem) Title() string        { return i.title }
func (i mainMenuItem) Description() string { return i.description }
func (i mainMenuItem) FilterValue() string { return strings.ToLower(i.title + " " + i.description) }

type settingsMenuItem struct {
	title       string
	description string
	action      string
}

func (i settingsMenuItem) Title() string        { return i.title }
func (i settingsMenuItem) Description() string { return i.description }
func (i settingsMenuItem) FilterValue() string { return strings.ToLower(i.title + " " + i.description) }

type settingsTUIModel struct {
	list           list.Model
	selectedAction string
}

func newSettingsTUIModel(items []list.Item, defaultIndex int) settingsTUIModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bfc7ff"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#14141f")).Background(lipgloss.Color("#8be9fd")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2f9"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Settings"
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true).PaddingLeft(1)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.InfiniteScrolling = false
	l.Select(defaultIndex)

	return settingsTUIModel{
		list:           l,
		selectedAction: "",
	}
}

func (m settingsTUIModel) Init() tea.Cmd { return nil }

func (m settingsTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(settingsMenuItem); ok {
				m.selectedAction = item.action
			}
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
	return m, cmd
}

func (m settingsTUIModel) View() string {
	frame := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475a")).
		Render(m.list.View())

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8be9fd")).
		MarginTop(1).
		Render("↑/↓ navigate   •   Enter select   •   q quit")

	return frame + "\n" + help
}

func newMainTUIModel(items []list.Item, defaultIndex int) mainTUIModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bfc7ff"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#14141f")).Background(lipgloss.Color("#8be9fd")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2f9"))

	l := list.New(items, delegate, 0, 0)
	l.Title = fmt.Sprintf("%s Main Menu", launcherShortName)
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true).PaddingLeft(1)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.InfiniteScrolling = false
	l.Select(defaultIndex)

	return mainTUIModel{
		list:         l,
		selectedIndex: defaultIndex,
	}
}

func (m mainTUIModel) Init() tea.Cmd { return nil }

func (m mainTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		case "s":
			// Quick access to settings
			m.selectedIndex = 1
			m.list.Select(1)
			return m, tea.Quit
		case "1":
			m.selectedIndex = 0
			m.list.Select(0)
			return m, tea.Quit
		case "2":
			m.selectedIndex = 1
			m.list.Select(1)
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
	m.selectedIndex = m.list.Index()
	return m, cmd
}

func (m mainTUIModel) View() string {
	frame := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475a")).
		Render(m.list.View())

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8be9fd")).
		MarginTop(1).
		Render("↑/↓ navigate   •   Enter select   •   s settings   •   q quit")

	return frame + "\n" + help
}

// runSettingsMenu displays and handles the settings menu using TUI
func runSettingsMenu(root string) {
	items := []list.Item{
		settingsMenuItem{title: "Change Memory Settings", description: fmt.Sprintf("Current: %d GB", settings.MemoryMB/1024), action: "memory"},
		settingsMenuItem{title: "Reset to Auto", description: "Use half of system RAM", action: "auto"},
		settingsMenuItem{title: "Save and Exit", description: "Save settings and return to main menu", action: "save"},
		settingsMenuItem{title: "Exit without Saving", description: "Discard changes and return to main menu", action: "exit"},
	}

	defaultIndex := 0

	model := newSettingsTUIModel(items, defaultIndex)
	prog := tea.NewProgram(model, tea.WithAltScreen())
	res, err := prog.Run()
	if err != nil {
		return
	}

	finalModel := res.(settingsTUIModel)
	if finalModel.selectedAction == "" {
		return // User cancelled
	}

	// Handle the selected action
	switch finalModel.selectedAction {
	case "memory":
		changeMemorySettingsTUI()
		// Recursively show settings menu again
		runSettingsMenu(root)
	case "auto":
		resetToAutoSettings(root)
		fmt.Printf("\n%s", successLine("Memory reset to auto settings"))
		fmt.Printf("%s", infoLine("Returning to settings menu..."))
		time.Sleep(2 * time.Second)
		runSettingsMenu(root)
	case "save":
		if err := saveSettings(root); err != nil {
			fmt.Printf("\n%s", warnLine(fmt.Sprintf("Failed to save settings: %v", err)))
		} else {
			fmt.Printf("\n%s", successLine("Settings saved successfully!"))
		}
	case "exit":
		fmt.Printf("\n%s", infoLine("Settings not saved."))
	}
}

// changeMemorySettingsTUI provides a TUI for changing memory settings
func changeMemorySettingsTUI() {
	// Create items for memory options
	memItems := []list.Item{}

	// Add memory options from 2GB to 16GB in sensible increments
	memoryOptions := []int{2, 4, 6, 8, 10, 12, 14, 16}

	for _, gb := range memoryOptions {
		desc := fmt.Sprintf("Allocate %d GB to the modpack", gb)
		memItems = append(memItems, memoryMenuItem{
			title:       fmt.Sprintf("%d GB", gb),
			description: desc,
			memoryMB:    gb * 1024,
		})
	}

	// Add custom option
	memItems = append(memItems, memoryMenuItem{
		title:       "Custom",
		description: "Enter custom GB amount",
		memoryMB:    -1, // Special value for custom
	})

	model := newMemoryTUIModel(memItems, 0)
	prog := tea.NewProgram(model, tea.WithAltScreen())
	res, err := prog.Run()
	if err != nil {
		return
	}

	finalModel := res.(memoryTUIModel)
	if finalModel.selected {
		if finalModel.memoryMB == -1 {
			// Custom values - fall back to console input
			changeMemorySettingsConsole()
		} else {
			// Update settings with selected memory amount
			settings.MemoryMB = finalModel.memoryMB

			fmt.Printf("\n%s", successLine("Memory settings updated:"))
			fmt.Printf("  ■ Memory: %d GB\n", finalModel.memoryMB/1024)
			fmt.Printf("%s", infoLine("Press Enter to continue..."))
			fmt.Scanln() // Wait for user to acknowledge
		}
	}
}

// changeMemorySettingsConsole provides console input for custom memory values
func changeMemorySettingsConsole() {
	fmt.Printf("\n%s", headerLine("Custom Memory Configuration"))
	fmt.Printf("%s", infoLine("Enter memory allocation in GB (2-16 GB range)"))
	fmt.Printf("%s", dividerLine())

	var memoryGB int

	for {
		fmt.Printf("  %s Memory allocation (2-16 GB): ", stepBulletStyle.Render("►"))
		_, err := fmt.Scanf("%d", &memoryGB)
		if err != nil || memoryGB < 2 || memoryGB > 16 {
			fmt.Printf("  %s %s\n", warnBulletStyle.Render("⚠"), warnTextStyle.Render("Please enter a number between 2 and 16"))
			continue
		}
		break
	}

	// Convert to MB and update settings
	settings.MemoryMB = memoryGB * 1024

	fmt.Printf("%s", dividerLine())
	fmt.Printf("%s", successLine("Memory settings updated:"))
	fmt.Printf("  ■ Memory: %d GB\n", memoryGB)
	fmt.Printf("%s", infoLine("Press Enter to continue..."))
	fmt.Scanln() // Wait for user to acknowledge
}

type memoryMenuItem struct {
	title       string
	description string
	memoryMB    int
}

func (i memoryMenuItem) Title() string        { return i.title }
func (i memoryMenuItem) Description() string { return i.description }
func (i memoryMenuItem) FilterValue() string { return strings.ToLower(i.title + " " + i.description) }

type memoryTUIModel struct {
	list       list.Model
	selected   bool
	memoryMB   int
}

func newMemoryTUIModel(items []list.Item, defaultIndex int) memoryTUIModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bfc7ff"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#14141f")).Background(lipgloss.Color("#8be9fd")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2f9"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Select Memory Allocation"
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true).PaddingLeft(1)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.InfiniteScrolling = false
	l.Select(defaultIndex)

	return memoryTUIModel{
		list:     l,
		selected: false,
		memoryMB: 0,
	}
}

func (m memoryTUIModel) Init() tea.Cmd { return nil }

func (m memoryTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(memoryMenuItem); ok {
				m.selected = true
				m.memoryMB = item.memoryMB
			}
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
	return m, cmd
}

func (m memoryTUIModel) View() string {
	frame := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475a")).
		Render(m.list.View())

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8be9fd")).
		MarginTop(1).
		Render("↑/↓ navigate   •   Enter select   •   q quit")

	return frame + "\n" + help
}