//go:build windows
// +build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows memory detection using GlobalMemoryStatusEx API
// This replaces the memory detection that was previously in config.go

// Windows memory status structure
type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

// getSystemMemoryInfo returns detailed memory information for Windows
func getSystemMemoryInfo() (*MEMORYSTATUSEX, error) {
	debugf("Retrieving Windows system memory information")
	var memInfo MEMORYSTATUSEX
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))

	kernel32 := windows.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GlobalMemoryStatusEx")

	debugf("Calling GlobalMemoryStatusEx API")
	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		debugf("GlobalMemoryStatusEx failed: %v", err)
		return nil, err
	}

	debugf("Memory info retrieved - Total: %d MB, Available: %d MB, Load: %d%%",
		memInfo.ullTotalPhys/(1024*1024),
		memInfo.ullAvailPhys/(1024*1024),
		memInfo.dwMemoryLoad)
	return &memInfo, nil
}

// getAvailableMemoryMB returns available memory in MB for Windows
func getAvailableMemoryMB() int {
	debugf("Getting available memory for Windows")
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		debugf("Failed to get system memory info, using fallback 4GB: %v", err)
		// Fallback to 4GB if API call fails
		return 4096
	}

	availableMB := int(memInfo.ullAvailPhys / (1024 * 1024))
	debugf("Available memory detected: %d MB", availableMB)
	return availableMB
}

// validateMemoryResult ensures the memory value is reasonable
func validateMemoryResult(totalMB int) int {
	debugf("Validating memory result: %d MB", totalMB)
	// Validate the result seems reasonable
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		logf("DEBUG: Memory result %d MB seems unreasonable, using 16GB fallback", totalMB)
		return 16384 // Use 16GB default if result seems invalid
	}
	logf("DEBUG: Memory result %d MB is valid", totalMB)
	return totalMB
}
