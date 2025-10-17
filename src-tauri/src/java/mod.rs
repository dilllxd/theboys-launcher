use crate::models::{JavaVersion, LauncherResult, LauncherError};
use crate::utils::system;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;
use std::process::Command;
use tracing::{info, warn, error};
use reqwest::Client;

/// Java installation status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JavaInstallation {
    pub version: String,
    pub path: String,
    pub is_64bit: bool,
    pub major_version: u32,
    pub is_managed: bool, // Whether installed by this launcher
    pub installation_date: Option<String>,
    pub size_bytes: Option<u64>,
}

/// Java Manager for handling Java detection and installation
pub struct JavaManager {
    pub installed_versions: HashMap<String, JavaInstallation>,
    pub required_versions: HashMap<String, String>, // MC version -> Java version
    client: Client,
}

impl JavaManager {
    /// Create a new Java Manager instance
    pub fn new() -> Self {
        let mut required_versions = HashMap::new();

        // Map Minecraft versions to required Java versions
        required_versions.insert("1.0".to_string(), "8".to_string());
        required_versions.insert("1.1".to_string(), "8".to_string());
        required_versions.insert("1.2".to_string(), "8".to_string());
        required_versions.insert("1.3".to_string(), "8".to_string());
        required_versions.insert("1.4".to_string(), "8".to_string());
        required_versions.insert("1.5".to_string(), "8".to_string());
        required_versions.insert("1.6".to_string(), "8".to_string());
        required_versions.insert("1.7".to_string(), "8".to_string());
        required_versions.insert("1.8".to_string(), "8".to_string());
        required_versions.insert("1.9".to_string(), "8".to_string());
        required_versions.insert("1.10".to_string(), "8".to_string());
        required_versions.insert("1.11".to_string(), "8".to_string());
        required_versions.insert("1.12".to_string(), "8".to_string());
        required_versions.insert("1.13".to_string(), "16".to_string());
        required_versions.insert("1.14".to_string(), "16".to_string());
        required_versions.insert("1.15".to_string(), "16".to_string());
        required_versions.insert("1.16".to_string(), "16".to_string());
        required_versions.insert("1.17".to_string(), "17".to_string());
        required_versions.insert("1.18".to_string(), "18".to_string());
        required_versions.insert("1.19".to_string(), "18".to_string());
        required_versions.insert("1.20".to_string(), "18".to_string());
        required_versions.insert("1.21".to_string(), "21".to_string());

        Self {
            installed_versions: HashMap::new(),
            required_versions,
            client: Client::new(),
        }
    }

    /// Initialize the Java Manager by scanning for installations
    pub async fn initialize(&mut self) -> LauncherResult<()> {
        info!("Initializing Java Manager");

        let java_versions = system::detect_java_installations()?;

        self.installed_versions.clear();
        for version in java_versions {
            let installation = JavaInstallation {
                version: version.version.clone(),
                path: version.path.clone(),
                is_64bit: version.is_64bit,
                major_version: version.major_version,
                is_managed: self.is_managed_installation(&version.path),
                installation_date: self.get_installation_date(&version.path),
                size_bytes: self.get_installation_size(&version.path),
            };

            self.installed_versions.insert(version.version.clone(), installation);
        }

        info!("Found {} Java installations", self.installed_versions.len());
        Ok(())
    }

    /// Get all installed Java versions
    pub fn get_installed_versions(&self) -> Vec<JavaInstallation> {
        self.installed_versions.values().cloned().collect()
    }

    /// Get required Java version for a Minecraft version
    pub fn get_required_java_version(&self, minecraft_version: &str) -> Option<String> {
        // Check exact version first
        if let Some(version) = self.required_versions.get(minecraft_version) {
            return Some(version.clone());
        }

        // Check for prefix matches (e.g., "1.20.1" should match "1.20")
        for (mc_version, java_version) in &self.required_versions {
            if minecraft_version.starts_with(mc_version) {
                return Some(java_version.clone());
            }
        }

        // Default to Java 21 for unknown versions
        Some("21".to_string())
    }

    /// Check if a specific Java version is installed
    pub fn is_java_version_installed(&self, java_version: &str) -> bool {
        self.installed_versions.values().any(|inst| {
            inst.version == java_version || inst.major_version.to_string() == java_version
        })
    }

    /// Get the best Java installation for a Minecraft version
    pub fn get_best_java_installation(&self, minecraft_version: &str) -> Option<JavaInstallation> {
        let required_version = self.get_required_java_version(minecraft_version)?;

        // First try to find exact version match
        for installation in self.installed_versions.values() {
            if installation.version == required_version || installation.major_version.to_string() == required_version {
                return Some(installation.clone());
            }
        }

        // Fallback to any compatible version (higher versions are usually compatible)
        let required_major = required_version.parse::<u32>().unwrap_or(8);
        for installation in self.installed_versions.values() {
            if installation.major_version >= required_major {
                return Some(installation.clone());
            }
        }

        None
    }

    /// Get download URL for a specific Java version from Adoptium
    pub async fn get_java_download_url(&self, java_version: &str) -> LauncherResult<String> {
        let platform = self.determine_platform();

        let api_url = format!(
            "https://api.adoptium.net/v3/assets/latest/{}/hotspot?os={}&architecture=x64&image_type=jre&release=latest",
            java_version, platform
        );

        let response = self.client.get(&api_url)
            .header("User-Agent", "TheBoys-Launcher/1.1.0")
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("Failed to fetch Java download info: {}", response.status())
            ));
        }

        let assets: Vec<serde_json::Value> = response.json().await?;

        if let Some(asset) = assets.first() {
            if let Some(download_url) = asset.get("binary").and_then(|b| b.get("package")).and_then(|p| p.get("link")).and_then(|l| l.as_str()) {
                return Ok(download_url.to_string());
            }
        }

        Err(LauncherError::DownloadFailed(
            "Could not find suitable Java download".to_string()
        ))
    }

    /// Determine the current platform string for Adoptium API
    fn determine_platform(&self) -> &'static str {
        #[cfg(target_os = "windows")]
        return "windows";

        #[cfg(target_os = "macos")]
        return "mac";

        #[cfg(target_os = "linux")]
        return "linux";

        #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
        return "unknown";
    }

    /// Get Java installation path for the launcher
    pub fn get_java_install_path(&self, java_version: &str) -> LauncherResult<String> {
        let mut base_path = dirs::home_dir()
            .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

        base_path.push("TheBoysLauncher");
        base_path.push("tools");
        base_path.push("java");
        base_path.push(format!("jre-{}", java_version));

        Ok(base_path.to_string_lossy().to_string())
    }

    /// Check if Java installation is managed by this launcher
    fn is_managed_installation(&self, path: &str) -> bool {
        path.contains("TheBoysLauncher") && path.contains("tools") && path.contains("java")
    }

    /// Get installation date for a Java installation
    fn get_installation_date(&self, path: &str) -> Option<String> {
        if let Ok(metadata) = std::fs::metadata(path) {
            if let Ok(modified) = metadata.modified() {
                return Some(format!("{}", modified.elapsed().unwrap_or_default().as_secs()));
            }
        }
        None
    }

    /// Get installation size for a Java installation
    fn get_installation_size(&self, path: &str) -> Option<u64> {
        if let Ok(size) = get_directory_size(Path::new(path)) {
            Some(size)
        } else {
            None
        }
    }

    /// Get all managed Java installations
    pub fn get_managed_installations(&self) -> Vec<JavaInstallation> {
        self.installed_versions.values()
            .filter(|inst| inst.is_managed)
            .cloned()
            .collect()
    }

    /// Remove a managed Java installation
    pub async fn remove_managed_installation(&mut self, java_version: &str) -> LauncherResult<()> {
        let install_path = self.get_java_install_path(java_version)?;

        if Path::new(&install_path).exists() {
            tokio::fs::remove_dir_all(&install_path).await
                .map_err(|e| LauncherError::FileSystem(
                    format!("Failed to remove Java installation: {}", e)
                ))?;

            // Remove from our tracked versions
            self.installed_versions.retain(|_, v| v.path != install_path);

            info!("Removed managed Java installation: {}", java_version);
        }

        Ok(())
    }

    /// Get Java compatibility information
    pub fn get_compatibility_info(&self, minecraft_version: &str) -> JavaCompatibilityInfo {
        let required_version = self.get_required_java_version(minecraft_version);
        let installed_compatible = self.get_best_java_installation(minecraft_version);

        JavaCompatibilityInfo {
            minecraft_version: minecraft_version.to_string(),
            required_java_version: required_version,
            has_compatible_java: installed_compatible.is_some(),
            recommended_installation: installed_compatible,
        }
    }

    /// Cleanup old managed Java installations
    pub async fn cleanup_old_installations(&mut self) -> LauncherResult<usize> {
        let managed = self.get_managed_installations();
        let mut removed_count = 0;

        // Keep only the latest 3 versions to save space
        let mut sorted_versions = managed;
        sorted_versions.sort_by(|a, b| b.major_version.cmp(&a.major_version));

        for (index, installation) in sorted_versions.iter().enumerate() {
            if index >= 3 { // Keep only top 3
                if let Err(e) = self.remove_managed_installation(&installation.version).await {
                    warn!("Failed to remove old Java installation {}: {}", installation.version, e);
                } else {
                    removed_count += 1;
                }
            }
        }

        if removed_count > 0 {
            info!("Cleaned up {} old Java installations", removed_count);
        }

        Ok(removed_count)
    }
}

/// Java compatibility information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JavaCompatibilityInfo {
    pub minecraft_version: String,
    pub required_java_version: Option<String>,
    pub has_compatible_java: bool,
    pub recommended_installation: Option<JavaInstallation>,
}

impl Default for JavaManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Calculate directory size recursively
fn get_directory_size(path: &Path) -> std::io::Result<u64> {
    let mut total_size = 0u64;

    if path.is_dir() {
        for entry in std::fs::read_dir(path)? {
            let entry = entry?;
            let entry_path = entry.path();

            if entry_path.is_dir() {
                total_size += get_directory_size(&entry_path)?;
            } else {
                total_size += entry.metadata()?.len();
            }
        }
    } else {
        total_size = std::fs::metadata(path)?.len();
    }

    Ok(total_size)
}

/// Global Java Manager instance
static mut JAVA_MANAGER: Option<JavaManager> = None;
static JAVA_MANAGER_INIT: std::sync::Once = std::sync::Once::new();

/// Get the global Java Manager instance
pub fn java_manager() -> &'static mut JavaManager {
    unsafe {
        JAVA_MANAGER_INIT.call_once(|| {
            JAVA_MANAGER = Some(JavaManager::new());
        });
        JAVA_MANAGER.as_mut().unwrap()
    }
}