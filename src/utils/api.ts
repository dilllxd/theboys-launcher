import { invoke } from '@tauri-apps/api/core';
import { LauncherResult } from '../types/api';
import toast from 'react-hot-toast';

// Generic wrapper for Tauri commands with error handling
export async function invokeCommand<T>(
  command: string,
  args?: any
): Promise<T> {
  try {
    const result = await invoke<LauncherResult<T>>(command, args);

    if (result.success && result.data !== undefined) {
      return result.data;
    } else {
      const errorMessage = result.error || 'Unknown error occurred';
      toast.error(errorMessage);
      throw new Error(errorMessage);
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    toast.error(`Failed to execute ${command}: ${message}`);
    throw error;
  }
}

// Specific API functions
export const api = {
  // Health and version
  healthCheck: () => invokeCommand<string>('health_check'),
  getAppVersion: () => invokeCommand<string>('get_app_version'),

  // Settings
  getSettings: () => invokeCommand<import('../types/launcher').LauncherSettings>('get_settings'),
  saveSettings: (settings: import('../types/launcher').LauncherSettings) =>
    invokeCommand<void>('save_settings', { settings }),
  resetSettings: () => invokeCommand<import('../types/launcher').LauncherSettings>('reset_settings'),
  detectSystemInfo: () => invokeCommand<import('../types/launcher').SystemInfo>('detect_system_info'),
  validateSetting: (key: string, value: string) =>
    invokeCommand<boolean>('validate_setting', { key, value }),
  browseForJava: () => invokeCommand<string | null>('browse_for_java'),
  browseForPrism: () => invokeCommand<string | null>('browse_for_prism'),
  browseForInstancesDir: () => invokeCommand<string | null>('browse_for_instances_dir'),

  // Modpacks
  getAvailableModpacks: () =>
    invokeCommand<import('../types/launcher').Modpack[]>('get_available_modpacks'),
  getInstalledModpacks: () =>
    invokeCommand<import('../types/launcher').InstalledModpack[]>('get_installed_modpacks'),
  getDefaultModpack: () =>
    invokeCommand<import('../types/launcher').Modpack | null>('get_default_modpack'),
  checkModpackUpdates: (modpackId: string) =>
    invokeCommand<import('../types/launcher').ModpackUpdate | null>('check_modpack_updates', { modpackId }),
  checkAllModpackUpdates: () =>
    invokeCommand<import('../types/launcher').ModpackUpdate[]>('check_all_modpack_updates'),
  selectDefaultModpack: (modpackId: string) =>
    invokeCommand<void>('select_default_modpack', { modpackId }),
  downloadModpack: (modpackId: string) =>
    invokeCommand<string>('download_modpack', { modpackId }),

  // Minecraft
  launchMinecraft: (instanceId: string) =>
    invokeCommand<void>('launch_minecraft', { instanceId }),

  // System
  checkJavaInstallation: () =>
    invokeCommand<import('../types/launcher').JavaVersion[]>('check_java_installation'),
  installPrismLauncher: (version?: string) =>
    invokeCommand<string>('install_prism_launcher', { version }),
  getSystemInfo: () =>
    invokeCommand<import('../types/launcher').SystemInfo>('get_system_info'),

  // Java Management
  detectJavaInstallations: () =>
    invokeCommand<import('../types/launcher').JavaInstallation[]>('detect_java_installations'),
  getManagedJavaInstallations: () =>
    invokeCommand<import('../types/launcher').JavaInstallation[]>('get_managed_java_installations'),
  getRequiredJavaVersion: (minecraftVersion: string) =>
    invokeCommand<string | null>('get_required_java_version', { minecraftVersion }),
  getBestJavaInstallation: (minecraftVersion: string) =>
    invokeCommand<import('../types/launcher').JavaInstallation | null>('get_best_java_installation', { minecraftVersion }),
  checkJavaCompatibility: (minecraftVersion: string) =>
    invokeCommand<import('../types/launcher').JavaCompatibilityInfo>('check_java_compatibility', { minecraftVersion }),
  installJavaVersion: (javaVersion: string) =>
    invokeCommand<string>('install_java_version', { javaVersion }),
  removeJavaInstallation: (javaVersion: string) =>
    invokeCommand<void>('remove_java_installation', { javaVersion }),
  cleanupJavaInstallations: () =>
    invokeCommand<number>('cleanup_java_installations'),
  getJavaDownloadInfo: (javaVersion: string) =>
    invokeCommand<[string, number]>('get_java_download_info', { javaVersion }),

  // Prism Management
  detectPrismInstallation: () =>
    invokeCommand<import('../types/launcher').PrismInstallation | null>('detect_prism_installation'),
  getPrismStatus: () =>
    invokeCommand<import('../types/launcher').PrismStatus>('get_prism_status'),
  checkPrismUpdates: () =>
    invokeCommand<import('../types/launcher').PrismUpdateInfo>('check_prism_updates'),
  installPrismLauncherNew: (version?: string) =>
    invokeCommand<string>('install_prism_launcher_new', { version }),
  uninstallPrismLauncher: () =>
    invokeCommand<void>('uninstall_prism_launcher'),
  getPrismInstallPath: () =>
    invokeCommand<string>('get_prism_install_path'),
  verifyPrismInstallation: (path: string) =>
    invokeCommand<boolean>('verify_prism_installation', { path }),
  getPrismDownloadInfo: (version?: string) =>
    invokeCommand<[string, number, string]>('get_prism_download_info', { version }),

  // Downloads
  getDownloadProgress: (downloadId: string) =>
    invokeCommand<import('../types/launcher').DownloadProgress | null>('get_download_progress', { downloadId }),
  cancelDownload: (downloadId: string) =>
    invokeCommand<void>('cancel_download', { downloadId }),
  pauseDownload: (downloadId: string) =>
    invokeCommand<void>('pause_download', { downloadId }),
  resumeDownload: (downloadId: string) =>
    invokeCommand<void>('resume_download', { downloadId }),
  removeDownload: (downloadId: string) =>
    invokeCommand<void>('remove_download', { downloadId }),
  setMaxConcurrentDownloads: (maxConcurrent: number) =>
    invokeCommand<void>('set_max_concurrent_downloads', { maxConcurrent }),

  // Instance Management
  createInstance: (config: import('../types/launcher').InstanceConfig) =>
    invokeCommand<import('../types/launcher').Instance>('create_instance', { config }),
  getInstances: () =>
    invokeCommand<import('../types/launcher').Instance[]>('get_instances'),
  getInstance: (instanceId: string) =>
    invokeCommand<import('../types/launcher').Instance | null>('get_instance', { instanceId }),
  getInstanceByName: (name: string) =>
    invokeCommand<import('../types/launcher').Instance | null>('get_instance_by_name', { name }),
  updateInstance: (instance: import('../types/launcher').Instance) =>
    invokeCommand<void>('update_instance', { instance }),
  deleteInstance: (instanceId: string) =>
    invokeCommand<void>('delete_instance', { instanceId }),
  launchInstance: (instanceId: string) =>
    invokeCommand<void>('launch_instance', { instanceId }),
  validateInstance: (instanceId: string) =>
    invokeCommand<import('../types/launcher').InstanceValidation>('validate_instance', { instanceId }),
  installModloader: (instanceId: string) =>
    invokeCommand<void>('install_modloader', { instanceId }),
  getModloaderVersions: (modloader: import('../types/launcher').Modloader, minecraftVersion: string) =>
    invokeCommand<string[]>('get_modloader_versions', { modloader, minecraftVersion }),
  repairInstance: (instanceId: string) =>
    invokeCommand<void>('repair_instance', { instanceId }),
  getInstanceStatus: (instanceId: string) =>
    invokeCommand<import('../types/launcher').InstanceStatus | null>('get_instance_status', { instanceId }),
  setInstanceStatus: (instanceId: string, status: import('../types/launcher').InstanceStatus) =>
    invokeCommand<void>('set_instance_status', { instanceId, status }),
  getInstanceLogs: (instanceId: string) =>
    invokeCommand<string[]>('get_instance_logs', { instanceId }),
  clearInstanceLogs: (instanceId: string) =>
    invokeCommand<void>('clear_instance_logs', { instanceId }),
  getInstanceStatistics: (instanceId: string) =>
    invokeCommand<import('../types/launcher').InstanceStatistics>('get_instance_statistics', { instanceId }),
};