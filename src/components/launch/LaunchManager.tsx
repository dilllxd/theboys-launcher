import React, { useState, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import {
  LaunchedProcess,
  type ProcessStatus,
  Instance
} from '../../types/launcher';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Progress } from '../ui/Progress';
import { invoke } from '@tauri-apps/api/core';
import { toast } from 'react-hot-toast';

interface LaunchManagerProps {
  instances: Instance[];
  onInstanceUpdate?: (instanceId: string) => void;
}

const LaunchManagerContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const SectionTitle = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-md) 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
`;

const ProcessGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: var(--spacing-md);
`;

const ProcessCard = styled(Card)`
  padding: var(--spacing-lg);
  position: relative;
  overflow: hidden;
`;

const ProcessHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-md);
`;

const ProcessInfo = styled.div`
  flex: 1;
`;

const ProcessName = styled.h4`
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
`;

const ProcessDetails = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-md);
`;

const ProcessDetail = styled.span`
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const ProcessStatus = styled.div<{ status: ProcessStatus }>`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: ${props => {
      switch (props.status) {
        case 'starting':
          return 'var(--color-warning-500)';
        case 'running':
          return 'var(--color-success-500)';
        case 'finished':
          return 'var(--color-primary-500)';
        case 'crashed':
          return 'var(--color-error-500)';
        case 'killed':
          return 'var(--color-gray-500)';
        default:
          return 'var(--color-gray-400)';
      }
    }};

    animation: ${props => props.status === 'starting' || props.status === 'running'
      ? 'pulse 2s infinite'
      : 'none'};
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
`;

const ProcessActions = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
`;

const LaunchTime = styled.div`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-sm);
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);

  .icon {
    font-size: 3rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }
`;

const ProcessLog = styled.div`
  background: var(--color-gray-50);
  border: 1px solid var(--color-border-light);
  border-radius: var(--border-radius-md);
  padding: var(--spacing-sm);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  max-height: 100px;
  overflow-y: auto;
  margin-top: var(--spacing-md);

  .log-entry {
    margin-bottom: var(--spacing-xs);

    &.error {
      color: var(--color-error-600);
    }

    &.success {
      color: var(--color-success-600);
    }
  }
`;

const formatDuration = (startMs: number): string => {
  if (!startMs) return 'Unknown';

  const now = Date.now();
  const duration = now - startMs;

  if (duration < 1000) {
    return `${duration}ms`;
  } else if (duration < 60000) {
    return `${Math.round(duration / 1000)}s`;
  } else {
    const minutes = Math.floor(duration / 60000);
    const seconds = Math.round((duration % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  }
};

const formatTimestamp = (timestamp: string): string => {
  return new Date(timestamp).toLocaleTimeString();
};

const getStatusLabel = (status: ProcessStatus): string => {
  switch (status) {
    case 'starting':
      return 'Starting...';
    case 'running':
      return 'Running';
    case 'finished':
      return 'Finished';
    case 'crashed':
      return 'Crashed';
    case 'killed':
      return 'Terminated';
    default:
      return 'Unknown';
  }
};

const getStatusVariant = (status: ProcessStatus): 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'outline' => {
  switch (status) {
    case 'starting':
      return 'warning';
    case 'running':
      return 'success';
    case 'finished':
      return 'primary';
    case 'crashed':
    case 'killed':
      return 'error';
    default:
      return 'outline';
  }
};

export const LaunchManager: React.FC<LaunchManagerProps> = ({
  instances,
  onInstanceUpdate: _onInstanceUpdate,
}) => {
  const [processes, setProcesses] = useState<LaunchedProcess[]>([]);

  // Load active processes on mount
  useEffect(() => {
    loadActiveProcesses();

    // Set up periodic updates
    const interval = setInterval(loadActiveProcesses, 5000);

    return () => clearInterval(interval);
  }, []);

  const loadActiveProcesses = useCallback(async () => {
    try {
      const activeProcesses = await invoke<LaunchedProcess[]>('get_active_processes');
      setProcesses(activeProcesses);
    } catch (error) {
      console.error('Failed to load active processes:', error);
    }
  }, []);

  const handleTerminateProcess = useCallback(async (launchId: string, instanceName: string) => {
    try {
      await invoke('terminate_process', { launchId });
      toast.success(`Terminated ${instanceName}`);
      await loadActiveProcesses();
    } catch (error) {
      console.error('Failed to terminate process:', error);
      toast.error(`Failed to terminate ${instanceName}: ${error}`);
    }
  }, [loadActiveProcesses]);

  const handleForceKillInstance = useCallback(async (instanceId: string, instanceName: string) => {
    try {
      const killedCount = await invoke<number>('force_kill_instance', { instanceId });
      toast.success(`Force killed ${killedCount} processes for ${instanceName}`);
      await loadActiveProcesses();
    } catch (error) {
      console.error('Failed to force kill instance:', error);
      toast.error(`Failed to force kill ${instanceName}: ${error}`);
    }
  }, [loadActiveProcesses]);

  const handleCleanup = useCallback(async () => {
    try {
      const cleanedCount = await invoke<number>('cleanup_finished_processes');
      if (cleanedCount > 0) {
        toast.success(`Cleaned up ${cleanedCount} finished processes`);
      }
      await loadActiveProcesses();
    } catch (error) {
      console.error('Failed to cleanup processes:', error);
      toast.error(`Failed to cleanup processes: ${error}`);
    }
  }, [loadActiveProcesses]);

  
  const getInstance = (instanceId: string): Instance | undefined => {
    return instances.find(i => i.id === instanceId);
  };

  const runningProcesses = processes.filter(p =>
    p.status === 'starting' || p.status === 'running'
  );
  const finishedProcesses = processes.filter(p =>
    p.status === 'finished' || p.status === 'crashed' || p.status === 'killed'
  );

  return (
    <LaunchManagerContainer>
      {/* Running Processes */}
      {runningProcesses.length > 0 && (
        <>
          <SectionTitle>
            🎮 Running Games ({runningProcesses.length})
            <Button
              variant="outline"
              size="sm"
              onClick={handleCleanup}
            >
              Cleanup Finished
            </Button>
          </SectionTitle>

          <ProcessGrid>
            <AnimatePresence>
              {runningProcesses.map((process) => (
                <motion.div
                  key={process.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  transition={{ duration: 0.3 }}
                >
                  <ProcessCard>
                    <ProcessHeader>
                      <ProcessInfo>
                        <ProcessName>{process.instanceName}</ProcessName>
                        <ProcessDetails>
                          <ProcessDetail>
                            <span>🆔</span>
                            <span>PID: {process.pid}</span>
                          </ProcessDetail>
                          <ProcessDetail>
                            <span>⏰</span>
                            <span>Started: {formatTimestamp(process.startedAt)}</span>
                          </ProcessDetail>
                          {process.launchTimeMs && (
                            <ProcessDetail>
                              <span>🚀</span>
                              <span>Launch: {process.launchTimeMs}ms</span>
                            </ProcessDetail>
                          )}
                        </ProcessDetails>
                      </ProcessInfo>
                      <Badge variant={getStatusVariant(process.status)}>
                        {getStatusLabel(process.status)}
                      </Badge>
                    </ProcessHeader>

                    <ProcessStatus status={process.status}>
                      <div className="status-indicator" />
                      <span>
                        {process.status === 'starting' && 'Game is starting...'}
                        {process.status === 'running' && 'Game is running'}
                      </span>
                    </ProcessStatus>

                    {process.status === 'starting' && (
                      <Progress value={75} size="sm" showPercentage={false} />
                    )}

                    <LaunchTime>
                      Running for {formatDuration(new Date(process.startedAt).getTime())}
                    </LaunchTime>

                    <ProcessActions>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleTerminateProcess(process.id, process.instanceName)}
                      >
                        Terminate
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleForceKillInstance(process.instanceId, process.instanceName)}
                        style={{ color: 'var(--color-error-600)' }}
                      >
                        Force Kill
                      </Button>
                    </ProcessActions>
                  </ProcessCard>
                </motion.div>
              ))}
            </AnimatePresence>
          </ProcessGrid>
        </>
      )}

      {/* Finished Processes */}
      {finishedProcesses.length > 0 && (
        <>
          <SectionTitle>
            📊 Recent Sessions ({finishedProcesses.length})
          </SectionTitle>

          <ProcessGrid>
            {finishedProcesses.map((process) => {
              const instance = getInstance(process.instanceId);
              const hasCrashInfo = process.crashReason || (process.exitCode && process.exitCode !== 0);

              return (
                <motion.div
                  key={process.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <ProcessCard>
                    <ProcessHeader>
                      <ProcessInfo>
                        <ProcessName>{process.instanceName}</ProcessName>
                        <ProcessDetails>
                          <ProcessDetail>
                            <span>⏰</span>
                            <span>Ended: {formatTimestamp(process.startedAt)}</span>
                          </ProcessDetail>
                          {process.exitCode !== undefined && (
                            <ProcessDetail>
                              <span>🔢</span>
                              <span>Exit Code: {process.exitCode}</span>
                            </ProcessDetail>
                          )}
                        </ProcessDetails>
                      </ProcessInfo>
                      <Badge variant={getStatusVariant(process.status)}>
                        {getStatusLabel(process.status)}
                      </Badge>
                    </ProcessHeader>

                    {hasCrashInfo && (
                      <ProcessLog>
                        {process.crashReason && (
                          <div className="log-entry error">
                            {process.crashReason}
                          </div>
                        )}
                        {process.exitCode && process.exitCode !== 0 && (
                          <div className="log-entry error">
                            Game exited with non-zero code: {process.exitCode}
                          </div>
                        )}
                      </ProcessLog>
                    )}

                    {instance && (
                      <LaunchTime>
                        Total Playtime: {Math.round(instance.totalPlaytime / 60)} minutes
                      </LaunchTime>
                    )}

                    <ProcessActions>
                      {process.status === 'crashed' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleForceKillInstance(process.instanceId, process.instanceName)}
                        >
                          Cleanup
                        </Button>
                      )}
                    </ProcessActions>
                  </ProcessCard>
                </motion.div>
              );
            })}
          </ProcessGrid>
        </>
      )}

      {/* Empty State */}
      {processes.length === 0 && (
        <Card>
          <EmptyState>
            <div className="icon">🎮</div>
            <h3>No Active Games</h3>
            <p>Launch an instance to see it appear here with real-time status updates.</p>
          </EmptyState>
        </Card>
      )}
    </LaunchManagerContainer>
  );
};