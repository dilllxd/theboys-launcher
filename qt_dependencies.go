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
	debugf("Detecting package manager on Linux")
	if runtime.GOOS != "linux" {
		debugf("Not on Linux, skipping package manager detection")
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
		debugf("Checking for package manager: %s", pm.Name)
		if _, err := exec.LookPath(pm.Name); err == nil {
			debugf("Found %s in PATH, verifying functionality", pm.Name)
			// Verify the package manager is actually working
			cmd := exec.Command("sh", "-c", pm.CheckCmd)
			output, err := cmd.CombinedOutput()
			if err == nil {
				debugf("Successfully verified package manager: %s", pm.Name)
				debugf("%s version output: %s", pm.Name, string(output))
				return &pm
			} else {
				debugf("Package manager %s found but verification failed: %v, output: %s", pm.Name, err, string(output))
			}
		} else {
			debugf("Package manager %s not found in PATH", pm.Name)
		}
	}

	debugf("No supported package manager found on this system")
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
	debugf("Checking Qt library dependencies")
	if runtime.GOOS != "linux" {
		debugf("Not on Linux, Qt libraries check skipped")
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
		debugf("Checking for library: %s", lib)
		if !isLibraryInstalled(lib) {
			debugf("Missing library: %s", lib)
			missingLibs = append(missingLibs, lib)
		} else {
			debugf("Found library: %s", lib)
		}
	}

	installed := len(missingLibs) == 0
	pm := getPackageManager()
	var packages []string

	if pm != nil {
		packages = getQtPackages(pm.Name)
		debugf("Qt packages to install for %s: %v", pm.Name, packages)
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
	debugf("Checking if library %s is installed", libraryName)
	// Use ldconfig -p to check for the library
	cmd := exec.Command("ldconfig", "-p")
	output, err := cmd.Output()
	if err != nil {
		debugf("ldconfig -p failed: %v, falling back to path checking", err)
		// Fallback to checking common library paths
		return checkLibraryPaths(libraryName)
	}

	debugf("Parsing ldconfig output for library %s", libraryName)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, libraryName) {
			debugf("Found library %s in ldconfig output: %s", libraryName, line)
			found = true
		}
	}

	if found {
		return true
	}

	debugf("Library %s not found in ldconfig output, checking paths", libraryName)
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

// ensurePatchelfInstalled ensures that patchelf is installed and available
// This function provides explicit verification and error handling for patchelf
func ensurePatchelfInstalled() error {
	if runtime.GOOS != "linux" {
		return nil // Only needed on Linux
	}

	logf("%s", stepLine("Ensuring patchelf is installed"))

	// First check if patchelf is already available
	if _, err := exec.LookPath("patchelf"); err == nil {
		// Verify it actually works by running it
		cmd := exec.Command("patchelf", "--version")
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			logf("%s", successLine(fmt.Sprintf("patchelf is already installed: %s", version)))
			return nil
		} else {
			logf("%s", warnLine("patchelf found but not working, attempting reinstallation"))
		}
	}

	// Get the package manager
	pm := getPackageManager()
	if pm == nil {
		return fmt.Errorf("no supported package manager found for patchelf installation")
	}

	logf("%s", infoLine(fmt.Sprintf("Installing patchelf using %s", pm.Name)))

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

	// Install patchelf
	installCmd := pm.InstallCmd + " patchelf"
	if pm.SudoRequired {
		installCmd = "sudo " + installCmd
	}

	logf("%s", infoLine("Installing patchelf..."))
	cmd := exec.Command("sh", "-c", installCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install patchelf: %w", err)
	}

	// Verify installation was successful
	logf("%s", stepLine("Verifying patchelf installation"))
	if _, err := exec.LookPath("patchelf"); err != nil {
		return fmt.Errorf("patchelf installation verification failed: patchelf not found in PATH")
	}

	// Test that patchelf actually works
	cmd = exec.Command("patchelf", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("patchelf installation verification failed: %w", err)
	}

	version := strings.TrimSpace(string(output))
	logf("%s", successLine(fmt.Sprintf("patchelf successfully installed: %s", version)))
	return nil
}
