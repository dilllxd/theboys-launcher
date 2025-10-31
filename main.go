//go:generate goversioninfo -64 -icon=icon.ico

// TheBoysLauncher - Minecraft bootstrapper with Fyne GUI
// - Self-updates from GitHub Releases (latest tag, no downgrades)
// - Stores data in user's home directory (~/.theboyslauncher)
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
//   go build -ldflags="-s -w -X main.version=v3.0.0" -o TheBoysLauncher
//
// Usage for players: run TheBoysLauncher from any location. Data will be stored in your home directory.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	runtime.LockOSThread()
	hideConsoleWindow()

	// Get executable path for potential use by GUI
	exePath, _ := os.Executable()

	opts := parseOptions()

	if opts.cleanupAfterUpdate {
		// This is a cleanup run after an update
		performUpdateCleanup(opts.cleanupOldExe, opts.cleanupNewExe)
		return
	}

	// Platform check now handled by platform abstraction
	// Windows hard block removed for cross-platform support

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
	logf("%s", infoLine(fmt.Sprintf("Detected system RAM: %d GB",
		roundToNearestGB(totalRAMMB()))))
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
		// Use platform-specific process management
		forceCloseAllProcesses(prismProcess)
		os.Exit(1)
	}()

	// Launch the GUI
	logf("Starting modern GUI interface...")
	gui := NewGUI(modpacks, root)

	gui.launchWithCallback(&prismProcess, root, exePath)
}
