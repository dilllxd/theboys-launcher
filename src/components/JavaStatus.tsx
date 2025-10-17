import React, { useState, useEffect } from 'react';
import { Card } from './ui/Card';
import { Button } from './ui/Button';
import { LoadingSpinner } from './ui/LoadingSpinner';
import { Progress } from './ui/Progress';
import { Badge } from './ui/Badge';
import { Tooltip } from './ui/Tooltip';
import { Modal } from './ui/Modal';
import {
  JavaInstallation,
  JavaCompatibilityInfo,
  DownloadProgress
} from '../types';
import { api } from '../utils/api';

interface JavaStatusProps {
  minecraftVersion?: string;
  className?: string;
}

export const JavaStatus: React.FC<JavaStatusProps> = ({
  minecraftVersion = '1.20.1',
  className = ''
}) => {
  const [installations, setInstallations] = useState<JavaInstallation[]>([]);
  const [managedInstallations, setManagedInstallations] = useState<JavaInstallation[]>([]);
  const [compatibility, setCompatibility] = useState<JavaCompatibilityInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [installModalOpen, setInstallModalOpen] = useState(false);
  const [selectedVersion, setSelectedVersion] = useState('');
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Available Java versions for installation
  const availableVersions = ['8', '11', '16', '17', '18', '21', '22'];

  useEffect(() => {
    loadJavaStatus();
    // Set up periodic progress checking
    const interval = setInterval(checkDownloadProgress, 1000);
    return () => clearInterval(interval);
  }, [minecraftVersion]);

  const loadJavaStatus = async () => {
    try {
      setLoading(true);
      setError(null);

      const [installationsData, managedData, compatibilityData] = await Promise.all([
        api.detectJavaInstallations(),
        api.getManagedJavaInstallations(),
        api.checkJavaCompatibility(minecraftVersion)
      ]);

      setInstallations(installationsData);
      setManagedInstallations(managedData);
      setCompatibility(compatibilityData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to load Java status:', err);
    } finally {
      setLoading(false);
    }
  };

  const checkDownloadProgress = async () => {
    if (downloadProgress && downloadProgress.status !== 'completed' && downloadProgress.status !== 'cancelled') {
      try {
        const progress = await api.getDownloadProgress(downloadProgress.id);
        if (progress) {
          setDownloadProgress(progress);

          if (progress.status === 'completed') {
            // Refresh the Java status after download completes
            setTimeout(loadJavaStatus, 2000);
          }
        }
      } catch (err) {
        console.error('Failed to get download progress:', err);
      }
    }
  };

  const handleInstallJava = async (version: string) => {
    try {
      setError(null);
      const downloadId = await api.installJavaVersion(version);

      // Get initial progress
      const progress = await api.getDownloadProgress(downloadId);
      if (progress) {
        setDownloadProgress(progress);
      }
      setInstallModalOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to install Java:', err);
    }
  };

  const handleRemoveJava = async (version: string) => {
    if (!confirm(`Are you sure you want to remove Java ${version}?`)) {
      return;
    }

    try {
      setError(null);
      await api.removeJavaInstallation(version);
      await loadJavaStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to remove Java:', err);
    }
  };

  const handleCleanup = async () => {
    if (!confirm('Are you sure you want to remove old Java installations? This will keep only the latest 3 versions.')) {
      return;
    }

    try {
      setError(null);
      const removedCount = await api.cleanupJavaInstallations();
      alert(`Successfully removed ${removedCount} old Java installations.`);
      await loadJavaStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      console.error('Failed to cleanup Java installations:', err);
    }
  };

  const formatSize = (bytes?: number) => {
    if (!bytes) return 'Unknown';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const getStatusColor = (hasCompatible: boolean) => {
    return hasCompatible ? 'success' : 'warning';
  };

  const isDownloading = downloadProgress &&
    (downloadProgress.status === 'pending' || downloadProgress.status === 'downloading');

  if (loading) {
    return (
      <Card className={`p-6 ${className}`}>
        <div className="flex items-center justify-center">
          <LoadingSpinner size="lg" />
          <span className="ml-3 text-gray-600">Loading Java status...</span>
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

      {/* Compatibility Status */}
      {compatibility && (
        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Java Compatibility</h3>
            <Badge
              variant={getStatusColor(compatibility.hasCompatibleJava) as any}
            >
              {compatibility.hasCompatibleJava ? 'Compatible' : 'Not Compatible'}
            </Badge>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Minecraft Version:</span>
              <span className="text-sm font-medium">{compatibility.minecraftVersion}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Required Java:</span>
              <span className="text-sm font-medium">{compatibility.requiredJavaVersion || 'Auto'}</span>
            </div>
            {compatibility.recommendedInstallation && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Recommended:</span>
                <span className="text-sm font-medium text-green-600">
                  Java {compatibility.recommendedInstallation.version}
                </span>
              </div>
            )}
          </div>

          {!compatibility.hasCompatibleJava && (
            <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <p className="text-sm text-yellow-800">
                No compatible Java installation found. Install Java {compatibility.requiredJavaVersion} to play this version.
              </p>
              <Button
                variant="primary"
                size="sm"
                className="mt-2"
                onClick={() => {
                  setSelectedVersion(compatibility.requiredJavaVersion || '21');
                  setInstallModalOpen(true);
                }}
              >
                Install Java {compatibility.requiredJavaVersion}
              </Button>
            </div>
          )}
        </Card>
      )}

      {/* All Java Installations */}
      <Card className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Java Installations</h3>
          <div className="flex space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setInstallModalOpen(true)}
            >
              Install Java
            </Button>
            {managedInstallations.length > 3 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleCleanup}
              >
                Cleanup Old
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={loadJavaStatus}
            >
              Refresh
            </Button>
          </div>
        </div>

        {installations.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <div className="text-4xl mb-3">☕</div>
            <p className="text-sm">No Java installations found</p>
            <Button
              variant="primary"
              size="sm"
              className="mt-3"
              onClick={() => {
                setSelectedVersion('21');
                setInstallModalOpen(true);
              }}
            >
              Install Java 21
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            {installations.map((installation, index) => (
              <div
                key={index}
                className={`flex items-center justify-between p-3 rounded-lg border ${
                  installation.isManaged
                    ? 'border-blue-200 bg-blue-50'
                    : 'border-gray-200 bg-gray-50'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <div className={`w-2 h-2 rounded-full ${
                    installation.is64bit ? 'bg-green-500' : 'bg-yellow-500'
                  }`} />
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="font-medium text-gray-900">
                        Java {installation.version}
                      </span>
                      {installation.isManaged && (
                        <Badge variant="primary" size="sm">Managed</Badge>
                      )}
                      {compatibility?.recommendedInstallation?.version === installation.version && (
                        <Badge variant="success" size="sm">Recommended</Badge>
                      )}
                    </div>
                    <div className="text-xs text-gray-500">
                      {installation.path}
                      <span className="ml-2">
                        {installation.is64bit ? '64-bit' : '32-bit'} •
                        {formatSize(installation.sizeBytes)}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <Tooltip content={installation.isManaged ? 'Managed installation' : 'System installation'}>
                    <span className={`text-sm ${
                      installation.isManaged ? 'text-blue-600' : 'text-gray-600'
                    }`}>
                      {installation.isManaged ? '📦' : '🖥️'}
                    </span>
                  </Tooltip>
                  {installation.isManaged && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemoveJava(installation.version)}
                    >
                      Remove
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* Managed Installations Summary */}
      {managedInstallations.length > 0 && (
        <Card className="p-4 border-blue-200 bg-blue-50">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="text-sm font-medium text-blue-900">Managed Installations</h4>
              <p className="text-xs text-blue-700">
                {managedInstallations.length} installation(s) managed by this launcher
              </p>
            </div>
            <div className="text-xs text-blue-600">
              Total: {formatSize(managedInstallations.reduce((sum, inst) => sum + (inst.sizeBytes || 0), 0))}
            </div>
          </div>
        </Card>
      )}

      {/* Install Java Modal */}
      <Modal
        isOpen={installModalOpen}
        onClose={() => setInstallModalOpen(false)}
        title="Install Java"
        size="md"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Select the Java version you want to install. The launcher will download and manage it for you.
          </p>

          <div className="grid grid-cols-2 gap-3">
            {availableVersions.map(version => (
              <Button
                key={version}
                variant={selectedVersion === version ? 'primary' : 'outline'}
                onClick={() => setSelectedVersion(version)}
                disabled={installations.some(inst => inst.version === version)}
              >
                Java {version}
                {installations.some(inst => inst.version === version) && ' ✓'}
              </Button>
            ))}
          </div>

          {selectedVersion && (
            <div className="p-3 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600">
                <strong>Java {selectedVersion}</strong> will be installed and managed by the launcher.
              </p>
            </div>
          )}

          <div className="flex justify-end space-x-3 pt-4 border-t">
            <Button
              variant="outline"
              onClick={() => setInstallModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={() => selectedVersion && handleInstallJava(selectedVersion)}
              disabled={!selectedVersion || installations.some(inst => inst.version === selectedVersion)}
            >
              Install Java {selectedVersion}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};