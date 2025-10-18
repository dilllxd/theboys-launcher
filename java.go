package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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

// Prefer Adoptium API (stable), fall back to GitHub release asset.
// We want: OS=windows, arch=x64, image_type=jre (or jdk for Java 16), vm=hotspot, latest for specified version.
func fetchJREURL(javaVersion string) (string, error) {
	// Java 16 only has JDK builds available, not JRE
	imageType := "jre"
	if javaVersion == "16" {
		imageType = "jdk"
	}

	// 1) Adoptium API (v3)
	adoptium := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=x64&image_type=%s&os=windows", javaVersion, imageType)
	req, _ := http.NewRequest("GET", adoptium, nil)
	req.Header.Set("User-Agent", getUserAgent("Adoptium"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var payload []struct {
			Binaries []struct {
				Package struct {
					Link string `json:"link"`
				} `json:"package"`
			} `json:"binaries"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			for _, v := range payload {
				for _, b := range v.Binaries {
					if b.Package.Link != "" && strings.HasSuffix(strings.ToLower(b.Package.Link), ".zip") {
						return b.Package.Link, nil
					}
				}
			}
		}
	} else if resp != nil {
		resp.Body.Close()
	}

	// 2) Fallback to GitHub Releases: adoptium/temurin{version}-binaries
	api := fmt.Sprintf("https://api.github.com/repos/adoptium/temurin%s-binaries/releases/latest", javaVersion)
	req2, _ := http.NewRequest("GET", api, nil)
	req2.Header.Set("User-Agent", getUserAgent("Adoptium"))
	req2.Header.Set("Cache-Control", "no-cache")
	req2.Header.Set("Pragma", "no-cache")
	resp2, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		return "", fmt.Errorf("adoptium api and github fallback failed: %v", err2)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("github adoptium status %d", resp2.StatusCode)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp2.Body).Decode(&rel); err != nil {
		return "", err
	}
	// Example pattern: OpenJDK{version}U-jre_x64_windows_hotspot_*.zip or OpenJDK{version}U-jdk_x64_windows_hotspot_*.zip
	re := regexp.MustCompile(fmt.Sprintf(`(?i)^OpenJDK%sU-%s_x64_windows_hotspot_.*\.zip$`, javaVersion, imageType))
	for _, a := range rel.Assets {
		if re.MatchString(a.Name) {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no suitable Java %s %s zip found", javaVersion, imageType)
}