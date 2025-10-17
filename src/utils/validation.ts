import { LauncherSettings } from '../types/launcher';

// Validation functions
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validateUrl = (url: string): boolean => {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
};

export const validateMemorySetting = (memoryMb: number): boolean => {
  return memoryMb >= 1024 && memoryMb <= 32768; // 1GB to 32GB
};

export const validateJavaPath = (path: string): boolean => {
  // Basic validation for Java path
  return path.length > 0 && (path.endsWith('java') || path.endsWith('java.exe'));
};

export const validateDirectoryName = (name: string): boolean => {
  // Check for invalid characters
  const invalidChars = /[<>:"/\\|?*]/;
  return name.length > 0 && name.length <= 255 && !invalidChars.test(name);
};

export const validateSettings = (settings: LauncherSettings): string[] => {
  const errors: string[] = [];

  if (!validateMemorySetting(settings.memoryMb)) {
    errors.push('Memory allocation must be between 1GB and 32GB');
  }

  if (settings.javaPath && !validateJavaPath(settings.javaPath)) {
    errors.push('Invalid Java path');
  }

  if (settings.instancesDir && !validateDirectoryName(settings.instancesDir)) {
    errors.push('Invalid instances directory name');
  }

  if (!['light', 'dark', 'system'].includes(settings.theme)) {
    errors.push('Invalid theme setting');
  }

  return errors;
};

export const validateModpackId = (id: string): boolean => {
  // Modpack IDs should be alphanumeric with hyphens and underscores
  const validId = /^[a-zA-Z0-9_-]+$/;
  return id.length > 0 && id.length <= 64 && validId.test(id);
};

export const validateInstanceId = (id: string): boolean => {
  // Similar to modpack ID validation
  return validateModpackId(id);
};

export const sanitizeFileName = (name: string): string => {
  // Remove or replace invalid characters
  return name
    .replace(/[<>:"/\\|?*]/g, '_')
    .replace(/\s+/g, '_')
    .substring(0, 255);
};

export const isValidMinecraftVersion = (version: string): boolean => {
  // Basic Minecraft version validation (e.g., "1.20.1", "1.19.4")
  const versionRegex = /^\d+\.\d+(\.\d+)?(-.+)?$/;
  return versionRegex.test(version);
};