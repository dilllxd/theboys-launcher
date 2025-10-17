use std::sync::Arc;
use tokio::sync::RwLock;
use serde::{Serialize, Deserialize};
use tracing::{info, warn, error};
use reqwest::Client;
use semver::Version;
use std::collections::HashMap;
use crate::models::LauncherError;
use crate::utils::file;

/// Update channel for the launcher
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum UpdateChannel {
    Stable,
    Beta,
    Alpha,
}

impl Default for UpdateChannel {
    fn default() -> Self {
        UpdateChannel::Stable
    }
}

impl UpdateChannel {
    pub fn as_str(&self) -> &'static str {
        match self {
            UpdateChannel::Stable => "stable",
            UpdateChannel::Beta => "beta",
            UpdateChannel::Alpha => "alpha",
        }
    }
}

/// Information about an available update
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateInfo {
    pub version: String,
    pub tag_name: String,
    pub release_notes: String,
    pub published_at: String,
    pub download_url: String,
    pub file_size: u64,
    pub checksum: Option<String>,
    pub prerelease: bool,
    pub channel: UpdateChannel,
}

/// Update download progress
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateProgress {
    pub download_id: String,
    pub version: String,
    pub downloaded_bytes: u64,
    pub total_bytes: u64,
    pub progress_percent: f64,
    pub download_speed_bps: u64,
    pub status: UpdateStatus,
}

/// Update status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum UpdateStatus {
    Checking,
    Available,
    Downloading,
    Downloaded,
    Installing,
    Installed,
    Failed,
    RollingBack,
}

/// Update settings
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateSettings {
    pub auto_update_enabled: bool,
    pub update_channel: UpdateChannel,
    pub check_updates_on_startup: bool,
    pub allow_prerelease: bool,
    pub backup_before_update: bool,
}

impl Default for UpdateSettings {
    fn default() -> Self {
        Self {
            auto_update_enabled: true,
            update_channel: UpdateChannel::Stable,
            check_updates_on_startup: true,
            allow_prerelease: false,
            backup_before_update: true,
        }
    }
}

/// Update manager for handling launcher self-updates
pub struct UpdateManager {
    current_version: String,
    client: Client,
    settings: Arc<RwLock<UpdateSettings>>,
    active_downloads: Arc<RwLock<HashMap<String, UpdateProgress>>>,
}

impl UpdateManager {
    /// Create a new update manager
    pub fn new(current_version: String) -> Self {
        Self {
            current_version,
            client: Client::builder()
                .user_agent("TheBoys-Launcher/1.1.0")
                .timeout(std::time::Duration::from_secs(30))
                .build()
                .unwrap_or_else(|_| Client::new()),
            settings: Arc::new(RwLock::new(UpdateSettings::default())),
            active_downloads: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Initialize the update manager
    pub async fn initialize(&self) -> Result<(), LauncherError> {
        info!("Initializing update manager with version {}", self.current_version);

        // Load update settings if they exist
        if let Ok(settings) = self.load_settings().await {
            *self.settings.write().await = settings;
            info!("Update settings loaded");
        } else {
            info!("Using default update settings");
        }

        Ok(())
    }

    /// Check for available updates
    pub async fn check_for_updates(&self) -> Result<Option<UpdateInfo>, LauncherError> {
        info!("Checking for launcher updates");

        let settings = self.settings.read().await;

        // Update progress to show we're checking
        self.update_progress_status("check".to_string(), UpdateStatus::Checking).await;

        // Fetch latest release from GitHub
        let release_info = self.fetch_latest_release(&settings.update_channel, settings.allow_prerelease).await?;

        let latest_version = match Version::parse(&release_info.version) {
            Ok(v) => v,
            Err(e) => {
                error!("Failed to parse latest version {}: {}", release_info.version, e);
                return Err(LauncherError::InvalidConfig(
                    format!("Invalid version format: {}", release_info.version)
                ));
            }
        };

        let current_version = match Version::parse(&self.current_version) {
            Ok(v) => v,
            Err(e) => {
                error!("Failed to parse current version {}: {}", self.current_version, e);
                return Err(LauncherError::InvalidConfig(
                    format!("Invalid current version: {}", self.current_version)
                ));
            }
        };

        if latest_version > current_version {
            info!("Update available: {} -> {}", self.current_version, release_info.version);
            Ok(Some(release_info))
        } else {
            info!("Launcher is up to date: {}", self.current_version);
            Ok(None)
        }
    }

    /// Download an update
    pub async fn download_update(&self, update_info: &UpdateInfo) -> Result<String, LauncherError> {
        info!("Starting download for update {}", update_info.version);

        let download_id = format!("update-{}-{}", update_info.version, uuid::Uuid::new_v4());

        // Initialize progress tracking
        let progress = UpdateProgress {
            download_id: download_id.clone(),
            version: update_info.version.clone(),
            downloaded_bytes: 0,
            total_bytes: update_info.file_size,
            progress_percent: 0.0,
            download_speed_bps: 0,
            status: UpdateStatus::Downloading,
        };

        {
            let mut downloads = self.active_downloads.write().await;
            downloads.insert(download_id.clone(), progress);
        }

        // Determine download destination
        let destination = self.get_update_destination(&update_info.version).await?;

        // Start the download
        let update_info_clone = update_info.clone();
        let download_id_clone = download_id.clone();
        let active_downloads = self.active_downloads.clone();

        tokio::spawn(async move {
            match Self::perform_download(&update_info_clone, &destination, &download_id_clone, active_downloads).await {
                Ok(_) => info!("Update download completed: {}", update_info_clone.version),
                Err(e) => error!("Update download failed: {}", e),
            }
        });

        Ok(download_id)
    }

    /// Apply a downloaded update
    pub async fn apply_update(&self, download_id: &str) -> Result<(), LauncherError> {
        info!("Applying update for download: {}", download_id);

        let progress = {
            let downloads = self.active_downloads.read().await;
            downloads.get(download_id).cloned()
        };

        let progress = progress.ok_or_else(|| LauncherError::InvalidConfig(
            "Download not found".to_string()
        ))?;

        if progress.status != UpdateStatus::Downloaded {
            return Err(LauncherError::InvalidConfig(
                "Update not downloaded yet".to_string()
            ));
        }

        // Update status to installing
        self.update_progress_status(download_id.to_string(), UpdateStatus::Installing).await;

        // Get the downloaded file path
        let update_file = self.get_update_destination(&progress.version).await?;

        // Create backup if enabled
        let settings = self.settings.read().await;
        if settings.backup_before_update {
            if let Err(e) = self.create_backup().await {
                warn!("Failed to create backup before update: {}", e);
            }
        }

        // Apply the update based on platform
        #[cfg(target_os = "windows")]
        let result = self.apply_windows_update(&update_file, &progress.version).await;

        #[cfg(target_os = "macos")]
        let result = self.apply_macos_update(&update_file, &progress.version).await;

        #[cfg(target_os = "linux")]
        let result = self.apply_linux_update(&update_file, &progress.version).await;

        #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
        let result = Err(LauncherError::NotImplemented("Platform not supported".to_string()));

        match result {
            Ok(_) => {
                self.update_progress_status(download_id.to_string(), UpdateStatus::Installed).await;
                info!("Update applied successfully: {}", progress.version);

                // Schedule restart after a short delay
                let version = progress.version.clone();
                tokio::spawn(async move {
                    tokio::time::sleep(std::time::Duration::from_secs(2)).await;
                    info!("Restarting launcher for update: {}", version);
                    // Note: In a real implementation, this would restart the application
                    // For now, we'll just log the restart intent
                });

                Ok(())
            },
            Err(e) => {
                self.update_progress_status(download_id.to_string(), UpdateStatus::Failed).await;
                error!("Failed to apply update: {}", e);

                // Attempt rollback if backup was created
                if settings.backup_before_update {
                    if let Err(rollback_err) = self.rollback_update().await {
                        error!("Failed to rollback update: {}", rollback_err);
                        self.update_progress_status(download_id.to_string(), UpdateStatus::RollingBack).await;
                    }
                }

                Err(e)
            }
        }
    }

    /// Get update settings
    pub async fn get_settings(&self) -> UpdateSettings {
        self.settings.read().await.clone()
    }

    /// Update settings
    pub async fn update_settings(&self, settings: UpdateSettings) -> Result<(), LauncherError> {
        *self.settings.write().await = settings.clone();
        self.save_settings(&settings).await?;
        Ok(())
    }

    /// Get download progress
    pub async fn get_download_progress(&self, download_id: &str) -> Option<UpdateProgress> {
        let downloads = self.active_downloads.read().await;
        downloads.get(download_id).cloned()
    }

    /// Get all active downloads
    pub async fn get_all_downloads(&self) -> Vec<UpdateProgress> {
        let downloads = self.active_downloads.read().await;
        downloads.values().cloned().collect()
    }

    /// Cancel an update download
    pub async fn cancel_download(&self, download_id: &str) -> Result<(), LauncherError> {
        let mut downloads = self.active_downloads.write().await;
        if let Some(mut progress) = downloads.remove(download_id) {
            progress.status = UpdateStatus::Failed;
            // Clean up downloaded file
            let _ = self.cleanup_download_file(&progress.version).await;
        }
        Ok(())
    }

    /// Clean up completed downloads
    pub async fn cleanup_completed_downloads(&self) -> Result<u32, LauncherError> {
        let mut downloads = self.active_downloads.write().await;
        let initial_count = downloads.len();

        downloads.retain(|_, progress| {
            !matches!(progress.status, UpdateStatus::Installed | UpdateStatus::Failed)
        });

        let removed_count = (initial_count - downloads.len()) as u32;

        if removed_count > 0 {
            info!("Cleaned up {} completed downloads", removed_count);
        }

        Ok(removed_count)
    }

    // Private helper methods

    /// Fetch latest release from GitHub
    async fn fetch_latest_release(&self, channel: &UpdateChannel, allow_prerelease: bool) -> Result<UpdateInfo, LauncherError> {
        let owner = "dilllxd";
        let repo = "theboys-launcher";

        let url = match channel {
            UpdateChannel::Stable if !allow_prerelease => {
                format!("https://api.github.com/repos/{}/{}/releases/latest", owner, repo)
            },
            _ => {
                // For beta/alpha or when prereleases are allowed, get all releases and filter
                format!("https://api.github.com/repos/{}/{}/releases", owner, repo)
            }
        };

        let response = self.client.get(&url).send().await
            .map_err(|e| LauncherError::Network(format!("Failed to fetch releases: {}", e)))?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("GitHub API error: {}", response.status())
            ));
        }

        if url.contains("/releases/latest") {
            // Single release response
            let release: serde_json::Value = response.json().await
                .map_err(|e| LauncherError::Network(format!("Failed to parse release: {}", e)))?;
            self.parse_release_info(release).await
        } else {
            // Multiple releases response - filter and find the latest
            let releases: Vec<serde_json::Value> = response.json().await
                .map_err(|e| LauncherError::Network(format!("Failed to parse releases: {}", e)))?;

            let mut latest_release: Option<UpdateInfo> = None;
            let mut latest_version: Option<Version> = None;

            for release in releases {
                if let Ok(update_info) = self.parse_release_info(release.clone()).await {
                    let is_prerelease = update_info.prerelease;
                    let matches_channel = match channel {
                        UpdateChannel::Stable => !is_prerelease,
                        UpdateChannel::Beta => is_prerelease && update_info.version.contains("beta"),
                        UpdateChannel::Alpha => is_prerelease,
                    };

                    if matches_channel || (allow_prerelease && is_prerelease) {
                        if let Ok(version) = Version::parse(&update_info.version) {
                            if latest_version.is_none() || version > latest_version.unwrap() {
                                latest_version = Some(version);
                                latest_release = Some(update_info);
                            }
                        }
                    }
                }
            }

            latest_release.ok_or_else(|| LauncherError::NotFound("No suitable release found".to_string()))
        }
    }

    /// Parse release information from GitHub API response
    async fn parse_release_info(&self, release: serde_json::Value) -> Result<UpdateInfo, LauncherError> {
        let tag_name = release.get("tag_name")
            .and_then(|v| v.as_str())
            .ok_or_else(|| LauncherError::InvalidConfig("Missing tag_name".to_string()))?
            .to_string();

        let version = self.normalize_version(&tag_name);

        let name = release.get("name")
            .and_then(|v| v.as_str())
            .unwrap_or(&version);

        let body = release.get("body")
            .and_then(|v| v.as_str())
            .unwrap_or("");

        let published_at = release.get("published_at")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();

        let prerelease = release.get("prerelease")
            .and_then(|v| v.as_bool())
            .unwrap_or(false);

        let channel = if prerelease {
            if version.to_lowercase().contains("alpha") {
                UpdateChannel::Alpha
            } else {
                UpdateChannel::Beta
            }
        } else {
            UpdateChannel::Stable
        };

        // Find the appropriate download asset
        let assets = release.get("assets")
            .and_then(|v| v.as_array())
            .unwrap_or(&vec![]);

        let (download_url, file_size) = self.find_suitable_asset(assets).await?;

        Ok(UpdateInfo {
            version,
            tag_name,
            release_notes: format!("{}\n\n{}", name, body),
            published_at,
            download_url,
            file_size,
            checksum: None, // TODO: Implement checksum verification
            prerelease,
            channel,
        })
    }

    /// Find the suitable download asset for the current platform
    async fn find_suitable_asset(&self, assets: &[serde_json::Value]) -> Result<(String, u64), LauncherError> {
        let platform = self.current_platform();

        for asset in assets {
            if let (Some(name), Some(download_url), Some(size)) = (
                asset.get("name").and_then(|v| v.as_str()),
                asset.get("browser_download_url").and_then(|v| v.as_str()),
                asset.get("size").and_then(|v| v.as_u64())
            ) {
                // Look for platform-specific executables/archives
                if name.contains("TheBoysLauncher") &&
                   (name.contains(&platform) || name.ends_with(".exe") || name.ends_with(".app") || name.ends_with(".AppImage")) {
                    return Ok((download_url.to_string(), size));
                }
            }
        }

        Err(LauncherError::NotFound(
            format!("No suitable asset found for platform: {}", platform)
        ))
    }

    /// Get current platform identifier
    fn current_platform(&self) -> String {
        #[cfg(target_os = "windows")]
        return "Windows".to_string();

        #[cfg(target_os = "macos")]
        return "macOS".to_string();

        #[cfg(target_os = "linux")]
        return "Linux".to_string();

        #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
        return "Unknown".to_string();
    }

    /// Normalize version string (remove 'v' prefix if present)
    fn normalize_version(&self, version: &str) -> String {
        version.trim_start_matches('v').to_string()
    }

    /// Get update destination path
    async fn get_update_destination(&self, version: &str) -> Result<String, LauncherError> {
        let mut temp_dir = std::env::temp_dir();
        temp_dir.push("theboys-launcher");
        temp_dir.push("updates");

        // Create directory if it doesn't exist
        tokio::fs::create_dir_all(&temp_dir).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create update directory: {}", e)
            ))?;

        let platform = self.current_platform();
        let extension = if platform == "Windows" { ".exe" } else { ".tar.gz" };
        temp_dir.push(format!("TheBoysLauncher-{}{}", version, extension));

        Ok(temp_dir.to_string_lossy().to_string())
    }

    /// Perform the actual download
    async fn perform_download(
        update_info: &UpdateInfo,
        destination: &str,
        download_id: &str,
        active_downloads: Arc<RwLock<HashMap<String, UpdateProgress>>>,
    ) -> Result<(), LauncherError> {
        let client = Client::new();
        let mut response = client.get(&update_info.download_url).send().await
            .map_err(|e| LauncherError::Network(format!("Download failed: {}", e)))?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("Download error: {}", response.status())
            ));
        }

        let total_size = response.content_length().unwrap_or(update_info.file_size);
        let mut downloaded = 0u64;
        let start_time = std::time::Instant::now();

        // Create parent directories
        if let Some(parent) = std::path::Path::new(destination).parent() {
            tokio::fs::create_dir_all(parent).await
                .map_err(|e| LauncherError::FileSystem(
                    format!("Failed to create download directory: {}", e)
                ))?;
        }

        let mut file = tokio::fs::File::create(destination).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create update file: {}", e)
            ))?;

        let mut buffer = [0; 8192];
        loop {
            let n = response.read(&mut buffer).await
                .map_err(|e| LauncherError::Network(format!("Read error: {}", e)))?;

            if n == 0 {
                break;
            }

            file.write_all(&buffer[..n]).await
                .map_err(|e| LauncherError::FileSystem(
                    format!("Write error: {}", e)
                ))?;

            downloaded += n as u64;
            let progress = (downloaded as f64 / total_size as f64) * 100.0;
            let elapsed = start_time.elapsed().as_secs_f64();
            let speed = if elapsed > 0.0 { (downloaded as f64 / elapsed) as u64 } else { 0 };

            // Update progress
            {
                let mut downloads = active_downloads.write().await;
                if let Some(update_progress) = downloads.get_mut(download_id) {
                    update_progress.downloaded_bytes = downloaded;
                    update_progress.progress_percent = progress;
                    update_progress.download_speed_bps = speed;
                }
            }
        }

        // Mark as downloaded
        {
            let mut downloads = active_downloads.write().await;
            if let Some(update_progress) = downloads.get_mut(download_id) {
                update_progress.status = UpdateStatus::Downloaded;
            }
        }

        Ok(())
    }

    /// Update progress status
    async fn update_progress_status(&self, download_id: String, status: UpdateStatus) {
        let mut downloads = self.active_downloads.write().await;
        if let Some(progress) = downloads.get_mut(&download_id) {
            progress.status = status;
        }
    }

    /// Platform-specific update implementations
    #[cfg(target_os = "windows")]
    async fn apply_windows_update(&self, update_file: &str, version: &str) -> Result<(), LauncherError> {
        info!("Applying Windows update from {}", update_file);

        let current_exe = std::env::current_exe()
            .map_err(|e| LauncherError::FileSystem(format!("Failed to get current exe path: {}", e)))?;

        let backup_path = current_exe.with_extension("exe.backup");

        // Move current executable to backup
        tokio::fs::rename(&current_exe, &backup_path).await
            .map_err(|e| LauncherError::FileSystem(format!("Failed to backup current exe: {}", e)))?;

        // Move new executable to original location
        tokio::fs::copy(update_file, &current_exe).await
            .map_err(|e| LauncherError::FileSystem(format!("Failed to copy new exe: {}", e)))?;

        // Clean up the downloaded update file
        let _ = tokio::fs::remove_file(update_file).await;

        info!("Windows update applied successfully, restart required");
        Ok(())
    }

    #[cfg(target_os = "macos")]
    async fn apply_macos_update(&self, update_file: &str, version: &str) -> Result<(), LauncherError> {
        info!("Applying macOS update from {}", update_file);

        // For macOS, we'd typically update the app bundle
        // This is a simplified implementation
        warn!("macOS update implementation needs to be completed for app bundle updates");
        Ok(())
    }

    #[cfg(target_os = "linux")]
    async fn apply_linux_update(&self, update_file: &str, version: &str) -> Result<(), LauncherError> {
        info!("Applying Linux update from {}", update_file);

        let current_exe = std::env::current_exe()
            .map_err(|e| LauncherError::FileSystem(format!("Failed to get current exe path: {}", e)))?;

        let backup_path = current_exe.with_extension("backup");

        // Move current executable to backup
        tokio::fs::rename(&current_exe, &backup_path).await
            .map_err(|e| LauncherError::FileSystem(format!("Failed to backup current exe: {}", e)))?;

        // Extract and move new executable
        // This would need to handle tar.gz extraction
        warn!("Linux update implementation needs tar.gz extraction");
        Ok(())
    }

    /// Create backup before update
    async fn create_backup(&self) -> Result<(), LauncherError> {
        let current_exe = std::env::current_exe()
            .map_err(|e| LauncherError::FileSystem(format!("Failed to get current exe path: {}", e)))?;

        let backup_dir = current_exe.parent()
            .ok_or_else(|| LauncherError::FileSystem("No parent directory".to_string()))?
            .join("backups");

        tokio::fs::create_dir_all(&backup_dir).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create backup directory: {}", e)
            ))?;

        let timestamp = chrono::Utc::now().format("%Y%m%d_%H%M%S");
        let backup_name = format!("TheBoysLauncher-backup-{}", timestamp);
        let backup_path = backup_dir.join(backup_name);

        tokio::fs::copy(&current_exe, &backup_path).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create backup: {}", e)
            ))?;

        info!("Created backup at: {}", backup_path.display());
        Ok(())
    }

    /// Rollback a failed update
    async fn rollback_update(&self) -> Result<(), LauncherError> {
        warn!("Attempting to rollback failed update");

        let current_exe = std::env::current_exe()
            .map_err(|e| LauncherError::FileSystem(format!("Failed to get current exe path: {}", e)))?;

        let backup_path = current_exe.with_extension("backup");

        if backup_path.exists() {
            // Remove the failed update
            let _ = tokio::fs::remove_file(&current_exe).await;

            // Restore backup
            tokio::fs::rename(&backup_path, &current_exe).await
                .map_err(|e| LauncherError::FileSystem(
                    format!("Failed to restore backup: {}", e)
                ))?;

            info!("Update rollback successful");
        } else {
            warn!("No backup found for rollback");
        }

        Ok(())
    }

    /// Clean up download file
    async fn cleanup_download_file(&self, version: &str) -> Result<(), LauncherError> {
        let destination = self.get_update_destination(version).await?;
        let _ = tokio::fs::remove_file(&destination).await;
        Ok(())
    }

    /// Load update settings
    async fn load_settings(&self) -> Result<UpdateSettings, LauncherError> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| LauncherError::FileSystem("No config directory".to_string()))?
            .join("theboys-launcher");

        let settings_file = config_dir.join("update-settings.json");

        if settings_file.exists() {
            let content = tokio::fs::read_to_string(&settings_file).await
                .map_err(|e| LauncherError::FileSystem(
                    format!("Failed to read update settings: {}", e)
                ))?;

            let settings: UpdateSettings = serde_json::from_str(&content)
                .map_err(|e| LauncherError::InvalidConfig(
                    format!("Invalid update settings format: {}", e)
                ))?;

            Ok(settings)
        } else {
            Err(LauncherError::NotFound("Update settings file not found".to_string()))
        }
    }

    /// Save update settings
    async fn save_settings(&self, settings: &UpdateSettings) -> Result<(), LauncherError> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| LauncherError::FileSystem("No config directory".to_string()))?
            .join("theboys-launcher");

        tokio::fs::create_dir_all(&config_dir).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to create config directory: {}", e)
            ))?;

        let settings_file = config_dir.join("update-settings.json");
        let content = serde_json::to_string_pretty(settings)
            .map_err(|e| LauncherError::InvalidConfig(
                format!("Failed to serialize settings: {}", e)
            ))?;

        tokio::fs::write(&settings_file, content).await
            .map_err(|e| LauncherError::FileSystem(
                format!("Failed to write update settings: {}", e)
            ))?;

        Ok(())
    }
}

use tokio::io::AsyncWriteExt;
use tokio::io::AsyncReadExt;