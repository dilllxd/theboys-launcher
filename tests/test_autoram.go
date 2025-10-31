package main

import (
	"testing"
)

// TestAutoRAMCalculation tests the AutoRAM calculation functionality
func TestAutoRAMCalculation(t *testing.T) {
	// Test cases for different system memory configurations
	testCases := []struct {
		systemRAMGB int
		expectedGB  int
	}{
		{4, 2},   // 4 GB system → AutoRAM 2 GB
		{8, 4},   // 8 GB system → AutoRAM 4 GB
		{16, 8},  // 16 GB system → AutoRAM 8 GB
		{32, 16}, // 32 GB system → AutoRAM 16 GB
		{64, 16}, // 64+ GB system → AutoRAM 16 GB (maxed out)
		{1, 2},   // 1 GB system → AutoRAM 2 GB (minimum)
		{128, 16}, // 128 GB system → AutoRAM 16 GB (maxed out)
	}

	for _, tc := range testCases {
		totalRAMMB := tc.systemRAMGB * 1024
		autoRAMMB := DefaultAutoMemoryMBWithTotal(totalRAMMB)
		autoRAMGB := autoRAMMB / 1024
		
		if autoRAMGB != tc.expectedGB {
			t.Errorf("FAIL: %d GB system → AutoRAM %d GB (expected %d GB)", 
				tc.systemRAMGB, autoRAMGB, tc.expectedGB)
		} else {
			t.Logf("PASS: %d GB system → AutoRAM %d GB", 
				tc.systemRAMGB, autoRAMGB)
		}
	}
}

// TestAutoRAMEdgeCases tests edge cases for AutoRAM calculation
func TestAutoRAMEdgeCases(t *testing.T) {
	// Test with 0 (failed detection)
	autoRAMMB := DefaultAutoMemoryMBWithTotal(0)
	expectedGB := 16 // fallback to 16GB
	if autoRAMMB/1024 != expectedGB {
		t.Errorf("FAIL: Failed detection (0) → AutoRAM %d GB (expected %d GB)", 
			autoRAMMB/1024, expectedGB)
	} else {
		t.Logf("PASS: Failed detection (0) → AutoRAM %d GB", autoRAMMB/1024)
	}
	
	// Test with negative value
	autoRAMMB = DefaultAutoMemoryMBWithTotal(-1024)
	if autoRAMMB/1024 != expectedGB {
		t.Errorf("FAIL: Invalid detection (-1GB) → AutoRAM %d GB (expected %d GB)", 
			autoRAMMB/1024, expectedGB)
	} else {
		t.Logf("PASS: Invalid detection (-1GB) → AutoRAM %d GB", autoRAMMB/1024)
	}
	
	// Test minimum boundary
	autoRAMMB = DefaultAutoMemoryMBWithTotal(2048) // 2GB system
	expectedGB = 2
	if autoRAMMB/1024 != expectedGB {
		t.Errorf("FAIL: 2 GB system → AutoRAM %d GB (expected %d GB)", 
			autoRAMMB/1024, expectedGB)
	} else {
		t.Logf("PASS: 2 GB system → AutoRAM %d GB", autoRAMMB/1024)
	}
	
	// Test maximum boundary
	autoRAMMB = DefaultAutoMemoryMBWithTotal(32768) // 32GB system
	expectedGB = 16
	if autoRAMMB/1024 != expectedGB {
		t.Errorf("FAIL: 32 GB system → AutoRAM %d GB (expected %d GB)", 
			autoRAMMB/1024, expectedGB)
	} else {
		t.Logf("PASS: 32 GB system → AutoRAM %d GB", autoRAMMB/1024)
	}
}

// DefaultAutoMemoryMBWithTotal is a helper function for testing that accepts a total RAM parameter
// This avoids conflicts with the actual DefaultAutoMemoryMB() function in config.go
func DefaultAutoMemoryMBWithTotal(totalRAM int) int {
	// If total memory detection fails, fall back to 32GB
	if totalRAM <= 0 {
		totalRAM = 32768 // fallback 32GB
	}
	
	// Calculate half of total memory and clamp to 2-16GB range
	auto := clampMemoryMB(totalRAM / 2)
	return auto
}