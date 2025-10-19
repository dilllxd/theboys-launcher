package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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
		_ = flattenOneLevel(jreDir)
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
		"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
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
					"PATH="+filepath.Join(jreDir, "bin")+";"+os.Getenv("PATH"),
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

	prismExe := filepath.Join(prismDir, PrismExeName)
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
