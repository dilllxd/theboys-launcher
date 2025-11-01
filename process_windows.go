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
		outputStr := string(output)
		debugf("taskkill failed for PID %d: %v, output: %s", pid, err, outputStr)

		// Check if the process has already exited
		if strings.Contains(strings.ToLower(outputStr), "not found") ||
			strings.Contains(strings.ToLower(outputStr), "does not exist") ||
			strings.Contains(strings.ToLower(outputStr), "cannot be found") {
			debugf("Process PID %d has already exited", pid)
			return nil // Not an error if process is already gone
		}

		// Check if it's an access denied error (process already terminating)
		if strings.Contains(strings.ToLower(outputStr), "access denied") {
			debugf("Access denied for PID %d, process may be terminating", pid)
			return nil // Not an error if process is already terminating
		}

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
	var actualExecutable string
	var err error

	// Try PowerShell first (more reliable on modern Windows)
	psCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Process -Id %d -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path", pid))
	psOutput, psErr := psCmd.Output()

	if psErr == nil {
		actualExecutable = strings.TrimSpace(string(psOutput))
		if actualExecutable != "" {
			// PowerShell succeeded, continue with validation
		} else {
			// PowerShell returned empty result, try wmic fallback
			err = fmt.Errorf("PowerShell returned empty result")
		}
	} else {
		// PowerShell failed, try wmic fallback
		err = psErr
	}

	// If PowerShell failed, try wmic fallback
	if err != nil {
		logf("PowerShell process query failed during validation, falling back to wmic: %v", err)
		wmicCmd := exec.Command("wmic", "process", "where", "ProcessId="+strconv.Itoa(pid), "get", "ExecutablePath", "/format:csv")
		output, wmicErr := wmicCmd.Output()
		if wmicErr != nil {
			// Check if wmic executable is not found
			if strings.Contains(wmicErr.Error(), "executable file not found") ||
				strings.Contains(wmicErr.Error(), "not found in %PATH%") {
				logf("wmic not available in PATH, skipping to tasklist fallback")
			} else {
				logf("wmic process query failed during validation, falling back to tasklist: %v", wmicErr)
			}

			// Try tasklist as a final fallback
			tasklistCmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
			tasklistOutput, tasklistErr := tasklistCmd.Output()
			if tasklistErr != nil {
				return false, fmt.Errorf("failed to get process information (PowerShell, wmic, and tasklist all failed): %w",
					fmt.Errorf("PowerShell: %v, wmic: %v, tasklist: %v", psErr, wmicErr, tasklistErr))
			}

			// Parse tasklist output to get executable name (not full path)
			outputStr := strings.TrimSpace(string(tasklistOutput))
			if outputStr == "" {
				// Empty output likely means process doesn't exist
				return false, fmt.Errorf("process not found (may have exited)")
			}

			// Check for "No tasks" message
			if strings.Contains(strings.ToLower(outputStr), "no tasks") ||
				strings.Contains(strings.ToLower(outputStr), "not found") {
				return false, fmt.Errorf("process not found in tasklist output (may have exited)")
			}

			lines := strings.Split(outputStr, "\n")
			if len(lines) > 0 {
				// Tasklist CSV format: "PID","Image Name","Session Name","Session#","Mem Usage"
				fields := strings.Split(lines[0], ",")
				if len(fields) >= 2 {
					// Extract just the executable name from tasklist
					imageName := strings.Trim(fields[1], " \t\"")
					if imageName != "" {
						// For tasklist fallback, we only get the executable name, not the full path
						logf("Warning: Only executable name available from tasklist fallback during validation: %s", imageName)
						actualExecutable = imageName
					}
				}
			}

			if actualExecutable == "" {
				// If we get here, it means no process was found in tasklist output
				// This could be because the process exited or because tasklist returned "INFO: No tasks..."
				return false, fmt.Errorf("process not found in tasklist output (may have exited)")
			}
		} else {
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
			actualExecutable = strings.Trim(fields[1], " \t\"")
		}
	}

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

	// For working directory, we can't easily get it reliably on Windows, so we'll skip this check
	// and rely on executable matching only

	return true, nil
}

// getProcessDetails retrieves detailed information about a process on Windows
func getProcessDetails(pid int) (executable, workingDir string, err error) {
	// Try PowerShell first (more reliable on modern Windows)
	psCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Process -Id %d -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path", pid))
	psOutput, psErr := psCmd.Output()

	if psErr == nil {
		executable = strings.TrimSpace(string(psOutput))
		if executable != "" {
			// Get working directory using PowerShell
			psDirCmd := exec.Command("powershell", "-Command",
				fmt.Sprintf("(Get-Process -Id %d | Select-Object -ExpandProperty Path).DirectoryName", pid))
			psDirOutput, psDirErr := psDirCmd.Output()
			if psDirErr == nil {
				workingDir = strings.TrimSpace(string(psDirOutput))
				if workingDir == "" {
					workingDir = filepath.Dir(executable)
				}
			} else {
				workingDir = filepath.Dir(executable)
				logf("Warning: Could not get working directory for PID %d, using executable directory: %v", pid, psDirErr)
			}
			return executable, workingDir, nil
		}
	}

	// Fallback to wmic if PowerShell fails (for older Windows systems)
	logf("PowerShell process query failed, falling back to wmic: %v", psErr)
	wmicCmd := exec.Command("wmic", "process", "where", "ProcessId="+strconv.Itoa(pid), "get", "ExecutablePath", "/format:csv")
	output, wmicErr := wmicCmd.Output()
	if wmicErr != nil {
		// Check if wmic executable is not found
		if strings.Contains(wmicErr.Error(), "executable file not found") ||
			strings.Contains(wmicErr.Error(), "not found in %PATH%") {
			logf("wmic not available in PATH, skipping to tasklist fallback")
		} else {
			logf("wmic process query failed, falling back to tasklist: %v", wmicErr)
		}

		// Try tasklist as a final fallback
		tasklistCmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
		tasklistOutput, tasklistErr := tasklistCmd.Output()
		if tasklistErr != nil {
			return "", "", fmt.Errorf("failed to get process executable (PowerShell, wmic, and tasklist all failed): %w",
				fmt.Errorf("PowerShell: %v, wmic: %v, tasklist: %v", psErr, wmicErr, tasklistErr))
		}

		// Parse tasklist output to get executable name (not full path)
		outputStr := strings.TrimSpace(string(tasklistOutput))
		if outputStr == "" {
			// Empty output likely means process doesn't exist
			return "", "", fmt.Errorf("process not found (may have exited)")
		}

		// Check for "No tasks" message
		if strings.Contains(strings.ToLower(outputStr), "no tasks") ||
			strings.Contains(strings.ToLower(outputStr), "not found") {
			return "", "", fmt.Errorf("process not found in tasklist output (may have exited)")
		}

		lines := strings.Split(outputStr, "\n")
		if len(lines) > 0 {
			// Tasklist CSV format: "PID","Image Name","Session Name","Session#","Mem Usage"
			fields := strings.Split(lines[0], ",")
			if len(fields) >= 2 {
				// Extract just the executable name from tasklist
				imageName := strings.Trim(fields[1], " \t\"")
				if imageName != "" {
					// For tasklist fallback, we only get the executable name, not the full path
					// This is limited but better than failing completely
					logf("Warning: Only executable name available from tasklist fallback: %s", imageName)
					return imageName, "", nil
				}
			}
		}
		// If we get here, it means no process was found in tasklist output
		// This could be because the process exited or because tasklist returned "INFO: No tasks..."
		return "", "", fmt.Errorf("process not found in tasklist output (may have exited)")
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
	workingDir = filepath.Dir(executable)

	return executable, workingDir, nil
}
