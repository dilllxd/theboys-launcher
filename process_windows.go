//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// Windows process management using taskkill
// This provides the same interface as the macOS version but uses Windows APIs

// killProcessByName kills all processes with the given name on Windows
func killProcessByName(processName string) error {
	cmd := exec.Command("taskkill", "/F", "/IM", processName)
	return cmd.Run()
}

// killProcessByPID kills a process and its children by PID on Windows
func killProcessByPID(pid int) error {
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	return cmd.Run()
}

// killPrismProcesses kills all Prism Launcher processes on Windows
func killPrismProcesses() error {
	return killProcessByName("PrismLauncher.exe")
}

// killJavaProcesses kills all Java processes on Windows
func killJavaProcesses() error {
	return killProcessByName("java.exe")
}

// forceCloseAllProcesses force-closes all game-related processes on Windows
func forceCloseAllProcesses(prismProcess *os.Process) error {
	logf("%s", warnLine("Force-closing Prism Launcher and Minecraft processes..."))

	// Force close all Prism processes
	if err := killPrismProcesses(); err != nil {
		logf("Warning: Failed to kill Prism processes: %v", err)
	}

	// Force close any Java processes (likely Minecraft)
	if err := killJavaProcesses(); err != nil {
		logf("Warning: Failed to kill Java processes: %v", err)
	}

	// Also close the specific Prism process we launched if we have it
	if prismProcess != nil && prismProcess.Pid > 0 {
		if err := killProcessByPID(prismProcess.Pid); err != nil {
			logf("Warning: Failed to kill Prism process %d: %v", prismProcess.Pid, err)
		} else {
			logf("Force-closed Prism process %d and related processes", prismProcess.Pid)
		}
	}

	logf("All game processes force-closed")
	return nil
}

// getProcessName returns the platform-specific process name
func getProcessName(baseName string) string {
	if runtime.GOOS == "windows" {
		return baseName + ".exe"
	}
	return baseName
}
