import React, { useState, useEffect } from 'react';
import { Card } from './ui/Card';
import { Button } from './ui/Button';
import { LoadingSpinner } from './ui/LoadingSpinner';
import { Progress } from './ui/Progress';
import { Badge } from './ui/Badge';
import { Tooltip } from './ui/Tooltip';
import type {
  PrismUpdateInfo,
  PrismStatus as IPrismStatus,
  DownloadProgress
} from '../types';
import { api } from '../utils/api';

interface PrismStatusProps {
  className?: string;
}

export const PrismStatus: React.FC<PrismStatusProps> = ({
  className = ''
}) => {
  const [status, setStatus] = useState<IPrismStatus | null>(null);
  const [updateInfo, setUpdateInfo] = useState<PrismUpdateInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [checkingUpdates, setCheckingUpdates] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showReleaseNotes, setShowReleaseNotes] = useState(false);

  useEffect(() => {
    loadPrismStatus();
    // Set up periodic progress checking
    const interval = setInterval(checkDownloadProgress, 1000);
    return () => clearInterval(interval);
  }, []);

  const loadPrismStatus = async () => {
    try {
      setLoading(true);
      setError(null);

      const statusData = await api.getPrismStatus();
      setStatus(statusData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to load Prism status:', err);
    } finally {
      setLoading(false);
    }
  };

  const checkForUpdates = async () => {
    try {
      setCheckingUpdates(true);
      setError(null);

      const updateData = await api.checkPrismUpdates();
      setUpdateInfo(updateData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to check for updates:', err);
    } finally {
      setCheckingUpdates(false);
    }
  };

  const checkDownloadProgress = async () => {
    if (downloadProgress && downloadProgress.status !== 'completed' && downloadProgress.status !== 'cancelled') {
      try {
        const progress = await api.getDownloadProgress(downloadProgress.id);
        if (progress) {
          setDownloadProgress(progress);

          if (progress.status === 'completed') {
            // Refresh the Prism status after download completes
            setTimeout(() => {
              loadPrismStatus();
              checkForUpdates();
            }, 2000);
          }
        }
      } catch (err) {
        console.error('Failed to get download progress:', err);
      }
    }
  };

  const handleInstallPrism = async () => {
    try {
      setError(null);
      const downloadId = await api.installPrismLauncherNew();

      // Get initial progress
      const progress = await api.getDownloadProgress(downloadId);
      if (progress) {
        setDownloadProgress(progress);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to install Prism:', err);
    }
  };

  const handleUninstallPrism = async () => {
    if (!confirm('Are you sure you want to uninstall Prism Launcher? This will remove the launcher and all its data.')) {
      return;
    }

    try {
      setError(null);
      await api.uninstallPrismLauncher();
      await loadPrismStatus();
      setUpdateInfo(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to uninstall Prism:', err);
    }
  };

  const formatSize = (bytes: number) => {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const formatDate = (seconds?: string) => {
    if (!seconds) return 'Unknown';
    const age = parseInt(seconds);
    const days = Math.floor(age / 86400);
    if (days > 30) return 'Long ago';
    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
    const hours = Math.floor((age % 86400) / 3600);
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    const minutes = Math.floor((age % 3600) / 60);
    return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
  };

  const getPlatformIcon = (platform: string) => {
    switch (platform.toLowerCase()) {
      case 'windows': return '🪟';
      case 'macos': return '🍎';
      case 'linux': return '🐧';
      default: return '💻';
    }
  };

  const getStatusColor = (isInstalled: boolean) => {
    return isInstalled ? 'success' : 'warning';
  };

  const getUpdateColor = (updateAvailable: boolean) => {
    return updateAvailable ? 'primary' : 'success';
  };

  const isDownloading = downloadProgress &&
    (downloadProgress.status === 'pending' || downloadProgress.status === 'downloading');

  if (loading) {
    return (
      <Card className={`p-6 ${className}`}>
        <div className="flex items-center justify-center">
          <LoadingSpinner size="lg" />
          <span className="ml-3 text-gray-600">Loading Prism Launcher status...</span>
        </div>
      </Card>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Error Display */}
      {error && (
        <Card className="p-4 border-red-200 bg-red-50">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <span className="text-red-400">⚠️</span>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <p className="text-sm text-red-700">{error}</p>
            </div>
            <div className="ml-auto pl-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setError(null)}
              >
                Dismiss
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Download Progress */}
      {isDownloading && (
        <Card className="p-4 border-blue-200 bg-blue-50">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-blue-800">
                Installing {downloadProgress.name}
              </h3>
              <Badge variant="primary" size="sm">
                {Math.round(downloadProgress.progressPercent)}%
              </Badge>
            </div>
            <Progress
              value={downloadProgress.progressPercent}
              showPercentage={false}
              color="primary"
            />
            <div className="flex items-center justify-between text-xs text-blue-600">
              <span>{formatSize(downloadProgress.downloadedBytes)} / {formatSize(downloadProgress.totalBytes)}</span>
              <span>{formatSize(downloadProgress.speedBps)}/s</span>
            </div>
          </div>
        </Card>
      )}

      {/* Main Status Card */}
      <Card className="p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-3">
            <div className="text-3xl">🔮</div>
            <div>
              <h2 className="text-xl font-bold text-gray-900">Prism Launcher</h2>
              <p className="text-sm text-gray-600">Minecraft launcher and instance manager</p>
            </div>
          </div>
          <Badge
            variant={getStatusColor(status?.isInstalled ?? false) as any}
            size="md"
          >
            {status?.isInstalled ? 'Installed' : 'Not Installed'}
          </Badge>
        </div>

        {status?.isInstalled && status.installation ? (
          <div className="space-y-4">
            {/* Installation Details */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">Version:</span>
                  <span className="text-sm font-medium">{status.installation.version}</span>
                  {status.isManaged && (
                    <Badge variant="primary" size="sm">Managed</Badge>
                  )}
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">Platform:</span>
                  <span className="text-sm">
                    {getPlatformIcon(status.installation.platform)} {status.installation.platform}
                  </span>
                  <span className="text-sm text-gray-500">({status.installation.architecture})</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">Size:</span>
                  <span className="text-sm font-medium">
                    {formatSize(status.installation.sizeBytes || 0)}
                  </span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">Installed:</span>
                  <span className="text-sm font-medium">
                    {formatDate(status.installation.installationDate)}
                  </span>
                </div>
              </div>
            </div>

            {/* Installation Path */}
            <div className="p-3 bg-gray-50 rounded-lg">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Installation Path:</span>
                <Tooltip content={status.installation.path}>
                  <span className="text-xs text-gray-500 font-mono truncate max-w-xs">
                    {status.installation.path}
                  </span>
                </Tooltip>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex items-center justify-between pt-4 border-t">
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={checkForUpdates}
                  loading={checkingUpdates}
                >
                  Check for Updates
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={loadPrismStatus}
                >
                  Refresh
                </Button>
              </div>
              {status.isManaged && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleUninstallPrism}
                  className="text-red-600 border-red-200 hover:bg-red-50"
                >
                  Uninstall
                </Button>
              )}
            </div>
          </div>
        ) : (
          /* Not Installed State */
          <div className="text-center py-8">
            <div className="text-6xl mb-4">📦</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">Prism Launcher Not Installed</h3>
            <p className="text-sm text-gray-600 mb-6">
              Prism Launcher is required to manage Minecraft instances and modpacks.
            </p>
            <Button
              variant="primary"
              onClick={handleInstallPrism}
              disabled={isDownloading || undefined}
            >
              {isDownloading ? 'Installing...' : 'Install Prism Launcher'}
            </Button>
          </div>
        )}
      </Card>

      {/* Update Information */}
      {updateInfo && status?.isInstalled && (
        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Update Information</h3>
            <Badge
              variant={getUpdateColor(updateInfo.updateAvailable) as any}
            >
              {updateInfo.updateAvailable ? 'Update Available' : 'Up to Date'}
            </Badge>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Current Version:</span>
              <span className="text-sm font-medium">
                {updateInfo.currentVersion || 'Unknown'}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Latest Version:</span>
              <span className="text-sm font-medium text-green-600">
                {updateInfo.latestVersion}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Download Size:</span>
              <span className="text-sm font-medium">
                {formatSize(updateInfo.sizeBytes)}
              </span>
            </div>
          </div>

          {updateInfo.updateAvailable && (
            <div className="mt-4 space-y-3">
              {updateInfo.releaseNotes && (
                <div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowReleaseNotes(!showReleaseNotes)}
                  >
                    {showReleaseNotes ? 'Hide' : 'Show'} Release Notes
                  </Button>
                  {showReleaseNotes && (
                    <div className="mt-2 p-3 bg-gray-50 rounded-lg max-h-32 overflow-y-auto">
                      <pre className="text-xs text-gray-700 whitespace-pre-wrap">
                        {updateInfo.releaseNotes}
                      </pre>
                    </div>
                  )}
                </div>
              )}
              <Button
                variant="primary"
                onClick={handleInstallPrism}
                disabled={isDownloading || undefined}
              >
                {isDownloading ? 'Downloading...' : `Update to ${updateInfo.latestVersion}`}
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* Platform Information */}
      {status && (
        <Card className="p-4 border-gray-200 bg-gray-50">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-lg">{getPlatformIcon(status.platform)}</span>
              <div>
                <h4 className="text-sm font-medium text-gray-900">Platform Information</h4>
                <p className="text-xs text-gray-600">
                  {status.platform} ({status.architecture})
                </p>
              </div>
            </div>
            <Badge variant="outline" size="sm">
              {status.isManaged ? 'Managed Installation' : 'External Installation'}
            </Badge>
          </div>
        </Card>
      )}
    </div>
  );
};