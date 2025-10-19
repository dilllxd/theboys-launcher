//go:generate goversioninfo -64 -icon=icon.ico

// TheBoysLauncher.exe - Minecraft bootstrapper for Windows with Fyne GUI
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Stores data in user's home directory (C:\Users\Username\.theboys-launcher)
// - Downloads Prism Launcher (portable) - prefers MinGW w64 on amd64
// - Downloads Java dynamically based on Minecraft version (Temurin JRE) (Adoptium API w/ GitHub fallback)
// - Downloads packwiz bootstrap dynamically (GitHub assets discovery)
// - Creates instance in launcher home directory, writes instance.cfg (name/RAM/Java)
// - Runs packwiz from the *instance root* (detects MultiMC/Prism mode)
// - Console output + logs/latest.log (rotates to logs/previous.log)
// - Optional cache-bust for the modpack URL: set THEBOYS_CACHEBUST=1
// - Uses Fyne GUI for modpack selection and configuration
// - Supports multiple modpacks via modpacks.json
//
// Build (set your version!):
//   go generate
//   go build -ldflags="-s -w -X main.version=v3.0.0" -o TheBoysLauncher.exe
//
// Usage for players: run TheBoysLauncher.exe from any location. Data will be stored in your home directory.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// VERY EARLY LOGGING - this should write even if everything else fails
	exePath, _ := os.Executable()
	launcherHome := filepath.Dir(exePath)
	logDir := filepath.Join(launcherHome, "logs")
	os.MkdirAll(logDir, 0755)

	runtime.LockOSThread()

	hideConsoleWindow()

	opts := parseOptions()

	if opts.cleanupAfterUpdate {
		// This is a cleanup run after an update
		performUpdateCleanup(opts.cleanupOldExe, opts.cleanupNewExe)
		return
	}

	if runtime.GOOS != "windows" {
		fail(errors.New("Windows only"))
	}

	root := getLauncherHome()

	// Set up emergency crash logger BEFORE anything else that might crash
	setupEmergencyCrashLogger(root)

	// 0) Logging: console + logs/latest.log (rotate previous.log)
	closeLog := setupLogging(root)
	defer closeLog()

	// Hide any console window that might have appeared during initialization
	hideConsoleWindow()

	// 1) Load settings
	if err := loadSettings(root); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to load settings: %v", err)))
	} else {
	}

	// Show beautiful welcome message
	logf("\n%s", headerLine(launcherName))
	logf("%s", dividerLine())
	logf("%s", infoLine(fmt.Sprintf("Version %s • Started at %s", version, time.Now().Format("3:04 PM"))))
	logf("%s", infoLine(fmt.Sprintf("Detected system RAM: %d GB", totalRAMMB()/1024)))
	logf("%s", infoLine(fmt.Sprintf("Memory allocation: %d GB", settings.MemoryMB/1024)))
	logf("%s", dividerLine())

	modpacks := loadModpacks(root)
	if len(modpacks) == 0 {
		fail(errors.New("no modpacks configured"))
	} else {
	}

	// Set up signal handling for force-closing Prism and Minecraft on launcher exit
	var prismProcess *os.Process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logf("%s", warnLine("Launcher interrupted, force-closing Prism Launcher and Minecraft..."))

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

	// Launch the GUI
	logf("Starting modern GUI interface...")
	gui := NewGUI(modpacks, root)

	gui.launchWithCallback(&prismProcess, root, exePath)
}
