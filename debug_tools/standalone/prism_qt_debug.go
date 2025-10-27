package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run prism_qt_debug.go <prism_directory>")
		os.Exit(1)
	}

	prismDir := os.Args[1]
	fmt.Printf("=== Prism Launcher Qt Debug Tool ===\n")
	fmt.Printf("Prism Directory: %s\n\n", prismDir)

	// 1. Check if environment variables are being set correctly
	fmt.Println("1. Environment Variable Analysis:")
	fmt.Println("================================")

	// Simulate the buildQtEnvironment function
	jreDir := filepath.Join(prismDir, "java", "jre17") // Assuming JRE 17
	qtEnv := buildQtEnvironmentDebug(prismDir, jreDir)

	fmt.Println("Qt environment variables that would be set:")
	for _, env := range qtEnv {
		if strings.Contains(env, "QT_") || strings.Contains(env, "LD_LIBRARY_PATH") {
			fmt.Printf("  %s\n", env)
		}
	}
	fmt.Println()

	// 2. Verify Qt plugin files exist at expected paths
	fmt.Println("2. Qt Plugin File Verification:")
	fmt.Println("===============================")

	pluginsDir := filepath.Join(prismDir, "plugins")
	if !fileExists(pluginsDir) {
		fmt.Printf("❌ Plugins directory does not exist: %s\n", pluginsDir)
	} else {
		fmt.Printf("✅ Plugins directory exists: %s\n", pluginsDir)

		// Check for critical plugin directories
		criticalDirs := []string{"platforms", "imageformats", "iconengines", "tls"}
		for _, dir := range criticalDirs {
			dirPath := filepath.Join(pluginsDir, dir)
			if fileExists(dirPath) {
				fmt.Printf("✅ Plugin directory exists: %s\n", dir)
			} else {
				fmt.Printf("❌ Plugin directory missing: %s\n", dir)
			}
		}

		// Check for critical plugin files
		criticalPlugins := []string{
			"platforms/libqxcb.so",
			"imageformats/libqjpeg.so",
			"iconengines/libqsvgicon.so",
			"tls/libqopensslbackend.so",
		}

		fmt.Println("\nCritical plugin files:")
		for _, plugin := range criticalPlugins {
			pluginPath := filepath.Join(pluginsDir, plugin)
			if fileExists(pluginPath) {
				fmt.Printf("✅ %s\n", plugin)
			} else {
				fmt.Printf("❌ %s\n", plugin)
			}
		}
	}
	fmt.Println()

	// 3. Check library directory
	fmt.Println("3. Library Directory Analysis:")
	fmt.Println("===============================")

	libDir := filepath.Join(prismDir, "lib")
	if !fileExists(libDir) {
		fmt.Printf("❌ Library directory does not exist: %s\n", libDir)
	} else {
		fmt.Printf("✅ Library directory exists: %s\n", libDir)

		// Check for Qt libraries
		qtLibs, err := filepath.Glob(filepath.Join(libDir, "libQt*.so*"))
		if err != nil {
			fmt.Printf("❌ Error scanning for Qt libraries: %v\n", err)
		} else if len(qtLibs) == 0 {
			fmt.Printf("❌ No Qt libraries found in %s\n", libDir)
		} else {
			fmt.Printf("✅ Found %d Qt libraries:\n", len(qtLibs))
			for _, lib := range qtLibs {
				libName := filepath.Base(lib)
				fmt.Printf("  - %s\n", libName)
			}
		}
	}
	fmt.Println()

	// 4. Check Prism executable
	fmt.Println("4. Prism Executable Analysis:")
	fmt.Println("=============================")

	prismExe := filepath.Join(prismDir, "PrismLauncher")
	if runtime.GOOS == "windows" {
		prismExe += ".exe"
	}

	if !fileExists(prismExe) {
		fmt.Printf("❌ Prism executable does not exist: %s\n", prismExe)
	} else {
		fmt.Printf("✅ Prism executable exists: %s\n", prismExe)

		// Check if it's executable
		if runtime.GOOS != "windows" {
			info, err := os.Stat(prismExe)
			if err != nil {
				fmt.Printf("❌ Error checking executable permissions: %v\n", err)
			} else if info.Mode().Perm()&0111 != 0 {
				fmt.Printf("✅ Prism executable has execute permissions\n")
			} else {
				fmt.Printf("❌ Prism executable lacks execute permissions\n")
			}
		}

		// Try to get library dependencies
		if runtime.GOOS == "linux" {
			fmt.Println("\nChecking Prism library dependencies:")
			cmd := exec.Command("ldd", prismExe)
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("❌ Error running ldd: %v\n", err)
			} else {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "Qt") || strings.Contains(line, "not found") {
						fmt.Printf("  %s\n", line)
					}
				}
			}
		}
	}
	fmt.Println()

	// 5. Check for wrapper scripts
	fmt.Println("5. Wrapper Script Analysis:")
	fmt.Println("===========================")

	wrapperScripts := []string{"launch-prism.sh", "prism-launcher.sh", "run-prism.sh"}
	for _, script := range wrapperScripts {
		scriptPath := filepath.Join(prismDir, script)
		if fileExists(scriptPath) {
			fmt.Printf("✅ Found wrapper script: %s\n", script)

			// Check script content
			content, err := os.ReadFile(scriptPath)
			if err != nil {
				fmt.Printf("❌ Error reading script: %v\n", err)
			} else {
				contentStr := string(content)
				if strings.Contains(contentStr, "QT_PLUGIN_PATH") {
					fmt.Printf("  ✅ Script sets QT_PLUGIN_PATH\n")
				} else {
					fmt.Printf("  ❌ Script does not set QT_PLUGIN_PATH\n")
				}
				if strings.Contains(contentStr, "LD_LIBRARY_PATH") {
					fmt.Printf("  ✅ Script sets LD_LIBRARY_PATH\n")
				} else {
					fmt.Printf("  ❌ Script does not set LD_LIBRARY_PATH\n")
				}
			}
		}
	}
	fmt.Println()

	// 6. Check system Qt installation
	fmt.Println("6. System Qt Analysis:")
	fmt.Println("=====================")

	if runtime.GOOS == "linux" {
		// Check for system Qt packages
		commands := []struct {
			cmd  string
			desc string
		}{
			{"dpkg -l | grep -i qt5", "Qt5 packages (dpkg)"},
			{"dpkg -l | grep -i qt6", "Qt6 packages (dpkg)"},
			{"rpm -qa | grep -i qt", "Qt packages (rpm)"},
		}

		for _, command := range commands {
			cmd := exec.Command("sh", "-c", command.cmd)
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("❌ Error checking %s: %v\n", command.desc, err)
			} else if len(strings.TrimSpace(string(output))) == 0 {
				fmt.Printf("ℹ️  No %s found\n", command.desc)
			} else {
				fmt.Printf("✅ Found %s:\n", command.desc)
				lines := strings.Split(string(output), "\n")
				for i, line := range lines {
					if i < 5 { // Limit output
						fmt.Printf("  %s\n", line)
					}
				}
				if len(lines) > 5 {
					fmt.Printf("  ... (%d more lines)\n", len(lines)-5)
				}
			}
		}
	}
	fmt.Println()

	// 7. Test environment variable setting
	fmt.Println("7. Environment Variable Test:")
	fmt.Println("============================")

	// Create a test script to verify environment variables
	testScript := filepath.Join(prismDir, "test_env.sh")
	if runtime.GOOS == "linux" {
		scriptContent := fmt.Sprintf(`#!/bin/bash
echo "=== Environment Test ==="
echo "QT_PLUGIN_PATH: $QT_PLUGIN_PATH"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
echo "JAVA_HOME: $JAVA_HOME"
echo "PATH: $PATH"
echo "========================"

# Test if Qt plugins are accessible
if [ -n "$QT_PLUGIN_PATH" ]; then
    echo "Testing Qt plugin access:"
    ls -la "$QT_PLUGIN_PATH"/platforms/libqxcb.so 2>/dev/null && echo "✅ libqxcb.so found" || echo "❌ libqxcb.so not found"
    ls -la "$QT_PLUGIN_PATH"/imageformats/libqjpeg.so 2>/dev/null && echo "✅ libqjpeg.so found" || echo "❌ libqjpeg.so not found"
else
    echo "❌ QT_PLUGIN_PATH not set"
fi
`, prismDir, jreDir)

		err := os.WriteFile(testScript, []byte(scriptContent), 0755)
		if err != nil {
			fmt.Printf("❌ Error creating test script: %v\n", err)
		} else {
			fmt.Printf("✅ Created test script: %s\n", testScript)

			// Run the test script with the environment
			cmd := exec.Command("bash", testScript)
			cmd.Dir = prismDir
			cmd.Env = append(os.Environ(), qtEnv...)

			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("❌ Error running test script: %v\n", err)
			} else {
				fmt.Printf("Test script output:\n%s\n", string(output))
			}

			// Clean up
			os.Remove(testScript)
		}
	}

	fmt.Println("\n=== Debug Analysis Complete ===")
}

// Standalone version of buildQtEnvironment function
func buildQtEnvironmentDebug(prismDir, jreDir string) []string {
	qtEnv := []string{
		"JAVA_HOME=" + jreDir,
		"PATH=" + buildPathEnv(filepath.Join(jreDir, "bin")),
	}

	// Add Qt-specific environment variables for Linux only
	if runtime.GOOS == "linux" {
		// Set Qt plugin path to bundled plugins directory
		qtPluginPath := filepath.Join(prismDir, "plugins")
		if fileExists(qtPluginPath) {
			qtEnv = append(qtEnv, "QT_PLUGIN_PATH="+qtPluginPath)
		}

		// Set library path to bundled libraries directory
		qtLibPath := filepath.Join(prismDir, "lib")
		if fileExists(qtLibPath) {
			// Prepend to LD_LIBRARY_PATH to prioritize bundled libraries
			existingLdPath := os.Getenv("LD_LIBRARY_PATH")
			if existingLdPath != "" {
				qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath+":"+existingLdPath)
			} else {
				qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath)
			}
		}

		// Additional Qt environment variables for better compatibility
		qtEnv = append(qtEnv, "QT_QPA_PLATFORM=xcb")           // Force X11 backend
		qtEnv = append(qtEnv, "QT_XCB_GL_INTEGRATION=xcb_glx") // OpenGL integration
	}

	return qtEnv
}

// Standalone version of BuildPathEnv function
func buildPathEnv(additionalPath string) string {
	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}
	return additionalPath + separator + os.Getenv("PATH")
}

// Standalone version of exists function
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
