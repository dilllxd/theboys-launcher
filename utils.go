package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// global writer used by log/fail and for piping subprocess output
var (
	out       io.Writer = os.Stdout
	activeLog *os.File
)

type logTeeWriter struct{}

func (logTeeWriter) Write(p []byte) (int, error) {
	// Write to console; ignore errors
	if len(p) > 0 {
		if _, err := os.Stdout.Write(p); err != nil {
			// fall back to Print in case Write fails
			fmt.Print(string(p))
		}
	}

	if activeLog != nil && len(p) > 0 {
		if _, err := activeLog.Write(p); err != nil {
			fmt.Printf("Warning: Failed to write to log file: %v\n", err)
			return len(p), nil
		}
		if err := activeLog.Sync(); err != nil {
			fmt.Printf("Warning: Failed to sync log file: %v\n", err)
		}
	}

	return len(p), nil
}

type launcherOptions struct {
	cleanupAfterUpdate bool
	cleanupOldExe      string
	cleanupNewExe      string
}

func parseOptions() launcherOptions {
	var opts launcherOptions
	flag.BoolVar(&opts.cleanupAfterUpdate, "cleanup-after-update", false, "internal use only")
	flag.StringVar(&opts.cleanupOldExe, "cleanup-old-exe", "", "internal use only")
	flag.StringVar(&opts.cleanupNewExe, "cleanup-new-exe", "", "internal use only")
	flag.Parse()
	return opts
}

func modpackLabel(mp Modpack) string {
	if name := strings.TrimSpace(mp.DisplayName); name != "" {
		return name
	}
	return mp.ID
}

func slugifyID(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "modpack"
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "modpack"
	}
	return result
}

func versionFileNameFor(mp Modpack) string { return "." + slugifyID(mp.ID) + "-version" }
func backupPrefixFor(mp Modpack) string    { return slugifyID(mp.ID) + "-backup-" }

// getLauncherHome is now implemented in platform-specific files
// This function is handled by platform_windows.go and platform_darwin.go

// -------------------- Helpers --------------------

func fail(err error) {
	msg := fmt.Sprintf("Error: %v", err)
	fmt.Fprintln(os.Stderr, msg)
	logf("%s", warnLine(msg))
	os.Exit(1)
}

func pause() {
	if os.Getenv(envNoPause) == "1" {
		return
	}
	fmt.Print("\nPress Enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func logf(format string, args ...interface{}) {
	// Create the log message
	message := fmt.Sprintf(format+"\n", args...)

	if out != nil {
		if _, err := fmt.Fprint(out, message); err != nil {
			fmt.Print(message)
			fmt.Printf("Warning: Failed to write to log output: %v\n", err)
		}
	} else {
		fmt.Print(message)
	}
}

// debugf only logs when debug mode is enabled
func debugf(format string, args ...interface{}) {
	// Only log if debug is enabled
	if settings.DebugEnabled {
		logf("DEBUG: "+format, args...)
	}
}

// -------------------- Log Setup --------------------

func setupLogging(root string) func() {
	logDir := filepath.Join(root, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
		return func() {}
	}

	// Rotate previous log
	previousLog := filepath.Join(logDir, "previous.log")
	currentLog := filepath.Join(logDir, "latest.log")

	if _, err := os.Stat(currentLog); err == nil {
		_ = os.Remove(previousLog)
		_ = os.Rename(currentLog, previousLog)
	}

	// Create new log file
	logFile, err := os.OpenFile(currentLog, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to create log file: %v\n", err)
		return func() {}
	}

	// Set global output to both console and the file
	activeLog = logFile
	out = logTeeWriter{}

	// Return cleanup function that flushes and closes
	return func() {
		if activeLog != nil {
			_ = activeLog.Sync()
			activeLog.Close()
			activeLog = nil
		}
		out = os.Stdout
	}
}

// -------------------- Emergency Crash Logging --------------------

func setupEmergencyCrashLogger(root string) {
	// Set up emergency crash logger
	logDir := filepath.Join(root, "logs")
	crashLogPath := filepath.Join(logDir, "crash.log")

	// Create panic handler
	defer func() {
		if r := recover(); r != nil {
			// Write crash to both console and crash file
			crashMsg := fmt.Sprintf("=== EMERGENCY CRASH ===\nTime: %s\nPanic: %v\nStack Trace:\n", time.Now().Format("2006-01-02 15:04:05"), r)

			// Print to console
			fmt.Print(crashMsg)

			// Get stack trace
			buf := make([]byte, 4096)
			stackLen := runtime.Stack(buf, false)
			stackTrace := string(buf[:stackLen])
			fmt.Print(stackTrace)

			// Write to crash file
			if file, err := os.OpenFile(crashLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				file.WriteString(crashMsg)
				file.WriteString(stackTrace)
				file.WriteString("\n=== END CRASH ===\n")
				file.Close()
				fmt.Printf("\nCrash details written to: %s\n", crashLogPath)
			}

			// Give time to read the message
			time.Sleep(3 * time.Second)
		}
	}()
}

// -------------------- Logging Helper Functions --------------------

func headerLine(title string) string {
	border := "═"
	padding := strings.Repeat(border, len(title)+4)
	return fmt.Sprintf("╔%s╗\n║ %s ║\n╚%s╝", padding, title, padding)
}

func sectionLine(title string) string {
	border := "═"
	padding := strings.Repeat(border, len(title)+4)
	return fmt.Sprintf("%s\n║ %s ║\n%s",
		fmt.Sprintf("╔%s╗", padding),
		fmt.Sprintf("║  %s  ║", title),
		fmt.Sprintf("╚%s╝", padding))
}

func stepLine(msg string) string {
	return fmt.Sprintf("  ● %s", msg)
}

func successLine(msg string) string {
	return fmt.Sprintf("  ✓ %s", msg)
}

func warnLine(msg string) string {
	return fmt.Sprintf("  ⚠ %s", msg)
}

func infoLine(msg string) string {
	return fmt.Sprintf("  ℹ %s", msg)
}

func dividerLine() string {
	return "────────────────────────────────────────"
}

// -------------------- UI Helper Functions --------------------

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func yesNoBox(prompt string, args ...string) bool {
	// For GUI mode, we default to true (user consent)
	// In a full GUI implementation, this would show a proper dialog
	// The function accepts variable arguments to match the old TUI interface
	return true
}

func waitEnter() {
	// No-op in GUI mode
}

// -------------------- File Operations --------------------

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Check if source file is executable and preserve permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Preserve the same permissions as the source file
	return os.WriteFile(dst, input, info.Mode())
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenOneLevel(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(path, entry.Name())
			subEntries, err := os.ReadDir(dirPath)
			if err != nil {
				return err
			}

			for _, subEntry := range subEntries {
				oldPath := filepath.Join(dirPath, subEntry.Name())
				newPath := filepath.Join(path, subEntry.Name())

				if err := os.Rename(oldPath, newPath); err != nil {
					return err
				}
			}

			// Remove the now-empty directory
			os.Remove(dirPath)
		}
	}

	return nil
}

// flattenJREExtraction handles platform-specific JRE extraction structures
func flattenJREExtraction(jreDir string) error {
	// First, flatten the top level (handles jdk-17.0.16+8-jre/ directory)
	if err := flattenOneLevel(jreDir); err != nil {
		return err
	}

	// On macOS, check if we have a Contents/Home structure and flatten it
	if runtime.GOOS == "darwin" {
		contentsPath := filepath.Join(jreDir, "Contents")
		if exists(contentsPath) {
			homePath := filepath.Join(contentsPath, "Home")
			if exists(homePath) {
				// Move everything from Contents/Home to the jreDir root
				entries, err := os.ReadDir(homePath)
				if err != nil {
					return err
				}

				for _, entry := range entries {
					oldPath := filepath.Join(homePath, entry.Name())
					newPath := filepath.Join(jreDir, entry.Name())

					if err := os.Rename(oldPath, newPath); err != nil {
						return err
					}
				}

				// Remove the now-empty Contents directory
				os.RemoveAll(contentsPath)
			}
		}
	}

	return nil
}
