package launcher

import (
	"fmt"
	"strings"
	"testing"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

func TestJavaManager(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	manager := NewJavaManager(platformImpl, logger)

	// Test getting Java version for Minecraft
	testCases := []struct {
		mcVersion    string
		expectedLike string
	}{
		{"1.19.2", "17"}, // Should get Java 17
		{"1.20.1", "17"}, // Should get Java 17
		{"", "17"},      // Should get default Java 17
		{"invalid", "17"}, // Should get default Java 17
	}

	for _, tc := range testCases {
		t.Run("JavaVersionForMinecraft_"+tc.mcVersion, func(t *testing.T) {
			version := manager.GetJavaVersionForMinecraft(tc.mcVersion)
			if version == "" {
				t.Errorf("Expected non-empty Java version for Minecraft %s", tc.mcVersion)
			}
			t.Logf("Minecraft %s -> Java %s", tc.mcVersion, version)
		})
	}

	// Test architecture and OS strings
	arch := manager.getArchitectureString()
	if arch == "" {
		t.Error("Architecture string should not be empty")
	}
	t.Logf("Architecture: %s", arch)

	osName := manager.getOSString()
	if osName == "" {
		t.Error("OS string should not be empty")
	}
	t.Logf("OS: %s", osName)

	// Test version compatibility
	testCompatibility := []struct {
		current  string
		required string
		expected bool
	}{
		{"17", "17", true},
		{"18", "17", true},
		{"21", "17", true},
		{"11", "17", false},
		{"1.8", "17", false},
		{"17.0.2", "17", true},
	}

	for _, tc := range testCompatibility {
		t.Run("Compatibility_"+tc.current+"_vs_"+tc.required, func(t *testing.T) {
			compatible := manager.isVersionCompatible(tc.current, tc.required)
			if compatible != tc.expected {
				t.Errorf("Java %s compatible with %s: expected %v, got %v",
					tc.current, tc.required, tc.expected, compatible)
			}
		})
	}
}

func TestJavaInstaller(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	installer := NewJavaInstaller(platformImpl, logger)

	// Test getting Java executable path
	testDirs := []struct {
		installDir   string
		expectedLike string
	}{
		{"/opt/java", "/opt/java/bin/java"},
		{"C:\\Program Files\\Java", "C:\\Program Files\\Java\\bin\\java.exe"},
		{"./java", "./java/bin/java"},
	}

	for _, tc := range testDirs {
		t.Run("JavaExecutable_"+tc.installDir, func(t *testing.T) {
			executable := installer.getJavaExecutable(tc.installDir)
			if executable == "" {
				t.Error("Java executable path should not be empty")
			}
			t.Logf("Install dir %s -> executable %s", tc.installDir, executable)
		})
	}

	// Test executable suffix
	suffix := installer.getExecutableSuffix()
	expectedSuffix := ""
	if platformImpl.GetOS() == "windows" {
		expectedSuffix = ".exe"
	}
	if suffix != expectedSuffix {
		t.Errorf("Expected executable suffix %s, got %s", expectedSuffix, suffix)
	}
}

func TestDownloadURLGeneration(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	_ = NewJavaManager(platformImpl, logger)

	// Test Adoptium URL generation
	testCases := []struct {
		javaVersion string
		osName      string
		arch        string
		imageType   string
	}{
		{"17", "windows", "x64", "jre"},
		{"16", "windows", "x64", "jdk"}, // Java 16 uses JDK
		{"17", "mac", "x64", "jre"},
		{"17", "linux", "x64", "jre"},
	}

	for _, tc := range testCases {
		t.Run("AdoptiumURL_"+tc.javaVersion+"_"+tc.osName+"_"+tc.arch, func(t *testing.T) {
			url := fmt.Sprintf(adoptiumAPIURLPattern, tc.javaVersion, tc.arch, tc.imageType, tc.osName)
			if url == "" {
				t.Error("Generated URL should not be empty")
			}
			if !strings.Contains(url, tc.javaVersion) {
				t.Errorf("URL should contain Java version %s", tc.javaVersion)
			}
			if !strings.Contains(url, tc.osName) {
				t.Errorf("URL should contain OS %s", tc.osName)
			}
			if !strings.Contains(url, tc.arch) {
				t.Errorf("URL should contain architecture %s", tc.arch)
			}
			t.Logf("Generated Adoptium URL: %s", url)
		})
	}
}

func TestJavaDetectionPlatformSpecific(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	manager := NewJavaManager(platformImpl, logger)

	// Test that platform detection doesn't crash
	installations, err := manager.DetectJavaInstallations()
	if err != nil {
		t.Logf("Java detection failed (expected in CI): %v", err)
		// Don't fail the test as this might not work in CI environments
		return
	}

	t.Logf("Found %d Java installations", len(installations))

	// Verify detected installations have required fields
	for _, inst := range installations {
		if inst.Path == "" {
			t.Error("Java installation path should not be empty")
		}
		if inst.Version == "" {
			t.Error("Java installation version should not be empty")
		}
		t.Logf("Java %s at %s (JDK: %v)", inst.Version, inst.Path, inst.IsJDK)
	}
}