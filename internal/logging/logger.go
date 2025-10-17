package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger interface defines the logging methods used throughout the application
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Initialize(logDir string) error
	GetLogDir() string
}

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// logger represents the application logger implementation
type logger struct {
	file     *os.File
	mu       sync.RWMutex
	level    LogLevel
	verbose  bool
	logDir   string
	logFile  string
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	return &logger{
		level:   INFO,
		verbose: false,
	}
}

// Initialize sets up the logging system
func (l *logger) Initialize(logDir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logDir = logDir

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Rotate old logs
	l.rotateLogs()

	// Create new log file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	l.logFile = filepath.Join(logDir, fmt.Sprintf("theboys-launcher-%s.log", timestamp))

	file, err := os.OpenFile(l.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	l.file = file

	// Set up standard logger to also write to file
	log.SetOutput(l.file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	l.Info("Logging initialized. Log file: %s", l.logFile)
	return nil
}

// SetLevel sets the minimum log level
func (l *logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetVerbose enables or disables verbose logging
func (l *logger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
}

// Close closes the log file
func (l *logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.Info("Closing log file")
		return l.file.Close()
	}
	return nil
}

// Debug logs a debug message
func (l *logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// log is the internal logging method
func (l *logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	// Always log to file if available
	if l.file != nil {
		l.file.WriteString(logEntry)
		l.file.Sync()
	}

	// Also output to console if verbose or error level
	if l.verbose || level >= WARN {
		fmt.Print(logEntry)
	}
}

// rotateLogs rotates old log files
func (l *logger) rotateLogs() {
	// Remove old log files (keep only last 5)
	entries, err := filepath.Glob(filepath.Join(l.logDir, "theboys-launcher-*.log"))
	if err != nil {
		return
	}

	if len(entries) > 5 {
		// Sort by modification time and remove oldest
		for i := 0; i < len(entries)-5; i++ {
			os.Remove(entries[i])
		}
	}

	// Check for latest.log and previous.log
	latestLog := filepath.Join(l.logDir, "latest.log")
	previousLog := filepath.Join(l.logDir, "previous.log")

	// If latest.log exists, move it to previous.log
	if _, err := os.Stat(latestLog); err == nil {
		os.Remove(previousLog)
		os.Rename(latestLog, previousLog)
	}
}

// GetLogFile returns the current log file path
func (l *logger) GetLogFile() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logFile
}

// GetLogDir returns the log directory
func (l *logger) GetLogDir() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logDir
}