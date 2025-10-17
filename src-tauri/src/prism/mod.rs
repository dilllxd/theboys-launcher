use crate::models::{LauncherResult, LauncherError};
use serde::{Deserialize, Serialize};
use std::path::Path;
use tracing::{info, warn, error};
use reqwest::Client;

/// Prism Launcher installation information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrismInstallation {
    pub version: String,
    pub path: String,
    pub executable_path: String,
    pub is_managed: bool,
    pub installation_date: Option<String>,
    pub size_bytes: Option<u64>,
    pub architecture: String,
    pub platform: String,
}

/// Prism update information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrismUpdateInfo {
    pub current_version: Option<String>,
    pub latest_version: String,
    pub update_available: bool,
    pub download_url: String,
    pub release_notes: Option<String>,
    pub size_bytes: u64,
}

/// Prism Manager for handling Prism Launcher detection and installation
pub struct PrismManager {
    client: Client,
    current_installation: Option<PrismInstallation>,
}

impl PrismManager {
    /// Create a new Prism Manager instance
    pub fn new() -> Self {
        Self {
            client: Client::new(),
            current_installation: None,
        }
    }

    /// Initialize the Prism Manager by scanning for installations
    pub async fn initialize(&mut self) -> LauncherResult<()> {
        info!("Initializing Prism Manager");

        // Check for existing installations
        if let Some(installation) = self.detect_prism_installation().await? {
            self.current_installation = Some(installation);
            info!("Found existing Prism Launcher installation");
        } else {
            info!("No existing Prism Launcher installation found");
        }

        Ok(())
    }

    /// Detect existing Prism Launcher installation
    async fn detect_prism_installation(&self) -> LauncherResult<Option<PrismInstallation>> {
        // Check common installation paths
        let common_paths = self.get_common_prism_paths();

        for path in common_paths {
            if let Some(installation) = self.check_prism_at_path(&path).await? {
                return Ok(Some(installation));
            }
        }

        // Check if it's in PATH
        if let Ok(prism_exe) = which::which("PrismLauncher") {
            if let Some(installation) = self.check_prism_executable(&prism_exe.to_string_lossy()).await? {
                return Ok(Some(installation));
            }
        }

        Ok(None)
    }

    /// Get common Prism Launcher installation paths
    fn get_common_prism_paths(&self) -> Vec<String> {
        match std::env::consts::OS {
            "windows" => vec![
                r"C:\Program Files\PrismLauncher".to_string(),
                r"C:\Program Files (x86)\PrismLauncher".to_string(),
                r"C:\Users".to_string(), // Will search user directories
            ],
            "macos" => vec![
                "/Applications/PrismLauncher.app".to_string(),
                "/Applications/Prism Launcher.app".to_string(),
            ],
            "linux" => vec![
                "/usr/bin".to_string(),
                "/usr/local/bin".to_string(),
                "/opt/prism-launcher".to_string(),
            ],
            _ => Vec::new(),
        }
    }

    /// Check if Prism Launcher exists at a given path
    pub async fn check_prism_at_path(&self, base_path: &str) -> LauncherResult<Option<PrismInstallation>> {
        let path = Path::new(base_path);

        if !path.exists() {
            return Ok(None);
        }

        // Check different executable names based on platform
        let executable_names = self.get_prism_executable_names();

        for executable_name in executable_names {
            let executable_path = if cfg!(target_os = "windows") {
                path.join(executable_name)
            } else {
                path.join(executable_name)
            };

            if executable_path.exists() {
                return Ok(self.check_prism_executable(&executable_path.to_string_lossy()).await?);
            }
        }

        // If it's a directory, search within it
        if path.is_dir() {
            if let Ok(entries) = std::fs::read_dir(path) {
                for entry in entries.flatten() {
                    let entry_path = entry.path();
                    if entry_path.is_file() {
                        let file_name = entry_path.file_name()
                            .and_then(|n| n.to_str())
                            .unwrap_or("");

                        if file_name.to_lowercase().contains("prism") &&
                           (file_name.to_lowercase().contains("launcher") || file_name.to_lowercase().contains(".exe")) {
                            if let Some(installation) = self.check_prism_executable(&entry_path.to_string_lossy()).await? {
                                return Ok(Some(installation));
                            }
                        }
                    }
                }
            }
        }

        Ok(None)
    }

    /// Get Prism Launcher executable names for current platform
    fn get_prism_executable_names(&self) -> Vec<String> {
        match std::env::consts::OS {
            "windows" => vec![
                "PrismLauncher.exe".to_string(),
                "Prism Launcher.exe".to_string(),
                "prismlauncher.exe".to_string(),
            ],
            "macos" => vec![
                "PrismLauncher".to_string(),
                "Prism Launcher".to_string(),
                "prismlauncher".to_string(),
            ],
            "linux" => vec![
                "PrismLauncher".to_string(),
                "prismlauncher".to_string(),
                "prism-launcher".to_string(),
            ],
            _ => Vec::new(),
        }
    }

    /// Check a specific Prism executable and get installation info
    async fn check_prism_executable(&self, executable_path: &str) -> LauncherResult<Option<PrismInstallation>> {
        let path = Path::new(executable_path);

        if !path.exists() {
            return Ok(None);
        }

        // Try to get version information
        let version = self.get_prism_version(executable_path).await
            .unwrap_or_else(|_| "Unknown".to_string());

        let installation = PrismInstallation {
            version: version.clone(),
            path: path.parent()
                .and_then(|p| p.to_str())
                .unwrap_or(executable_path)
                .to_string(),
            executable_path: executable_path.to_string(),
            is_managed: self.is_managed_installation(path),
            installation_date: self.get_installation_date(path),
            size_bytes: self.get_installation_size(path),
            architecture: std::env::consts::ARCH.to_string(),
            platform: std::env::consts::OS.to_string(),
        };

        info!("Found Prism Launcher {} at {}", version, executable_path);
        Ok(Some(installation))
    }

    /// Get Prism Launcher version
    async fn get_prism_version(&self, executable_path: &str) -> LauncherResult<String> {
        let output = tokio::process::Command::new(executable_path)
            .arg("--version")
            .output()
            .await;

        match output {
            Ok(output) => {
                let version_info = String::from_utf8_lossy(&output.stdout);
                // Extract version from output
                if let Some(line) = version_info.lines().next() {
                    Ok(line.trim().to_string())
                } else {
                    Ok("Unknown".to_string())
                }
            }
            Err(_) => Ok("Unknown".to_string()),
        }
    }

    /// Check if Prism installation is managed by this launcher
    fn is_managed_installation(&self, path: &Path) -> bool {
        path.to_string_lossy().contains("TheBoysLauncher") &&
        path.to_string_lossy().contains("tools") &&
        path.to_string_lossy().contains("prism")
    }

    /// Get installation date for Prism
    fn get_installation_date(&self, path: &Path) -> Option<String> {
        if let Ok(metadata) = std::fs::metadata(path) {
            if let Ok(modified) = metadata.modified() {
                return Some(format!("{}", modified.elapsed().unwrap_or_default().as_secs()));
            }
        }
        None
    }

    /// Get installation size for Prism
    fn get_installation_size(&self, path: &Path) -> Option<u64> {
        if path.is_file() {
            if let Ok(metadata) = std::fs::metadata(path) {
                return Some(metadata.len());
            }
        } else if path.is_dir() {
            if let Ok(size) = get_directory_size(path) {
                return Some(size);
            }
        }
        None
    }

    /// Get current Prism installation
    pub fn get_current_installation(&self) -> Option<PrismInstallation> {
        self.current_installation.clone()
    }

    /// Check for Prism Launcher updates
    pub async fn check_for_updates(&self) -> LauncherResult<PrismUpdateInfo> {
        info!("Checking for Prism Launcher updates");

        let current_version = self.current_installation.as_ref().map(|i| i.version.clone());

        // Get latest release from GitHub API
        let api_url = "https://api.github.com/repos/PrismLauncher/PrismLauncher/releases/latest";

        let response = self.client.get(api_url)
            .header("User-Agent", "TheBoys-Launcher/1.1.0")
            .header("Accept", "application/vnd.github.v3+json")
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("Failed to fetch Prism release info: {}", response.status())
            ));
        }

        let release_info: serde_json::Value = response.json().await?;

        let latest_version = release_info.get("tag_name")
            .and_then(|v| v.as_str())
            .unwrap_or("latest")
            .trim_start_matches('v')
            .to_string();

        let release_notes = release_info.get("body")
            .and_then(|b| b.as_str())
            .map(|s| s.to_string());

        // Find the appropriate download URL for this platform
        let download_url = self.get_platform_download_url(&release_info)?;
        let size_bytes = self.get_download_size(&download_url).await.unwrap_or(0);

        let update_available = match (&current_version, &latest_version) {
            (Some(current), latest) => {
                // Simple version comparison - could be improved with semver
                current != latest && latest != "latest"
            }
            (None, _) => true, // No current installation, so update is available
        };

        Ok(PrismUpdateInfo {
            current_version,
            latest_version,
            update_available,
            download_url,
            release_notes,
            size_bytes,
        })
    }

    /// Get platform-specific download URL from release info
    fn get_platform_download_url(&self, release_info: &serde_json::Value) -> LauncherResult<String> {
        let platform = std::env::consts::OS;
        let arch = std::env::consts::ARCH;

        if let Some(assets) = release_info.get("assets").and_then(|a| a.as_array()) {
            for asset in assets {
                if let (Some(name), Some(download_url)) = (
                    asset.get("name").and_then(|n| n.as_str()),
                    asset.get("browser_download_url").and_then(|u| u.as_str())
                ) {
                    if self.is_platform_compatible(name, platform, arch) {
                        return Ok(download_url.to_string());
                    }
                }
            }
        }

        Err(LauncherError::DownloadFailed(
            "Could not find suitable Prism Launcher download for this platform".to_string()
        ))
    }

    /// Check if asset name is compatible with current platform
    fn is_platform_compatible(&self, name: &str, platform: &str, _arch: &str) -> bool {
        let name_lower = name.to_lowercase();

        // Check for portable builds (preferred)
        if !name_lower.contains("portable") {
            return false;
        }

        match platform {
            "windows" => {
                name_lower.contains("windows") &&
                (name_lower.contains("x64") || name_lower.contains("amd64")) &&
                (name_lower.contains("msvc") || name_lower.contains("mingw"))
            }
            "macos" => {
                name_lower.contains("macos") || name_lower.contains("darwin")
            }
            "linux" => {
                name_lower.contains("linux") &&
                (name_lower.contains("x64") || name_lower.contains("amd64"))
            }
            _ => false,
        }
    }

    /// Get download size for a URL
    async fn get_download_size(&self, url: &str) -> LauncherResult<u64> {
        let response = self.client.head(url)
            .header("User-Agent", "TheBoys-Launcher/1.1.0")
            .send()
            .await?;

        Ok(response.content_length().unwrap_or(0))
    }

    /// Get Prism installation path for the launcher
    pub fn get_prism_install_path(&self) -> LauncherResult<String> {
        let mut base_path = dirs::home_dir()
            .ok_or_else(|| LauncherError::FileSystem("Could not find home directory".to_string()))?;

        base_path.push("TheBoysLauncher");
        base_path.push("tools");
        base_path.push("prism");

        Ok(base_path.to_string_lossy().to_string())
    }

    /// Install Prism Launcher
    pub async fn install_prism(&mut self, version: Option<String>) -> LauncherResult<String> {
        info!("Installing Prism Launcher, version: {:?}", version);

        // Check for updates to get download info
        let update_info = self.check_for_updates().await?;

        if let Some(v) = version {
            if v != update_info.latest_version {
                return Err(LauncherError::InvalidConfig(
                    format!("Specific version installation not yet implemented. Latest available: {}", update_info.latest_version)
                ));
            }
        }

        let install_path = self.get_prism_install_path()?;

        // Create installation directory
        tokio::fs::create_dir_all(&install_path).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create Prism directory: {}", e)
            ))?;

        // Return download URL for the download manager to handle
        Ok(update_info.download_url)
    }

    /// Uninstall managed Prism installation
    pub async fn uninstall_prism(&mut self) -> LauncherResult<()> {
        info!("Uninstalling managed Prism Launcher installation");

        if let Some(installation) = &self.current_installation {
            if installation.is_managed {
                tokio::fs::remove_dir_all(&installation.path).await
                    .map_err(|e| LauncherError::FileSystem(
                        format!("Failed to remove Prism installation: {}", e)
                    ))?;

                self.current_installation = None;
                info!("Successfully uninstalled Prism Launcher");
            } else {
                return Err(LauncherError::PermissionDenied(
                    "Cannot uninstall externally managed Prism installation".to_string()
                ));
            }
        } else {
            return Err(LauncherError::InvalidConfig(
                "No Prism installation found".to_string()
            ));
        }

        Ok(())
    }

    /// Get Prism status information
    pub fn get_prism_status(&self) -> PrismStatus {
        let installation = self.get_current_installation();

        PrismStatus {
            is_installed: installation.is_some(),
            installation: installation.clone(),
            is_managed: installation.as_ref().map(|i| i.is_managed).unwrap_or(false),
            platform: std::env::consts::OS.to_string(),
            architecture: std::env::consts::ARCH.to_string(),
        }
    }
}

/// Prism status information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrismStatus {
    pub is_installed: bool,
    pub installation: Option<PrismInstallation>,
    pub is_managed: bool,
    pub platform: String,
    pub architecture: String,
}

impl Default for PrismManager {
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

/// Global Prism Manager instance
static mut PRISM_MANAGER: Option<PrismManager> = None;
static PRISM_MANAGER_INIT: std::sync::Once = std::sync::Once::new();

/// Get the global Prism Manager instance
pub fn prism_manager() -> &'static mut PrismManager {
    unsafe {
        PRISM_MANAGER_INIT.call_once(|| {
            PRISM_MANAGER = Some(PrismManager::new());
        });
        PRISM_MANAGER.as_mut().unwrap()
    }
}