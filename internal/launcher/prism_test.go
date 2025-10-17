package launcher

import (
	"os"
	"testing"

	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

func TestPrismManager(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	manager := NewPrismManager(platformImpl, logger)

	// Test getting Prism executable path
	testCases := []struct {
		installDir string
		expected   string
	}{
		{"/opt/prism", "/opt/prism/PrismLauncher"},
		{"C:\\Prism", "C:\\Prism\\PrismLauncher.exe"},
		{"./prism", "./prism/PrismLauncher"},
	}

	for _, tc := range testCases {
		t.Run("PrismExecutable_"+tc.installDir, func(t *testing.T) {
			executable := manager.GetPrismExecutable(tc.installDir)
			if executable == "" {
				t.Error("Prism executable path should not be empty")
			}
			t.Logf("Install dir %s -> executable %s", tc.installDir, executable)
		})
	}

	// Test memory calculation
	minMB, maxMB := manager.calculateOptimalMemory()
	if minMB <= 0 || maxMB <= 0 {
		t.Error("Memory values should be positive")
	}
	if maxMB < minMB {
		t.Error("Max memory should be greater than or equal to min memory")
	}
	t.Logf("Memory allocation: %dMB - %dMB", minMB, maxMB)

	// Test modloader component creation
	modloaderTestCases := []struct {
		name        string
		modLoader   string
		loaderVersion string
		minecraft   string
		shouldSucceed bool
	}{
		{"Forge", "forge", "47.2.0", "1.20.1", true},
		{"Fabric", "fabric", "0.14.21", "1.20.1", true},
		{"Quilt", "quilt", "0.21.0", "1.20.1", true},
		{"Unknown", "unknown", "1.0.0", "1.20.1", false},
	}

	for _, tc := range modloaderTestCases {
		t.Run("ModloaderComponent_"+tc.name, func(t *testing.T) {
			packInfo := &PackInfo{
				Minecraft:     tc.minecraft,
				ModLoader:     tc.modLoader,
				LoaderVersion: tc.loaderVersion,
			}

			component := manager.createModloaderComponent(packInfo)

			if tc.shouldSucceed && component == nil {
				t.Errorf("Expected modloader component for %s", tc.name)
			} else if !tc.shouldSucceed && component != nil {
				t.Errorf("Expected no modloader component for %s", tc.name)
			}

			if component != nil {
				t.Logf("Created %s component: %s", tc.name, component["cachedName"])
			}
		})
	}

	// Test user agent generation
	userAgent := manager.getUserAgent("Test")
	expected := "TheBoys-Test/dev"
	if userAgent != expected {
		t.Errorf("Expected user agent %s, got %s", expected, userAgent)
	}
}

func TestInstanceManager(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	// Create mock dependencies
	prismManager := NewPrismManager(platformImpl, logger)
	javaManager := NewJavaManager(platformImpl, logger)

	manager := NewInstanceManager(platformImpl, logger, prismManager, javaManager)

	// Test instance ID generation
	testCases := []struct {
		modpackID string
		prefix    string
	}{
		{"winterpack", "winterpack-"},
		{"Example Pack", "example-pack-"},
		{"test123", "test123-"},
	}

	for _, tc := range testCases {
		t.Run("GenerateInstanceID_"+tc.modpackID, func(t *testing.T) {
			instanceID := manager.generateInstanceID(tc.modpackID)
			if instanceID == "" {
				t.Error("Instance ID should not be empty")
			}
			if !manager.startsWith(instanceID, tc.prefix) {
				t.Errorf("Instance ID should start with %s, got %s", tc.prefix, instanceID)
			}
			t.Logf("Generated instance ID: %s", instanceID)
		})
	}

	// Test getting instances directory
	instancesDir := manager.getInstancesDir()
	if instancesDir == "" {
		t.Error("Instances directory should not be empty")
	}
	t.Logf("Instances directory: %s", instancesDir)

	// Test metadata path generation
	instanceID := "test-instance"
	metadataPath := manager.getInstanceMetadataPath(instanceID)
	if metadataPath == "" {
		t.Error("Metadata path should not be empty")
	}
	if !manager.contains(metadataPath, instanceID+".json") {
		t.Errorf("Metadata path should contain instance ID: %s", instanceID)
	}
	t.Logf("Metadata path: %s", metadataPath)
}

func TestPackInfoParsing(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()

	manager := NewInstanceManager(platformImpl, logger, nil, nil)

	// Test valid pack.toml content
	validToml := `name = "Test Modpack"

[[versions]]
minecraft = "1.20.1"

[[versions.loaders]]
loader = "fabric"
loader-version = "0.14.21"
`

	packInfo, err := manager.parsePackTomlContent(validToml)
	if err != nil {
		t.Fatalf("Failed to parse valid pack.toml: %v", err)
	}

	if packInfo.Name != "Test Modpack" {
		t.Errorf("Expected name 'Test Modpack', got %s", packInfo.Name)
	}
	if packInfo.Minecraft != "1.20.1" {
		t.Errorf("Expected minecraft '1.20.1', got %s", packInfo.Minecraft)
	}
	if packInfo.ModLoader != "fabric" {
		t.Errorf("Expected modloader 'fabric', got %s", packInfo.ModLoader)
	}
	if packInfo.LoaderVersion != "0.14.21" {
		t.Errorf("Expected loader version '0.14.21', got %s", packInfo.LoaderVersion)
	}

	t.Logf("Parsed pack info: %+v", packInfo)

	// Test invalid pack.toml content
	invalidToml := `name = "Test Modpack"

# Missing versions section entirely
`

	_, err = manager.parsePackTomlContent(invalidToml)
	if err == nil {
		t.Error("Expected error for invalid pack.toml")
	}
}

// Helper methods for testing
func (p *PrismManager) startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (i *InstanceManager) startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (i *InstanceManager) contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (i *InstanceManager) parsePackTomlContent(content string) (*PackInfo, error) {
	// Create a temporary file for testing
	tmpFile := "/tmp/test_pack.toml"
	file, err := os.Create(tmpFile)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return nil, err
	}

	return i.parsePackToml(tmpFile)
}