use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use serde::{Serialize, Deserialize};
use tracing::{error, info};
use chrono::{DateTime, Utc};
use crate::models::LauncherError;

/// Error context for better error reporting
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorContext {
    pub operation: String,
    pub component: String,
    pub user_friendly_message: String,
    pub technical_details: String,
    pub timestamp: DateTime<Utc>,
    pub error_code: Option<String>,
    pub recovery_suggestions: Vec<String>,
}

/// Enhanced error handler for comprehensive error management
pub struct ErrorHandler {
    error_history: Arc<RwLock<Vec<ErrorContext>>>,
    recovery_strategies: HashMap<String, Vec<String>>,
}

impl ErrorHandler {
    pub fn new() -> Self {
        let mut handler = Self {
            error_history: Arc::new(RwLock::new(Vec::new())),
            recovery_strategies: HashMap::new(),
        };

        // Initialize recovery strategies
        handler.init_recovery_strategies();

        handler
    }

    fn init_recovery_strategies(&mut self) {
        // Network error recovery strategies
        self.recovery_strategies.insert(
            "Network".to_string(),
            vec![
                "Check your internet connection".to_string(),
                "Try again in a few moments".to_string(),
                "If the problem persists, contact support".to_string(),
            ],
        );

        // File system error recovery strategies
        self.recovery_strategies.insert(
            "FileSystem".to_string(),
            vec![
                "Check if you have sufficient disk space".to_string(),
                "Ensure the launcher has write permissions".to_string(),
                "Try running the launcher as administrator".to_string(),
            ],
        );

        // Java error recovery strategies
        self.recovery_strategies.insert(
            "JavaNotFound".to_string(),
            vec![
                "Install Java from https://adoptium.net".to_string(),
                "Check if Java is in your system PATH".to_string(),
                "Configure Java path in launcher settings".to_string(),
            ],
        );

        // Prism Launcher error recovery strategies
        self.recovery_strategies.insert(
            "PrismNotFound".to_string(),
            vec![
                "Install Prism Launcher from the settings page".to_string(),
                "Verify Prism Launcher installation path".to_string(),
                "Download and install Prism Launcher manually".to_string(),
            ],
        );

        // Download error recovery strategies
        self.recovery_strategies.insert(
            "DownloadFailed".to_string(),
            vec![
                "Check your internet connection".to_string(),
                "Try downloading again".to_string(),
                "Clear download cache and retry".to_string(),
            ],
        );

        // Launch error recovery strategies
        self.recovery_strategies.insert(
            "LaunchFailed".to_string(),
            vec![
                "Check if Java is properly installed".to_string(),
                "Verify instance configuration".to_string(),
                "Check available system memory".to_string(),
                "Try launching with reduced memory settings".to_string(),
            ],
        );
    }

    /// Handle and enhance an error with context
    pub async fn handle_error(
        &self,
        error: &LauncherError,
        operation: &str,
        component: &str,
    ) -> LauncherError {
        let (error_type, technical_details) = self.extract_error_info(error);

        let user_friendly_message = self.create_user_friendly_message(error, operation);
        let recovery_suggestions = self.get_recovery_suggestions(&error_type);
        let error_code = self.generate_error_code(&error_type);

        let context = ErrorContext {
            operation: operation.to_string(),
            component: component.to_string(),
            user_friendly_message,
            technical_details,
            timestamp: Utc::now(),
            error_code,
            recovery_suggestions,
        };

        // Log the error with context
        error!(
            "Error in {}::{}: {} | Technical: {} | Code: {}",
            component,
            operation,
            context.user_friendly_message,
            context.technical_details,
            context.error_code.as_ref().unwrap_or(&"UNKNOWN".to_string())
        );

        // Store in error history
        {
            let mut history = self.error_history.write().await;
            history.push(context.clone());

            // Keep only last 100 errors
            if history.len() > 100 {
                history.remove(0);
            }
        }

        // Return enhanced error
        LauncherError::Process(format!("{}: {}", context.user_friendly_message, context.technical_details))
    }

    /// Get recent errors for troubleshooting
    pub async fn get_recent_errors(&self, limit: usize) -> Vec<ErrorContext> {
        let history = self.error_history.read().await;
        let start = if history.len() > limit {
            history.len() - limit
        } else {
            0
        };

        history[start..].to_vec()
    }

    /// Clear error history
    pub async fn clear_error_history(&self) {
        let mut history = self.error_history.write().await;
        history.clear();
        info!("Error history cleared");
    }

    /// Generate troubleshooting report
    pub async fn generate_troubleshooting_report(&self) -> String {
        let history = self.error_history.read().await;

        if history.is_empty() {
            return "No errors recorded.".to_string();
        }

        let mut report = String::new();
        report.push_str("# TheBoys Launcher Troubleshooting Report\n\n");
        report.push_str(&format!("Generated: {}\n\n", Utc::now().format("%Y-%m-%d %H:%M:%S UTC")));

        // Error frequency analysis
        let mut error_counts: HashMap<String, usize> = HashMap::new();
        for context in history.iter() {
            let error_type = self.extract_error_type(&context.technical_details);
            *error_counts.entry(error_type).or_insert(0) += 1;
        }

        report.push_str("## Error Frequency Analysis\n\n");
        for (error_type, count) in error_counts.iter() {
            report.push_str(&format!("- {}: {} occurrences\n", error_type, count));
        }

        report.push_str("\n## Recent Errors\n\n");
        for (i, context) in history.iter().rev().take(10).enumerate() {
            report.push_str(&format!("### Error #{}\n", i + 1));
            report.push_str(&format!("- **Component**: {}\n", context.component));
            report.push_str(&format!("- **Operation**: {}\n", context.operation));
            report.push_str(&format!("- **Message**: {}\n", context.user_friendly_message));
            report.push_str(&format!("- **Code**: {}\n", context.error_code.as_ref().unwrap_or(&"N/A".to_string())));
            report.push_str(&format!("- **Time**: {}\n", context.timestamp.format("%Y-%m-%d %H:%M:%S UTC")));

            if !context.recovery_suggestions.is_empty() {
                report.push_str("- **Recovery Suggestions**:\n");
                for suggestion in &context.recovery_suggestions {
                    report.push_str(&format!("  - {}\n", suggestion));
                }
            }
            report.push_str("\n");
        }

        report.push_str("## Recommended Actions\n\n");
        if let Some((most_common, _)) = error_counts.iter().max_by_key(|(_, count)| *count) {
            if let Some(suggestions) = self.recovery_strategies.get(most_common) {
                report.push_str(&format!("Based on recent errors, consider the following:\n"));
                for suggestion in suggestions {
                    report.push_str(&format!("- {}\n", suggestion));
                }
            }
        }

        report
    }

    fn extract_error_info(&self, error: &LauncherError) -> (String, String) {
        match error {
            LauncherError::Network(msg) => ("Network".to_string(), format!("Network error: {}", msg)),
            LauncherError::FileSystem(msg) => ("FileSystem".to_string(), format!("File system error: {}", msg)),
            LauncherError::JavaNotFound => ("JavaNotFound".to_string(), "Java installation not found".to_string()),
            LauncherError::PrismNotFound => ("PrismNotFound".to_string(), "Prism Launcher not found".to_string()),
            LauncherError::InvalidConfig(msg) => ("InvalidConfig".to_string(), format!("Configuration error: {}", msg)),
            LauncherError::DownloadFailed(msg) => ("DownloadFailed".to_string(), format!("Download failed: {}", msg)),
            LauncherError::Process(msg) => ("Process".to_string(), format!("Process error: {}", msg)),
            LauncherError::Serialization(msg) => ("Serialization".to_string(), format!("Serialization error: {}", msg)),
            LauncherError::PermissionDenied(msg) => ("PermissionDenied".to_string(), format!("Permission denied: {}", msg)),
            LauncherError::ModpackNotFound(msg) => ("ModpackNotFound".to_string(), format!("Modpack not found: {}", msg)),
            LauncherError::InstanceNotFound(msg) => ("InstanceNotFound".to_string(), format!("Instance not found: {}", msg)),
            LauncherError::LaunchFailed(msg) => ("LaunchFailed".to_string(), format!("Launch failed: {}", msg)),
            LauncherError::ProcessNotFound(msg) => ("ProcessNotFound".to_string(), format!("Process not found: {}", msg)),
            LauncherError::ProcessTermination(msg) => ("ProcessTermination".to_string(), format!("Process termination failed: {}", msg)),
            LauncherError::NotImplemented(msg) => ("NotImplemented".to_string(), format!("Feature not implemented: {}", msg)),
            LauncherError::NotFound(msg) => ("NotFound".to_string(), format!("Resource not found: {}", msg)),
            LauncherError::UpdateFailed(msg) => ("UpdateFailed".to_string(), format!("Update failed: {}", msg)),
        }
    }

    fn extract_error_type(&self, technical_details: &str) -> String {
        if technical_details.contains("Network") {
            "Network".to_string()
        } else if technical_details.contains("File system") {
            "FileSystem".to_string()
        } else if technical_details.contains("Java") {
            "JavaNotFound".to_string()
        } else if technical_details.contains("Prism") {
            "PrismNotFound".to_string()
        } else if technical_details.contains("Download") {
            "DownloadFailed".to_string()
        } else if technical_details.contains("Launch") {
            "LaunchFailed".to_string()
        } else {
            "Unknown".to_string()
        }
    }

    fn create_user_friendly_message(&self, error: &LauncherError, operation: &str) -> String {
        match error {
            LauncherError::Network(_) => format!("Failed to connect to the server while {}", operation),
            LauncherError::FileSystem(_) => format!("File system error while {}", operation),
            LauncherError::JavaNotFound => "Java installation is required but not found".to_string(),
            LauncherError::PrismNotFound => "Prism Launcher is required but not found".to_string(),
            LauncherError::InvalidConfig(_) => format!("Invalid configuration while {}", operation),
            LauncherError::DownloadFailed(_) => format!("Failed to download required files while {}", operation),
            LauncherError::Process(_) => format!("Process error while {}", operation),
            LauncherError::Serialization(_) => format!("Data format error while {}", operation),
            LauncherError::PermissionDenied(_) => format!("Permission denied while {}", operation),
            LauncherError::ModpackNotFound(_) => format!("Modpack not found while {}", operation),
            LauncherError::InstanceNotFound(_) => format!("Instance not found while {}", operation),
            LauncherError::LaunchFailed(_) => format!("Failed to launch while {}", operation),
            LauncherError::ProcessNotFound(_) => format!("Process not found while {}", operation),
            LauncherError::ProcessTermination(_) => format!("Failed to terminate process while {}", operation),
            LauncherError::NotImplemented(_) => format!("This feature is not yet implemented: {}", operation),
            LauncherError::NotFound(_) => format!("Required resource not found while {}", operation),
            LauncherError::UpdateFailed(_) => format!("Update failed while {}", operation),
        }
    }

    fn get_recovery_suggestions(&self, error_type: &str) -> Vec<String> {
        self.recovery_strategies
            .get(error_type)
            .cloned()
            .unwrap_or_else(|| {
                vec![
                    "Try again".to_string(),
                    "If the problem persists, contact support".to_string(),
                ]
            })
    }

    fn generate_error_code(&self, error_type: &str) -> Option<String> {
        Some(format!("{}{:04}", error_type.chars().take(3).collect::<String>().to_uppercase(),
            chrono::Utc::now().timestamp() as u32 % 10000))
    }
}

/// Macro for enhanced error handling
#[macro_export]
macro_rules! handle_error {
    ($result:expr, $operation:expr, $component:expr) => {
        match $result {
            Ok(value) => value,
            Err(error) => {
                let handler = $crate::utils::error_handler::ErrorHandler::new();
                return Err(handler.handle_error(&error, $operation, $component).await);
            }
        }
    };
}

/// Global error handler instance
use std::sync::OnceLock;
static ERROR_HANDLER: OnceLock<ErrorHandler> = OnceLock::new();

pub fn get_error_handler() -> &'static ErrorHandler {
    ERROR_HANDLER.get_or_init(|| ErrorHandler::new())
}