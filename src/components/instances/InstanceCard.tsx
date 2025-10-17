import React from 'react';
import styled from 'styled-components';
import { Instance, InstanceStatus } from '../../types/launcher';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import LaunchButton from '../launch/LaunchButton';

interface InstanceCardProps {
  instance: Instance;
  onLaunch: (instanceId: string) => void;
  onEdit: (instanceId: string) => void;
  onDelete: (instanceId: string) => void;
  onRepair: (instanceId: string) => void;
  onInstallModloader: (instanceId: string) => void;
}

const InstanceCardContainer = styled(Card)`
  padding: var(--spacing-lg);
  transition: all 0.2s ease;
  cursor: pointer;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }
`;

const InstanceHeader = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
`;

const InstanceIcon = styled.div`
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
  border-radius: var(--border-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
`;

const InstanceInfo = styled.div`
  flex: 1;
`;

const InstanceName = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
`;

const InstanceDetails = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
`;

const InstanceDetail = styled.span`
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
`;

const StatusBadge = styled.span<{ status: InstanceStatus }>`
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;

  background: ${props => {
    switch (props.status) {
      case 'ready':
        return 'var(--color-success-100)';
      case 'running':
        return 'var(--color-primary-100)';
      case 'installing':
      case 'updating':
        return 'var(--color-warning-100)';
      case 'broken':
        return 'var(--color-error-100)';
      case 'needsUpdate':
        return 'var(--color-warning-100)';
      default:
        return 'var(--color-gray-100)';
    }
  }};

  color: ${props => {
    switch (props.status) {
      case 'ready':
        return 'var(--color-success-700)';
      case 'running':
        return 'var(--color-primary-700)';
      case 'installing':
      case 'updating':
        return 'var(--color-warning-700)';
      case 'broken':
        return 'var(--color-error-700)';
      case 'needsUpdate':
        return 'var(--color-warning-700)';
      default:
        return 'var(--color-gray-700)';
    }
  }};
`;

const InstanceActions = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
`;

const PlayTime = styled.div`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-sm);
`;

const formatPlayTime = (seconds: number): string => {
  if (seconds === 0) return 'Never played';

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours === 0) {
    return `${minutes} minutes`;
  } else if (minutes === 0) {
    return `${hours} hour${hours > 1 ? 's' : ''}`;
  } else {
    return `${hours}h ${minutes}m`;
  }
};

const formatLastPlayed = (lastPlayed?: string): string => {
  if (!lastPlayed) return 'Never';

  const date = new Date(lastPlayed);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
  return `${Math.floor(diffDays / 365)} years ago`;
};

export const InstanceCard: React.FC<InstanceCardProps> = ({
  instance,
  onLaunch: _onLaunch,
  onEdit,
  onDelete,
  onRepair,
  onInstallModloader,
}) => {
  const handleEdit = () => onEdit(instance.id);
  const handleDelete = () => onDelete(instance.id);
  const handleRepair = () => onRepair(instance.id);
  const handleInstallModloader = () => onInstallModloader(instance.id);

  const needsRepair = instance.status === 'broken';
  const needsModloaderInstall = instance.status === 'installing';

  return (
    <InstanceCardContainer>
      <InstanceHeader>
        <InstanceIcon>
          {instance.name.charAt(0).toUpperCase()}
        </InstanceIcon>
        <InstanceInfo>
          <InstanceName>{instance.name}</InstanceName>
          <InstanceDetails>
            <InstanceDetail>
              <span>📦</span>
              <span>Minecraft {instance.minecraftVersion}</span>
            </InstanceDetail>
            <InstanceDetail>
              <span>⚙️</span>
              <span>{instance.loaderType}</span>
              {instance.loaderVersion && (
                <span>({instance.loaderVersion})</span>
              )}
            </InstanceDetail>
            <InstanceDetail>
              <span>💾</span>
              <span>{instance.memoryMb}MB</span>
            </InstanceDetail>
          </InstanceDetails>
        </InstanceInfo>
        <StatusBadge status={instance.status}>
          {instance.status.replace(/([A-Z])/g, ' $1').trim()}
        </StatusBadge>
      </InstanceHeader>

      <InstanceActions>
        <LaunchButton
          instance={instance}
          size="sm"
          onLaunchComplete={(launchId) => {
            // Handle successful launch
            console.log(`Instance ${instance.name} launched with ID: ${launchId}`);
          }}
          onError={(error) => {
            // Handle launch error
            console.error(`Failed to launch ${instance.name}:`, error);
          }}
        />

        <Button
          variant="outline"
          size="sm"
          onClick={handleEdit}
        >
          Edit
        </Button>

        {needsRepair && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleRepair}
          >
            Repair
          </Button>
        )}

        {needsModloaderInstall && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleInstallModloader}
          >
            Install Modloader
          </Button>
        )}

        <Button
          variant="ghost"
          size="sm"
          onClick={handleDelete}
          style={{ color: 'var(--color-error-600)' }}
        >
          Delete
        </Button>
      </InstanceActions>

      <PlayTime>
        {formatPlayTime(instance.totalPlaytime)} total • Last played {formatLastPlayed(instance.lastPlayed)}
      </PlayTime>
    </InstanceCardContainer>
  );
};