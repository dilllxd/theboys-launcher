import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Progress } from '../ui/Progress';
import { Badge } from '../ui/Badge';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { Tooltip } from '../ui/Tooltip';
import { Modal } from '../ui/Modal';
import { formatFileSize, formatTimeAgo } from '../../utils/format';

interface ModpackUpdate {
  modpack_id: string;
  current_version: string;
  latest_version: string;
  update_available: boolean;
  changelog_url?: string;
  download_url: string;
  size_bytes: number;
}

interface BackupInfo {
  id: string;
  instance_id: string;
  backup_date: string;
  version: string;
  size_bytes: number;
  backup_path: string;
  description?: string;
}

interface PackInstallProgress {
  instance_id: string;
  modpack_id: string;
  step: 'Downloading' | 'Extracting' | 'ParsingManifest' | 'DownloadingDependencies' | 'InstallingFiles' | 'ConfiguringInstance' | 'Completed';
  progress_percent: number;
  message: string;
  status: 'Running' | 'Completed' | 'Failed' | 'Cancelled';
}

interface ManualDownload {
  id: string;
  filename: string;
  url: string;
  checksum?: string;
  size: number;
  download_type: string;
  instructions?: string;
}

interface UpdateOptions {
  create_backup: boolean;
  backup_description?: string;
  force_update: boolean;
  allow_downgrade: boolean;
}

export const UpdateManager: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [updates, setUpdates] = useState<ModpackUpdate[]>([]);
  const [backups, setBackups] = useState<BackupInfo[]>([]);
  const [installProgress, setInstallProgress] = useState<Map<string, PackInstallProgress>>(new Map());
  const [manualDownloads, setManualDownloads] = useState<Map<string, ManualDownload[]>>(new Map());
  const [isLoading, setIsLoading] = useState(false);
  const [showBackupModal, setShowBackupModal] = useState(false);
  const [showRestoreModal, setShowRestoreModal] = useState(false);
  const [showManualDownloadsModal, setShowManualDownloadsModal] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState<BackupInfo | null>(null);
  const [currentInstall, setCurrentInstall] = useState<string | null>(null);
  const [backupDescription, setBackupDescription] = useState('');

  // Check for updates on component mount
  useEffect(() => {
    checkForUpdates();
    loadBackups();

    // Listen for progress updates
    const unlistenProgress = listen('install-progress', (event) => {
      const progress = event.payload as PackInstallProgress;
      setInstallProgress(prev => new Map(prev).set(progress.instance_id, progress));
    });

    return () => {
      unlistenProgress.then(fn => fn());
    };
  }, [instanceId]);

  const checkForUpdates = async () => {
    setIsLoading(true);
    try {
      const update = await invoke<ModpackUpdate | null>('check_instance_updates', {
        instanceId
      });

      if (update) {
        setUpdates([update]);
      } else {
        setUpdates([]);
      }
    } catch (error) {
      console.error('Failed to check for updates:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadBackups = async () => {
    try {
      const backups = await invoke<BackupInfo[]>('get_instance_backups', {
        instanceId
      });
      setBackups(backups);
    } catch (error) {
      console.error('Failed to load backups:', error);
    }
  };

  const createBackup = async () => {
    try {
      const backup = await invoke<BackupInfo>('create_instance_backup', {
        instanceId,
        description: backupDescription || undefined
      });

      setBackups(prev => [backup, ...prev]);
      setShowBackupModal(false);
      setBackupDescription('');
    } catch (error) {
      console.error('Failed to create backup:', error);
    }
  };

  const restoreBackup = async (backupId: string) => {
    try {
      await invoke('restore_instance_backup', {
        backupId,
        instanceId
      });

      setShowRestoreModal(false);
      setSelectedBackup(null);
      loadBackups();
    } catch (error) {
      console.error('Failed to restore backup:', error);
    }
  };

  const deleteBackup = async (backupId: string) => {
    try {
      await invoke('delete_instance_backup', { backupId });
      setBackups(prev => prev.filter(b => b.id !== backupId));
    } catch (error) {
      console.error('Failed to delete backup:', error);
    }
  };

  const startUpdate = async (update: ModpackUpdate, options: UpdateOptions) => {
    try {
      const installId = await invoke<string>('install_modpack_with_packwiz', {
        instanceId,
        packUrl: update.download_url,
        options
      });

      setCurrentInstall(installId);

      // Monitor progress
      const progressInterval = setInterval(async () => {
        try {
          const progress = await invoke<PackInstallProgress | null>('get_pack_install_progress', {
            installId
          });

          if (progress) {
            setInstallProgress(prev => new Map(prev).set(instanceId, progress));

            if (progress.status === 'Completed' || progress.status === 'Failed' || progress.status === 'Cancelled') {
              clearInterval(progressInterval);
              setCurrentInstall(null);

              // Check for manual downloads if failed
              if (progress.status === 'Failed') {
                const manualDls = await invoke<ManualDownload[]>('get_manual_downloads', { installId });
                if (manualDls.length > 0) {
                  setManualDownloads(prev => new Map(prev).set(installId, manualDls));
                  setShowManualDownloadsModal(true);
                }
              }
            }
          }
        } catch (error) {
          console.error('Failed to get progress:', error);
        }
      }, 1000);

    } catch (error) {
      console.error('Failed to start update:', error);
    }
  };

  const confirmManualDownload = async (installId: string, filename: string, localPath: string) => {
    try {
      await invoke('confirm_manual_download', {
        installId,
        filename,
        localPath
      });

      // Refresh manual downloads list
      const manualDls = await invoke<ManualDownload[]>('get_manual_downloads', { installId });
      setManualDownloads(prev => new Map(prev).set(installId, manualDls));
    } catch (error) {
      console.error('Failed to confirm manual download:', error);
    }
  };

  const cancelInstall = async (installId: string) => {
    try {
      await invoke('cancel_modpack_installation', { installId });
      setCurrentInstall(null);
      setInstallProgress(prev => {
        const newMap = new Map(prev);
        newMap.delete(instanceId);
        return newMap;
      });
    } catch (error) {
      console.error('Failed to cancel installation:', error);
    }
  };

  const getStepIcon = (step: string) => {
    switch (step) {
      case 'Downloading': return '⬇️';
      case 'Extracting': return '📦';
      case 'ParsingManifest': return '📋';
      case 'DownloadingDependencies': return '🔗';
      case 'InstallingFiles': return '📁';
      case 'ConfiguringInstance': return '⚙️';
      case 'Completed': return '✅';
      default: return '⏳';
    }
  };

  const progress = installProgress.get(instanceId);
  const manualDls = currentInstall ? manualDownloads.get(currentInstall) : [];

  return (
    <div className="update-manager">
      {/* Update Status Card */}
      <Card className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Updates</h3>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={checkForUpdates}
              disabled={isLoading}
            >
              {isLoading ? <LoadingSpinner size="sm" /> : '🔄'} Check Updates
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowBackupModal(true)}
            >
              💾 Create Backup
            </Button>
          </div>
        </div>

        {updates.length > 0 ? (
          <div className="space-y-4">
            {updates.map((update, index) => (
              <div key={index} className="border rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Badge variant="success">Update Available</Badge>
                    <span className="text-sm text-gray-600">
                      {update.current_version} → {update.latest_version}
                    </span>
                  </div>
                  <span className="text-sm text-gray-500">
                    {formatFileSize(update.size_bytes)}
                  </span>
                </div>

                <div className="mb-3">
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <span>📦 {update.modpack_id}</span>
                    {update.changelog_url && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => window.open(update.changelog_url, '_blank')}
                      >
                        📄 Changelog
                      </Button>
                    )}
                  </div>
                </div>

                {progress && progress.modpack_id === update.modpack_id ? (
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm">
                      <span>{getStepIcon(progress.step)}</span>
                      <span>{progress.message}</span>
                      {progress.status === 'Running' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => currentInstall && cancelInstall(currentInstall)}
                        >
                          ❌ Cancel
                        </Button>
                      )}
                    </div>
                    <Progress value={progress.progress_percent} />
                    <div className="text-xs text-gray-500 text-center">
                      {Math.round(progress.progress_percent)}% - {progress.step}
                    </div>
                  </div>
                ) : (
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={() => startUpdate(update, {
                        create_backup: true,
                        backup_description: `Before update to ${update.latest_version}`,
                        force_update: false,
                        allow_downgrade: false
                      })}
                    >
                      Update with Backup
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => startUpdate(update, {
                        create_backup: false,
                        force_update: false,
                        allow_downgrade: false
                      })}
                    >
                      Update Directly
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            {isLoading ? (
              <div className="flex items-center justify-center gap-2">
                <LoadingSpinner size="sm" />
                <span>Checking for updates...</span>
              </div>
            ) : (
              <div>
                <span className="text-4xl mb-2 block">✅</span>
                <span>No updates available</span>
              </div>
            )}
          </div>
        )}
      </Card>

      {/* Backups Card */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Backups</h3>
          <span className="text-sm text-gray-500">{backups.length} backups</span>
        </div>

        {backups.length > 0 ? (
          <div className="space-y-3">
            {backups.map((backup) => (
              <div key={backup.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium">v{backup.version}</span>
                    <Badge variant="outline">{formatFileSize(backup.size_bytes)}</Badge>
                  </div>
                  <div className="text-sm text-gray-500">
                    {formatTimeAgo(backup.backup_date)}
                    {backup.description && (
                      <span className="block text-xs">{backup.description}</span>
                    )}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSelectedBackup(backup);
                      setShowRestoreModal(true);
                    }}
                  >
                    🔄 Restore
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => deleteBackup(backup.id)}
                  >
                    🗑️ Delete
                  </Button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <span className="text-4xl mb-2 block">💾</span>
            <span>No backups available</span>
          </div>
        )}
      </Card>

      {/* Backup Modal */}
      <Modal
        isOpen={showBackupModal}
        onClose={() => setShowBackupModal(false)}
        title="Create Backup"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Backup Description (optional)
            </label>
            <input
              type="text"
              className="w-full p-2 border rounded-lg"
              placeholder="e.g., Before major update"
              value={backupDescription}
              onChange={(e) => setBackupDescription(e.target.value)}
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowBackupModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={createBackup}>
              Create Backup
            </Button>
          </div>
        </div>
      </Modal>

      {/* Restore Modal */}
      <Modal
        isOpen={showRestoreModal}
        onClose={() => setShowRestoreModal(false)}
        title="Restore Backup"
      >
        {selectedBackup && (
          <div className="space-y-4">
            <div className="p-3 border rounded-lg bg-yellow-50">
              <p className="text-sm text-yellow-800">
                ⚠️ Restoring this backup will replace the current instance state.
              </p>
            </div>
            <div>
              <p className="text-sm"><strong>Version:</strong> {selectedBackup.version}</p>
              <p className="text-sm"><strong>Size:</strong> {formatFileSize(selectedBackup.size_bytes)}</p>
              <p className="text-sm"><strong>Created:</strong> {formatTimeAgo(selectedBackup.backup_date)}</p>
              {selectedBackup.description && (
                <p className="text-sm"><strong>Description:</strong> {selectedBackup.description}</p>
              )}
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => setShowRestoreModal(false)}
              >
                Cancel
              </Button>
              <Button onClick={() => restoreBackup(selectedBackup.id)}>
                Restore Backup
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Manual Downloads Modal */}
      <Modal
        isOpen={showManualDownloadsModal}
        onClose={() => setShowManualDownloadsModal(false)}
        title="Manual Downloads Required"
      >
        {manualDls && manualDls.length > 0 && (
          <div className="space-y-4">
            <div className="p-3 border rounded-lg bg-blue-50">
              <p className="text-sm text-blue-800">
                📥 Some files require manual download. Please download them from the provided links
                and select the downloaded files.
              </p>
            </div>
            <div className="space-y-3">
              {manualDls.map((download, index) => (
                <div key={index} className="border rounded-lg p-3">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium">{download.filename}</span>
                    <Badge variant="outline">{formatFileSize(download.size)}</Badge>
                  </div>
                  <div className="flex gap-2 mb-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => window.open(download.url, '_blank')}
                    >
                      🔗 Download Link
                    </Button>
                    {download.instructions && (
                      <Tooltip content={download.instructions}>
                        <Button variant="outline" size="sm">
                          ℹ️ Instructions
                        </Button>
                      </Tooltip>
                    )}
                  </div>
                  <input
                    type="file"
                    className="w-full p-2 border rounded text-sm"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      if (file && currentInstall) {
                        confirmManualDownload(currentInstall, download.filename, file.name);
                      }
                    }}
                  />
                </div>
              ))}
            </div>
            <div className="flex justify-end">
              <Button onClick={() => setShowManualDownloadsModal(false)}>
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};