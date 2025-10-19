package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// -------------------- packwiz bootstrap URL discovery --------------------

// downloadPackwizInstaller downloads the main packwiz-installer.jar using our non-GitHub API method
func downloadPackwizInstaller(destPath string) error {
	releasesURL := "https://github.com/packwiz/packwiz-installer/releases"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(releasesURL)
	if err != nil {
		return fmt.Errorf("failed to fetch packwiz-installer releases page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("packwiz-installer releases page returned status %d", resp.StatusCode)
	}

	// Read HTML content
	htmlBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read packwiz-installer releases page HTML: %w", err)
	}
	html := string(htmlBody)

	// Extract the first (latest) release tag from the releases page
	tagPattern := `/packwiz/packwiz-installer/releases/tag/([^"]+)`
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindStringSubmatch(html)

	if len(tagMatches) < 2 {
		return errors.New("could not find any packwiz-installer release tags")
	}

	latestTag := tagMatches[1]

	// Look for the main packwiz-installer.jar file (not bootstrap)
	assetPatterns := []string{
		fmt.Sprintf("packwiz-installer-%s.jar", latestTag),
		"packwiz-installer.jar", // Generic fallback
	}

	for _, assetName := range assetPatterns {
		assetURL := fmt.Sprintf("https://github.com/packwiz/packwiz-installer/releases/download/%s/%s", latestTag, assetName)

		// Verify the asset exists by making a HEAD request
		headReq, err := http.NewRequest("HEAD", assetURL, nil)
		if err != nil {
			continue
		}
		headReq.Header.Set("User-Agent", getUserAgent("General"))

		headResp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			continue
		}
		headResp.Body.Close()

		if headResp.StatusCode == 200 {
			// Download the file
			logf("Downloading packwiz-installer.jar from: %s", assetURL)
			return downloadTo(assetURL, destPath, 0644)
		}
	}

	return errors.New("no packwiz-installer.jar assets found")
}

func fetchPackwizBootstrapURL() (string, error) {
	// Use GitHub's releases page to find the latest packwiz bootstrap without API
	releasesURL := "https://github.com/packwiz/packwiz-installer-bootstrap/releases"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(releasesURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch packwiz releases page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("packwiz releases page returned status %d", resp.StatusCode)
	}

	// Read HTML content
	htmlBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read packwiz releases page HTML: %w", err)
	}
	html := string(htmlBody)

	// Extract the first (latest) release tag from the releases page
	tagPattern := `/packwiz/packwiz-installer-bootstrap/releases/tag/([^"]+)`
	tagRe := regexp.MustCompile(tagPattern)
	tagMatches := tagRe.FindStringSubmatch(html)

	if len(tagMatches) < 2 {
		return "", errors.New("could not find any packwiz bootstrap release tags")
	}

	latestTag := tagMatches[1]

	// Try common asset patterns for packwiz bootstrap
	possibleAssets := []string{
		fmt.Sprintf("packwiz-installer-bootstrap-%s.jar", latestTag),
		fmt.Sprintf("packwiz-installer-bootstrap%s", getExecutableExtension()), // Platform-specific bootstrap
		"packwiz-installer-bootstrap.jar",              // Generic fallback
	}

	for _, assetName := range possibleAssets {
		assetURL := fmt.Sprintf("https://github.com/packwiz/packwiz-installer-bootstrap/releases/download/%s/%s", latestTag, assetName)

		// Verify the asset exists by making a HEAD request
		headReq, err := http.NewRequest("HEAD", assetURL, nil)
		if err != nil {
			continue
		}
		headReq.Header.Set("User-Agent", getUserAgent("General"))

		headResp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			continue
		}
		headResp.Body.Close()

		if headResp.StatusCode == 200 {
			return assetURL, nil
		}
	}

	return "", errors.New("no packwiz bootstrap assets found")
}

// -------------------- Modpack Version Checking --------------------

// PackConfig represents the structure of a pack.toml file
type PackConfig struct {
	Version  string       `toml:"version"`
	Versions PackVersions `toml:"versions"`
}

// PackVersions represents the [versions] section from pack.toml
type PackVersions struct {
	Minecraft string `toml:"minecraft"`
	Forge     string `toml:"forge"`
	Fabric    string `toml:"fabric"`
	Quilt     string `toml:"quilt"`
	NeoForge  string `toml:"neoforge"`
}

// PackInfo holds the complete modpack information from pack.toml
type PackInfo struct {
	Version       string
	Minecraft     string
	ModLoader     string // "forge", "fabric", "quilt", "neoforge"
	LoaderVersion string
}

// fetchPackInfo reads the remote pack.toml and extracts all version information
func fetchPackInfo(packURL string) (*PackInfo, error) {
	req, err := http.NewRequest("GET", packURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", getUserAgent("General"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, packURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var packConfig PackConfig
	if err := toml.Unmarshal(body, &packConfig); err != nil {
		return nil, fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	if packConfig.Version == "" {
		return nil, errors.New("no version found in pack.toml")
	}

	// Determine modloader and versions
	info := &PackInfo{
		Version:   packConfig.Version,
		Minecraft: packConfig.Versions.Minecraft,
	}

	// Determine which modloader is being used
	if packConfig.Versions.Forge != "" {
		info.ModLoader = "forge"
		info.LoaderVersion = packConfig.Versions.Forge
	} else if packConfig.Versions.Fabric != "" {
		info.ModLoader = "fabric"
		info.LoaderVersion = packConfig.Versions.Fabric
	} else if packConfig.Versions.Quilt != "" {
		info.ModLoader = "quilt"
		info.LoaderVersion = packConfig.Versions.Quilt
	} else if packConfig.Versions.NeoForge != "" {
		info.ModLoader = "neoforge"
		info.LoaderVersion = packConfig.Versions.NeoForge
	} else {
		return nil, errors.New("no supported modloader found in pack.toml [versions] section")
	}

	return info, nil
}

// fetchRemotePackVersion fetches the remote pack.toml and extracts the version
func fetchRemotePackVersion(packURL string) (string, error) {
	req, err := http.NewRequest("GET", packURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", getUserAgent("General"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, packURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var packConfig PackConfig
	if err := toml.Unmarshal(body, &packConfig); err != nil {
		return "", fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	if packConfig.Version == "" {
		return "", errors.New("no version found in pack.toml")
	}

	return packConfig.Version, nil
}

// getLocalPackVersion gets the version from our local version tracking file
func getLocalPackVersion(mp Modpack, instDir string) (string, error) {
	versionFilePath := filepath.Join(instDir, versionFileNameFor(mp))

	// Check if our version file exists
	if !exists(versionFilePath) {
		logf("Debug: %s version file not found at %s", modpackLabel(mp), versionFilePath)
		return "", nil // No version file exists
	}

	body, err := os.ReadFile(versionFilePath)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(body))
	logf("Debug: Found local %s version %s at %s", modpackLabel(mp), version, versionFilePath)
	return version, nil
}

// saveLocalVersion saves the current modpack version to our tracking file
func saveLocalVersion(mp Modpack, instDir, version string) error {
	versionFilePath := filepath.Join(instDir, versionFileNameFor(mp))

	if err := os.WriteFile(versionFilePath, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to save local version: %w", err)
	}

	logf("Debug: Saved local %s version %s to %s", modpackLabel(mp), version, versionFilePath)
	return nil
}

// checkModpackUpdate checks if there's a modpack update available
func checkModpackUpdate(modpack Modpack, instDir string) (bool, string, string, error) {
	remoteVersion, err := fetchRemotePackVersion(modpack.PackURL)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to fetch remote modpack version: %w", err)
	}

	localVersion, err := getLocalPackVersion(modpack, instDir)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get local modpack version: %w", err)
	}

	packName := modpackLabel(modpack)

	// If no local version exists, we need to install
	if localVersion == "" {
		logf("No local %s found, will install version %s", packName, remoteVersion)
		return true, "", remoteVersion, nil
	}

	// Compare versions
	if localVersion != remoteVersion {
		logf("%s update available: %s → %s", packName, localVersion, remoteVersion)
		return true, localVersion, remoteVersion, nil
	}

	logf("%s is up to date (%s)", packName, localVersion)
	return false, localVersion, remoteVersion, nil
}

// -------------------- Modpack Backup & Restore --------------------

// createModpackBackup creates a backup of the current modpack before updating
func createModpackBackup(mp Modpack, mcDir string) (string, error) {
	packName := modpackLabel(mp)
	// Clean up old backups (keep only the 3 most recent)
	if err := cleanupOldBackups(mp, mcDir, 3); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to clean old backups: %v", err)))
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupName := backupPrefixFor(mp) + timestamp
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(mcDir)))
	backupPath := filepath.Join(rootDir, "util", "backups", backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Directories to backup
	dirsToBackup := []string{"mods", "config", "resourcepacks", "shaderpacks"}

	logf("%s", stepLine(fmt.Sprintf("Creating backup %s for %s", backupName, packName)))

	var backedUpItems []string
	for _, dir := range dirsToBackup {
		srcPath := filepath.Join(mcDir, dir)
		dstPath := filepath.Join(backupPath, dir)

		if exists(srcPath) {
			if err := copyDir(srcPath, dstPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to backup %s: %v", dir, err)))
			} else {
				backedUpItems = append(backedUpItems, dir)
			}
		}
	}

	// Backup our version file if it exists
	versionFile := versionFileNameFor(mp)
	versionFileSrc := filepath.Join(filepath.Dir(mcDir), versionFile)
	versionFileDst := filepath.Join(backupPath, versionFile)
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to backup version file: %v", err)))
		} else {
			backedUpItems = append(backedUpItems, versionFile)
		}
	}

	if len(backedUpItems) == 0 {
		logf("%s", warnLine(fmt.Sprintf("No files found to backup for %s", packName)))
		return "", nil
	}

	logf("%s", successLine(fmt.Sprintf("Backup created for %s: %s (items: %s)", packName, backupName, strings.Join(backedUpItems, ", "))))
	return backupPath, nil
}

// restoreModpackBackup restores from a backup if the update fails
func restoreModpackBackup(mp Modpack, backupPath, mcDir string) error {
	if backupPath == "" || !exists(backupPath) {
		return errors.New("no backup available to restore")
	}

	packName := modpackLabel(mp)
	logf("%s", stepLine(fmt.Sprintf("Restoring %s from backup", packName)))

	// Remove current modpack directories
	dirsToRemove := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	for _, dir := range dirsToRemove {
		dirPath := filepath.Join(mcDir, dir)
		if exists(dirPath) {
			if err := os.RemoveAll(dirPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to remove %s during restore: %v", dir, err)))
			}
		}
	}

	// Restore backup
	dirsToRestore := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	var restoredItems []string

	for _, dir := range dirsToRestore {
		srcPath := filepath.Join(backupPath, dir)
		dstPath := filepath.Join(mcDir, dir)

		if exists(srcPath) {
			if err := copyDir(srcPath, dstPath); err != nil {
				logf("%s", warnLine(fmt.Sprintf("Failed to restore %s: %v", dir, err)))
			} else {
				restoredItems = append(restoredItems, dir)
			}
		}
	}

	// Restore our version file
	versionFile := versionFileNameFor(mp)
	versionFileSrc := filepath.Join(backupPath, versionFile)
	versionFileDst := filepath.Join(filepath.Dir(mcDir), versionFile)
	if exists(versionFileSrc) {
		if err := copyFile(versionFileSrc, versionFileDst); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to restore version file: %v", err)))
		} else {
			restoredItems = append(restoredItems, versionFile)
		}
	}

	if len(restoredItems) == 0 {
		return errors.New("nothing to restore from backup")
	}

	logf("%s", successLine(fmt.Sprintf("Restored %s: %s", packName, strings.Join(restoredItems, ", "))))
	return nil
}

// cleanupOldBackups removes old backups, keeping only the most recent ones
func cleanupOldBackups(mp Modpack, mcDir string, keepCount int) error {
	packName := modpackLabel(mp)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(mcDir)))
	backupsDir := filepath.Join(rootDir, "util", "backups")
	if !exists(backupsDir) {
		return nil
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return err
	}

	var backupDirs []os.DirEntry
	prefix := backupPrefixFor(mp)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backupDirs = append(backupDirs, entry)
		}
	}

	// Sort by name (which includes timestamp, so this sorts by time)
	if len(backupDirs) <= keepCount {
		return nil
	}

	// Remove oldest backups (keep the most recent ones)
	sort.Slice(backupDirs, func(i, j int) bool {
		return backupDirs[i].Name() > backupDirs[j].Name() // newer names first
	})

	toRemove := backupDirs[keepCount:]
	for _, entry := range toRemove {
		removePath := filepath.Join(backupsDir, entry.Name())
		if err := os.RemoveAll(removePath); err != nil {
			logf("%s", warnLine(fmt.Sprintf("Failed to remove old %s backup %s: %v", packName, entry.Name(), err)))
		} else {
			logf("%s", successLine(fmt.Sprintf("Removed old %s backup: %s", packName, entry.Name())))
		}
	}

	return nil
}
