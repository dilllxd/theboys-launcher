package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run prism_portable_structure_debug.go <prism_directory_or_url>")
		fmt.Println("  - If URL is provided, it will download and analyze portable build")
		fmt.Println("  - If directory is provided, it will analyze extracted build")
		os.Exit(1)
	}

	input := os.Args[1]
	fmt.Printf("=== Prism Launcher Portable Structure Debug Tool ===\n")
	fmt.Printf("Input: %s\n\n", input)

	var analysisDir string
	var err error

	// Check if input is a URL or directory
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		// Download and analyze portable build
		analysisDir, err = downloadAndAnalyze(input)
		if err != nil {
			fmt.Printf("Error downloading and analyzing: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(analysisDir) // Clean up temp directory
	} else {
		// Analyze existing directory
		analysisDir = input
		if !exists(analysisDir) {
			fmt.Printf("Directory does not exist: %s\n", analysisDir)
			os.Exit(1)
		}
	}

	// Analyze structure
	analyzePrismStructure(analysisDir)
}

func downloadAndAnalyze(url string) (string, error) {
	fmt.Printf("Downloading Prism Launcher portable build from: %s\n", url)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "prism-debug-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	// Read all data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract based on file extension
	if strings.HasSuffix(strings.ToLower(url), ".tar.gz") || strings.HasSuffix(strings.ToLower(url), ".tgz") {
		err = extractTarGz(data, tempDir)
	} else if strings.HasSuffix(strings.ToLower(url), ".zip") {
		err = extractZip(data, tempDir)
	} else {
		return "", fmt.Errorf("unsupported file format")
	}

	if err != nil {
		return "", fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Printf("Downloaded and extracted to: %s\n", tempDir)
	return tempDir, nil
}

func extractTarGz(data []byte, dest string) error {
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func extractZip(data []byte, dest string) error {
	// This is a simplified zip extraction - in a real implementation you'd use archive/zip
	return fmt.Errorf("zip extraction not implemented in this debug tool")
}

func analyzePrismStructure(dir string) {
	fmt.Printf("=== Analyzing Prism Launcher Structure ===\n")
	fmt.Printf("Directory: %s\n\n", dir)

	// 1. Check top-level structure
	fmt.Println("1. Top-level directory structure:")
	fmt.Println("===============================")

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	var hasPrismExecutable, hasLibDir, hasPluginsDir, hasJavaDir bool
	var topLevelDirs []string

	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			topLevelDirs = append(topLevelDirs, name)
			if name == "lib" {
				hasLibDir = true
			} else if name == "plugins" {
				hasPluginsDir = true
			} else if name == "java" {
				hasJavaDir = true
			}
		} else {
			if strings.HasPrefix(name, "PrismLauncher") || name == "PrismLauncher" {
				if runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe") {
					hasPrismExecutable = true
				} else if runtime.GOOS != "windows" {
					hasPrismExecutable = true
				}
			}
			fmt.Printf("  FILE: %s\n", name)
		}
	}

	fmt.Printf("  Directories: %v\n", topLevelDirs)
	fmt.Printf("  Has Prism executable: %v\n", hasPrismExecutable)
	fmt.Printf("  Has lib directory: %v\n", hasLibDir)
	fmt.Printf("  Has plugins directory: %v\n", hasPluginsDir)
	fmt.Printf("  Has java directory: %v\n", hasJavaDir)
	fmt.Println()

	// 2. Analyze lib directory
	if hasLibDir {
		fmt.Println("2. Library directory analysis:")
		fmt.Println("=============================")

		libDir := filepath.Join(dir, "lib")
		analyzeLibDirectory(libDir)
		fmt.Println()
	}

	// 3. Analyze plugins directory
	if hasPluginsDir {
		fmt.Println("3. Plugins directory analysis:")
		fmt.Println("==============================")

		pluginsDir := filepath.Join(dir, "plugins")
		analyzePluginsDirectory(pluginsDir)
		fmt.Println()
	}

	// 4. Check for wrapper scripts
	fmt.Println("4. Wrapper script analysis:")
	fmt.Println("==========================")

	wrapperScripts := []string{"launch-prism.sh", "prism-launcher.sh", "run-prism.sh", "PrismLauncher.sh"}
	for _, script := range wrapperScripts {
		scriptPath := filepath.Join(dir, script)
		if exists(scriptPath) {
			fmt.Printf("✅ Found wrapper script: %s\n", script)
			analyzeWrapperScript(scriptPath)
		}
	}
	fmt.Println()

	// 5. Check for configuration files
	fmt.Println("5. Configuration file analysis:")
	fmt.Println("==============================")

	configFiles := []string{"prismlauncher.cfg", "prismlauncher.ini"}
	for _, config := range configFiles {
		configPath := filepath.Join(dir, config)
		if exists(configPath) {
			fmt.Printf("✅ Found config file: %s\n", config)
			analyzeConfigFile(configPath)
		}
	}
	fmt.Println()

	// 6. Summary and recommendations
	fmt.Println("6. Summary and Recommendations:")
	fmt.Println("=================================")

	if !hasPrismExecutable {
		fmt.Println("❌ CRITICAL: No Prism Launcher executable found")
	} else {
		fmt.Println("✅ Prism Launcher executable found")
	}

	if !hasLibDir {
		fmt.Println("❌ CRITICAL: No lib directory found - Qt libraries may be missing")
	} else {
		fmt.Println("✅ lib directory found")
	}

	if !hasPluginsDir {
		fmt.Println("❌ CRITICAL: No plugins directory found - Qt plugins may be missing")
	} else {
		fmt.Println("✅ plugins directory found")
	}

	// Check for common issues
	checkCommonIssues(dir)
}

func analyzeLibDirectory(libDir string) {
	files, err := os.ReadDir(libDir)
	if err != nil {
		fmt.Printf("Error reading lib directory: %v\n", err)
		return
	}

	var qtLibCount, otherLibCount int
	var qtLibs []string

	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, "libQt") && strings.HasSuffix(name, ".so") {
			qtLibCount++
			qtLibs = append(qtLibs, name)
		} else if strings.HasSuffix(name, ".so") {
			otherLibCount++
		}
	}

	fmt.Printf("  Total Qt libraries found: %d\n", qtLibCount)
	fmt.Printf("  Total other libraries found: %d\n", otherLibCount)

	if qtLibCount > 0 {
		fmt.Println("  Qt libraries:")
		for _, lib := range qtLibs {
			fmt.Printf("    - %s\n", lib)
		}
	}

	if qtLibCount == 0 {
		fmt.Println("  ❌ WARNING: No Qt libraries found in lib directory")
	}
}

func analyzePluginsDirectory(pluginsDir string) {
	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		fmt.Printf("Error reading plugins directory: %v\n", err)
		return
	}

	var pluginDirs []string
	var pluginFiles []string

	for _, file := range files {
		if file.IsDir() {
			pluginDirs = append(pluginDirs, file.Name())
		} else {
			pluginFiles = append(pluginFiles, file.Name())
		}
	}

	fmt.Printf("  Plugin subdirectories: %v\n", pluginDirs)
	fmt.Printf("  Plugin files: %v\n", pluginFiles)

	// Check critical plugin directories
	criticalDirs := []string{"platforms", "imageformats", "iconengines", "tls"}
	for _, dir := range criticalDirs {
		dirPath := filepath.Join(pluginsDir, dir)
		if exists(dirPath) {
			fmt.Printf("  ✅ Critical plugin directory exists: %s\n", dir)

			// List files in critical directories
			if files, err := os.ReadDir(dirPath); err == nil {
				var soFiles []string
				for _, file := range files {
					if strings.HasSuffix(file.Name(), ".so") {
						soFiles = append(soFiles, file.Name())
					}
				}
				if len(soFiles) > 0 {
					fmt.Printf("    Files in %s: %v\n", dir, soFiles)
				}
			}
		} else {
			fmt.Printf("  ❌ Critical plugin directory missing: %s\n", dir)
		}
	}
}

func analyzeWrapperScript(scriptPath string) {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("  Error reading script: %v\n", err)
		return
	}

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

	if strings.Contains(contentStr, "export") {
		fmt.Printf("  ✅ Script uses export statements\n")
	}
}

func analyzeConfigFile(configPath string) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("  Error reading config: %v\n", err)
		return
	}

	contentStr := string(content)

	if strings.Contains(contentStr, "Portable=true") {
		fmt.Printf("  ✅ Portable mode enabled\n")
	} else {
		fmt.Printf("  ❌ Portable mode not enabled\n")
	}

	if strings.Contains(contentStr, "JavaDir=") {
		fmt.Printf("  ✅ Java directory configured\n")
	}
}

func checkCommonIssues(dir string) {
	fmt.Println("\nCommon Issues Check:")

	// Check for nested directory structure
	prismExe := filepath.Join(dir, "PrismLauncher")
	if runtime.GOOS == "windows" {
		prismExe += ".exe"
	}

	if !exists(prismExe) {
		// Check if there's a nested directory structure
		files, _ := os.ReadDir(dir)
		for _, file := range files {
			if file.IsDir() && strings.Contains(file.Name(), "PrismLauncher") {
				nestedPath := filepath.Join(dir, file.Name())
				nestedExe := filepath.Join(nestedPath, "PrismLauncher")
				if runtime.GOOS == "windows" {
					nestedExe += ".exe"
				}
				if exists(nestedExe) {
					fmt.Printf("❌ ISSUE: Nested directory structure detected - Prism executable is in subdirectory: %s\n", file.Name())
					fmt.Printf("   RECOMMENDATION: Extract files directly to target directory\n")
				}
			}
		}
	}

	// Check for missing Qt dependencies
	libDir := filepath.Join(dir, "lib")
	if exists(libDir) {
		qtCoreLib := filepath.Join(libDir, "libQt6Core.so.6")
		if !exists(qtCoreLib) {
			qt5CoreLib := filepath.Join(libDir, "libQt5Core.so.5")
			if !exists(qt5CoreLib) {
				fmt.Printf("❌ ISSUE: No Qt core libraries found (neither Qt5 nor Qt6)\n")
			} else {
				fmt.Printf("ℹ️  INFO: Qt5 libraries detected\n")
			}
		} else {
			fmt.Printf("ℹ️  INFO: Qt6 libraries detected\n")
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
