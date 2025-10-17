package launcher

import (
	"testing"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/config"
)

func TestModpackValidation(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()
	configManager := config.NewManager(platformImpl, logger)
	modpackConfigManager := config.NewModpackManager(platformImpl, logger)

	manager := NewModpackManager(configManager, modpackConfigManager, platformImpl, logger)

	// Test valid modpack
	validModpack := types.Modpack{
		ID:           "test",
		DisplayName:  "Test Modpack",
		PackURL:      "https://example.com/pack.toml",
		InstanceName: "Test Instance",
	}

	err := manager.ValidateModpack(validModpack)
	if err != nil {
		t.Fatalf("Valid modpack failed validation: %v", err)
	}

	// Test invalid modpacks
	testCases := []struct {
		name    string
		modpack types.Modpack
	}{
		{
			name: "Empty ID",
			modpack: types.Modpack{
				ID:           "",
				DisplayName:  "Test Modpack",
				PackURL:      "https://example.com/pack.toml",
				InstanceName: "Test Instance",
			},
		},
		{
			name: "Empty DisplayName",
			modpack: types.Modpack{
				ID:           "test",
				DisplayName:  "",
				PackURL:      "https://example.com/pack.toml",
				InstanceName: "Test Instance",
			},
		},
		{
			name: "Empty PackURL",
			modpack: types.Modpack{
				ID:           "test",
				DisplayName:  "Test Modpack",
				PackURL:      "",
				InstanceName: "Test Instance",
			},
		},
		{
			name: "Invalid URL",
			modpack: types.Modpack{
				ID:           "test",
				DisplayName:  "Test Modpack",
				PackURL:      "not-a-url",
				InstanceName: "Test Instance",
			},
		},
		{
			name: "Empty InstanceName",
			modpack: types.Modpack{
				ID:           "test",
				DisplayName:  "Test Modpack",
				PackURL:      "https://example.com/pack.toml",
				InstanceName: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.ValidateModpack(tc.modpack)
			if err == nil {
				t.Errorf("Expected validation error for %s, but got none", tc.name)
			}
		})
	}
}

func TestNormalizeModpacks(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()
	configManager := config.NewManager(platformImpl, logger)
	modpackConfigManager := config.NewModpackManager(platformImpl, logger)

	manager := NewModpackManager(configManager, modpackConfigManager, platformImpl, logger)

	// Test with valid and invalid modpacks
	input := []types.Modpack{
		{
			ID:           "valid1",
			DisplayName:  "Valid Modpack 1",
			PackURL:      "https://example.com/pack1.toml",
			InstanceName: "Valid Instance 1",
		},
		{
			ID:           "", // Invalid
			DisplayName:  "Invalid Modpack",
			PackURL:      "https://example.com/invalid.toml",
			InstanceName: "Invalid Instance",
		},
		{
			ID:           "valid2",
			DisplayName:  "Valid Modpack 2",
			PackURL:      "https://example.com/pack2.toml",
			InstanceName: "Valid Instance 2",
		},
		{
			ID:           "valid1", // Duplicate
			DisplayName:  "Duplicate Modpack",
			PackURL:      "https://example.com/duplicate.toml",
			InstanceName: "Duplicate Instance",
		},
	}

	result := manager.normalizeModpacks(input)

	// Should only have 2 valid modpacks (no duplicates, no invalid)
	if len(result) != 2 {
		t.Fatalf("Expected 2 normalized modpacks, got %d", len(result))
	}

	// Check that valid modpacks are present
	foundValid1 := false
	foundValid2 := false
	for _, mp := range result {
		if mp.ID == "valid1" {
			foundValid1 = true
		}
		if mp.ID == "valid2" {
			foundValid2 = true
		}
	}

	if !foundValid1 {
		t.Error("valid1 modpack not found in result")
	}
	if !foundValid2 {
		t.Error("valid2 modpack not found in result")
	}
}

func TestGetModpackLabel(t *testing.T) {
	logger := logging.NewLogger()
	logger.SetVerbose(false)
	platformImpl := platform.NewPlatform()
	configManager := config.NewManager(platformImpl, logger)
	modpackConfigManager := config.NewModpackManager(platformImpl, logger)

	manager := NewModpackManager(configManager, modpackConfigManager, platformImpl, logger)

	// Test with display name
	modpack1 := types.Modpack{
		ID:          "test",
		DisplayName: "Test Modpack",
	}
	label1 := manager.GetModpackLabel(modpack1)
	if label1 != "Test Modpack" {
		t.Errorf("Expected 'Test Modpack', got '%s'", label1)
	}

	// Test without display name (fallback to ID)
	modpack2 := types.Modpack{
		ID:          "test2",
		DisplayName: "",
	}
	label2 := manager.GetModpackLabel(modpack2)
	if label2 != "test2" {
		t.Errorf("Expected 'test2', got '%s'", label2)
	}

	// Test with whitespace-only display name
	modpack3 := types.Modpack{
		ID:          "test3",
		DisplayName: "   ",
	}
	label3 := manager.GetModpackLabel(modpack3)
	if label3 != "test3" {
		t.Errorf("Expected 'test3', got '%s'", label3)
	}
}