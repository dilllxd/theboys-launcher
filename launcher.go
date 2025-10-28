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

		// Qt debug variables for comprehensive logging
		qtEnv = append(qtEnv, "QT_DEBUG_PLUGINS=1")         // Enable detailed plugin loading information
		qtEnv = append(qtEnv, "QT_LOGGING_RULES*=true")     // Enable comprehensive logging
		qtEnv = append(qtEnv, "QT_DEBUG_PLUGINS_VERBOSE=1") // More verbose plugin debugging
		qtEnv = append(qtEnv, "QT_QPA_VERBOSE=1")           // QPA platform debugging
		qtEnv = append(qtEnv, "QT_XCB_DEBUG=1")             // XCB backend debugging

		logf("DEBUG: Enabled Qt debug variables for troubleshooting")
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

	logf("%s", stepLine("Fixing Qt plugin RPATH settings and permissions"))

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

	// First, fix permissions for all .so files in plugins directory
	logf("%s", stepLine("Fixing plugin permissions"))
	if err := fixPluginPermissions(pluginsDir); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to fix plugin permissions: %v", err)))
		// Don't fail the entire operation, just log the warning
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

				// Verify RPATH was set correctly
				if err := verifyRPATH(pluginPath, relativeLibPath); err != nil {
					errMsg := fmt.Sprintf("RPATH verification failed for %s: %v", plugin, err)
					logf("%s", warnLine(errMsg))
					errors = append(errors, errMsg)
				} else {
					logf("Verified RPATH for %s", plugin)
				}
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

			// Verify RPATH was set correctly
			if err := verifyRPATH(path, relativeLibPath); err != nil {
				logf("RPATH verification failed for %s: %v", filepath.Base(path), err)
			} else {
				logf("Verified RPATH for %s", filepath.Base(path))
			}
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

// fixPluginPermissions fixes permissions for all .so files in the plugins directory
func fixPluginPermissions(pluginsDir string) error {
	return filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.so files
		if info.IsDir() || !strings.HasSuffix(path, ".so") {
			return nil
		}

		// Set execute permissions (755) for plugin files
		currentMode := info.Mode()
		newMode := currentMode | 0111 // Add execute bit for owner, group, and others

		if newMode != currentMode {
			if err := os.Chmod(path, newMode); err != nil {
				logf("Failed to set permissions for %s: %v", filepath.Base(path), err)
				return err
			}
			logf("Fixed permissions for %s", filepath.Base(path))
		}

		return nil
	})
}

// verifyRPATH verifies that the RPATH was set correctly for a plugin
func verifyRPATH(pluginPath, expectedRPATH string) error {
	// Use patchelf to check the current RPATH
	cmd := exec.Command("patchelf", "--print-rpath", pluginPath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read RPATH: %w", err)
	}

	actualRPATH := strings.TrimSpace(string(output))

	// For $ORIGIN-based paths, we need to check if the expected pattern is contained
	if strings.HasPrefix(expectedRPATH, "$ORIGIN") {
		if !strings.Contains(actualRPATH, expectedRPATH) {
			return fmt.Errorf("RPATH mismatch: expected to contain '%s', got '%s'", expectedRPATH, actualRPATH)
		}
	} else {
		// For absolute paths, check exact match
		if actualRPATH != expectedRPATH {
			return fmt.Errorf("RPATH mismatch: expected '%s', got '%s'", expectedRPATH, actualRPATH)
		}
	}

	return nil
}

// isSharedLibrary checks if a file is a valid shared library using the file command
func isSharedLibrary(filePath string) bool {
	// Check if file command is available
	if _, err := exec.LookPath("file"); err != nil {
		logf("%s", warnLine("file command not found, skipping library type checking"))
		return true // Assume it's valid if we can't check
	}

	cmd := exec.Command("file", filePath)
	output, err := cmd.Output()
	if err != nil {
		logf("Failed to check file type for %s: %v", filepath.Base(filePath), err)
		return false
	}

	outputStr := strings.ToLower(string(output))
	// Check for shared library indicators
	return strings.Contains(outputStr, "shared object") ||
		strings.Contains(outputStr, "pie executable") ||
		strings.Contains(outputStr, "dynamically linked")
}

// checkPluginDependencies checks if plugins are valid shared libraries and can find their dependencies
func checkPluginDependencies(prismDir string) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	logf("%s", stepLine("Checking plugin dependencies"))

	// Use the actual base directory where Prism executable is located
	actualPrismDir := getPrismBaseDir(prismDir)
	pluginsDir := filepath.Join(actualPrismDir, "plugins")

	if !exists(pluginsDir) {
		logf("%s", warnLine("Plugins directory not found, skipping dependency checking"))
		return nil
	}

	// Check critical plugins
	criticalPlugins := []string{
		"platforms/libqxcb.so",
		"imageformats/libqjpeg.so",
		"iconengines/libqsvgicon.so",
	}

	var invalidPlugins []string
	var missingDeps []string
	var checkedPlugins int
	var validPlugins int

	// Check if ldd is available for dependency checking
	lddAvailable := true
	if _, err := exec.LookPath("ldd"); err != nil {
		logf("%s", warnLine("ldd not found, will only validate plugin file types"))
		lddAvailable = false
	}

	for _, plugin := range criticalPlugins {
		pluginPath := filepath.Join(pluginsDir, plugin)
		if exists(pluginPath) {
			checkedPlugins++
			logf("Checking plugin: %s", plugin)

			// First validate that it's a proper shared library
			if !isSharedLibrary(pluginPath) {
				errMsg := fmt.Sprintf("%s: not a valid shared library (may be corrupted or wrong file type)", plugin)
				invalidPlugins = append(invalidPlugins, errMsg)
				logf("%s", warnLine(errMsg))
				continue
			}

			validPlugins++
			logf("Plugin %s: valid shared library", plugin)

			// Only check dependencies if ldd is available
			if lddAvailable {
				logf("Checking dependencies for %s", plugin)
				cmd := exec.Command("ldd", pluginPath)
				output, err := cmd.Output()
				if err != nil {
					logf("%s", warnLine(fmt.Sprintf("Failed to check dependencies for %s: %v", plugin, err)))
					continue
				}

				// Parse ldd output for missing dependencies
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "not found") {
						// Extract the library name
						parts := strings.Fields(line)
						if len(parts) > 0 {
							missingLib := parts[0]
							missingDeps = append(missingDeps, fmt.Sprintf("%s: %s", plugin, missingLib))
							logf("%s", warnLine(fmt.Sprintf("Missing dependency for %s: %s", plugin, missingLib)))
						}
					}
				}
			}
		} else {
			logf("Plugin not found: %s (skipping dependency check)", plugin)
		}
	}

	// Check all .so files in plugins directory recursively
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

		checkedPlugins++
		pluginName := filepath.Base(path)
		logf("Checking additional plugin: %s", pluginName)

		// First validate that it's a proper shared library
		if !isSharedLibrary(path) {
			errMsg := fmt.Sprintf("%s: not a valid shared library (may be corrupted or wrong file type)", pluginName)
			invalidPlugins = append(invalidPlugins, errMsg)
			logf("%s", warnLine(errMsg))
			return nil
		}

		validPlugins++
		logf("Plugin %s: valid shared library", pluginName)

		// Only check dependencies if ldd is available
		if lddAvailable {
			logf("Checking dependencies for %s", pluginName)
			cmd := exec.Command("ldd", path)
			output, err := cmd.Output()
			if err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to check dependencies for %s: %v", pluginName, err)))
				return nil
			}

			// Parse ldd output for missing dependencies
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "not found") {
					// Extract the library name
					parts := strings.Fields(line)
					if len(parts) > 0 {
						missingLib := parts[0]
						missingDeps = append(missingDeps, fmt.Sprintf("%s: %s", pluginName, missingLib))
						logf("%s", warnLine(fmt.Sprintf("Missing dependency for %s: %s", pluginName, missingLib)))
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Error during recursive plugin checking: %v", err)))
	}

	// Report results
	if len(invalidPlugins) > 0 {
		logf("%s", warnLine(fmt.Sprintf("Found %d invalid/corrupted plugins:", len(invalidPlugins))))
		for _, plugin := range invalidPlugins {
			logf("  - %s", plugin)
		}
		logf("%s", warnLine("Invalid plugins may need to be reinstalled or extracted from the Prism Launcher archive"))
	}

	if len(missingDeps) > 0 {
		logf("%s", warnLine(fmt.Sprintf("Found %d missing dependencies across %d plugins", len(missingDeps), validPlugins)))
		logf("Missing dependencies:")
		for _, dep := range missingDeps {
			logf("  - %s", dep)
		}
		return fmt.Errorf("plugin dependencies missing: %s", strings.Join(missingDeps, "; "))
	} else if validPlugins > 0 {
		logf("%s", successLine(fmt.Sprintf("All %d plugins validated with resolved dependencies", validPlugins)))
	} else {
		logf("%s", warnLine("No valid plugins found"))
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

// -------------------- Error Analysis Functions --------------------

// analyzePrismError analyzes error output from Prism to identify common issues
func analyzePrismError(stderrStr, stdoutStr string) []string {
	var issues []string
	combinedOutput := stderrStr + "\n" + stdoutStr

	// Log the raw error output for debugging
	logf("=== Prism Error Analysis ===")
	logf("Raw stderr output:")
	logf("%s", stderrStr)
	if stdoutStr != "" {
		logf("Raw stdout output:")
		logf("%s", stdoutStr)
	}

	// Check for unusual colon format that might indicate Prism-specific errors
	if strings.Contains(combinedOutput, ":") && !strings.Contains(combinedOutput, " ") {
		lines := strings.Split(combinedOutput, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, ":") && !strings.Contains(line, " ") {
				logf("%s", warnLine(fmt.Sprintf("Detected unusual colon format error: %s", line)))
				issues = append(issues, fmt.Sprintf("Unusual error format: %s", line))
			}
		}
	}

	// Check for common Qt library issues
	if strings.Contains(combinedOutput, "cannot open shared object file") {
		issues = append(issues, "Missing shared library - Qt dependencies may be incomplete")
		logf("%s", warnLine("Detected missing shared library error"))
	}

	// Check for Qt platform plugin issues
	if strings.Contains(combinedOutput, "Could not load the Qt platform plugin") {
		issues = append(issues, "Qt platform plugin loading failed - check plugin permissions and RPATH")
		logf("%s", warnLine("Detected Qt platform plugin issue"))
	}

	// Check for permission issues
	if strings.Contains(combinedOutput, "Permission denied") {
		issues = append(issues, "Permission denied - check file and directory permissions")
		logf("%s", warnLine("Detected permission issue"))
	}

	// Check for GL/GLX issues
	if strings.Contains(combinedOutput, "GLX") || strings.Contains(combinedOutput, "OpenGL") {
		issues = append(issues, "Graphics/GLX issue - may need graphics drivers or different Qt platform")
		logf("%s", warnLine("Detected graphics/GLX issue"))
	}

	// Check for Java-related issues
	if strings.Contains(combinedOutput, "JAVA_HOME") || strings.Contains(combinedOutput, "java.lang") {
		issues = append(issues, "Java configuration issue - check JAVA_HOME and Java path")
		logf("%s", warnLine("Detected Java configuration issue"))
	}

	// Check for patchelf-related issues
	if strings.Contains(combinedOutput, "patchelf") || strings.Contains(combinedOutput, "RPATH") {
		issues = append(issues, "RPATH/library linking issue - patchelf may have failed")
		logf("%s", warnLine("Detected RPATH/library linking issue"))
	}

	logf("=== End Prism Error Analysis ===")
	return issues
}

// provideErrorContext provides user-friendly error context based on identified issues
func provideErrorContext(issues []string) {
	if len(issues) == 0 {
		logf("%s", infoLine("No specific issues identified in error output"))
		return
	}

	logf("%s", sectionLine("Error Analysis Results"))
	logf("%s", warnLine(fmt.Sprintf("Identified %d issue(s):", len(issues))))

	for i, issue := range issues {
		logf("%d. %s", i+1, issue)
	}

	logf("%s", sectionLine("Recommended Solutions"))

	for _, issue := range issues {
		switch {
		case strings.Contains(issue, "Missing shared library"):
			logf("%s", infoLine("• Run: sudo apt install libqt6core6t64 libqt6gui6 libqt6widgets6 libqt6network6 libqt6svg6"))
			logf("%s", infoLine("• Ensure patchelf is installed: sudo apt install patchelf"))
		case strings.Contains(issue, "Qt platform plugin"):
			logf("%s", infoLine("• Check plugin permissions in the plugins directory"))
			logf("%s", infoLine("• Verify RPATH settings with: readelf -d plugins/platforms/libqxcb.so"))
		case strings.Contains(issue, "Permission denied"):
			logf("%s", infoLine("• Fix permissions: chmod +x plugins/**/*.so"))
			logf("%s", infoLine("• Check directory ownership: ls -la prism/"))
		case strings.Contains(issue, "Graphics/GLX"):
			logf("%s", infoLine("• Try different Qt platform: export QT_QPA_PLATFORM=wayland"))
			logf("%s", infoLine("• Update graphics drivers"))
		case strings.Contains(issue, "Java configuration"):
			logf("%s", infoLine("• Verify Java installation: java -version"))
			logf("%s", infoLine("• Check JAVA_HOME is set correctly"))
		case strings.Contains(issue, "RPATH/library linking"):
			logf("%s", infoLine("• Reinstall patchelf: sudo apt install --reinstall patchelf"))
			logf("%s", infoLine("• Manually fix RPATH: patchelf --set-rpath '$ORIGIN/../lib' plugins/**/*.so"))
		case strings.Contains(issue, "Unusual error format"):
			logf("%s", infoLine("• This may be a Prism Launcher internal error"))
			logf("%s", infoLine("• Try launching Prism GUI directly for more details"))
		}
	}
}

// createPrismWrapperScript creates a wrapper script for launching Prism with proper environment
func createPrismWrapperScript(prismDir, jreDir string) (string, error) {
	if runtime.GOOS != "linux" {
		return "", nil // Only needed on Linux
	}

	logf("%s", stepLine("Creating Prism wrapper script"))

	// Create wrapper script path
	wrapperPath := filepath.Join(prismDir, "launch-prism-wrapper.sh")

	// Get the actual Prism executable path
	prismExe := GetPrismExecutablePath(prismDir)

	// Build the wrapper script content
	wrapperContent := fmt.Sprintf(`#!/bin/bash
# Prism Launcher Wrapper Script
# This script ensures proper environment setup for Prism Launcher

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRISM_EXE="%s"
JRE_DIR="%s"

# Log function for debugging
log_debug() {
	   echo "[WRAPPER DEBUG] $1" >&2
}

log_debug "Starting Prism wrapper script"
log_debug "Script directory: $SCRIPT_DIR"
log_debug "Prism executable: $PRISM_EXE"
log_debug "JRE directory: $JRE_DIR"

# Ensure patchelf is installed
if ! command -v patchelf &> /dev/null; then
	   echo "[WRAPPER ERROR] patchelf is not installed" >&2
	   echo "[WRAPPER INFO] Attempting to install patchelf..." >&2
	   
	   # Try to install patchelf
	   if command -v apt &> /dev/null; then
	       sudo apt update && sudo apt install -y patchelf
	   elif command -v dnf &> /dev/null; then
	       sudo dnf install -y patchelf
	   elif command -v yum &> /dev/null; then
	       sudo yum install -y patchelf
	   elif command -v pacman &> /dev/null; then
	       sudo pacman -S --noconfirm patchelf
	   else
	       echo "[WRAPPER ERROR] Cannot install patchelf - no supported package manager found" >&2
	       exit 1
	   fi
	   
	   # Verify installation
	   if ! command -v patchelf &> /dev/null; then
	       echo "[WRAPPER ERROR] patchelf installation failed" >&2
	       exit 1
	   fi
	   
	   echo "[WRAPPER SUCCESS] patchelf installed successfully" >&2
fi

# Fix RPATH for Qt plugins if needed
PLUGINS_DIR="$SCRIPT_DIR/plugins"
LIB_DIR="$SCRIPT_DIR/lib"

if [ -d "$PLUGINS_DIR" ]; then
	   log_debug "Fixing RPATH for Qt plugins"
	   
	   # Fix critical plugins
	   for plugin in platforms/libqxcb.so imageformats/libqjpeg.so iconengines/libqsvgicon.so; do
	       plugin_path="$PLUGINS_DIR/$plugin"
	       if [ -f "$plugin_path" ]; then
	           log_debug "Fixing RPATH for $plugin"
	           # Calculate relative path to lib directory
	           rel_path="$(dirname "$plugin_path" | sed "s|$SCRIPT_DIR||" | sed 's|^/||' | sed 's|[^/][^/]*|..|g')/lib"
	           if [ "$rel_path" = "/lib" ]; then
	               rel_path="$LIB_DIR"
	           fi
	           
	           # Set RPATH
	           if patchelf --set-rpath "\$ORIGIN/$rel_path" "$plugin_path" 2>/dev/null; then
	               log_debug "Successfully fixed RPATH for $plugin"
	           else
	               log_debug "Failed to fix RPATH for $plugin (may not be an ELF file)"
	           fi
	       fi
	   done
fi

# Set environment variables
export JAVA_HOME="$JRE_DIR"
export PATH="$JRE_DIR/bin:$PATH"

# Qt environment variables
if [ -d "$PLUGINS_DIR" ]; then
	   export QT_PLUGIN_PATH="$PLUGINS_DIR"
	   log_debug "Set QT_PLUGIN_PATH=$QT_PLUGIN_PATH"
fi

if [ -d "$LIB_DIR" ]; then
	   export LD_LIBRARY_PATH="$LIB_DIR${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
	   log_debug "Set LD_LIBRARY_PATH=$LD_LIBRARY_PATH"
fi

# Qt debug variables
export QT_DEBUG_PLUGINS=1
export QT_LOGGING_RULES="*=true"
export QT_QPA_PLATFORM=xcb
export QT_XCB_GL_INTEGRATION=xcb_glx

log_debug "Environment variables set"
log_debug "Launching Prism: $PRISM_EXE $*"

# Launch Prism with all arguments passed through
exec "$PRISM_EXE" "$@"
`, prismExe, jreDir)

	// Write the wrapper script
	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {
		return "", fmt.Errorf("failed to create wrapper script: %w", err)
	}

	logf("%s", successLine(fmt.Sprintf("Created wrapper script: %s", wrapperPath)))
	return wrapperPath, nil
}

// launchPrismWithWrapper launches Prism using the wrapper script approach
func launchPrismWithWrapper(prismDir, jreDir, instanceName string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("wrapper script approach only supported on Linux")
	}

	logf("%s", stepLine("Launching Prism using wrapper script approach"))

	// Create wrapper script
	wrapperPath, err := createPrismWrapperScript(prismDir, jreDir)
	if err != nil {
		return fmt.Errorf("failed to create wrapper script: %w", err)
	}

	// Launch using the wrapper script
	var cmd *exec.Cmd
	if instanceName != "" {
		cmd = exec.Command(wrapperPath, "--dir", ".", "--launch", instanceName)
	} else {
		cmd = exec.Command(wrapperPath, "--dir", ".")
	}

	cmd.Dir = prismDir
	cmd.Env = append(os.Environ(), buildQtEnvironment(prismDir, jreDir)...)

	// Capture output for error analysis
	var stdoutBuf, stderrBuf bytes.Buffer
	multiWriter := io.MultiWriter(out, &stdoutBuf)
	multiErrWriter := io.MultiWriter(out, &stderrBuf)

	cmd.Stdout = multiWriter
	cmd.Stderr = multiErrWriter

	logf("DEBUG: Starting wrapper launch with command: %s", cmd.String())

	// Start the process
	if err := cmd.Start(); err != nil {
		// Analyze the error output
		stderrStr := stderrBuf.String()
		stdoutStr := stdoutBuf.String()
		issues := analyzePrismError(stderrStr, stdoutStr)
		provideErrorContext(issues)

		return fmt.Errorf("failed to launch Prism with wrapper: %w", err)
	}

	logf("%s", successLine(fmt.Sprintf("Prism launched via wrapper (PID: %d)", cmd.Process.Pid)))

	// Wait for the process to complete
	err = cmd.Wait()

	// Analyze output even if the process completed
	stderrStr := stderrBuf.String()
	stdoutStr := stdoutBuf.String()
	if err != nil || stderrStr != "" {
		issues := analyzePrismError(stderrStr, stdoutStr)
		if len(issues) > 0 {
			provideErrorContext(issues)
		}
	}

	return err
}

// launchPrismDirect launches Prism directly with enhanced error handling
func launchPrismDirect(prismExe, prismDir, jreDir, instanceName, packName string, prismProcess **os.Process) error {
	logf("%s", stepLine("Attempting direct Prism launch"))

	// Launch the instance directly (this should not show the Prism GUI)
	launch := exec.Command(prismExe, "--dir", ".", "--launch", instanceName)
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

	// Capture both stdout and stderr for better error reporting
	var stdoutBuf, stderrBuf bytes.Buffer
	multiWriter := io.MultiWriter(out, &stdoutBuf)
	multiErrWriter := io.MultiWriter(out, &stderrBuf)

	launch.Stdout = multiWriter
	launch.Stderr = multiErrWriter

	logf("DEBUG: Starting Prism launch with command: %s", launch.String())

	// Start the process and wait for it to complete (keeps console open)
	if err := launch.Start(); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to launch %s: %v", packName, err)))

		// Log captured output for debugging
		if stdoutBuf.Len() > 0 {
			logf("DEBUG: Captured stdout before failure:\n%s", stdoutBuf.String())
		}
		if stderrBuf.Len() > 0 {
			logf("DEBUG: Captured stderr before failure:\n%s", stderrBuf.String())
		}

		// Analyze the error output
		stderrStr := stderrBuf.String()
		stdoutStr := stdoutBuf.String()
		issues := analyzePrismError(stderrStr, stdoutStr)
		provideErrorContext(issues)

		return fmt.Errorf("direct launch failed: %w", err)
	}

	// Store the process reference for signal handling
	*prismProcess = launch.Process
	logf("%s", successLine(fmt.Sprintf("%s launched (PID: %d)", packName, launch.Process.Pid)))

	// Wait for the game process to complete
	err := launch.Wait()

	// Log completion output for debugging
	if stdoutBuf.Len() > 0 {
		logf("DEBUG: Process stdout:\n%s", stdoutBuf.String())
	}
	if stderrBuf.Len() > 0 {
		logf("DEBUG: Process stderr:\n%s", stderrBuf.String())
	}

	// Check if the process exited with an error
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Prism process exited with error: %v", err)))

		// Analyze the output for common issues using our new analysis function
		stderrStr := stderrBuf.String()
		stdoutStr := stdoutBuf.String()
		issues := analyzePrismError(stderrStr, stdoutStr)

		// Provide user-friendly error context and solutions
		provideErrorContext(issues)

		return fmt.Errorf("Prism process exited with error: %w", err)
	}

	return nil
}

// launchPrismGUIFallback launches Prism GUI as a fallback
func launchPrismGUIFallback(prismExe, prismDir, jreDir, packName string, prismProcess **os.Process) error {
	logf("%s", stepLine("Opening Prism Launcher UI instead"))
	launchFallback := exec.Command(prismExe, "--dir", ".")
	launchFallback.Dir = prismDir

	// Use the same Qt environment setup for fallback
	qtEnv := buildQtEnvironment(prismDir, jreDir)
	launchFallback.Env = append(os.Environ(), qtEnv...)

	// Capture fallback output as well
	var fallbackStdout, fallbackStderr bytes.Buffer
	launchFallback.Stdout = io.MultiWriter(out, &fallbackStdout)
	launchFallback.Stderr = io.MultiWriter(out, &fallbackStderr)

	if err := launchFallback.Start(); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to open Prism Launcher UI: %v", err)))

		// Log captured fallback output for debugging
		if fallbackStdout.Len() > 0 {
			logf("DEBUG: Captured fallback stdout:\n%s", fallbackStdout.String())
		}
		if fallbackStderr.Len() > 0 {
			logf("DEBUG: Captured fallback stderr:\n%s", fallbackStderr.String())
		}

		// Analyze the error output
		stderrStr := fallbackStderr.String()
		stdoutStr := fallbackStdout.String()
		issues := analyzePrismError(stderrStr, stdoutStr)
		provideErrorContext(issues)

		return fmt.Errorf("GUI fallback failed: %w", err)
	}

	*prismProcess = launchFallback.Process
	logf("%s", successLine(fmt.Sprintf("Prism Launcher UI launched for %s (PID: %d)", packName, launchFallback.Process.Pid)))

	// Wait for the GUI process to complete
	err := launchFallback.Wait()

	// Log fallback completion output for debugging
	if fallbackStdout.Len() > 0 {
		logf("DEBUG: Fallback process stdout:\n%s", fallbackStdout.String())
	}
	if fallbackStderr.Len() > 0 {
		logf("DEBUG: Fallback process stderr:\n%s", fallbackStderr.String())
	}

	// Analyze output even if process completed successfully
	if err != nil || fallbackStderr.Len() > 0 {
		stderrStr := fallbackStderr.String()
		stdoutStr := fallbackStdout.String()
		issues := analyzePrismError(stderrStr, stdoutStr)
		if len(issues) > 0 {
			provideErrorContext(issues)
		}
	}

	return err
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

	// Ensure patchelf is installed before attempting RPATH fixes
	if runtime.GOOS == "linux" {
		logf("%s", stepLine("Ensuring patchelf is installed"))
		if err := ensurePatchelfInstalled(); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to ensure patchelf is installed: %v", err)))
			logf("%s", warnLine("RPATH fixing may not work properly without patchelf"))
		}
	}

	// Check plugin dependencies before launching
	if runtime.GOOS == "linux" {
		if err := checkPluginDependencies(prismDir); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Plugin dependency check failed: %v", err)))
			// Don't fail the launch, but warn the user
		}
	}

	// Try multiple launch approaches with fallbacks
	var launchErr error

	// Approach 1: Direct launch with enhanced error handling
	launchErr = launchPrismDirect(prismExe, prismDir, jreDir, modpack.InstanceName, packName, prismProcess)

	if launchErr != nil {
		logf("%s", warnLine(fmt.Sprintf("Direct launch failed: %v", launchErr)))

		// Approach 2: Wrapper script approach (Linux only)
		if runtime.GOOS == "linux" {
			logf("%s", stepLine("Attempting wrapper script launch"))
			launchErr = launchPrismWithWrapper(prismDir, jreDir, modpack.InstanceName)
			if launchErr != nil {
				logf("%s", warnLine(fmt.Sprintf("Wrapper script launch failed: %v", launchErr)))
			} else {
				logf("%s", successLine("Prism launched successfully via wrapper script"))
				return
			}
		}

		// Approach 3: Fallback to GUI launch
		logf("%s", stepLine("Attempting fallback to Prism GUI"))
		launchErr = launchPrismGUIFallback(prismExe, prismDir, jreDir, packName, prismProcess)
		if launchErr != nil {
			logf("%s", warnLine(fmt.Sprintf("GUI fallback launch failed: %v", launchErr)))
			logf("%s", warnLine("All launch attempts failed"))
		} else {
			logf("%s", successLine("Prism launched successfully via GUI fallback"))
		}
	} else {
		logf("%s", successLine("Prism launched successfully via direct launch"))
	}

	logf("%s", successLine(fmt.Sprintf("Prism Launcher closed for %s", packName)))
}
