package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// -------------------- MultiMC Instance Creation --------------------

func createMultiMCInstance(modpack Modpack, packInfo *PackInfo, instDir, javaExe string) error {
	memoryMB := MemoryForModpack(modpack)
	minMB, maxMB := memoryMB, memoryMB

	// Create instance.cfg
	instanceLines := []string{
		"InstanceType=OneSix", // Use OneSix not Minecraft
		"name=" + modpack.InstanceName,
		"iconKey=default",
		"OverrideMemory=true",
		fmt.Sprintf("MinMemAlloc=%d", minMB),
		fmt.Sprintf("MaxMemAlloc=%d", maxMB),
		"OverrideJava=true",
		"JavaPath=" + filepath.ToSlash(javaExe),
		"AutomaticJava=false",
		"Notes=Managed by " + launcherName,
	}

	// Build components dynamically based on pack info
	lwjglInfo := getLWJGLVersionForMinecraft(packInfo.Minecraft)
	lwjglVersion := lwjglInfo.Version
	lwjglUID := lwjglInfo.UID
	lwjglName := lwjglInfo.Name

	components := []interface{}{
		map[string]interface{}{
			"cachedName":     lwjglName,
			"cachedVersion":  lwjglVersion,
			"cachedVolatile": true,
			"dependencyOnly": true,
			"uid":            lwjglUID,
			"version":        lwjglVersion,
		},
		map[string]interface{}{
			"cachedName":    "Minecraft",
			"cachedVersion": packInfo.Minecraft,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"suggests": lwjglVersion,
					"uid":      lwjglUID,
				},
			},
			"important": true,
			"uid":       "net.minecraft",
			"version":   packInfo.Minecraft,
		},
	}

	// Add modloader component based on what's detected
	var modloaderComponent map[string]interface{}
	switch packInfo.ModLoader {
	case "forge":
		modloaderComponent = map[string]interface{}{
			"cachedName":    "Forge",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "net.minecraftforge",
			"version": packInfo.LoaderVersion,
		}
	case "fabric":
		modloaderComponent = map[string]interface{}{
			"cachedName":    "Fabric Loader",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "net.fabricmc.fabric-loader",
			"version": packInfo.LoaderVersion,
		}
	case "quilt":
		modloaderComponent = map[string]interface{}{
			"cachedName":    "Quilt Loader",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "org.quiltmc.quilt-loader",
			"version": packInfo.LoaderVersion,
		}
	case "neoforge":
		modloaderComponent = map[string]interface{}{
			"cachedName":    "NeoForge",
			"cachedVersion": packInfo.LoaderVersion,
			"cachedRequires": []interface{}{
				map[string]interface{}{
					"equals": packInfo.Minecraft,
					"uid":    "net.minecraft",
				},
			},
			"uid":     "net.neoforged.neoforge",
			"version": packInfo.LoaderVersion,
		}
	}

	components = append(components, modloaderComponent)

	// Create mmc-pack.json with dynamic components
	mmcPack := map[string]interface{}{
		"formatVersion": 1,
		"components":    components,
	}

	// Create pack.json for MultiMC format with matching components
	pack := map[string]interface{}{
		"formatVersion": 3,
		"components":    components,
	}

	// Write all the required MultiMC files (only if they don't exist)
	instanceCfgPath := filepath.Join(instDir, "instance.cfg")
	mmcPackPath := filepath.Join(instDir, "mmc-pack.json")
	packJsonPath := filepath.Join(instDir, "pack.json")

	if !exists(instanceCfgPath) {
		if err := os.WriteFile(instanceCfgPath, []byte(strings.Join(instanceLines, "\n")+"\n"), 0644); err != nil {
			return err
		}
	}

	if !exists(mmcPackPath) {
		mmcPackData, err := json.MarshalIndent(mmcPack, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(mmcPackPath, mmcPackData, 0644); err != nil {
			return err
		}
	}

	if !exists(packJsonPath) {
		packData, err := json.MarshalIndent(pack, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(packJsonPath, packData, 0644); err != nil {
			return err
		}
	}

	return nil
}

func installModLoaderForInstance(instDir, javaBin string, packInfo *PackInfo) error {
	switch packInfo.ModLoader {
	case "forge":
		return installForgeForInstance(instDir, javaBin, packInfo)
	case "fabric":
		return installFabricForInstance(instDir, javaBin, packInfo)
	case "quilt":
		return installQuiltForInstance(instDir, javaBin, packInfo)
	case "neoforge":
		return installNeoForgeForInstance(instDir, javaBin, packInfo)
	default:
		return fmt.Errorf("unsupported modloader: %s", packInfo.ModLoader)
	}
}

func updateInstanceMemory(instDir string, memoryMB int) error {
	instanceCfgPath := filepath.Join(instDir, "instance.cfg")
	if !exists(instanceCfgPath) {
		return nil
	}

	data, err := os.ReadFile(instanceCfgPath)
	if err != nil {
		return err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var updated []string
	var hasMin, hasMax, hasOverride bool

	for _, line := range lines {
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "MinMemAlloc="):
			line = fmt.Sprintf("MinMemAlloc=%d", memoryMB)
			hasMin = true
		case strings.HasPrefix(line, "MaxMemAlloc="):
			line = fmt.Sprintf("MaxMemAlloc=%d", memoryMB)
			hasMax = true
		case strings.HasPrefix(line, "OverrideMemory="):
			line = "OverrideMemory=true"
			hasOverride = true
		}
		updated = append(updated, line)
	}

	if !hasOverride {
		updated = append(updated, "OverrideMemory=true")
	}
	if !hasMin {
		updated = append(updated, fmt.Sprintf("MinMemAlloc=%d", memoryMB))
	}
	if !hasMax {
		updated = append(updated, fmt.Sprintf("MaxMemAlloc=%d", memoryMB))
	}

	output := strings.Join(updated, "\n") + "\n"
	return os.WriteFile(instanceCfgPath, []byte(output), 0644)
}

func installForgeForInstance(instDir, javaBin string, packInfo *PackInfo) error {
	mcDir := filepath.Join(instDir, "minecraft") // Use minecraft not .minecraft

	// Check for Forge installation in MultiMC/Prism instance structure
	forgeJar := filepath.Join(mcDir, "libraries", "net", "minecraftforge", "forge", fmt.Sprintf("%s-%s", packInfo.Minecraft, packInfo.LoaderVersion), fmt.Sprintf("forge-%s-%s-universal.jar", packInfo.Minecraft, packInfo.LoaderVersion))
	mmcPackFile := filepath.Join(instDir, "mmc-pack.json")

	// Check if Forge is already installed
	if exists(forgeJar) && exists(mmcPackFile) {
		logf("Forge already completely installed in instance")
		return nil
	}

	// Download Forge installer
	forgeURL := fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/forge-%s-%s-installer.jar", packInfo.Minecraft, packInfo.LoaderVersion, packInfo.Minecraft, packInfo.LoaderVersion)
	utilDir := filepath.Join(filepath.Dir(instDir), "..", "..", "util")
	installerPath := filepath.Join(utilDir, "forge-installer.jar")

	logf("Downloading Forge installer...")
	if err := downloadTo(forgeURL, installerPath, 0644); err != nil {
		return fmt.Errorf("failed to download Forge installer: %w", err)
	}

	// Run Forge installer
	logf("Installing Forge...")
	fmt.Fprintf(out, "Running Forge installer... (this may take a few minutes)\n")

	cmd := exec.Command(javaBin, "-jar", installerPath, "--installClient", "--installServer")
	cmd.Dir = mcDir
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+filepath.Dir(filepath.Dir(javaBin)),
		"PATH="+filepath.Dir(filepath.Dir(javaBin))+";"+os.Getenv("PATH"),
	)

	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Forge installer failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(out, "✓ Forge installation completed successfully\n")

	// Clean up installer
	_ = os.Remove(installerPath)

	return nil
}

func installFabricForInstance(instDir, javaBin string, packInfo *PackInfo) error {
	mcDir := filepath.Join(instDir, "minecraft")

	// Download Fabric installer
	fabricURL := fmt.Sprintf("https://maven.fabricmc.net/net/fabricmc/fabric-installer/%s/fabric-installer-%s.jar", packInfo.LoaderVersion, packInfo.LoaderVersion)
	utilDir := filepath.Join(filepath.Dir(instDir), "..", "..", "util")
	installerPath := filepath.Join(utilDir, "fabric-installer.jar")

	logf("Downloading Fabric installer...")
	if err := downloadTo(fabricURL, installerPath, 0644); err != nil {
		return fmt.Errorf("failed to download Fabric installer: %w", err)
	}

	// Run Fabric installer
	logf("Installing Fabric Loader...")
	fmt.Fprintf(out, "Running Fabric installer... (this may take a few minutes)\n")

	cmd := exec.Command(javaBin, "-jar", installerPath, "client", "-dir", mcDir, "-mcversion", packInfo.Minecraft)
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+filepath.Dir(filepath.Dir(javaBin)),
		"PATH="+filepath.Dir(filepath.Dir(javaBin))+";"+os.Getenv("PATH"),
	)

	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Fabric installer failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(out, "✓ Fabric Loader installation completed successfully\n")

	// Clean up installer
	_ = os.Remove(installerPath)

	return nil
}

func installQuiltForInstance(instDir, javaBin string, packInfo *PackInfo) error {
	mcDir := filepath.Join(instDir, "minecraft")

	// Download Quilt installer
	quiltURL := fmt.Sprintf("https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/%s/quilt-installer-%s.jar", packInfo.LoaderVersion, packInfo.LoaderVersion)
	utilDir := filepath.Join(filepath.Dir(instDir), "..", "..", "util")
	installerPath := filepath.Join(utilDir, "quilt-installer.jar")

	logf("Downloading Quilt installer...")
	if err := downloadTo(quiltURL, installerPath, 0644); err != nil {
		return fmt.Errorf("failed to download Quilt installer: %w", err)
	}

	// Run Quilt installer
	logf("Installing Quilt Loader...")
	fmt.Fprintf(out, "Running Quilt installer... (this may take a few minutes)\n")

	cmd := exec.Command(javaBin, "-jar", installerPath, "install", "client", "--dir", mcDir, "--minecraft-version", packInfo.Minecraft)
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+filepath.Dir(filepath.Dir(javaBin)),
		"PATH="+filepath.Dir(filepath.Dir(javaBin))+";"+os.Getenv("PATH"),
	)

	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Quilt installer failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(out, "✓ Quilt Loader installation completed successfully\n")

	// Clean up installer
	_ = os.Remove(installerPath)

	return nil
}

func installNeoForgeForInstance(instDir, javaBin string, packInfo *PackInfo) error {
	mcDir := filepath.Join(instDir, "minecraft")

	// Download NeoForge installer
	neoforgeURL := fmt.Sprintf("https://maven.neoforged.net/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", packInfo.LoaderVersion, packInfo.LoaderVersion)
	utilDir := filepath.Join(filepath.Dir(instDir), "..", "..", "util")
	installerPath := filepath.Join(utilDir, "neoforge-installer.jar")

	logf("Downloading NeoForge installer...")
	if err := downloadTo(neoforgeURL, installerPath, 0644); err != nil {
		return fmt.Errorf("failed to download NeoForge installer: %w", err)
	}

	// Run NeoForge installer
	logf("Installing NeoForge...")
	fmt.Fprintf(out, "Running NeoForge installer... (this may take a few minutes)\n")

	cmd := exec.Command(javaBin, "-jar", installerPath, "--install-client", "--install-server")
	cmd.Dir = mcDir
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+filepath.Dir(filepath.Dir(javaBin)),
		"PATH="+filepath.Dir(filepath.Dir(javaBin))+";"+os.Getenv("PATH"),
	)

	// Hide console window on Windows
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("NeoForge installer failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(out, "✓ NeoForge installation completed successfully\n")

	// Clean up installer
	_ = os.Remove(installerPath)

	return nil
}
