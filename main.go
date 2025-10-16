// WinterPack.exe — portable Minecraft bootstrapper for Windows
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Fully portable: writes only beside the EXE (no AppData)
// - Downloads Prism Launcher (portable) — prefers MinGW w64 on amd64
// - Downloads Java 21 (Temurin JRE) dynamically (Adoptium API w/ GitHub fallback)
// - Downloads packwiz bootstrap dynamically (GitHub assets discovery)
// - Creates instance beside the EXE, writes instance.cfg (name/RAM/Java)
// - Runs packwiz from the *instance root* (detects MultiMC/Prism mode)
// - Console output + logs/latest.log (rotates to logs/previous.log)
// - Keeps console open on error (press Enter), disable with WINTERPACK_NOPAUSE=1
// - Optional cache-bust for PACK_URL: set WINTERPACK_CACHEBUST=1
//
// Build (set your version!):
//   go build -ldflags="-s -w -X main.version=v1.0.3" -o WinterPack.exe
//
// Usage for players: put WinterPack.exe in any writable folder and run it.

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
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
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	// Use Raw (recommended for byte-stable text)
	PACK_URL      = "https://raw.githubusercontent.com/dilllxd/winterpack-modpack/main/pack.toml"
	INSTANCE_NAME = "WinterPack"

	// Self-update source (GitHub Releases of this EXE)
	UPDATE_OWNER = "dilllxd"             // GitHub username/org
	UPDATE_REPO  = "winterpack-launcher" // repo that hosts releases for this EXE
	UPDATE_ASSET = "WinterPack.exe"      // asset name in releases to download
)

// Optional: show MessageBox popups (false = log to console/file)
var interactive = false

// Populated at build time via -X main.version=vX.Y.Z
var version = "dev"

// global writer used by log/fail and for piping subprocess output
var out io.Writer = os.Stdout

// -------------------- MAIN --------------------

func main() {
	if runtime.GOOS != "windows" {
		msgBox("Windows only", "WinterPack", 0)
		return
	}

	exePath, _ := os.Executable()
	root := filepath.Dir(exePath)

	// 0) Logging: console + logs/latest.log (rotate previous.log)
	closeLog := setupLogging(root)
	defer closeLog()

	logf("=== WinterPack Launcher started %s ===", time.Now().Format(time.RFC1123))
	logf("Version: %s", version)

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

	// Run the launcher logic directly
	runLauncherLogic(root, exePath, &prismProcess)
}

// -------------------- Launcher Logic --------------------

func runLauncherLogic(root, exePath string, prismProcess **os.Process) {
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

	fmt.Fprintf(out, "Setting up prerequisites...\n")

	// Prism Launcher
	if err := ensurePrism(prismDir); err != nil {
		fail(err)
	} else {
		fmt.Fprintf(out, "✓ Prism Launcher ready\n")
	}

	// Java 21
	if !exists(javaBin) {
		fmt.Fprintf(out, "Setting up Java 21...\n")
		jreURL, err := fetchJRE21ZipURL()
		if err != nil {
			fail(fmt.Errorf("failed to resolve Java 21 download: %w", err))
		}
		if err := downloadAndUnzipTo(jreURL, jreDir); err != nil {
			fail(err)
		}
		// Flatten typical top-level folder so jre21\bin\java.exe exists
		_ = flattenOneLevel(jreDir)
		if !exists(javaBin) {
			fail(errors.New("Java 21 installation looks incomplete (bin/java.exe not found)"))
		}
		fmt.Fprintf(out, "✓ Java 21 installed\n")
	} else {
		fmt.Fprintf(out, "✓ Java 21 already installed\n")
	}

	// Packwiz bootstrap
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
		fmt.Fprintf(out, "✓ Packwiz bootstrap installed\n")
	} else {
		fmt.Fprintf(out, "✓ Packwiz bootstrap already installed\n")
	}

	// 3) Create proper MultiMC/Prism instance first
	instDir := filepath.Join(prismDir, "instances", INSTANCE_NAME)
	mcDir := filepath.Join(instDir, "minecraft") // Use minecraft, not .minecraft
	if err := os.MkdirAll(mcDir, 0755); err != nil {
		fail(err)
	}

	// 4) Create proper MultiMC/Prism instance structure with Forge 1.20.1 (only if needed)
	instanceConfigFile := filepath.Join(instDir, "instance.cfg")
	mmcPackFile := filepath.Join(instDir, "mmc-pack.json")

	needsInstanceCreation := !exists(instanceConfigFile) || !exists(mmcPackFile)
	if needsInstanceCreation {
		logf("Creating MultiMC/Prism instance with Forge 1.20.1...")
		if err := createMultiMCInstance(instDir, javaBin); err != nil {
			fail(fmt.Errorf("failed to create MultiMC instance: %w", err))
		}
	} else {
		logf("Instance already exists, skipping creation...")
	}

	// 5) Install Forge properly for the instance (only if needed)
	// Check for Forge installation in MultiMC/Prism instance structure
	forgeJar := filepath.Join(mcDir, "libraries", "net", "minecraftforge", "forge", "1.20.1-47.4.0", "forge-1.20.1-47.4.0-universal.jar")

	forgeInstalled := exists(forgeJar) && exists(mmcPackFile)

	if !forgeInstalled {
		logf("Installing Forge 1.20.1 for the instance...")
		if err := installForgeForInstance(instDir, javaBin); err != nil {
			fail(fmt.Errorf("failed to install Forge: %w", err))
		}
	} else {
		logf("Forge already installed, skipping installation...")
	}

	// 6) Check for modpack updates
	logf("Checking for modpack updates...")
	updateAvailable, localVersion, remoteVersion, err := checkModpackUpdate(instDir)
	if err != nil {
		logf("Warning: Failed to check modpack updates: %v", err)
		// Continue with packwiz anyway
		updateAvailable = true // Assume update needed to be safe
	}

	// 7) Now run packwiz from within the instance to install/update the modpack
	var action string
	var backupPath string

	if updateAvailable {
		if localVersion == "" {
			action = "Installing"
		} else {
			action = "Updating"
			// Create backup before updating
			logf("Creating backup before update...")
			backupPath, err = createModpackBackup(mcDir)
			if err != nil {
				logf("Warning: Backup creation failed: %v", err)
			}
		}
		if localVersion == "" {
			logf("%s modpack version %s with packwiz…", action, remoteVersion)
		} else {
			logf("%s modpack %s → %s with packwiz…", action, localVersion, remoteVersion)
		}
	} else {
		logf("Modpack is up to date, verifying installation with packwiz…")
	}

	packURL := PACK_URL
	if os.Getenv("WINTERPACK_CACHEBUST") == "1" {
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
			logf("Packwiz update failed, attempting to restore from backup...")
			if restoreErr := restoreModpackBackup(backupPath, mcDir); restoreErr != nil {
				logf("Warning: Failed to restore backup: %v", restoreErr)
			} else {
				logf("Successfully restored previous modpack version")
			}
		}
		fail(fmt.Errorf("packwiz update failed: %w", err))
	}

	// Post-update verification and version saving
	if updateAvailable {
		logf("Verifying installation was completed successfully...")

		// Save the version that packwiz just installed
		if err := saveLocalVersion(instDir, remoteVersion); err != nil {
			logf("Warning: Failed to save local version: %v", err)
		} else {
			logf("✓ Installation completed successfully! Now running version %s", remoteVersion)
		}
	} else {
		logf("✓ Installation verification completed")
	}

	// 8) Launch WinterPack instance directly
	logf("Launching WinterPack instance...")

	prismExe := filepath.Join(prismDir, "PrismLauncher.exe")
	launch := exec.Command(prismExe, "--dir", ".", "--launch", INSTANCE_NAME)
	launch.Dir = prismDir
	launch.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	launch.Stdout, launch.Stderr = out, out

	// Start the process and wait for it to complete (keeps console open)
	if err := launch.Start(); err != nil {
		logf("Failed to launch instance: %v", err)
		logf("Falling back to GUI mode...")
		launchFallback := exec.Command(prismExe, "--dir", ".")
		launchFallback.Dir = prismDir
		launchFallback.Env = append(os.Environ(),
			"JAVA_HOME="+jreDir,
			"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
		)
		launchFallback.Stdout, launchFallback.Stderr = out, out
		if err := launchFallback.Start(); err != nil {
			logf("Failed to launch GUI mode: %v", err)
			return
		}
		*prismProcess = launchFallback.Process
		logf("WinterPack GUI launched (PID: %d)", launchFallback.Process.Pid)
		// Wait for the GUI process to complete
		launchFallback.Wait()
	} else {
		// Store the process reference for signal handling
		*prismProcess = launch.Process
		logf("WinterPack instance launched (PID: %d)", launch.Process.Pid)
		// Wait for the game process to complete
		launch.Wait()
	}

	logf("Prism Launcher closed")
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
		msgBox(fmt.Sprintf(format, a...), "WinterPack", 0)
	} else {
		fmt.Fprintf(out, format+"\n", a...)
	}
}

func pauseIfWanted() {
	if os.Getenv("WINTERPACK_NOPAUSE") == "1" {
		return
	}
	fmt.Fprint(out, "\nPress Enter to exit…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func fail(err error) {
	if interactive {
		msgBox("Error: "+err.Error(), "WinterPack", 0)
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
		logf("WinterPack up to date (%s).", version)
		return nil
	case 1:
		// local > remote → don't downgrade
		logf("Local launcher (%s) is newer than latest release (%s). Skipping update.", version, tag)
		return nil
	case -1:
		// remote > local → proceed
	}

	logf("New WinterPack available: %s (current %s). Updating…", tag, version)

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
	req.Header.Set("User-Agent", "WinterPack-Updater/1.0")
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
	req.Header.Set("User-Agent", "WinterPack/1.0")
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

func ensurePrism(dir string) error {
	if exists(filepath.Join(dir, "PrismLauncher.exe")) {
		return nil
	}
	url, err := fetchLatestPrismPortableURL()
	if err != nil {
		return err
	}
	logf("Downloading Prism portable: %s", url)
	if err := downloadAndUnzipTo(url, dir); err != nil {
		return err
	}
	// Force portable mode
	cfg := filepath.Join(dir, "prismlauncher.cfg")
	_ = os.WriteFile(cfg, []byte("Portable=true\n"), 0644)
	return nil
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
	req.Header.Set("User-Agent", "WinterPack-Prism/1.0")
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
	req.Header.Set("User-Agent", "WinterPack-Adoptium/1.0")
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
	req2.Header.Set("User-Agent", "WinterPack-Adoptium/1.0")
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
	req.Header.Set("User-Agent", "WinterPack-Packwiz/1.0")
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

func createMultiMCInstance(instDir, javaExe string) error {
	minMB, maxMB := autoRAM()

	// Create instance.cfg
	instanceLines := []string{
		"InstanceType=OneSix", // Use OneSix not Minecraft
		"name=" + INSTANCE_NAME,
		"iconKey=default",
		"OverrideMemory=true",
		fmt.Sprintf("MinMemAlloc=%d", minMB),
		fmt.Sprintf("MaxMemAlloc=%d", maxMB),
		"OverrideJava=true",
		"JavaPath=" + javaExe,
		"Notes=Managed by WinterPack launcher",
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
				"uid": "net.minecraftforge",
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
				"uid": "net.minecraftforge",
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
func fetchRemotePackVersion() (string, error) {
	req, err := http.NewRequest("GET", PACK_URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "WinterPack/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, PACK_URL)
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
func getLocalPackVersion(instDir string) (string, error) {
	versionFilePath := filepath.Join(instDir, ".winterpack-version")

	// Check if our version file exists
	if !exists(versionFilePath) {
		logf("Debug: WinterPack version file not found at %s", versionFilePath)
		return "", nil // No version file exists
	}

	body, err := os.ReadFile(versionFilePath)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(body))
	logf("Debug: Found local WinterPack version %s at %s", version, versionFilePath)
	return version, nil
}

// saveLocalVersion saves the current modpack version to our tracking file
func saveLocalVersion(instDir, version string) error {
	versionFilePath := filepath.Join(instDir, ".winterpack-version")

	if err := os.WriteFile(versionFilePath, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to save local version: %w", err)
	}

	logf("Debug: Saved local WinterPack version %s to %s", version, versionFilePath)
	return nil
}

// checkModpackUpdate checks if there's a modpack update available
func checkModpackUpdate(instDir string) (bool, string, string, error) {
	remoteVersion, err := fetchRemotePackVersion()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch remote modpack version: %w", err)
	}

	localVersion, err := getLocalPackVersion(instDir)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get local modpack version: %w", err)
	}

	// If no local version exists, we need to install
	if localVersion == "" {
		logf("No local modpack found, will install version %s", remoteVersion)
		return true, "", remoteVersion, nil
	}

	// Compare versions
	if localVersion != remoteVersion {
		logf("Modpack update available: %s → %s", localVersion, remoteVersion)
		return true, localVersion, remoteVersion, nil
	}

	logf("Modpack is up to date (%s)", localVersion)
	return false, localVersion, remoteVersion, nil
}

// -------------------- Modpack Backup & Restore --------------------

// createModpackBackup creates a backup of the current modpack before updating
func createModpackBackup(mcDir string) (string, error) {
	// Clean up old backups (keep only the 3 most recent)
	if err := cleanupOldBackups(mcDir, 3); err != nil {
		logf("Warning: Failed to cleanup old backups: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupName := fmt.Sprintf("winterpack-backup-%s", timestamp)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(mcDir)))
	backupPath := filepath.Join(rootDir, "util", "backups", backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Directories to backup
	dirsToBackup := []string{"mods", "config", "resourcepacks", "shaderpacks"}

	logf("Creating backup at %s...", backupName)

	var backedUpItems []string
	for _, dir := range dirsToBackup {
		srcPath := filepath.Join(mcDir, dir)
		dstPath := filepath.Join(backupPath, dir)

		if exists(srcPath) {
			if err := copyDir(srcPath, dstPath); err != nil {
				logf("Warning: Failed to backup %s: %v", dir, err)
			} else {
				backedUpItems = append(backedUpItems, dir)
			}
		}
	}

	// Backup our version file if it exists
	versionFileSrc := filepath.Join(filepath.Dir(mcDir), ".winterpack-version")
	versionFileDst := filepath.Join(backupPath, ".winterpack-version")
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("Warning: Failed to backup version file: %v", err)
		} else {
			backedUpItems = append(backedUpItems, ".winterpack-version")
		}
	}

	if len(backedUpItems) == 0 {
		logf("No modpack files found to backup")
		return "", nil
	}

	logf("Backup created: %s (items: %s)", backupName, strings.Join(backedUpItems, ", "))
	return backupPath, nil
}

// restoreModpackBackup restores from a backup if the update fails
func restoreModpackBackup(backupPath, mcDir string) error {
	if backupPath == "" || !exists(backupPath) {
		return errors.New("no backup available to restore")
	}

	logf("Restoring from backup...")

	// Remove current modpack directories
	dirsToRemove := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	for _, dir := range dirsToRemove {
		dirPath := filepath.Join(mcDir, dir)
		if exists(dirPath) {
			if err := os.RemoveAll(dirPath); err != nil {
				logf("Warning: Failed to remove %s during restore: %v", dir, err)
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
				logf("Warning: Failed to restore %s: %v", dir, err)
			} else {
				restoredItems = append(restoredItems, dir)
			}
		}
	}

	// Restore our version file
	versionFileSrc := filepath.Join(backupPath, ".winterpack-version")
	versionFileDst := filepath.Join(filepath.Dir(mcDir), ".winterpack-version")
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("Warning: Failed to restore version file: %v", err)
		} else {
			restoredItems = append(restoredItems, ".winterpack-version")
		}
	}

	if len(restoredItems) == 0 {
		return errors.New("nothing to restore from backup")
	}

	logf("Restored: %s", strings.Join(restoredItems, ", "))
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
func cleanupOldBackups(mcDir string, keepCount int) error {
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
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "winterpack-backup-") {
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
			logf("Warning: Failed to remove old backup %s: %v", entry.Name(), err)
		} else {
			logf("Removed old backup: %s", entry.Name())
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
	req.Header.Set("User-Agent", "WinterPack/1.0")

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
	req.Header.Set("User-Agent", "WinterPack/1.0")

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
				for j := i + 1; j < i + 10 && j < len(lines); j++ {
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

		if yesNoBox("Some downloads failed. Open remaining pages in browser?", "WinterPack – Download Failed") == 6 {
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
				if yesNoBox("Open the pages again?", "WinterPack") == 6 {
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
