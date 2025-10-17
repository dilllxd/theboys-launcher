import { useState, useEffect, useCallback } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { DownloadProgress } from '../types/launcher';

export const useDownloads = () => {
  const [downloads, setDownloads] = useState<DownloadProgress[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load all downloads
  const loadDownloads = useCallback(async () => {
    try {
      setError(null);
      const allDownloads = await invoke<DownloadProgress[]>('get_all_downloads');
      setDownloads(allDownloads);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load downloads';
      setError(errorMessage);
      console.error('Failed to load downloads:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Get specific download progress
  const getDownloadProgress = useCallback(async (downloadId: string): Promise<DownloadProgress | null> => {
    try {
      const progress = await invoke<DownloadProgress | null>('get_download_progress', { downloadId });
      return progress;
    } catch (err) {
      console.error('Failed to get download progress:', err);
      return null;
    }
  }, []);

  // Start a new download
  const startDownload = useCallback(async (name: string, url: string, destination: string): Promise<string | null> => {
    try {
      setError(null);
      const downloadId = await invoke<string>('download_file', { name, url, destination });

      // Refresh downloads list
      await loadDownloads();

      return downloadId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start download';
      setError(errorMessage);
      console.error('Failed to start download:', err);
      return null;
    }
  }, [loadDownloads]);

  // Pause a download
  const pauseDownload = useCallback(async (downloadId: string): Promise<boolean> => {
    try {
      setError(null);
      await invoke('pause_download', { downloadId });
      await loadDownloads();
      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to pause download';
      setError(errorMessage);
      console.error('Failed to pause download:', err);
      return false;
    }
  }, [loadDownloads]);

  // Resume a download
  const resumeDownload = useCallback(async (downloadId: string): Promise<boolean> => {
    try {
      setError(null);
      await invoke('resume_download', { downloadId });
      await loadDownloads();
      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to resume download';
      setError(errorMessage);
      console.error('Failed to resume download:', err);
      return false;
    }
  }, [loadDownloads]);

  // Cancel a download
  const cancelDownload = useCallback(async (downloadId: string): Promise<boolean> => {
    try {
      setError(null);
      await invoke('cancel_download', { downloadId });
      await loadDownloads();
      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to cancel download';
      setError(errorMessage);
      console.error('Failed to cancel download:', err);
      return false;
    }
  }, [loadDownloads]);

  // Remove a download from tracking
  const removeDownload = useCallback(async (downloadId: string): Promise<boolean> => {
    try {
      setError(null);
      await invoke('remove_download', { downloadId });
      await loadDownloads();
      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to remove download';
      setError(errorMessage);
      console.error('Failed to remove download:', err);
      return false;
    }
  }, [loadDownloads]);

  // Download Prism Launcher
  const downloadPrismLauncher = useCallback(async (version?: string): Promise<string | null> => {
    try {
      setError(null);
      const downloadId = await invoke<string>('download_prism_launcher', { version });
      await loadDownloads();
      return downloadId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to download Prism Launcher';
      setError(errorMessage);
      console.error('Failed to download Prism Launcher:', err);
      return null;
    }
  }, [loadDownloads]);

  // Download Java
  const downloadJava = useCallback(async (minecraftVersion: string): Promise<string | null> => {
    try {
      setError(null);
      const downloadId = await invoke<string>('download_java', { minecraftVersion });
      await loadDownloads();
      return downloadId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to download Java';
      setError(errorMessage);
      console.error('Failed to download Java:', err);
      return null;
    }
  }, [loadDownloads]);

  // Download packwiz bootstrap
  const downloadPackwizBootstrap = useCallback(async (): Promise<string | null> => {
    try {
      setError(null);
      const downloadId = await invoke<string>('download_packwiz_bootstrap');
      await loadDownloads();
      return downloadId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to download packwiz bootstrap';
      setError(errorMessage);
      console.error('Failed to download packwiz bootstrap:', err);
      return null;
    }
  }, [loadDownloads]);

  // Set max concurrent downloads
  const setMaxConcurrentDownloads = useCallback(async (maxConcurrent: number): Promise<boolean> => {
    try {
      setError(null);
      await invoke('set_max_concurrent_downloads', { maxConcurrent });
      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to set max concurrent downloads';
      setError(errorMessage);
      console.error('Failed to set max concurrent downloads:', err);
      return false;
    }
  }, []);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Get download statistics
  const getStats = useCallback(() => {
    const activeDownloads = downloads.filter(d => d.status === 'downloading').length;
    const pausedDownloads = downloads.filter(d => d.status === 'paused').length;
    const completedDownloads = downloads.filter(d => d.status === 'completed').length;
    const failedDownloads = downloads.filter(d => typeof d.status === 'object' && d.status.failed).length;
    const totalDownloads = downloads.length;

    return {
      activeDownloads,
      pausedDownloads,
      completedDownloads,
      failedDownloads,
      totalDownloads,
    };
  }, [downloads]);

  // Get active downloads
  const getActiveDownloads = useCallback(() => {
    return downloads.filter(d => d.status === 'downloading' || d.status === 'paused');
  }, [downloads]);

  // Get completed downloads
  const getCompletedDownloads = useCallback(() => {
    return downloads.filter(d => d.status === 'completed');
  }, [downloads]);

  // Get failed downloads
  const getFailedDownloads = useCallback(() => {
    return downloads.filter(d => typeof d.status === 'object' && d.status.failed);
  }, [downloads]);

  // Load downloads on mount
  useEffect(() => {
    loadDownloads();

    // Set up periodic updates for active downloads
    const interval = setInterval(() => {
      if (downloads.some(d => d.status === 'downloading')) {
        loadDownloads();
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [loadDownloads, downloads]);

  return {
    downloads,
    isLoading,
    error,
    loadDownloads,
    getDownloadProgress,
    startDownload,
    pauseDownload,
    resumeDownload,
    cancelDownload,
    removeDownload,
    downloadPrismLauncher,
    downloadJava,
    downloadPackwizBootstrap,
    setMaxConcurrentDownloads,
    clearError,
    getStats,
    getActiveDownloads,
    getCompletedDownloads,
    getFailedDownloads,
  };
};