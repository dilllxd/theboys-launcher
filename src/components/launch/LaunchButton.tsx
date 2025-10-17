import React, { useState, useEffect } from 'react';
import styled, { keyframes } from 'styled-components';
import { motion } from 'framer-motion';
import { Button } from '../ui/Button';
import { Instance } from '../../types/launcher';
import { invoke } from '@tauri-apps/api/core';
import { toast } from 'react-hot-toast';

interface LaunchButtonProps {
  instance: Instance;
  onLaunchComplete?: (launchId: string) => void;
  onError?: (error: string) => void;
  disabled?: boolean;
  size?: 'sm' | 'md' | 'lg';
  showStatus?: boolean;
}

const pulseAnimation = keyframes`
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.05); opacity: 0.8; }
`;

const LaunchButtonContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
`;

const LaunchButtonWrapper = styled(motion.div)<{ $launching: boolean }>`
  position: relative;

  ${props => props.$launching && `
    animation: ${pulseAnimation} 2s infinite;
  `}
`;

const LaunchProgress = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
  overflow: hidden;

  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg,
      transparent,
      rgba(255, 255, 255, 0.3),
      transparent
    );
    animation: shimmer 2s infinite;
  }
`;


const LaunchIcon = styled.span<{ $spinning?: boolean }>`
  display: inline-block;
  margin-right: var(--spacing-xs);

  ${props => props.$spinning && `
    animation: spin 1s linear infinite;
  `}
`;


const LaunchStatus = styled.div`
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const LaunchMetrics = styled.div`
  display: flex;
  gap: var(--spacing-md);
  margin-top: var(--spacing-xs);
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
`;

const Metric = styled.span`
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const StatusIndicator = styled.div<{ $status: 'idle' | 'launching' | 'running' | 'error' }>`
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: ${props => {
    switch (props.$status) {
      case 'idle':
        return 'var(--color-gray-400)';
      case 'launching':
        return 'var(--color-warning-500)';
      case 'running':
        return 'var(--color-success-500)';
      case 'error':
        return 'var(--color-error-500)';
      default:
        return 'var(--color-gray-400)';
    }
  }};
`;

const LaunchButton: React.FC<LaunchButtonProps> = ({
  instance,
  onLaunchComplete,
  onError,
  disabled = false,
  size = 'md',
  showStatus = true,
}) => {
  const [launching, setLaunching] = useState(false);
  const [launchTime, setLaunchTime] = useState<number | null>(null);
  const [_launchId, setLaunchId] = useState<string | null>(null);
  const [launchStatus, setLaunchStatus] = useState<'idle' | 'launching' | 'running' | 'error'>('idle');

  // Check if instance is already running
  useEffect(() => {
    // This would be enhanced with real-time status updates
    if (instance.status === 'running') {
      setLaunchStatus('running');
    }
  }, [instance.status]);

  const handleLaunch = async () => {
    if (launching || disabled) return;

    setLaunching(true);
    setLaunchStatus('launching');
    const startTime = Date.now();

    try {
      toast.loading(`Launching ${instance.name}...`, { id: 'launch-progress' });

      const newLaunchId = await invoke<string>('launch_instance', {
        instanceId: instance.id
      });

      setLaunchId(newLaunchId);

      // Simulate launch progress (in real implementation, this would be from backend updates)
      let progress = 0;
      const progressInterval = setInterval(() => {
        progress += 10;

        if (progress >= 100) {
          clearInterval(progressInterval);
          const totalTime = Date.now() - startTime;
          setLaunchTime(totalTime);
          setLaunching(false);
          setLaunchStatus('running');

          toast.success(`${instance.name} launched successfully!`, {
            id: 'launch-progress',
            duration: 3000,
          });

          onLaunchComplete?.(newLaunchId);
        }
      }, 200);

    } catch (error) {
      console.error('Launch failed:', error);
      setLaunching(false);
      setLaunchStatus('error');

      toast.error(`Failed to launch ${instance.name}: ${error}`, {
        id: 'launch-progress',
      });

      onError?.(error as string);
    }
  };

  const canLaunch = instance.status === 'ready' && !launching && !disabled;
  const isRunning = instance.status === 'running' || launchStatus === 'running';

  const getLaunchText = () => {
    if (launching) return 'Launching...';
    if (isRunning) return 'Running';
    if (instance.status === 'installing') return 'Installing...';
    if (instance.status === 'updating') return 'Updating...';
    if (instance.status === 'broken') return 'Broken';
    if (instance.status === 'needsUpdate') return 'Update Required';
    return 'Launch';
  };

  const getLaunchIcon = () => {
    if (launching) return '🚀';
    if (isRunning) return '▶️';
    if (instance.status === 'installing' || instance.status === 'updating') return '⏳';
    if (instance.status === 'broken') return '⚠️';
    return '🎮';
  };

  const getButtonVariant = (): 'primary' | 'secondary' | 'outline' | 'ghost' => {
    if (launching) return 'primary';
    if (isRunning) return 'outline';
    if (instance.status === 'broken') return 'outline';
    if (instance.status === 'needsUpdate') return 'outline';
    return 'primary';
  };

  const getStatusText = () => {
    if (launching) return 'Starting game...';
    if (isRunning) return 'Game is running';
    if (instance.status === 'ready') return 'Ready to launch';
    if (instance.status === 'installing') return 'Installing modloader...';
    if (instance.status === 'updating') return 'Updating instance...';
    if (instance.status === 'broken') return 'Instance needs repair';
    if (instance.status === 'needsUpdate') return 'Update available';
    return 'Unknown status';
  };

  return (
    <LaunchButtonContainer>
      <LaunchButtonWrapper
        $launching={launching}
        whileHover={canLaunch ? { scale: 1.02 } : {}}
        whileTap={canLaunch ? { scale: 0.98 } : {}}
      >
        <Button
          variant={getButtonVariant()}
          size={size}
          onClick={handleLaunch}
          disabled={!canLaunch}
          loading={launching}
          style={{
            minWidth: '120px',
            position: 'relative',
          }}
        >
          <LaunchIcon $spinning={launching}>
            {getLaunchIcon()}
          </LaunchIcon>
          {getLaunchText()}

          {launching && (
            <LaunchProgress />
          )}
        </Button>
      </LaunchButtonWrapper>

      {showStatus && (
        <LaunchStatus>
          <StatusIndicator $status={launchStatus} />
          <span>{getStatusText()}</span>
        </LaunchStatus>
      )}

      {launchTime && (
        <LaunchMetrics>
          <Metric>
            <span>⚡</span>
            <span>{launchTime}ms</span>
          </Metric>
          <Metric>
            <span>💾</span>
            <span>{instance.memoryMb}MB</span>
          </Metric>
        </LaunchMetrics>
      )}
    </LaunchButtonContainer>
  );
};

export default LaunchButton;