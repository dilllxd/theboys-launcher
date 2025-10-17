package launcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/platform"
)

// ProcessManager handles process management and signal handling
type ProcessManager struct {
	platform         platform.Platform
	logger           logging.Logger
	childProcesses   []*ManagedProcess
	shutdownChan     chan os.Signal
	cleanupCompleted bool
}

// ManagedProcess represents a managed child process
type ManagedProcess struct {
	Name     string
	PID      int
	Cmd      interface{} // Could be *exec.Cmd or other process types
	Started  time.Time
	KillFunc func() error
}

// NewProcessManager creates a new process manager
func NewProcessManager(platform platform.Platform, logger logging.Logger) *ProcessManager {
	return &ProcessManager{
		platform:       platform,
		logger:         logger,
		childProcesses: make([]*ManagedProcess, 0),
		shutdownChan:   make(chan os.Signal, 1),
	}
}

// SetupSignalHandling sets up signal handlers for graceful shutdown
func (pm *ProcessManager) SetupSignalHandling() {
	pm.logger.Info("Setting up signal handling for graceful shutdown")

	// Handle different signals based on platform
	if runtime.GOOS == "windows" {
		// Windows-specific signals
		signal.Notify(pm.shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	} else {
		// Unix-like systems
		signal.Notify(pm.shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	}

	// Start signal handler goroutine
	go pm.handleSignals()
}

// handleSignals processes incoming shutdown signals
func (pm *ProcessManager) handleSignals() {
	sig := <-pm.shutdownChan
	pm.logger.Info("Received shutdown signal: %v", sig)

	// Perform graceful shutdown
	pm.GracefulShutdown()
}

// GracefulShutdown performs a graceful shutdown of all child processes
func (pm *ProcessManager) GracefulShutdown() {
	pm.logger.Info("Starting graceful shutdown process")

	if pm.cleanupCompleted {
		pm.logger.Debug("Cleanup already completed")
		return
	}

	// Give processes time to exit gracefully
	gracefulTimeout := 10 * time.Second
	forceTimeout := 5 * time.Second

	// Step 1: Try to stop processes gracefully
	pm.logger.Info("Attempting graceful shutdown of %d child processes", len(pm.childProcesses))

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	for _, proc := range pm.childProcesses {
		if proc.KillFunc != nil {
			select {
			case <-ctx.Done():
				goto forceShutdown
			default:
				pm.logger.Debug("Attempting graceful shutdown of process: %s (PID: %d)", proc.Name, proc.PID)
				if err := proc.KillFunc(); err != nil {
					pm.logger.Warn("Failed to gracefully shutdown process %s: %v", proc.Name, err)
				}
				// Give process some time to exit
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

forceShutdown:
	// Step 2: Force kill remaining processes if necessary
	if len(pm.childProcesses) > 0 {
		pm.logger.Warn("Force killing remaining processes after timeout")
		forceCtx, forceCancel := context.WithTimeout(context.Background(), forceTimeout)
		defer forceCancel()

		for _, proc := range pm.childProcesses {
			select {
			case <-forceCtx.Done():
				break
			default:
				pm.forceKillProcess(proc)
			}
		}
	}

	// Step 3: Perform additional cleanup
	pm.performFinalCleanup()

	pm.cleanupCompleted = true
	pm.logger.Info("Graceful shutdown completed")
}

// forceKillProcess forcefully kills a process
func (pm *ProcessManager) forceKillProcess(proc *ManagedProcess) {
	if proc.PID <= 0 {
		return
	}

	pm.logger.Debug("Force killing process: %s (PID: %d)", proc.Name, proc.PID)

	if runtime.GOOS == "windows" {
		// Windows: Use taskkill command
		if err := pm.forceKillWindowsProcess(proc.PID); err != nil {
			pm.logger.Warn("Failed to force kill Windows process %d: %v", proc.PID, err)
		}
	} else {
		// Unix-like systems: Use SIGKILL
		if err := syscall.Kill(proc.PID, syscall.SIGKILL); err != nil {
			pm.logger.Warn("Failed to force kill Unix process %d: %v", proc.PID, err)
		}
	}
}

// forceKillWindowsProcess uses taskkill to force kill a Windows process
func (pm *ProcessManager) forceKillWindowsProcess(pid int) error {
	// Find and kill the process tree (parent and children)
	if err := pm.runCommand("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid)); err != nil {
		return err
	}
	return nil
}

// runCommand executes a system command
func (pm *ProcessManager) runCommand(name string, args ...string) error {
	// This would use the platform abstraction to run commands
	// For now, just log the command
	pm.logger.Debug("Running command: %s %v", name, args)
	return nil
}

// RegisterProcess registers a child process for management
func (pm *ProcessManager) RegisterProcess(name string, pid int, killFunc func() error) {
	proc := &ManagedProcess{
		Name:     name,
		PID:      pid,
		Started:  time.Now(),
		KillFunc: killFunc,
	}

	pm.childProcesses = append(pm.childProcesses, proc)
	pm.logger.Info("Registered child process: %s (PID: %d)", name, pid)
}

// UnregisterProcess removes a child process from management
func (pm *ProcessManager) UnregisterProcess(pid int) {
	for i, proc := range pm.childProcesses {
		if proc.PID == pid {
			pm.childProcesses = append(pm.childProcesses[:i], pm.childProcesses[i+1:]...)
			pm.logger.Info("Unregistered child process: %s (PID: %d)", proc.Name, pid)
			return
		}
	}
}

// GetActiveProcesses returns the list of active child processes
func (pm *ProcessManager) GetActiveProcesses() []*ManagedProcess {
	return pm.childProcesses
}

// IsProcessRunning checks if a process is still running
func (pm *ProcessManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	if runtime.GOOS == "windows" {
		// Windows: Check using tasklist
		return pm.checkWindowsProcess(pid)
	} else {
		// Unix-like systems: Send signal 0 (doesn't kill the process)
		err := syscall.Kill(pid, 0)
		return err == nil
	}
}

// checkWindowsProcess checks if a Windows process is running
func (pm *ProcessManager) checkWindowsProcess(pid int) bool {
	// This would use tasklist or other Windows-specific methods
	// For now, assume it's running
	return true
}

// CleanupMinecraftProcesses finds and terminates Minecraft/PrismLauncher processes
func (pm *ProcessManager) CleanupMinecraftProcesses() error {
	pm.logger.Info("Cleaning up Minecraft and PrismLauncher processes")

	// Common Minecraft-related process names
	processNames := []string{
		"java",           // Minecraft Java process
		"prism-launcher", // PrismLauncher
		"minecraft",      // Minecraft (if running directly)
		"javaw",          // Windows Java without console
	}

	if runtime.GOOS == "windows" {
		return pm.cleanupWindowsProcesses(processNames)
	} else {
		return pm.cleanupUnixProcesses(processNames)
	}
}

// cleanupWindowsProcesses cleans up Windows processes
func (pm *ProcessManager) cleanupWindowsProcesses(processNames []string) error {
	pm.logger.Debug("Cleaning up Windows processes")

	for _, name := range processNames {
		// Use taskkill to terminate all processes with the given name
		err := pm.runCommand("taskkill", "/F", "/IM", fmt.Sprintf("%s.exe", name))
		if err != nil {
			pm.logger.Debug("No %s processes found to clean up", name)
		} else {
			pm.logger.Info("Terminated %s processes", name)
		}
	}

	return nil
}

// cleanupUnixProcesses cleans up Unix processes
func (pm *ProcessManager) cleanupUnixProcesses(processNames []string) error {
	pm.logger.Debug("Cleaning up Unix processes")

	for _, name := range processNames {
		// Use pkill to terminate all processes with the given name
		err := pm.runCommand("pkill", "-f", name)
		if err != nil {
			pm.logger.Debug("No %s processes found to clean up", name)
		} else {
			pm.logger.Info("Terminated %s processes", name)
		}
	}

	return nil
}

// performFinalCleanup performs final cleanup tasks
func (pm *ProcessManager) performFinalCleanup() {
	pm.logger.Debug("Performing final cleanup")

	// Clear child processes list
	pm.childProcesses = make([]*ManagedProcess, 0)

	// Additional platform-specific cleanup
	if runtime.GOOS == "windows" {
		pm.performWindowsCleanup()
	} else {
		pm.performUnixCleanup()
	}
}

// performWindowsCleanup performs Windows-specific cleanup
func (pm *ProcessManager) performWindowsCleanup() {
	pm.logger.Debug("Performing Windows-specific cleanup")
	// Windows-specific cleanup tasks could go here
}

// performUnixCleanup performs Unix-specific cleanup
func (pm *ProcessManager) performUnixCleanup() {
	pm.logger.Debug("Performing Unix-specific cleanup")
	// Unix-specific cleanup tasks could go here
}

// WaitForExit waits for the application to exit
func (pm *ProcessManager) WaitForExit() {
	<-pm.shutdownChan
}

// Shutdown triggers a graceful shutdown
func (pm *ProcessManager) Shutdown() {
	pm.logger.Info("Triggering graceful shutdown")
	select {
	case pm.shutdownChan <- syscall.SIGTERM:
		pm.logger.Debug("Shutdown signal sent")
	default:
		pm.logger.Debug("Shutdown already in progress")
	}
}

// GetProcessInfo returns information about managed processes
func (pm *ProcessManager) GetProcessInfo() map[string]interface{} {
	info := map[string]interface{}{
		"total_processes":    len(pm.childProcesses),
		"cleanup_completed":  pm.cleanupCompleted,
		"processes":         make([]map[string]interface{}, 0),
	}

	for _, proc := range pm.childProcesses {
		procInfo := map[string]interface{}{
			"name":    proc.Name,
			"pid":     proc.PID,
			"started": proc.Started.Format(time.RFC3339),
			"running": pm.IsProcessRunning(proc.PID),
		}
		info["processes"] = append(info["processes"].([]map[string]interface{}), procInfo)
	}

	return info
}