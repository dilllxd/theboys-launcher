#[cfg(test)]
mod tests {
    use std::path::PathBuf;
    use tempfile::TempDir;
    use serde_json;
    use theboys_launcher::utils::config::*;
    use theboys_launcher::models::{LauncherSettings, LauncherError};
    use tokio::fs;

    #[tokio::test]
    async fn test_load_settings_default() -> Result<(), Box<dyn std::error::Error>> {
        // Test loading settings when file doesn't exist (should return defaults)
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        let settings = load_settings_from_dir(&config_dir).await?;

        // Should return default settings
        assert_eq!(settings.memory_mb, 4096); // Default memory
        assert_eq!(settings.theme, "dark"); // Default theme
        assert!(settings.instances_dir.is_some());

        Ok(())
    }

    #[tokio::test]
    async fn test_save_and_load_settings() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        // Create custom settings
        let original_settings = LauncherSettings {
            memory_mb: 8192,
            theme: "light".to_string(),
            java_path: Some("/path/to/java".to_string()),
            prism_path: Some("/path/to/prism".to_string()),
            instances_dir: Some("/custom/instances".to_string()),
            auto_update: true,
            max_concurrent_downloads: 3,
            ..Default::default()
        };

        // Save settings
        save_settings_to_dir(&original_settings, &config_dir).await?;

        // Load settings back
        let loaded_settings = load_settings_from_dir(&config_dir).await?;

        // Verify settings match
        assert_eq!(original_settings.memory_mb, loaded_settings.memory_mb);
        assert_eq!(original_settings.theme, loaded_settings.theme);
        assert_eq!(original_settings.java_path, loaded_settings.java_path);
        assert_eq!(original_settings.prism_path, loaded_settings.prism_path);
        assert_eq!(original_settings.instances_dir, loaded_settings.instances_dir);
        assert_eq!(original_settings.auto_update, loaded_settings.auto_update);
        assert_eq!(original_settings.max_concurrent_downloads, loaded_settings.max_concurrent_downloads);

        Ok(())
    }

    #[tokio::test]
    async fn test_validate_settings_valid() -> Result<(), Box<dyn std::error::Error>> {
        let valid_settings = LauncherSettings {
            memory_mb: 4096,
            theme: "dark".to_string(),
            java_path: Some("/usr/bin/java".to_string()),
            prism_path: Some("/usr/bin/prism".to_string()),
            instances_dir: Some("/valid/path".to_string()),
            auto_update: true,
            max_concurrent_downloads: 3,
            ..Default::default()
        };

        // Should not panic on valid settings
        let result = validate_settings(&valid_settings);
        assert!(result.is_ok());

        Ok(())
    }

    #[tokio::test]
    fn test_validate_settings_invalid_memory() {
        let test_cases = vec![
            (0, "Memory too low"),
            (1023, "Memory just below minimum"),
            (32769, "Memory just above maximum"),
            (100000, "Memory way too high"),
        ];

        for (memory_mb, description) in test_cases {
            let invalid_settings = LauncherSettings {
                memory_mb,
                ..Default::default()
            };

            let result = validate_settings(&invalid_settings);
            assert!(result.is_err(), "Should fail: {}", description);

            match result.unwrap_err() {
                LauncherError::InvalidConfig(_) => {}, // Expected
                other => panic!("Expected InvalidConfig error, got: {:?}", other),
            }
        }
    }

    #[tokio::test]
    fn test_validate_settings_invalid_theme() {
        let invalid_themes = vec!["", "invalid", "DARK", "Light", "system-dark"];

        for theme in invalid_themes {
            let invalid_settings = LauncherSettings {
                theme: theme.to_string(),
                ..Default::default()
            };

            let result = validate_settings(&invalid_settings);
            assert!(result.is_err(), "Should fail for theme: {}", theme);

            match result.unwrap_err() {
                LauncherError::InvalidConfig(_) => {}, // Expected
                other => panic!("Expected InvalidConfig error for theme '{}', got: {:?}", theme, other),
            }
        }
    }

    #[tokio::test]
    fn test_validate_settings_invalid_downloads() {
        let invalid_counts = vec![0, 11, 50];

        for max_downloads in invalid_counts {
            let invalid_settings = LauncherSettings {
                max_concurrent_downloads: max_downloads,
                ..Default::default()
            };

            let result = validate_settings(&invalid_settings);
            assert!(result.is_err(), "Should fail for max_downloads: {}", max_downloads);

            match result.unwrap_err() {
                LauncherError::InvalidConfig(_) => {}, // Expected
                other => panic!("Expected InvalidConfig error for max_downloads {}, got: {:?}", max_downloads, other),
            }
        }
    }

    #[tokio::test]
    async fn test_load_settings_corrupted_file() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        // Create corrupted JSON file
        let config_file = config_dir.join("settings.json");
        fs::create_dir_all(&config_dir).await?;
        fs::write(&config_file, "{ invalid json content").await?;

        // Should return default settings when file is corrupted
        let settings = load_settings_from_dir(&config_dir).await?;

        // Should return default settings
        assert_eq!(settings.memory_mb, 4096);
        assert_eq!(settings.theme, "dark");

        Ok(())
    }

    #[tokio::test]
    async fn test_get_default_config_dir() {
        let config_dir = get_default_config_dir();

        // Should return a valid path
        assert!(!config_dir.as_os_str().is_empty());

        // Should include the app name
        let path_str = config_dir.to_string_lossy();
        assert!(path_str.contains("TheBoysLauncher") || path_str.contains("theboys-launcher"));
    }

    #[tokio::test]
    async fn test_get_default_instances_dir() {
        let instances_dir = get_default_instances_dir();

        // Should return a valid path
        assert!(!instances_dir.as_os_str().is_empty());

        // Should include the app name
        let path_str = instances_dir.to_string_lossy();
        assert!(path_str.contains("TheBoysLauncher") || path_str.contains("theboys-launcher"));
        assert!(path_str.contains("instances"));
    }

    #[tokio::test]
    async fn test_reset_settings() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        // Save custom settings first
        let custom_settings = LauncherSettings {
            memory_mb: 8192,
            theme: "light".to_string(),
            ..Default::default()
        };

        save_settings_to_dir(&custom_settings, &config_dir).await?;

        // Reset to defaults
        let default_settings = reset_settings_to_dir(&config_dir).await?;

        // Verify defaults are restored
        assert_eq!(default_settings.memory_mb, 4096);
        assert_eq!(default_settings.theme, "dark");

        // Verify settings are saved to disk
        let loaded_settings = load_settings_from_dir(&config_dir).await?;
        assert_eq!(default_settings.memory_mb, loaded_settings.memory_mb);
        assert_eq!(default_settings.theme, loaded_settings.theme);

        Ok(())
    }

    #[tokio::test]
    async fn test_migrate_settings_from_v1() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        // Create old v1 settings format
        let old_settings = serde_json::json!({
            "memory": 2048,           // Old field name
            "theme": "light",
            "java_path": "/old/java",
            "download_dir": "/old/downloads"  // Old field name
        });

        let config_file = config_dir.join("settings.json");
        fs::create_dir_all(&config_dir).await?;
        fs::write(&config_file, old_settings.to_string()).await?;

        // Migrate settings
        let migrated_settings = load_settings_from_dir(&config_dir).await?;

        // Verify migration worked
        assert_eq!(migrated_settings.memory_mb, 2048);
        assert_eq!(migrated_settings.theme, "light");
        assert_eq!(migrated_settings.java_path, Some("/old/java".to_string()));

        // Verify file was updated to new format
        let content = fs::read_to_string(&config_file).await?;
        let parsed: serde_json::Value = serde_json::from_str(&content)?;
        assert!(parsed.get("memory_mb").is_some());
        assert!(parsed.get("instances_dir").is_some());

        Ok(())
    }

    #[tokio::test]
    async fn test_settings_file_permissions() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();

        let settings = LauncherSettings::default();

        // Save settings
        save_settings_to_dir(&settings, &config_dir).await?;

        // Check file permissions (should be readable/writable by owner)
        let config_file = config_dir.join("settings.json");
        let metadata = fs::metadata(&config_file).await?;

        // File should exist and not be empty
        assert!(metadata.len() > 0);

        // File should be readable
        let content = fs::read_to_string(&config_file).await?;
        assert!(!content.is_empty());

        Ok(())
    }

    #[tokio::test]
    async fn test_concurrent_settings_access() -> Result<(), Box<dyn std::error::Error>> {
        let temp_dir = TempDir::new()?;
        let config_dir = temp_dir.path().to_path_buf();
        let config_dir_clone = config_dir.clone();

        // Test concurrent access to settings
        let handle1 = tokio::spawn(async move {
            let settings = LauncherSettings {
                memory_mb: 4096,
                theme: "dark".to_string(),
                ..Default::default()
            };
            save_settings_to_dir(&settings, &config_dir).await
        });

        let handle2 = tokio::spawn(async move {
            load_settings_from_dir(&config_dir_clone).await
        });

        // Both operations should complete without errors
        let result1 = handle1.await??;
        let result2 = handle2.await??;

        assert!(result1.is_ok());
        assert!(result2.is_ok());

        Ok(())
    }

    #[tokio::test]
    async fn test_settings_serialization_roundtrip() -> Result<(), Box<dyn std::error::Error>> {
        let original_settings = LauncherSettings {
            memory_mb: 6144,
            theme: "system".to_string(),
            java_path: Some("/test/java".to_string()),
            prism_path: Some("/test/prism".to_string()),
            instances_dir: Some("/test/instances".to_string()),
            auto_update: false,
            max_concurrent_downloads: 5,
            update_channel: "beta".to_string(),
            allow_prerelease: true,
            ..Default::default()
        };

        // Serialize to JSON
        let json = serde_json::to_string_pretty(&original_settings)?;

        // Deserialize back
        let deserialized_settings: LauncherSettings = serde_json::from_str(&json)?;

        // Verify all fields match
        assert_eq!(original_settings.memory_mb, deserialized_settings.memory_mb);
        assert_eq!(original_settings.theme, deserialized_settings.theme);
        assert_eq!(original_settings.java_path, deserialized_settings.java_path);
        assert_eq!(original_settings.prism_path, deserialized_settings.prism_path);
        assert_eq!(original_settings.instances_dir, deserialized_settings.instances_dir);
        assert_eq!(original_settings.auto_update, deserialized_settings.auto_update);
        assert_eq!(original_settings.max_concurrent_downloads, deserialized_settings.max_concurrent_downloads);
        assert_eq!(original_settings.update_channel, deserialized_settings.update_channel);
        assert_eq!(original_settings.allow_prerelease, deserialized_settings.allow_prerelease);

        Ok(())
    }
}