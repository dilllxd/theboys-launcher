package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// -------------------- Java Version Detection --------------------

// getJavaVersionForMinecraft fetches compatible Java versions from PrismLauncher meta-launcher GitHub
func getJavaVersionForMinecraft(mcVersion string) string {
	// Clean version string for GitHub path
	cleanVersion := strings.TrimSpace(mcVersion)
	if cleanVersion == "" {
		return "17" // default fallback
	}

	// Construct GitHub URL for PrismLauncher meta-launcher data
	url := fmt.Sprintf("https://raw.githubusercontent.com/PrismLauncher/meta-launcher/refs/heads/master/net.minecraft/%s.json", cleanVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to create request for Java compatibility data: %v", err)))
		return "17" // default fallback
	}
	req.Header.Set("User-Agent", getUserAgent("Java"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to fetch Java compatibility data for Minecraft %s: %v", cleanVersion, err)))
		return "17" // default fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logf("%s", warnLine(fmt.Sprintf("Java compatibility data not found for Minecraft %s (HTTP %d)", cleanVersion, resp.StatusCode)))
		return "17" // default fallback
	}

	// Parse the JSON response
	var data struct {
		CompatibleJavaMajors []int `json:"compatibleJavaMajors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to parse Java compatibility data for Minecraft %s: %v", cleanVersion, err)))
		return "17" // default fallback
	}

	// If we have compatible Java versions, select the best one
	if len(data.CompatibleJavaMajors) > 0 {
		// Choose the newest compatible Java version (prefer higher versions)
		bestJava := data.CompatibleJavaMajors[0]
		for _, javaVersion := range data.CompatibleJavaMajors {
			if javaVersion > bestJava {
				bestJava = javaVersion
			}
		}

		logf("%s", successLine(fmt.Sprintf("Found Java %d compatible with Minecraft %s from PrismLauncher meta", bestJava, cleanVersion)))
		return strconv.Itoa(bestJava)
	}

	logf("%s", warnLine(fmt.Sprintf("No Java compatibility data found for Minecraft %s", cleanVersion)))
	return "17" // default fallback
}

// LWJGLInfo holds version and UID information for LWJGL
type LWJGLInfo struct {
	Version string
	UID     string
	Name    string
}

// getLWJGLVersionForMinecraft fetches LWJGL version from PrismLauncher meta-launcher GitHub
func getLWJGLVersionForMinecraft(mcVersion string) LWJGLInfo {
	// Clean version string for GitHub path
	cleanVersion := strings.TrimSpace(mcVersion)
	if cleanVersion == "" {
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}

	// Construct GitHub URL for PrismLauncher meta-launcher data
	url := fmt.Sprintf("https://raw.githubusercontent.com/PrismLauncher/meta-launcher/refs/heads/master/net.minecraft/%s.json", cleanVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to create request for LWJGL data: %v", err)))
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}
	req.Header.Set("User-Agent", getUserAgent("LWJGL"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to fetch LWJGL data for Minecraft %s: %v", cleanVersion, err)))
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logf("%s", warnLine(fmt.Sprintf("LWJGL data not found for Minecraft %s (HTTP %d)", cleanVersion, resp.StatusCode)))
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}

	// Parse the JSON response
	var data struct {
		Requires []struct {
			Suggests string `json:"suggests"`
			UID      string `json:"uid"`
		} `json:"requires"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logf("%s", warnLine(fmt.Sprintf("Failed to parse LWJGL data for Minecraft %s: %v", cleanVersion, err)))
		return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
	}

	// Look for LWJGL requirement
	for _, req := range data.Requires {
		if req.UID == "org.lwjgl" || req.UID == "org.lwjgl3" {
			if req.Suggests != "" {
				var name string
				if req.UID == "org.lwjgl" {
					name = "LWJGL 2"
				} else {
					name = "LWJGL 3"
				}
				logf("%s", successLine(fmt.Sprintf("Found LWJGL %s (%s) for Minecraft %s from PrismLauncher meta", req.Suggests, name, cleanVersion)))
				return LWJGLInfo{Version: req.Suggests, UID: req.UID, Name: name}
			}
		}
	}

	logf("%s", warnLine(fmt.Sprintf("No LWJGL data found for Minecraft %s", cleanVersion)))
	return LWJGLInfo{Version: "3.3.3", UID: "org.lwjgl3", Name: "LWJGL 3"} // default fallback
}

// -------------------- Java URL discovery --------------------

// getPlatformJavaParams returns platform-specific parameters for Adoptium API
func getPlatformJavaParams() (osName, arch string) {
	switch runtime.GOOS {
	case "darwin":
		osName = "mac"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
		} else {
			arch = "x64"
		}
	case "windows":
		osName = "windows"
		arch = "x64"
	default:
		osName = "linux"
		arch = "x64"
	}
	return osName, arch
}

// Prefer Adoptium API (stable), fall back to GitHub release asset.
// We want: OS=windows, arch=x64, image_type=jre (or jdk for Java 16), vm=hotspot, latest for specified version.
func fetchJREURL(javaVersion string) (string, error) {
	// Java 16 only has JDK builds available, not JRE
	imageType := "jre"
	if javaVersion == "16" {
		imageType = "jdk"
	}

	// 1) Primary: Adoptium API (v3) - most reliable method
	osName, arch := getPlatformJavaParams()
	adoptium := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=%s&image_type=%s&os=%s", javaVersion, arch, imageType, osName)
	req, _ := http.NewRequest("GET", adoptium, nil)
	req.Header.Set("User-Agent", getUserAgent("Adoptium"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var payload []struct {
			Binary struct {
				Package struct {
					Link string `json:"link"`
					Name string `json:"name"`
				} `json:"package"`
			} `json:"binary"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			for _, v := range payload {
				// Prefer zip files (packages) over installers
				if v.Binary.Package.Link != "" && strings.HasSuffix(strings.ToLower(v.Binary.Package.Link), ".zip") {
					return v.Binary.Package.Link, nil
				}
			}
		}
	} else if resp != nil {
		resp.Body.Close()
	}

	// 2) Fallback: GitHub Releases - use simple URL construction without scraping
	// Only used if Adoptium API fails
	releaseURL := fmt.Sprintf("https://github.com/adoptium/temurin%s-binaries/releases/latest", javaVersion)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp2, err2 := client.Get(releaseURL)
	if err2 != nil {
		return "", fmt.Errorf("adoptium api and github fallback failed: %v", err2)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("github adoptium status %d", resp2.StatusCode)
	}

	// Extract tag from the final redirected URL
	finalURL := resp2.Request.URL.String()
	latestTag := ""
	if strings.Contains(finalURL, "/releases/tag/") {
		parts := strings.Split(finalURL, "/releases/tag/")
		if len(parts) > 1 {
			latestTag = parts[1]
		}
	}

	if latestTag == "" {
		return "", fmt.Errorf("could not extract tag from GitHub redirect URL: %s", finalURL)
	}

	// Generate platform-specific asset name
	assetName := generateJavaAssetName(javaVersion, imageType, osName, arch, latestTag)
	assetURL := fmt.Sprintf("https://github.com/adoptium/temurin%s-binaries/releases/download/%s/%s", javaVersion, latestTag, assetName)

	return assetURL, nil
}

// generateJavaAssetName creates platform-specific asset names for Adoptium releases
func generateJavaAssetName(javaVersion, imageType, osName, arch, tag string) string {
	tagWithoutJdk := strings.TrimPrefix(tag, "jdk")
	// Remove hyphens from the tag part for the asset name (e.g., "8u462-b08" -> "8u462b08")
	tagWithoutHyphens := strings.ReplaceAll(tagWithoutJdk, "-", "")

	switch osName {
	case "windows":
		return fmt.Sprintf("OpenJDK%sU-%s_x64_windows_hotspot_%s.zip", javaVersion, imageType, tagWithoutHyphens)
	case "mac":
		if arch == "aarch64" {
			return fmt.Sprintf("OpenJDK%sU-%s_aarch64_mac_hotspot_%s.tar.gz", javaVersion, imageType, tagWithoutHyphens)
		}
		return fmt.Sprintf("OpenJDK%sU-%s_x64_mac_hotspot_%s.tar.gz", javaVersion, imageType, tagWithoutHyphens)
	case "linux":
		return fmt.Sprintf("OpenJDK%sU-%s_x64_linux_hotspot_%s.tar.gz", javaVersion, imageType, tagWithoutHyphens)
	default:
		// Fallback to Windows pattern
		return fmt.Sprintf("OpenJDK%sU-%s_x64_windows_hotspot_%s.zip", javaVersion, imageType, tagWithoutHyphens)
	}
}
