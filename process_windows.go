//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Windows process management using taskkill
// This provides the same interface as the macOS version but uses Windows APIs

// killProcessByName kills all processes with the given name on Windows
func killProcessByName(processName string) error {
	debugf("Attempting to kill processes by name: %s", processName)
	cmd := exec.Command("taskkill", "/F", "/IM", processName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		debugf("taskkill failed for %s: %v, output: %s", processName, err, string(output))
		return err
	}
	debugf("Successfully killed processes named %s, output: %s", processName, string(output))
	return nil
}

// killProcessByPID kills a process and its children by PID on Windows
func killProcessByPID(pid int) error {
	debugf("Attempting to kill process tree for PID %d", pid)
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		debugf("taskkill failed for PID %d: %v, output: %s", pid, err, string(output))
		return err
	}
	debugf("Successfully killed process tree for PID %d, output: %s", pid, string(output))
	return nil
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

// isProcessRunning checks if a process with the given PID is still running on Windows
func isProcessRunning(pid int) (bool, error) {
	debugf("Checking if process PID %d is running", pid)
	// Use tasklist to check if the process is still running
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		debugf("Failed to check process status for PID %d: %v", pid, err)
		return false, fmt.Errorf("failed to check process status: %w", err)
	}

	outputStr := string(output)
	isRunning := len(strings.TrimSpace(outputStr)) > 0
	debugf("Process PID %d running status: %t (output: %s)", pid, isRunning, outputStr)
	return isRunning, nil
}

// validateProcessIdentity validates that a process matches the expected executable and working directory on Windows
func validateProcessIdentity(pid int, expectedExecutable, expectedWorkingDir string) (bool, error) {
	// Get process information using wmic
	cmd := exec.Command("wmic", "process", "where", "ProcessId="+strconv.Itoa(pid), "get", "ExecutablePath,WorkingSetSize", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get process information: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return false, fmt.Errorf("process not found")
	}

	// Parse CSV output (skip header line)
	fields := strings.Split(lines[1], ",")
	if len(fields) < 2 {
		return false, fmt.Errorf("invalid process information format")
	}

	// Get executable path from wmic output
	actualExecutable := strings.Trim(fields[1], " \t\"")

	// Normalize paths for comparison
	actualExecutable = strings.ToLower(actualExecutable)
	expectedExecutable = strings.ToLower(expectedExecutable)

	// Check if executable paths match (allowing for case differences)
	if !strings.EqualFold(actualExecutable, expectedExecutable) {
		// Also check basename in case full paths differ
		actualBase := filepath.Base(actualExecutable)
		expectedBase := filepath.Base(expectedExecutable)
		if !strings.EqualFold(actualBase, expectedBase) {
			return false, nil
		}
	}

	// For working directory, we can't easily get it from wmic, so we'll skip this check on Windows
	// and rely on executable matching only

	return true, nil
}

// getProcessDetails retrieves detailed information about a process on Windows
func getProcessDetails(pid int) (executable, workingDir string, err error) {
	// Get executable path using wmic
	cmd := exec.Command("wmic", "process", "where", "ProcessId="+strconv.Itoa(pid), "get", "ExecutablePath", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get process executable: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("process not found")
	}

	// Parse CSV output (skip header line)
	fields := strings.Split(lines[1], ",")
	if len(fields) < 2 {
		return "", "", fmt.Errorf("invalid process information format")
	}

	executable = strings.Trim(fields[1], " \t\"")

	// Get working directory using PowerShell (more reliable than wmic for this)
	psCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("(Get-Process -Id %d | Select-Object -ExpandProperty Path).DirectoryName", pid))
	psOutput, psErr := psCmd.Output()
	if psErr != nil {
		// Fallback: use directory of executable
		workingDir = filepath.Dir(executable)
		logf("Warning: Could not get working directory for PID %d, using executable directory: %v", pid, psErr)
	} else {
		workingDir = strings.TrimSpace(string(psOutput))
		if workingDir == "" {
			workingDir = filepath.Dir(executable)
		}
	}

	return executable, workingDir, nil
}
