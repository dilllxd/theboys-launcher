//go:build linux
// +build linux

package main

import (
	"os"
	"strconv"
	"strings"
)

// Linux memory detection using /proc/meminfo
// This provides the same interface as the Windows/macOS versions but uses Linux /proc filesystem

// LinuxMemoryInfo represents memory information on Linux
type LinuxMemoryInfo struct {
	TotalMemory     uint64
	AvailableMemory uint64
	MemoryLoad      uint32
}

// getSystemMemoryInfo returns detailed memory information for Linux
func getSystemMemoryInfo() (*LinuxMemoryInfo, error) {
	debugf("Reading Linux memory information from /proc/meminfo")
	// Read memory information from /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		debugf("Failed to read /proc/meminfo: %v", err)
		return nil, err
	}

	var totalMemory, availableMemory uint64
	lines := strings.Split(string(data), "\n")
	debugf("Parsing %d lines from /proc/meminfo", len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			debugf("Failed to parse memory value from line: %s", line)
			continue
		}

		// Convert from KB to bytes
		valueBytes := value * 1024

		switch fields[0] {
		case "MemTotal:":
			totalMemory = valueBytes
			debugf("Found total memory: %d KB (%d MB)", value, valueBytes/(1024*1024))
		case "MemAvailable:":
			availableMemory = valueBytes
			debugf("Found available memory: %d KB (%d MB)", value, valueBytes/(1024*1024))
		}
	}

	// Calculate memory load percentage
	var memoryLoad uint32
	if totalMemory > 0 {
		usedMemory := totalMemory - availableMemory
		memoryLoad = uint32((usedMemory * 100) / totalMemory)
		debugf("Memory load calculated: %d%% (used: %d MB, available: %d MB)",
			memoryLoad, usedMemory/(1024*1024), availableMemory/(1024*1024))
	}

	return &LinuxMemoryInfo{
		TotalMemory:     totalMemory,
		AvailableMemory: availableMemory,
		MemoryLoad:      memoryLoad,
	}, nil
}

// getAvailableMemoryMB returns available memory in MB for Linux
func getAvailableMemoryMB() int {
	debugf("Getting available memory for Linux")
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		debugf("Failed to get system memory info, using fallback 4GB: %v", err)
		// Fallback to 4GB if we can't read meminfo
		return 4096
	}

	availableMB := int(memInfo.AvailableMemory / (1024 * 1024))
	debugf("Available memory detected on Linux: %d MB", availableMB)
	return availableMB
}

// validateMemoryResult ensures the memory value is reasonable on Linux
func validateMemoryResult(totalMB int) int {
	// Validate the result seems reasonable for Linux
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		return 16384 // Use 16GB default if result seems invalid
	}
	return totalMB
}
