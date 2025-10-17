// Mock Wails bindings for development
// These will be replaced by actual Wails bindings during proper build

export const GetModpacks = async () => {
  return [];
};

export const GetInstances = async () => {
  return [];
};

export const GetSettings = async () => {
  return {
    theme: 'dark',
    autoUpdates: true,
    keepLauncherOpen: true,
    minMemory: 2048,
    maxMemory: 4096,
    autoDetectJava: true,
    jvmArgs: '',
    instanceDirectory: '',
    autoBackup: true,
    maxBackups: 5,
    downloadTimeout: 60,
    maxConcurrentDownloads: 3,
    useProxy: false,
    proxyUrl: '',
    debugMode: false,
    pauseOnError: true,
  };
};

export const GetJavaInstallations = async () => {
  return [];
};

export const SelectModpack = async (modpackID) => {
  console.log('Mock: SelectModpack called with:', modpackID);
  return null;
};

export const RefreshModpacks = async () => {
  console.log('Mock: RefreshModpacks called');
};

export const CreateInstance = async (modpack) => {
  console.log('Mock: CreateInstance called with:', modpack);
  return null;
};

export const GetInstance = async (instanceID) => {
  console.log('Mock: GetInstance called with:', instanceID);
  return null;
};

export const LaunchInstance = async (instanceID) => {
  console.log('Mock: LaunchInstance called with:', instanceID);
};

export const DeleteInstance = async (instanceID) => {
  console.log('Mock: DeleteInstance called with:', instanceID);
};

export const GetBestJavaInstallation = async (mcVersion) => {
  console.log('Mock: GetBestJavaInstallation called with:', mcVersion);
  return null;
};

export const GetJavaVersionForMinecraft = async (mcVersion) => {
  console.log('Mock: GetJavaVersionForMinecraft called with:', mcVersion);
  return '17';
};

export const DownloadJava = async (javaVersion, installDir) => {
  console.log('Mock: DownloadJava called with:', javaVersion, installDir);
};

export const IsPrismInstalled = async () => {
  console.log('Mock: IsPrismInstalled called');
  return false;
};

export const InstallModpackWithPackwiz = async (instanceID, progressCallback) => {
  console.log('Mock: InstallModpackWithPackwiz called with:', instanceID);
};

export const CheckModpackUpdate = async (instanceID) => {
  console.log('Mock: CheckModpackUpdate called with:', instanceID);
  return [false, '', '', null];
};

export const UpdateModpack = async (instanceID, progressCallback) => {
  console.log('Mock: UpdateModpack called with:', instanceID);
};

export const GetLWJGLVersionForMinecraft = async (mcVersion) => {
  console.log('Mock: GetLWJGLVersionForMinecraft called with:', mcVersion);
  return null;
};

export const CheckForUpdates = async () => {
  console.log('Mock: CheckForUpdates called');
  return null;
};

export const DownloadUpdate = async (downloadURL, progressCallback) => {
  console.log('Mock: DownloadUpdate called with:', downloadURL);
  return '';
};

export const InstallUpdate = async (updatePath) => {
  console.log('Mock: InstallUpdate called with:', updatePath);
};

export const GetVersion = async () => {
  return '1.0.0-dev';
};