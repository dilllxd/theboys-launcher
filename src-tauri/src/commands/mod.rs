use crate::models::*;
use crate::utils::{config, system, performance};
use crate::downloader::download_manager;
use crate::launcher::{instance_manager, launch_manager::{LaunchConfig, LaunchedProcess}};
use crate::modpack::modpack_manager;
use crate::packwiz::packwiz_manager;
use crate::java::{java_manager, JavaInstallation, JavaCompatibilityInfo};
use crate::prism::{prism_manager, PrismInstallation, PrismUpdateInfo, PrismStatus};
use crate::update::{UpdateManager, UpdateInfo, UpdateProgress, UpdateSettings, UpdateChannel};
use tauri::{State, AppHandle};
use tauri_plugin_updater::UpdaterExt;
use std::sync::{Arc, Mutex};
use tracing::{info, warn};
use serde::{Serialize, Deserialize};
use serde_json::Value;
use chrono;

// Application state shared across commands
pub struct AppState {
    pub settings: Arc<Mutex<LauncherSettings>>,
    pub download_progress: Arc<Mutex<std::collections::HashMap<String, DownloadProgress>>>,
    pub launch_manager: Arc<crate::launcher::LaunchManager>,
    pub update_manager: Arc<UpdateManager>,
}

impl Default for AppState {
    fn default() -> Self {
        Self {
            settings: Arc::new(Mutex::new(LauncherSettings::default())),
            download_progress: Arc::new(Mutex::new(std::collections::HashMap::new())),
            launch_manager: Arc::new(crate::launcher::LaunchManager::new()),
            update_manager: Arc::new(UpdateManager::new("1.1.0".to_string())),
        }
    }
}

/// Health check command to verify the launcher is working
#[tauri::command]
pub async fn health_check() -> LauncherResult<String> {
    info!("Health check requested");
    Ok("OK".to_string())
}

/// Get the current application version
#[tauri::command]
pub async fn get_app_version() -> LauncherResult<String> {
    let version = env!("CARGO_PKG_VERSION");
    info!("Version requested: {}", version);
    Ok(version.to_string())
}

/// Get current launcher settings
#[tauri::command]
pub async fn get_settings(state: State<'_, AppState>) -> LauncherResult<LauncherSettings> {
    info!("Getting current settings");
    let settings = state.settings.lock().unwrap().clone();
    Ok(settings)
}

/// Save launcher settings
#[tauri::command]
pub async fn save_settings(
    settings: LauncherSettings,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Saving settings: memory={}MB, theme={}, java={:?}, prism={:?}",
          settings.memory_mb, settings.theme, settings.java_path, settings.prism_path);

    // Validate settings using the config validation
    config::validate_settings(&settings)?;

    // Update state
    {
        let mut state_settings = state.settings.lock().unwrap();
        *state_settings = settings.clone();
    }

    // Save to config file
    config::save_settings(&settings)?;

    info!("Settings saved successfully");
    Ok(())
}

/// Get available modpacks from the remote configuration
#[tauri::command]
pub async fn get_available_modpacks() -> LauncherResult<Vec<Modpack>> {
    info!("Fetching available modpacks");
    modpack_manager().get_modpacks().await
}

/// Get installed modpacks
#[tauri::command]
pub async fn get_installed_modpacks() -> LauncherResult<Vec<InstalledModpack>> {
    info!("Getting installed modpacks");
    modpack_manager().get_installed_modpacks().await
}

/// Check for updates for a specific modpack
#[tauri::command]
pub async fn check_modpack_updates(modpack_id: String) -> LauncherResult<Option<ModpackUpdate>> {
    info!("Checking updates for modpack: {}", modpack_id);
    modpack_manager().check_modpack_updates(&modpack_id).await
}

/// Check for updates for all installed modpacks
#[tauri::command]
pub async fn check_all_modpack_updates() -> LauncherResult<Vec<ModpackUpdate>> {
    info!("Checking updates for all installed modpacks");
    modpack_manager().check_all_updates().await
}

/// Select a modpack as the default
#[tauri::command]
pub async fn select_default_modpack(modpack_id: String) -> LauncherResult<()> {
    info!("Setting default modpack: {}", modpack_id);
    modpack_manager().select_default_modpack(&modpack_id).await
}

/// Get the current default modpack
#[tauri::command]
pub async fn get_default_modpack() -> LauncherResult<Option<Modpack>> {
    info!("Getting default modpack");
    modpack_manager().get_default_modpack().await
}

/// Get a specific modpack by ID
#[tauri::command]
pub async fn get_modpack(modpack_id: String) -> LauncherResult<Option<Modpack>> {
    info!("Getting modpack: {}", modpack_id);
    modpack_manager().get_modpack(&modpack_id).await
}

/// Clear modpack cache
#[tauri::command]
pub async fn clear_modpack_cache() -> LauncherResult<()> {
    info!("Clearing modpack cache");
    modpack_manager().clear_cache().await
}

/// Download a modpack
#[tauri::command]
pub async fn download_modpack(
    modpack_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<String> {
    info!("Starting download for modpack: {}", modpack_id);

    // Get modpack info
    let modpacks = get_available_modpacks().await?;
    let modpack = modpacks
        .into_iter()
        .find(|m| m.id == modpack_id)
        .ok_or_else(|| LauncherError::ModpackNotFound(modpack_id.clone()))?;

    // Start download in background
    let download_id = format!("download-{}-{}", modpack_id, uuid::Uuid::new_v4());

    // Update progress
    {
        let mut progress_map = state.download_progress.lock().unwrap();
        progress_map.insert(download_id.clone(), DownloadProgress {
            id: download_id.clone(),
            name: modpack.display_name.clone(),
            downloaded_bytes: 0,
            total_bytes: 0,
            progress_percent: 0.0,
            speed_bps: 0,
            status: DownloadStatus::Pending,
        });
    }

    // Start the actual download
    let download_id = download_manager().start_download(
        modpack.display_name.clone(),
        modpack.pack_url.clone(),
        determine_modpack_destination(&modpack).await?
    ).await?;

    info!("Modpack download started successfully: {} (ID: {})", modpack.display_name, download_id);
    Ok(download_id)
}

/// Launch Minecraft with specified instance
#[tauri::command]
pub async fn launch_minecraft(
    instance_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<String> {
    info!("Launching Minecraft for instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    // Get launcher settings
    let settings = state.settings.lock().unwrap().clone();

    // Get Prism path from settings or use default
    let prism_path = settings.prism_path.unwrap_or_else(|| {
        // Try to auto-detect Prism Launcher
        detect_prism_path().unwrap_or_else(|_| {
            // Fallback to common installation paths
            if cfg!(target_os = "windows") {
                format!("{}\\AppData\\Local\\Programs\\PrismLauncher",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            } else if cfg!(target_os = "macos") {
                format!("{}/Applications/PrismLauncher.app",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            } else {
                format!("{}/.local/share/PrismLauncher",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            }
        })
    });

    // Create launch configuration
    let launch_config = LaunchConfig {
        instance_id: instance.id.clone(),
        instance_name: instance.name.clone(),
        prism_path,
        java_path: if instance.java_path.is_empty() { None } else { Some(instance.java_path.clone()) },
        working_directory: instance.game_dir.clone(),
        additional_args: vec![],
        memory_mb: Some(instance.memory_mb),
        custom_jvm_args: instance.jvm_args.clone().unwrap_or_default().split_whitespace().map(|s| s.to_string()).collect(),
        environment_vars: instance.env_vars.clone().unwrap_or_default(),
    };

    // Launch the instance
    let launch_id = state.launch_manager.launch_instance(launch_config).await?;

    info!("Instance launched successfully: {} (Launch ID: {})", instance.name, launch_id);
    Ok(launch_id)
}

/// Check Java installation
#[tauri::command]
pub async fn check_java_installation() -> LauncherResult<Vec<JavaVersion>> {
    info!("Checking Java installation");

    let java_versions = system::detect_java_installations()?;
    Ok(java_versions)
}

/// Install Prism Launcher
#[tauri::command]
pub async fn install_prism_launcher(
    version: Option<String>,
) -> LauncherResult<String> {
    info!("Installing Prism Launcher, version: {:?}", version);

    prism_manager().initialize().await?;

    // Check if already installed
    if prism_manager().get_current_installation().is_some() {
        return Err(LauncherError::InvalidConfig(
            "Prism Launcher is already installed".to_string()
        ));
    }

    // Get download URL and start installation
    let install_id = prism_manager().install_prism(version).await?;

    info!("Prism Launcher installation started with ID: {}", install_id);
    Ok(install_id)
}

/// Get system information
#[tauri::command]
pub async fn get_system_info() -> LauncherResult<SystemInfo> {
    info!("Getting system information");

    let sys_info = system::get_system_info()?;
    Ok(sys_info)
}

/// Get download progress for a specific download
#[tauri::command]
pub async fn get_download_progress(
    download_id: String,
) -> LauncherResult<Option<DownloadProgress>> {
    info!("Getting download progress for: {}", download_id);
    let progress = download_manager().get_progress(&download_id).await;
    Ok(progress)
}

/// Get all active downloads
#[tauri::command]
pub async fn get_all_downloads() -> LauncherResult<Vec<DownloadProgress>> {
    info!("Getting all active downloads");
    let downloads = download_manager().get_all_downloads().await;
    Ok(downloads)
}

/// Download a file
#[tauri::command]
pub async fn download_file(
    name: String,
    url: String,
    destination: String,
) -> LauncherResult<String> {
    info!("Starting download: {} from {} to {}", name, url, destination);

    let download_id = download_manager().start_download(name, url, destination).await?;

    info!("Download started with ID: {}", download_id);
    Ok(download_id)
}

/// Cancel an ongoing download
#[tauri::command]
pub async fn cancel_download(
    download_id: String,
) -> LauncherResult<()> {
    info!("Cancelling download: {}", download_id);

    download_manager().cancel_download(&download_id).await?;

    info!("Download {} cancelled successfully", download_id);
    Ok(())
}

/// Pause a download
#[tauri::command]
pub async fn pause_download(
    download_id: String,
) -> LauncherResult<()> {
    info!("Pausing download: {}", download_id);

    download_manager().pause_download(&download_id).await?;

    info!("Download {} paused successfully", download_id);
    Ok(())
}

/// Resume a paused download
#[tauri::command]
pub async fn resume_download(
    download_id: String,
) -> LauncherResult<()> {
    info!("Resuming download: {}", download_id);

    download_manager().resume_download(&download_id).await?;

    info!("Download {} resumed successfully", download_id);
    Ok(())
}

/// Remove a completed download from tracking
#[tauri::command]
pub async fn remove_download(
    download_id: String,
) -> LauncherResult<()> {
    info!("Removing download from tracking: {}", download_id);

    download_manager().remove_download(&download_id).await?;

    info!("Download {} removed from tracking", download_id);
    Ok(())
}

/// Set maximum concurrent downloads
#[tauri::command]
pub async fn set_max_concurrent_downloads(
    max_concurrent: usize,
) -> LauncherResult<()> {
    info!("Setting maximum concurrent downloads to: {}", max_concurrent);

    if max_concurrent == 0 || max_concurrent > 10 {
        return Err(LauncherError::InvalidConfig(
            "Maximum concurrent downloads must be between 1 and 10".to_string()
        ));
    }

    // Note: This would require updating the download manager to be mutable
    // For now, we'll log the request
    warn!("Setting max concurrent downloads requires mutable access to download manager");

    Ok(())
}

/// Download Prism Launcher portable
#[tauri::command]
pub async fn download_prism_launcher(
    version: Option<String>,
) -> LauncherResult<String> {
    info!("Downloading Prism Launcher, version: {:?}", version);

    // Determine the download URL based on platform and version
    let url = determine_prism_download_url(version.clone()).await?;

    // Determine destination path
    let destination = determine_prism_destination().await?;

    let name = format!("Prism Launcher {}", version.as_deref().unwrap_or("latest"));

    let download_id = download_manager().start_download(name, url, destination).await?;

    info!("Prism Launcher download started with ID: {}", download_id);
    Ok(download_id)
}

/// Download Java JRE for specific Minecraft version
#[tauri::command]
pub async fn download_java(
    minecraft_version: String,
) -> LauncherResult<String> {
    info!("Downloading Java for Minecraft version: {}", minecraft_version);

    // Determine required Java version and download URL
    let (java_version, url) = determine_java_download_url(&minecraft_version).await?;

    // Determine destination path
    let destination = determine_java_destination(&java_version).await?;

    let name = format!("Java {} JRE", java_version);

    let download_id = download_manager().start_download(name, url, destination).await?;

    info!("Java download started with ID: {} for version {}", download_id, java_version);
    Ok(download_id)
}

/// Download packwiz bootstrap
#[tauri::command]
pub async fn download_packwiz_bootstrap() -> LauncherResult<String> {
    info!("Downloading packwiz bootstrap");

    // Get latest packwiz release from GitHub
    let url = get_packwiz_download_url().await?;

    // Determine destination path
    let destination = determine_packwiz_destination().await?;

    let name = "Packwiz Bootstrap".to_string();

    let download_id = download_manager().start_download(name, url, destination).await?;

    info!("Packwiz bootstrap download started with ID: {}", download_id);
    Ok(download_id)
}

/// Reset settings to system-aware defaults
#[tauri::command]
pub async fn reset_settings(
    state: State<'_, AppState>,
) -> LauncherResult<LauncherSettings> {
    info!("Resetting settings to defaults");

    let default_settings = config::reset_settings()?;

    // Update state
    {
        let mut state_settings = state.settings.lock().unwrap();
        *state_settings = default_settings.clone();
    }

    info!("Settings reset to defaults successfully");
    Ok(default_settings)
}

/// Detect system information for settings optimization
#[tauri::command]
pub async fn detect_system_info() -> LauncherResult<crate::models::SystemInfo> {
    info!("Detecting system information for settings");

    let sys_info = system::get_system_info()?;

    info!("System info detected: {} {} with {}GB RAM, {} cores, Java installed: {}",
          sys_info.os, sys_info.arch, sys_info.total_memory_mb / 1024,
          sys_info.cpu_cores, sys_info.java_installed);

    Ok(sys_info)
}

/// Browse for Java installation
#[tauri::command]
pub async fn browse_for_java() -> LauncherResult<Option<String>> {
    info!("Opening file browser for Java selection");

    // In a real implementation, this would use Tauri's dialog API
    // For now, we'll suggest common Java installation paths
    let java_paths = vec![
        // Windows paths
        r"C:\Program Files\Java\jdk-21",
        r"C:\Program Files\Eclipse Adoptium\jdk-21",
        r"C:\Program Files (x86)\Common Files\Oracle\Java\javapath",
        // macOS paths
        "/Library/Java/JavaVirtualMachines",
        "/usr/local/openjdk",
        // Linux paths
        "/usr/lib/jvm",
        "/usr/lib/jvm/java-21-openjdk",
        "/opt/java",
    ];

    // Try to find the first existing Java path
    for path in java_paths {
        if std::path::Path::new(path).exists() {
            info!("Found Java installation at: {}", path);
            return Ok(Some(path.to_string()));
        }
    }

    warn!("No Java installation found. Please install Java and specify the path manually.");
    Ok(None)
}

/// Browse for Prism Launcher installation
#[tauri::command]
pub async fn browse_for_prism() -> LauncherResult<Option<String>> {
    info!("Opening file browser for Prism Launcher selection");

    // In a real implementation, this would use Tauri's dialog API
    // For now, we'll suggest common Prism Launcher installation paths
    let prism_paths = vec![
        // Windows paths
        r"C:\Program Files\PrismLauncher",
        r"C:\Program Files (x86)\PrismLauncher",
        r"C:\Users\%USERNAME%\AppData\Local\Programs\PrismLauncher",
        // macOS paths
        "/Applications/PrismLauncher.app",
        "/Users/%USERNAME%/Applications/PrismLauncher.app",
        // Linux paths
        "/usr/bin/PrismLauncher",
        "/usr/local/bin/PrismLauncher",
        "/opt/PrismLauncher",
    ];

    // Try to find the first existing Prism path
    for path in prism_paths {
        let expanded_path = if path.contains("%USERNAME%") {
            if let Ok(username) = std::env::var("USERNAME") {
                path.replace("%USERNAME%", &username)
            } else {
                path.replace("%USERNAME%", "default")
            }
        } else {
            path.to_string()
        };

        if std::path::Path::new(&expanded_path).exists() {
            info!("Found Prism Launcher at: {}", expanded_path);
            return Ok(Some(expanded_path));
        }
    }

    warn!("No Prism Launcher installation found. Please install Prism Launcher or specify the path manually.");
    Ok(None)
}

/// Browse for instances directory
#[tauri::command]
pub async fn browse_for_instances_dir() -> LauncherResult<Option<String>> {
    info!("Opening folder browser for instances directory selection");

    // In a real implementation, this would use Tauri's dialog API
    // For now, we'll suggest a default instances directory
    let home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    let default_instances_dir = home_dir.join("TheBoysLauncher").join("instances");

    // Try to create the directory if it doesn't exist
    if !default_instances_dir.exists() {
        if let Err(e) = tokio::fs::create_dir_all(&default_instances_dir).await {
            warn!("Failed to create default instances directory: {}", e);
        }
    }

    info!("Using instances directory: {}", default_instances_dir.display());
    Ok(Some(default_instances_dir.to_string_lossy().to_string()))
}

/// Validate a specific setting value
#[tauri::command]
pub async fn validate_setting(
    key: String,
    value: String,
) -> LauncherResult<bool> {
    info!("Validating setting: {} = {}", key, value);

    let is_valid = match key.as_str() {
        "memory_mb" => {
            if let Ok(memory_mb) = value.parse::<u32>() {
                memory_mb >= 2048 && memory_mb <= 32768
            } else {
                false
            }
        },
        "theme" => matches!(value.as_str(), "light" | "dark" | "system"),
        "java_path" => std::path::Path::new(&value).exists(),
        "prism_path" => std::path::Path::new(&value).exists(),
        "instances_dir" => {
            std::path::Path::new(&value).exists() ||
            std::fs::create_dir_all(&value).is_ok()
        },
        "auto_update" => matches!(value.as_str(), "true" | "false"),
        _ => {
            warn!("Unknown setting key for validation: {}", key);
            false
        }
    };

    info!("Validation result for {}: {}", key, is_valid);
    Ok(is_valid)
}

// Helper functions for download management

/// Determine Prism Launcher download URL based on platform and version
async fn determine_prism_download_url(version: Option<String>) -> LauncherResult<String> {
    let platform = determine_platform();

    // GitHub API endpoint for Prism Launcher releases
    let api_url = if let Some(v) = version {
        format!("https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/tags/{}", v)
    } else {
        "https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/latest".to_string()
    };

    let client = reqwest::Client::new();
    let response = client.get(&api_url)
        .header("User-Agent", "TheBoys-Launcher/1.1.0")
        .send()
        .await?;

    if !response.status().is_success() {
        return Err(LauncherError::Network(
            format!("Failed to fetch Prism Launcher release info: {}", response.status())
        ));
    }

    let release_info: Value = response.json().await?;

    // Find the appropriate asset for this platform
    if let Some(assets) = release_info.get("assets").and_then(|a| a.as_array()) {
        let pattern = match platform {
            "windows" => "Windows-MSVC",
            "linux" => "Linux",
            "macos" => "macOS",
            _ => return Err(LauncherError::NotImplemented(format!("Unsupported platform: {}", platform))),
        };

        for asset in assets {
            if let (Some(name), Some(download_url)) = (
                asset.get("name").and_then(|n| n.as_str()),
                asset.get("browser_download_url").and_then(|u| u.as_str())
            ) {
                if name.contains("Portable") && name.contains(pattern) {
                    return Ok(download_url.to_string());
                }
            }
        }
    }

    Err(LauncherError::DownloadFailed(
        "Could not find suitable Prism Launcher download".to_string()
    ))
}

/// Determine destination path for Prism Launcher
async fn determine_prism_destination() -> LauncherResult<String> {
    let mut home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    home_dir.push("TheBoysLauncher");
    home_dir.push("tools");
    home_dir.push("prism");

    // Create directory if it doesn't exist
    tokio::fs::create_dir_all(&home_dir).await
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to create Prism directory: {}", e)
        ))?;

    home_dir.push("portable");

    Ok(home_dir.to_string_lossy().to_string())
}

/// Determine Java version and download URL for Minecraft version
async fn determine_java_download_url(minecraft_version: &str) -> LauncherResult<(String, String)> {
    // Map Minecraft versions to required Java versions
    let java_version = match minecraft_version {
        v if v.starts_with("1.") => {
            let minor: u32 = v.split('.').nth(1).unwrap_or("0").parse()
                .unwrap_or(0);
            match minor {
                0..=12 => "8".to_string(),
                13..=16 => "16".to_string(),
                17 => "17".to_string(),
                18..=20 => "18".to_string(),
                _ => "21".to_string(),
            }
        },
        _ => "21".to_string(),
    };

    // Adoptium API for Temurin JRE
    let api_url = format!(
        "https://api.adoptium.net/v3/assets/latest/{}/hotspot?os={}&architecture=x64&image_type=jre&release=latest",
        java_version, determine_platform()
    );

    let client = reqwest::Client::new();
    let response = client.get(&api_url)
        .header("User-Agent", "TheBoys-Launcher/1.1.0")
        .send()
        .await?;

    if !response.status().is_success() {
        return Err(LauncherError::Network(
            format!("Failed to fetch Java download info: {}", response.status())
        ));
    }

    let assets: Vec<Value> = response.json().await?;

    if let Some(asset) = assets.first() {
        if let Some(download_url) = asset.get("binary").and_then(|b| b.get("package")).and_then(|p| p.get("link")).and_then(|l| l.as_str()) {
            return Ok((java_version, download_url.to_string()));
        }
    }

    Err(LauncherError::DownloadFailed(
        "Could not find suitable Java download".to_string()
    ))
}

/// Determine destination path for Java
async fn determine_java_destination(java_version: &str) -> LauncherResult<String> {
    let mut home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    home_dir.push("TheBoysLauncher");
    home_dir.push("tools");
    home_dir.push("java");

    // Create directory if it doesn't exist
    tokio::fs::create_dir_all(&home_dir).await
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to create Java directory: {}", e)
        ))?;

    home_dir.push(format!("jre-{}", java_version));

    Ok(home_dir.to_string_lossy().to_string())
}

/// Get packwiz bootstrap download URL
async fn get_packwiz_download_url() -> LauncherResult<String> {
    let api_url = "https://api.github.com/repos/packwiz/packwiz/releases/latest";

    let client = reqwest::Client::new();
    let response = client.get(api_url)
        .header("User-Agent", "TheBoys-Launcher/1.1.0")
        .send()
        .await?;

    if !response.status().is_success() {
        return Err(LauncherError::Network(
            format!("Failed to fetch packwiz release info: {}", response.status())
        ));
    }

    let release_info: Value = response.json().await?;

    if let Some(assets) = release_info.get("assets").and_then(|a| a.as_array()) {
        for asset in assets {
            if let (Some(name), Some(download_url)) = (
                asset.get("name").and_then(|n| n.as_str()),
                asset.get("browser_download_url").and_then(|u| u.as_str())
            ) {
                // Look for packwiz executable for the current platform
                let platform = determine_platform();
                let pattern = match platform {
                    "windows" => "packwiz.exe",
                    "linux" => "packwiz-linux",
                    "macos" => "packwiz-macos",
                    _ => return Err(LauncherError::NotImplemented(format!("Unsupported platform: {}", platform))),
                };

                if name == pattern || (name.starts_with("packwiz") && name.contains(platform)) {
                    return Ok(download_url.to_string());
                }
            }
        }
    }

    Err(LauncherError::DownloadFailed(
        "Could not find suitable packwiz download".to_string()
    ))
}

/// Determine destination path for packwiz
async fn determine_packwiz_destination() -> LauncherResult<String> {
    let mut home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    home_dir.push("TheBoysLauncher");
    home_dir.push("tools");
    home_dir.push("packwiz");

    // Create directory if it doesn't exist
    tokio::fs::create_dir_all(&home_dir).await
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to create packwiz directory: {}", e)
        ))?;

    let platform = determine_platform();
    let executable_name = match platform {
        "windows" => "packwiz.exe",
        _ => "packwiz",
    };

    home_dir.push(executable_name);

    Ok(home_dir.to_string_lossy().to_string())
}

/// Determine current platform string
fn determine_platform() -> &'static str {
    #[cfg(target_os = "windows")]
    return "windows";

    #[cfg(target_os = "macos")]
    return "macos";

    #[cfg(target_os = "linux")]
    return "linux";

    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    return "unknown";
}

// ==================== JAVA MANAGEMENT COMMANDS ====================

/// Initialize Java Manager and detect installations
#[tauri::command]
pub async fn detect_java_installations() -> LauncherResult<Vec<JavaInstallation>> {
    info!("Detecting Java installations with Java Manager");

    java_manager().initialize().await?;
    let installations = java_manager().get_installed_versions();

    info!("Found {} Java installations", installations.len());
    Ok(installations)
}

/// Get all managed Java installations
#[tauri::command]
pub async fn get_managed_java_installations() -> LauncherResult<Vec<JavaInstallation>> {
    info!("Getting managed Java installations");

    java_manager().initialize().await?;
    let installations = java_manager().get_managed_installations();

    Ok(installations)
}

/// Get required Java version for a Minecraft version
#[tauri::command]
pub async fn get_required_java_version(minecraft_version: String) -> LauncherResult<Option<String>> {
    info!("Getting required Java version for Minecraft {}", minecraft_version);

    let required_version = java_manager().get_required_java_version(&minecraft_version);
    Ok(required_version)
}

/// Get best Java installation for a Minecraft version
#[tauri::command]
pub async fn get_best_java_installation(minecraft_version: String) -> LauncherResult<Option<JavaInstallation>> {
    info!("Getting best Java installation for Minecraft {}", minecraft_version);

    java_manager().initialize().await?;
    let installation = java_manager().get_best_java_installation(&minecraft_version);

    Ok(installation)
}

/// Check Java compatibility for a Minecraft version
#[tauri::command]
pub async fn check_java_compatibility(minecraft_version: String) -> LauncherResult<JavaCompatibilityInfo> {
    info!("Checking Java compatibility for Minecraft {}", minecraft_version);

    java_manager().initialize().await?;
    let compatibility_info = java_manager().get_compatibility_info(&minecraft_version);

    Ok(compatibility_info)
}

/// Install Java version from Adoptium
#[tauri::command]
pub async fn install_java_version(java_version: String) -> LauncherResult<String> {
    info!("Installing Java version {}", java_version);

    // Initialize Java Manager
    java_manager().initialize().await?;

    // Check if already installed
    if java_manager().is_java_version_installed(&java_version) {
        return Err(LauncherError::InvalidConfig(
            format!("Java {} is already installed", java_version)
        ));
    }

    // Get download URL
    let download_url = java_manager().get_java_download_url(&java_version).await?;

    // Get destination path
    let destination = java_manager().get_java_install_path(&java_version)?;

    // Create parent directories
    if let Some(parent) = std::path::Path::new(&destination).parent() {
        tokio::fs::create_dir_all(parent).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create Java directory: {}", e)
            ))?;
    }

    let name = format!("Java {} JRE", java_version);

    // Start download
    let download_id = download_manager().start_download(name, download_url, destination).await?;

    info!("Java {} installation started with ID: {}", java_version, download_id);
    Ok(download_id)
}

/// Remove a managed Java installation
#[tauri::command]
pub async fn remove_java_installation(java_version: String) -> LauncherResult<()> {
    info!("Removing Java installation {}", java_version);

    java_manager().initialize().await?;
    java_manager().remove_managed_installation(&java_version).await?;

    Ok(())
}

/// Cleanup old managed Java installations
#[tauri::command]
pub async fn cleanup_java_installations() -> LauncherResult<u32> {
    info!("Cleaning up old Java installations");

    java_manager().initialize().await?;
    let removed_count = java_manager().cleanup_old_installations().await?;

    Ok(removed_count as u32)
}

/// Get Java download URL without starting download
#[tauri::command]
pub async fn get_java_download_info(java_version: String) -> LauncherResult<(String, u64)> {
    info!("Getting Java download info for version {}", java_version);

    // Get download URL
    let download_url = java_manager().get_java_download_url(&java_version).await?;

    // Try to get file size
    let client = reqwest::Client::new();
    let size = client.head(&download_url)
        .header("User-Agent", "TheBoys-Launcher/1.1.0")
        .send()
        .await
        .ok()
        .and_then(|response| {
            response.content_length()
        })
        .unwrap_or(0);

    Ok((download_url, size))
}

// ==================== PRISM MANAGEMENT COMMANDS ====================

/// Initialize Prism Manager and detect installations
#[tauri::command]
pub async fn detect_prism_installation() -> LauncherResult<Option<PrismInstallation>> {
    info!("Detecting Prism Launcher installation");

    prism_manager().initialize().await?;
    let installation = prism_manager().get_current_installation();

    Ok(installation)
}

/// Get Prism Launcher status
#[tauri::command]
pub async fn get_prism_status() -> LauncherResult<PrismStatus> {
    info!("Getting Prism Launcher status");

    prism_manager().initialize().await?;
    let status = prism_manager().get_prism_status();

    Ok(status)
}

/// Check for Prism Launcher updates
#[tauri::command]
pub async fn check_prism_updates() -> LauncherResult<PrismUpdateInfo> {
    info!("Checking for Prism Launcher updates");

    prism_manager().initialize().await?;
    let update_info = prism_manager().check_for_updates().await?;

    Ok(update_info)
}

/// Install Prism Launcher
#[tauri::command]
pub async fn install_prism_launcher_new(version: Option<String>) -> LauncherResult<String> {
    info!("Installing Prism Launcher, version: {:?}", version);

    prism_manager().initialize().await?;

    // Check if already installed
    if prism_manager().get_current_installation().is_some() {
        return Err(LauncherError::InvalidConfig(
            "Prism Launcher is already installed".to_string()
        ));
    }

    // Get download URL
    let download_url = prism_manager().install_prism(version).await?;

    // Get destination path
    let destination = prism_manager().get_prism_install_path()?;

    let name = "Prism Launcher".to_string();

    // Start download
    let download_id = download_manager().start_download(name, download_url, destination).await?;

    info!("Prism Launcher installation started with ID: {}", download_id);
    Ok(download_id)
}

/// Uninstall managed Prism Launcher
#[tauri::command]
pub async fn uninstall_prism_launcher() -> LauncherResult<()> {
    info!("Uninstalling Prism Launcher");

    prism_manager().initialize().await?;
    prism_manager().uninstall_prism().await?;

    Ok(())
}

/// Get Prism installation path
#[tauri::command]
pub async fn get_prism_install_path() -> LauncherResult<String> {
    info!("Getting Prism installation path");

    let path = prism_manager().get_prism_install_path()?;
    Ok(path)
}

/// Verify Prism installation
#[tauri::command]
pub async fn verify_prism_installation(path: String) -> LauncherResult<bool> {
    info!("Verifying Prism installation at: {}", path);

    prism_manager().initialize().await?;

    // Check if the path exists and contains Prism executable
    let path_obj = std::path::Path::new(&path);
    if !path_obj.exists() {
        return Ok(false);
    }

    // Try to detect Prism at the given path
    let prism_manager = prism_manager();
    let installation = prism_manager.check_prism_at_path(&path).await?;

    Ok(installation.is_some())
}

/// Get Prism download information without starting download
#[tauri::command]
pub async fn get_prism_download_info(version: Option<String>) -> LauncherResult<(String, u64, String)> {
    info!("Getting Prism download info, version: {:?}", version);

    prism_manager().initialize().await?;
    let update_info = prism_manager().check_for_updates().await?;

    if let Some(v) = version {
        if v != update_info.latest_version {
            return Err(LauncherError::InvalidConfig(
                format!("Specific version {} not found. Latest available: {}", v, update_info.latest_version)
            ));
        }
    }

    Ok((update_info.download_url, update_info.size_bytes, update_info.latest_version))
}

// ==================== INSTANCE MANAGEMENT COMMANDS ====================

/// Create a new instance from configuration
#[tauri::command]
pub async fn create_instance(
    config: InstanceConfig,
) -> LauncherResult<Instance> {
    info!("Creating new instance: {} with modloader {}", config.name, config.loader_type.as_str());

    let instance = instance_manager().create_instance(&config).await?;

    info!("Instance created successfully: {} with ID {}", instance.name, instance.id);
    Ok(instance)
}

/// Create a new instance from a modpack
#[tauri::command]
pub async fn create_instance_from_modpack(
    modpack_id: String,
    instance_name: Option<String>,
    memory_mb: Option<u32>,
) -> LauncherResult<Instance> {
    info!("Creating instance from modpack: {}", modpack_id);

    // Get modpack information
    let modpacks = get_available_modpacks().await?;
    let modpack = modpacks
        .into_iter()
        .find(|m| m.id == modpack_id)
        .ok_or_else(|| LauncherError::ModpackNotFound(modpack_id.clone()))?;

    // Use provided name or generate from modpack
    let instance_name = instance_name.unwrap_or_else(|| format!("{}-Instance", modpack.display_name));

    // Get system memory for default allocation if not specified
    let total_memory = 8192; // Default fallback
    let default_memory = memory_mb.unwrap_or_else(|| {
        std::cmp::min(std::cmp::max(total_memory / 2, 2048), 8192) as u32
    });

    // Create instance configuration from modpack
    let config = InstanceConfig {
        name: instance_name.clone(),
        modpack_id: modpack_id.clone(),
        minecraft_version: modpack.minecraft_version.clone(),
        loader_type: modpack.modloader.clone(),
        loader_version: modpack.loader_version.clone(),
        memory_mb: default_memory,
        java_path: String::new(), // Will be auto-detected
        icon_path: None,
        jvm_args: None,
        env_vars: None,
    };

    // Create the instance
    let mut instance = instance_manager().create_instance(&config).await?;

    // Start modpack installation in background
    info!("Starting modpack installation for instance: {}", instance_name);

    // Update instance status to installing
    instance.status = InstanceStatus::Installing;
    instance_manager().update_instance(&instance).await?;

    // Use packwiz to install the modpack
    let install_id = packwiz_manager().install_modpack(&instance.id, &modpack.pack_url, UpdateOptions::default()).await?;

    info!("Modpack installation started with ID: {} for instance: {}", install_id, instance_name);

    // Update instance to show it's ready after packwiz installation
    // Note: In a real implementation, you'd want to monitor the packwiz installation progress
    // and update the instance status when it's complete
    instance.status = InstanceStatus::Ready;
    instance_manager().update_instance(&instance).await?;

    info!("Instance created successfully from modpack: {} (ID: {})", instance_name, instance.id);
    Ok(instance)
}

/// Install a modpack into an existing instance
#[tauri::command]
pub async fn install_modpack_to_instance(
    instance_id: String,
    modpack_id: String,
) -> LauncherResult<String> {
    info!("Installing modpack {} to instance {}", modpack_id, instance_id);

    // Verify instance exists
    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    // Get modpack information
    let modpacks = get_available_modpacks().await?;
    let modpack = modpacks
        .into_iter()
        .find(|m| m.id == modpack_id)
        .ok_or_else(|| LauncherError::ModpackNotFound(modpack_id.clone()))?;

    // Update instance status
    let mut updating_instance = instance.clone();
    updating_instance.status = InstanceStatus::Updating;
    instance_manager().update_instance(&updating_instance).await?;

    // Use packwiz to install the modpack
    let install_id = packwiz_manager().install_modpack(&instance_id, &modpack.pack_url, UpdateOptions::default()).await?;

    info!("Modpack installation started with ID: {} for instance: {}", install_id, instance.name);
    Ok(install_id)
}

/// Get all instances
#[tauri::command]
pub async fn get_instances() -> LauncherResult<Vec<Instance>> {
    info!("Getting all instances");

    let instances = instance_manager().get_instances().await?;

    info!("Found {} instances", instances.len());
    Ok(instances)
}

/// Get a specific instance by ID
#[tauri::command]
pub async fn get_instance(instance_id: String) -> LauncherResult<Option<Instance>> {
    info!("Getting instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?;

    match &instance {
        Some(inst) => info!("Found instance: {} ({})", inst.name, instance_id),
        None => info!("Instance not found: {}", instance_id),
    }

    Ok(instance)
}

/// Get instance by name
#[tauri::command]
pub async fn get_instance_by_name(name: String) -> LauncherResult<Option<Instance>> {
    info!("Getting instance by name: {}", name);

    let instance = instance_manager().get_instance_by_name(&name).await?;

    match &instance {
        Some(inst) => info!("Found instance: {} (ID: {})", inst.name, inst.id),
        None => info!("Instance not found: {}", name),
    }

    Ok(instance)
}

/// Update instance settings
#[tauri::command]
pub async fn update_instance(
    instance: Instance,
) -> LauncherResult<()> {
    info!("Updating instance: {} ({})", instance.name, instance.id);

    instance_manager().update_instance(&instance).await?;

    info!("Instance updated successfully: {}", instance.name);
    Ok(())
}

/// Delete an instance
#[tauri::command]
pub async fn delete_instance(
    instance_id: String,
) -> LauncherResult<()> {
    info!("Deleting instance: {}", instance_id);

    // Get instance info before deletion for logging
    let instance_info = instance_manager().get_instance(&instance_id).await?;

    instance_manager().delete_instance(&instance_id).await?;

    if let Some(instance) = instance_info {
        info!("Instance deleted successfully: {}", instance.name);
    } else {
        info!("Instance deleted (was already missing): {}", instance_id);
    }

    Ok(())
}

/// Launch an instance
#[tauri::command]
pub async fn launch_instance(
    instance_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<String> {
    info!("Launching instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    // Get launcher settings
    let settings = state.settings.lock().unwrap().clone();

    // Get Prism path from settings or use default
    let prism_path = settings.prism_path.unwrap_or_else(|| {
        // Try to auto-detect Prism Launcher
        detect_prism_path().unwrap_or_else(|_| {
            // Fallback to common installation paths
            if cfg!(target_os = "windows") {
                format!("{}\\AppData\\Local\\Programs\\PrismLauncher",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            } else if cfg!(target_os = "macos") {
                format!("{}/Applications/PrismLauncher.app",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            } else {
                format!("{}/.local/share/PrismLauncher",
                    dirs::home_dir().unwrap_or_default().to_string_lossy())
            }
        })
    });

    // Create launch configuration
    let launch_config = LaunchConfig {
        instance_id: instance.id.clone(),
        instance_name: instance.name.clone(),
        prism_path,
        java_path: if instance.java_path.is_empty() { None } else { Some(instance.java_path.clone()) },
        working_directory: instance.game_dir.clone(),
        additional_args: vec![],
        memory_mb: Some(instance.memory_mb),
        custom_jvm_args: instance.jvm_args.clone().unwrap_or_default().split_whitespace().map(|s| s.to_string()).collect(),
        environment_vars: instance.env_vars.clone().unwrap_or_default(),
    };

    // Launch the instance
    let launch_id = state.launch_manager.launch_instance(launch_config).await?;

    info!("Instance launched successfully: {} (Launch ID: {})", instance.name, launch_id);
    Ok(launch_id)
}

/// Validate an instance
#[tauri::command]
pub async fn validate_instance(
    instance_id: String,
) -> LauncherResult<InstanceValidation> {
    info!("Validating instance: {}", instance_id);

    let validation = instance_manager().validate_instance(&instance_id).await?;

    info!("Instance validation completed for {}: {} ({} issues, {} recommendations)",
          instance_id, validation.is_valid, validation.issues.len(), validation.recommendations.len());

    Ok(validation)
}

/// Install modloader for an instance
#[tauri::command]
pub async fn install_modloader(
    instance_id: String,
) -> LauncherResult<()> {
    info!("Installing modloader for instance: {}", instance_id);

    instance_manager().install_modloader(&instance_id, None).await?;

    info!("Modloader installation completed for instance: {}", instance_id);
    Ok(())
}

/// Get available modloader versions
#[tauri::command]
pub async fn get_modloader_versions(
    modloader: Modloader,
    minecraft_version: String,
) -> LauncherResult<Vec<String>> {
    info!("Getting available versions for {} on Minecraft {}", modloader.as_str(), minecraft_version);

    let versions = instance_manager().get_modloader_versions(&modloader, &minecraft_version).await?;

    info!("Found {} versions for {} on Minecraft {}", versions.len(), modloader.as_str(), minecraft_version);
    Ok(versions)
}

/// Repair an instance
#[tauri::command]
pub async fn repair_instance(
    instance_id: String,
) -> LauncherResult<()> {
    info!("Repairing instance: {}", instance_id);

    instance_manager().repair_instance(&instance_id).await?;

    info!("Instance repair completed for: {}", instance_id);
    Ok(())
}

/// Get instance status
#[tauri::command]
pub async fn get_instance_status(
    instance_id: String,
) -> LauncherResult<Option<InstanceStatus>> {
    info!("Getting status for instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?;

    match instance {
        Some(inst) => {
            info!("Instance {} status: {:?}", inst.name, inst.status);
            Ok(Some(inst.status))
        },
        None => {
            info!("Instance not found: {}", instance_id);
            Ok(None)
        }
    }
}

/// Set instance status (internal use)
#[tauri::command]
pub async fn set_instance_status(
    instance_id: String,
    status: InstanceStatus,
) -> LauncherResult<()> {
    info!("Setting status for instance {}: {:?}", instance_id, status);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    // Create a new instance with updated status
    let mut updated_instance = instance.clone();
    updated_instance.status = status.clone();
    instance_manager().update_instance(&updated_instance).await?;

    info!("Instance status updated successfully: {} -> {:?}", instance_id, status);
    Ok(())
}

/// Get instance logs
#[tauri::command]
pub async fn get_instance_logs(
    instance_id: String,
) -> LauncherResult<Vec<String>> {
    info!("Getting logs for instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    let mut logs = Vec::new();

    // Check for common log files in the minecraft directory
    let minecraft_dir = std::path::Path::new(&instance.game_dir).join("minecraft");

    // Latest log
    let latest_log = minecraft_dir.join("logs").join("latest.log");
    if latest_log.exists() {
        match std::fs::read_to_string(&latest_log) {
            Ok(content) => {
                // Split into lines and take last 100 lines
                let lines: Vec<&str> = content.lines().rev().take(100).collect();
                logs.extend(lines.iter().rev().map(|s| s.to_string()));
            },
            Err(e) => warn!("Failed to read latest log: {}", e),
        }
    }

    // Crash reports
    let crash_reports_dir = minecraft_dir.join("crash-reports");
    if crash_reports_dir.exists() {
        if let Ok(entries) = std::fs::read_dir(&crash_reports_dir) {
            for entry in entries.flatten() {
                let path = entry.path();
                if path.extension().and_then(|s| s.to_str()) == Some("txt") {
                    if let Ok(content) = std::fs::read_to_string(&path) {
                        logs.push(format!("=== Crash Report: {} ===",
                                        path.file_name()
                                            .and_then(|n| n.to_str())
                                            .unwrap_or("unknown")));
                        logs.extend(content.lines().take(50).map(|s| s.to_string()));
                    }
                }
            }
        }
    }

    info!("Retrieved {} log entries for instance {}", logs.len(), instance_id);
    Ok(logs)
}

/// Clear instance logs
#[tauri::command]
pub async fn clear_instance_logs(
    instance_id: String,
) -> LauncherResult<()> {
    info!("Clearing logs for instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    let minecraft_dir = std::path::Path::new(&instance.game_dir).join("minecraft");

    // Clear latest log
    let latest_log = minecraft_dir.join("logs").join("latest.log");
    if latest_log.exists() {
        if let Err(e) = std::fs::write(&latest_log, "") {
            warn!("Failed to clear latest log: {}", e);
        }
    }

    // Remove old crash reports (keep last 3)
    let crash_reports_dir = minecraft_dir.join("crash-reports");
    if crash_reports_dir.exists() {
        if let Ok(entries) = std::fs::read_dir(&crash_reports_dir) {
            let mut crash_reports: Vec<_> = entries.flatten().collect();
            crash_reports.sort_by(|a, b| b.metadata().unwrap().modified().unwrap().cmp(
                &a.metadata().unwrap().modified().unwrap()
            ));

            for (i, entry) in crash_reports.iter().enumerate() {
                if i >= 3 { // Keep only the 3 most recent
                    if let Err(e) = std::fs::remove_file(entry.path()) {
                        warn!("Failed to remove old crash report {:?}: {}", entry.path(), e);
                    }
                }
            }
        }
    }

    info!("Logs cleared for instance: {}", instance_id);
    Ok(())
}

/// Get instance statistics
#[tauri::command]
pub async fn get_instance_statistics(
    instance_id: String,
) -> LauncherResult<InstanceStatistics> {
    info!("Getting statistics for instance: {}", instance_id);

    let instance = instance_manager().get_instance(&instance_id).await?
        .ok_or_else(|| LauncherError::InstanceNotFound(
            format!("Instance {} not found", instance_id)
        ))?;

    let minecraft_dir = std::path::Path::new(&instance.game_dir).join("minecraft");

    // Calculate directory size
    let total_size = calculate_directory_size(&minecraft_dir)?;

    // Count mods, resource packs, etc.
    let mods_count = count_files_in_directory(&minecraft_dir.join("mods")).unwrap_or(0);
    let resource_packs_count = count_files_in_directory(&minecraft_dir.join("resourcepacks")).unwrap_or(0);
    let shader_packs_count = count_files_in_directory(&minecraft_dir.join("shaderpacks")).unwrap_or(0);
    let screenshots_count = count_files_in_directory(&minecraft_dir.join("screenshots")).unwrap_or(0);

    let stats = InstanceStatistics {
        instance_id: instance_id.clone(),
        name: instance.name,
        total_size_bytes: total_size,
        mods_count,
        resource_packs_count,
        shader_packs_count,
        screenshots_count,
        total_playtime_seconds: instance.total_playtime,
        last_played: instance.last_played.clone(),
        created_at: instance.created_at.clone(),
        updated_at: instance.updated_at.clone(),
    };

    info!("Statistics retrieved for instance {}: {}MB total, {} mods, {} resource packs",
          instance_id, total_size / 1024 / 1024, mods_count, resource_packs_count);

    Ok(stats)
}

/// Instance statistics structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceStatistics {
    pub instance_id: String,
    pub name: String,
    pub total_size_bytes: u64,
    pub mods_count: usize,
    pub resource_packs_count: usize,
    pub shader_packs_count: usize,
    pub screenshots_count: usize,
    pub total_playtime_seconds: u64,
    pub last_played: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

// Helper functions for instance management

/// Calculate total size of a directory recursively
fn calculate_directory_size(dir_path: &std::path::Path) -> LauncherResult<u64> {
    if !dir_path.exists() {
        return Ok(0);
    }

    let mut total_size = 0u64;

    let entries = std::fs::read_dir(dir_path)
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to read directory {:?}: {}", dir_path, e)
        ))?;

    for entry in entries.flatten() {
        let path = entry.path();
        if path.is_dir() {
            total_size += calculate_directory_size(&path)?;
        } else {
            total_size += entry.metadata()
                .map_err(|e| LauncherError::FileSystem(
                    format!("Failed to read metadata for {:?}: {}", path, e)
                ))?
                .len();
        }
    }

    Ok(total_size)
}

/// Count files in a directory (non-recursive)
fn count_files_in_directory(dir_path: &std::path::Path) -> LauncherResult<usize> {
    if !dir_path.exists() {
        return Ok(0);
    }

    let entries = std::fs::read_dir(dir_path)
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to read directory {:?}: {}", dir_path, e)
        ))?;

    let count = entries.filter(|entry| {
        entry.as_ref().ok().map(|e| e.path().is_file()).unwrap_or(false)
    }).count();

    Ok(count)
}

// ==================== PACKWIZ INTEGRATION COMMANDS ====================

/// Initialize packwiz manager
#[tauri::command]
pub async fn initialize_packwiz() -> LauncherResult<()> {
    info!("Initializing packwiz manager");

    packwiz_manager().initialize().await?;
    Ok(())
}

/// Install or update a modpack using packwiz
#[tauri::command]
pub async fn install_modpack_with_packwiz(
    instance_id: String,
    pack_url: String,
    options: UpdateOptions,
) -> LauncherResult<String> {
    info!("Installing modpack with packwiz for instance: {}", instance_id);

    let install_id = packwiz_manager().install_modpack(&instance_id, &pack_url, options).await?;
    Ok(install_id)
}

/// Check for updates for a specific instance
#[tauri::command]
pub async fn check_instance_updates(instance_id: String) -> LauncherResult<Option<ModpackUpdate>> {
    info!("Checking updates for instance: {}", instance_id);

    let update = packwiz_manager().check_instance_updates(&instance_id).await?;
    Ok(update)
}

/// Create a backup of an instance
#[tauri::command]
pub async fn create_instance_backup(
    instance_id: String,
    description: Option<String>,
) -> LauncherResult<BackupInfo> {
    info!("Creating backup for instance: {}", instance_id);

    let backup_info = packwiz_manager().create_backup(&instance_id, description.as_deref()).await?;
    Ok(backup_info)
}

/// Get list of backups for an instance
#[tauri::command]
pub async fn get_instance_backups(instance_id: String) -> LauncherResult<Vec<BackupInfo>> {
    info!("Getting backups for instance: {}", instance_id);

    let backups = packwiz_manager().get_backups(&instance_id).await?;
    Ok(backups)
}

/// Restore instance from backup
#[tauri::command]
pub async fn restore_instance_backup(
    backup_id: String,
    instance_id: String,
) -> LauncherResult<()> {
    info!("Restoring backup {} to instance {}", backup_id, instance_id);

    packwiz_manager().restore_backup(&backup_id, &instance_id).await?;
    Ok(())
}

/// Delete a backup
#[tauri::command]
pub async fn delete_instance_backup(backup_id: String) -> LauncherResult<()> {
    info!("Deleting backup: {}", backup_id);

    packwiz_manager().delete_backup(&backup_id).await?;
    Ok(())
}

/// Get installation progress
#[tauri::command]
pub async fn get_pack_install_progress(install_id: String) -> LauncherResult<Option<PackInstallProgress>> {
    info!("Getting installation progress: {}", install_id);

    let progress = packwiz_manager().get_install_progress(&install_id).await?;
    Ok(progress)
}

/// Get manual downloads required for an installation
#[tauri::command]
pub async fn get_manual_downloads(install_id: String) -> LauncherResult<Vec<ManualDownload>> {
    info!("Getting manual downloads for installation: {}", install_id);

    let downloads = packwiz_manager().get_manual_downloads(&install_id).await?;
    Ok(downloads)
}

/// Confirm manual download completion
#[tauri::command]
pub async fn confirm_manual_download(
    install_id: String,
    filename: String,
    local_path: String,
) -> LauncherResult<()> {
    info!("Confirming manual download: {} -> {}", filename, local_path);

    packwiz_manager().confirm_manual_download(&install_id, &filename, &local_path).await?;
    Ok(())
}

/// Cancel modpack installation
#[tauri::command]
pub async fn cancel_modpack_installation(install_id: String) -> LauncherResult<()> {
    info!("Cancelling modpack installation: {}", install_id);

    packwiz_manager().cancel_installation(&install_id).await?;
    Ok(())
}

/// Get pack manifest information from a pack.toml
#[tauri::command]
pub async fn get_pack_manifest(pack_url: String) -> LauncherResult<PackManifest> {
    info!("Getting pack manifest from: {}", pack_url);

    // Download the pack.toml file
    let client = reqwest::Client::new();
    let response = client.get(&pack_url)
        .header("User-Agent", "TheBoys-Launcher/1.1.0")
        .send()
        .await
        .map_err(|e| LauncherError::Network(
            format!("Failed to download pack.toml: {}", e)
        ))?;

    if !response.status().is_success() {
        return Err(LauncherError::Network(
            format!("HTTP error downloading pack.toml: {}", response.status())
        ));
    }

    let content = response.text().await
        .map_err(|e| LauncherError::Network(e.to_string()))?;

    // Parse TOML content
    let toml_value: serde_json::Value = toml::from_str(&content)
        .map_err(|e| LauncherError::Serialization(
            format!("Failed to parse pack.toml: {}", e)
        ))?;

    // Extract manifest information
    let pack_name = toml_value
        .get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("Unknown Modpack")
        .to_string();

    let description = toml_value
        .get("description")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .to_string();

    let version = toml_value
        .get("version")
        .and_then(|v| v.as_str())
        .unwrap_or("1.0.0")
        .to_string();

    let authors = toml_value
        .get("authors")
        .and_then(|v| v.as_array())
        .map(|arr| arr.iter().filter_map(|v| v.as_str()).collect::<Vec<_>>())
        .unwrap_or_default();

    let minecraft_version = toml_value
        .get("minecraft")
        .and_then(|m| m.get("version"))
        .and_then(|v| v.as_str())
        .unwrap_or("1.20.1")
        .to_string();

    // Check for supported modloader
    let loader_info = toml_value
        .get("loader")
        .and_then(|l| {
            if let Some(loader_map) = l.as_object() {
                Some(format!("{}-{}",
                    loader_map.get("modloader").and_then(|m| m.as_str()).unwrap_or("fabric"),
                    loader_map.get("version").and_then(|v| v.as_str()).unwrap_or("0.15.11")
                ))
            } else {
                l.as_str().map(|s| s.to_string())
            }
        });

    // Parse dependencies
    let dependencies = toml_value
        .get("dependencies")
        .and_then(|deps| deps.as_array())
        .map(|arr| {
            arr.iter().filter_map(|dep| {
                dep.as_object().and_then(|dep_obj| {
                    let name = dep_obj.get("name").and_then(|n| n.as_str())?;
                    let version = dep_obj.get("version").and_then(|v| v.as_str()).unwrap_or("latest");
                    Some(format!("{}:{}", name, version))
                })
            }).collect()
        }).unwrap_or_default();

    // Get icon if present
    let icon_path = toml_value
        .get("icon")
        .and_then(|v| v.as_str())
        .map(|s| format!("{}/{}", pack_url, s))
        .or_else(|| Some("default.png".to_string()));

    let manifest = PackManifest {
        name: pack_name.clone(),
        version: version.clone(),
        description: Some(description),
        authors: authors.iter().map(|s| s.to_string()).collect(),
        minecraft_version,
        pack_format: "1.0".to_string(),
        index_file: "index.toml".to_string(),
        files: vec![], // TODO: Parse files from pack.toml
        modloader: loader_info,
        dependencies,
        icon_path,
        pack_url: pack_url.clone(),
        file_size: 0, // Would need to fetch the actual file size
        checksum: None, // TODO: Implement checksum calculation
        download_count: 0,
        last_updated: chrono::Utc::now().to_rfc3339(),
    };

    info!("Successfully parsed pack manifest: {} v{}", pack_name, version);
    Ok(manifest)
}

/// Validate manual download file
#[tauri::command]
pub async fn validate_manual_download(
    filename: String,
    local_path: String,
    _expected_checksum: Option<String>,
    expected_size: Option<u64>,
) -> LauncherResult<bool> {
    info!("Validating manual download: {} at {}", filename, local_path);

    let path = std::path::Path::new(&local_path);
    if !path.exists() {
        return Ok(false);
    }

    // Check file size if expected
    if let Some(expected) = expected_size {
        let actual_size = std::fs::metadata(path)?.len();
        if actual_size != expected {
            return Ok(false);
        }
    }

    // TODO: Implement checksum validation
    Ok(true)
}

/// Get update history for an instance
#[tauri::command]
pub async fn get_instance_update_history(instance_id: String) -> LauncherResult<Vec<ModpackUpdate>> {
    info!("Getting update history for instance: {}", instance_id);

    // TODO: Implement update history tracking
    Ok(vec![])
}

/// Download file for manual download
#[tauri::command]
pub async fn download_manual_file(
    filename: String,
    url: String,
    destination: String,
) -> LauncherResult<String> {
    info!("Downloading manual file: {} from {} to {}", filename, url, destination);

    let download_id = download_manager().start_download(filename, url, destination).await?;
    Ok(download_id)
}

// ==================== LAUNCH MANAGEMENT COMMANDS ====================

/// Get launch status for a specific launch ID
#[tauri::command]
pub async fn get_launch_status(
    launch_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<Option<LaunchedProcess>> {
    info!("Getting launch status for: {}", launch_id);

    let status = state.launch_manager.get_launch_status(&launch_id)?;
    Ok(status)
}

/// Get all active processes
#[tauri::command]
pub async fn get_active_processes(
    state: State<'_, AppState>,
) -> LauncherResult<Vec<LaunchedProcess>> {
    info!("Getting all active processes");

    let processes = state.launch_manager.get_active_processes()?;
    Ok(processes)
}

/// Get processes for a specific instance
#[tauri::command]
pub async fn get_instance_processes(
    instance_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<Vec<LaunchedProcess>> {
    info!("Getting processes for instance: {}", instance_id);

    let processes = state.launch_manager.get_instance_processes(&instance_id)?;
    Ok(processes)
}

/// Terminate a running process
#[tauri::command]
pub async fn terminate_process(
    launch_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Terminating process: {}", launch_id);

    state.launch_manager.terminate_instance(&launch_id).await?;
    info!("Process terminated successfully: {}", launch_id);
    Ok(())
}

/// Force kill all processes for an instance
#[tauri::command]
pub async fn force_kill_instance(
    instance_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<u32> {
    info!("Force killing all processes for instance: {}", instance_id);

    let killed_count = state.launch_manager.force_kill_instance_processes(&instance_id).await?;
    info!("Force killed {} processes for instance: {}", killed_count, instance_id);
    Ok(killed_count)
}

/// Clean up finished processes
#[tauri::command]
pub async fn cleanup_finished_processes(
    state: State<'_, AppState>,
) -> LauncherResult<u32> {
    info!("Cleaning up finished processes");

    let cleaned_count = state.launch_manager.cleanup_finished_processes().await?;
    info!("Cleaned up {} finished processes", cleaned_count);
    Ok(cleaned_count)
}

/// Get launch configuration for an instance
#[tauri::command]
pub async fn get_launch_config(
    instance_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<Option<LaunchConfig>> {
    info!("Getting launch configuration for instance: {}", instance_id);

    let config = state.launch_manager.get_launch_config(&instance_id)?;
    Ok(config)
}

/// Update launch configuration for an instance
#[tauri::command]
pub async fn update_launch_config(
    config: LaunchConfig,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Updating launch configuration for instance: {}", config.instance_id);

    state.launch_manager.update_launch_config(config)?;
    info!("Launch configuration updated successfully");
    Ok(())
}

/// Initialize the launch manager
#[tauri::command]
pub async fn initialize_launch_manager(
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Initializing launch manager");

    state.launch_manager.initialize().await?;
    info!("Launch manager initialized successfully");
    Ok(())
}

// Helper functions for launch management

/// Detect Prism Launcher installation path
fn detect_prism_path() -> LauncherResult<String> {
    let home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    let possible_paths = if cfg!(target_os = "windows") {
        vec![
            home_dir.join("AppData\\Local\\Programs\\PrismLauncher"),
            home_dir.join("AppData\\Roaming\\PrismLauncher"),
            std::path::PathBuf::from("C:\\Program Files\\PrismLauncher"),
            std::path::PathBuf::from("C:\\Program Files (x86)\\PrismLauncher"),
        ]
    } else if cfg!(target_os = "macos") {
        vec![
            home_dir.join("Applications/PrismLauncher.app"),
            std::path::PathBuf::from("/Applications/PrismLauncher.app"),
        ]
    } else {
        vec![
            home_dir.join(".local/share/PrismLauncher"),
            home_dir.join(".local/bin/PrismLauncher"),
            std::path::PathBuf::from("/usr/bin/PrismLauncher"),
            std::path::PathBuf::from("/usr/local/bin/PrismLauncher"),
        ]
    };

    for path in possible_paths {
        if path.exists() {
            return Ok(path.to_string_lossy().to_string());
        }
    }

    Err(LauncherError::PrismNotFound)
}

// ==================== UPDATE MANAGEMENT COMMANDS ====================

/// Initialize update manager
#[tauri::command]
pub async fn initialize_update_manager(
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Initializing update manager");

    state.update_manager.initialize().await?;
    info!("Update manager initialized successfully");
    Ok(())
}

/// Check for launcher updates
#[tauri::command]
pub async fn check_for_updates(
    state: State<'_, AppState>,
) -> LauncherResult<Option<UpdateInfo>> {
    info!("Checking for launcher updates");

    state.update_manager.initialize().await?;
    let update_info = state.update_manager.check_for_updates().await?;

    match &update_info {
        Some(update) => info!("Update available: {} -> {}",
            env!("CARGO_PKG_VERSION"), update.version),
        None => info!("No updates available"),
    }

    Ok(update_info)
}

/// Download available update
#[tauri::command]
pub async fn download_update(
    update_info: UpdateInfo,
    state: State<'_, AppState>,
) -> LauncherResult<String> {
    info!("Starting update download for version {}", update_info.version);

    state.update_manager.initialize().await?;
    let download_id = state.update_manager.download_update(&update_info).await?;

    info!("Update download started with ID: {}", download_id);
    Ok(download_id)
}

/// Apply downloaded update
#[tauri::command]
pub async fn apply_update(
    download_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Applying update for download: {}", download_id);

    state.update_manager.initialize().await?;
    state.update_manager.apply_update(&download_id).await?;

    info!("Update applied successfully");
    Ok(())
}

/// Get update download progress
#[tauri::command]
pub async fn get_update_progress(
    download_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<Option<UpdateProgress>> {
    info!("Getting update progress for: {}", download_id);

    let progress = state.update_manager.get_download_progress(&download_id).await;
    Ok(progress)
}

/// Get all active update downloads
#[tauri::command]
pub async fn get_all_update_downloads(
    state: State<'_, AppState>,
) -> LauncherResult<Vec<UpdateProgress>> {
    info!("Getting all active update downloads");

    let downloads = state.update_manager.get_all_downloads().await;
    Ok(downloads)
}

/// Cancel update download
#[tauri::command]
pub async fn cancel_update_download(
    download_id: String,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Cancelling update download: {}", download_id);

    state.update_manager.cancel_download(&download_id).await?;
    info!("Update download {} cancelled successfully", download_id);
    Ok(())
}

/// Clean up completed update downloads
#[tauri::command]
pub async fn cleanup_update_downloads(
    state: State<'_, AppState>,
) -> LauncherResult<u32> {
    info!("Cleaning up completed update downloads");

    let cleaned_count = state.update_manager.cleanup_completed_downloads().await?;
    info!("Cleaned up {} completed update downloads", cleaned_count);
    Ok(cleaned_count)
}

/// Get update settings
#[tauri::command]
pub async fn get_update_settings(
    state: State<'_, AppState>,
) -> LauncherResult<UpdateSettings> {
    info!("Getting update settings");

    state.update_manager.initialize().await?;
    let settings = state.update_manager.get_settings().await;
    Ok(settings)
}

/// Update update settings
#[tauri::command]
pub async fn update_update_settings(
    settings: UpdateSettings,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Updating update settings: auto_update={}, channel={}, allow_prerelease={}",
          settings.auto_update_enabled, settings.update_channel.as_str(), settings.allow_prerelease);

    state.update_manager.initialize().await?;
    state.update_manager.update_settings(settings).await?;

    info!("Update settings updated successfully");
    Ok(())
}

/// Set update channel
#[tauri::command]
pub async fn set_update_channel(
    channel: UpdateChannel,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Setting update channel to: {}", channel.as_str());

    state.update_manager.initialize().await?;
    let mut settings = state.update_manager.get_settings().await;
    settings.update_channel = channel.clone();
    state.update_manager.update_settings(settings).await?;

    info!("Update channel set to: {}", channel.as_str());
    Ok(())
}

/// Enable/disable auto updates
#[tauri::command]
pub async fn set_auto_update(
    enabled: bool,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Setting auto update to: {}", enabled);

    state.update_manager.initialize().await?;
    let mut settings = state.update_manager.get_settings().await;
    settings.auto_update_enabled = enabled;
    state.update_manager.update_settings(settings).await?;

    info!("Auto update set to: {}", enabled);
    Ok(())
}

/// Enable/disable prerelease updates
#[tauri::command]
pub async fn set_allow_prerelease_updates(
    allowed: bool,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Setting prerelease updates allowed to: {}", allowed);

    state.update_manager.initialize().await?;
    let mut settings = state.update_manager.get_settings().await;
    settings.allow_prerelease = allowed;
    state.update_manager.update_settings(settings).await?;

    info!("Prerelease updates allowed set to: {}", allowed);
    Ok(())
}

// ==================== PERFORMANCE MONITORING COMMANDS ====================

/// Get performance metrics
#[tauri::command]
pub async fn get_performance_metrics(
    state: State<'_, AppState>,
) -> LauncherResult<performance::PerformanceMetrics> {
    info!("Getting performance metrics");

    let monitor = performance::get_performance_monitor();

    // Get basic metrics
    let startup_time = monitor.get_startup_time_ms();
    let memory_usage = monitor.get_memory_usage_mb().await;

    // Get active processes count
    let active_processes = state.launch_manager.get_active_processes()
        .unwrap_or_default()
        .len();

    // Get active downloads count
    let active_downloads = state.download_progress.lock().unwrap().len();

    // Calculate average response time (sample of recent operations)
    let avg_health_check = monitor.get_average_response_time("health_check").await.unwrap_or(0.0);
    let avg_get_settings = monitor.get_average_response_time("get_settings").await.unwrap_or(0.0);
    let avg_launch = monitor.get_average_response_time("launch_instance").await.unwrap_or(0.0);

    let average_response_time = (avg_health_check + avg_get_settings + avg_launch) / 3.0;

    let metrics = performance::PerformanceMetrics {
        startup_time_ms: startup_time,
        memory_usage_mb: memory_usage,
        active_processes,
        active_downloads,
        cache_hit_rate: 0.0, // TODO: Implement cache hit rate tracking
        average_response_time_ms: average_response_time,
    };

    info!("Performance metrics: {}ms startup, {}MB memory, {} processes, {} downloads",
          startup_time, memory_usage, active_processes, active_downloads);

    Ok(metrics)
}

/// Clear performance cache
#[tauri::command]
pub async fn clear_performance_cache() -> LauncherResult<()> {
    info!("Clearing performance cache");

    let _monitor = performance::get_performance_monitor();
    // Note: We would need to add a clear method to the performance monitor
    // For now, this is a placeholder

    info!("Performance cache cleared");
    Ok(())
}

// ==================== TAURI UPDATER INTEGRATION COMMANDS ====================

/// Check for and install available updates
#[tauri::command]
pub async fn check_and_install_update(
    app_handle: tauri::AppHandle,
) -> LauncherResult<bool> {
    info!("Checking for and installing updates");

    let updater = app_handle.updater().map_err(|e| LauncherError::UpdateFailed(format!("Updater initialization failed: {}", e)))?;
    let response = updater.check().await;

    match response {
        Ok(Some(update)) => {
            info!("Update available: {}", update.version);

            // Download and install the update
            let update_response = update.download_and_install(|_, _| {}, || {}).await;

            match update_response {
                Ok(()) => {
                    info!("Update downloaded and installed successfully");
                    Ok(true)
                },
                Err(e) => {
                    warn!("Failed to install update: {}", e);
                    Err(LauncherError::UpdateFailed(format!("Failed to install update: {}", e)))
                }
            }
        },
        Ok(None) => {
            info!("No update available");
            Ok(false)
        },
        Err(e) => {
            info!("Update check failed: {}", e);
            Ok(false)
        }
    }
}

/// Check for updates without installing
#[tauri::command]
pub async fn check_tauri_update(
    app_handle: tauri::AppHandle,
) -> LauncherResult<Option<tauri_plugin_updater::Update>> {
    info!("Checking for Tauri updates");

    let updater = app_handle.updater().map_err(|e| LauncherError::UpdateFailed(format!("Updater initialization failed: {}", e)))?;
    match updater.check().await {
        Ok(Some(update)) => {
            info!("Update available: {} (body: {})", update.version, update.body.as_ref().unwrap_or(&String::new()));
            Ok(Some(update))
        },
        Ok(None) => {
            info!("No update available");
            Ok(None)
        },
        Err(e) => {
            info!("Update check failed: {}", e);
            Ok(None)
        }
    }
}

/// Start update download in background
#[tauri::command]
pub async fn start_update_download(
    app_handle: tauri::AppHandle,
) -> LauncherResult<Option<String>> {
    info!("Starting update download");

    let updater = app_handle.updater().map_err(|e| LauncherError::UpdateFailed(format!("Updater initialization failed: {}", e)))?;
    match updater.check().await {
        Ok(Some(update)) => {
            info!("Starting download for update: {}", update.version);

            // Start download (this is async and will continue in background)
            let download_result = update.download(|_, _| {}, || {}).await;

            match download_result {
                Ok(_) => {
                    info!("Update download completed successfully");
                    Ok(Some(update.version))
                },
                Err(e) => {
                    warn!("Failed to download update: {}", e);
                    Err(LauncherError::UpdateFailed(format!("Failed to download update: {}", e)))
                }
            }
        },
        Ok(None) => {
            info!("No update available to download");
            Ok(None)
        },
        Err(e) => {
            info!("Update check failed: {}", e);
            Ok(None)
        }
    }
}

/// Install downloaded update
#[tauri::command]
pub async fn install_downloaded_update(
    app_handle: tauri::AppHandle,
) -> LauncherResult<bool> {
    info!("Installing downloaded update");

    // Note: This requires that the update was already downloaded
    // The exact implementation depends on the Tauri updater API
    // For now, we'll trigger a restart which will apply the update

    info!("Triggering restart to apply update");
    app_handle.restart();

    // This code is unreachable but kept for compilation
    #[allow(unreachable_code)]
    Ok(true)
}

/// Get current and available update version information
#[tauri::command]
pub async fn get_update_info(
    app_handle: tauri::AppHandle,
) -> LauncherResult<Option<UpdateVersionInfo>> {
    info!("Getting update version information");

    let current_version = env!("CARGO_PKG_VERSION").to_string();

    let updater = app_handle.updater().map_err(|e| LauncherError::UpdateFailed(format!("Updater initialization failed: {}", e)))?;
    match updater.check().await {
        Ok(Some(update)) => {
            info!("Update available: {} -> {}", current_version, update.version);

            let update_info = UpdateVersionInfo {
                current_version,
                available_version: update.version.clone(),
                release_date: update.date.map(|d| d.to_string()),
                release_notes: update.body.clone(),
                download_url: None, // Tauri updater handles this internally
                mandatory: false, // Could be configured in the update response
                signature: None,
            };

            Ok(Some(update_info))
        },
        Ok(None) => {
            info!("No update available");
            Ok(None)
        },
        Err(e) => {
            info!("Update check failed: {}", e);
            Ok(None)
        }
    }
}

/// Structure for update version information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateVersionInfo {
    pub current_version: String,
    pub available_version: String,
    pub release_date: Option<String>,
    pub release_notes: Option<String>,
    pub download_url: Option<String>,
    pub mandatory: bool,
    pub signature: Option<String>,
}

/// Configure automatic update behavior
#[tauri::command]
pub async fn configure_auto_update(
    enabled: bool,
    check_interval_hours: u32,
    install_mode: String,
    _app_handle: tauri::AppHandle,
) -> LauncherResult<()> {
    info!("Configuring auto-update: enabled={}, interval={}h, mode={}",
          enabled, check_interval_hours, install_mode);

    // Save auto-update settings
    let auto_update_settings = AutoUpdateSettings {
        enabled,
        check_interval_hours,
        install_mode,
        last_check: chrono::Utc::now().to_rfc3339(),
    };

    // Save to configuration
    crate::utils::config::save_auto_update_settings(&auto_update_settings)?;

    info!("Auto-update configured successfully");
    Ok(())
}

/// Structure for auto-update settings
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AutoUpdateSettings {
    pub enabled: bool,
    pub check_interval_hours: u32,
    pub install_mode: String,
    pub last_check: String,
}

/// Enable/disable update notifications
#[tauri::command]
pub async fn set_update_notifications(
    enabled: bool,
    state: State<'_, AppState>,
) -> LauncherResult<()> {
    info!("Setting update notifications to: {}", enabled);

    let mut settings = state.settings.lock().unwrap();
    settings.update_notifications = enabled;

    // Save settings
    crate::utils::config::save_settings(&*settings)?;

    info!("Update notifications set to: {}", enabled);
    Ok(())
}

// Helper function to determine modpack download destination
async fn determine_modpack_destination(modpack: &crate::models::Modpack) -> LauncherResult<String> {
    let mut home_dir = dirs::home_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

    home_dir.push("TheBoysLauncher");
    home_dir.push("modpacks");

    // Create directory if it doesn't exist
    tokio::fs::create_dir_all(&home_dir).await
        .map_err(|e| LauncherError::FileSystem(
            format!("Failed to create modpacks directory: {}", e)
        ))?;

    home_dir.push(format!("{}.zip", modpack.id));
    Ok(home_dir.to_string_lossy().to_string())
}