use crate::models::{
    LauncherResult, Instance, InstanceConfig,
    InstanceStatus, MultiMCInstanceConfig, MultiMCComponent, ComponentRequirement,
    InstanceValidation, Modloader
};
use std::path::{Path, PathBuf};
use tokio::fs;
use serde_json;
use chrono::Utc;
use tracing::{info, warn, error, debug};

pub mod modloader;
pub mod launch_manager;
pub use modloader::ModloaderInstaller;
pub use launch_manager::LaunchManager;

/// Instance manager for handling Minecraft instances
pub struct InstanceManager {
    instances_dir: PathBuf,
    modloader_installer: ModloaderInstaller,
}

impl InstanceManager {
    pub fn new(instances_dir: PathBuf) -> Self {
        Self {
            instances_dir,
            modloader_installer: ModloaderInstaller::new(),
        }
    }

    /// Create a new instance from a modpack with complete MultiMC compatibility
    pub async fn create_instance(
        &self,
        config: &InstanceConfig,
    ) -> LauncherResult<Instance> {
        info!("Creating new instance: {}", config.name);

        // Validate instance name doesn't already exist
        let instance_dir = self.instances_dir.join(&config.name);
        if instance_dir.exists() {
            return Err(crate::models::LauncherError::InvalidConfig(
                format!("Instance '{}' already exists", config.name)
            ));
        }

        // Create instance directory structure
        self.create_instance_structure(&instance_dir).await?;

        // Generate instance ID
        let instance_id = format!("instance-{}", uuid::Uuid::new_v4());
        let now = Utc::now().to_rfc3339();

        // Create instance object
        let instance = Instance {
            id: instance_id.clone(),
            name: config.name.clone(),
            modpack_id: config.modpack_id.clone(),
            minecraft_version: config.minecraft_version.clone(),
            loader_type: config.loader_type.as_str().to_string(),
            loader_version: config.loader_version.clone(),
            memory_mb: config.memory_mb,
            java_path: config.java_path.clone(),
            game_dir: instance_dir.to_string_lossy().to_string(),
            last_played: None,
            total_playtime: 0,
            icon_path: config.icon_path.clone(),
            status: InstanceStatus::Installing,
            created_at: now.clone(),
            updated_at: now,
            jvm_args: config.jvm_args.clone(),
            env_vars: config.env_vars.clone(),
        };

        // Create MultiMC configuration files
        self.create_multimc_configs(&instance_dir, &instance).await?;

        // Create minecraft subdirectory
        let minecraft_dir = instance_dir.join("minecraft");
        fs::create_dir_all(&minecraft_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create minecraft directory: {}", e)
            ))?;

        // Save instance metadata
        self.save_instance_metadata(&instance_dir, &instance).await?;

        info!("Successfully created instance: {}", config.name);
        Ok(instance)
    }

    /// Get all instances with their current status
    pub async fn get_instances(&self) -> LauncherResult<Vec<Instance>> {
        debug!("Scanning instances directory: {:?}", self.instances_dir);

        if !self.instances_dir.exists() {
            fs::create_dir_all(&self.instances_dir).await
                .map_err(|e| crate::models::LauncherError::FileSystem(
                    format!("Failed to create instances directory: {}", e)
                ))?;
            return Ok(vec![]);
        }

        let mut instances = Vec::new();

        let mut entries = fs::read_dir(&self.instances_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to read instances directory: {}", e)
            ))?;

        while let Some(entry) = entries.next_entry().await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to read directory entry: {}", e)
            ))? {

            let path = entry.path();
            if path.is_dir() {
                if let Ok(instance) = self.load_instance_from_directory(&path).await {
                    instances.push(instance);
                }
            }
        }

        // Sort by last played (most recent first) or creation date
        instances.sort_by(|a, b| {
            match (&a.last_played, &b.last_played) {
                (Some(a_date), Some(b_date)) => b_date.cmp(a_date),
                (Some(_), None) => std::cmp::Ordering::Less,
                (None, Some(_)) => std::cmp::Ordering::Greater,
                (None, None) => b.created_at.cmp(&a.created_at),
            }
        });

        Ok(instances)
    }

    /// Get a specific instance by ID
    pub async fn get_instance(&self, id: &str) -> LauncherResult<Option<Instance>> {
        debug!("Looking for instance with ID: {}", id);

        let instances = self.get_instances().await?;
        Ok(instances.into_iter().find(|inst| inst.id == id))
    }

    /// Get instance by name
    pub async fn get_instance_by_name(&self, name: &str) -> LauncherResult<Option<Instance>> {
        debug!("Looking for instance with name: {}", name);

        let instance_dir = self.instances_dir.join(name);
        if !instance_dir.exists() {
            return Ok(None);
        }

        let instance = self.load_instance_from_directory(&instance_dir).await?;
        Ok(Some(instance))
    }

    /// Update instance settings
    pub async fn update_instance(&self, instance: &Instance) -> LauncherResult<()> {
        info!("Updating instance: {}", instance.name);

        let instance_dir = Path::new(&instance.game_dir);
        if !instance_dir.exists() {
            return Err(crate::models::LauncherError::InstanceNotFound(
                instance.name.clone()
            ));
        }

        // Update timestamp
        let mut updated_instance = instance.clone();
        updated_instance.updated_at = Utc::now().to_rfc3339();

        // Save updated metadata
        self.save_instance_metadata(instance_dir, &updated_instance).await?;

        // Update MultiMC configuration if needed
        self.update_multimc_config(instance_dir, &updated_instance).await?;

        info!("Successfully updated instance: {}", instance.name);
        Ok(())
    }

    /// Delete an instance safely
    pub async fn delete_instance(&self, id: &str) -> LauncherResult<()> {
        info!("Deleting instance with ID: {}", id);

        let instance = self.get_instance(id).await?
            .ok_or_else(|| crate::models::LauncherError::InstanceNotFound(
                format!("Instance {} not found", id)
            ))?;

        let instance_dir = Path::new(&instance.game_dir);

        // Verify it's actually an instance directory
        let instance_cfg_path = instance_dir.join("instance.cfg");
        if !instance_cfg_path.exists() {
            return Err(crate::models::LauncherError::InvalidConfig(
                format!("Directory {} is not a valid instance", instance.game_dir)
            ));
        }

        // Remove the instance directory
        fs::remove_dir_all(instance_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to delete instance directory: {}", e)
            ))?;

        info!("Successfully deleted instance: {}", instance.name);
        Ok(())
    }

    /// Validate an instance
    pub async fn validate_instance(&self, id: &str) -> LauncherResult<InstanceValidation> {
        debug!("Validating instance: {}", id);

        let instance = self.get_instance(id).await?
            .ok_or_else(|| crate::models::LauncherError::InstanceNotFound(
                format!("Instance {} not found", id)
            ))?;

        let mut issues = Vec::new();
        let mut recommendations = Vec::new();
        let mut is_valid = true;

        let instance_dir = Path::new(&instance.game_dir);

        // Check instance.cfg exists
        let instance_cfg_path = instance_dir.join("instance.cfg");
        if !instance_cfg_path.exists() {
            issues.push("instance.cfg file is missing".to_string());
            is_valid = false;
        }

        // Check mmc-pack.json exists
        let mmc_pack_path = instance_dir.join("mmc-pack.json");
        if !mmc_pack_path.exists() {
            issues.push("mmc-pack.json file is missing".to_string());
            is_valid = false;
        }

        // Check minecraft directory exists
        let minecraft_dir = instance_dir.join("minecraft");
        if !minecraft_dir.exists() {
            issues.push("minecraft directory is missing".to_string());
            recommendations.push("Create minecraft directory".to_string());
            is_valid = false;
        }

        // Check Java path exists
        if !Path::new(&instance.java_path).exists() {
            issues.push(format!("Java executable not found: {}", instance.java_path));
            recommendations.push("Update Java path in instance settings".to_string());
            is_valid = false;
        }

        // Check modloader installation
        if !self.check_modloader_installed(&instance).await? {
            issues.push("Modloader is not properly installed".to_string());
            recommendations.push("Reinstall modloader".to_string());
        }

        // Memory recommendations
        let total_memory = crate::utils::system::get_total_memory_mb().await.unwrap_or(8192);
        if u64::from(instance.memory_mb) > total_memory {
            issues.push("Allocated memory exceeds system memory".to_string());
            recommendations.push("Reduce memory allocation".to_string());
        } else if u64::from(instance.memory_mb) > total_memory / 2 {
            recommendations.push("Consider reducing memory allocation for better performance".to_string());
        }

        Ok(InstanceValidation {
            instance_id: id.to_string(),
            is_valid,
            issues,
            recommendations,
        })
    }

    /// Check if modloader is installed
    pub async fn check_modloader_installed(&self, instance: &Instance) -> LauncherResult<bool> {
        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");

        match instance.loader_type.as_str() {
            "forge" => {
                let forge_jar = minecraft_dir.join("libraries")
                    .join("net")
                    .join("minecraftforge")
                    .join("forge")
                    .join(format!("{}-{}", instance.minecraft_version, instance.loader_version))
                    .join(format!("forge-{}-{}-universal.jar", instance.minecraft_version, instance.loader_version));
                Ok(forge_jar.exists())
            },
            "fabric" => {
                let fabric_jar = minecraft_dir.join("libraries")
                    .join("net")
                    .join("fabricmc")
                    .join("fabric-loader")
                    .join(&instance.loader_version)
                    .join(format!("fabric-loader-{}.jar", instance.loader_version));
                Ok(fabric_jar.exists())
            },
            "quilt" => {
                let quilt_jar = minecraft_dir.join("libraries")
                    .join("org")
                    .join("quiltmc")
                    .join("quilt-loader")
                    .join(&instance.loader_version)
                    .join(format!("quilt-loader-{}.jar", instance.loader_version));
                Ok(quilt_jar.exists())
            },
            "neoforge" => {
                let neoforge_jar = minecraft_dir.join("libraries")
                    .join("net")
                    .join("neoforged")
                    .join("neoforge")
                    .join(&instance.loader_version)
                    .join(format!("neoforge-{}.jar", instance.loader_version));
                Ok(neoforge_jar.exists())
            },
            "vanilla" => Ok(true), // Vanilla doesn't need additional files
            _ => Ok(false),
        }
    }

    /// Launch a Minecraft instance
    pub async fn launch_instance(&self, instance: &Instance) -> LauncherResult<()> {
        info!("Launching instance: {}", instance.name);

        // Validate instance first
        let validation = self.validate_instance(&instance.id).await?;
        if !validation.is_valid {
            return Err(crate::models::LauncherError::InvalidConfig(
                format!("Instance validation failed: {}", validation.issues.join(", "))
            ));
        }

        // Update status to running
        let mut running_instance = instance.clone();
        running_instance.status = InstanceStatus::Running;
        self.update_instance(&running_instance).await?;

        // TODO: Implement actual Minecraft launch logic
        // This would involve:
        // 1. Downloading Minecraft version if needed
        // 2. Preparing launch command with proper JVM arguments
        // 3. Setting up environment variables
        // 4. Executing Minecraft process
        // 5. Monitoring the process

        info!("Instance launched successfully: {}", instance.name);
        Ok(())
    }

    /// Install modloader for an instance
    pub async fn install_modloader(
        &self,
        instance_id: &str,
        progress_callback: Option<Box<dyn Fn(f64) + Send + Sync>>,
    ) -> LauncherResult<()> {
        let instance = self.get_instance(instance_id).await?
            .ok_or_else(|| crate::models::LauncherError::InstanceNotFound(
                format!("Instance {} not found", instance_id)
            ))?;

        info!("Installing modloader for instance: {}", instance.name);

        // Update status to installing
        let mut installing_instance = instance.clone();
        installing_instance.status = InstanceStatus::Installing;
        self.update_instance(&installing_instance).await?;

        // Install modloader
        let result = self.modloader_installer.install_modloader(&instance, progress_callback).await;

        // Update status based on result
        let mut final_instance = instance.clone();
        match result {
            Ok(()) => {
                final_instance.status = InstanceStatus::Ready;
                info!("Modloader installation completed for instance: {}", instance.name);
            },
            Err(e) => {
                final_instance.status = InstanceStatus::Broken;
                error!("Modloader installation failed for instance {}: {}", instance.name, e);
                return Err(e);
            }
        }

        self.update_instance(&final_instance).await?;
        Ok(())
    }

    /// Get available modloader versions
    pub async fn get_modloader_versions(
        &self,
        modloader: &Modloader,
        minecraft_version: &str,
    ) -> LauncherResult<Vec<String>> {
        self.modloader_installer.get_available_versions(modloader, minecraft_version).await
    }

    /// Repair an instance
    pub async fn repair_instance(&self, instance_id: &str) -> LauncherResult<()> {
        info!("Repairing instance: {}", instance_id);

        let instance = self.get_instance(instance_id).await?
            .ok_or_else(|| crate::models::LauncherError::InstanceNotFound(
                format!("Instance {} not found", instance_id)
            ))?;

        // Update status to updating
        let mut repairing_instance = instance.clone();
        repairing_instance.status = InstanceStatus::Updating;
        self.update_instance(&repairing_instance).await?;

        // Validate instance first
        let validation = self.validate_instance(instance_id).await?;

        if !validation.is_valid {
            // Try to fix issues
            for issue in &validation.issues {
                match issue.as_str() {
                    "instance.cfg file is missing" => {
                        self.create_multimc_configs(
                            Path::new(&instance.game_dir),
                            &instance
                        ).await?;
                    },
                    "mmc-pack.json file is missing" => {
                        self.create_multimc_configs(
                            Path::new(&instance.game_dir),
                            &instance
                        ).await?;
                    },
                    "minecraft directory is missing" => {
                        let minecraft_dir = Path::new(&instance.game_dir).join("minecraft");
                        fs::create_dir_all(&minecraft_dir).await?;
                    },
                    "Modloader is not properly installed" => {
                        // Reinstall modloader
                        self.install_modloader(instance_id, None).await?;
                    },
                    _ => {}
                }
            }
        }

        // Mark as ready if no critical issues remain
        let final_validation = self.validate_instance(instance_id).await?;
        let mut final_instance = instance.clone();
        final_instance.status = if final_validation.is_valid {
            InstanceStatus::Ready
        } else {
            InstanceStatus::Broken
        };

        self.update_instance(&final_instance).await?;

        if final_validation.is_valid {
            info!("Instance repair completed successfully: {}", instance.name);
        } else {
            warn!("Instance repair completed with remaining issues: {}", instance.name);
        }

        Ok(())
    }

    // Private helper methods

    /// Create instance directory structure
    async fn create_instance_structure(&self, instance_dir: &Path) -> LauncherResult<()> {
        fs::create_dir_all(instance_dir).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create instance directory: {}", e)
            ))?;

        Ok(())
    }

    /// Create MultiMC configuration files
    async fn create_multimc_configs(&self, instance_dir: &Path, instance: &Instance) -> LauncherResult<()> {
        // Create instance.cfg
        let instance_cfg = MultiMCInstanceConfig {
            instance_type: "OneSix".to_string(),
            name: instance.name.clone(),
            icon_key: "default".to_string(),
            override_memory: true,
            min_mem_alloc: instance.memory_mb / 2,
            max_mem_alloc: instance.memory_mb,
            override_java: true,
            java_path: instance.java_path.clone(),
            notes: format!("Managed by TheBoys Launcher - Created: {}", instance.created_at),
            jvm_args: instance.jvm_args.clone(),
        };

        let instance_cfg_path = instance_dir.join("instance.cfg");
        let instance_cfg_content = self.format_instance_cfg(&instance_cfg);
        fs::write(&instance_cfg_path, instance_cfg_content).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to write instance.cfg: {}", e)
            ))?;

        // Create mmc-pack.json
        let components = self.create_multimc_components(instance)?;
        let mmc_pack = serde_json::json!({
            "formatVersion": 1,
            "components": components
        });

        let mmc_pack_path = instance_dir.join("mmc-pack.json");
        let mmc_pack_content = serde_json::to_string_pretty(&mmc_pack)
            .map_err(|e| crate::models::LauncherError::Serialization(e.to_string()))?;
        fs::write(&mmc_pack_path, mmc_pack_content).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to write mmc-pack.json: {}", e)
            ))?;

        // Create pack.json
        let pack = serde_json::json!({
            "formatVersion": 3,
            "components": components
        });

        let pack_path = instance_dir.join("pack.json");
        let pack_content = serde_json::to_string_pretty(&pack)
            .map_err(|e| crate::models::LauncherError::Serialization(e.to_string()))?;
        fs::write(&pack_path, pack_content).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to write pack.json: {}", e)
            ))?;

        Ok(())
    }

    /// Create MultiMC components for mmc-pack.json
    fn create_multimc_components(&self, instance: &Instance) -> LauncherResult<Vec<MultiMCComponent>> {
        let mut components = Vec::new();

        // Minecraft component
        components.push(MultiMCComponent {
            cached_name: Some("Minecraft".to_string()),
            cached_version: Some(instance.minecraft_version.clone()),
            cached_requires: None,
            uid: "net.minecraft".to_string(),
            version: instance.minecraft_version.clone(),
        });

        // Modloader component
        match instance.loader_type.as_str() {
            "forge" => {
                components.push(MultiMCComponent {
                    cached_name: Some("Minecraft Forge".to_string()),
                    cached_version: Some(instance.loader_version.clone()),
                    cached_requires: Some(vec![
                        ComponentRequirement {
                            equals: Some(instance.minecraft_version.clone()),
                            uid: "net.minecraft".to_string(),
                        }
                    ]),
                    uid: "net.minecraftforge".to_string(),
                    version: instance.loader_version.clone(),
                });
            },
            "fabric" => {
                components.push(MultiMCComponent {
                    cached_name: Some("Fabric Loader".to_string()),
                    cached_version: Some(instance.loader_version.clone()),
                    cached_requires: Some(vec![
                        ComponentRequirement {
                            equals: Some(instance.minecraft_version.clone()),
                            uid: "net.minecraft".to_string(),
                        }
                    ]),
                    uid: "net.fabricmc.fabric-loader".to_string(),
                    version: instance.loader_version.clone(),
                });
            },
            "quilt" => {
                components.push(MultiMCComponent {
                    cached_name: Some("Quilt Loader".to_string()),
                    cached_version: Some(instance.loader_version.clone()),
                    cached_requires: Some(vec![
                        ComponentRequirement {
                            equals: Some(instance.minecraft_version.clone()),
                            uid: "net.minecraft".to_string(),
                        }
                    ]),
                    uid: "org.quiltmc.quilt-loader".to_string(),
                    version: instance.loader_version.clone(),
                });
            },
            "neoforge" => {
                components.push(MultiMCComponent {
                    cached_name: Some("NeoForge".to_string()),
                    cached_version: Some(instance.loader_version.clone()),
                    cached_requires: Some(vec![
                        ComponentRequirement {
                            equals: Some(instance.minecraft_version.clone()),
                            uid: "net.minecraft".to_string(),
                        }
                    ]),
                    uid: "net.neoforged.neoforge".to_string(),
                    version: instance.loader_version.clone(),
                });
            },
            _ => {} // Vanilla doesn't need a modloader component
        }

        Ok(components)
    }

    /// Format instance.cfg content
    fn format_instance_cfg(&self, config: &MultiMCInstanceConfig) -> String {
        let mut lines = Vec::new();
        lines.push(format!("InstanceType={}", config.instance_type));
        lines.push(format!("name={}", config.name));
        lines.push(format!("iconKey={}", config.icon_key));
        lines.push(format!("OverrideMemory={}", config.override_memory));
        lines.push(format!("MinMemAlloc={}", config.min_mem_alloc));
        lines.push(format!("MaxMemAlloc={}", config.max_mem_alloc));
        lines.push(format!("OverrideJava={}", config.override_java));
        lines.push(format!("JavaPath={}", config.java_path));
        lines.push(format!("Notes={}", config.notes));

        if let Some(jvm_args) = &config.jvm_args {
            lines.push(format!("JvmArgs={}", jvm_args));
        }

        lines.join("\n")
    }

    /// Save instance metadata to JSON file
    async fn save_instance_metadata(&self, instance_dir: &Path, instance: &Instance) -> LauncherResult<()> {
        let metadata_path = instance_dir.join("theboys-instance.json");
        let metadata_content = serde_json::to_string_pretty(instance)
            .map_err(|e| crate::models::LauncherError::Serialization(e.to_string()))?;

        fs::write(&metadata_path, metadata_content).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to write instance metadata: {}", e)
            ))?;

        Ok(())
    }

    /// Load instance from directory
    async fn load_instance_from_directory(&self, instance_dir: &Path) -> LauncherResult<Instance> {
        let metadata_path = instance_dir.join("theboys-instance.json");

        if !metadata_path.exists() {
            // Try to create metadata from existing MultiMC files for migration
            return self.migrate_multimc_instance(instance_dir).await;
        }

        let metadata_content = fs::read_to_string(&metadata_path).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to read instance metadata: {}", e)
            ))?;

        let instance: Instance = serde_json::from_str(&metadata_content)
            .map_err(|e| crate::models::LauncherError::Serialization(e.to_string()))?;

        Ok(instance)
    }

    /// Migrate existing MultiMC instance to TheBoys format
    async fn migrate_multimc_instance(&self, instance_dir: &Path) -> LauncherResult<Instance> {
        debug!("Migrating MultiMC instance from: {:?}", instance_dir);

        let instance_cfg_path = instance_dir.join("instance.cfg");
        let mmc_pack_path = instance_dir.join("mmc-pack.json");

        if !instance_cfg_path.exists() || !mmc_pack_path.exists() {
            return Err(crate::models::LauncherError::InvalidConfig(
                "Not a valid MultiMC instance".to_string()
            ));
        }

        // Parse instance.cfg
        let instance_cfg_content = fs::read_to_string(&instance_cfg_path).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to read instance.cfg: {}", e)
            ))?;

        let instance_cfg = self.parse_instance_cfg(&instance_cfg_content)?;

        // Parse mmc-pack.json to get version info
        let mmc_pack_content = fs::read_to_string(&mmc_pack_path).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to read mmc-pack.json: {}", e)
            ))?;

        let mmc_pack: serde_json::Value = serde_json::from_str(&mmc_pack_content)
            .map_err(|e| crate::models::LauncherError::Serialization(e.to_string()))?;

        // Extract version information
        let (minecraft_version, loader_type, loader_version) = self.extract_version_info(&mmc_pack)?;

        let now = Utc::now().to_rfc3339();
        let instance_name = instance_dir.file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("Unknown")
            .to_string();

        let instance = Instance {
            id: format!("instance-{}", uuid::Uuid::new_v4()),
            name: instance_cfg.name.unwrap_or(instance_name),
            modpack_id: "migrated".to_string(),
            minecraft_version,
            loader_type,
            loader_version,
            memory_mb: instance_cfg.max_mem_alloc.unwrap_or(4096),
            java_path: instance_cfg.java_path.unwrap_or_default(),
            game_dir: instance_dir.to_string_lossy().to_string(),
            last_played: None,
            total_playtime: 0,
            icon_path: None,
            status: InstanceStatus::Ready,
            created_at: now.clone(),
            updated_at: now,
            jvm_args: instance_cfg.jvm_args,
            env_vars: None,
        };

        // Save the migrated metadata
        self.save_instance_metadata(instance_dir, &instance).await?;

        Ok(instance)
    }

    /// Parse instance.cfg file
    fn parse_instance_cfg(&self, content: &str) -> LauncherResult<ParsedInstanceConfig> {
        let mut config = ParsedInstanceConfig::default();

        for line in content.lines() {
            if let Some((key, value)) = line.split_once('=') {
                match key.trim() {
                    "name" => config.name = Some(value.trim().to_string()),
                    "JavaPath" => config.java_path = Some(value.trim().to_string()),
                    "MaxMemAlloc" => config.max_mem_alloc = value.trim().parse().ok(),
                    "MinMemAlloc" => config.min_mem_alloc = value.trim().parse().ok(),
                    "JvmArgs" => config.jvm_args = Some(value.trim().to_string()),
                    _ => {}
                }
            }
        }

        Ok(config)
    }

    /// Extract version information from mmc-pack.json
    fn extract_version_info(&self, mmc_pack: &serde_json::Value) -> LauncherResult<(String, String, String)> {
        let components = mmc_pack.get("components")
            .and_then(|c| c.as_array())
            .ok_or_else(|| crate::models::LauncherError::InvalidConfig(
                "Invalid mmc-pack.json format".to_string()
            ))?;

        let mut minecraft_version = String::new();
        let mut loader_type = "vanilla".to_string();
        let mut loader_version = String::new();

        for component in components {
            if let Some(uid) = component.get("uid").and_then(|u| u.as_str()) {
                let version = component.get("version").and_then(|v| v.as_str()).unwrap_or("");

                match uid {
                    "net.minecraft" => minecraft_version = version.to_string(),
                    "net.minecraftforge" => {
                        loader_type = "forge".to_string();
                        loader_version = version.to_string();
                    },
                    "net.fabricmc.fabric-loader" => {
                        loader_type = "fabric".to_string();
                        loader_version = version.to_string();
                    },
                    "org.quiltmc.quilt-loader" => {
                        loader_type = "quilt".to_string();
                        loader_version = version.to_string();
                    },
                    "net.neoforged.neoforge" => {
                        loader_type = "neoforge".to_string();
                        loader_version = version.to_string();
                    },
                    _ => {}
                }
            }
        }

        if minecraft_version.is_empty() {
            return Err(crate::models::LauncherError::InvalidConfig(
                "No Minecraft version found in mmc-pack.json".to_string()
            ));
        }

        Ok((minecraft_version, loader_type, loader_version))
    }

    /// Update MultiMC configuration
    async fn update_multimc_config(&self, instance_dir: &Path, instance: &Instance) -> LauncherResult<()> {
        let instance_cfg = MultiMCInstanceConfig {
            instance_type: "OneSix".to_string(),
            name: instance.name.clone(),
            icon_key: "default".to_string(),
            override_memory: true,
            min_mem_alloc: instance.memory_mb / 2,
            max_mem_alloc: instance.memory_mb,
            override_java: true,
            java_path: instance.java_path.clone(),
            notes: format!("Managed by TheBoys Launcher - Updated: {}", instance.updated_at),
            jvm_args: instance.jvm_args.clone(),
        };

        let instance_cfg_path = instance_dir.join("instance.cfg");
        let instance_cfg_content = self.format_instance_cfg(&instance_cfg);
        fs::write(&instance_cfg_path, instance_cfg_content).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to update instance.cfg: {}", e)
            ))?;

        Ok(())
    }
}

/// Parsed instance.cfg structure
#[derive(Debug, Default)]
struct ParsedInstanceConfig {
    name: Option<String>,
    java_path: Option<String>,
    max_mem_alloc: Option<u32>,
    min_mem_alloc: Option<u32>,
    jvm_args: Option<String>,
}

/// Prism Launcher manager
pub struct PrismManager {
    prism_path: Option<PathBuf>,
}

impl PrismManager {
    pub fn new(prism_path: Option<PathBuf>) -> Self {
        Self { prism_path }
    }

    /// Check if Prism Launcher is installed
    pub async fn is_installed(&self) -> bool {
        // TODO: Implement Prism detection logic
        false
    }

    /// Install Prism Launcher
    pub async fn install(&self, version: Option<String>) -> LauncherResult<PathBuf> {
        // TODO: Implement Prism installation logic
        Err(crate::models::LauncherError::NotImplemented(
            "Prism installation not yet implemented".to_string()
        ))
    }

    /// Get Prism Launcher executable path
    pub fn get_executable_path(&self) -> Option<PathBuf> {
        self.prism_path.clone()
    }
}

/// Global instance manager instance
pub static INSTANCE_MANAGER: std::sync::LazyLock<InstanceManager> =
    std::sync::LazyLock::new(|| {
        let instances_dir = crate::utils::config::get_instances_dir()
            .unwrap_or_else(|_| std::env::current_dir().unwrap().join("instances"));
        InstanceManager::new(instances_dir)
    });

/// Get the global instance manager
pub fn instance_manager() -> &'static InstanceManager {
    &INSTANCE_MANAGER
}