import React from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { Progress } from './ui/Progress';
import { Badge } from './ui/Badge';
import { Tooltip } from './ui/Tooltip';
import { Button } from './ui/Button';
import { DownloadProgress as DownloadProgressType, DownloadStatus } from '../types/launcher';
import {
  Download,
  Pause,
  Play,
  X,
  CheckCircle,
  XCircle,
  Clock
} from 'lucide-react';

const DownloadProgressContainer = styled(motion.div)`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  background-color: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
`;

const DownloadHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-sm);
`;

const DownloadInfo = styled.div`
  flex: 1;
  min-width: 0;
`;

const DownloadName = styled.div`
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-xs);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const DownloadDetails = styled.div`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
`;

const DownloadActions = styled.div`
  display: flex;
  gap: var(--spacing-xs);
  align-items: center;
  flex-shrink: 0;
`;

const StatusBadge = styled(Badge)<{ status: DownloadStatus }>`
  ${({ status }) => {
    switch (status) {
      case 'downloading':
        return 'background-color: var(--color-primary); color: white;';
      case 'paused':
        return 'background-color: var(--color-warning); color: white;';
      case 'completed':
        return 'background-color: var(--color-success); color: white;';
      case 'cancelled':
        return 'background-color: var(--color-bg-tertiary); color: var(--color-text-secondary);';
      case 'pending':
        return 'background-color: var(--color-bg-secondary); color: var(--color-text-secondary);';
      default:
        if (typeof status === 'object' && status.failed) {
          return 'background-color: var(--color-error); color: white;';
        }
        return 'background-color: var(--color-bg-secondary); color: var(--color-text-secondary);';
    }
  }}
`;

interface DownloadProgressProps {
  download: DownloadProgressType;
  onPause?: (downloadId: string) => void;
  onResume?: (downloadId: string) => void;
  onCancel?: (downloadId: string) => void;
  onRemove?: (downloadId: string) => void;
  showActions?: boolean;
  compact?: boolean;
  className?: string;
}

export const DownloadProgressComponent: React.FC<DownloadProgressProps> = ({
  download,
  onPause,
  onResume,
  onCancel,
  onRemove,
  showActions = true,
  compact = false,
  className
}) => {
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatSpeed = (bps: number): string => {
    return formatBytes(bps) + '/s';
  };

  const getStatusIcon = (status: DownloadStatus) => {
    switch (status) {
      case 'downloading':
        return <Download size={compact ? 14 : 16} />;
      case 'paused':
        return <Pause size={compact ? 14 : 16} />;
      case 'completed':
        return <CheckCircle size={compact ? 14 : 16} />;
      case 'cancelled':
        return <XCircle size={compact ? 14 : 16} />;
      case 'pending':
        return <Clock size={compact ? 14 : 16} />;
      default:
        if (typeof status === 'object' && status.failed) {
          return <XCircle size={compact ? 14 : 16} />;
        }
        return <Clock size={compact ? 14 : 16} />;
    }
  };

  const getStatusText = (status: DownloadStatus): string => {
    switch (status) {
      case 'downloading':
        return 'Downloading';
      case 'paused':
        return 'Paused';
      case 'completed':
        return 'Completed';
      case 'cancelled':
        return 'Cancelled';
      case 'pending':
        return 'Pending';
      default:
        if (typeof status === 'object' && status.failed) {
          return `Failed: ${status.failed}`;
        }
        return 'Unknown';
    }
  };

  const getProgressColor = (status: DownloadStatus): 'primary' | 'success' | 'warning' | 'error' => {
    switch (status) {
      case 'downloading':
        return 'primary';
      case 'paused':
        return 'warning';
      case 'completed':
        return 'success';
      case 'cancelled':
        return 'error';
      default:
        if (typeof status === 'object' && status.failed) {
          return 'error';
        }
        return 'primary';
    }
  };

  const handlePauseResume = () => {
    if (download.status === 'downloading' && onPause) {
      onPause(download.id);
    } else if (download.status === 'paused' && onResume) {
      onResume(download.id);
    }
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel(download.id);
    }
  };

  const handleRemove = () => {
    if (onRemove) {
      onRemove(download.id);
    }
  };

  return (
    <DownloadProgressContainer
      className={className}
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      transition={{ duration: 0.2 }}
    >
      <DownloadHeader>
        <DownloadInfo>
          <DownloadName>
            {getStatusIcon(download.status)}
            <span style={{
              fontSize: compact ? 'var(--font-size-sm)' : 'var(--font-size-base)'
            }}>
              {download.name}
            </span>
            <StatusBadge status={download.status} size="sm">
              {getStatusText(download.status)}
            </StatusBadge>
          </DownloadName>
          <DownloadDetails>
            <span>{formatBytes(download.downloadedBytes)}</span>
            {download.totalBytes > 0 && (
              <span>/ {formatBytes(download.totalBytes)}</span>
            )}
            {download.speedBps > 0 && (
              <span>• {formatSpeed(download.speedBps)}</span>
            )}
          </DownloadDetails>
        </DownloadInfo>

        {showActions && (
          <DownloadActions>
            {download.status === 'downloading' && (
              <Tooltip content="Pause">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handlePauseResume}
                >
                  <Pause size={compact ? 14 : 16} />
                </Button>
              </Tooltip>
            )}
            {download.status === 'paused' && (
              <Tooltip content="Resume">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handlePauseResume}
                >
                  <Play size={compact ? 14 : 16} />
                </Button>
              </Tooltip>
            )}
            {(download.status === 'downloading' || download.status === 'paused') && (
              <Tooltip content="Cancel">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleCancel}
                >
                  <X size={compact ? 14 : 16} />
                </Button>
              </Tooltip>
            )}
            {(download.status === 'completed' ||
              download.status === 'cancelled' ||
              (typeof download.status === 'object' && download.status.failed)) && (
              <Tooltip content="Remove">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleRemove}
                >
                  <X size={compact ? 14 : 16} />
                </Button>
              </Tooltip>
            )}
          </DownloadActions>
        )}
      </DownloadHeader>

      {(download.status === 'downloading' || download.status === 'paused') ? (
        <Progress
          value={download.progressPercent}
          color={getProgressColor(download.status)}
          showPercentage={!compact}
          size={compact ? 'sm' : 'md'}
        />
      ) : download.status === 'completed' ? (
        <Progress
          value={100}
          color="success"
          showPercentage={!compact}
          size={compact ? 'sm' : 'md'}
        />
      ) : null}
    </DownloadProgressContainer>
  );
};

export default DownloadProgressComponent;