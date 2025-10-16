// TheBoysLauncher.exe - portable Minecraft bootstrapper for Windows
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Fully portable: writes only beside the EXE (no AppData)
// - Downloads Prism Launcher (portable) - prefers MinGW w64 on amd64
// - Downloads Java 21 (Temurin JRE) dynamically (Adoptium API w/ GitHub fallback)
// - Downloads packwiz bootstrap dynamically (GitHub assets discovery)
// - Creates instance beside the EXE, writes instance.cfg (name/RAM/Java)
// - Runs packwiz from the *instance root* (detects MultiMC/Prism mode)
// - Console output + logs/latest.log (rotates to logs/previous.log)
// - Keeps console open on error (press Enter), disable with THEBOYS_NOPAUSE=1
// - Optional cache-bust for the modpack URL: set THEBOYS_CACHEBUST=1
// - Default launch opens an interactive TUI for modpack selection; use --cli for unattended console mode
// - Supports multiple modpacks via modpacks.json (falls back to built-in defaults)
//
// Build (set your version!):
//   go build -ldflags="-s -w -X main.version=v1.0.3" -o TheBoysLauncher.exe
//
// Usage for players: put TheBoysLauncher.exe in any writable folder and run it.
// Optional CLI: TheBoysLauncher.exe --cli [--modpack <id>] or --list-modpacks.

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	launcherName      = "TheBoys Launcher"
	launcherShortName = "TheBoys"
	launcherExeName   = "TheBoysLauncher.exe"

	// Self-update source (GitHub Releases of this EXE)
	UPDATE_OWNER      = "dilllxd"
	UPDATE_REPO       = "theboys-launcher"
	UPDATE_ASSET      = launcherExeName
	remoteModpacksURL = "https://raw.githubusercontent.com/dilllxd/theboys-launcher/refs/heads/main/modpacks.json"

	envCacheBust = "THEBOYS_CACHEBUST"
	envNoPause   = "THEBOYS_NOPAUSE"
)

type Modpack struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	PackURL      string `json:"packUrl"`
	InstanceName string `json:"instanceName"`
	Description  string `json:"description"`
	Default      bool   `json:"default,omitempty"`
}

var defaultModpackID string

// Optional: show MessageBox popups (false = log to console/file)
var interactive = false

// Populated at build time via -X main.version=vX.Y.Z
var version = "dev"

// global writer used by log/fail and for piping subprocess output
var out io.Writer = os.Stdout

var (
	sectionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")).Bold(true)
	stepBulletStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true)
	stepTextStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2"))
	successBulletStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
	successTextStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b"))
	warnBulletStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Bold(true)
	warnTextStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c"))
)

func sectionLine(title string) string {
	return sectionStyle.Render(strings.ToUpper(title))
}

func stepLine(msg string) string {
	return fmt.Sprintf("%s %s", stepBulletStyle.Render("→"), stepTextStyle.Render(msg))
}

func successLine(msg string) string {
	return fmt.Sprintf("%s %s", successBulletStyle.Render("✓"), successTextStyle.Render(msg))
}

func warnLine(msg string) string {
	return fmt.Sprintf("%s %s", warnBulletStyle.Render("!"), warnTextStyle.Render(msg))
}

func modpackLabel(mp Modpack) string {
	if name := strings.TrimSpace(mp.DisplayName); name != "" {
		return name
	}
	return mp.ID
}

func slugifyID(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "modpack"
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "modpack"
	}
	return result
}

func versionFileNameFor(mp Modpack) string { return "." + slugifyID(mp.ID) + "-version" }
func backupPrefixFor(mp Modpack) string    { return slugifyID(mp.ID) + "-backup-" }

// -------------------- MAIN --------------------

type launcherOptions struct {
	useCLI       bool
	modpackID    string
	listModpacks bool
}

func parseOptions() launcherOptions {
	opts := launcherOptions{}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.BoolVar(&opts.useCLI, "cli", false, "Run the launcher in plain console mode (skip the interactive TUI).")
	flag.StringVar(&opts.modpackID, "modpack", "", "Select the modpack to launch by ID.")
	flag.BoolVar(&opts.listModpacks, "list-modpacks", false, "List available modpacks and exit.")
	flag.Parse()

	return opts
}

func loadModpacks(root string) []Modpack {
	remote, err := fetchRemoteModpacks(remoteModpacksURL, 5*time.Second)
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
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

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

func main() {
	runtime.LockOSThread()

	opts := parseOptions()

	if runtime.GOOS != "windows" {
		msgBox("Windows only", launcherShortName, 0)
		return
	}

	exePath, _ := os.Executable()
	root := filepath.Dir(exePath)

	// 0) Logging: console + logs/latest.log (rotate previous.log)
	closeLog := setupLogging(root)
	defer closeLog()

	logf("%s", sectionLine(launcherName))
	logf("%s", stepLine(fmt.Sprintf("Session started %s", time.Now().Format(time.RFC1123))))
	logf("%s", stepLine(fmt.Sprintf("Version: %s", version)))

	modpacks := loadModpacks(root)
	if len(modpacks) == 0 {
		fail(errors.New("no modpacks configured"))
	}

	if opts.listModpacks {
		printModpackList(modpacks)
		return
	}

	selectedModpack, err := selectModpack(modpacks, opts.modpackID)
	if err != nil {
		fail(err)
	}

	// Set up signal handling for force-closing Prism and Minecraft on launcher exit
	var prismProcess *os.Process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logf("Launcher interrupted, force-closing Prism Launcher and Minecraft...")

		// Force close all Prism processes
		cmd := exec.Command("taskkill", "/F", "/IM", "PrismLauncher.exe")
		cmd.Run()

		// Force close any Java processes (likely Minecraft)
		javaCmd := exec.Command("taskkill", "/F", "/IM", "java.exe")
		javaCmd.Run()

		// Also close the specific Prism process we launched if we have it
		if prismProcess != nil && prismProcess.Pid > 0 {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", prismProcess.Pid))
			killCmd.Run()
			logf("Force-closed Prism process %d and related processes", prismProcess.Pid)
		}

		logf("All game processes force-closed")
		os.Exit(1)
	}()

	if opts.useCLI {
		logf("Launching (CLI) modpack: %s (%s)", modpackLabel(selectedModpack), selectedModpack.ID)
		runLauncherLogic(root, exePath, selectedModpack, &prismProcess)
		return
	}

	chosen, confirmed, err := runLauncherTUI(modpacks, selectedModpack)
	if err != nil {
		fail(err)
	}
	if !confirmed {
		logf("No modpack selected; exiting.")
		return
	}

	selectedModpack = chosen
	logf("Launching modpack: %s (%s)", modpackLabel(selectedModpack), selectedModpack.ID)
	runLauncherLogic(root, exePath, selectedModpack, &prismProcess)
}

// -------------------- Launcher Logic --------------------

func runLauncherLogic(root, exePath string, modpack Modpack, prismProcess **os.Process) {
	packName := modpackLabel(modpack)
	// 1) Self-update (best-effort; skips downgrades)
	if err := selfUpdate(root, exePath); err != nil {
		logf("Update check failed: %v", err)
	}

	// 2) Ensure prerequisites — organize directories cleanly
	prismDir := filepath.Join(root, "prism")
	utilDir := filepath.Join(root, "util")
	jreDir := filepath.Join(utilDir, "jre21")
	javaBin := filepath.Join(jreDir, "bin", "java.exe")
	bootstrapExe := filepath.Join(utilDir, "packwiz-installer-bootstrap.exe")
	bootstrapJar := filepath.Join(utilDir, "packwiz-installer-bootstrap.jar")

	// Create util directory for miscellaneous files
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		fail(fmt.Errorf("failed to create util directory: %w", err))
	}

	logf("%s", sectionLine("Preparing Environment"))

	logf("%s", stepLine("Ensuring Prism Launcher portable build"))
	prismDownloaded, err := ensurePrism(prismDir)
	if err != nil {
		fail(err)
	}
	if prismDownloaded {
		logf("%s", successLine("Prism Launcher downloaded"))
	} else {
		logf("%s", successLine("Prism Launcher ready"))
	}

	if !exists(javaBin) {
		logf("%s", stepLine("Installing Temurin JRE 21"))
		jreURL, err := fetchJRE21ZipURL()
		if err != nil {
			fail(fmt.Errorf("failed to resolve Java 21 download: %w", err))
		}
		if err := downloadAndUnzipTo(jreURL, jreDir); err != nil {
			fail(err)
		}
		_ = flattenOneLevel(jreDir)
		if !exists(javaBin) {
			fail(errors.New("Java 21 installation looks incomplete (bin/java.exe not found)"))
		}
		logf("%s", successLine("Java 21 installed"))
	} else {
		logf("%s", successLine("Java 21 already installed"))
	}

	logf("%s", stepLine("Ensuring packwiz bootstrap"))
	if !exists(bootstrapExe) && !exists(bootstrapJar) {
		pwURL, err := fetchPackwizBootstrapURL()
		if err != nil {
			fail(fmt.Errorf("failed to resolve packwiz bootstrap: %w", err))
		}
		target := bootstrapExe
		if strings.HasSuffix(strings.ToLower(pwURL), ".jar") {
			target = bootstrapJar
		}
		if err := downloadTo(pwURL, target, 0755); err != nil {
			fail(err)
		}
		logf("%s", successLine("Packwiz bootstrap installed"))
	} else {
		logf("%s", successLine("Packwiz bootstrap already installed"))
	}

	// 3) Create proper MultiMC/Prism instance first
	instDir := filepath.Join(prismDir, "instances", modpack.InstanceName)
	mcDir := filepath.Join(instDir, "minecraft") // Use minecraft, not .minecraft
	if err := os.MkdirAll(mcDir, 0755); err != nil {
		fail(err)
	}

	logf("%s", sectionLine("Instance Setup"))

	instanceConfigFile := filepath.Join(instDir, "instance.cfg")
	mmcPackFile := filepath.Join(instDir, "mmc-pack.json")

	needsInstanceCreation := !exists(instanceConfigFile) || !exists(mmcPackFile)
	if needsInstanceCreation {
		logf("%s", stepLine("Creating Prism instance structure with Forge 1.20.1"))
		if err := createMultiMCInstance(modpack, instDir, javaBin); err != nil {
			fail(fmt.Errorf("failed to create MultiMC instance: %w", err))
		}
		logf("%s", successLine("Instance structure ready"))
	} else {
		logf("%s", successLine("Instance structure already present"))
	}

	forgeJar := filepath.Join(mcDir, "libraries", "net", "minecraftforge", "forge", "1.20.1-47.4.0", "forge-1.20.1-47.4.0-universal.jar")
	forgeInstalled := exists(forgeJar) && exists(mmcPackFile)

	if !forgeInstalled {
		logf("%s", stepLine("Installing Forge 1.20.1"))
		if err := installForgeForInstance(instDir, javaBin); err != nil {
			fail(fmt.Errorf("failed to install Forge: %w", err))
		}
		logf("%s", successLine("Forge ready"))
	} else {
		logf("%s", successLine("Forge already installed"))
	}

	// 6) Check for modpack updates
	logf("%s", sectionLine("Modpack Sync"))
	logf("%s", stepLine("Checking for modpack updates"))
	updateAvailable, localVersion, remoteVersion, err := checkModpackUpdate(modpack, instDir)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to check modpack updates: %v", err)))
		updateAvailable = true
	}

	var action string
	var backupPath string

	if updateAvailable {
		if localVersion == "" {
			action = fmt.Sprintf("Installing %s version %s", packName, remoteVersion)
			logf("%s", stepLine(action))
		} else {
			action = fmt.Sprintf("Updating %s %s → %s", packName, localVersion, remoteVersion)
			logf("%s", stepLine(action))
			logf("%s", stepLine("Creating safety backup before update"))
			backupPath, err = createModpackBackup(modpack, mcDir)
			if err != nil {
				logf("%s", warnLine(fmt.Sprintf("Backup creation failed: %v", err)))
			}
		}
	} else {
		logf("%s", successLine("Modpack already up to date"))
		logf("%s", stepLine("Verifying installation with packwiz"))
	}

	packURL := modpack.PackURL
	if os.Getenv(envCacheBust) == "1" {
		sep := "?"
		if strings.Contains(packURL, "?") {
			sep = "&"
		}
		packURL = packURL + sep + "cb=" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Show progress indicator for packwiz operation
	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()

	go func() {
		for range progressTicker.C {
			if updateAvailable {
				logf("%s in progress... (this may take several minutes)", action)
			} else {
				logf("Verifying installation...")
			}
		}
	}()

	var cmd *exec.Cmd
	if exists(bootstrapExe) {
		cmd = exec.Command(bootstrapExe, "-g", packURL) // run from minecraft directory
	} else if exists(bootstrapJar) {
		cmd = exec.Command(javaBin, "-jar", bootstrapJar, "-g", packURL)
	} else {
		fail(errors.New("packwiz bootstrap not found after download"))
	}
	cmd.Dir = mcDir // critical: minecraft directory so packwiz installs mods in correct place
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	var buf bytes.Buffer
	mw := io.MultiWriter(out, &buf)
	cmd.Stdout, cmd.Stderr = mw, mw

	progressTicker.Stop() // Stop progress ticker before running packwiz
	err = cmd.Run()
	if err != nil {
		// Parse packwiz output for manual-download instructions
		items := parsePackwizManuals(buf.String())
		if len(items) > 0 {
			assistManualFromPackwiz(items)
			// Retry ONCE after user saves files, but create a new command to avoid "already started" error
			var retryCmd *exec.Cmd
			if exists(bootstrapExe) {
				retryCmd = exec.Command(bootstrapExe, "-g", packURL)
			} else if exists(bootstrapJar) {
				retryCmd = exec.Command(javaBin, "-jar", bootstrapJar, "-g", packURL)
			}
			if retryCmd != nil {
				retryCmd.Dir = mcDir // also run from minecraft directory
				retryCmd.Env = append(os.Environ(),
					"JAVA_HOME="+jreDir,
					"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
				)
				retryCmd.Stdout, retryCmd.Stderr = out, out
				err = retryCmd.Run()
			}
		}
	}

	if err != nil {
		// Update failed - attempt to restore from backup if we have one
		if backupPath != "" {
			logf("%s", warnLine("Packwiz update failed, attempting to restore from backup"))
			if restoreErr := restoreModpackBackup(modpack, backupPath, mcDir); restoreErr != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to restore backup: %v", restoreErr)))
			} else {
				logf("%s", successLine("Restored previous modpack state"))
			}
		}
		fail(fmt.Errorf("packwiz update failed: %w", err))
	}

	// Post-update verification and version saving
	if updateAvailable {
		logf("%s", stepLine("Verifying installation"))

		// Save the version that packwiz just installed
		if err := saveLocalVersion(modpack, instDir, remoteVersion); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to save local version: %v", err)))
		} else {
			logf("%s", successLine(fmt.Sprintf("%s now running version %s", packName, remoteVersion)))
		}
	} else {
		logf("%s", successLine(fmt.Sprintf("%s installation verification completed", packName)))
	}

	// 8) Launch selected instance directly
	logf("%s", sectionLine("Launching"))
	logf("%s", stepLine(fmt.Sprintf("Launching %s", packName)))

	prismExe := filepath.Join(prismDir, "PrismLauncher.exe")
	launch := exec.Command(prismExe, "--dir", ".", "--launch", modpack.InstanceName)
	launch.Dir = prismDir
	launch.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	launch.Stdout, launch.Stderr = out, out

	// Start the process and wait for it to complete (keeps console open)
	if err := launch.Start(); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to launch %s: %v", packName, err)))
		logf("%s", stepLine("Opening Prism Launcher UI instead"))
		launchFallback := exec.Command(prismExe, "--dir", ".")
		launchFallback.Dir = prismDir
		launchFallback.Env = append(os.Environ(),
			"JAVA_HOME="+jreDir,
			"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
		)
		launchFallback.Stdout, launchFallback.Stderr = out, out
		if err := launchFallback.Start(); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to open Prism Launcher UI: %v", err)))
			return
		}
		*prismProcess = launchFallback.Process
		logf("%s", successLine(fmt.Sprintf("Prism Launcher UI launched for %s (PID: %d)", packName, launchFallback.Process.Pid)))
		// Wait for the GUI process to complete
		launchFallback.Wait()
	} else {
		// Store the process reference for signal handling
		*prismProcess = launch.Process
		logf("%s", successLine(fmt.Sprintf("%s launched (PID: %d)", packName, launch.Process.Pid)))
		// Wait for the game process to complete
		launch.Wait()
	}

	logf("%s", successLine(fmt.Sprintf("Prism Launcher closed for %s", packName)))
}

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
	return finalModel.selected, finalModel.confirmed, nil
}

type tuiModel struct {
	list      list.Model
	selected  Modpack
	confirmed bool
}

func newTUIModel(modpacks []Modpack, defaultIndex int) tuiModel {
	items := make([]list.Item, len(modpacks))
	for i, mp := range modpacks {
		items[i] = modpackListItem{modpack: mp}
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
	l.Select(defaultIndex)

	return tuiModel{
		list:     l,
		selected: modpacks[defaultIndex],
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
			if item, ok := m.list.SelectedItem().(modpackListItem); ok {
				m.selected = item.modpack
			}
			m.confirmed = true
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
	if item, ok := m.list.SelectedItem().(modpackListItem); ok {
		m.selected = item.modpack
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
		Render("↑/↓ navigate   •   Enter launch   •   q quit")

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

// -------------------- Logging --------------------

func setupLogging(root string) func() {
	logDir := filepath.Join(root, "logs")
	_ = os.MkdirAll(logDir, 0755)
	latest := filepath.Join(logDir, "latest.log")
	prev := filepath.Join(logDir, "previous.log")

	// rotate previous
	if _, err := os.Stat(latest); err == nil {
		_ = os.Remove(prev)         // best-effort
		_ = os.Rename(latest, prev) // best-effort
	}

	f, err := os.OpenFile(latest, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// fall back to console only
		out = os.Stdout
		return func() {}
	}

	// Mirror to console + file
	out = io.MultiWriter(os.Stdout, f)

	return func() { _ = f.Close() }
}

func logf(format string, a ...any) {
	if interactive {
		msgBox(fmt.Sprintf(format, a...), launcherShortName, 0)
	} else {
		fmt.Fprintf(out, format+"\n", a...)
	}
}

func pauseIfWanted() {
	if os.Getenv(envNoPause) == "1" {
		return
	}
	fmt.Fprint(out, "\nPress Enter to exit…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func fail(err error) {
	if interactive {
		msgBox("Error: "+err.Error(), launcherShortName, 0)
	} else {
		fmt.Fprintf(out, "Error: %v\n", err)
	}
	pauseIfWanted()
	os.Exit(1)
}

// -------------------- Self-update (no downgrades) --------------------

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func selfUpdate(root, exePath string) error {
	tag, assetURL, err := fetchLatestAsset(UPDATE_OWNER, UPDATE_REPO, UPDATE_ASSET)
	if err != nil || tag == "" || assetURL == "" {
		return err
	}

	localTag := normalizeTag(version)
	remoteTag := normalizeTag(tag)

	switch compareSemver(localTag, remoteTag) {
	case 0:
		logf("%s up to date (%s).", launcherShortName, version)
		return nil
	case 1:
		// local > remote → don't downgrade
		logf("Local launcher (%s) is newer than latest release (%s). Skipping update.", version, tag)
		return nil
	case -1:
		// remote > local → proceed
	}

	logf("New %s available: %s (current %s). Updating…", launcherShortName, tag, version)

	tmpNew := exePath + ".new"
	if err := downloadTo(assetURL, tmpNew, 0755); err != nil {
		return err
	}

	// Windows can't replace a running EXE. Use a tiny batch to swap.
	updater := filepath.Join(root, "update_launcher.bat")
	up := fmt.Sprintf(`@echo off
set EXE="%s"
set NEW="%s"
:loop
tasklist /FI "IMAGENAME eq %s" | find /I "%s" >nul && (timeout /t 1 >nul & goto loop)
move /Y %s %s >nul
start "" %s
del "%%~f0"
`, filepath.Base(exePath), filepath.Base(tmpNew),
		filepath.Base(exePath), filepath.Base(exePath),
		filepath.Base(tmpNew), filepath.Base(exePath),
		filepath.Base(exePath))
	if err := os.WriteFile(updater, []byte(up), 0644); err != nil {
		return err
	}
	cmd := exec.Command("cmd", "/c", "start", "/min", "", filepath.Base(updater))
	cmd.Dir = root
	_ = cmd.Start()
	os.Exit(0)
	return nil
}

func fetchLatestAsset(owner, repo, wantName string) (tag, url string, err error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", "TheBoys-Updater/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", "", err
	}
	for _, a := range r.Assets {
		if a.Name == wantName {
			return r.TagName, a.BrowserDownloadURL, nil
		}
	}
	return r.TagName, "", errors.New("asset not found in latest release")
}

// --- Version helpers (semver-ish) ---

func normalizeTag(t string) string {
	// Strip leading "v" or "V" and whitespace
	t = strings.TrimSpace(t)
	if len(t) > 0 && (t[0] == 'v' || t[0] == 'V') {
		t = t[1:]
	}
	// Drop any pre-release/build suffix ("-rc.1", "+meta")
	if i := strings.IndexAny(t, "-+"); i >= 0 {
		t = t[:i]
	}
	return t
}

func parseSemverInts(t string) (major, minor, patch int) {
	parts := strings.Split(t, ".")
	get := func(i int) int {
		if i >= len(parts) || parts[i] == "" {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	return get(0), get(1), get(2)
}

// returns -1 if a<b, 0 if equal, +1 if a>b
func compareSemver(a, b string) int {
	amaj, amin, apat := parseSemverInts(a)
	bmaj, bmin, bpat := parseSemverInts(b)
	if amaj != bmaj {
		if amaj < bmaj {
			return -1
		}
		return 1
	}
	if amin != bmin {
		if amin < bmin {
			return -1
		}
		return 1
	}
	if apat != bpat {
		if apat < bpat {
			return -1
		}
		return 1
	}
	return 0
}

// -------------------- Downloads / Unzip --------------------

type progressWriter struct {
	total      int64
	downloaded int64
	filename   string
	startTime  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	err := error(nil)
	pw.downloaded += int64(n)

	// Update progress every 1MB or every second
	if pw.downloaded%1048576 == 0 || time.Since(pw.startTime) > time.Second {
		pw.updateProgress()
		pw.startTime = time.Now()
	}

	return n, err
}

func (pw *progressWriter) updateProgress() {
	if pw.total > 0 {
		percent := float64(pw.downloaded) / float64(pw.total) * 100

		// Calculate download speed
		elapsed := time.Since(pw.startTime).Seconds()
		if elapsed > 0 {
			speedMBps := (float64(pw.downloaded) / (1024 * 1024)) / elapsed
			fmt.Fprintf(out, "\rDownloading %s (%.1f MB/s, %d%%)", pw.filename, speedMBps, int(percent))
		} else {
			fmt.Fprintf(out, "\rDownloading %s (%d%%)", pw.filename, int(percent))
		}
	}
}

func downloadTo(url, path string, mode os.FileMode) error {
	b, err := downloadWithProgress(url)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, mode)
}

func downloadAndUnzipTo(url, dest string) error {
	b, err := download(url)
	if err != nil {
		return err
	}
	return unzipBytesTo(b, dest)
}

func download(url string) ([]byte, error) {
	return downloadWithProgress(url)
}

func downloadWithProgress(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "TheBoys/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	// Get file info for progress
	contentLength := resp.ContentLength
	filename := filepath.Base(url)

	// Create progress writer
	pw := &progressWriter{
		total:     contentLength,
		filename:  filename,
		startTime: time.Now(),
	}

	// If we don't know the content length, show indefinite progress
	if contentLength <= 0 {
		fmt.Fprintf(out, "Downloading %s...", filename)
		return io.ReadAll(resp.Body)
	}

	// Read with progress tracking
	body, err := io.ReadAll(io.TeeReader(resp.Body, pw))
	if err != nil {
		return nil, err
	}

	// Show completion
	fmt.Fprintf(out, "\nDownloaded %s (%.1f MB)\n", filename, float64(contentLength)/(1024*1024))

	return body, nil
}

func unzipBytesTo(b []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		p := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outf, err := os.Create(p)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(outf, rc)
		outf.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// -------------------- Prism + Instance --------------------

func ensurePrism(dir string) (bool, error) {
	if exists(filepath.Join(dir, "PrismLauncher.exe")) {
		return false, nil
	}
	url, err := fetchLatestPrismPortableURL()
	if err != nil {
		return false, err
	}
	logf("%s", stepLine(fmt.Sprintf("Downloading Prism portable build: %s", url)))
	if err := downloadAndUnzipTo(url, dir); err != nil {
		return false, err
	}
	// Force portable mode
	cfg := filepath.Join(dir, "prismlauncher.cfg")
	_ = os.WriteFile(cfg, []byte("Portable=true\n"), 0644)
	return true, nil
}

type prismRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Prefer MinGW w64 portable on amd64; fall back to MSVC portable.
// On arm64, use MSVC arm64 portable.
func fetchLatestPrismPortableURL() (string, error) {
	api := "https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/latest"
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", "TheBoys-Prism/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var rel prismRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}

	// Build priority patterns by arch
	type pat struct{ re *regexp.Regexp }
	var patterns []pat

	if runtime.GOARCH == "amd64" {
		// 1) MinGW w64 portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MinGW-w64-Portable-.*\.zip$`)})
		// 2) MSVC portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MSVC-Portable-.*\.zip$`)})
	} else if runtime.GOARCH == "arm64" {
		// MSVC arm64 portable zip
		patterns = append(patterns, pat{regexp.MustCompile(`(?i)Windows-MSVC-arm64-Portable-.*\.zip$`)})
	}

	// Fallbacks for unexpected naming: generic portable zips
	patterns = append(patterns,
		pat{regexp.MustCompile(`(?i)Windows-.*Portable-.*\.zip$`)},
		pat{regexp.MustCompile(`(?i)Windows-.*\.zip$`)},
	)

	// Search in priority order
	for _, p := range patterns {
		for _, a := range rel.Assets {
			if p.re.MatchString(a.Name) {
				return a.URL, nil
			}
		}
	}
	return "", errors.New("no suitable Prism portable asset found in latest release")
}

// -------------------- Java 21 URL discovery --------------------

// Prefer Adoptium API (stable), fall back to GitHub release asset.
// We want: OS=windows, arch=x64, image_type=jre, vm=hotspot, latest for 21.
func fetchJRE21ZipURL() (string, error) {
	// 1) Adoptium API (v3)
	adoptium := "https://api.adoptium.net/v3/assets/latest/21/hotspot?architecture=x64&image_type=jre&os=windows"
	req, _ := http.NewRequest("GET", adoptium, nil)
	req.Header.Set("User-Agent", "TheBoys-Adoptium/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var payload []struct {
			Binaries []struct {
				Package struct {
					Link string `json:"link"`
				} `json:"package"`
			} `json:"binaries"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			for _, v := range payload {
				for _, b := range v.Binaries {
					if b.Package.Link != "" && strings.HasSuffix(strings.ToLower(b.Package.Link), ".zip") {
						return b.Package.Link, nil
					}
				}
			}
		}
	} else if resp != nil {
		resp.Body.Close()
	}

	// 2) Fallback to GitHub Releases: adoptium/temurin21-binaries
	api := "https://api.github.com/repos/adoptium/temurin21-binaries/releases/latest"
	req2, _ := http.NewRequest("GET", api, nil)
	req2.Header.Set("User-Agent", "TheBoys-Adoptium/1.0")
	req2.Header.Set("Cache-Control", "no-cache")
	req2.Header.Set("Pragma", "no-cache")
	resp2, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return "", fmt.Errorf("adoptium api and github fallback failed: %v", err2)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("github adoptium status %d", resp2.StatusCode)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp2.Body).Decode(&rel); err != nil {
		return "", err
	}
	// Example pattern: OpenJDK21U-jre_x64_windows_hotspot_*.zip
	re := regexp.MustCompile(`(?i)^OpenJDK21U-jre_x64_windows_hotspot_.*\.zip$`)
	for _, a := range rel.Assets {
		if re.MatchString(a.Name) {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", errors.New("no suitable Java 21 JRE zip found")
}

// -------------------- packwiz bootstrap URL discovery --------------------

func fetchPackwizBootstrapURL() (string, error) {
	api := "https://api.github.com/repos/packwiz/packwiz-installer-bootstrap/releases/latest"
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", "TheBoys-Packwiz/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	// Prefer Windows exe if present, else accept jar (cross-platform)
	prefs := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^packwiz-installer-bootstrap-.*windows.*amd64.*\.exe$`),
		regexp.MustCompile(`(?i)^packwiz-installer-bootstrap-.*windows.*\.exe$`),
		regexp.MustCompile(`(?i)^packwiz-installer-bootstrap\.jar$`),
	}
	for _, re := range prefs {
		for _, a := range rel.Assets {
			if re.MatchString(a.Name) {
				return a.BrowserDownloadURL, nil
			}
		}
	}
	return "", errors.New("no suitable packwiz bootstrap asset found")
}

// -------------------- MultiMC Instance Creation --------------------

func createMultiMCInstance(modpack Modpack, instDir, javaExe string) error {
	minMB, maxMB := autoRAM()

	// Create instance.cfg
	instanceLines := []string{
		"InstanceType=OneSix", // Use OneSix not Minecraft
		"name=" + modpack.InstanceName,
		"iconKey=default",
		"OverrideMemory=true",
		fmt.Sprintf("MinMemAlloc=%d", minMB),
		fmt.Sprintf("MaxMemAlloc=%d", maxMB),
		"OverrideJava=true",
		"JavaPath=" + javaExe,
		"Notes=Managed by " + launcherName,
	}

	// Create mmc-pack.json with proper components including LWJGL 3
	mmcPack := map[string]interface{}{
		"formatVersion": 1,
		"components": []interface{}{
			map[string]interface{}{
				"cachedName":     "LWJGL 3",
				"cachedVersion":  "3.3.1",
				"cachedVolatile": true,
				"dependencyOnly": true,
				"uid":            "org.lwjgl3",
				"version":        "3.3.1",
			},
			map[string]interface{}{
				"cachedName":    "Minecraft",
				"cachedVersion": "1.20.1",
				"cachedRequires": []interface{}{
					map[string]interface{}{
						"suggests": "3.3.1",
						"uid":      "org.lwjgl3",
					},
				},
				"important": true,
				"uid":       "net.minecraft",
				"version":   "1.20.1",
			},
			map[string]interface{}{
				"cachedName":    "Forge",
				"cachedVersion": "47.4.0", // Update to match working instance
				"cachedRequires": []interface{}{
					map[string]interface{}{
						"equals": "1.20.1",
						"uid":    "net.minecraft",
					},
				},
				"uid":     "net.minecraftforge",
				"version": "47.4.0",
			},
		},
	}

	// Create pack.json for MultiMC format with matching components
	pack := map[string]interface{}{
		"formatVersion": 3,
		"components": []interface{}{
			map[string]interface{}{
				"cachedName":     "LWJGL 3",
				"cachedVersion":  "3.3.1",
				"cachedVolatile": true,
				"dependencyOnly": true,
				"uid":            "org.lwjgl3",
				"version":        "3.3.1",
			},
			map[string]interface{}{
				"cachedName":    "Minecraft",
				"cachedVersion": "1.20.1",
				"cachedRequires": []interface{}{
					map[string]interface{}{
						"suggests": "3.3.1",
						"uid":      "org.lwjgl3",
					},
				},
				"important": true,
				"uid":       "net.minecraft",
				"version":   "1.20.1",
			},
			map[string]interface{}{
				"cachedName":    "Forge",
				"cachedVersion": "47.4.0", // Update to match working instance
				"cachedRequires": []interface{}{
					map[string]interface{}{
						"equals": "1.20.1",
						"uid":    "net.minecraft",
					},
				},
				"uid":     "net.minecraftforge",
				"version": "47.4.0",
			},
		},
	}

	// Write all the required MultiMC files (only if they don't exist)
	instanceCfgPath := filepath.Join(instDir, "instance.cfg")
	mmcPackPath := filepath.Join(instDir, "mmc-pack.json")
	packJsonPath := filepath.Join(instDir, "pack.json")

	if !exists(instanceCfgPath) {
		if err := os.WriteFile(instanceCfgPath, []byte(strings.Join(instanceLines, "\n")+"\n"), 0644); err != nil {
			return err
		}
	}

	if !exists(mmcPackPath) {
		mmcPackData, err := json.MarshalIndent(mmcPack, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(mmcPackPath, mmcPackData, 0644); err != nil {
			return err
		}
	}

	if !exists(packJsonPath) {
		packData, err := json.MarshalIndent(pack, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(packJsonPath, packData, 0644); err != nil {
			return err
		}
	}

	return nil
}

func installForgeForInstance(instDir, javaBin string) error {
	mcDir := filepath.Join(instDir, "minecraft") // Use minecraft not .minecraft

	// Check for Forge installation in MultiMC/Prism instance structure
	forgeJar := filepath.Join(mcDir, "libraries", "net", "minecraftforge", "forge", "1.20.1-47.4.0", "forge-1.20.1-47.4.0-universal.jar")
	mmcPackFile := filepath.Join(instDir, "mmc-pack.json")

	// Check if Forge is already installed
	if exists(forgeJar) && exists(mmcPackFile) {
		logf("Forge already completely installed in instance")
		return nil
	}

	// Download Forge installer
	forgeURL := "https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.0/forge-1.20.1-47.4.0-installer.jar"
	utilDir := filepath.Join(filepath.Dir(instDir), "..", "..", "util")
	installerPath := filepath.Join(utilDir, "forge-installer.jar")

	logf("Downloading Forge installer...")
	if err := downloadTo(forgeURL, installerPath, 0644); err != nil {
		return fmt.Errorf("failed to download Forge installer: %w", err)
	}

	// Run Forge installer
	logf("Installing Forge...")
	fmt.Fprintf(out, "Running Forge installer... (this may take a few minutes)\n")

	cmd := exec.Command(javaBin, "-jar", installerPath, "--installClient", "--installServer")
	cmd.Dir = mcDir
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+filepath.Dir(filepath.Dir(javaBin)),
		"PATH="+filepath.Dir(filepath.Dir(javaBin))+";"+os.Getenv("PATH"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Forge installer failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(out, "✓ Forge installation completed successfully\n")

	// Clean up installer
	_ = os.Remove(installerPath)

	return nil
}

// -------------------- Modpack Version Checking --------------------

// PackConfig represents the structure of a pack.toml file
type PackConfig struct {
	Version string `toml:"version"`
}

// fetchRemotePackVersion fetches the remote pack.toml and extracts the version
func fetchRemotePackVersion(packURL string) (string, error) {
	req, err := http.NewRequest("GET", packURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "TheBoys/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, packURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var packConfig PackConfig
	if err := toml.Unmarshal(body, &packConfig); err != nil {
		return "", fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	if packConfig.Version == "" {
		return "", errors.New("no version found in pack.toml")
	}

	return packConfig.Version, nil
}

// getLocalPackVersion gets the version from our local version tracking file
func getLocalPackVersion(mp Modpack, instDir string) (string, error) {
	versionFilePath := filepath.Join(instDir, versionFileNameFor(mp))

	// Check if our version file exists
	if !exists(versionFilePath) {
		logf("Debug: %s version file not found at %s", modpackLabel(mp), versionFilePath)
		return "", nil // No version file exists
	}

	body, err := os.ReadFile(versionFilePath)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(body))
	logf("Debug: Found local %s version %s at %s", modpackLabel(mp), version, versionFilePath)
	return version, nil
}

// saveLocalVersion saves the current modpack version to our tracking file
func saveLocalVersion(mp Modpack, instDir, version string) error {
	versionFilePath := filepath.Join(instDir, versionFileNameFor(mp))

	if err := os.WriteFile(versionFilePath, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to save local version: %w", err)
	}

	logf("Debug: Saved local %s version %s to %s", modpackLabel(mp), version, versionFilePath)
	return nil
}

// checkModpackUpdate checks if there's a modpack update available
func checkModpackUpdate(modpack Modpack, instDir string) (bool, string, string, error) {
	remoteVersion, err := fetchRemotePackVersion(modpack.PackURL)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch remote modpack version: %w", err)
	}

	localVersion, err := getLocalPackVersion(modpack, instDir)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get local modpack version: %w", err)
	}

	packName := modpackLabel(modpack)

	// If no local version exists, we need to install
	if localVersion == "" {
		logf("No local %s found, will install version %s", packName, remoteVersion)
		return true, "", remoteVersion, nil
	}

	// Compare versions
	if localVersion != remoteVersion {
		logf("%s update available: %s → %s", packName, localVersion, remoteVersion)
		return true, localVersion, remoteVersion, nil
	}

	logf("%s is up to date (%s)", packName, localVersion)
	return false, localVersion, remoteVersion, nil
}

// -------------------- Modpack Backup & Restore --------------------

// createModpackBackup creates a backup of the current modpack before updating
func createModpackBackup(mp Modpack, mcDir string) (string, error) {
	packName := modpackLabel(mp)
	// Clean up old backups (keep only the 3 most recent)
	if err := cleanupOldBackups(mp, mcDir, 3); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to clean old backups: %v", err)))
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupName := backupPrefixFor(mp) + timestamp
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(mcDir)))
	backupPath := filepath.Join(rootDir, "util", "backups", backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Directories to backup
	dirsToBackup := []string{"mods", "config", "resourcepacks", "shaderpacks"}

	logf("%s", stepLine(fmt.Sprintf("Creating backup %s for %s", backupName, packName)))

	var backedUpItems []string
	for _, dir := range dirsToBackup {
		srcPath := filepath.Join(mcDir, dir)
		dstPath := filepath.Join(backupPath, dir)

		if exists(srcPath) {
			if err := copyDir(srcPath, dstPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to backup %s: %v", dir, err)))
			} else {
				backedUpItems = append(backedUpItems, dir)
			}
		}
	}

	// Backup our version file if it exists
	versionFile := versionFileNameFor(mp)
	versionFileSrc := filepath.Join(filepath.Dir(mcDir), versionFile)
	versionFileDst := filepath.Join(backupPath, versionFile)
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to backup version file: %v", err)))
		} else {
			backedUpItems = append(backedUpItems, versionFile)
		}
	}

	if len(backedUpItems) == 0 {
		logf("%s", warnLine(fmt.Sprintf("No files found to backup for %s", packName)))
		return "", nil
	}

	logf("%s", successLine(fmt.Sprintf("Backup created for %s: %s (items: %s)", packName, backupName, strings.Join(backedUpItems, ", "))))
	return backupPath, nil
}

// restoreModpackBackup restores from a backup if the update fails
func restoreModpackBackup(mp Modpack, backupPath, mcDir string) error {
	if backupPath == "" || !exists(backupPath) {
		return errors.New("no backup available to restore")
	}

	packName := modpackLabel(mp)
	logf("%s", stepLine(fmt.Sprintf("Restoring %s from backup", packName)))

	// Remove current modpack directories
	dirsToRemove := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	for _, dir := range dirsToRemove {
		dirPath := filepath.Join(mcDir, dir)
		if exists(dirPath) {
			if err := os.RemoveAll(dirPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to remove %s during restore: %v", dir, err)))
			}
		}
	}

	// Restore backup
	dirsToRestore := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	var restoredItems []string

	for _, dir := range dirsToRestore {
		srcPath := filepath.Join(backupPath, dir)
		dstPath := filepath.Join(mcDir, dir)

		if exists(srcPath) {
			if err := copyDir(srcPath, dstPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to restore %s: %v", dir, err)))
			} else {
				restoredItems = append(restoredItems, dir)
			}
		}
	}

	// Restore our version file
	versionFile := versionFileNameFor(mp)
	versionFileSrc := filepath.Join(backupPath, versionFile)
	versionFileDst := filepath.Join(filepath.Dir(mcDir), versionFile)
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to restore version file: %v", err)))
		} else {
			restoredItems = append(restoredItems, versionFile)
		}
	}

	if len(restoredItems) == 0 {
		return errors.New("nothing to restore from backup")
	}

	logf("%s", successLine(fmt.Sprintf("Restored %s: %s", packName, strings.Join(restoredItems, ", "))))
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// cleanupOldBackups removes old backups, keeping only the most recent ones
func cleanupOldBackups(mp Modpack, mcDir string, keepCount int) error {
	packName := modpackLabel(mp)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(mcDir)))
	backupsDir := filepath.Join(rootDir, "util", "backups")
	if !exists(backupsDir) {
		return nil
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return err
	}

	var backupDirs []os.DirEntry
	prefix := backupPrefixFor(mp)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backupDirs = append(backupDirs, entry)
		}
	}

	// Sort by name (which includes timestamp, so this sorts by time)
	if len(backupDirs) <= keepCount {
		return nil
	}

	// Remove oldest backups (keep the most recent ones)
	sort.Slice(backupDirs, func(i, j int) bool {
		return backupDirs[i].Name() > backupDirs[j].Name() // newer names first
	})

	toRemove := backupDirs[keepCount:]
	for _, entry := range toRemove {
		removePath := filepath.Join(backupsDir, entry.Name())
		if err := os.RemoveAll(removePath); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to remove old %s backup %s: %v", packName, entry.Name(), err)))
		} else {
			logf("%s", successLine(fmt.Sprintf("Removed old %s backup: %s", packName, entry.Name())))
		}
	}

	return nil
}

// -------------------- Helpers --------------------

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func flattenOneLevel(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return nil
	}
	top := filepath.Join(dir, entries[0].Name())
	items, err := os.ReadDir(top)
	if err != nil {
		return err
	}
	for _, it := range items {
		src := filepath.Join(top, it.Name())
		dst := filepath.Join(dir, it.Name())
		_ = os.Rename(src, dst)
	}
	_ = os.Remove(top)
	return nil
}

func msgBox(text, title string, flags uintptr) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	_, _, _ = proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		flags)
}

// Auto RAM: ~50% of total, min 4096 MB, max 16384 MB for modern systems
func autoRAM() (minMB, maxMB int) {
	total := totalRAMMB()
	if total <= 0 {
		total = 65536 // fallback if detection fails (assume 64GB)
	}

	// Use 25-30% of total RAM for modpacks, but cap at 16GB
	maxMB = int(float64(total) * 0.30)
	if maxMB > 16384 {
		maxMB = 16384
	}

	// Floor at 4096 for modern modpacks
	if maxMB < 4096 {
		maxMB = 4096
	}

	// Min mem = half of max, but not below 4096
	minMB = max(4096, maxMB/2)
	return
}

// totalRAMMB returns total system memory in MB
func totalRAMMB() int {
	type mstat struct {
		dwLen uint32
		load  uint32
		total uint64
		avail uint64
		a     uint64
		b     uint64
		c     uint64
		d     uint64
	}
	k32 := syscall.NewLazyDLL("kernel32.dll")
	proc := k32.NewProc("GlobalMemoryStatusEx")
	var s mstat
	s.dwLen = uint32(unsafe.Sizeof(s))
	r1, _, _ := proc.Call(uintptr(unsafe.Pointer(&s)))
	if r1 == 0 {
		return 0
	}
	return int(s.total / (1024 * 1024))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// -------------------- CurseForge Direct Download --------------------

// downloadFromCurseForge attempts to download JAR files directly from CurseForge URLs
func downloadFromCurseForge(url, destPath string) error {
	// Handle CurseForge URLs with retry logic
	if strings.Contains(url, "curseforge.com") {
		return downloadCurseForgeFileWithRetry(url, destPath, 3)
	}

	// For non-CurseForge URLs, fall back to regular download
	return downloadTo(url, destPath, 0644)
}

// downloadCurseForgeFileWithRetry attempts to download from CurseForge with multiple retry attempts
func downloadCurseForgeFileWithRetry(pageURL, destPath string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logf("  Attempt %d/%d...", attempt, maxRetries)

		// Remove any existing partial download
		if exists(destPath) {
			os.Remove(destPath)
		}

		err := downloadCurseForgeFile(pageURL, destPath)
		if err == nil {
			return nil // Success!
		}

		lastErr = err
		logf("  Failed: %v", err)

		// Don't wait on the last attempt
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			logf("  Retrying in %d seconds...", waitTime/time.Second)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadCurseForgeFile converts CurseForge file URLs to direct download URLs and downloads the file
func downloadCurseForgeFile(pageURL, destPath string) error {
	// Extract project info from the URL
	game, category, projectSlug, fileID, err := parseCurseForgeFileURL(pageURL)
	if err != nil {
		return fmt.Errorf("failed to parse CurseForge URL: %w", err)
	}

	// Method 1: Try the simple download URL format first
	downloadURL := fmt.Sprintf("https://www.curseforge.com/%s/%s/%s/download/%s", game, category, projectSlug, fileID)
	if err := tryDirectDownload(downloadURL, destPath); err == nil {
		return nil
	}

	// Method 2: Scrape project ID from the file page itself and use API
	projectID, err := getProjectIDFromFilePage(pageURL)
	if err == nil && projectID != "" {
		apiURL := fmt.Sprintf("https://www.curseforge.com/api/v1/mods/%s/files/%s/download", projectID, fileID)
		if err := tryDirectDownload(apiURL, destPath); err == nil {
			return nil
		}
	}

	// Method 3: Fallback to parsing the file page for download links
	return downloadCurseForgeFromPage(pageURL, destPath)
}

// getProjectIDFromFilePage scrapes the project ID from the CurseForge file page itself
func getProjectIDFromFilePage(filePageURL string) (string, error) {
	req, err := http.NewRequest("GET", filePageURL, nil)
	if err != nil {
		return "", err
	}

	// Add realistic browser headers to avoid 403 errors
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body

	// Check if content is gzipped and decompress if needed
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	html := string(body)

	// Look for the specific pattern: <dt>Project ID</dt><dd><div class="project-id-container"><span class="project-id">433760</span>
	projectIDPattern := regexp.MustCompile(`<dt>Project ID</dt>\s*<dd>\s*<div[^>]*class="project-id-container"[^>]*>\s*<span[^>]*class="project-id"[^>]*>(\d+)</span>`)
	matches := projectIDPattern.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// More flexible patterns - try many different ways the project ID might appear
	patterns := []string{
		// Various HTML patterns for project ID
		`<span[^>]*class="project-id"[^>]*>(\d+)</span>`,
		`<dt>Project ID</dt>\s*<dd>(\d+)</dd>`,
		`<dt>Project ID</dt>\s*<dd>\s*(\d+)\s*</dd>`,
		`<div[^>]*project-id[^>]*>(\d+)</div>`,
		`data-project-id="(\d+)"`,
		`project-id="(\d+)"`,

		// JSON patterns in embedded data
		`"project_id":\s*(\d+)`,
		`"projectId":\s*(\d+)`,
		`"project":\s*\{[^}]*"id":\s*(\d+)`,
		`"project":\{[^}]*"id":(\d+)`,
		`globalThis\.project[^=]*=\s*\{[^}]*"id":\s*(\d+)`,
		`window\.project[^=]*=\s*\{[^}]*"id":\s*(\d+)`,
		`"eagerProject":\{[^}]*"id":\s*(\d+)`,
		`"projectData":\{[^}]*"id":\s*(\d+)`,
		`"data":[^}]*"id":\s*(\d+)`,

		// More general numeric patterns in project context
		`"file_id":\d+.*?"project_id":\s*(\d+)`,
		`"fileId":\d+.*?"projectId":\s*(\d+)`,
		`"slug":"[^"]*".*?"id":(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			projectID := matches[1]
			if len(projectID) > 0 && len(projectID) < 10 {
				return projectID, nil
			}
		}
	}

	return "", fmt.Errorf("project ID not found in file page")
}

// parseCurseForgeFileURL extracts game, category, project slug, and file ID from a CurseForge file URL
func parseCurseForgeFileURL(url string) (game, category, projectSlug, fileID string, err error) {
	// Pattern: https://www.curseforge.com/minecraft/mc-mods/mod-name/files/1234567
	// Pattern: https://www.curseforge.com/minecraft/texture-packs/pack-name/files/1234567
	re := regexp.MustCompile(`https://www\.curseforge\.com/([^/]+)/([^/]+)/([^/]+)/files/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 5 {
		return "", "", "", "", fmt.Errorf("invalid CurseForge URL format")
	}

	return matches[1], matches[2], matches[3], matches[4], nil
}

// tryDirectDownload attempts to download from a URL that should be a direct download link
func tryDirectDownload(downloadURL, destPath string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "TheBoys/1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirects for direct downloads
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Check if we got a file (check content type and disposition)
	contentType := resp.Header.Get("Content-Type")
	contentDisp := resp.Header.Get("Content-Disposition")

	isFile := strings.Contains(contentType, "application/java-archive") ||
		strings.Contains(contentType, "application/zip") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentDisp, ".jar") ||
		strings.Contains(contentDisp, ".zip")

	if !isFile {
		return fmt.Errorf("response doesn't appear to be a file (Content-Type: %s)", contentType)
	}

	// Download the file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return os.WriteFile(destPath, body, 0644)
}

// downloadCurseForgeFromPage fallback method that parses the page HTML
func downloadCurseForgeFromPage(pageURL, destPath string) error {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "TheBoys/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch CurseForge page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body

	// Check if content is gzipped and decompress if needed
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read page content: %w", err)
	}

	// Look for download link patterns in the HTML
	downloadURL := extractCurseForgeDownloadLink(string(body))
	if downloadURL != "" {
		return downloadTo(downloadURL, destPath, 0644)
	}

	return fmt.Errorf("could not extract direct download link from CurseForge page")
}

// extractCurseForgeDownloadLink attempts to find the direct download URL from CurseForge page HTML
func extractCurseForgeDownloadLink(html string) string {
	// Look for various CurseForge download patterns
	patterns := []string{
		`"downloadUrl":"([^"]+\.jar)"`,
		`"url":"([^"]+\.jar)"`,
		`href="([^"]+\.jar)"`,
		`data-download="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			// Unescape JSON-encoded URL if needed
			url := strings.ReplaceAll(matches[1], "\\u0026", "&")
			url = strings.ReplaceAll(url, "\\", "")
			return url
		}
	}

	return ""
}

// -------------------- Packwiz manual download parsing & assist --------------------

type manualItem struct {
	Name string // optional; may be empty
	URL  string
	Path string // absolute path from packwiz message
}

// Parse lines like:
// "Mod Name: ... Please go to https://... and save this file to C:\...\mods\file.jar"
// The actual packwiz output format: "Mod Name: java.lang.Exception: Please go to URL and save this file to PATH"
func parsePackwizManuals(s string) []manualItem {
	// Pattern to match the exact packwiz output format:
	// "Aquaculture Delight (A Farmer's Delight Add-on): java.lang.Exception: This mod is excluded from the CurseForge API and must be downloaded manually."
	// "Please go to https://www.curseforge.com/minecraft/mc-mods/aquaculture-delight/files/6259758 and save this file to C:\path\to\file.jar"

	// Use a more flexible approach to find mod errors and URLs
	lines := strings.Split(s, "\n")
	seen := map[string]bool{}
	var items []manualItem

	for i, line := range lines {
		// Look for lines with mod errors
		if strings.Contains(line, "java.lang.Exception: This mod is excluded from the CurseForge API and must be downloaded manually.") {
			// Extract mod name from the beginning of the line
			if colonIndex := strings.Index(line, ":"); colonIndex > 0 {
				name := strings.TrimSpace(line[:colonIndex])

				// Filter out non-mod entries
				if strings.Contains(name, "Current version") ||
					strings.Contains(name, "at link.infra") ||
					strings.Contains(name, "java.base") ||
					len(name) == 0 {
					continue
				}

				// Look for the corresponding download URL in the next few lines
				for j := i + 1; j < i+10 && j < len(lines); j++ {
					nextLine := strings.TrimSpace(lines[j])
					if strings.Contains(nextLine, "Please go to ") && strings.Contains(nextLine, "curseforge.com") {
						// Extract URL and path
						if strings.Contains(nextLine, " and save this file to ") {
							parts := strings.Split(nextLine, " and save this file to ")
							if len(parts) == 2 {
								url := strings.TrimSpace(strings.TrimPrefix(parts[0], "Please go to "))
								path := strings.TrimSpace(parts[1])

								// Validate URL
								if strings.Contains(url, "curseforge.com") && url != "" && path != "" {
									key := url + "|" + strings.ToLower(path)
									if !seen[key] {
										seen[key] = true
										items = append(items, manualItem{Name: name, URL: url, Path: path})
										break // Found the URL for this mod, move to next mod
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return items
}

func assistManualFromPackwiz(items []manualItem) {
	if len(items) == 0 {
		return
	}

	logf("Downloading %d manual mod(s) directly from CurseForge...", len(items))

	// Ensure destination folders exist
	for _, it := range items {
		_ = os.MkdirAll(filepath.Dir(it.Path), 0755)
	}

	// Download all files directly
	var failedItems []manualItem
	for _, it := range items {
		logf("Downloading %s...", it.Name)
		logf("  From: %s", it.URL)
		logf("  To:   %s", it.Path)

		if err := downloadFromCurseForge(it.URL, it.Path); err != nil {
			logf("  Failed: %v", err)
			failedItems = append(failedItems, it)
		} else {
			logf("  ✓ Downloaded successfully")
		}
	}

	// Handle any failed downloads
	if len(failedItems) > 0 {
		logf("\n%d download(s) failed. These may require manual download:", len(failedItems))
		for _, it := range failedItems {
			logf(" - %s\n   %s\n   Save as: %s", it.Name, it.URL, it.Path)
		}

		if yesNoBox("Some downloads failed. Open remaining pages in browser?", launcherName+" - Download Failed") == 6 {
			for _, it := range failedItems {
				_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
			}

			// Wait for manual downloads
			for {
				logf("\nPress Enter after saving the files to re-check…")
				waitEnter()

				still := failedItems[:0]
				for _, it := range failedItems {
					if !exists(it.Path) {
						still = append(still, it)
					}
				}
				failedItems = still
				if len(failedItems) == 0 {
					logf("All manual items found. Continuing…")
					return
				}

				logf("Still missing:")
				for _, it := range failedItems {
					logf(" - %s -> %s", it.Name, it.Path)
				}
				if yesNoBox("Open the pages again?", launcherName) == 6 {
					for _, it := range failedItems {
						_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
					}
				}
			}
		}
	} else {
		logf("All manual downloads completed successfully!")
	}
}

// A neutral "press enter" we can use inside flows without saying "exit"
func waitEnter() {
	fmt.Fprint(out, "\nPress Enter to continue…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// Returns IDYES(6) or IDNO(7)
func yesNoBox(text, title string) int {
	const (
		MB_YESNO = 0x00000004
		IDYES    = 6
	)
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	r, _, _ := proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		MB_YESNO,
	)
	return int(r)
}
