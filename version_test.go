package main

import (
	"testing"
)

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.2.3", "1.2.3"},
		{"V1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"v1.2.3-dev.abc123", "1.2.3-dev.abc123"},
		{"  v1.2.3  ", "1.2.3"},
		{"", ""},
	}

	for _, test := range tests {
		result := normalizeTag(test.input)
		if result != test.expected {
			t.Errorf("normalizeTag(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestParseSemverInts(t *testing.T) {
	tests := []struct {
		input         string
		expectedMajor int
		expectedMinor int
		expectedPatch int
	}{
		{"1.2.3", 1, 2, 3},
		{"10.20.30", 10, 20, 30},
		{"1.2.3-dev.abc123", 1, 2, 3},
		{"1.2", 1, 2, 0},
		{"1", 1, 0, 0},
		{"", 0, 0, 0},
		{"invalid", 0, 0, 0},
	}

	for _, test := range tests {
		major, minor, patch := parseSemverInts(test.input)
		if major != test.expectedMajor || minor != test.expectedMinor || patch != test.expectedPatch {
			t.Errorf("parseSemverInts(%q) = (%d, %d, %d), want (%d, %d, %d)",
				test.input, major, minor, patch, test.expectedMajor, test.expectedMinor, test.expectedPatch)
		}
	}
}

func TestGetPrerelease(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", ""},
		{"1.2.3-dev.abc123", "dev.abc123"},
		{"1.2.3-alpha", "alpha"},
		{"1.2.3-beta.2", "beta.2"},
		{"1.2.3-rc.1", "rc.1"},
		{"", ""},
	}

	for _, test := range tests {
		result := getPrerelease(test.input)
		if result != test.expected {
			t.Errorf("getPrerelease(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestComparePrerelease(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		// Stable vs prerelease
		{"", "alpha", 1}, // stable > prerelease
		{"beta", "", -1}, // prerelease < stable
		{"", "", 0},      // both stable

		// Same prerelease
		{"alpha", "alpha", 0},
		{"dev.abc123", "dev.abc123", 0},

		// Different prereleases (lexicographical)
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},
		{"dev.abc123", "dev.def456", -1}, // abc < def
		{"dev.def456", "dev.abc123", 1},  // def > abc

		// Numeric comparison
		{"rc.1", "rc.2", -1},
		{"rc.2", "rc.1", 1},
		{"dev.10", "dev.2", 1},

		// Numeric vs alphanumeric
		{"rc.1", "rc.beta", -1}, // numeric < alphanumeric
		{"rc.beta", "rc.1", 1},  // alphanumeric > numeric

		// Different lengths
		{"dev.abc", "dev.abc.123", -1},
		{"dev.abc.123", "dev.abc", 1},
	}

	for _, test := range tests {
		result := comparePrerelease(test.a, test.b)
		if result != test.expected {
			t.Errorf("comparePrerelease(%q, %q) = %d, want %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		// Basic version comparison
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.3.0", -1},
		{"1.3.0", "1.2.3", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},

		// Equal versions
		{"1.2.3", "1.2.3", 0},
		{"1.2.3-dev.abc123", "1.2.3-dev.abc123", 0},

		// Stable vs prerelease (same core version)
		{"1.2.3", "1.2.3-dev.abc123", 1},  // stable > prerelease
		{"1.2.3-dev.abc123", "1.2.3", -1}, // prerelease < stable

		// Different core versions with prereleases
		{"1.2.3", "1.2.4-dev.abc123", -1}, // 1.2.3 < 1.2.4-dev
		{"1.2.4-dev.abc123", "1.2.3", 1},  // 1.2.4-dev > 1.2.3

		// Prerelease comparison (same core version)
		{"1.2.3-dev.abc123", "1.2.3-dev.def456", -1}, // abc < def
		{"1.2.3-dev.def456", "1.2.3-dev.abc123", 1},  // def > abc

		// Complex dev versions
		{"3.2.27-dev.5c0625a", "3.2.27", -1}, // dev < stable
		{"3.2.27", "3.2.27-dev.5c0625a", 1},  // stable > dev

		// Version with missing components
		{"1.2", "1.2.0", 0},
		{"1.2.0", "1.2", 0},
		{"1", "1.0.0", 0},
		{"1.0.0", "1", 0},
	}

	for _, test := range tests {
		result := compareSemver(test.a, test.b)
		if result != test.expected {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestVersionComparisonScenarios(t *testing.T) {
	// Test real-world scenarios from the launcher

	// Scenario 1: Current dev version vs latest stable
	currentDev := "3.2.27-dev.5c0625a"
	latestStable := "3.2.26"
	result := compareSemver(currentDev, latestStable)
	if result <= 0 {
		t.Errorf("Dev version %q should be newer than stable %q", currentDev, latestStable)
	}

	// Scenario 2: Current stable vs newer dev
	currentStable := "3.2.26"
	newerDev := "3.2.27-dev.abc123"
	result = compareSemver(currentStable, newerDev)
	if result >= 0 {
		t.Errorf("Stable %q should be older than dev %q", currentStable, newerDev)
	}

	// Scenario 3: Same version, different prerelease
	dev1 := "3.2.27-dev.abc123"
	dev2 := "3.2.27-def456"
	result = compareSemver(dev1, dev2)
	// Note: Based on current implementation, "abc123" is considered newer than "def456"
	// This test documents current behavior, even if it seems counterintuitive
	if result <= 0 {
		t.Errorf("%q should be newer than %q (got %d)", dev1, dev2, result)
	}

	// Scenario 4: Major version jump
	oldVersion := "2.5.0"
	newVersion := "3.0.0"
	result = compareSemver(oldVersion, newVersion)
	if result >= 0 {
		t.Errorf("%q should be older than %q", oldVersion, newVersion)
	}
}

func TestIsPrereleaseTag(t *testing.T) {
	tests := []struct {
		input       string
		expected    bool
		description string
	}{
		// Stable versions (should return false)
		{"v1.2.3", false, "Standard stable version"},
		{"1.2.3", false, "Stable version without v prefix"},
		{"v3.2.27", false, "Stable version from bug report"},
		{"v10.20.30", false, "Multi-digit stable version"},
		{"v1.0.0", false, "Zero patch version"},
		{"v1.2.0", false, "Zero patch version with minor"},

		// Dev versions (should return true)
		{"v1.2.3-dev", true, "Simple dev version"},
		{"v1.2.3-dev.abc123", true, "Dev version with hash"},
		{"v3.2.30-dev.adcb1ae", true, "Dev version from bug report"},
		{"v1.0.0-dev", true, "Dev version with zero patch"},
		{"dev", false, "Simple dev without version (no dash)"},
		{"v1.2.3-DEV", true, "Uppercase dev"},
		{"v1.2.3-Dev", true, "Mixed case dev"},

		// Beta versions (should return true)
		{"v1.2.3-beta", true, "Simple beta version"},
		{"v1.2.3-beta.2", true, "Beta with number"},
		{"v1.2.3-BETA", true, "Uppercase beta"},
		{"v1.2.3-Beta", true, "Mixed case beta"},

		// Release candidate versions (should return true)
		{"v1.2.3-rc", true, "Simple rc version"},
		{"v1.2.3-rc.1", true, "RC with number"},
		{"v1.2.3-rc.10", true, "RC with multi-digit number"},
		{"v1.2.3-RC", true, "Uppercase RC"},
		{"v1.2.3-Rc", true, "Mixed case RC"},

		// Alpha versions (should return true)
		{"v1.2.3-alpha", true, "Simple alpha version"},
		{"v1.2.3-alpha.2", true, "Alpha with number"},
		{"v1.2.3-ALPHA", true, "Uppercase alpha"},
		{"v1.2.3-Alpha", true, "Mixed case alpha"},

		// Pre versions (should return true)
		{"v1.2.3-pre", true, "Simple pre version"},
		{"v1.2.3-pre.1", true, "Pre with number"},
		{"v1.2.3-PRE", true, "Uppercase pre"},
		{"v1.2.3-Pre", true, "Mixed case pre"},

		// Edge cases
		{"", false, "Empty string"},
		{"v", false, "Just v prefix"},
		{"1.2.3-", false, "Trailing dash without indicator"},
		{"v1.2.3-snapshot", false, "Unknown prerelease type"},
		{"v1.2.3-test", false, "Unknown prerelease type"},
		{"v1.2.3-build", false, "Unknown prerelease type"},

		// Complex versions
		{"v1.2.3-dev.abc123+build.456", true, "Dev version with build metadata"},
		{"v1.2.3-beta.1+build.123", true, "Beta version with build metadata"},
		{"v1.2.3+build.456", false, "Stable version with build metadata"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := isPrereleaseTag(test.input)
			if result != test.expected {
				t.Errorf("isPrereleaseTag(%q) = %v, want %v (%s)",
					test.input, result, test.expected, test.description)
			}
		})
	}
}

func TestIsPrereleaseTagCaseSensitivity(t *testing.T) {
	// Test that the function is case insensitive
	testCases := []struct {
		prefix   string
		expected bool
	}{
		{"dev", true},
		{"DEV", true},
		{"Dev", true},
		{"dEv", true},
		{"beta", true},
		{"BETA", true},
		{"Beta", true},
		{"rc", true},
		{"RC", true},
		{"Rc", true},
		{"alpha", true},
		{"ALPHA", true},
		{"Alpha", true},
		{"pre", true},
		{"PRE", true},
		{"Pre", true},
		{"stable", false},
		{"release", false},
		{"final", false},
	}

	for _, tc := range testCases {
		version := "v1.2.3-" + tc.prefix
		result := isPrereleaseTag(version)
		if result != tc.expected {
			t.Errorf("isPrereleaseTag(%q) = %v, want %v", version, result, tc.expected)
		}
	}
}

func TestIsPrereleaseTagPosition(t *testing.T) {
	// Test that the prerelease indicator must be after a dash
	testCases := []struct {
		version     string
		expected    bool
		description string
	}{
		{"v1.2.3-dev", true, "Prerelease after dash"},
		{"v1.2.3dev", false, "Prerelease without dash"},
		{"v1.2-dev.3", true, "Prerelease in middle with dash"},
		{"v1.2dev.3", false, "Prerelease in middle without dash"},
		{"dev-v1.2.3", false, "Prerelease before version"},
		{"vdev1.2.3", false, "Prerelease embedded in version"},
		{"v1.2.3-dev-beta", true, "Multiple prerelease indicators"},
		{"v1.2.3-beta-dev", true, "Multiple prerelease indicators reversed"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := isPrereleaseTag(tc.version)
			if result != tc.expected {
				t.Errorf("isPrereleaseTag(%q) = %v, want %v", tc.version, result, tc.expected)
			}
		})
	}
}
