use crate::models::*;
use crate::downloader::DownloadManager;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use tokio::process::Command;
use std::time::{SystemTime, UNIX_EPOCH};
use std::io::{Read, Write};
use tokio::fs;
use tokio::sync::RwLock;
use tokio::io::AsyncWriteExt;
use tracing::{info, warn, error, debug};
use uuid::Uuid;

/// Packwiz manager for handling packwiz operations
pub struct PackwizManager {
    bootstrap_path: PathBuf,
    instances_path: PathBuf,
    temp_path: PathBuf,
    download_manager: std::sync::Arc<tokio::sync::Mutex<DownloadManager>>,
    progress: RwLock<std::collections::HashMap<String, PackInstallProgress>>,
    manual_downloads: RwLock<std::collections::HashMap<String, Vec<ManualDownload>>>,
}

impl PackwizManager {
    pub fn new<P: AsRef<Path>>(instances_path: P) -> Self {
        let instances_path = instances_path.as_ref().to_path_buf();
        let tools_path = instances_path.join("..").join("tools");
        let bootstrap_path = tools_path.join("packwiz").join("packwiz.exe");
        let temp_path = instances_path.join("..").join("temp");

        Self {
            bootstrap_path,
            instances_path,
            temp_path,
            download_manager: std::sync::Arc::new(tokio::sync::Mutex::new(
                DownloadManager::new()
            )),
            progress: RwLock::new(std::collections::HashMap::new()),
            manual_downloads: RwLock::new(std::collections::HashMap::new()),
        }
    }

    /// Initialize the packwiz manager
    pub async fn initialize(&self) -> LauncherResult<()> {
        info!("Initializing PackwizManager");

        // Create necessary directories
        fs::create_dir_all(&self.instances_path).await?;
        fs::create_dir_all(&self.temp_path).await?;
        fs::create_dir_all(self.bootstrap_path.parent().unwrap()).await?;

        // Ensure packwiz bootstrap is available
        if !self.bootstrap_path.exists() {
            info!("Packwiz bootstrap not found, downloading");
            self.download_bootstrap().await?;
        }

        info!("PackwizManager initialized successfully");
        Ok(())
    }

    /// Download packwiz bootstrap
    async fn download_bootstrap(&self) -> LauncherResult<()> {
        let bootstrap_url = self.get_packwiz_download_url().await?;
        let _bootstrap_dir = self.bootstrap_path.parent().unwrap();

        info!("Downloading packwiz bootstrap from: {}", bootstrap_url);

        let client = reqwest::Client::new();
        let response = client.get(&bootstrap_url)
            .header("User-Agent", "TheBoys-Launcher/1.1.0")
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("Failed to download packwiz bootstrap: {}", response.status())
            ));
        }

        let bytes = response.bytes().await?;
        fs::write(&self.bootstrap_path, bytes).await?;

        // Mark as executable on Unix systems
        #[cfg(not(target_os = "windows"))]
        {
            use std::os::unix::fs::PermissionsExt;
            let mut perms = fs::metadata(&self.bootstrap_path).await?.permissions();
            perms.set_mode(0o755);
            fs::set_permissions(&self.bootstrap_path, perms).await?;
        }

        info!("Packwiz bootstrap downloaded successfully");
        Ok(())
    }

    /// Get packwiz download URL based on platform
    async fn get_packwiz_download_url(&self) -> LauncherResult<String> {
        let api_url = "https://api.github.com/repos/packwiz/packwiz-installer-bootstrap/releases/latest";

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

        let release_info: serde_json::Value = response.json().await?;

        if let Some(assets) = release_info.get("assets").and_then(|a| a.as_array()) {
            let platform = determine_platform();
            let pattern = match platform {
                "windows" => "packwiz-installer-bootstrap-.*windows.*amd64.*\\.exe$",
                "linux" => "packwiz-installer-bootstrap-.*linux.*amd64$",
                "macos" => "packwiz-installer-bootstrap-.*darwin.*amd64$",
                _ => return Err(LauncherError::NotImplemented(format!("Unsupported platform: {}", platform))),
            };

            let regex = regex::Regex::new(pattern).unwrap();

            for asset in assets {
                if let (Some(name), Some(download_url)) = (
                    asset.get("name").and_then(|n| n.as_str()),
                    asset.get("browser_download_url").and_then(|u| u.as_str())
                ) {
                    if regex.is_match(name) {
                        return Ok(download_url.to_string());
                    }
                }
            }
        }

        Err(LauncherError::DownloadFailed(
            "Could not find suitable packwiz bootstrap download".to_string()
        ))
    }

    /// Install or update a modpack using packwiz
    pub async fn install_modpack(
        &self,
        instance_id: &str,
        pack_url: &str,
        options: UpdateOptions,
    ) -> LauncherResult<String> {
        info!("Installing modpack for instance {} from {}", instance_id, pack_url);

        let install_id = Uuid::new_v4().to_string();
        let instance_path = self.instances_path.join(instance_id);

        // Create instance directory if it doesn't exist
        fs::create_dir_all(&instance_path).await?;

        // Create backup if requested and instance exists
        if options.create_backup && self.is_instance_installed(instance_id).await {
            self.create_backup(instance_id, options.backup_description.as_deref()).await?;
        }

        // Update progress
        self.update_progress(&install_id, PackInstallProgress {
            instance_id: instance_id.to_string(),
            modpack_id: pack_url.to_string(),
            step: PackInstallStep::Downloading,
            progress_percent: 0.0,
            message: "Starting modpack installation...".to_string(),
            status: PackInstallStatus::Running,
        }).await;

        // Run packwiz install in instance directory
        let result = self.run_packwiz_install(&instance_path, pack_url, &install_id).await;

        match result {
            Ok(_) => {
                self.update_progress(&install_id, PackInstallProgress {
                    instance_id: instance_id.to_string(),
                    modpack_id: pack_url.to_string(),
                    step: PackInstallStep::Completed,
                    progress_percent: 100.0,
                    message: "Installation completed successfully".to_string(),
                    status: PackInstallStatus::Completed,
                }).await;

                info!("Modpack installation completed for instance {}", instance_id);
                Ok(install_id)
            }
            Err(e) => {
                self.update_progress(&install_id, PackInstallProgress {
                    instance_id: instance_id.to_string(),
                    modpack_id: pack_url.to_string(),
                    step: PackInstallStep::Downloading,
                    progress_percent: 0.0,
                    message: format!("Installation failed: {}", e),
                    status: PackInstallStatus::Failed(e.to_string()),
                }).await;

                error!("Modpack installation failed for instance {}: {}", instance_id, e);
                Err(e)
            }
        }
    }

    /// Run packwiz install command
    async fn run_packwiz_install(
        &self,
        instance_path: &Path,
        pack_url: &str,
        install_id: &str,
    ) -> LauncherResult<()> {
        info!("Running packwiz install in {}", instance_path.display());

        // Update progress
        self.update_progress(install_id, PackInstallProgress {
            instance_id: instance_path.file_name().unwrap().to_string_lossy().to_string(),
            modpack_id: pack_url.to_string(),
            step: PackInstallStep::Extracting,
            progress_percent: 10.0,
            message: "Initializing packwiz...".to_string(),
            status: PackInstallStatus::Running,
        }).await;

        // Check for manual downloads first
        if let Ok(manual_downloads) = self.check_manual_downloads(pack_url).await {
            if !manual_downloads.is_empty() {
                self.manual_downloads.write().await.insert(install_id.to_string(), manual_downloads.clone());

                self.update_progress(install_id, PackInstallProgress {
                    instance_id: instance_path.file_name().unwrap().to_string_lossy().to_string(),
                    modpack_id: pack_url.to_string(),
                    step: PackInstallStep::DownloadingDependencies,
                    progress_percent: 20.0,
                    message: format!("{} manual downloads required", manual_downloads.len()),
                    status: PackInstallStatus::Failed("Manual downloads required".to_string()),
                }).await;

                return Err(LauncherError::DownloadFailed(
                    format!("{} manual downloads required", manual_downloads.len())
                ));
            }
        }

        // Execute packwiz command
        let mut cmd = Command::new(&self.bootstrap_path);
        cmd.current_dir(instance_path)
           .arg("install")
           .arg(pack_url)
           .stdout(Stdio::piped())
           .stderr(Stdio::piped());

        let mut child = cmd.spawn()
            .map_err(|e| LauncherError::Process(
                format!("Failed to start packwiz: {}", e)
            ))?;

        // Monitor progress
        let stdout = child.stdout.take().unwrap();
        let stderr = child.stderr.take().unwrap();

        use tokio::io::{AsyncBufReadExt, BufReader};

        let _install_id_clone = install_id.to_string();
        let _instance_path_clone = instance_path.to_path_buf();
        let _pack_url_clone = pack_url.to_string();

        // Monitor stdout
        let stdout_task = tokio::spawn(async move {
            let reader = BufReader::new(stdout);
            let mut lines = reader.lines();

            while let Ok(Some(line)) = lines.next_line().await {
                debug!("packwiz stdout: {}", line);

                // Update progress based on output
                let progress = parse_packwiz_progress(&line);
                if let Some((_step, _percent, _message)) = progress {
                    // Update progress in the manager
                    // Note: This would need access to self, so we'd need to restructure
                }
            }
        });

        // Monitor stderr
        let stderr_task = tokio::spawn(async move {
            let reader = BufReader::new(stderr);
            let mut lines = reader.lines();

            while let Ok(Some(line)) = lines.next_line().await {
                warn!("packwiz stderr: {}", line);
            }
        });

        // Wait for completion
        let status = child.wait().await
            .map_err(|e| LauncherError::Process(
                format!("Failed to wait for packwiz: {}", e)
            ))?;

        // Wait for output tasks to complete
        let _ = tokio::try_join!(stdout_task, stderr_task);

        if status.success() {
            info!("packwiz install completed successfully");
            Ok(())
        } else {
            let code = status.code().unwrap_or(-1);
            Err(LauncherError::Process(
                format!("packwiz install failed with exit code: {}", code)
            ))
        }
    }

    /// Check for manual downloads required by a modpack
    async fn check_manual_downloads(&self, _pack_url: &str) -> LauncherResult<Vec<ManualDownload>> {
        // This would analyze the pack.toml manifest to identify manual downloads
        // For now, return empty - would need to implement TOML parsing
        Ok(vec![])
    }

    /// Create a backup of an instance
    pub async fn create_backup(
        &self,
        instance_id: &str,
        description: Option<&str>,
    ) -> LauncherResult<BackupInfo> {
        info!("Creating backup for instance: {}", instance_id);

        let instance_path = self.instances_path.join(instance_id);
        if !instance_path.exists() {
            return Err(LauncherError::InstanceNotFound(
                format!("Instance {} not found", instance_id)
            ));
        }

        let backup_id = Uuid::new_v4().to_string();
        let backups_dir = self.instances_path.join("..").join("backups");
        fs::create_dir_all(&backups_dir).await?;

        let backup_path = backups_dir.join(format!("{}.zip", backup_id));
        let timestamp = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();

        // Create zip archive
        self.create_zip_archive(&instance_path, &backup_path).await?;

        // Get backup size
        let size_bytes = fs::metadata(&backup_path).await?.len();

        // Get current version from instance info
        let version = self.get_instance_version(instance_id).await
            .unwrap_or_else(|_| "unknown".to_string());

        let backup_info = BackupInfo {
            id: backup_id.clone(),
            instance_id: instance_id.to_string(),
            backup_date: timestamp.to_string(),
            version,
            size_bytes,
            backup_path: backup_path.to_string_lossy().to_string(),
            description: description.map(|s| s.to_string()),
        };

        // Save backup info
        let info_path = backups_dir.join(format!("{}.json", backup_id));
        fs::write(info_path, serde_json::to_string_pretty(&backup_info)?).await?;

        info!("Backup created successfully: {}", backup_id);
        Ok(backup_info)
    }

    /// Create a zip archive
    async fn create_zip_archive(&self, source: &Path, destination: &Path) -> LauncherResult<()> {
        use std::fs::File;
        use zip::ZipWriter;

        let file = File::create(destination)?;
        let mut zip = ZipWriter::new(file);

        self.add_to_zip(&mut zip, source, "").await?;
        zip.finish()?;

        Ok(())
    }

    /// Recursively add directory to zip
    async fn add_to_zip(
        &self,
        zip: &mut zip::ZipWriter<std::fs::File>,
        path: &Path,
        prefix: &str,
    ) -> LauncherResult<()> {
        let mut entries = fs::read_dir(path).await?;

        while let Some(entry) = entries.next_entry().await? {
            let entry_path = entry.path();
            let file_name = entry.file_name();
            let name = file_name.to_string_lossy();
            let zip_name = if prefix.is_empty() {
                name.to_string()
            } else {
                format!("{}/{}", prefix, name)
            };

            if entry_path.is_dir() {
                let options: zip::write::FileOptions<'_, ()> = zip::write::FileOptions::default()
                    .compression_method(zip::CompressionMethod::Deflated);
                zip.add_directory(zip_name.clone(), options)
                    .map_err(|e| LauncherError::FileSystem(
                        format!("Failed to add directory to zip: {}", e)
                    ))?;
                Box::pin(self.add_to_zip(zip, &entry_path, &zip_name)).await?;
            } else {
                let options: zip::write::FileOptions<'_, ()> = zip::write::FileOptions::default()
                    .compression_method(zip::CompressionMethod::Deflated);
                zip.start_file(zip_name, options)
                    .map_err(|e| LauncherError::FileSystem(
                        format!("Failed to start file in zip: {}", e)
                    ))?;
                let contents = fs::read(&entry_path).await?;
                zip.write_all(&contents)
                    .map_err(|e| LauncherError::FileSystem(
                        format!("Failed to write to zip: {}", e)
                    ))?;
            }
        }

        Ok(())
    }

    /// Restore instance from backup
    pub async fn restore_backup(&self, backup_id: &str, instance_id: &str) -> LauncherResult<()> {
        info!("Restoring backup {} to instance {}", backup_id, instance_id);

        let backups_dir = self.instances_path.join("..").join("backups");
        let backup_path = backups_dir.join(format!("{}.zip", backup_id));
        let info_path = backups_dir.join(format!("{}.json", backup_id));

        if !backup_path.exists() || !info_path.exists() {
            return Err(LauncherError::InstanceNotFound(
                format!("Backup {} not found", backup_id)
            ));
        }

        // Load backup info
        let info_content = fs::read_to_string(&info_path).await?;
        let backup_info: BackupInfo = serde_json::from_str(&info_content)?;

        // Create/replace instance directory
        let instance_path = self.instances_path.join(instance_id);
        if instance_path.exists() {
            fs::remove_dir_all(&instance_path).await?;
        }
        fs::create_dir_all(&instance_path).await?;

        // Extract backup
        self.extract_zip_archive(&backup_path, &instance_path).await?;

        // Update instance info
        self.update_instance_info(instance_id, &backup_info).await?;

        info!("Backup restored successfully");
        Ok(())
    }

    /// Extract zip archive
    async fn extract_zip_archive(&self, source: &Path, destination: &Path) -> LauncherResult<()> {
        use std::fs::File;
        use zip::ZipArchive;

        let file = File::open(source)?;
        let mut archive = ZipArchive::new(file).map_err(|e| LauncherError::FileSystem(
            format!("Failed to open zip archive: {}", e)
        ))?;

        // Collect all file info first, then extract
        let mut files_to_extract = Vec::new();

        for i in 0..archive.len() {
            let mut file = archive.by_index(i).map_err(|e| LauncherError::FileSystem(
                format!("Failed to read file from archive: {}", e)
            ))?;

            let path = destination.join(file.mangled_name());
            let file_name = file.name().to_string();
            let is_dir = file_name.ends_with('/');

            // Read all file data before any async operations
            let file_data = if !is_dir {
                let mut buffer = Vec::new();
                file.read_to_end(&mut buffer).map_err(|e| LauncherError::FileSystem(
                    format!("Failed to read from archive: {}", e)
                ))?;
                Some(buffer)
            } else {
                None
            };

            // Extract Unix mode before dropping the file
            #[cfg(unix)]
            let unix_mode = if !is_dir {
                use std::os::unix::fs::PermissionsExt;
                file.unix_mode()
            } else {
                None
            };

            #[cfg(not(unix))]
            let unix_mode = ();

            files_to_extract.push((path, is_dir, file_data, unix_mode));
        }

        // Now extract all files without holding any ZipFile references
        for (path, is_dir, file_data, unix_mode) in files_to_extract {
            if is_dir {
                fs::create_dir_all(&path).await?;
            } else {
                if let Some(parent) = path.parent() {
                    fs::create_dir_all(parent).await?;
                }
                let mut outfile = fs::File::create(&path).await?;

                if let Some(data) = file_data {
                    outfile.write_all(&data).await?;
                }
            }

            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                if let Some(mode) = unix_mode {
                    let mut perms = fs::metadata(&path).await?.permissions();
                    perms.set_mode(mode);
                    fs::set_permissions(&path, perms).await?;
                }
            }
        }

        Ok(())
    }

    /// Get list of backups for an instance
    pub async fn get_backups(&self, instance_id: &str) -> LauncherResult<Vec<BackupInfo>> {
        let backups_dir = self.instances_path.join("..").join("backups");
        if !backups_dir.exists() {
            return Ok(vec![]);
        }

        let mut backups = Vec::new();
        let mut entries = fs::read_dir(&backups_dir).await?;

        while let Some(entry) = entries.next_entry().await? {
            let path = entry.path();
            if path.extension().and_then(|s| s.to_str()) == Some("json") {
                if let Ok(content) = fs::read_to_string(&path).await {
                    if let Ok(backup) = serde_json::from_str::<BackupInfo>(&content) {
                        if backup.instance_id == instance_id {
                            backups.push(backup);
                        }
                    }
                }
            }
        }

        // Sort by date (newest first)
        backups.sort_by(|a, b| b.backup_date.cmp(&a.backup_date));
        Ok(backups)
    }

    /// Delete a backup
    pub async fn delete_backup(&self, backup_id: &str) -> LauncherResult<()> {
        info!("Deleting backup: {}", backup_id);

        let backups_dir = self.instances_path.join("..").join("backups");
        let backup_path = backups_dir.join(format!("{}.zip", backup_id));
        let info_path = backups_dir.join(format!("{}.json", backup_id));

        if backup_path.exists() {
            fs::remove_file(&backup_path).await?;
        }
        if info_path.exists() {
            fs::remove_file(&info_path).await?;
        }

        info!("Backup deleted successfully");
        Ok(())
    }

    /// Check for updates for a specific instance
    pub async fn check_instance_updates(&self, instance_id: &str) -> LauncherResult<Option<ModpackUpdate>> {
        let _current_version = self.get_instance_version(instance_id).await?;
        let _pack_url = self.get_instance_pack_url(instance_id).await?;

        // This would need to check the remote version
        // For now, return None - would need to implement remote version checking
        Ok(None)
    }

    /// Get manual downloads for an installation
    pub async fn get_manual_downloads(&self, install_id: &str) -> LauncherResult<Vec<ManualDownload>> {
        let downloads = self.manual_downloads.read().await;
        Ok(downloads.get(install_id).cloned().unwrap_or_default())
    }

    /// Confirm manual download completion
    pub async fn confirm_manual_download(
        &self,
        install_id: &str,
        filename: &str,
        local_path: &str,
    ) -> LauncherResult<()> {
        info!("Confirming manual download: {} -> {}", filename, local_path);

        // Validate the file
        let path = Path::new(local_path);
        if !path.exists() {
            return Err(LauncherError::FileSystem(
                format!("File not found: {}", local_path)
            ));
        }

        // Update manual downloads status
        {
            let mut downloads = self.manual_downloads.write().await;
            if let Some(instance_downloads) = downloads.get_mut(install_id) {
                if let Some(index) = instance_downloads.iter().position(|d| d.filename == filename) {
                    instance_downloads.remove(index);
                }
            }
        }

        // If all manual downloads are confirmed, continue installation
        let remaining_downloads = {
            let downloads = self.manual_downloads.read().await;
            downloads.get(install_id).map(|d| d.len()).unwrap_or(0)
        };

        if remaining_downloads == 0 {
            info!("All manual downloads confirmed, continuing installation");
            // Continue with the installation process
            // This would resume the packwiz install
        }

        Ok(())
    }

    /// Get installation progress
    pub async fn get_install_progress(&self, install_id: &str) -> LauncherResult<Option<PackInstallProgress>> {
        let progress = self.progress.read().await;
        Ok(progress.get(install_id).cloned())
    }

    /// Cancel installation
    pub async fn cancel_installation(&self, install_id: &str) -> LauncherResult<()> {
        info!("Cancelling installation: {}", install_id);

        self.update_progress(install_id, PackInstallProgress {
            instance_id: "".to_string(),
            modpack_id: "".to_string(),
            step: PackInstallStep::Downloading,
            progress_percent: 0.0,
            message: "Installation cancelled".to_string(),
            status: PackInstallStatus::Cancelled,
        }).await;

        // Clean up any temporary files
        let temp_path = self.temp_path.join(install_id);
        if temp_path.exists() {
            fs::remove_dir_all(&temp_path).await?;
        }

        Ok(())
    }

    // Helper methods

    async fn update_progress(&self, install_id: &str, progress: PackInstallProgress) {
        let mut progress_map = self.progress.write().await;
        progress_map.insert(install_id.to_string(), progress);
    }

    async fn is_instance_installed(&self, instance_id: &str) -> bool {
        self.instances_path.join(instance_id).exists()
    }

    async fn get_instance_version(&self, instance_id: &str) -> LauncherResult<String> {
        let instance_path = self.instances_path.join(instance_id);
        let pack_toml = instance_path.join("pack.toml");

        if pack_toml.exists() {
            let content = fs::read_to_string(&pack_toml).await?;
            let toml_value: toml::Value = toml::from_str(&content)
                .map_err(|e| LauncherError::Serialization(
                    format!("Failed to parse pack.toml: {}", e)
                ))?;

            if let Some(version) = toml_value.get("version")
                .and_then(|v| v.as_str()) {
                return Ok(version.to_string());
            }
        }

        Err(LauncherError::InstanceNotFound(
            "Could not determine instance version".to_string()
        ))
    }

    async fn get_instance_pack_url(&self, instance_id: &str) -> LauncherResult<String> {
        let instance_path = self.instances_path.join(instance_id);
        let pack_toml = instance_path.join("pack.toml");

        if pack_toml.exists() {
            let content = fs::read_to_string(&pack_toml).await?;
            let toml_value: toml::Value = toml::from_str(&content)
                .map_err(|e| LauncherError::Serialization(
                    format!("Failed to parse pack.toml: {}", e)
                ))?;

            if let Some(download_url) = toml_value.get("download")
                .and_then(|v| v.get("url"))
                .and_then(|u| u.as_str()) {
                return Ok(download_url.to_string());
            }
        }

        Err(LauncherError::InstanceNotFound(
            "Could not determine pack URL".to_string()
        ))
    }

    async fn update_instance_info(&self, instance_id: &str, backup_info: &BackupInfo) -> LauncherResult<()> {
        let instance_path = self.instances_path.join(instance_id);
        let info_file = instance_path.join("instance_info.json");

        // This would update the instance information with backup data
        // For now, just create a basic info file
        let info = serde_json::json!({
            "id": instance_id,
            "restored_from_backup": backup_info.id,
            "restored_date": chrono::Utc::now().to_rfc3339(),
            "backup_version": backup_info.version
        });

        fs::write(info_file, serde_json::to_string_pretty(&info)?).await?;
        Ok(())
    }
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

/// Parse packwiz progress from output line
fn parse_packwiz_progress(line: &str) -> Option<(PackInstallStep, f64, String)> {
    // This would parse packwiz output to determine progress
    // For now, return None - would need to implement actual parsing

    if line.contains("Downloading") {
        Some((PackInstallStep::Downloading, 30.0, "Downloading files...".to_string()))
    } else if line.contains("Installing") {
        Some((PackInstallStep::InstallingFiles, 70.0, "Installing files...".to_string()))
    } else if line.contains("Complete") {
        Some((PackInstallStep::Completed, 100.0, "Installation complete".to_string()))
    } else {
        None
    }
}

/// Global packwiz manager instance
pub static PACKWIZ_MANAGER: std::sync::LazyLock<PackwizManager> =
    std::sync::LazyLock::new(|| {
        let instances_dir = dirs::home_dir()
            .unwrap_or_else(|| std::env::current_dir().unwrap())
            .join("TheBoysLauncher")
            .join("instances");

        PackwizManager::new(instances_dir)
    });

/// Get the global packwiz manager
pub fn packwiz_manager() -> &'static PackwizManager {
    &PACKWIZ_MANAGER
}