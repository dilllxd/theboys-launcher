use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Represents a modpack configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Modpack {
    pub id: String,
    pub display_name: String,
    pub pack_url: String,
    pub instance_name: String,
    pub description: String,
    pub default: bool,
    pub version: String,
    pub minecraft_version: String,
    pub modloader: Modloader,
    pub loader_version: String,
}

/// Modloader types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Modloader {
    Vanilla,
    Forge,
    Fabric,
    Quilt,
    NeoForge,
}

impl Modloader {
    pub fn as_str(&self) -> &'static str {
        match self {
            Modloader::Vanilla => "vanilla",
            Modloader::Forge => "forge",
            Modloader::Fabric => "fabric",
            Modloader::Quilt => "quilt",
            Modloader::NeoForge => "neoforge",
        }
    }

    pub fn from_str(s: &str) -> Self {
        match s.to_lowercase().as_str() {
            "vanilla" => Modloader::Vanilla,
            "forge" => Modloader::Forge,
            "fabric" => Modloader::Fabric,
            "quilt" => Modloader::Quilt,
            "neoforge" => Modloader::NeoForge,
            _ => Modloader::Vanilla, // Default fallback
        }
    }
}

/// Represents a locally installed modpack
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstalledModpack {
    pub modpack: Modpack,
    pub installed_version: String,
    pub install_path: String,
    pub install_date: String,
    pub last_played: Option<String>,
    pub total_playtime: u64,
    pub update_available: bool,
    pub size_bytes: u64,
}

/// Modpack update information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModpackUpdate {
    pub modpack_id: String,
    pub current_version: String,
    pub latest_version: String,
    pub update_available: bool,
    pub changelog_url: Option<String>,
    pub download_url: String,
    pub size_bytes: u64,
}

/// Launcher settings that can be configured by users
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LauncherSettings {
    pub memory_mb: u32,
    pub java_path: Option<String>,
    pub prism_path: Option<String>,
    pub instances_dir: Option<String>,
    pub auto_update: bool,
    pub theme: String, // "light", "dark", or "system"
    pub default_modpack_id: Option<String>,
    pub update_notifications: bool,
}

impl Default for LauncherSettings {
    fn default() -> Self {
        Self {
            memory_mb: 4096, // Default 4GB
            java_path: None,
            prism_path: None,
            instances_dir: None,
            auto_update: true,
            theme: "system".to_string(),
            default_modpack_id: None,
            update_notifications: true,
        }
    }
}

/// System information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemInfo {
    pub os: String,
    pub arch: String,
    pub total_memory_mb: u64,
    pub available_memory_mb: u64,
    pub cpu_cores: usize,
    pub java_installed: bool,
    pub java_versions: Vec<JavaVersion>,
}

/// Java installation information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JavaVersion {
    pub version: String,
    pub path: String,
    pub is_64bit: bool,
    pub major_version: u32,
}

/// Download progress information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DownloadProgress {
    pub id: String,
    pub name: String,
    pub downloaded_bytes: u64,
    pub total_bytes: u64,
    pub progress_percent: f64,
    pub speed_bps: u64,
    pub status: DownloadStatus,
}

/// Download status
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum DownloadStatus {
    Pending,
    Downloading,
    Paused,
    Completed,
    Failed(String),
    Cancelled,
}

/// Instance creation configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceConfig {
    pub name: String,
    pub modpack_id: String,
    pub minecraft_version: String,
    pub loader_type: Modloader,
    pub loader_version: String,
    pub memory_mb: u32,
    pub java_path: String,
    pub icon_path: Option<String>,
    pub jvm_args: Option<String>,
    pub env_vars: Option<std::collections::HashMap<String, String>>,
}

/// Instance status
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum InstanceStatus {
    Ready,
    NeedsUpdate,
    Broken,
    Installing,
    Running,
    Updating,
}

/// Instance information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Instance {
    pub id: String,
    pub name: String,
    pub modpack_id: String,
    pub minecraft_version: String,
    pub loader_type: String, // "vanilla", "forge", "fabric", etc.
    pub loader_version: String,
    pub memory_mb: u32,
    pub java_path: String,
    pub game_dir: String,
    pub last_played: Option<String>,
    pub total_playtime: u64, // in seconds
    pub icon_path: Option<String>,
    pub status: InstanceStatus,
    pub created_at: String,
    pub updated_at: String,
    pub jvm_args: Option<String>,
    pub env_vars: Option<std::collections::HashMap<String, String>>,
}

/// MultiMC instance configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MultiMCInstanceConfig {
    pub instance_type: String,
    pub name: String,
    pub icon_key: String,
    pub override_memory: bool,
    pub min_mem_alloc: u32,
    pub max_mem_alloc: u32,
    pub override_java: bool,
    pub java_path: String,
    pub notes: String,
    pub jvm_args: Option<String>,
}

/// MultiMC component for mmc-pack.json
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MultiMCComponent {
    pub cached_name: Option<String>,
    pub cached_version: Option<String>,
    pub cached_requires: Option<Vec<ComponentRequirement>>,
    pub uid: String,
    pub version: String,
}

/// Component requirement
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComponentRequirement {
    pub equals: Option<String>,
    pub uid: String,
}

/// Instance validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceValidation {
    pub instance_id: String,
    pub is_valid: bool,
    pub issues: Vec<String>,
    pub recommendations: Vec<String>,
}

/// Error types for the launcher
#[derive(Debug, Clone, Serialize, Deserialize, thiserror::Error)]
pub enum LauncherError {
    #[error("Network error: {0}")]
    Network(String),

    #[error("File system error: {0}")]
    FileSystem(String),

    #[error("Java not found")]
    JavaNotFound,

    #[error("Prism Launcher not found")]
    PrismNotFound,

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Download failed: {0}")]
    DownloadFailed(String),

    #[error("Process error: {0}")]
    Process(String),

    #[error("Serialization error: {0}")]
    Serialization(String),

    #[error("Permission denied: {0}")]
    PermissionDenied(String),

    #[error("Modpack not found: {0}")]
    ModpackNotFound(String),

    #[error("Instance not found: {0}")]
    InstanceNotFound(String),

    #[error("Launch failed: {0}")]
    LaunchFailed(String),

    #[error("Process not found: {0}")]
    ProcessNotFound(String),

    #[error("Process termination failed: {0}")]
    ProcessTermination(String),

    #[error("Update failed: {0}")]
    UpdateFailed(String),

    #[error("Not implemented: {0}")]
    NotImplemented(String),

    #[error("Not found: {0}")]
    NotFound(String),
}

impl From<reqwest::Error> for LauncherError {
    fn from(err: reqwest::Error) -> Self {
        LauncherError::Network(err.to_string())
    }
}

impl From<std::io::Error> for LauncherError {
    fn from(err: std::io::Error) -> Self {
        LauncherError::FileSystem(err.to_string())
    }
}

impl From<serde_json::Error> for LauncherError {
    fn from(err: serde_json::Error) -> Self {
        LauncherError::Serialization(err.to_string())
    }
}

impl From<zip::result::ZipError> for LauncherError {
    fn from(err: zip::result::ZipError) -> Self {
        LauncherError::FileSystem(format!("Zip error: {}", err))
    }
}

/// Backup information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackupInfo {
    pub id: String,
    pub instance_id: String,
    pub backup_date: String,
    pub version: String,
    pub size_bytes: u64,
    pub backup_path: String,
    pub description: Option<String>,
}

/// Pack installation progress
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackInstallProgress {
    pub instance_id: String,
    pub modpack_id: String,
    pub step: PackInstallStep,
    pub progress_percent: f64,
    pub message: String,
    pub status: PackInstallStatus,
}

/// Pack installation steps
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PackInstallStep {
    Downloading,
    Extracting,
    ParsingManifest,
    DownloadingDependencies,
    InstallingFiles,
    ConfiguringInstance,
    Completed,
}

/// Pack installation status
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PackInstallStatus {
    Running,
    Completed,
    Failed(String),
    Cancelled,
}

/// Manual download requirement
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ManualDownload {
    pub id: String,
    pub filename: String,
    pub url: String,
    pub checksum: Option<String>,
    pub size: u64,
    pub download_type: String, // "direct", "adfoc.us", etc.
    pub instructions: Option<String>,
}

/// Pack manifest information (from pack.toml)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackManifest {
    pub name: String,
    pub version: String,
    pub description: Option<String>,
    pub authors: Vec<String>,
    pub minecraft_version: String,
    pub pack_format: String,
    pub index_file: String,
    pub files: Vec<PackFile>,
    pub modloader: Option<String>,
    pub dependencies: Vec<String>,
    pub icon_path: Option<String>,
    pub pack_url: String,
    pub file_size: u64,
    pub checksum: Option<String>,
    pub download_count: u64,
    pub last_updated: String,
}

/// File entry in pack manifest
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackFile {
    pub file: String,
    pub hash: String,
    pub hash_format: String,
    pub download: String,
    pub filesize: u64,
    pub required: bool,
    pub side: Option<String>, // "client", "server", "both"
    pub metadata: Option<serde_json::Value>,
}

/// Update options
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateOptions {
    pub create_backup: bool,
    pub backup_description: Option<String>,
    pub force_update: bool,
    pub allow_downgrade: bool,
}

impl Default for UpdateOptions {
    fn default() -> Self {
        Self {
            create_backup: true,
            backup_description: None,
            force_update: false,
            allow_downgrade: false,
        }
    }
}

/// Result type for launcher operations
pub type LauncherResult<T> = Result<T, LauncherError>;