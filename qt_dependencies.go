package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// QtDependencyInfo holds information about Qt dependencies
type QtDependencyInfo struct {
	Installed      bool
	MissingLibs    []string
	PackageManager string
	Packages       []string
}

// PackageManagerInfo holds information about a package manager
type PackageManagerInfo struct {
	Name         string
	CheckCmd     string
	InstallCmd   string
	UpdateCmd    string
	SearchCmd    string
	SudoRequired bool
}

// getPackageManager detects the available package manager on the system
func getPackageManager() *PackageManagerInfo {
	if runtime.GOOS != "linux" {
		return nil
	}

	// List of package managers to check in order of preference
	packageManagers := []PackageManagerInfo{
		{
			Name:         "apt",
			CheckCmd:     "apt --version",
			InstallCmd:   "apt install -y",
			UpdateCmd:    "apt update",
			SearchCmd:    "apt search",
			SudoRequired: true,
		},
		{
			Name:         "dnf",
			CheckCmd:     "dnf --version",
			InstallCmd:   "dnf install -y",
			UpdateCmd:    "dnf check-update",
			SearchCmd:    "dnf search",
			SudoRequired: true,
		},
		{
			Name:         "yum",
			CheckCmd:     "yum --version",
			InstallCmd:   "yum install -y",
			UpdateCmd:    "yum check-update",
			SearchCmd:    "yum search",
			SudoRequired: true,
		},
		{
			Name:         "pacman",
			CheckCmd:     "pacman --version",
			InstallCmd:   "pacman -S --noconfirm",
			UpdateCmd:    "pacman -Sy",
			SearchCmd:    "pacman -Ss",
			SudoRequired: true,
		},
		{
			Name:         "zypper",
			CheckCmd:     "zypper --version",
			InstallCmd:   "zypper install -y",
			UpdateCmd:    "zypper refresh",
			SearchCmd:    "zypper search",
			SudoRequired: true,
		},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.Name); err == nil {
			// Verify the package manager is actually working
			cmd := exec.Command("sh", "-c", pm.CheckCmd)
			if err := cmd.Run(); err == nil {
				logf("Detected package manager: %s", pm.Name)
				return &pm
			}
		}
	}

	logf("No supported package manager found")
	return nil
}

// getQtPackages returns the Qt packages to install based on the distribution
func getQtPackages(packageManager string) []string {
	switch packageManager {
	case "apt":
		return []string{
			"libqt6core6t64",
			"libqt6gui6",
			"libqt6widgets6",
			"libqt6network6",
			"libqt6svg6",
			"libxcb-cursor0",
			"patchelf", // Add patchelf for RPATH fixing
		}
	case "dnf", "yum":
		return []string{
			"qt6-qtbase",
			"qt6-qtgui",
			"qt6-qtwidgets",
			"qt6-qtnetwork",
			"qt6-qtsvg",
			"libxcb-cursor",
			"patchelf", // Add patchelf for RPATH fixing
		}
	case "pacman":
		return []string{
			"qt6-base",
			"qt6-svg",
			"libxcb",
			"patchelf", // Add patchelf for RPATH fixing
		}
	case "zypper":
		return []string{
			"libQt6Core6",
			"libQt6Gui6",
			"libQt6Widgets6",
			"libQt6Network6",
			"libQt6Svg6",
			"libxcb-cursor0",
			"patchelf", // Add patchelf for RPATH fixing
		}
	default:
		return []string{}
	}
}

// checkQtLibraries checks if required Qt libraries are installed
func checkQtLibraries() *QtDependencyInfo {
	if runtime.GOOS != "linux" {
		return &QtDependencyInfo{Installed: true}
	}

	logf("Checking Qt library dependencies...")

	// List of critical Qt libraries to check
	criticalLibs := []string{
		"libQt6Core.so.6",
		"libQt6Gui.so.6",
		"libQt6Widgets.so.6",
		"libQt6Network.so.6",
		"libQt6Svg.so.6",
		"libxcb-cursor.so.0",
	}

	var missingLibs []string

	// Check each library using ldconfig
	for _, lib := range criticalLibs {
		if !isLibraryInstalled(lib) {
			missingLibs = append(missingLibs, lib)
		}
	}

	installed := len(missingLibs) == 0
	pm := getPackageManager()
	var packages []string

	if pm != nil {
		packages = getQtPackages(pm.Name)
	}

	result := &QtDependencyInfo{
		Installed:      installed,
		MissingLibs:    missingLibs,
		PackageManager: "",
		Packages:       packages,
	}

	if pm != nil {
		result.PackageManager = pm.Name
	}

	if installed {
		logf("%s", successLine("All Qt libraries are installed"))
	} else {
		logf("%s", warnLine(fmt.Sprintf("Missing Qt libraries: %s", strings.Join(missingLibs, ", "))))
	}

	return result
}

// isLibraryInstalled checks if a library is available on the system
func isLibraryInstalled(libraryName string) bool {
	// Use ldconfig -p to check for the library
	cmd := exec.Command("ldconfig", "-p")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to checking common library paths
		return checkLibraryPaths(libraryName)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, libraryName) {
			return true
		}
	}

	// Fallback to checking common library paths
	return checkLibraryPaths(libraryName)
}

// checkLibraryPaths checks common library paths for the library
func checkLibraryPaths(libraryName string) bool {
	commonPaths := []string{
		"/usr/lib/x86_64-linux-gnu/",
		"/usr/lib/",
		"/usr/lib64/",
		"/lib/x86_64-linux-gnu/",
		"/lib/",
		"/lib64/",
		"/usr/local/lib/",
		"/usr/local/lib64/",
	}

	for _, path := range commonPaths {
		libPath := filepath.Join(path, libraryName)
		if exists(libPath) {
			return true
		}
	}

	return false
}

// promptUserForPermission prompts the user for permission to install packages
func promptUserForPermission(depInfo *QtDependencyInfo) bool {
	if depInfo.Installed || depInfo.PackageManager == "" {
		return true // No installation needed
	}

	logf("%s", sectionLine("Qt Dependency Installation"))
	logf("%s", infoLine("Prism Launcher requires Qt libraries to run properly."))
	logf("%s", infoLine(fmt.Sprintf("Missing libraries: %s", strings.Join(depInfo.MissingLibs, ", "))))
	logf("%s", infoLine(fmt.Sprintf("Package manager: %s", depInfo.PackageManager)))
	logf("%s", infoLine(fmt.Sprintf("Packages to install: %s", strings.Join(depInfo.Packages, ", "))))

	// In GUI mode, we'll use a dialog, but for now we'll use console prompt
	fmt.Print("\nWould you like to install these Qt dependencies? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logf("%s", warnLine("Failed to read user input, skipping installation"))
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// installQtDependencies installs the required Qt packages
func installQtDependencies(depInfo *QtDependencyInfo) error {
	if depInfo.Installed || depInfo.PackageManager == "" {
		return nil // No installation needed
	}

	pm := getPackageManager()
	if pm == nil {
		return fmt.Errorf("no supported package manager found")
	}

	logf("%s", stepLine("Installing Qt dependencies..."))

	// Update package lists first
	if pm.UpdateCmd != "" {
		updateCmd := pm.UpdateCmd
		if pm.SudoRequired {
			updateCmd = "sudo " + updateCmd
		}

		logf("%s", infoLine("Updating package lists..."))
		cmd := exec.Command("sh", "-c", updateCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to update package lists: %v", err)))
			// Continue anyway, as the update might not be critical
		}
	}

	// Install packages
	for _, pkg := range depInfo.Packages {
		installCmd := pm.InstallCmd + " " + pkg
		if pm.SudoRequired {
			installCmd = "sudo " + installCmd
		}

		logf("%s", infoLine(fmt.Sprintf("Installing %s...", pkg)))
		cmd := exec.Command("sh", "-c", installCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to install %s: %v", pkg, err)))
			return fmt.Errorf("failed to install package %s: %w", pkg, err)
		}

		logf("%s", successLine(fmt.Sprintf("Successfully installed %s", pkg)))
	}

	logf("%s", successLine("Qt dependencies installed successfully"))
	return nil
}

// ensureQtDependencies checks and installs Qt dependencies if needed
func ensureQtDependencies() error {
	if runtime.GOOS != "linux" {
		return nil // Only needed on Linux
	}

	// Check Qt dependencies
	depInfo := checkQtLibraries()

	if depInfo.Installed {
		return nil // All dependencies are already installed
	}

	// Prompt user for permission
	if !promptUserForPermission(depInfo) {
		logf("%s", warnLine("User declined Qt dependency installation"))
		return fmt.Errorf("Qt dependencies are required but installation was declined")
	}

	// Install dependencies
	if err := installQtDependencies(depInfo); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to install Qt dependencies: %v", err)))
		return fmt.Errorf("failed to install Qt dependencies: %w", err)
	}

	// Verify installation
	logf("%s", stepLine("Verifying Qt dependency installation..."))
	newDepInfo := checkQtLibraries()

	if !newDepInfo.Installed {
		logf("%s", warnLine("Qt dependency installation verification failed"))
		return fmt.Errorf("Qt dependency installation verification failed")
	}

	logf("%s", successLine("Qt dependencies successfully installed and verified"))
	return nil
}
