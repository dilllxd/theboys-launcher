// WinterPack.exe — portable Minecraft bootstrapper for Windows
// - Self-updates from GitHub Releases (latest tag)
// - Fully portable: writes only beside the EXE (no AppData)
// - Downloads Prism Launcher (portable), Java 21, packwiz bootstrap
// - Updates your packwiz pack, auto-sets RAM, launches Prism
//
// Build (set your version!):
//   go build -ldflags="-s -w -H=windowsgui -X main.version=v1.0.0" -o WinterPack.exe
//
// Usage for players: put WinterPack.exe in any writable folder and run it.

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// -------------------- CONFIG: EDIT THESE --------------------

const (
	// Your hosted packwiz URL (GitHub Pages or any static host)
	PACK_URL      = "https://dilllxd.github.io/winterpack-modpack/pack.toml"
	INSTANCE_NAME = "WinterPack"

	// Self-update source (GitHub Releases of this EXE)
	UPDATE_OWNER = "dilllxd"       // GitHub username/org
	UPDATE_REPO  = "winterpack-launcher"   // repo that hosts releases for this EXE
	UPDATE_ASSET = "WinterPack.exe"        // asset name in releases to download

	// Official downloads
	PRISM_ZIP_URL    = "https://github.com/PrismLauncher/PrismLauncher/releases/latest/download/PrismLauncher-Windows-MSVC-portable-64.zip"
	PACKWIZ_BOOT_URL = "https://github.com/packwiz/packwiz-installer-bootstrap/releases/latest/download/packwiz-installer-bootstrap-windows-amd64.exe"
	// Java 21 (Temurin JRE, Windows x64)
	JRE_ZIP_URL = "https://github.com/adoptium/temurin21-binaries/releases/latest/download/OpenJDK21U-jre_x64_windows_hotspot.zip"
)

// Optional: show MessageBox popups (false = log to console)
var interactive = false

// Populated at build time via -X main.version=vX.Y.Z
var version = "dev"

// -------------------- MAIN --------------------

func main() {
	if runtime.GOOS != "windows" {
		msgBox("Windows only", "WinterPack", 0)
		return
	}
	exePath, _ := os.Executable()
	root := filepath.Dir(exePath)

	// 1) Self-update (best-effort)
	if err := selfUpdate(root, exePath); err != nil {
		log("Update check failed: %v", err)
	}

	// 2) Ensure prerequisites
	prismDir := filepath.Join(root, "prism")
	jreDir := filepath.Join(root, "jre21")
	javaBin := filepath.Join(jreDir, "bin", "java.exe")
	bootstrap := filepath.Join(root, "packwiz-installer-bootstrap.exe")

	if err := ensurePrism(prismDir); err != nil { fail(err) }
	if !exists(javaBin) {
		log("Downloading Java 21…")
		if err := downloadAndUnzipTo(JRE_ZIP_URL, jreDir); err != nil { fail(err) }
	}
	if !exists(bootstrap) {
		log("Downloading packwiz bootstrap…")
		if err := downloadTo(PACKWIZ_BOOT_URL, bootstrap, 0755); err != nil { fail(err) }
	}

	// 3) Prepare instance
	instDir := filepath.Join(prismDir, "instances", INSTANCE_NAME)
	mcDir := filepath.Join(instDir, ".minecraft")
	if err := os.MkdirAll(mcDir, 0755); err != nil { fail(err) }

	// 4) Update pack
	log("Updating modpack…")
	cmd := exec.Command(bootstrap, "-s", mcDir, "-g", PACK_URL)
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil { fail(fmt.Errorf("packwiz update failed: %w", err)) }

	// 5) Write instance config (RAM)
	if err := writeInstanceCfg(instDir); err != nil { fail(err) }

	// 6) Launch Prism (portable)
	log("Launching Prism…")
	prismExe := filepath.Join(prismDir, "PrismLauncher.exe")
	launch := exec.Command(prismExe, "--dir", ".", "--launch", INSTANCE_NAME)
	launch.Dir = prismDir
	launch.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
	)
	_ = launch.Run()
}

// -------------------- Self-update --------------------

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func selfUpdate(root, exePath string) error {
	tag, assetURL, err := fetchLatestAsset(UPDATE_OWNER, UPDATE_REPO, UPDATE_ASSET)
	if err != nil || tag == "" || assetURL == "" { return err }
	if tag == version {
		log("WinterPack up to date (%s).", version)
		return nil
	}
	log("New WinterPack available: %s (current %s). Updating…", tag, version)

	tmpNew := exePath + ".new"
	if err := downloadTo(assetURL, tmpNew, 0755); err != nil { return err }

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
	if err != nil { return "", "", err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil { return "", "", err }
	for _, a := range r.Assets {
		if a.Name == wantName {
			return r.TagName, a.BrowserDownloadURL, nil
		}
	}
	return r.TagName, "", errors.New("asset not found in latest release")
}

// -------------------- Downloads / Unzip --------------------

func downloadTo(url, path string, mode os.FileMode) error {
	b, err := download(url)
	if err != nil { return err }
	return os.WriteFile(path, b, mode)
}

func downloadAndUnzipTo(url, dest string) error {
	b, err := download(url)
	if err != nil { return err }
	return unzipBytesTo(b, dest)
}

func download(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "WinterPack/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func unzipBytesTo(b []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil { return err }
	for _, f := range r.File {
		p := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0755); err != nil { return err }
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil { return err }
		rc, err := f.Open(); if err != nil { return err }
		out, err := os.Create(p); if err != nil { rc.Close(); return err }
		_, err = io.Copy(out, rc)
		out.Close(); rc.Close()
		if err != nil { return err }
	}
	return nil
}

// -------------------- Prism + Instance --------------------

func ensurePrism(dir string) error {
	if exists(filepath.Join(dir, "PrismLauncher.exe")) { return nil }
	log("Downloading Prism portable…")
	if err := downloadAndUnzipTo(PRISM_ZIP_URL, dir); err != nil { return err }
	// Force portable mode
	cfg := filepath.Join(dir, "prismlauncher.cfg")
	_ = os.WriteFile(cfg, []byte("Portable=true\n"), 0644)
	return nil
}

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
	if err := os.MkdirAll(instDir, 0755); err != nil { return err }
	return os.WriteFile(filepath.Join(instDir, "instance.cfg"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// -------------------- Helpers --------------------

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func log(fmtStr string, a ...any) {
	if interactive {
		msgBox(fmt.Sprintf(fmtStr, a...), "WinterPack", 0)
	} else {
		fmt.Printf(fmtStr+"\n", a...)
	}
}

func fail(err error) {
	if interactive {
		msgBox("Error: "+err.Error(), "WinterPack", 0)
	} else {
		fmt.Println("Error:", err)
	}
	os.Exit(1)
}

// Auto RAM: ~50% of total (cap 10 GB, floor 2/4 GB)
func autoRAM() (minMB, maxMB int) {
	total := totalRAMMB()
	if total <= 0 { total = 8192 }
	maxMB = min(int(float64(total)*0.5), 10240)
	if maxMB < 4096 && total >= 8192 { maxMB = 4096 }
	minMB = max(2048, maxMB/2)
	return
}

func totalRAMMB() int {
	type mstat struct {
		dwLen uint32; load uint32; total, avail, a,b,c,d uint64
	}
	k32 := syscall.NewLazyDLL("kernel32.dll")
	proc := k32.NewProc("GlobalMemoryStatusEx")
	var s mstat
	s.dwLen = uint32(unsafe.Sizeof(s))
	r1,_,_ := proc.Call(uintptr(unsafe.Pointer(&s)))
	if r1 == 0 { return 0 }
	return int(s.total / (1024 * 1024))
}

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }

// Simple Windows MessageBox (optional)
func msgBox(text, title string, flags uintptr) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	_, _, _ = proc.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), flags)
}
