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
	var totalMemory uint64
	var freeMemory uint64
	size := uint64(8)

	// Get total physical memory
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
		return nil, err
	}

	// Get free memory (vm page free count)
	var freeCount uint64
	size = uint64(8)
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
		// If we can't get free memory, estimate 25% available
		freeMemory = totalMemory / 4
	} else {
		// Convert page count to bytes (assuming 4KB pages)
		pageSize := uint64(4096)
		freeMemory = freeCount * pageSize
	}

	// Calculate memory load percentage
	memoryLoad := uint32(((totalMemory - freeMemory) * 100) / totalMemory)

	return &DarwinMemoryInfo{
		TotalMemory:     totalMemory,
		AvailableMemory: freeMemory,
		MemoryLoad:      memoryLoad,
	}, nil
}

// getAvailableMemoryMB returns available memory in MB for macOS
func getAvailableMemoryMB() int {
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		// Fallback to 4GB if sysctl fails
		return 4096
	}

	// Convert bytes to megabytes
	return int(memInfo.AvailableMemory / (1024 * 1024))
}

// validateMemoryResult ensures the memory value is reasonable on macOS
func validateMemoryResult(totalMB int) int {
	// Validate the result seems reasonable for macOS
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		return 16384 // Use 16GB default if result seems invalid
	}
	return totalMB
}