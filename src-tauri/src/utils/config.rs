use crate::models::{LauncherSettings, LauncherResult, LauncherError};
use std::path::PathBuf;
use dirs;
use tracing::{info, warn};

/// Get the launcher configuration directory
pub fn get_config_dir() -> LauncherResult<PathBuf> {
    let config_dir = dirs::config_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find config directory".to_string()))?
        .join("theboys-launcher");

    std::fs::create_dir_all(&config_dir)?;
    Ok(config_dir)
}

/// Get the settings file path
pub fn get_settings_file() -> LauncherResult<PathBuf> {
    Ok(get_config_dir()?.join("settings.json"))
}

/// Load settings from file with system-aware defaults
pub fn load_settings() -> LauncherResult<LauncherSettings> {
    let settings_file = get_settings_file()?;

    if !settings_file.exists() {
        info!("Settings file not found, creating with system-aware defaults");
        let default_settings = create_system_aware_defaults()?;
        save_settings(&default_settings)?;
        return Ok(default_settings);
    }

    info!("Loading settings from {:?}", settings_file);
    let content = std::fs::read_to_string(&settings_file)
        .map_err(|e| LauncherError::FileSystem(format!("Failed to read settings file: {}", e)))?;

    let mut settings: LauncherSettings = serde_json::from_str(&content)
        .map_err(|e| LauncherError::Serialization(format!("Failed to parse settings: {}", e)))?;

    // Validate and sanitize settings
    sanitize_settings(&mut settings)?;

    Ok(settings)
}

/// Save settings to file with atomic write to prevent corruption
pub fn save_settings(settings: &LauncherSettings) -> LauncherResult<()> {
    let settings_file = get_settings_file()?;

    info!("Saving settings to {:?}", settings_file);

    // Create backup of existing settings if it exists
    if settings_file.exists() {
        let backup_path = settings_file.with_extension("json.bak");
        if let Err(e) = std::fs::copy(&settings_file, &backup_path) {
            warn!("Failed to create settings backup: {}", e);
        }
    }

    let content = serde_json::to_string_pretty(settings)
        .map_err(|e| LauncherError::Serialization(format!("Failed to serialize settings: {}", e)))?;

    // Atomic write: write to temporary file first, then rename
    let temp_file = settings_file.with_extension("json.tmp");
    std::fs::write(&temp_file, content)
        .map_err(|e| LauncherError::FileSystem(format!("Failed to write temporary settings file: {}", e)))?;

    std::fs::rename(&temp_file, &settings_file)
        .map_err(|e| LauncherError::FileSystem(format!("Failed to rename settings file: {}", e)))?;

    info!("Settings saved successfully");
    Ok(())
}

/// Get the instances directory
pub fn get_instances_dir() -> LauncherResult<PathBuf> {
    let instances_dir = get_config_dir()?.join("instances");
    std::fs::create_dir_all(&instances_dir)?;
    Ok(instances_dir)
}

/// Get the downloads directory
pub fn get_downloads_dir() -> LauncherResult<PathBuf> {
    let downloads_dir = get_config_dir()?.join("downloads");
    std::fs::create_dir_all(&downloads_dir)?;
    Ok(downloads_dir)
}

/// Get the logs directory
pub fn get_logs_dir() -> LauncherResult<PathBuf> {
    let logs_dir = get_config_dir()?.join("logs");
    std::fs::create_dir_all(&logs_dir)?;
    Ok(logs_dir)
}

/// Get the cache directory
pub fn get_cache_dir() -> LauncherResult<PathBuf> {
    let cache_dir = dirs::cache_dir()
        .ok_or_else(|| LauncherError::FileSystem("Could not find cache directory".to_string()))?
        .join("theboys-launcher");

    std::fs::create_dir_all(&cache_dir)?;
    Ok(cache_dir)
}

/// Create system-aware default settings
fn create_system_aware_defaults() -> LauncherResult<LauncherSettings> {
    use crate::utils::system;

    // Get system information to calculate optimal defaults
    let sys_info = system::get_system_info().unwrap_or_else(|_| {
        warn!("Failed to get system info, using conservative defaults");
        crate::models::SystemInfo {
            os: std::env::consts::OS.to_string(),
            arch: std::env::consts::ARCH.to_string(),
            total_memory_mb: 8192,
            available_memory_mb: 4096,
            cpu_cores: 4,
            java_installed: false,
            java_versions: vec![],
        }
    });

    // Calculate optimal memory allocation (50% of total RAM, 2-16GB range)
    let total_ram_gb = sys_info.total_memory_mb / 1024;
    let recommended_memory_gb = (total_ram_gb / 2).max(2).min(16);
    let recommended_memory_mb = recommended_memory_gb * 1024;

    info!("Using system-aware defaults: {}GB RAM detected, recommending {}GB for Minecraft",
          total_ram_gb, recommended_memory_gb);

    // Detect default theme based on system (simplified)
    let default_theme = "system"; // Could be enhanced to detect system theme

    // Try to detect existing Prism Launcher installation
    let default_prism_path = detect_prism_launcher();

    // Try to detect Java installation
    let default_java_path = detect_default_java(&sys_info);

    Ok(LauncherSettings {
        memory_mb: recommended_memory_mb as u32,
        java_path: default_java_path,
        prism_path: default_prism_path,
        instances_dir: None, // Use default instances directory
        auto_update: true,
        theme: default_theme.to_string(),
        default_modpack_id: None,
        update_notifications: true,
    })
}

/// Sanitize and validate settings values
fn sanitize_settings(settings: &mut LauncherSettings) -> LauncherResult<()> {
    // Validate memory allocation (2GB - 32GB range)
    if settings.memory_mb < 2048 {
        warn!("Memory allocation too low ({}MB), setting to minimum 2048MB", settings.memory_mb);
        settings.memory_mb = 2048;
    } else if settings.memory_mb > 32768 {
        warn!("Memory allocation too high ({}MB), setting to maximum 32768MB", settings.memory_mb);
        settings.memory_mb = 32768;
    }

    // Validate theme
    match settings.theme.as_str() {
        "light" | "dark" | "system" => {},
        _ => {
            warn!("Invalid theme '{}', setting to 'system'", settings.theme);
            settings.theme = "system".to_string();
        }
    }

    // Validate Java path if provided
    if let Some(ref java_path) = settings.java_path {
        if !std::path::Path::new(java_path).exists() {
            warn!("Java path does not exist: {}, clearing", java_path);
            settings.java_path = None;
        }
    }

    // Validate Prism path if provided
    if let Some(ref prism_path) = settings.prism_path {
        if !std::path::Path::new(prism_path).exists() {
            warn!("Prism Launcher path does not exist: {}, clearing", prism_path);
            settings.prism_path = None;
        }
    }

    // Validate instances directory if provided
    if let Some(ref instances_dir) = settings.instances_dir {
        let path = std::path::Path::new(instances_dir);
        if !path.exists() {
            if let Err(e) = std::fs::create_dir_all(path) {
                warn!("Failed to create instances directory {}: {}, clearing", instances_dir, e);
                settings.instances_dir = None;
            }
        }
    }

    Ok(())
}

/// Detect existing Prism Launcher installation
fn detect_prism_launcher() -> Option<String> {
    let common_paths = get_common_prism_paths();

    for path in common_paths {
        if std::path::Path::new(&path).exists() {
            info!("Found Prism Launcher at: {}", path);
            return Some(path);
        }
    }

    info!("No existing Prism Launcher installation found");
    None
}

/// Get common Prism Launcher installation paths
fn get_common_prism_paths() -> Vec<String> {
    match std::env::consts::OS {
        "windows" => vec![
            r"C:\Program Files\Prism Launcher\PrismLauncher.exe".to_string(),
            r"C:\Program Files (x86)\Prism Launcher\PrismLauncher.exe".to_string(),
            // Expand environment variables for user-specific paths
            std::env::var("LOCALAPPDATA")
                .unwrap_or_else(|_| r"C:\Users\Default\AppData\Local".to_string())
                + r"\Programs\Prism Launcher\PrismLauncher.exe",
        ],
        "macos" => vec![
            "/Applications/Prism Launcher.app/Contents/MacOS/Prism Launcher".to_string(),
            "/Applications/PrismLauncher.app/Contents/MacOS/PrismLauncher".to_string(),
        ],
        "linux" => vec![
            "/usr/bin/prismlauncher".to_string(),
            "/usr/local/bin/prismlauncher".to_string(),
            "/opt/prismlauncher/bin/prismlauncher".to_string(),
            "~/.local/bin/prismlauncher".to_string(),
        ],
        _ => Vec::new(),
    }
}

/// Detect default Java installation
fn detect_default_java(sys_info: &crate::models::SystemInfo) -> Option<String> {
    // Prefer the latest 64-bit Java version
    for java_version in &sys_info.java_versions {
        if java_version.is_64bit && java_version.major_version >= 17 {
            info!("Found suitable Java {} at: {}", java_version.version, java_version.path);
            return Some(java_version.path.clone());
        }
    }

    // Fallback to any 64-bit Java
    for java_version in &sys_info.java_versions {
        if java_version.is_64bit {
            info!("Using fallback Java {} at: {}", java_version.version, java_version.path);
            return Some(java_version.path.clone());
        }
    }

    info!("No suitable Java installation found");
    None
}

/// Reset settings to system-aware defaults
pub fn reset_settings() -> LauncherResult<LauncherSettings> {
    info!("Resetting settings to system-aware defaults");
    let default_settings = create_system_aware_defaults()?;
    save_settings(&default_settings)?;
    Ok(default_settings)
}

/// Validate settings without saving
pub fn validate_settings(settings: &LauncherSettings) -> LauncherResult<()> {
    let mut temp_settings = settings.clone();
    sanitize_settings(&mut temp_settings)?;
    Ok(())
}

/// Save auto-update settings to file
pub fn save_auto_update_settings(settings: &crate::commands::AutoUpdateSettings) -> LauncherResult<()> {
    let settings_file = get_config_dir()?.join("auto_update.json");

    info!("Saving auto-update settings to {:?}", settings_file);

    let content = serde_json::to_string_pretty(settings)
        .map_err(|e| LauncherError::Serialization(format!("Failed to serialize auto-update settings: {}", e)))?;

    std::fs::write(&settings_file, content)
        .map_err(|e| LauncherError::FileSystem(format!("Failed to write auto-update settings file: {}", e)))?;

    info!("Auto-update settings saved successfully");
    Ok(())
}