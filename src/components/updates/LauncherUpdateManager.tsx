import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Progress } from '../ui/Progress';
import { Badge } from '../ui/Badge';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { Modal } from '../ui/Modal';
import { Select } from '../ui/Select';
import { Checkbox } from '../ui/Checkbox';
import { formatFileSize, formatTimeAgo } from '../../utils/format';
import type {
  UpdateInfo,
  UpdateProgress,
  UpdateSettings,
  UpdateChannel,
  UpdateStatus
} from '../../types/launcher';

interface LauncherUpdateManagerProps {
  className?: string;
}

export const LauncherUpdateManager: React.FC<LauncherUpdateManagerProps> = ({
  className = ''
}) => {
  const [currentVersion, setCurrentVersion] = useState<string>('');
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
  const [updateProgress, setUpdateProgress] = useState<Map<string, UpdateProgress>>(new Map());
  const [updateSettings, setUpdateSettings] = useState<UpdateSettings | null>(null);
  const [_isLoading, _setIsLoading] = useState(false);
  const [isChecking, setIsChecking] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [activeDownloadId, setActiveDownloadId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Initialize component
  useEffect(() => {
    initializeUpdateManager();
    loadCurrentVersion();
    loadUpdateSettings();

    // Check for updates on startup if enabled
    const checkUpdatesOnStartup = async () => {
      const settings = await loadUpdateSettings();
      if (settings?.checkUpdatesOnStartup) {
        await checkForUpdates();
      }
    };

    checkUpdatesOnStartup();
  }, []);

  // Monitor update progress
  useEffect(() => {
    if (activeDownloadId) {
      const interval = setInterval(async () => {
        try {
          const progress = await invoke<UpdateProgress | null>('get_update_progress', {
            downloadId: activeDownloadId
          });

          if (progress) {
            setUpdateProgress(prev => new Map(prev).set(activeDownloadId, progress));

            // Clear interval when download is complete or failed
            if (progress.status === 'installed' || progress.status === 'failed') {
              clearInterval(interval);
              setActiveDownloadId(null);

              if (progress.status === 'installed') {
                // Show success message and optionally restart
                setUpdateInfo(null);
              }
            }
          }
        } catch (err) {
          console.error('Failed to get update progress:', err);
        }
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [activeDownloadId]);

  const initializeUpdateManager = async () => {
    try {
      await invoke('initialize_update_manager');
    } catch (err) {
      console.error('Failed to initialize update manager:', err);
      setError('Failed to initialize update manager');
    }
  };

  const loadCurrentVersion = async () => {
    try {
      const version = await invoke<string>('get_app_version');
      setCurrentVersion(version);
    } catch (err) {
      console.error('Failed to get current version:', err);
    }
  };

  const loadUpdateSettings = async () => {
    try {
      const settings = await invoke<UpdateSettings>('get_update_settings');
      setUpdateSettings(settings);
      return settings;
    } catch (err) {
      console.error('Failed to load update settings:', err);
      return null;
    }
  };

  const checkForUpdates = async () => {
    setIsChecking(true);
    setError(null);

    try {
      const update = await invoke<UpdateInfo | null>('check_for_updates');
      setUpdateInfo(update);
    } catch (err) {
      console.error('Failed to check for updates:', err);
      setError('Failed to check for updates');
    } finally {
      setIsChecking(false);
    }
  };

  const startUpdate = async () => {
    if (!updateInfo) return;

    setError(null);

    try {
      const downloadId = await invoke<string>('download_update', {
        updateInfo: updateInfo
      });

      setActiveDownloadId(downloadId);
      setShowUpdateModal(false);
    } catch (err) {
      console.error('Failed to start update:', err);
      setError('Failed to start update download');
    }
  };

  const applyUpdate = async (downloadId: string) => {
    try {
      await invoke('apply_update', { downloadId });
      // The launcher will restart after successful update
    } catch (err) {
      console.error('Failed to apply update:', err);
      setError('Failed to apply update');
    }
  };

  const cancelUpdate = async (downloadId: string) => {
    try {
      await invoke('cancel_update_download', { downloadId });
      setActiveDownloadId(null);
      setUpdateProgress(prev => {
        const newMap = new Map(prev);
        newMap.delete(downloadId);
        return newMap;
      });
    } catch (err) {
      console.error('Failed to cancel update:', err);
      setError('Failed to cancel update');
    }
  };

  const saveUpdateSettings = async (newSettings: UpdateSettings) => {
    try {
      await invoke('update_update_settings', { settings: newSettings });
      setUpdateSettings(newSettings);
    } catch (err) {
      console.error('Failed to save update settings:', err);
      setError('Failed to save update settings');
    }
  };

  const getChannelBadgeColor = (channel: UpdateChannel) => {
    switch (channel) {
      case 'stable': return 'success';
      case 'beta': return 'warning';
      case 'alpha': return 'error';
      default: return 'outline';
    }
  };

  const getStatusIcon = (status: UpdateStatus) => {
    switch (status) {
      case 'checking': return '🔍';
      case 'available': return '⬇️';
      case 'downloading': return '⏬';
      case 'downloaded': return '✅';
      case 'installing': return '⚙️';
      case 'installed': return '🎉';
      case 'failed': return '❌';
      case 'rollingBack': return '🔄';
      default: return '❓';
    }
  };

  const getStatusColor = (status: UpdateStatus) => {
    switch (status) {
      case 'checking': return 'text-blue-600';
      case 'available': return 'text-green-600';
      case 'downloading': return 'text-blue-600';
      case 'downloaded': return 'text-green-600';
      case 'installing': return 'text-yellow-600';
      case 'installed': return 'text-green-600';
      case 'failed': return 'text-red-600';
      case 'rollingBack': return 'text-orange-600';
      default: return 'text-gray-600';
    }
  };

  const activeProgress = activeDownloadId ? updateProgress.get(activeDownloadId) : null;

  return (
    <div className={`launcher-update-manager ${className}`}>
      {/* Current Version Card */}
      <Card className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold mb-2">Launcher Version</h3>
            <div className="flex items-center gap-3">
              <span className="text-2xl font-bold">{currentVersion}</span>
              <Badge variant="outline">Current</Badge>
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={checkForUpdates}
              disabled={isChecking}
            >
              {isChecking ? <LoadingSpinner size="sm" /> : '🔄'} Check Updates
            </Button>
            <Button
              variant="outline"
              onClick={() => setShowSettingsModal(true)}
            >
              ⚙️ Settings
            </Button>
          </div>
        </div>
      </Card>

      {/* Error Display */}
      {error && (
        <Card className="mb-6 border-red-200 bg-red-50">
          <div className="flex items-center gap-2 text-red-800">
            <span>❌</span>
            <span>{error}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setError(null)}
              className="ml-auto"
            >
              ✕
            </Button>
          </div>
        </Card>
      )}

      {/* Update Available Card */}
      {updateInfo && (
        <Card className="mb-6 border-green-200 bg-green-50">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-2xl">🎉</span>
                <div>
                  <h3 className="text-lg font-semibold text-green-800">Update Available!</h3>
                  <p className="text-green-600">
                    Version {updateInfo.version} is ready to download
                  </p>
                </div>
              </div>
              <Badge variant={getChannelBadgeColor(updateInfo.channel)}>
                {updateInfo.channel.toUpperCase()}
              </Badge>
            </div>

            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-600">Current Version:</span>
                <span className="ml-2 font-medium">{currentVersion}</span>
              </div>
              <div>
                <span className="text-gray-600">New Version:</span>
                <span className="ml-2 font-medium text-green-600">{updateInfo.version}</span>
              </div>
              <div>
                <span className="text-gray-600">Download Size:</span>
                <span className="ml-2 font-medium">{formatFileSize(updateInfo.fileSize)}</span>
              </div>
              <div>
                <span className="text-gray-600">Released:</span>
                <span className="ml-2 font-medium">
                  {formatTimeAgo(updateInfo.publishedAt)}
                </span>
              </div>
            </div>

            {updateInfo.releaseNotes && (
              <div className="mt-4">
                <h4 className="font-medium mb-2">Release Notes:</h4>
                <div className="bg-white p-3 rounded border text-sm max-h-32 overflow-y-auto">
                  <pre className="whitespace-pre-wrap font-sans">
                    {updateInfo.releaseNotes}
                  </pre>
                </div>
              </div>
            )}

            <div className="flex gap-2">
              <Button onClick={() => setShowUpdateModal(true)}>
                Download & Install Update
              </Button>
              <Button
                variant="outline"
                onClick={() => setUpdateInfo(null)}
              >
                Dismiss
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Active Update Progress */}
      {activeProgress && (
        <Card className="mb-6 border-blue-200 bg-blue-50">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{getStatusIcon(activeProgress.status)}</span>
                <div>
                  <h3 className="text-lg font-semibold">Updating Launcher</h3>
                  <p className={`capitalize ${getStatusColor(activeProgress.status)}`}>
                    {activeProgress.status.replace(/([A-Z])/g, ' $1').trim()}
                  </p>
                </div>
              </div>
              <Badge variant="outline">
                {activeProgress.version}
              </Badge>
            </div>

            {(activeProgress.status === 'downloading' || activeProgress.status === 'installing') && (
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>
                    {activeProgress.status === 'downloading' ? 'Downloading' : 'Installing'}
                  </span>
                  <span>{Math.round(activeProgress.progressPercent)}%</span>
                </div>
                <Progress
                  value={activeProgress.progressPercent}
                  color="primary"
                />
                {activeProgress.status === 'downloading' && (
                  <div className="flex justify-between text-xs text-gray-600">
                    <span>
                      {formatFileSize(activeProgress.downloadedBytes)} / {formatFileSize(activeProgress.totalBytes)}
                    </span>
                    <span>{formatFileSize(activeProgress.downloadSpeedBps)}/s</span>
                  </div>
                )}
              </div>
            )}

            {activeProgress.status === 'downloaded' && (
              <div className="flex gap-2">
                <Button onClick={() => applyUpdate(activeProgress.downloadId)}>
                  Install Update Now
                </Button>
                <Button
                  variant="outline"
                  onClick={() => cancelUpdate(activeProgress.downloadId)}
                >
                  Cancel
                </Button>
              </div>
            )}

            {activeProgress.status === 'failed' && (
              <div className="flex gap-2">
                <Button onClick={() => cancelUpdate(activeProgress.downloadId)}>
                  Clear Failed Update
                </Button>
                <Button
                  variant="outline"
                  onClick={checkForUpdates}
                >
                  Try Again
                </Button>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* No Updates Available */}
      {!updateInfo && !isChecking && !activeProgress && (
        <Card className="text-center py-8">
          <span className="text-4xl mb-2 block">✅</span>
          <h3 className="text-lg font-semibold mb-2">Up to Date</h3>
          <p className="text-gray-600 mb-4">
            You're running the latest version of TheBoys Launcher
          </p>
          <Button variant="outline" onClick={checkForUpdates}>
            Check Again
          </Button>
        </Card>
      )}

      {/* Update Settings Modal */}
      <Modal
        isOpen={showSettingsModal}
        onClose={() => setShowSettingsModal(false)}
        title="Update Settings"
      >
        {updateSettings && (
          <div className="space-y-4">
            <div>
              <Checkbox
                id="auto-update"
                label="Automatically download and install updates"
                checked={updateSettings.autoUpdateEnabled}
                onChange={(e) => {
                  saveUpdateSettings({
                    ...updateSettings,
                    autoUpdateEnabled: e.target.checked
                  });
                }}
              />
              <p className="text-sm text-gray-600 mt-1">
                When enabled, updates will be downloaded and installed automatically
              </p>
            </div>

            <div>
              <Checkbox
                id="check-startup"
                label="Check for updates on startup"
                checked={updateSettings.checkUpdatesOnStartup}
                onChange={(e) => {
                  saveUpdateSettings({
                    ...updateSettings,
                    checkUpdatesOnStartup: e.target.checked
                  });
                }}
              />
              <p className="text-sm text-gray-600 mt-1">
                Automatically check for new updates when the launcher starts
              </p>
            </div>

            <div>
              <Checkbox
                id="backup-update"
                label="Create backup before updating"
                checked={updateSettings.backupBeforeUpdate}
                onChange={(e) => {
                  saveUpdateSettings({
                    ...updateSettings,
                    backupBeforeUpdate: e.target.checked
                  });
                }}
              />
              <p className="text-sm text-gray-600 mt-1">
                Create a backup of the current launcher before applying updates
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">
                Update Channel
              </label>
              <Select
                value={updateSettings.updateChannel}
                onChange={(e) => {
                  saveUpdateSettings({
                    ...updateSettings,
                    updateChannel: e.target.value as UpdateChannel
                  });
                }}
              >
                <option value="stable">Stable - Recommended for most users</option>
                <option value="beta">Beta - Test new features before release</option>
                <option value="alpha">Alpha - Latest features (may be unstable)</option>
              </Select>
            </div>

            <div>
              <Checkbox
                id="allow-prerelease"
                label="Include pre-release versions in stable channel"
                checked={updateSettings.allowPrerelease}
                onChange={(e) => {
                  saveUpdateSettings({
                    ...updateSettings,
                    allowPrerelease: e.target.checked
                  });
                }}
              />
              <p className="text-sm text-gray-600 mt-1">
                Show beta and alpha versions even when on stable channel
              </p>
            </div>

            <div className="flex justify-end">
              <Button onClick={() => setShowSettingsModal(false)}>
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Update Confirmation Modal */}
      <Modal
        isOpen={showUpdateModal}
        onClose={() => setShowUpdateModal(false)}
        title="Confirm Update"
      >
        {updateInfo && (
          <div className="space-y-4">
            <div className="p-3 border rounded-lg bg-blue-50">
              <p className="text-sm text-blue-800">
                📦 The launcher will download and install version {updateInfo.version}.
                The application will restart automatically after installation.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-600">Version:</span>
                <span className="ml-2 font-medium">{updateInfo.version}</span>
              </div>
              <div>
                <span className="text-gray-600">Size:</span>
                <span className="ml-2 font-medium">{formatFileSize(updateInfo.fileSize)}</span>
              </div>
              <div>
                <span className="text-gray-600">Channel:</span>
                <span className="ml-2 font-medium">{updateInfo.channel}</span>
              </div>
              <div>
                <span className="text-gray-600">Type:</span>
                <span className="ml-2 font-medium">
                  {updateInfo.prerelease ? 'Pre-release' : 'Stable Release'}
                </span>
              </div>
            </div>

            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => setShowUpdateModal(false)}
              >
                Cancel
              </Button>
              <Button onClick={startUpdate}>
                Download & Install
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};