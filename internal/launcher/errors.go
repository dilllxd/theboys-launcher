package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"theboys-launcher/internal/logging"
	"theboys-launcher/internal/platform"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	NetworkError ErrorType = iota
	FileSystemError
	ProcessError
	ConfigurationError
	DownloadError
	InstallationError
	JavaError
	PrismError
	ModpackError
	BackupError
	UpdateError
	UnknownError
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityLow ErrorSeverity = iota
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// LauncherError represents a structured error with context
type LauncherError struct {
	Type        ErrorType
	Severity    ErrorSeverity
	Code        string
	Message     string
	Details     string
	Timestamp   time.Time
	Retryable   bool
	UserAction  string
	TechDetails map[string]interface{}
}

// Error (error interface implementation)
func (e *LauncherError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s (%s)", e.Type, e.Message, e.Details, e.Code)
	}
	return fmt.Sprintf("[%s] %s (%s)", e.Type, e.Message, e.Code)
}

// ErrorHandler handles advanced error scenarios and user guidance
type ErrorHandler struct {
	platform    platform.Platform
	logger      logging.Logger
	errorHistory []*LauncherError
	pauseOnError bool
	debugMode    bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(platform platform.Platform, logger logging.Logger) *ErrorHandler {
	return &ErrorHandler{
		platform:     platform,
		logger:       logger,
		errorHistory: make([]*LauncherError, 0),
		pauseOnError: true,
		debugMode:    false,
	}
}

// SetPauseOnError configures whether to pause on errors
func (eh *ErrorHandler) SetPauseOnError(pause bool) {
	eh.pauseOnError = pause
}

// SetDebugMode enables or disables debug mode
func (eh *ErrorHandler) SetDebugMode(debug bool) {
	eh.debugMode = debug
}

// HandleError handles an error with advanced logic
func (eh *ErrorHandler) HandleError(err error, context map[string]interface{}) *LauncherError {
	launcherErr := eh.createLauncherError(err, context)
	eh.errorHistory = append(eh.errorHistory, launcherErr)

	eh.logError(launcherErr)
	eh.saveErrorToFile(launcherErr)

	if launcherErr.Severity >= ErrorSeverityHigh {
		eh.showUserGuidance(launcherErr)
	}

	if eh.pauseOnError && launcherErr.Severity >= ErrorSeverityMedium {
		eh.promptUserAction(launcherErr)
	}

	return launcherErr
}

// createLauncherError creates a structured error from a generic error
func (eh *ErrorHandler) createLauncherError(err error, context map[string]interface{}) *LauncherError {
	if launcherErr, ok := err.(*LauncherError); ok {
		return launcherErr
	}

	errStr := err.Error()
	launcherErr := &LauncherError{
		Timestamp:   time.Now(),
		TechDetails: make(map[string]interface{}),
	}

	// Add context to tech details
	for k, v := range context {
		launcherErr.TechDetails[k] = v
	}

	// Analyze error type and details
	eh.analyzeError(errStr, launcherErr)

	return launcherErr
}

// analyzeError analyzes an error string to determine type and suggest actions
func (eh *ErrorHandler) analyzeError(errStr string, launcherErr *LauncherError) {
	lowerErr := strings.ToLower(errStr)

	// Network errors
	if strings.Contains(lowerErr, "connection") ||
	   strings.Contains(lowerErr, "timeout") ||
	   strings.Contains(lowerErr, "network") ||
	   strings.Contains(lowerErr, "dns") {
		launcherErr.Type = NetworkError
		launcherErr.Severity = ErrorSeverityMedium
		launcherErr.Code = "NETWORK_ERROR"
		launcherErr.Message = "Network connection issue"
		launcherErr.Details = "Unable to connect to the server or service"
		launcherErr.Retryable = true
		launcherErr.UserAction = "Check your internet connection and try again"
	}

	// File system errors
	if strings.Contains(lowerErr, "permission denied") ||
	   strings.Contains(lowerErr, "access denied") ||
	   strings.Contains(lowerErr, "file not found") ||
	   strings.Contains(lowerErr, "no such file") ||
	   strings.Contains(lowerErr, "disk full") ||
	   strings.Contains(lowerErr, "read-only") {
		launcherErr.Type = FileSystemError
		launcherErr.Code = "FILESYSTEM_ERROR"

		if strings.Contains(lowerErr, "permission denied") || strings.Contains(lowerErr, "access denied") {
			launcherErr.Severity = ErrorSeverityHigh
			launcherErr.Message = "File access permission error"
			launcherErr.Details = "The launcher doesn't have permission to access files or directories"
			launcherErr.Retryable = false
			launcherErr.UserAction = "Run the launcher as administrator or check file permissions"
		} else if strings.Contains(lowerErr, "file not found") || strings.Contains(lowerErr, "no such file") {
			launcherErr.Severity = ErrorSeverityMedium
			launcherErr.Message = "File or directory not found"
			launcherErr.Details = "Required file or directory is missing"
			launcherErr.Retryable = true
			launcherErr.UserAction = "Reinstall the launcher or missing components"
		} else {
			launcherErr.Severity = ErrorSeverityHigh
			launcherErr.Message = "File system error"
			launcherErr.Details = "Error accessing or writing to files"
			launcherErr.Retryable = false
			launcherErr.UserAction = "Check disk space and file permissions"
		}
	}

	// Process errors
	if strings.Contains(lowerErr, "process") ||
	   strings.Contains(lowerErr, "already running") ||
	   strings.Contains(lowerErr, "exit status") ||
	   strings.Contains(lowerErr, "command not found") {
		launcherErr.Type = ProcessError
		launcherErr.Code = "PROCESS_ERROR"

		if strings.Contains(lowerErr, "already running") {
			launcherErr.Severity = ErrorSeverityMedium
			launcherErr.Message = "Process already running"
			launcherErr.Details = "Another instance of the application is already running"
			launcherErr.Retryable = true
			launcherErr.UserAction = "Close the other instance and try again"
		} else {
			launcherErr.Severity = ErrorSeverityMedium
			launcherErr.Message = "Process execution error"
			launcherErr.Details = "Failed to execute or manage a process"
			launcherErr.Retryable = true
			launcherErr.UserAction = "Check system requirements and try again"
		}
	}

	// Java errors
	if strings.Contains(lowerErr, "java") ||
	   strings.Contains(lowerErr, "jdk") ||
	   strings.Contains(lowerErr, "jre") {
		launcherErr.Type = JavaError
		launcherErr.Code = "JAVA_ERROR"
		launcherErr.Severity = ErrorSeverityHigh
		launcherErr.Message = "Java runtime error"
		launcherErr.Details = "Issue with Java installation or compatibility"
		launcherErr.Retryable = true
		launcherErr.UserAction = "Install or update Java to the required version"
	}

	// Download errors
	if strings.Contains(lowerErr, "download") ||
	   strings.Contains(lowerErr, "http") ||
	   strings.Contains(lowerErr, "404") ||
	   strings.Contains(lowerErr, "500") {
		launcherErr.Type = DownloadError
		launcherErr.Severity = ErrorSeverityMedium
		launcherErr.Code = "DOWNLOAD_ERROR"
		launcherErr.Message = "Download failed"
		launcherErr.Details = "Failed to download required files"
		launcherErr.Retryable = true
		launcherErr.UserAction = "Check internet connection and try again, or use manual download"
	}

	// Configuration errors
	if strings.Contains(lowerErr, "config") ||
	   strings.Contains(lowerErr, "invalid") ||
	   strings.Contains(lowerErr, "parse") {
		launcherErr.Type = ConfigurationError
		launcherErr.Severity = ErrorSeverityMedium
		launcherErr.Code = "CONFIG_ERROR"
		launcherErr.Message = "Configuration error"
		launcherErr.Details = "Invalid or corrupted configuration"
		launcherErr.Retryable = false
		launcherErr.UserAction = "Reset configuration or check settings file"
	}

	// Default to unknown error
	if launcherErr.Type == 0 {
		launcherErr.Type = UnknownError
		launcherErr.Severity = ErrorSeverityMedium
		launcherErr.Code = "UNKNOWN_ERROR"
		launcherErr.Message = "Unexpected error occurred"
		launcherErr.Details = errStr
		launcherErr.Retryable = true
		launcherErr.UserAction = "Try again or contact support if the problem persists"
	}
}

// logError logs an error with appropriate level
func (eh *ErrorHandler) logError(err *LauncherError) {
	switch err.Severity {
	case ErrorSeverityLow:
		eh.logger.Debug("Error: %s", err.Error())
	case ErrorSeverityMedium:
		eh.logger.Warn("Error: %s", err.Error())
	case ErrorSeverityHigh, ErrorSeverityCritical:
		eh.logger.Error("Error: %s", err.Error())
	}
}

// saveErrorToFile saves detailed error information to a file
func (eh *ErrorHandler) saveErrorToFile(err *LauncherError) {
	logDir := eh.logger.GetLogDir()
	if logDir == "" {
		return
	}

	// Create error entry
	errorEntry := map[string]interface{}{
		"timestamp":   err.Timestamp.Format(time.RFC3339),
		"type":        err.Type,
		"severity":    err.Severity,
		"code":        err.Code,
		"message":     err.Message,
		"details":     err.Details,
		"retryable":   err.Retryable,
		"user_action": err.UserAction,
		"tech_details": err.TechDetails,
		"platform":    runtime.GOOS,
		"arch":        runtime.GOARCH,
	}

	// Append to error log file
	// In a real implementation, this would append to a JSON array file
	_ = errorEntry
}

// showUserGuidance shows user-friendly guidance for errors
func (eh *ErrorHandler) showUserGuidance(err *LauncherError) {
	guidance := eh.generateUserGuidance(err)
	eh.logger.Info("User Guidance: %s", guidance)
}

// generateUserGuidance generates user-friendly guidance for an error
func (eh *ErrorHandler) generateUserGuidance(err *LauncherError) string {
	guidance := fmt.Sprintf("ERROR: %s\n\n", err.Message)
	guidance += fmt.Sprintf("Details: %s\n", err.Details)
	guidance += fmt.Sprintf("Recommended Action: %s\n\n", err.UserAction)

	if err.Retryable {
		guidance += "This error can be retried. You may try again.\n"
	} else {
		guidance += "This error requires manual intervention before retrying.\n"
	}

	if eh.debugMode {
		guidance += "\nTechnical Details:\n"
		for k, v := range err.TechDetails {
			guidance += fmt.Sprintf("  %s: %v\n", k, v)
		}
	}

	return guidance
}

// promptUserAction prompts the user for action (in GUI context)
func (eh *ErrorHandler) promptUserAction(err *LauncherError) {
	// In a GUI application, this would show a dialog
	// For now, just log the prompt
	eh.logger.Info("User prompt required for error: %s", err.Message)

	// Create a prompt file that the GUI can read
	promptFile := filepath.Join(eh.logger.GetLogDir(), "user_prompt.txt")
	promptText := fmt.Sprintf("ACTION REQUIRED\n\n%s\n\n%s",
		err.Message,
		eh.generateUserGuidance(err))

	if err := os.WriteFile(promptFile, []byte(promptText), 0644); err != nil {
		eh.logger.Warn("Failed to create user prompt file: %v", err)
	}
}

// ShouldRetry determines if an error should be retried
func (eh *ErrorHandler) ShouldRetry(err error) bool {
	if launcherErr, ok := err.(*LauncherError); ok {
		return launcherErr.Retryable
	}
	return true // Default to retryable for unknown errors
}

// GetErrorHistory returns the history of errors
func (eh *ErrorHandler) GetErrorHistory() []*LauncherError {
	return eh.errorHistory
}

// ClearErrorHistory clears the error history
func (eh *ErrorHandler) ClearErrorHistory() {
	eh.errorHistory = make([]*LauncherError, 0)
}

// GetRecentErrors returns recent errors within the specified duration
func (eh *ErrorHandler) GetRecentErrors(duration time.Duration) []*LauncherError {
	cutoff := time.Now().Add(-duration)
	recent := make([]*LauncherError, 0)

	for _, err := range eh.errorHistory {
		if err.Timestamp.After(cutoff) {
			recent = append(recent, err)
		}
	}

	return recent
}

// HasRecentCriticalErrors checks if there have been recent critical errors
func (eh *ErrorHandler) HasRecentCriticalErrors(duration time.Duration) bool {
	recent := eh.GetRecentErrors(duration)
	for _, err := range recent {
		if err.Severity == ErrorSeverityCritical {
			return true
		}
	}
	return false
}

// CreateNetworkError creates a network error
func (eh *ErrorHandler) CreateNetworkError(message, details string) *LauncherError {
	return &LauncherError{
		Type:        NetworkError,
		Severity:    ErrorSeverityMedium,
		Code:        "NETWORK_ERROR",
		Message:     message,
		Details:     details,
		Timestamp:   time.Now(),
		Retryable:   true,
		UserAction:  "Check internet connection and try again",
		TechDetails: make(map[string]interface{}),
	}
}

// CreateFileSystemError creates a file system error
func (eh *ErrorHandler) CreateFileSystemError(message, details string, severity ErrorSeverity) *LauncherError {
	return &LauncherError{
		Type:        FileSystemError,
		Severity:    severity,
		Code:        "FILESYSTEM_ERROR",
		Message:     message,
		Details:     details,
		Timestamp:   time.Now(),
		Retryable:   severity < ErrorSeverityHigh,
		UserAction:  "Check file permissions and disk space",
		TechDetails: make(map[string]interface{}),
	}
}

// CreateProcessError creates a process error
func (eh *ErrorHandler) CreateProcessError(message, details string) *LauncherError {
	return &LauncherError{
		Type:        ProcessError,
		Severity:    ErrorSeverityMedium,
		Code:        "PROCESS_ERROR",
		Message:     message,
		Details:     details,
		Timestamp:   time.Now(),
		Retryable:   true,
		UserAction:  "Check system requirements and try again",
		TechDetails: make(map[string]interface{}),
	}
}

// String implementations for better error display
func (et ErrorType) String() string {
	switch et {
	case NetworkError:
		return "Network"
	case FileSystemError:
		return "FileSystem"
	case ProcessError:
		return "Process"
	case ConfigurationError:
		return "Configuration"
	case DownloadError:
		return "Download"
	case InstallationError:
		return "Installation"
	case JavaError:
		return "Java"
	case PrismError:
		return "PrismLauncher"
	case ModpackError:
		return "Modpack"
	case BackupError:
		return "Backup"
	case UpdateError:
		return "Update"
	default:
		return "Unknown"
	}
}

func (es ErrorSeverity) String() string {
	switch es {
	case ErrorSeverityLow:
		return "Low"
	case ErrorSeverityMedium:
		return "Medium"
	case ErrorSeverityHigh:
		return "High"
	case ErrorSeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}