//go:build linux
// +build linux

package main

import (
	"os"
	"os/exec"
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
func killProcessByName(processName string) error {
	// Use pkill to kill processes by name
	cmd := exec.Command("pkill", "-f", processName)
	return cmd.Run()
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