use std::collections::HashMap;
use std::process::{Command, Stdio};
use std::sync::{Arc, Mutex};
use tokio::process::{Child, Command as AsyncCommand};
use tracing::{info, warn, error};
use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::models::{LauncherError, LauncherResult};

/// Launch configuration for a game instance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LaunchConfig {
    pub instance_id: String,
    pub instance_name: String,
    pub prism_path: String,
    pub java_path: Option<String>,
    pub working_directory: String,
    pub additional_args: Vec<String>,
    pub memory_mb: Option<u32>,
    pub custom_jvm_args: Vec<String>,
    pub environment_vars: HashMap<String, String>,
}

/// Status of a launched process
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ProcessStatus {
    Starting,
    Running,
    Finished,
    Crashed,
    Killed,
}

/// Information about a launched game process
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LaunchedProcess {
    pub id: String,
    pub instance_id: String,
    pub instance_name: String,
    pub pid: u32,
    pub status: ProcessStatus,
    pub started_at: DateTime<Utc>,
    pub exit_code: Option<i32>,
    pub crash_reason: Option<String>,
    pub launch_time_ms: Option<u64>,
}

/// Launch manager for handling game processes
pub struct LaunchManager {
    active_processes: Arc<Mutex<HashMap<String, LaunchedProcess>>>,
    launch_configurations: Arc<Mutex<HashMap<String, LaunchConfig>>>,
    process_handles: Arc<Mutex<HashMap<String, Child>>>,
}

impl LaunchManager {
    /// Create a new launch manager
    pub fn new() -> Self {
        Self {
            active_processes: Arc::new(Mutex::new(HashMap::new())),
            launch_configurations: Arc::new(Mutex::new(HashMap::new())),
            process_handles: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    /// Launch a game instance with the given configuration
    pub async fn launch_instance(&self, config: LaunchConfig) -> LauncherResult<String> {
        let launch_id = Uuid::new_v4().to_string();
        let start_time = std::time::Instant::now();

        info!("Starting launch for instance {} ({})", config.instance_name, config.instance_id);

        // Store the launch configuration
        {
            let mut configs = self.launch_configurations.lock().unwrap();
            configs.insert(config.instance_id.clone(), config.clone());
        }

        // Create process info
        let process_info = LaunchedProcess {
            id: launch_id.clone(),
            instance_id: config.instance_id.clone(),
            instance_name: config.instance_name.clone(),
            pid: 0, // Will be set after process starts
            status: ProcessStatus::Starting,
            started_at: Utc::now(),
            exit_code: None,
            crash_reason: None,
            launch_time_ms: None,
        };

        // Store initial process info
        {
            let mut processes = self.active_processes.lock().unwrap();
            processes.insert(launch_id.clone(), process_info);
        }

        // Build the command to launch Prism Launcher
        let mut cmd = self.build_launch_command(&config)?;

        info!("Executing command: {:?}", cmd);

        // Start the process
        match cmd.spawn() {
            Ok(mut child) => {
                let pid = child.id().unwrap_or(0);

                // Update process info with PID and launch time
                {
                    let mut processes = self.active_processes.lock().unwrap();
                    if let Some(process) = processes.get_mut(&launch_id) {
                        process.pid = pid;
                        process.status = ProcessStatus::Running;
                        process.launch_time_ms = Some(start_time.elapsed().as_millis() as u64);
                    }
                }

                // Store the process handle for later management
                {
                    let mut handles = self.process_handles.lock().unwrap();
                    handles.insert(launch_id.clone(), child);
                }

                // Start monitoring the process
                self.monitor_process(launch_id.clone());

                info!("Successfully launched instance {} (PID: {})", config.instance_name, pid);
                Ok(launch_id)
            },
            Err(e) => {
                error!("Failed to launch instance {}: {}", config.instance_name, e);

                // Update process info to reflect failure
                {
                    let mut processes = self.active_processes.lock().unwrap();
                    if let Some(process) = processes.get_mut(&launch_id) {
                        process.status = ProcessStatus::Crashed;
                        process.crash_reason = Some(format!("Failed to start process: {}", e));
                    }
                }

                Err(LauncherError::LaunchFailed(
                    format!("Failed to launch instance {}: {}", config.instance_name, e)
                ))
            }
        }
    }

    /// Build the launch command for Prism Launcher
    fn build_launch_command(&self, config: &LaunchConfig) -> LauncherResult<AsyncCommand> {
        let prism_executable = if cfg!(target_os = "windows") {
            "PrismLauncher.exe"
        } else if cfg!(target_os = "macos") {
            "PrismLauncher"
        } else {
            "PrismLauncher"
        };

        let mut cmd = AsyncCommand::new(prism_executable);

        // Set working directory to Prism directory
        cmd.current_dir(&config.prism_path);

        // Add Prism Launcher arguments
        cmd.arg("--dir").arg(".");
        cmd.arg("--launch").arg(&config.instance_name);

        // Add custom JVM arguments if provided
        if !config.custom_jvm_args.is_empty() {
            cmd.arg("--jvm").args(&config.custom_jvm_args);
        }

        // Set memory allocation if specified
        if let Some(memory_mb) = config.memory_mb {
            cmd.arg("--memory").arg(memory_mb.to_string());
        }

        // Add additional arguments
        cmd.args(&config.additional_args);

        // Set environment variables
        cmd.envs(&config.environment_vars);

        // Set Java home if specified
        if let Some(java_path) = &config.java_path {
            cmd.env("JAVA_HOME", java_path);

            // Add Java bin to PATH
            let java_bin = if cfg!(target_os = "windows") {
                format!("{};{}",
                    std::path::Path::new(java_path).join("bin").to_string_lossy(),
                    std::env::var("PATH").unwrap_or_default()
                )
            } else {
                format!(":{}",
                    std::path::Path::new(java_path).join("bin").to_string_lossy()
                )
            };
            cmd.env("PATH", java_bin);
        }

        // Set stdout and stderr for logging
        cmd.stdout(Stdio::piped());
        cmd.stderr(Stdio::piped());

        Ok(cmd)
    }

    /// Monitor a running process
    fn monitor_process(&self, launch_id: String) {
        let processes = Arc::clone(&self.active_processes);
        let handles = Arc::clone(&self.process_handles);

        tokio::spawn(async move {
            // Extract the child process from handles
            let child_opt = {
                let mut handles_guard = handles.lock().unwrap();
                handles_guard.remove(&launch_id)
            };

            // Wait for the process to finish
            let (exit_code, crash_reason) = if let Some(mut child) = child_opt {
                match child.wait().await {
                    Ok(status) => {
                        let exit_code = status.code();
                        let crash_reason = if status.success() {
                            None
                        } else {
                            Some(format!("Process exited with code: {:?}", exit_code))
                        };
                        (exit_code, crash_reason)
                    },
                    Err(e) => (None, Some(format!("Process error: {}", e)))
                }
            } else {
                (None, Some("Process handle not found".to_string()))
            };

            // Update process status
            {
                let mut processes_guard = processes.lock().unwrap();
                if let Some(process) = processes_guard.get_mut(&launch_id) {
                    process.exit_code = exit_code;
                    process.crash_reason = crash_reason;
                    process.status = if exit_code == Some(0) {
                        ProcessStatus::Finished
                    } else {
                        ProcessStatus::Crashed
                    };
                }
            }

            info!("Process monitoring completed for launch ID: {}", launch_id);
        });
    }

    /// Get the status of a launched process
    pub fn get_launch_status(&self, launch_id: &str) -> LauncherResult<Option<LaunchedProcess>> {
        let processes = self.active_processes.lock().unwrap();
        Ok(processes.get(launch_id).cloned())
    }

    /// Get all active processes
    pub fn get_active_processes(&self) -> LauncherResult<Vec<LaunchedProcess>> {
        let processes = self.active_processes.lock().unwrap();
        Ok(processes.values().cloned().collect())
    }

    /// Get processes for a specific instance
    pub fn get_instance_processes(&self, instance_id: &str) -> LauncherResult<Vec<LaunchedProcess>> {
        let processes = self.active_processes.lock().unwrap();
        Ok(processes
            .values()
            .filter(|p| p.instance_id == instance_id)
            .cloned()
            .collect())
    }

    /// Terminate a running process
    pub async fn terminate_instance(&self, launch_id: &str) -> LauncherResult<()> {
        info!("Terminating process with launch ID: {}", launch_id);

        let child = {
            let mut handles = self.process_handles.lock().unwrap();
            handles.remove(launch_id)
        };

        if let Some(mut child) = child {
            // Try to terminate gracefully first
            match child.kill().await {
                Ok(_) => {
                    info!("Successfully terminated process: {}", launch_id);

                    // Update process status
                    let mut processes = self.active_processes.lock().unwrap();
                    if let Some(process) = processes.get_mut(launch_id) {
                        process.status = ProcessStatus::Killed;
                        process.crash_reason = Some("Terminated by user".to_string());
                    }
                },
                Err(e) => {
                    warn!("Failed to terminate process {}: {}", launch_id, e);
                    return Err(LauncherError::ProcessTermination(
                        format!("Failed to terminate process {}: {}", launch_id, e)
                    ));
                }
            }
        } else {
            warn!("Process not found for termination: {}", launch_id);
            return Err(LauncherError::ProcessNotFound(
                format!("Process {} not found", launch_id)
            ));
        }

        Ok(())
    }

    /// Force kill all processes for an instance (including child processes)
    pub async fn force_kill_instance_processes(&self, instance_id: &str) -> LauncherResult<u32> {
        info!("Force killing all processes for instance: {}", instance_id);

        let mut killed_count = 0;

        // Kill managed processes first
        let launch_ids: Vec<String> = {
            let processes = self.active_processes.lock().unwrap();
            processes
                .values()
                .filter(|p| p.instance_id == instance_id)
                .map(|p| p.id.clone())
                .collect()
        };

        for launch_id in launch_ids {
            if let Err(e) = self.terminate_instance(&launch_id).await {
                warn!("Failed to terminate process {}: {}", launch_id, e);
            } else {
                killed_count += 1;
            }
        }

        // Force kill any remaining processes using system commands
        if cfg!(target_os = "windows") {
            // Kill Prism Launcher processes
            if let Ok(_) = tokio::process::Command::new("taskkill")
                .args(&["/F", "/IM", "PrismLauncher.exe"])
                .output()
                .await {
                killed_count += 1;
            }

            // Kill Java processes (be careful not to kill other Java applications)
            // This is a simplified approach - in production, you might want to be more selective
            if let Ok(_) = tokio::process::Command::new("taskkill")
                .args(&["/F", "/IM", "java.exe"])
                .output()
                .await {
                killed_count += 1;
            }
        } else if cfg!(target_os = "linux") {
            // Kill Prism Launcher processes on Linux
            if let Ok(_) = tokio::process::Command::new("pkill")
                .args(&["-f", "PrismLauncher"])
                .output()
                .await {
                killed_count += 1;
            }

            // Kill Java processes related to Minecraft
            if let Ok(_) = tokio::process::Command::new("pkill")
                .args(&["-f", "minecraft"])
                .output()
                .await {
                killed_count += 1;
            }
        } else if cfg!(target_os = "macos") {
            // Kill Prism Launcher processes on macOS
            if let Ok(_) = tokio::process::Command::new("pkill")
                .args(&["-f", "PrismLauncher"])
                .output()
                .await {
                killed_count += 1;
            }
        }

        info!("Force killed {} processes for instance {}", killed_count, instance_id);
        Ok(killed_count)
    }

    /// Clean up finished processes
    pub async fn cleanup_finished_processes(&self) -> LauncherResult<u32> {
        let mut to_remove = Vec::new();

        {
            let processes = self.active_processes.lock().unwrap();
            for (id, process) in processes.iter() {
                match process.status {
                    ProcessStatus::Finished | ProcessStatus::Crashed | ProcessStatus::Killed => {
                        to_remove.push(id.clone());
                    },
                    _ => {}
                }
            }
        }

        // Remove finished processes
        {
            let mut processes = self.active_processes.lock().unwrap();
            for id in &to_remove {
                processes.remove(id);
            }
        }

        // Also clean up process handles
        {
            let mut handles = self.process_handles.lock().unwrap();
            for id in &to_remove {
                handles.remove(id);
            }
        }

        if !to_remove.is_empty() {
            info!("Cleaned up {} finished processes", to_remove.len());
        }

        Ok(to_remove.len() as u32)
    }

    /// Get launch configuration for an instance
    pub fn get_launch_config(&self, instance_id: &str) -> LauncherResult<Option<LaunchConfig>> {
        let configs = self.launch_configurations.lock().unwrap();
        Ok(configs.get(instance_id).cloned())
    }

    /// Update launch configuration for an instance
    pub fn update_launch_config(&self, config: LaunchConfig) -> LauncherResult<()> {
        let mut configs = self.launch_configurations.lock().unwrap();
        configs.insert(config.instance_id.clone(), config);
        Ok(())
    }

    /// Initialize the launch manager and clean up any orphaned processes
    pub async fn initialize(&self) -> LauncherResult<()> {
        info!("Initializing Launch Manager");

        // Clean up any existing process data from previous sessions
        self.cleanup_finished_processes().await?;

        // Detect any running Prism Launcher processes from previous launcher sessions
        self.detect_orphaned_processes().await?;

        info!("Launch Manager initialized successfully");
        Ok(())
    }

    /// Detect and handle orphaned processes from previous launcher sessions
    async fn detect_orphaned_processes(&self) -> LauncherResult<()> {
        info!("Detecting orphaned processes from previous sessions");

        // This is a simplified implementation
        // In a production environment, you might want to:
        // 1. Check for running Prism Launcher processes
        // 2. Verify if they were launched by this launcher
        // 3. Either adopt them or clean them up

        if cfg!(target_os = "windows") {
            // Check for running Prism Launcher processes
            if let Ok(output) = tokio::process::Command::new("tasklist")
                .args(&["/FI", "IMAGENAME eq PrismLauncher.exe"])
                .output()
                .await {

                let output_str = String::from_utf8_lossy(&output.stdout);
                if output_str.contains("PrismLauncher.exe") {
                    info!("Found existing Prism Launcher processes - they may be from previous sessions");
                }
            }
        }

        info!("Orphaned process detection completed");
        Ok(())
    }
}

impl Default for LaunchManager {
    fn default() -> Self {
        Self::new()
    }
}

impl Drop for LaunchManager {
    fn drop(&mut self) {
        info!("LaunchManager being dropped - cleaning up active processes");

        // This is a synchronous cleanup during drop
        // In production, you might want to ensure all processes are properly terminated
        if let Ok(mut handles) = self.process_handles.try_lock() {
            for (id, child) in handles.drain() {
                // Try to kill the process synchronously using std::process::Child
                // Since we can't await in drop, we'll try to get the process id and kill it manually
                if let Some(pid) = child.id() {
                    if cfg!(target_os = "windows") {
                        // Use taskkill on Windows
                        if let Err(_) = std::process::Command::new("taskkill")
                            .args(&["/F", "/PID", &pid.to_string()])
                            .output() {
                            warn!("Failed to kill process {} during cleanup", id);
                        } else {
                            info!("Successfully killed process {} during cleanup", id);
                        }
                    } else {
                        // Use kill on Unix systems
                        if let Err(_) = std::process::Command::new("kill")
                            .arg(pid.to_string())
                            .output() {
                            warn!("Failed to kill process {} during cleanup", id);
                        } else {
                            info!("Successfully killed process {} during cleanup", id);
                        }
                    }
                }
            }
        }
    }
}