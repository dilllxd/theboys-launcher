// API types and error handling

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: any;
}

// Tauri command return types
export type LauncherResult<T> = Promise<{
  success: boolean;
  data?: T;
  error?: string;
}>;

// Command parameter types
export interface DownloadModpackParams {
  modpackId: string;
}

export interface LaunchMinecraftParams {
  instanceId: string;
}

export interface SaveSettingsParams {
  settings: import('./launcher').LauncherSettings;
}

export interface InstallPrismParams {
  version?: string;
}

// Event types for real-time updates
export interface DownloadProgressEvent {
  downloadId: string;
  progress: import('./launcher').DownloadProgress;
}

export interface InstanceStatusEvent {
  instanceId: string;
  status: 'starting' | 'running' | 'stopped' | 'error';
  details?: string;
}

// Query keys for React Query
export const queryKeys = {
  health: ['health'],
  version: ['version'],
  settings: ['settings'],
  modpacks: ['modpacks'],
  installedModpacks: ['installedModpacks'],
  modpackUpdates: ['modpackUpdates'],
  defaultModpack: ['defaultModpack'],
  systemInfo: ['systemInfo'],
  javaVersions: ['javaVersions'],
  javaInstallations: ['javaInstallations'],
  managedJavaInstallations: ['managedJavaInstallations'],
  javaCompatibility: (minecraftVersion: string) => ['javaCompatibility', minecraftVersion],
  prismStatus: ['prismStatus'],
  prismInstallation: ['prismInstallation'],
  prismUpdates: ['prismUpdates'],
  instances: ['instances'],
  downloadProgress: (downloadId: string) => ['downloadProgress', downloadId],
} as const;