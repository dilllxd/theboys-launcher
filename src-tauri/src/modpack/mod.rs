use crate::models::{LauncherResult, Modpack, InstalledModpack, ModpackUpdate, Modloader, LauncherError};
use reqwest;
use serde_json;
use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::fs;
use tracing::{info, warn, error, debug};

/// Modpack manager for handling remote and local modpack operations
pub struct ModpackManager {
    client: reqwest::Client,
    cache_dir: PathBuf,
    remote_url: String,
}

impl ModpackManager {
    pub fn new<P: AsRef<Path>>(cache_dir: P, remote_url: String) -> Self {
        Self {
            client: reqwest::Client::builder()
                .user_agent("TheBoys-Launcher/1.1.0")
                .timeout(std::time::Duration::from_secs(30))
                .build()
                .unwrap_or_else(|e| {
                    error!("Failed to create HTTP client: {}", e);
                    reqwest::Client::new()
                }),
            cache_dir: cache_dir.as_ref().to_path_buf(),
            remote_url,
        }
    }

    /// Fetch available modpacks from remote URL with caching
    pub async fn get_modpacks(&self) -> LauncherResult<Vec<Modpack>> {
        info!("Fetching modpacks from: {}", self.remote_url);

        // Try to load from cache first
        if let Ok(cached_modpacks) = self.load_cached_modpacks().await {
            if !cached_modpacks.is_empty() {
                debug!("Loaded {} modpacks from cache", cached_modpacks.len());
                return Ok(cached_modpacks);
            }
        }

        // Fetch from remote
        match self.fetch_remote_modpacks().await {
            Ok(modpacks) => {
                // Cache the results
                if let Err(e) = self.cache_modpacks(&modpacks).await {
                    warn!("Failed to cache modpacks: {}", e);
                }
                info!("Successfully fetched {} modpacks from remote", modpacks.len());
                Ok(modpacks)
            }
            Err(e) => {
                warn!("Failed to fetch remote modpacks: {}", e);
                // Try to return cached modpacks as fallback
                match self.load_cached_modpacks().await {
                    Ok(cached_modpacks) if !cached_modpacks.is_empty() => {
                        info!("Returning cached modpacks as fallback");
                        Ok(cached_modpacks)
                    }
                    _ => Err(e)
                }
            }
        }
    }

    /// Fetch modpacks directly from remote URL
    async fn fetch_remote_modpacks(&self) -> LauncherResult<Vec<Modpack>> {
        let response = self.client
            .get(&self.remote_url)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(LauncherError::Network(
                format!("HTTP {}: {}", response.status(), response.status().canonical_reason().unwrap_or("Unknown"))
            ));
        }

        let text = response.text().await?;
        let modpacks: Vec<Modpack> = serde_json::from_str(&text)
            .map_err(|e| LauncherError::Serialization(format!("Failed to parse modpacks.json: {}", e)))?;

        if modpacks.is_empty() {
            return Err(LauncherError::ModpackNotFound("No modpacks found in remote configuration".to_string()));
        }

        Ok(modpacks)
    }

    /// Load modpacks from cache file
    async fn load_cached_modpacks(&self) -> LauncherResult<Vec<Modpack>> {
        let cache_file = self.cache_dir.join("modpacks.json");
        if !cache_file.exists() {
            return Ok(vec![]);
        }

        let content = fs::read_to_string(&cache_file).await?;
        let cached: serde_json::Value = serde_json::from_str(&content)?;

        // Check cache age (24 hours)
        if let Some(timestamp) = cached.get("timestamp").and_then(|v| v.as_u64()) {
            let now = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs();

            if now - timestamp > 86400 { // 24 hours
                debug!("Cache expired, ignoring");
                return Ok(vec![]);
            }
        }

        if let Some(modpacks) = cached.get("modpacks").and_then(|v| serde_json::from_value::<Vec<Modpack>>(v.clone()).ok()) {
            Ok(modpacks)
        } else {
            Ok(vec![])
        }
    }

    /// Cache modpacks to local file
    async fn cache_modpacks(&self, modpacks: &[Modpack]) -> LauncherResult<()> {
        fs::create_dir_all(&self.cache_dir).await?;

        let cache_file = self.cache_dir.join("modpacks.json");
        let cache_data = serde_json::json!({
            "timestamp": SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            "modpacks": modpacks
        });

        fs::write(cache_file, serde_json::to_string_pretty(&cache_data)?).await?;
        Ok(())
    }

    /// Get installed modpacks
    pub async fn get_installed_modpacks(&self) -> LauncherResult<Vec<InstalledModpack>> {
        let instances_dir = self.cache_dir.join("instances");
        if !instances_dir.exists() {
            return Ok(vec![]);
        }

        let mut installed_modpacks = Vec::new();
        let mut entries = fs::read_dir(&instances_dir).await?;

        while let Some(entry) = entries.next_entry().await? {
            let path = entry.path();
            if path.is_dir() {
                if let Ok(installed) = self.load_installed_modpack(&path).await {
                    installed_modpacks.push(installed);
                }
            }
        }

        Ok(installed_modpacks)
    }

    /// Load installed modpack information from instance directory
    async fn load_installed_modpack(&self, instance_path: &Path) -> LauncherResult<InstalledModpack> {
        let info_file = instance_path.join("instance_info.json");
        if !info_file.exists() {
            return Err(LauncherError::InstanceNotFound("Instance info file not found".to_string()));
        }

        let content = fs::read_to_string(&info_file).await?;
        let info: serde_json::Value = serde_json::from_str(&content)?;

        // Parse modpack info
        let modpack: Modpack = serde_json::from_value(info.get("modpack").cloned().unwrap_or_default())?;

        // Calculate directory size
        let size_bytes = self.calculate_directory_size(instance_path).await?;

        // Parse installation info
        let installed_version = info.get("installed_version")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string();

        let install_date = info.get("install_date")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string();

        let last_played = info.get("last_played")
            .and_then(|v| v.as_str())
            .map(|s| s.to_string());

        let total_playtime = info.get("total_playtime")
            .and_then(|v| v.as_u64())
            .unwrap_or(0);

        Ok(InstalledModpack {
            modpack,
            installed_version,
            install_path: instance_path.to_string_lossy().to_string(),
            install_date,
            last_played,
            total_playtime,
            update_available: false, // Will be updated by check_for_updates
            size_bytes,
        })
    }

    /// Calculate total size of a directory recursively
    async fn calculate_directory_size(&self, path: &Path) -> LauncherResult<u64> {
        let mut total_size = 0u64;
        let mut entries = fs::read_dir(path).await?;

        while let Some(entry) = entries.next_entry().await? {
            let entry_path = entry.path();
            if entry_path.is_dir() {
                total_size += Box::pin(self.calculate_directory_size(&entry_path)).await?;
            } else {
                total_size += entry.metadata().await?.len();
            }
        }

        Ok(total_size)
    }

    /// Check for updates for a specific modpack
    pub async fn check_modpack_updates(&self, modpack_id: &str) -> LauncherResult<Option<ModpackUpdate>> {
        // Get installed version
        let installed_modpacks = self.get_installed_modpacks().await?;
        let installed_modpack = installed_modpacks
            .iter()
            .find(|m| m.modpack.id == modpack_id);

        if let Some(installed) = installed_modpack {
            // Get latest version from remote
            let remote_modpacks = self.get_modpacks().await?;
            let remote_modpack = remote_modpacks
                .iter()
                .find(|m| m.id == modpack_id);

            if let Some(remote) = remote_modpack {
                let update_available = installed.installed_version != remote.version;

                Ok(Some(ModpackUpdate {
                    modpack_id: modpack_id.to_string(),
                    current_version: installed.installed_version.clone(),
                    latest_version: remote.version.clone(),
                    update_available,
                    changelog_url: None, // TODO: Implement changelog support
                    download_url: remote.pack_url.clone(),
                    size_bytes: installed.size_bytes,
                }))
            } else {
                Err(LauncherError::ModpackNotFound(format!("Modpack {} not found in remote", modpack_id)))
            }
        } else {
            Err(LauncherError::InstanceNotFound(format!("Modpack {} not installed", modpack_id)))
        }
    }

    /// Check for updates for all installed modpacks
    pub async fn check_all_updates(&self) -> LauncherResult<Vec<ModpackUpdate>> {
        let installed_modpacks = self.get_installed_modpacks().await?;
        let mut updates = Vec::new();

        for installed in installed_modpacks {
            if let Ok(Some(update)) = self.check_modpack_updates(&installed.modpack.id).await {
                if update.update_available {
                    updates.push(update);
                }
            }
        }

        Ok(updates)
    }

    /// Select a modpack as default
    pub async fn select_default_modpack(&self, modpack_id: &str) -> LauncherResult<()> {
        let modpacks = self.get_modpacks().await?;

        // Verify modpack exists
        if !modpacks.iter().any(|m| m.id == modpack_id) {
            return Err(LauncherError::ModpackNotFound(format!("Modpack {} not found", modpack_id)));
        }

        // Update cache with new default
        let mut updated_modpacks = modpacks;
        for modpack in &mut updated_modpacks {
            modpack.default = modpack.id == modpack_id;
        }

        self.cache_modpacks(&updated_modpacks).await?;
        info!("Set {} as default modpack", modpack_id);

        Ok(())
    }

    /// Get the default modpack
    pub async fn get_default_modpack(&self) -> LauncherResult<Option<Modpack>> {
        let modpacks = self.get_modpacks().await?;
        Ok(modpacks.into_iter().find(|m| m.default))
    }

    /// Get a specific modpack by ID
    pub async fn get_modpack(&self, modpack_id: &str) -> LauncherResult<Option<Modpack>> {
        let modpacks = self.get_modpacks().await?;
        Ok(modpacks.into_iter().find(|m| m.id == modpack_id))
    }

    /// Clear modpack cache
    pub async fn clear_cache(&self) -> LauncherResult<()> {
        let cache_file = self.cache_dir.join("modpacks.json");
        if cache_file.exists() {
            fs::remove_file(cache_file).await?;
        }
        info!("Modpack cache cleared");
        Ok(())
    }
}

/// Global modpack manager instance
pub static MODPACK_MANAGER: std::sync::LazyLock<ModpackManager> =
    std::sync::LazyLock::new(|| {
        let cache_dir = dirs::cache_dir()
            .unwrap_or_else(|| std::env::current_dir().unwrap().join("cache"))
            .join("theboys-launcher");

        let remote_url = std::env::var("THEBOYS_MODPACKS_URL")
            .unwrap_or_else(|_| "https://raw.githubusercontent.com/dilllxd/theboys-launcher/refs/heads/main/modpacks.json".to_string());

        ModpackManager::new(cache_dir, remote_url)
    });

/// Get the global modpack manager
pub fn modpack_manager() -> &'static ModpackManager {
    &MODPACK_MANAGER
}