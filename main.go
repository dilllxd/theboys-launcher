// TheBoysLauncher.exe - portable Minecraft bootstrapper for Windows
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Fully portable: writes only beside the EXE (no AppData)
// - Downloads Prism Launcher (portable) - prefers MinGW w64 on amd64
// - Downloads Java dynamically based on Minecraft version (Temurin JRE) (Adoptium API w/ GitHub fallback)
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
	runtime.LockOSThread()

	opts := parseOptions()

	if opts.cleanupAfterUpdate {
		// This is a cleanup run after an update
		performUpdateCleanup(opts.cleanupOldExe, opts.cleanupNewExe)
		return
	}

	if runtime.GOOS != "windows" {
		fail(errors.New("Windows only"))
	}

	exePath, _ := os.Executable()
	root := filepath.Dir(exePath)

	// 0) Logging: console + logs/latest.log (rotate previous.log)
	closeLog := setupLogging(root)
	defer closeLog()

	// 1) Load settings
	if err := loadSettings(root); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to load settings: %v", err)))
	}

	// Show beautiful welcome message
	logf("\n%s", headerLine(launcherName))
	logf("%s", dividerLine())
	logf("%s", infoLine(fmt.Sprintf("Version %s • Started at %s", version, time.Now().Format("3:04 PM"))))
	logf("%s", infoLine(fmt.Sprintf("Memory allocation: %d GB", settings.MemoryMB/1024)))
	logf("%s", dividerLine())

	// Check for launcher updates immediately on startup
	logf("%s", sectionLine("Launcher Update Check"))
	if err := selfUpdate(root, exePath); err != nil {
		logf("Update check failed: %v", err)
	}

	modpacks := loadModpacks(root)
	if len(modpacks) == 0 {
		fail(errors.New("no modpacks configured"))
	}

	if opts.listModpacks {
		printModpackList(modpacks)
		return
	}

	if opts.openSettings {
		runSettingsMenu(root)
		return
	}

	// Default to TUI mode unless explicitly using CLI
	if !opts.useCLI {
		// TUI Mode - Show main menu with modpack selection and settings
		var selectedModpack Modpack
		var settingsChosen bool

		for {
			selectedModpack, settingsChosen = runMainTUI(modpacks)
			if settingsChosen {
				runSettingsMenu(root)
				// After settings, show main menu again (continue loop)
				continue
			}

			if selectedModpack.ID == "" {
				logf("%s", infoLine("No modpack selected. Exiting."))
				return
			}

			// User selected a modpack, break the loop and launch it
			break
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

		logf("%s", infoLine(fmt.Sprintf("Launching modpack: %s (%s)", modpackLabel(selectedModpack), selectedModpack.ID)))
		runLauncherLogic(root, exePath, selectedModpack, &prismProcess)
		return
	}

	// CLI Mode - Use console interface
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

	logf("Launching (CLI) modpack: %s (%s)", modpackLabel(selectedModpack), selectedModpack.ID)
	runLauncherLogic(root, exePath, selectedModpack, &prismProcess)
}