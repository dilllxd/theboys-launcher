package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// -------------------- Qt Environment Helper Functions --------------------

// buildQtEnvironment builds Qt-specific environment variables for Linux
func buildQtEnvironment(prismDir, jreDir string) []string {
	qtEnv := []string{
		"JAVA_HOME=" + jreDir,
		"PATH=" + BuildPathEnv(filepath.Join(jreDir, "bin")),
	}

	// Add Qt-specific environment variables for Linux only
	if runtime.GOOS == "linux" {
		// Use the actual base directory where Prism executable is located
		actualPrismDir := getPrismBaseDir(prismDir)

		// Set Qt plugin path to bundled plugins directory
		qtPluginPath := filepath.Join(actualPrismDir, "plugins")
		if exists(qtPluginPath) {
			qtEnv = append(qtEnv, "QT_PLUGIN_PATH="+qtPluginPath)
			logf("DEBUG: Setting QT_PLUGIN_PATH=%s", qtPluginPath)
		}

		// Set library path to bundled libraries directory
		qtLibPath := filepath.Join(actualPrismDir, "lib")
		if exists(qtLibPath) {
			// Prepend to LD_LIBRARY_PATH to prioritize bundled libraries
			existingLdPath := os.Getenv("LD_LIBRARY_PATH")
			if existingLdPath != "" {
				qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath+":"+existingLdPath)
			} else {
				qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath)
			}
			logf("DEBUG: Setting LD_LIBRARY_PATH=%s", qtLibPath)
		}

		// Additional Qt environment variables for better compatibility
		qtEnv = append(qtEnv, "QT_QPA_PLATFORM=xcb")           // Force X11 backend
		qtEnv = append(qtEnv, "QT_XCB_GL_INTEGRATION=xcb_glx") // OpenGL integration
	}

	return qtEnv
}

// logQtEnvironment logs Qt environment setup for debugging
func logQtEnvironment(prismDir string) {
	if runtime.GOOS != "linux" {
		return
	}

	// Use the actual base directory where Prism executable is located
	actualPrismDir := getPrismBaseDir(prismDir)

	logf("=== Qt Environment Debug ===")
	logf("Original Prism Directory: %s", prismDir)
	logf("Actual Prism Base Directory: %s", actualPrismDir)

	// Check directory structure
	pluginsDir := filepath.Join(actualPrismDir, "plugins")
	libDir := filepath.Join(actualPrismDir, "lib")

	logf("Plugins Directory: %s (exists: %v)", pluginsDir, exists(pluginsDir))
	logf("Lib Directory: %s (exists: %v)", libDir, exists(libDir))

	// List plugin directories if they exist
	if exists(pluginsDir) {
		files, err := os.ReadDir(pluginsDir)
		if err == nil {
			logf("Plugin directories found:")
			for _, file := range files {
				if file.IsDir() {
					logf("  - %s", file.Name())
				}
			}
		}
	}

	// Check for critical plugin files
	criticalPlugins := []string{
		"platforms/libqxcb.so",
		"imageformats/libqjpeg.so",
		"iconengines/libqsvgicon.so",
	}

	for _, plugin := range criticalPlugins {
		pluginPath := filepath.Join(pluginsDir, plugin)
		logf("Plugin %s: %v", plugin, exists(pluginPath))
	}

	logf("=== End Qt Environment Debug ===")
}

// fixQtPluginRPATH fixes RPATH settings in Qt plugins on Linux systems
// This ensures plugins can find the bundled Qt libraries
func fixQtPluginRPATH(prismDir string) error {
	// Only apply this fix on Linux systems
	if runtime.GOOS != "linux" {
		return nil
	}

	logf("%s", stepLine("Fixing Qt plugin RPATH settings"))

	// Check if patchelf is available
	if _, err := exec.LookPath("patchelf"); err != nil {
		logf("%s", warnLine("patchelf not found, skipping RPATH fixing (install patchelf for better Qt compatibility)"))
		return nil // Not an error, just skip the fix
	}

	// Use the actual base directory where Prism executable is located
	actualPrismDir := getPrismBaseDir(prismDir)
	pluginsDir := filepath.Join(actualPrismDir, "plugins")
	libDir := filepath.Join(actualPrismDir, "lib")

	logf("DEBUG: Using plugins directory: %s", pluginsDir)
	logf("DEBUG: Using lib directory: %s", libDir)

	if !exists(pluginsDir) {
		logf("%s", warnLine("Plugins directory not found, skipping RPATH fixing"))
		return nil
	}

	// Fix RPATH for critical Qt plugins
	criticalPlugins := []string{
		"platforms/libqxcb.so",
		"imageformats/libqjpeg.so",
		"imageformats/libqgif.so",
		"imageformats/libqwebp.so",
		"iconengines/libqsvgicon.so",
		"platforms/libqminimal.so",
		"platforms/libqoffscreen.so",
		"platforms/libqvnc.so",
	}

	var fixedCount int
	var errors []string

	for _, plugin := range criticalPlugins {
		pluginPath := filepath.Join(pluginsDir, plugin)
		if exists(pluginPath) {
			// Calculate relative path from plugin to lib directory
			// This handles both flat and nested directory structures
			relativeLibPath := calculateRelativePath(pluginPath, libDir)

			cmd := exec.Command("patchelf", "--set-rpath", relativeLibPath, pluginPath)
			if err := cmd.Run(); err != nil {
				errMsg := fmt.Sprintf("Failed to fix RPATH for %s: %v", plugin, err)
				logf("%s", warnLine(errMsg))
				errors = append(errors, errMsg)
			} else {
				logf("Fixed RPATH for %s (using %s)", plugin, relativeLibPath)
				fixedCount++
			}
		} else {
			logf("Plugin not found: %s (skipping)", plugin)
		}
	}

	// Fix all .so files in plugins directory recursively as a fallback
	err := filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.so files
		if info.IsDir() || !strings.HasSuffix(path, ".so") {
			return nil
		}

		// Skip files we already processed
		for _, plugin := range criticalPlugins {
			if strings.HasSuffix(path, plugin) {
				return nil
			}
		}

		// Calculate relative path from plugin to lib directory
		relativeLibPath := calculateRelativePath(path, libDir)

		// Fix RPATH for additional plugins
		cmd := exec.Command("patchelf", "--set-rpath", relativeLibPath, path)
		if err := cmd.Run(); err != nil {
			logf("Failed to fix RPATH for %s: %v", filepath.Base(path), err)
		} else {
			logf("Fixed RPATH for additional plugin: %s (using %s)", filepath.Base(path), relativeLibPath)
			fixedCount++
		}

		return nil
	})

	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Error during recursive RPATH fixing: %v", err)))
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		logf("%s", warnLine(fmt.Sprintf("RPATH fixing completed with %d errors", len(errors))))
	} else if fixedCount > 0 {
		logf("%s", successLine(fmt.Sprintf("Successfully fixed RPATH for %d Qt plugins", fixedCount)))
	} else {
		logf("%s", warnLine("No Qt plugins found to fix RPATH for"))
	}

	return nil
}

// calculateRelativePath calculates the relative path from a plugin to the lib directory
func calculateRelativePath(pluginPath, libDir string) string {
	// Get the directory containing the plugin
	pluginDir := filepath.Dir(pluginPath)

	// Calculate relative path from plugin directory to lib directory
	relPath, err := filepath.Rel(pluginDir, libDir)
	if err != nil {
		// Fallback to $ORIGIN/../../lib if calculation fails
		logf("DEBUG: Failed to calculate relative path, using fallback: %v", err)
		return "$ORIGIN/../../lib"
	}

	// Convert to $ORIGIN-based relative path for RPATH
	if relPath == "." {
		return "$ORIGIN"
	}

	// Replace backslashes with forward slashes for consistency
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	// Prefix with $ORIGIN for RPATH compatibility
	return "$ORIGIN/" + relPath
}

// -------------------- Launcher Logic --------------------

func runLauncherLogic(root, exePath string, modpack Modpack, prismProcess **os.Process, progressCb func(stage string, step, total int)) {
	packName := modpackLabel(modpack)
	// Note: Update check already happened at startup in main()

	totalSteps := 8
	currentStep := 0
	report := func(stage string) {
		currentStep++
		if progressCb != nil {
			progressCb(stage, currentStep, totalSteps)
		}
	}

	report("Reading modpack configuration")

	// 0) Read pack.toml to get correct Minecraft and modloader versions
	logf("%s", stepLine("Reading modpack configuration"))
	packInfo, err := fetchPackInfo(modpack.PackURL)
	if err != nil {
		fail(fmt.Errorf("failed to read modpack configuration: %w", err))
	}
	logf("%s", successLine(fmt.Sprintf("Detected: Minecraft %s with %s %s", packInfo.Minecraft, packInfo.ModLoader, packInfo.LoaderVersion)))

	// 1) Ensure prerequisites — organize directories cleanly
	prismDir := filepath.Join(root, "prism")
	utilDir := filepath.Join(root, "util")
	prismJavaDir := filepath.Join(prismDir, "java")

	// Determine required Java version based on Minecraft version
	requiredJavaVersion := getJavaVersionForMinecraft(packInfo.Minecraft)
	jreDir := filepath.Join(prismJavaDir, "jre"+requiredJavaVersion)
	javaBin := filepath.Join(jreDir, "bin", JavaBinName)
	javawBin := filepath.Join(jreDir, "bin", JavawBinName)
	bootstrapExe := filepath.Join(utilDir, "packwiz-installer-bootstrap"+getExecutableExtension())
	bootstrapJar := filepath.Join(utilDir, "packwiz-installer-bootstrap.jar")

	// Create util directory for miscellaneous files
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		fail(fmt.Errorf("failed to create util directory: %w", err))
	}

	// Create Prism Java directory for managed Java runtimes
	if err := os.MkdirAll(prismJavaDir, 0755); err != nil {
		fail(fmt.Errorf("failed to create Prism Java directory: %w", err))
	}

	logf("%s", sectionLine("Preparing Environment"))

	report("Ensuring Prism Launcher")
	logf("%s", stepLine("Ensuring Prism Launcher portable build"))

	// Check and install Qt dependencies if needed (Linux only)
	if runtime.GOOS == "linux" {
		logf("%s", stepLine("Checking Qt dependencies"))
		if err := ensureQtDependencies(); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Qt dependency check failed: %v", err)))
			// Don't fail the entire operation, just warn the user
			logf("%s", warnLine("Prism Launcher may fail to start without Qt dependencies"))
		}
	}

	prismDownloaded, err := ensurePrism(prismDir)
	if err != nil {
		fail(err)
	}
	if prismDownloaded {
		logf("%s", successLine("Prism Launcher downloaded"))
	} else {
		logf("%s", successLine("Prism Launcher ready"))
	}

	report("Ensuring Java runtime")
	if !exists(javaBin) || !exists(javawBin) {
		logf("%s", stepLine(fmt.Sprintf("Installing Temurin JRE %s", requiredJavaVersion)))
		jreURL, err := fetchJREURL(requiredJavaVersion)
		if err != nil {
			fail(fmt.Errorf("failed to resolve Java %s download: %w", requiredJavaVersion, err))
		}
		if err := downloadAndUnzipTo(jreURL, jreDir); err != nil {
			fail(err)
		}
		_ = flattenJREExtraction(jreDir)
		if !exists(javaBin) || !exists(javawBin) {
			fail(fmt.Errorf("Java %s installation looks incomplete (bin/%s or bin/%s not found)", requiredJavaVersion, JavaBinName, JavawBinName))
		}
		logf("%s", successLine(fmt.Sprintf("Java %s installed", requiredJavaVersion)))
	} else {
		logf("%s", successLine(fmt.Sprintf("Java %s already installed", requiredJavaVersion)))
	}

	report("Ensuring packwiz bootstrap")
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

	report("Preparing modpack instance")
	instanceConfigFile := filepath.Join(instDir, "instance.cfg")
	mmcPackFile := filepath.Join(instDir, "mmc-pack.json")

	needsInstanceCreation := !exists(instanceConfigFile) || !exists(mmcPackFile)
	if needsInstanceCreation {
		logf("%s", stepLine(fmt.Sprintf("Creating Prism instance structure with %s %s", packInfo.ModLoader, packInfo.LoaderVersion)))
		if err := createMultiMCInstance(modpack, packInfo, instDir, javawBin); err != nil {
			fail(fmt.Errorf("failed to create MultiMC instance: %w", err))
		}
		logf("%s", successLine("Instance structure ready"))
	} else {
		logf("%s", successLine("Instance structure already present"))
	}

	// Check if the modloader is already installed
	var modloaderInstalled bool
	if packInfo.ModLoader == "forge" {
		forgeJar := filepath.Join(mcDir, "libraries", "net", "minecraftforge", "forge", fmt.Sprintf("%s-%s", packInfo.Minecraft, packInfo.LoaderVersion), fmt.Sprintf("forge-%s-%s-universal.jar", packInfo.Minecraft, packInfo.LoaderVersion))
		modloaderInstalled = exists(forgeJar) && exists(mmcPackFile)
	} else {
		// For other modloaders, check mmc-pack.json exists
		modloaderInstalled = exists(mmcPackFile)
	}

	if !modloaderInstalled {
		logf("%s", stepLine(fmt.Sprintf("Installing %s %s", packInfo.ModLoader, packInfo.LoaderVersion)))
		if err := installModLoaderForInstance(instDir, javaBin, packInfo); err != nil {
			fail(fmt.Errorf("failed to install %s: %w", packInfo.ModLoader, err))
		}
		logf("%s", successLine(fmt.Sprintf("%s ready", strings.Title(packInfo.ModLoader))))
	} else {
		logf("%s", successLine(fmt.Sprintf("%s already installed", strings.Title(packInfo.ModLoader))))
	}

	// 6) Check for modpack updates
	logf("%s", sectionLine("Modpack Sync"))
	logf("%s", stepLine("Checking for modpack updates"))
	updateAvailable, localVersion, remoteVersion, err := checkModpackUpdate(modpack, instDir)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to check modpack updates: %v", err)))
		updateAvailable = true
	}

	report("Checking modpack updates")
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

	report("Synchronizing modpack files")
	go func() {
		for range progressTicker.C {
			if updateAvailable {
				logf("%s in progress... (this may take several minutes)", action)
			} else {
				logf("Verifying installation...")
			}
		}
	}()

	// Ensure packwiz-installer.jar is available
	mainJarPath := filepath.Join(utilDir, "packwiz-installer.jar")
	logf("DEBUG: Checking for packwiz-installer.jar at: %s", mainJarPath)
	if !exists(mainJarPath) {
		logf("%s", stepLine("Downloading packwiz-installer.jar"))
		logf("DEBUG: Starting download of packwiz-installer.jar...")
		if err := downloadPackwizInstaller(mainJarPath); err != nil {
			logf("DEBUG: downloadPackwizInstaller failed: %v", err)
			fail(fmt.Errorf("failed to download packwiz-installer.jar: %w", err))
		}
		logf("%s", successLine("packwiz-installer.jar downloaded"))
		logf("DEBUG: packwiz-installer.jar download completed")
	} else {
		logf("DEBUG: packwiz-installer.jar already exists")
	}

	var cmd *exec.Cmd
	if exists(bootstrapExe) {
		cmd = exec.Command(bootstrapExe, "--bootstrap-no-update", "--bootstrap-main-jar", mainJarPath, "-g", packURL) // run from minecraft directory
		logf("DEBUG: Using packwiz bootstrap EXE: %s", bootstrapExe)
	} else if exists(bootstrapJar) {
		cmd = exec.Command(javaBin, "-jar", bootstrapJar, "--bootstrap-no-update", "--bootstrap-main-jar", mainJarPath, "-g", packURL)
		logf("DEBUG: Using packwiz bootstrap JAR: %s", bootstrapJar)
		logf("DEBUG: Java binary: %s", javaBin)
	} else {
		fail(errors.New("packwiz bootstrap not found after download"))
	}
	cmd.Dir = mcDir // critical: minecraft directory so packwiz installs mods in correct place
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+jreDir,
		"PATH="+BuildPathEnv(filepath.Join(jreDir, "bin")),
	)

	logf("DEBUG: Packwiz command: %s", cmd.String())
	logf("DEBUG: Packwiz working directory: %s", cmd.Dir)
	logf("DEBUG: Packwiz environment: JAVA_HOME=%s", jreDir)

	// Set platform-specific process attributes
	setPackwizProcessAttributes(cmd)

	var buf bytes.Buffer
	mw := io.MultiWriter(out, &buf)
	cmd.Stdout, cmd.Stderr = mw, mw

	logf("DEBUG: Starting packwiz execution...")
	progressTicker.Stop() // Stop progress ticker before running packwiz
	err = cmd.Run()
	logf("DEBUG: Packwiz execution completed with error: %v", err)
	if err != nil {
		// Parse packwiz output for manual-download instructions
		items := parsePackwizManuals(buf.String())
		if len(items) > 0 {
			assistManualFromPackwiz(items)
			// Retry ONCE after user saves files, but create a new command to avoid "already started" error
			var retryCmd *exec.Cmd
			if exists(bootstrapExe) {
				retryCmd = exec.Command(bootstrapExe, "--bootstrap-no-update", "--bootstrap-main-jar", mainJarPath, "-g", packURL)
			} else if exists(bootstrapJar) {
				retryCmd = exec.Command(javaBin, "-jar", bootstrapJar, "--bootstrap-no-update", "--bootstrap-main-jar", mainJarPath, "-g", packURL)
			}
			if retryCmd != nil {
				retryCmd.Dir = mcDir // also run from minecraft directory
				retryCmd.Env = append(os.Environ(),
					"JAVA_HOME="+jreDir,
					"PATH="+BuildPathEnv(filepath.Join(jreDir, "bin")),
				)

				// Set platform-specific process attributes for retry
				setPackwizRetryProcessAttributes(retryCmd)

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

	report("Launching via Prism")

	// Update global JavaPath in prismlauncher.cfg for this modpack
	logf("%s", stepLine("Updating Prism Java configuration"))
	if err := updatePrismJavaPath(prismDir, javawBin); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to update Prism Java path: %v", err)))
	}

	prismExe := GetPrismExecutablePath(prismDir)

	// On macOS, check if Prism exists in the local directory, otherwise use /Applications
	if runtime.GOOS == "darwin" && !exists(prismExe) {
		applicationsPrism := filepath.Join("/Applications", "Prism Launcher.app", "Contents", "MacOS", "PrismLauncher")
		if exists(applicationsPrism) {
			prismExe = applicationsPrism
			logf("Using Prism Launcher from /Applications folder")
		} else {
			logf("Warning: Prism Launcher not found at %s or %s", prismExe, applicationsPrism)
		}
	}

	logf("DEBUG: Using Prism executable: %s", prismExe)
	logf("DEBUG: Working directory: %s", prismDir)

	// Log Qt environment setup for debugging
	logQtEnvironment(prismDir)

	// Launch the instance directly (this should not show the Prism GUI)
	launch := exec.Command(prismExe, "--dir", ".", "--launch", modpack.InstanceName)
	launch.Dir = prismDir

	// Build Qt environment variables
	qtEnv := buildQtEnvironment(prismDir, jreDir)
	launch.Env = append(os.Environ(), qtEnv...)

	// Log environment variables for debugging
	if runtime.GOOS == "linux" {
		for _, env := range launch.Env {
			if strings.Contains(env, "QT_") || strings.Contains(env, "LD_LIBRARY_PATH") {
				logf("DEBUG: Environment variable: %s", env)
			}
		}
	}

	launch.Stdout, launch.Stderr = out, out

	// Start the process and wait for it to complete (keeps console open)
	if err := launch.Start(); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to launch %s: %v", packName, err)))
		logf("%s", stepLine("Opening Prism Launcher UI instead"))
		launchFallback := exec.Command(prismExe, "--dir", ".")
		launchFallback.Dir = prismDir

		// Use the same Qt environment setup for fallback
		qtEnv := buildQtEnvironment(prismDir, jreDir)
		launchFallback.Env = append(os.Environ(), qtEnv...)

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
