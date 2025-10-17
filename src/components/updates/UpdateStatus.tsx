import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { Tooltip } from '../ui/Tooltip';
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

interface UpdateStatusProps {
  instanceId?: string;
  compact?: boolean;
  showCheckButton?: boolean;
}

export const UpdateStatus: React.FC<UpdateStatusProps> = ({
  instanceId,
  compact = false,
  showCheckButton = true
}) => {
  const [updates, setUpdates] = useState<ModpackUpdate[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [lastChecked, setLastChecked] = useState<Date | null>(null);

  useEffect(() => {
    if (instanceId) {
      checkForUpdates();
    }
  }, [instanceId]);

  const checkForUpdates = async () => {
    if (!instanceId) return;

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
      setLastChecked(new Date());
    } catch (error) {
      console.error('Failed to check for updates:', error);
    } finally {
      setIsLoading(false);
    }
  };

  if (compact) {
    return (
      <div className="flex items-center gap-2">
        {isLoading ? (
          <LoadingSpinner size="sm" />
        ) : updates.length > 0 ? (
          <Tooltip content={`${updates.length} update(s) available`}>
            <Badge variant="success" className="cursor-pointer">
              🔄 {updates.length}
            </Badge>
          </Tooltip>
        ) : (
          <Badge variant="outline">
            ✅ Up to date
          </Badge>
        )}
        {showCheckButton && (
          <Button
            variant="ghost"
            size="sm"
            onClick={checkForUpdates}
            disabled={isLoading}
          >
            🔄
          </Button>
        )}
      </div>
    );
  }

  return (
    <Card className="update-status">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="text-2xl">
            {isLoading ? '🔄' : updates.length > 0 ? '⬆️' : '✅'}
          </div>
          <div>
            <h3 className="font-semibold">Updates</h3>
            <p className="text-sm text-gray-600">
              {isLoading ? 'Checking...' :
               updates.length > 0 ?
               `${updates.length} update(s) available` :
               'Up to date'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {lastChecked && (
            <span className="text-xs text-gray-500">
              Checked {formatTimeAgo(lastChecked.toISOString())}
            </span>
          )}
          {showCheckButton && (
            <Button
              variant="outline"
              size="sm"
              onClick={checkForUpdates}
              disabled={isLoading}
            >
              {isLoading ? <LoadingSpinner size="sm" /> : 'Check'}
            </Button>
          )}
        </div>
      </div>

      {updates.length > 0 && !compact && (
        <div className="mt-4 space-y-2 border-t pt-4">
          {updates.map((update, index) => (
            <div key={index} className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-2">
                <span className="font-medium">{update.modpack_id}</span>
                <Badge variant="success">
                  {update.current_version} → {update.latest_version}
                </Badge>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-gray-500">{formatFileSize(update.size_bytes)}</span>
                {update.changelog_url && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => window.open(update.changelog_url, '_blank')}
                  >
                    📄
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
};