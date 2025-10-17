// Launcher data types that match the Rust backend

export interface Modpack {
  id: string;
  displayName: string;
  packUrl: string;
  instanceName: string;
  description: string;
  default: boolean;
  version: string;
  minecraftVersion: string;
  modloader: Modloader;
  loaderVersion: string;
}

export type Modloader = 'vanilla' | 'forge' | 'fabric' | 'quilt' | 'neoforge';

export interface InstalledModpack {
  modpack: Modpack;
  installedVersion: string;
  installPath: string;
  installDate: string;
  lastPlayed?: string;
  totalPlaytime: number;
  updateAvailable: boolean;
  sizeBytes: number;
}

export interface ModpackUpdate {
  modpackId: string;
  currentVersion: string;
  latestVersion: string;
  updateAvailable: boolean;
  changelogUrl?: string;
  downloadUrl: string;
  sizeBytes: number;
}

export interface LauncherSettings {
  memoryMb: number;
  javaPath?: string;
  prismPath?: string;
  instancesDir?: string;
  autoUpdate: boolean;
  theme: 'light' | 'dark' | 'system';
  defaultModpackId?: string;
}

export interface SystemInfo {
  os: string;
  arch: string;
  totalMemoryMb: number;
  availableMemoryMb: number;
  cpuCores: number;
  javaInstalled: boolean;
  javaVersions: JavaVersion[];
}

export interface JavaVersion {
  version: string;
  path: string;
  is64bit: boolean;
  majorVersion: number;
}

export interface JavaInstallation {
  version: string;
  path: string;
  is64bit: boolean;
  majorVersion: number;
  isManaged: boolean;
  installationDate?: string;
  sizeBytes?: number;
}

export interface JavaCompatibilityInfo {
  minecraftVersion: string;
  requiredJavaVersion?: string;
  hasCompatibleJava: boolean;
  recommendedInstallation?: JavaInstallation;
}

export interface PrismInstallation {
  version: string;
  path: string;
  executablePath: string;
  isManaged: boolean;
  installationDate?: string;
  sizeBytes?: number;
  architecture: string;
  platform: string;
}

export interface PrismUpdateInfo {
  currentVersion?: string;
  latestVersion: string;
  updateAvailable: boolean;
  downloadUrl: string;
  releaseNotes?: string;
  sizeBytes: number;
}

export interface PrismStatus {
  isInstalled: boolean;
  installation?: PrismInstallation;
  isManaged: boolean;
  platform: string;
  architecture: string;
}

export interface DownloadProgress {
  id: string;
  name: string;
  downloadedBytes: number;
  totalBytes: number;
  progressPercent: number;
  speedBps: number;
  status: DownloadStatus;
}

export type DownloadStatus =
  | 'pending'
  | 'downloading'
  | 'paused'
  | 'completed'
  | { failed: string }
  | 'cancelled';

// Instance management types
export interface InstanceConfig {
  name: string;
  modpackId: string;
  minecraftVersion: string;
  loaderType: Modloader;
  loaderVersion: string;
  memoryMb: number;
  javaPath: string;
  iconPath?: string;
  jvmArgs?: string;
  envVars?: Record<string, string>;
}

export type InstanceStatus = 'ready' | 'needsUpdate' | 'broken' | 'installing' | 'running' | 'updating';

export interface Instance {
  id: string;
  name: string;
  modpackId: string;
  minecraftVersion: string;
  loaderType: string; // "vanilla", "forge", "fabric", etc.
  loaderVersion: string;
  memoryMb: number;
  javaPath: string;
  gameDir: string;
  lastPlayed?: string;
  totalPlaytime: number; // in seconds
  iconPath?: string;
  status: InstanceStatus;
  createdAt: string;
  updatedAt: string;
  jvmArgs?: string;
  envVars?: Record<string, string>;
}

export interface InstanceValidation {
  instanceId: string;
  isValid: boolean;
  issues: string[];
  recommendations: string[];
}

export interface InstanceStatistics {
  instanceId: string;
  name: string;
  totalSizeBytes: number;
  modsCount: number;
  resourcePacksCount: number;
  shaderPacksCount: number;
  screenshotsCount: number;
  totalPlaytimeSeconds: number;
  lastPlayed?: string;
  createdAt: string;
  updatedAt: string;
}

export type ThemeMode = 'light' | 'dark' | 'system';

// Navigation types
export type NavigationItem = {
  id: string;
  label: string;
  icon: string;
  path: string;
};

// Component props types
export interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  onClick?: (event?: React.MouseEvent) => void;
  className?: string;
  type?: 'button' | 'submit' | 'reset';
  style?: React.CSSProperties;
}

export interface CardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  className?: string;
  interactive?: boolean;
  onClick?: () => void;
}

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

export interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  variant?: 'spinner' | 'dots';
}

export interface ProgressProps {
  value: number; // 0-100
  max?: number;
  showLabel?: boolean;
  showPercentage?: boolean;
  color?: 'primary' | 'success' | 'warning' | 'error';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  label?: string;
}

// Launch management types
export interface LaunchConfig {
  instanceId: string;
  instanceName: string;
  prismPath: string;
  javaPath?: string;
  workingDirectory: string;
  additionalArgs: string[];
  memoryMb?: number;
  customJvmArgs: string[];
  environmentVars: Record<string, string>;
}

export type ProcessStatus = 'starting' | 'running' | 'finished' | 'crashed' | 'killed';

export interface LaunchedProcess {
  id: string;
  instanceId: string;
  instanceName: string;
  pid: number;
  status: ProcessStatus;
  startedAt: string;
  exitCode?: number;
  crashReason?: string;
  launchTimeMs?: number;
}

// Launcher self-update types
export type UpdateChannel = 'stable' | 'beta' | 'alpha';

export interface UpdateInfo {
  version: string;
  tagName: string;
  releaseNotes: string;
  publishedAt: string;
  downloadUrl: string;
  fileSize: number;
  checksum?: string;
  prerelease: boolean;
  channel: UpdateChannel;
}

export interface UpdateProgress {
  downloadId: string;
  version: string;
  downloadedBytes: number;
  totalBytes: number;
  progressPercent: number;
  downloadSpeedBps: number;
  status: UpdateStatus;
}

export type UpdateStatus = 'checking' | 'available' | 'downloading' | 'downloaded' | 'installing' | 'installed' | 'failed' | 'rollingBack';

export interface UpdateSettings {
  autoUpdateEnabled: boolean;
  updateChannel: UpdateChannel;
  checkUpdatesOnStartup: boolean;
  allowPrerelease: boolean;
  backupBeforeUpdate: boolean;
}