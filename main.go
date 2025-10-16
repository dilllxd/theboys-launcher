// WinterPack.exe — portable Minecraft bootstrapper for Windows
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Fully portable: writes only beside the EXE (no AppData)
// - Downloads Prism Launcher (portable) — prefers MinGW w64 on amd64
// - Downloads Java 21 (Temurin JRE) dynamically (Adoptium API w/ GitHub fallback)
// - Downloads packwiz bootstrap dynamically (GitHub assets discovery)
// - Updates your packwiz pack, auto-sets RAM, launches Prism
// - Parses packwiz "manual download" errors: opens all pages, waits, retries once
// - Console output + logs/latest.log (rotates to logs/previous.log)
// - Keeps console open on error (press Enter), disable with WINTERPACK_NOPAUSE=1
//
// Build (set your version!):
//   go build -ldflags="-s -w -X main.version=v1.0.1" -o WinterPack.exe
//
// Usage for players: put WinterPack.exe in any writable folder and run it.

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	// Your hosted packwiz URL (GitHub Pages or any static host)
	PACK_URL      = "https://dilllxd.github.io/winterpack-modpack/pack.toml"
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

	// 1) Self-update (best-effort; skips downgrades)
	if err := selfUpdate(root, exePath); err != nil {
		logf("Update check failed: %v", err)
	}

	// 2) Ensure prerequisites
	prismDir := filepath.Join(root, "prism")
	jreDir := filepath.Join(root, "jre21")
	javaBin := filepath.Join(jreDir, "bin", "java.exe")
	bootstrapExe := filepath.Join(root, "packwiz-installer-bootstrap.exe")
	bootstrapJar := filepath.Join(root, "packwiz-installer-bootstrap.jar")

	if err := ensurePrism(prismDir); err != nil {
		fail(err)
	}
	if !exists(javaBin) {
		logf("Resolving Java 21 download…")
		jreURL, err := fetchJRE21ZipURL()
		if err != nil {
			fail(fmt.Errorf("failed to resolve Java 21 download: %w", err))
		}
		logf("Downloading Java 21… %s", jreURL)
		if err := downloadAndUnzipTo(jreURL, jreDir); err != nil {
			fail(err)
		}
		// Flatten typical top-level folder so jre21\bin\java.exe exists
		_ = flattenOneLevel(jreDir)
		if !exists(javaBin) {
			fail(errors.New("Java 21 installation looks incomplete (bin/java.exe not found)"))
		}
	}
	if !exists(bootstrapExe) && !exists(bootstrapJar) {
		logf("Resolving packwiz bootstrap…")
		pwURL, err := fetchPackwizBootstrapURL()
		if err != nil {
			fail(fmt.Errorf("failed to resolve packwiz bootstrap: %w", err))
		}
		logf("Downloading packwiz bootstrap… %s", pwURL)
		target := bootstrapExe
		if strings.HasSuffix(strings.ToLower(pwURL), ".jar") {
			target = bootstrapJar
		}
		if err := downloadTo(pwURL, target, 0755); err != nil {
			fail(err)
		}
	}

	// 3) Prepare instance
	instDir := filepath.Join(prismDir, "instances", INSTANCE_NAME)
	mcDir := filepath.Join(instDir, ".minecraft")
	if err := os.MkdirAll(mcDir, 0755); err != nil {
		fail(err)
	}

	// 4) Update pack (capture output; on manual-needed, assist and retry once)
	logf("Updating modpack…")
	var cmd *exec.Cmd
	if exists(bootstrapExe) {
		cmd = exec.Command(bootstrapExe, "-g", PACK_URL) // no -s path!
	} else if exists(bootstrapJar) {
		cmd = exec.Command(filepath.Join(jreDir, "bin", "java.exe"),
			"-jar", bootstrapJar, "-g", PACK_URL)
	} else {
		fail(errors.New("packwiz bootstrap not found after download"))
	}
	cmd.Dir = mcDir // run inside the .minecraft folder
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)

	var buf bytes.Buffer
	mw := io.MultiWriter(out, &buf)
	cmd.Stdout, cmd.Stderr = mw, mw

	err := cmd.Run()
	if err != nil {
		// Parse packwiz output for manual-download instructions
		items := parsePackwizManuals(buf.String())
		if len(items) > 0 {
			assistManualFromPackwiz(items)
			// Retry ONCE after user saves files
			buf.Reset()
			err = cmd.Run()
		}
	}
	if err != nil {
		fail(fmt.Errorf("packwiz update failed: %w", err))
	}

	// 5) Write instance config (RAM)
	if err := writeInstanceCfg(instDir); err != nil {
		fail(err)
	}

	// 6) Launch Prism (portable)
	logf("Launching Prism…")
	prismExe := filepath.Join(prismDir, "PrismLauncher.exe")
	launch := exec.Command(prismExe, "--dir", ".", "--launch", INSTANCE_NAME)
	launch.Dir = prismDir
	launch.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	launch.Stdout, launch.Stderr = out, out
	_ = launch.Run()
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

// A neutral "press enter" we can use inside flows without saying "exit"
func waitEnter() {
	fmt.Fprint(out, "\nPress Enter to continue…")
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

func downloadTo(url, path string, mode os.FileMode) error {
	b, err := download(url)
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
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "WinterPack/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
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

// -------------------- Instance config --------------------

func writeInstanceCfg(instDir string) error {
	minMB, maxMB := autoRAM()
	lines := []string{
		"InstanceType=Minecraft",
		"iconKey=default",
		"OverrideMemory=true",
		fmt.Sprintf("MinMemAlloc=%d", minMB),
		fmt.Sprintf("MaxMemAlloc=%d", maxMB),
		"Notes=Managed by WinterPack launcher",
	}
	if err := os.MkdirAll(instDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(instDir, "instance.cfg"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
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

// Auto RAM: ~50% of total (cap 10 GB, floor 2/4 GB)
func autoRAM() (minMB, maxMB int) {
	total := totalRAMMB()
	if total <= 0 {
		total = 8192
	}
	maxMB = min(int(float64(total)*0.5), 10240)
	if maxMB < 4096 && total >= 8192 {
		maxMB = 4096
	}
	minMB = max(2048, maxMB/2)
	return
}

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

// -------------------- Packwiz manual download parsing & assist --------------------

type manualItem struct {
	Name string // optional; may be empty
	URL  string
	Path string // absolute path from packwiz message
}

// Parse lines like:
// "<Mod Name>: ... Please go to https://... and save this file to C:\...\mods\file.jar"
func parsePackwizManuals(s string) []manualItem {
	re := regexp.MustCompile(`(?s)(?m)^(?P<name>.+?):.*?Please go to (?P<url>https?://\S+)\s+and save this file to (?P<path>.+?)(?:\r?\n|$)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var items []manualItem
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		u := strings.TrimSpace(m[2])
		p := strings.TrimSpace(m[3])
		key := u + "|" + strings.ToLower(p)
		if seen[key] {
			continue
		}
		seen[key] = true
		items = append(items, manualItem{Name: name, URL: u, Path: p})
	}
	return items
}

func assistManualFromPackwiz(items []manualItem) {
	if len(items) == 0 {
		return
	}

	// Ask to open all pages
	if yesNoBox("Some mods must be downloaded manually.\n\nOpen ALL download pages now?", "WinterPack – Manual Downloads") == 6 {
		for _, it := range items {
			_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
			logf(" - %s\n   %s\n   Save as: %s", it.Name, it.URL, it.Path)
		}
	} else {
		for _, it := range items {
			logf(" - %s\n   %s\n   Save as: %s", it.Name, it.URL, it.Path)
		}
	}

	// Ensure destination folders exist
	for _, it := range items {
		_ = os.MkdirAll(filepath.Dir(it.Path), 0755)
	}

	// Loop until all files present
	for {
		logf("\nPress Enter after saving the files to re-check…")
		waitEnter()

		still := items[:0]
		for _, it := range items {
			if !exists(it.Path) {
				still = append(still, it)
			}
		}
		items = still
		if len(items) == 0 {
			logf("All manual items found. Continuing…")
			return
		}

		logf("Still missing:")
		for _, it := range items {
			logf(" - %s -> %s", it.Name, it.Path)
		}
		if yesNoBox("Open the pages again?", "WinterPack") == 6 {
			for _, it := range items {
				_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", it.URL).Start()
			}
		}
	}
}
