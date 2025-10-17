import React from 'react';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { Card } from './ui/Card';
import { Button } from './ui/Button';
import { Progress } from './ui/Progress';
import { Badge } from './ui/Badge';
import { Tooltip } from './ui/Tooltip';
import { useDownloads } from '../hooks/useDownloads';
import { DownloadStatus } from '../types/launcher';
import {
  Download,
  Pause,
  Play,
  X,
  Trash2,
  RefreshCw,
  CheckCircle,
  XCircle,
  Clock,
  HardDrive,
  Zap,
  AlertCircle
} from 'lucide-react';

const DownloadManagerContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  height: 100%;
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
`;

const Title = styled.h2`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
`;

const DownloadList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  flex: 1;
  overflow-y: auto;
  min-height: 0;
`;

const DownloadItem = styled(motion.div)`
  margin-bottom: var(--spacing-sm);
`;

const DownloadHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-sm);
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
`;

const DownloadActions = styled.div`
  display: flex;
  gap: var(--spacing-xs);
  align-items: center;
`;

const ProgressSection = styled.div`
  margin: var(--spacing-sm) 0;
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

const EmptyState = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--color-text-secondary);
  text-align: center;
  gap: var(--spacing-md);
`;

const EmptyStateIcon = styled.div`
  font-size: 3rem;
  opacity: 0.5;
`;

const StatsContainer = styled.div`
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-sm) 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
`;

const Stat = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const ErrorContainer = styled.div`
  background-color: var(--color-error);
  color: white;
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-md);
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const ErrorMessage = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
`;

const ErrorCloseButton = styled.button`
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: var(--spacing-xs);
  border-radius: var(--radius-sm);

  &:hover {
    background-color: rgba(255, 255, 255, 0.2);
  }
`;

interface DownloadManagerProps {
  className?: string;
}

export const DownloadManager: React.FC<DownloadManagerProps> = ({ className }) => {
  const {
    downloads,
    isLoading,
    error,
    loadDownloads,
    pauseDownload,
    resumeDownload,
    cancelDownload,
    removeDownload,
    clearError,
  } = useDownloads();

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
        return <Download size={16} />;
      case 'paused':
        return <Pause size={16} />;
      case 'completed':
        return <CheckCircle size={16} />;
      case 'cancelled':
        return <XCircle size={16} />;
      case 'pending':
        return <Clock size={16} />;
      default:
        if (typeof status === 'object' && status.failed) {
          return <XCircle size={16} />;
        }
        return <Clock size={16} />;
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

  const handlePauseResume = async (downloadId: string, status: DownloadStatus) => {
    if (status === 'downloading') {
      await pauseDownload(downloadId);
    } else if (status === 'paused') {
      await resumeDownload(downloadId);
    }
  };

  const handleCancel = async (downloadId: string) => {
    await cancelDownload(downloadId);
  };

  const handleRemove = async (downloadId: string) => {
    await removeDownload(downloadId);
  };

  const getActiveDownloadsCount = () => {
    return downloads.filter(d => d.status === 'downloading').length;
  };

  const getCompletedDownloadsCount = () => {
    return downloads.filter(d => d.status === 'completed').length;
  };

  if (isLoading) {
    return (
      <DownloadManagerContainer className={className}>
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          Loading downloads...
        </div>
      </DownloadManagerContainer>
    );
  }

  return (
    <DownloadManagerContainer className={className}>
      <Header>
        <Title>
          <Download size={20} />
          Downloads
        </Title>
        <Button variant="ghost" size="sm" onClick={loadDownloads}>
          <RefreshCw size={16} />
        </Button>
      </Header>

      {error && (
        <ErrorContainer>
          <ErrorMessage>
            <AlertCircle size={16} />
            {error}
          </ErrorMessage>
          <ErrorCloseButton onClick={clearError}>
            <X size={16} />
          </ErrorCloseButton>
        </ErrorContainer>
      )}

      <StatsContainer>
        <Stat>
          <Zap size={14} />
          Active: {getActiveDownloadsCount()}
        </Stat>
        <Stat>
          <CheckCircle size={14} />
          Completed: {getCompletedDownloadsCount()}
        </Stat>
        <Stat>
          <HardDrive size={14} />
          Total: {downloads.length}
        </Stat>
      </StatsContainer>

      <DownloadList>
        <AnimatePresence>
          {downloads.map((download) => (
            <DownloadItem
              key={download.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ duration: 0.2 }}
            >
              <Card>
                <DownloadHeader>
                  <DownloadInfo>
                    <DownloadName>
                      {getStatusIcon(download.status)}
                      {download.name}
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
                  <DownloadActions>
                    {download.status === 'downloading' && (
                      <Tooltip content="Pause">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handlePauseResume(download.id, download.status)}
                        >
                          <Pause size={16} />
                        </Button>
                      </Tooltip>
                    )}
                    {download.status === 'paused' && (
                      <Tooltip content="Resume">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handlePauseResume(download.id, download.status)}
                        >
                          <Play size={16} />
                        </Button>
                      </Tooltip>
                    )}
                    {(download.status === 'downloading' || download.status === 'paused') && (
                      <Tooltip content="Cancel">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleCancel(download.id)}
                        >
                          <X size={16} />
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
                          onClick={() => handleRemove(download.id)}
                        >
                          <Trash2 size={16} />
                        </Button>
                      </Tooltip>
                    )}
                  </DownloadActions>
                </DownloadHeader>

                {download.status === 'downloading' || download.status === 'paused' ? (
                  <ProgressSection>
                    <Progress
                      value={download.progressPercent}
                      color={getProgressColor(download.status)}
                      showPercentage={true}
                      size="sm"
                    />
                  </ProgressSection>
                ) : download.status === 'completed' ? (
                  <ProgressSection>
                    <Progress
                      value={100}
                      color="success"
                      showPercentage={true}
                      size="sm"
                    />
                  </ProgressSection>
                ) : null}
              </Card>
            </DownloadItem>
          ))}
        </AnimatePresence>

        {downloads.length === 0 && (
          <EmptyState>
            <EmptyStateIcon>
              <Download size={48} />
            </EmptyStateIcon>
            <div>No downloads yet</div>
            <div style={{ fontSize: 'var(--font-size-sm)' }}>
              Downloads will appear here when you start downloading modpacks, tools, or resources
            </div>
          </EmptyState>
        )}
      </DownloadList>
    </DownloadManagerContainer>
  );
};

export default DownloadManager;