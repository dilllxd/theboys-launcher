import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { X, Download } from 'lucide-react';
import type { UpdateInfo } from '../../types/launcher';
import { formatFileSize, formatTimeAgo } from '../../utils/format';

interface UpdateNotificationProps {
  onUpdateAvailable?: (updateInfo: UpdateInfo) => void;
  onDismiss?: () => void;
  className?: string;
}

export const UpdateNotification: React.FC<UpdateNotificationProps> = ({
  onUpdateAvailable,
  onDismiss,
  className = ''
}) => {
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [hasChecked, setHasChecked] = useState(false);

  // Check for updates on component mount
  useEffect(() => {
    checkForUpdates();
  }, []);

  // Auto-hide notification after 30 seconds
  useEffect(() => {
    if (isVisible && updateInfo) {
      const timer = setTimeout(() => {
        setIsVisible(false);
        onDismiss?.();
      }, 30000);

      return () => clearTimeout(timer);
    }
  }, [isVisible, updateInfo, onDismiss]);

  const checkForUpdates = async () => {
    if (hasChecked) return;

    setIsLoading(true);
    try {
      const update = await invoke<UpdateInfo | null>('check_for_updates');
      if (update) {
        setUpdateInfo(update);
        setIsVisible(true);
        onUpdateAvailable?.(update);
      }
    } catch (error) {
      console.error('Failed to check for updates:', error);
    } finally {
      setIsLoading(false);
      setHasChecked(true);
    }
  };

  const handleDownload = () => {
    if (updateInfo) {
      onUpdateAvailable?.(updateInfo);
      setIsVisible(false);
    }
  };

  const handleDismiss = () => {
    setIsVisible(false);
    onDismiss?.();
  };

  const getChannelColor = (channel: string) => {
    switch (channel) {
      case 'stable': return 'bg-green-100 text-green-800 border-green-200';
      case 'beta': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'alpha': return 'bg-red-100 text-red-800 border-red-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  if (isLoading) {
    return null; // Don't show anything while loading
  }

  if (!isVisible || !updateInfo) {
    return null;
  }

  return (
    <div className={`fixed top-4 right-4 z-50 max-w-md animate-in slide-in-from-top duration-300 ${className}`}>
      <Card className="border-l-4 border-l-blue-500 shadow-lg">
        <div className="p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center gap-2">
              <div className="flex items-center justify-center w-8 h-8 bg-blue-100 rounded-full">
                <Download className="w-4 h-4 text-blue-600" />
              </div>
              <div>
                <h4 className="font-semibold text-sm">Update Available</h4>
                <p className="text-xs text-gray-600">TheBoys Launcher</p>
              </div>
            </div>
            <button
              onClick={handleDismiss}
              className="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Badge className={getChannelColor(updateInfo.channel)}>
                {updateInfo.channel.toUpperCase()}
              </Badge>
              <span className="text-sm font-medium">v{updateInfo.version}</span>
              <span className="text-xs text-gray-500">
                {formatFileSize(updateInfo.fileSize)}
              </span>
            </div>

            <div className="text-sm text-gray-700">
              <p className="line-clamp-2">
                A new version of the launcher is available with improvements and bug fixes.
              </p>
              {updateInfo.releaseNotes && (
                <details className="mt-2">
                  <summary className="cursor-pointer text-blue-600 hover:text-blue-800 text-xs">
                    View release notes
                  </summary>
                  <div className="mt-2 p-2 bg-gray-50 rounded text-xs max-h-24 overflow-y-auto">
                    <pre className="whitespace-pre-wrap font-sans">
                      {updateInfo.releaseNotes}
                    </pre>
                  </div>
                </details>
              )}
            </div>

            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">
                Released {formatTimeAgo(updateInfo.publishedAt)}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDismiss}
                >
                  Later
                </Button>
                <Button
                  size="sm"
                  onClick={handleDownload}
                >
                  Download
                </Button>
              </div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
};

// Minimal update notification for header
export const MinimalUpdateNotification: React.FC<{
  updateInfo: UpdateInfo;
  onClick?: () => void;
  className?: string;
}> = ({ updateInfo, onClick, className = '' }) => {
  return (
    <div
      className={`flex items-center gap-2 px-3 py-1.5 bg-blue-100 text-blue-800 rounded-full text-xs font-medium cursor-pointer hover:bg-blue-200 transition-colors ${className}`}
      onClick={onClick}
    >
      <Download className="w-3 h-3" />
      <span>Update v{updateInfo.version}</span>
      <span className="bg-blue-200 px-1.5 py-0.5 rounded-full text-xs">
        {formatFileSize(updateInfo.fileSize)}
      </span>
    </div>
  );
};

// Update status indicator for header
export const UpdateStatusIndicator: React.FC<{
  className?: string;
  onSettingsClick?: () => void;
}> = ({ className = '', onSettingsClick }) => {
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    checkForUpdates();
  }, []);

  const checkForUpdates = async () => {
    setIsLoading(true);
    try {
      const update = await invoke<UpdateInfo | null>('check_for_updates');
      setUpdateInfo(update);
    } catch (error) {
      console.error('Failed to check for updates:', error);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className={`flex items-center gap-2 px-3 py-1.5 bg-gray-100 text-gray-600 rounded-full text-xs ${className}`}>
        <div className="w-3 h-3 border-2 border-gray-300 border-t-gray-600 rounded-full animate-spin" />
        <span>Checking...</span>
      </div>
    );
  }

  if (updateInfo) {
    return (
      <MinimalUpdateNotification
        updateInfo={updateInfo}
        onClick={onSettingsClick}
        className={className}
      />
    );
  }

  return (
    <div className={`flex items-center gap-2 px-3 py-1.5 bg-green-100 text-green-800 rounded-full text-xs ${className}`}>
      <span className="w-2 h-2 bg-green-600 rounded-full" />
      <span>Up to date</span>
    </div>
  );
};