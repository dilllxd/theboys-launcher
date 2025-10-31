//go:build linux
// +build linux

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

// Linux process management using pkill and kill
// This provides the same interface as the Windows/macOS versions but uses Linux commands

// killProcessByName kills all processes with the given name on Linux
func killProcessByName(processName string) error {
	// Use pkill to kill processes by name
	cmd := exec.Command("pkill", "-f", processName)
	return cmd.Run()
}

// killProcessByPID kills a process and its children by PID on Linux
func killProcessByPID(pid int) error {
	// First try to kill the process tree using pkill with parent PID
	cmd := exec.Command("pkill", "-P", strconv.Itoa(pid))
	err := cmd.Run()

	// Also kill the specific process directly
	killCmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	if killErr := killCmd.Run(); killErr != nil {
		// If kill fails, try to use pkill on the PID
		pkillCmd := exec.Command("pkill", strconv.Itoa(pid))
		return pkillCmd.Run()
	}

	return err
}

// killPrismProcesses kills all Prism Launcher processes on Linux
func killPrismProcesses() error {
	// Kill Prism launcher processes
	processes := []string{
		"PrismLauncher",
		"prismlauncher",
		"prism-launcher",
	}

	for _, process := range processes {
		if err := killProcessByName(process); err != nil {
			logf("Warning: Failed to kill Prism process pattern '%s': %v", process, err)
		}
	}

	return nil
}

// killJavaProcesses kills all Java processes on Linux
func killJavaProcesses() error {
	// Kill Java processes but be more selective to avoid system Java processes
	cmd := exec.Command("pgrep", "-f", "java")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if len(output) == 0 {
		return nil // No Java processes found
	}

	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, pid := range pids {
		if pid != "" {
			// Check if this is a Minecraft-related Java process
			if isMinecraftJavaProcess(pid) {
				killCmd := exec.Command("kill", "-9", pid)
				killCmd.Run()
			}
		}
	}

	return nil
}

// isMinecraftJavaProcess checks if a Java process is likely Minecraft-related
func isMinecraftJavaProcess(pid string) bool {
	// Get the command line for the process
	cmd := exec.Command("ps", "-p", pid, "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	command := string(output)
	// Check for Minecraft indicators
	minecraftIndicators := []string{
		"minecraft",
		".minecraft",
		"PrismLauncher",
		"MultiMC",
		"forge",
		"fabric",
		"quilt",
		"prismlauncher",
	}

	commandLower := strings.ToLower(command)
	for _, indicator := range minecraftIndicators {
		if strings.Contains(commandLower, indicator) {
			return true
		}
	}

	return false
}

// forceCloseAllProcesses force-closes all game-related processes on Linux
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

// isProcessRunning checks if a process with the given PID is still running on Linux
func isProcessRunning(pid int) (bool, error) {
	// Use ps to check if the process is still running
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid=")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check process status: %w", err)
	}

	// ps returns the PID if the process exists, empty otherwise
	return strings.TrimSpace(string(output)) == strconv.Itoa(pid), nil
}

// validateProcessIdentity validates that a process matches the expected executable and working directory on Linux
func validateProcessIdentity(pid int, expectedExecutable, expectedWorkingDir string) (bool, error) {
	// Get process information using ps
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=", "-o", "cwd=")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get process information: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 1 {
		return false, fmt.Errorf("process not found")
	}

	// Parse ps output (format: command\ncwd)
	parts := strings.Split(lines[0], "\n")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid process information format")
	}

	actualCommand := strings.TrimSpace(parts[0])
	actualWorkingDir := strings.TrimSpace(parts[1])

	// Normalize paths for comparison
	expectedExecutable = strings.ToLower(expectedExecutable)
	actualCommand = strings.ToLower(actualCommand)

	// Check if executable names match (allowing for case differences)
	if !strings.Contains(actualCommand, filepath.Base(expectedExecutable)) {
		return false, nil
	}

	// Check working directory if provided
	if expectedWorkingDir != "" {
		expectedWorkingDir = strings.ToLower(expectedWorkingDir)
		actualWorkingDir = strings.ToLower(actualWorkingDir)

		// Normalize paths for comparison
		if !strings.EqualFold(filepath.Clean(actualWorkingDir), filepath.Clean(expectedWorkingDir)) {
			// Try to match by basename if full paths differ
			if filepath.Base(actualWorkingDir) != filepath.Base(expectedWorkingDir) {
				return false, nil
			}
		}
	}

	return true, nil
}

// getProcessDetails retrieves detailed information about a process on Linux
func getProcessDetails(pid int) (executable, workingDir string, err error) {
	// Get executable path using /proc filesystem
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	if executableBytes, err := os.Readlink(exePath); err == nil {
		executable = executableBytes
	} else {
		// Fallback to ps command
		cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
		output, psErr := cmd.Output()
		if psErr != nil {
			return "", "", fmt.Errorf("failed to get process executable: %w", psErr)
		}

		// Extract executable from command line
		command := strings.TrimSpace(string(output))
		parts := strings.Fields(command)
		if len(parts) > 0 {
			executable = parts[0]
		}
	}

	// Get working directory using /proc filesystem
	cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
	if workingDirBytes, err := os.Readlink(cwdPath); err == nil {
		workingDir = workingDirBytes
	} else {
		// Fallback to ps command
		cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "cwd=")
		output, psErr := cmd.Output()
		if psErr != nil {
			// If we can't get working directory, use executable directory as fallback
			if executable != "" {
				workingDir = filepath.Dir(executable)
			}
			logf("Warning: Could not get working directory for PID %d, using executable directory: %v", pid, psErr)
		} else {
			workingDir = strings.TrimSpace(string(output))
		}
	}

	return executable, workingDir, nil
}
