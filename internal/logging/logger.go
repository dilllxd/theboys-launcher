// Package logging provides structured logging for the TheBoys Launcher
package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Level represents the log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level     Level
	logger    *log.Logger
	logFile   *os.File
	enableGUI bool
}

// Config holds logger configuration
type Config struct {
	Level      Level
	LogPath    string
	EnableGUI  bool
	MaxSizeMB  int
	MaxBackups int
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return NewLoggerWithConfig(Config{
		Level:     InfoLevel,
		LogPath:   getLogPath(),
		EnableGUI: true,
	})
}

// NewLoggerWithConfig creates a new logger with custom configuration
func NewLoggerWithConfig(config Config) *Logger {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(config.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	// Open log file
	logFile, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		logFile = nil
	}

	logger := &Logger{
		level:     config.Level,
		logger:    log.New(os.Stdout, "", 0),
		logFile:   logFile,
		enableGUI: config.EnableGUI,
	}

	// Log startup
	logger.log(InfoLevel, "Logger initialized", map[string]interface{}{
		"level":    config.Level.String(),
		"logPath":  config.LogPath,
		"platform": runtime.GOOS,
	})

	return logger
}

// Close closes the logger and cleans up resources
func (l *Logger) Close() {
	if l.logFile != nil {
		l.log(InfoLevel, "Logger shutting down", nil)
		l.logFile.Close()
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Debugf logs a debug message with structured data
func (l *Logger) Debugf(format string, data map[string]interface{}, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), data)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Infof logs an info message with structured data
func (l *Logger) Infof(format string, data map[string]interface{}, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), data)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Warnf logs a warning message with structured data
func (l *Logger) Warnf(format string, data map[string]interface{}, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), data)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Errorf logs an error message with structured data
func (l *Logger) Errorf(format string, data map[string]interface{}, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), data)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

// Fatalf logs a fatal message with structured data and exits
func (l *Logger) Fatalf(format string, data map[string]interface{}, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), data)
	os.Exit(1)
}

// log is the internal logging method
func (l *Logger) log(level Level, message string, data map[string]interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	var caller string
	if ok {
		caller = filepath.Base(file) + ":" + fmt.Sprintf("%d", line)
	} else {
		caller = "unknown"
	}

	// Create log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s %s (%s)", timestamp, level.String(), message, caller)

	// Add structured data if provided
	if len(data) > 0 {
		var parts []string
		for k, v := range data {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		logEntry += fmt.Sprintf(" {%s}", strings.Join(parts, ", "))
	}

	// Log to console
	l.logger.Println(logEntry)

	// Log to file if available
	if l.logFile != nil {
		l.logFile.WriteString(logEntry + "\n")
		l.logFile.Sync()
	}
}

// getLogPath returns the appropriate log path for the current platform
func getLogPath() string {
	if runtime.GOOS == "windows" {
		// For Windows, use logs directory next to executable
		exePath, _ := os.Executable()
		return filepath.Join(filepath.Dir(exePath), "logs", "theboys.log")
	}

	// For Unix-like systems, use user cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "/tmp/theboys.log"
	}
	return filepath.Join(cacheDir, "theboys", "theboys.log")
}