// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::env;
use tracing::{info, error, warn};
use tracing_subscriber;

mod commands;
mod models;
mod utils;
mod downloader;
mod launcher;
mod modpack;
mod packwiz;
mod java;
mod prism;
mod update;

use commands::*;

fn main() {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    // Get version from environment or use default
    let version = env::var("CARGO_PKG_VERSION").unwrap_or_else(|_| "1.1.0".to_string());
    info!("TheBoys Launcher v{} starting", version);

    // Load initial settings
    let initial_settings = match utils::config::load_settings() {
        Ok(settings) => {
            info!("Initial settings loaded successfully");
            settings
        }
        Err(e) => {
            warn!("Failed to load initial settings, using defaults: {}", e);
            models::LauncherSettings::default()
        }
    };

    // Initialize update manager
    let update_manager = std::sync::Arc::new(update::UpdateManager::new(version.clone()));

    // Initialize application state
    let app_state = commands::AppState {
        settings: std::sync::Arc::new(std::sync::Mutex::new(initial_settings)),
        download_progress: std::sync::Arc::new(std::sync::Mutex::new(std::collections::HashMap::new())),
        launch_manager: std::sync::Arc::new(launcher::LaunchManager::new()),
        update_manager,
    };

    // Initialize Tauri application
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .manage(app_state)
        .invoke_handler(tauri::generate_handler![
            health_check,
            get_app_version,
            get_settings,
            save_settings,
            reset_settings,
            detect_system_info,
            validate_setting,
            browse_for_java,
            browse_for_prism,
            browse_for_instances_dir,
            get_available_modpacks,
            get_installed_modpacks,
            check_modpack_updates,
            check_all_modpack_updates,
            select_default_modpack,
            get_default_modpack,
            get_modpack,
            clear_modpack_cache,
            download_modpack,
            launch_minecraft,
            check_java_installation,
            install_prism_launcher,
            get_system_info,
            // Download management commands
            download_file,
            get_download_progress,
            get_all_downloads,
            cancel_download,
            pause_download,
            resume_download,
            remove_download,
            set_max_concurrent_downloads,
            download_prism_launcher,
            download_java,
            download_packwiz_bootstrap,
            // Java Management commands
            detect_java_installations,
            get_managed_java_installations,
            get_required_java_version,
            get_best_java_installation,
            check_java_compatibility,
            install_java_version,
            remove_java_installation,
            cleanup_java_installations,
            get_java_download_info,
            // Prism Management commands
            detect_prism_installation,
            get_prism_status,
            check_prism_updates,
            install_prism_launcher_new,
            uninstall_prism_launcher,
            get_prism_install_path,
            verify_prism_installation,
            get_prism_download_info,
            // Instance Management commands
            create_instance,
            create_instance_from_modpack,
            install_modpack_to_instance,
            get_instances,
            get_instance,
            get_instance_by_name,
            update_instance,
            delete_instance,
            launch_instance,
            validate_instance,
            install_modloader,
            get_modloader_versions,
            repair_instance,
            get_instance_status,
            set_instance_status,
            get_instance_logs,
            clear_instance_logs,
            get_instance_statistics,
            // Launch Management commands
            get_launch_status,
            get_active_processes,
            get_instance_processes,
            terminate_process,
            force_kill_instance,
            cleanup_finished_processes,
            get_launch_config,
            update_launch_config,
            initialize_launch_manager,
            // Update Management commands
            initialize_update_manager,
            check_for_updates,
            download_update,
            apply_update,
            get_update_progress,
            get_all_update_downloads,
            cancel_update_download,
            cleanup_update_downloads,
            get_update_settings,
            update_update_settings,
            set_update_channel,
            set_auto_update,
            set_allow_prerelease_updates,
            // Packwiz Integration commands
            initialize_packwiz,
            install_modpack_with_packwiz,
            check_instance_updates,
            create_instance_backup,
            get_instance_backups,
            restore_instance_backup,
            delete_instance_backup,
            get_pack_install_progress,
            get_manual_downloads,
            confirm_manual_download,
            cancel_modpack_installation,
            get_pack_manifest,
            validate_manual_download,
            get_instance_update_history,
            download_manual_file,
            // Performance Monitoring commands
            get_performance_metrics,
            clear_performance_cache,
            // Tauri Updater Integration commands
            check_and_install_update,
            // check_tauri_update, // Temporarily disabled due to async issues
            start_update_download,
            install_downloaded_update,
            get_update_info,
            configure_auto_update,
            set_update_notifications
        ])
        .run(tauri::generate_context!())
        .map_err(|e| {
            error!("Failed to run Tauri application: {}", e);
            e
        })
        .expect("error while running tauri application");
}