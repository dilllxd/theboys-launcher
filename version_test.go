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
