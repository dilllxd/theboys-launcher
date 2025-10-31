//go:build darwin
// +build darwin

package main

import (
	"syscall"
	"unsafe"
)

// macOS memory detection using sysctl
// This provides the same interface as the Windows version but uses macOS APIs

// DarwinMemoryInfo represents memory information on macOS
// Using a structure similar to Windows for compatibility
type DarwinMemoryInfo struct {
	TotalMemory     uint64
	AvailableMemory uint64
	MemoryLoad      uint32
}

// getSystemMemoryInfo returns detailed memory information for macOS
func getSystemMemoryInfo() (*DarwinMemoryInfo, error) {
	logf("DEBUG: Retrieving macOS system memory information")
	var totalMemory uint64
	var freeMemory uint64
	size := uint64(8)

	// Get total physical memory
	logf("DEBUG: Calling sysctl for hw.memsize")
	_, _, err := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&[]byte("hw.memsize")[0])),
		uintptr(len("hw.memsize")),
		uintptr(unsafe.Pointer(&totalMemory)),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if err != 0 {
		logf("DEBUG: Failed to get total memory via sysctl: %v", err)
		return nil, err
	}
	logf("DEBUG: Total memory detected: %d MB", totalMemory/(1024*1024))

	// Get free memory (vm page free count)
	var freeCount uint64
	size = uint64(8)
	logf("DEBUG: Calling sysctl for vm.page_free_count")
	_, _, err = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&[]byte("vm.page_free_count")[0])),
		uintptr(len("vm.page_free_count")),
		uintptr(unsafe.Pointer(&freeCount)),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if err != 0 {
		logf("DEBUG: Failed to get free memory via sysctl: %v", err)
		// If we can't get free memory, estimate 25% available
		freeMemory = totalMemory / 4
		logf("DEBUG: Using estimated free memory: %d MB (25%% of total)", freeMemory/(1024*1024))
	} else {
		// Convert page count to bytes (assuming 4KB pages)
		pageSize := uint64(4096)
		freeMemory = freeCount * pageSize
		logf("DEBUG: Free memory detected: %d MB (%d pages)", freeMemory/(1024*1024), freeCount)
	}

	// Calculate memory load percentage
	memoryLoad := uint32(((totalMemory - freeMemory) * 100) / totalMemory)
	logf("DEBUG: Memory load calculated: %d%%", memoryLoad)

	return &DarwinMemoryInfo{
		TotalMemory:     totalMemory,
		AvailableMemory: freeMemory,
		MemoryLoad:      memoryLoad,
	}, nil
}

// getAvailableMemoryMB returns available memory in MB for macOS
func getAvailableMemoryMB() int {
	logf("DEBUG: Getting available memory for macOS")
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		logf("DEBUG: Failed to get system memory info, using fallback 4GB: %v", err)
		// Fallback to 4GB if sysctl fails
		return 4096
	}

	availableMB := int(memInfo.AvailableMemory / (1024 * 1024))
	logf("DEBUG: Available memory detected on macOS: %d MB", availableMB)
	return availableMB
}

// validateMemoryResult ensures the memory value is reasonable on macOS
func validateMemoryResult(totalMB int) int {
	// Validate the result seems reasonable for macOS
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		return 16384 // Use 16GB default if result seems invalid
	}
	return totalMB
}
