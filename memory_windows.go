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
	var memInfo MEMORYSTATUSEX
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))

	kernel32 := windows.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GlobalMemoryStatusEx")

	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		return nil, err
	}

	return &memInfo, nil
}

// getAvailableMemoryMB returns available memory in MB for Windows
func getAvailableMemoryMB() int {
	memInfo, err := getSystemMemoryInfo()
	if err != nil {
		// Fallback to 4GB if API call fails
		return 4096
	}

	// Convert bytes to megabytes
	return int(memInfo.ullAvailPhys / (1024 * 1024))
}

// validateMemoryResult ensures the memory value is reasonable
func validateMemoryResult(totalMB int) int {
	// Validate the result seems reasonable
	if totalMB < 1024 || totalMB > 1024*1024 { // Less than 1GB or more than 1TB
		return 16384 // Use 16GB default if result seems invalid
	}
	return totalMB
}
