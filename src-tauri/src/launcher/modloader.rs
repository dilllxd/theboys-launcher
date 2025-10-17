use crate::models::{LauncherResult, Modloader, Instance};
use crate::downloader::download_manager;
use std::path::{Path, PathBuf};
use tokio::fs;
use tokio::process::Command;
use serde_json;
use tracing::{info, warn, error, debug};

/// Modloader installer for handling Forge, Fabric, Quilt, and NeoForge installations
pub struct ModloaderInstaller {
    temp_dir: PathBuf,
}

impl ModloaderInstaller {
    pub fn new() -> Self {
        Self {
            temp_dir: std::env::temp_dir().join("theboys-launcher"),
        }
    }

    /// Install a modloader for an instance
    pub async fn install_modloader(
        &self,
        instance: &Instance,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        info!("Installing modloader {} for instance {}", instance.loader_type, instance.name);

        match instance.loader_type.as_str() {
            "forge" => self.install_forge(instance, progress_callback).await,
            "fabric" => self.install_fabric(instance, progress_callback).await,
            "quilt" => self.install_quilt(instance, progress_callback).await,
            "neoforge" => self.install_neoforge(instance, progress_callback).await,
            "vanilla" => {
                info!("Vanilla instance, no modloader installation needed");
                Ok(())
            },
            _ => Err(crate::models::LauncherError::InvalidConfig(
                format!("Unsupported modloader: {}", instance.loader_type)
            )),
        }
    }

    /// Install Forge modloader
    async fn install_forge(
        &self,
        instance: &Instance,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        info!("Installing Forge {} for Minecraft {}", instance.loader_version, instance.minecraft_version);

        // Create temp directory
        fs::create_dir_all(&self.temp_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create temp directory: {}", e)
            ))?;

        // Download Forge installer
        let forge_url = self.get_forge_download_url(instance).await?;
        let installer_path = self.temp_dir.join(format!("forge-{}-{}-installer.jar",
            instance.minecraft_version, instance.loader_version));

        if let Some(callback) = &progress_callback {
            callback(0.1);
        }

        // Use the download manager for proper progress tracking
        let download_id = crate::downloader::download_manager()
            .start_download(
                format!("Forge Installer {}-{}", instance.minecraft_version, instance.loader_version),
                forge_url,
                installer_path.to_string_lossy().to_string()
            ).await
            .map_err(|e| crate::models::LauncherError::DownloadFailed(e.to_string()))?;

        // Monitor download progress
        let mut last_progress = 0.1;
        while let Some(progress) = crate::downloader::download_manager().get_progress(&download_id).await {
            match progress.status {
                crate::models::DownloadStatus::Downloading => {
                    let current_progress = 0.1 + (progress.progress_percent / 100.0) * 0.3; // 10% to 40%
                    if current_progress - last_progress > 0.05 { // Update every 5%
                        if let Some(callback) = &progress_callback {
                            callback(current_progress);
                        }
                        last_progress = current_progress;
                    }
                },
                crate::models::DownloadStatus::Completed => {
                    break;
                },
                crate::models::DownloadStatus::Failed(_) => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Forge installer download failed")
                    ));
                },
                crate::models::DownloadStatus::Cancelled => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Forge installer download was cancelled")
                    ));
                },
                _ => {
                    tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
                }
            }
        }

        if let Some(callback) = &progress_callback {
            callback(0.4);
        }

        // Run Forge installer
        self.run_forge_installer(instance, &installer_path).await?;

        if let Some(callback) = &progress_callback {
            callback(0.9);
        }

        // Verify installation
        if !self.verify_forge_installation(instance).await? {
            return Err(crate::models::LauncherError::Process(
                "Forge installation verification failed".to_string()
            ));
        }

        // Cleanup temp files
        if let Err(e) = fs::remove_file(&installer_path).await {
            warn!("Failed to cleanup Forge installer: {}", e);
        }

        if let Some(callback) = &progress_callback {
            callback(1.0);
        }

        info!("Forge installation completed successfully");
        Ok(())
    }

    /// Install Fabric modloader
    async fn install_fabric(
        &self,
        instance: &Instance,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        info!("Installing Fabric {} for Minecraft {}", instance.loader_version, instance.minecraft_version);

        // Create temp directory
        fs::create_dir_all(&self.temp_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create temp directory: {}", e)
            ))?;

        // Download Fabric installer
        let fabric_url = format!("https://maven.fabricmc.net/net/fabricmc/fabric-installer/{}/fabric-installer-{}.jar",
            instance.loader_version, instance.loader_version);
        let installer_path = self.temp_dir.join(format!("fabric-installer-{}.jar", instance.loader_version));

        if let Some(callback) = &progress_callback {
            callback(0.1);
        }

        // Use the download manager for proper progress tracking
        let download_id = crate::downloader::download_manager()
            .start_download(
                format!("Fabric Installer {}", instance.loader_version),
                fabric_url,
                installer_path.to_string_lossy().to_string()
            ).await
            .map_err(|e| crate::models::LauncherError::DownloadFailed(e.to_string()))?;

        // Monitor download progress
        let mut last_progress = 0.1;
        while let Some(progress) = crate::downloader::download_manager().get_progress(&download_id).await {
            match progress.status {
                crate::models::DownloadStatus::Downloading => {
                    let current_progress = 0.1 + (progress.progress_percent / 100.0) * 0.3; // 10% to 40%
                    if current_progress - last_progress > 0.05 { // Update every 5%
                        if let Some(callback) = &progress_callback {
                            callback(current_progress);
                        }
                        last_progress = current_progress;
                    }
                },
                crate::models::DownloadStatus::Completed => {
                    break;
                },
                crate::models::DownloadStatus::Failed(_) => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Fabric installer download failed")
                    ));
                },
                crate::models::DownloadStatus::Cancelled => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Fabric installer download was cancelled")
                    ));
                },
                _ => {
                    tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
                }
            }
        }

        if let Some(callback) = &progress_callback {
            callback(0.4);
        }

        // Run Fabric installer
        self.run_fabric_installer(instance, &installer_path).await?;

        if let Some(callback) = &progress_callback {
            callback(0.9);
        }

        // Verify installation
        if !self.verify_fabric_installation(instance).await? {
            return Err(crate::models::LauncherError::Process(
                "Fabric installation verification failed".to_string()
            ));
        }

        // Cleanup temp files
        if let Err(e) = fs::remove_file(&installer_path).await {
            warn!("Failed to cleanup Fabric installer: {}", e);
        }

        if let Some(callback) = &progress_callback {
            callback(1.0);
        }

        info!("Fabric installation completed successfully");
        Ok(())
    }

    /// Install Quilt modloader
    async fn install_quilt(
        &self,
        instance: &Instance,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        info!("Installing Quilt {} for Minecraft {}", instance.loader_version, instance.minecraft_version);

        // Create temp directory
        fs::create_dir_all(&self.temp_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create temp directory: {}", e)
            ))?;

        // Download Quilt installer
        let quilt_url = format!("https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/{}/quilt-installer-{}.jar",
            instance.loader_version, instance.loader_version);
        let installer_path = self.temp_dir.join(format!("quilt-installer-{}.jar", instance.loader_version));

        if let Some(callback) = &progress_callback {
            callback(0.1);
        }

        // Use the download manager for proper progress tracking
        let download_id = crate::downloader::download_manager()
            .start_download(
                format!("Quilt Installer {}", instance.loader_version),
                quilt_url,
                installer_path.to_string_lossy().to_string()
            ).await
            .map_err(|e| crate::models::LauncherError::DownloadFailed(e.to_string()))?;

        // Monitor download progress
        let mut last_progress = 0.1;
        while let Some(progress) = crate::downloader::download_manager().get_progress(&download_id).await {
            match progress.status {
                crate::models::DownloadStatus::Downloading => {
                    let current_progress = 0.1 + (progress.progress_percent / 100.0) * 0.3; // 10% to 40%
                    if current_progress - last_progress > 0.05 { // Update every 5%
                        if let Some(callback) = &progress_callback {
                            callback(current_progress);
                        }
                        last_progress = current_progress;
                    }
                },
                crate::models::DownloadStatus::Completed => {
                    break;
                },
                crate::models::DownloadStatus::Failed(_) => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Quilt installer download failed")
                    ));
                },
                crate::models::DownloadStatus::Cancelled => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("Quilt installer download was cancelled")
                    ));
                },
                _ => {
                    tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
                }
            }
        }

        if let Some(callback) = &progress_callback {
            callback(0.4);
        }

        // Run Quilt installer
        self.run_quilt_installer(instance, &installer_path).await?;

        if let Some(callback) = &progress_callback {
            callback(0.9);
        }

        // Verify installation
        if !self.verify_quilt_installation(instance).await? {
            return Err(crate::models::LauncherError::Process(
                "Quilt installation verification failed".to_string()
            ));
        }

        // Cleanup temp files
        if let Err(e) = fs::remove_file(&installer_path).await {
            warn!("Failed to cleanup Quilt installer: {}", e);
        }

        if let Some(callback) = &progress_callback {
            callback(1.0);
        }

        info!("Quilt installation completed successfully");
        Ok(())
    }

    /// Install NeoForge modloader
    async fn install_neoforge(
        &self,
        instance: &Instance,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        info!("Installing NeoForge {} for Minecraft {}", instance.loader_version, instance.minecraft_version);

        // Create temp directory
        fs::create_dir_all(&self.temp_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create temp directory: {}", e)
            ))?;

        // Download NeoForge installer
        let neoforge_url = self.get_neoforge_download_url(instance).await?;
        let installer_path = self.temp_dir.join(format!("neoforge-{}-installer.jar", instance.loader_version));

        if let Some(callback) = &progress_callback {
            callback(0.1);
        }

        // Use the download manager for proper progress tracking
        let download_id = crate::downloader::download_manager()
            .start_download(
                format!("NeoForge Installer {}", instance.loader_version),
                neoforge_url,
                installer_path.to_string_lossy().to_string()
            ).await
            .map_err(|e| crate::models::LauncherError::DownloadFailed(e.to_string()))?;

        // Monitor download progress
        let mut last_progress = 0.1;
        while let Some(progress) = crate::downloader::download_manager().get_progress(&download_id).await {
            match progress.status {
                crate::models::DownloadStatus::Downloading => {
                    let current_progress = 0.1 + (progress.progress_percent / 100.0) * 0.3; // 10% to 40%
                    if current_progress - last_progress > 0.05 { // Update every 5%
                        if let Some(callback) = &progress_callback {
                            callback(current_progress);
                        }
                        last_progress = current_progress;
                    }
                },
                crate::models::DownloadStatus::Completed => {
                    break;
                },
                crate::models::DownloadStatus::Failed(_) => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("NeoForge installer download failed")
                    ));
                },
                crate::models::DownloadStatus::Cancelled => {
                    return Err(crate::models::LauncherError::DownloadFailed(
                        format!("NeoForge installer download was cancelled")
                    ));
                },
                _ => {
                    tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
                }
            }
        }

        if let Some(callback) = &progress_callback {
            callback(0.4);
        }

        // Run NeoForge installer
        self.run_neoforge_installer(instance, &installer_path).await?;

        if let Some(callback) = &progress_callback {
            callback(0.9);
        }

        // Verify installation
        if !self.verify_neoforge_installation(instance).await? {
            return Err(crate::models::LauncherError::Process(
                "NeoForge installation verification failed".to_string()
            ));
        }

        // Cleanup temp files
        if let Err(e) = fs::remove_file(&installer_path).await {
            warn!("Failed to cleanup NeoForge installer: {}", e);
        }

        if let Some(callback) = &progress_callback {
            callback(1.0);
        }

        info!("NeoForge installation completed successfully");
        Ok(())
    }

    /// Get Forge download URL using version manifest
    async fn get_forge_download_url(&self, instance: &Instance) -> LauncherResult<String> {
        // Forge version manifest URL
        let manifest_url = format!("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json");

        let response = reqwest::get(&manifest_url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let manifest: serde_json::Value = response.json().await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        // Look for the specific version
        let promo_key = format!("{}-recommended", instance.minecraft_version);
        let version = manifest.get("promos")
            .and_then(|p| p.get(&promo_key))
            .and_then(|v| v.as_str())
            .unwrap_or(&instance.loader_version);

        let download_url = format!(
            "https://maven.minecraftforge.net/net/minecraftforge/forge/{}/forge-{}-installer.jar",
            format!("{}-{}", instance.minecraft_version, version),
            format!("{}-{}", instance.minecraft_version, version)
        );

        Ok(download_url)
    }

    /// Get NeoForge download URL
    async fn get_neoforge_download_url(&self, instance: &Instance) -> LauncherResult<String> {
        // NeoForge Maven URL pattern
        let download_url = format!(
            "https://maven.neoforged.net/api/maven/versions/releases/net/neoforged/neoforge/{}",
            instance.loader_version
        );

        // First get the version list to find the correct file
        let response = reqwest::get(&download_url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        if response.status().is_success() {
            let versions: serde_json::Value = response.json().await
                .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

            if let Some(versions_array) = versions.get("versions").and_then(|v| v.as_array()) {
                if let Some(version_info) = versions_array.last() {
                    if let Some(version_str) = version_info.as_str() {
                        return Ok(format!(
                            "https://maven.neoforged.net/releases/net/neoforged/neoforge/{}/neoforge-{}-installer.jar",
                            version_str, version_str
                        ));
                    }
                }
            }
        }

        // Fallback to common pattern
        Ok(format!(
            "https://maven.neoforged.net/releases/net/neoforged/neoforge/{}/neoforge-{}-installer.jar",
            instance.loader_version, instance.loader_version
        ))
    }

    /// Run Forge installer
    async fn run_forge_installer(&self, instance: &Instance, installer_path: &Path) -> LauncherResult<()> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");

        let output = Command::new(&instance.java_path)
            .arg("-jar")
            .arg(installer_path)
            .arg("--installClient")
            .arg("--dir")
            .arg(&minecraft_dir)
            .output()
            .await
            .map_err(|e| crate::models::LauncherError::Process(
                format!("Failed to run Forge installer: {}", e)
            ))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            error!("Forge installer failed. stderr: {}, stdout: {}", stderr, stdout);
            return Err(crate::models::LauncherError::Process(
                format!("Forge installer failed: {}", stderr)
            ));
        }

        debug!("Forge installer output: {}", String::from_utf8_lossy(&output.stdout));
        Ok(())
    }

    /// Run Fabric installer
    async fn run_fabric_installer(&self, instance: &Instance, installer_path: &Path) -> LauncherResult<()> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");

        let output = Command::new(&instance.java_path)
            .arg("-jar")
            .arg(installer_path)
            .arg("client")
            .arg("-dir")
            .arg(&minecraft_dir)
            .arg("-mcversion")
            .arg(&instance.minecraft_version)
            .output()
            .await
            .map_err(|e| crate::models::LauncherError::Process(
                format!("Failed to run Fabric installer: {}", e)
            ))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            error!("Fabric installer failed. stderr: {}, stdout: {}", stderr, stdout);
            return Err(crate::models::LauncherError::Process(
                format!("Fabric installer failed: {}", stderr)
            ));
        }

        debug!("Fabric installer output: {}", String::from_utf8_lossy(&output.stdout));
        Ok(())
    }

    /// Run Quilt installer
    async fn run_quilt_installer(&self, instance: &Instance, installer_path: &Path) -> LauncherResult<()> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");

        let output = Command::new(&instance.java_path)
            .arg("-jar")
            .arg(installer_path)
            .arg("install")
            .arg("client")
            .arg("--install-dir")
            .arg(&minecraft_dir)
            .arg("--minecraft-version")
            .arg(&instance.minecraft_version)
            .output()
            .await
            .map_err(|e| crate::models::LauncherError::Process(
                format!("Failed to run Quilt installer: {}", e)
            ))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            error!("Quilt installer failed. stderr: {}, stdout: {}", stderr, stdout);
            return Err(crate::models::LauncherError::Process(
                format!("Quilt installer failed: {}", stderr)
            ));
        }

        debug!("Quilt installer output: {}", String::from_utf8_lossy(&output.stdout));
        Ok(())
    }

    /// Run NeoForge installer
    async fn run_neoforge_installer(&self, instance: &Instance, installer_path: &Path) -> LauncherResult<()> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");

        let output = Command::new(&instance.java_path)
            .arg("-jar")
            .arg(installer_path)
            .arg("--install-client")
            .arg("--dir")
            .arg(&minecraft_dir)
            .output()
            .await
            .map_err(|e| crate::models::LauncherError::Process(
                format!("Failed to run NeoForge installer: {}", e)
            ))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            let stdout = String::from_utf8_lossy(&output.stdout);
            error!("NeoForge installer failed. stderr: {}, stdout: {}", stderr, stdout);
            return Err(crate::models::LauncherError::Process(
                format!("NeoForge installer failed: {}", stderr)
            ));
        }

        debug!("NeoForge installer output: {}", String::from_utf8_lossy(&output.stdout));
        Ok(())
    }

    /// Verify Forge installation
    async fn verify_forge_installation(&self, instance: &Instance) -> LauncherResult<bool> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");
        let forge_jar = minecraft_dir.join("libraries")
            .join("net")
            .join("minecraftforge")
            .join("forge")
            .join(format!("{}-{}", instance.minecraft_version, instance.loader_version))
            .join(format!("forge-{}-{}-universal.jar", instance.minecraft_version, instance.loader_version));

        Ok(forge_jar.exists())
    }

    /// Verify Fabric installation
    async fn verify_fabric_installation(&self, instance: &Instance) -> LauncherResult<bool> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");
        let fabric_jar = minecraft_dir.join("libraries")
            .join("net")
            .join("fabricmc")
            .join("fabric-loader")
            .join(&instance.loader_version)
            .join(format!("fabric-loader-{}.jar", instance.loader_version));

        Ok(fabric_jar.exists())
    }

    /// Verify Quilt installation
    async fn verify_quilt_installation(&self, instance: &Instance) -> LauncherResult<bool> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");
        let quilt_jar = minecraft_dir.join("libraries")
            .join("org")
            .join("quiltmc")
            .join("quilt-loader")
            .join(&instance.loader_version)
            .join(format!("quilt-loader-{}.jar", instance.loader_version));

        Ok(quilt_jar.exists())
    }

    /// Verify NeoForge installation
    async fn verify_neoforge_installation(&self, instance: &Instance) -> LauncherResult<bool> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");
        let neoforge_jar = minecraft_dir.join("libraries")
            .join("net")
            .join("neoforged")
            .join("neoforge")
            .join(&instance.loader_version)
            .join(format!("neoforge-{}.jar", instance.loader_version));

        Ok(neoforge_jar.exists())
    }

    /// Get available versions for a modloader
    pub async fn get_available_versions(&self, modloader: &Modloader, minecraft_version: &str) -> LauncherResult<Vec<String>> {
        match modloader {
            Modloader::Forge => self.get_forge_versions(minecraft_version).await,
            Modloader::Fabric => self.get_fabric_versions(minecraft_version).await,
            Modloader::Quilt => self.get_quilt_versions(minecraft_version).await,
            Modloader::NeoForge => self.get_neoforge_versions(minecraft_version).await,
            Modloader::Vanilla => Ok(vec!["latest".to_string()]),
        }
    }

    /// Get available Forge versions
    async fn get_forge_versions(&self, minecraft_version: &str) -> LauncherResult<Vec<String>> {
        let manifest_url = "https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json";

        let response = reqwest::get(manifest_url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let manifest: serde_json::Value = response.json().await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let mut versions = Vec::new();
        let promos = manifest.get("promos").and_then(|p| p.as_object());

        if let Some(promos) = promos {
            // Look for recommended version first
            let recommended_key = format!("{}-recommended", minecraft_version);
            if let Some(version) = promos.get(&recommended_key).and_then(|v| v.as_str()) {
                versions.push(version.to_string());
            }

            // Look for latest version
            let latest_key = format!("{}-latest", minecraft_version);
            if let Some(version) = promos.get(&latest_key).and_then(|v| v.as_str()) {
                if !versions.contains(&version.to_string()) {
                    versions.push(version.to_string());
                }
            }
        }

        Ok(versions)
    }

    /// Get available Fabric versions
    async fn get_fabric_versions(&self, minecraft_version: &str) -> LauncherResult<Vec<String>> {
        let url = "https://meta.fabricmc.net/v2/versions/loader";

        let response = reqwest::get(url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let versions: Vec<serde_json::Value> = response.json().await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let mut loader_versions = Vec::new();
        for version in versions {
            if let Some(loader_version) = version.get("loader").and_then(|v| v.as_str()) {
                // Check if this version supports the Minecraft version
                if let Some(minecraft_versions) = version.get("game").and_then(|v| v.as_array()) {
                    for mc_version in minecraft_versions {
                        if let Some(version_str) = mc_version.as_str() {
                            if version_str == minecraft_version {
                                loader_versions.push(loader_version.to_string());
                                break;
                            }
                        }
                    }
                }
            }
        }

        // Sort versions (newest first) and limit to recent versions
        loader_versions.sort_by(|a, b| b.cmp(a));
        loader_versions.truncate(10);

        Ok(loader_versions)
    }

    /// Get available Quilt versions
    async fn get_quilt_versions(&self, minecraft_version: &str) -> LauncherResult<Vec<String>> {
        let url = "https://meta.quiltmc.org/v3/versions/loader";

        let response = reqwest::get(url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let versions: Vec<serde_json::Value> = response.json().await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        let mut loader_versions = Vec::new();
        for version in versions {
            if let Some(loader_version) = version.get("version").and_then(|v| v.as_str()) {
                // Check if this version supports the Minecraft version
                if let Some(minecraft_versions) = version.get("game").and_then(|v| v.as_array()) {
                    for mc_version in minecraft_versions {
                        if let Some(version_str) = mc_version.as_str() {
                            if version_str == minecraft_version {
                                loader_versions.push(loader_version.to_string());
                                break;
                            }
                        }
                    }
                }
            }
        }

        // Sort versions (newest first) and limit to recent versions
        loader_versions.sort_by(|a, b| b.cmp(a));
        loader_versions.truncate(10);

        Ok(loader_versions)
    }

    /// Get available NeoForge versions
    async fn get_neoforge_versions(&self, minecraft_version: &str) -> LauncherResult<Vec<String>> {
        let url = "https://maven.neoforged.net/api/maven/versions/releases/net/neoforged/neoforge";

        let response = reqwest::get(url).await
            .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

        if response.status().is_success() {
            let versions_data: serde_json::Value = response.json().await
                .map_err(|e| crate::models::LauncherError::Network(e.to_string()))?;

            if let Some(versions) = versions_data.get("versions").and_then(|v| v.as_array()) {
                let mut version_strings = Vec::new();
                for version in versions {
                    if let Some(version_str) = version.as_str() {
                        // For NeoForge, we need to check if the version supports the Minecraft version
                        // This is a simplified check - in a full implementation, you'd want to
                        // check the NeoForge version manifest for compatibility
                        if version_str.starts_with(minecraft_version) ||
                           (!version_str.starts_with("1.") && minecraft_version.starts_with("1.20")) {
                            version_strings.push(version_str.to_string());
                        }
                    }
                }

                // Sort versions (newest first) and limit to recent versions
                version_strings.sort_by(|a, b| b.cmp(a));
                version_strings.truncate(10);

                return Ok(version_strings);
            }
        }

        // Fallback - return some common versions
        match minecraft_version {
            "1.20.1" => Ok(vec!["47.1.79".to_string(), "47.1.78".to_string()]),
            "1.20.2" => Ok(vec!["47.2.20".to_string(), "47.2.19".to_string()]),
            "1.20.4" => Ok(vec!["47.3.12".to_string(), "47.3.11".to_string()]),
            "1.20.6" => Ok(vec!["47.4.2".to_string(), "47.4.1".to_string()]),
            _ => Ok(vec![]),
        }
    }
}

impl Default for ModloaderInstaller {
    fn default() -> Self {
        Self::new()
    }
}