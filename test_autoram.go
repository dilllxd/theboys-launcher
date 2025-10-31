package main

import (
	"fmt"
)

// Simulate the AutoRAM calculation functions from config.go
func clampMemoryMB(mb int) int {
	if mb < 2048 {
		return 2048
	}
	if mb > 16384 {
		return 16384
	}
	return mb
}

func DefaultAutoMemoryMB(totalRAM int) int {
	// If total memory detection fails, fall back to 32GB
	if totalRAM <= 0 {
		totalRAM = 32768 // fallback 32GB
	}
	
	// Calculate half of total memory and clamp to 2-16GB range
	auto := clampMemoryMB(totalRAM / 2)
	return auto
}

func main() {
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

	fmt.Println("AutoRAM Calculation Test Results:")
	fmt.Println("=================================")
	
	for _, tc := range testCases {
		totalRAMMB := tc.systemRAMGB * 1024
		autoRAMMB := DefaultAutoMemoryMB(totalRAMMB)
		autoRAMGB := autoRAMMB / 1024
		
		status := "✓ PASS"
		if autoRAMGB != tc.expectedGB {
			status = "✗ FAIL"
		}
		
		fmt.Printf("%s %d GB system → AutoRAM %d GB (expected %d GB)\n", 
			status, tc.systemRAMGB, autoRAMGB, tc.expectedGB)
	}
	
	// Test edge cases
	fmt.Println("\nEdge Case Tests:")
	fmt.Println("================")
	
	// Test with 0 (failed detection)
	autoRAMMB := DefaultAutoMemoryMB(0)
	fmt.Printf("Failed detection (0) → AutoRAM %d GB (fallback to 16GB)\n", autoRAMMB/1024)
	
	// Test with negative value
	autoRAMMB = DefaultAutoMemoryMB(-1024)
	fmt.Printf("Invalid detection (-1GB) → AutoRAM %d GB (fallback to 16GB)\n", autoRAMMB/1024)
	
	// Test minimum boundary
	autoRAMMB = DefaultAutoMemoryMB(2048) // 2GB system
	fmt.Printf("2 GB system → AutoRAM %d GB (minimum)\n", autoRAMMB/1024)
	
	// Test maximum boundary
	autoRAMMB = DefaultAutoMemoryMB(32768) // 32GB system
	fmt.Printf("32 GB system → AutoRAM %d GB (maximum)\n", autoRAMMB/1024)
}