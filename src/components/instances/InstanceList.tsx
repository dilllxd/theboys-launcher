import React, { useState } from 'react';
import styled from 'styled-components';
import { Instance } from '../../types/launcher';
import { InstanceCard } from './InstanceCard';
import { Button } from '../ui/Button';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import toast from 'react-hot-toast';

interface InstanceListProps {
  instances: Instance[];
  loading?: boolean;
  onLaunch: (instanceId: string) => Promise<void>;
  onEdit: (instanceId: string) => void;
  onDelete: (instanceId: string) => Promise<void>;
  onRepair: (instanceId: string) => Promise<void>;
  onInstallModloader: (instanceId: string) => Promise<void>;
  onCreateInstance: () => void;
}

const InstanceListContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const ListHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
`;

const ListTitle = styled.h2`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
`;

const ListActions = styled.div`
  display: flex;
  gap: var(--spacing-md);
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
`;

const EmptyStateIcon = styled.div`
  font-size: 4rem;
  margin-bottom: var(--spacing-lg);
  opacity: 0.5;
`;

const EmptyStateTitle = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm) 0;
`;

const EmptyStateDescription = styled.p`
  color: var(--color-text-secondary);
  margin: 0 0 var(--spacing-lg) 0;
`;

const InstanceGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: var(--spacing-lg);
`;

const LoadingOverlay = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--spacing-xl);
`;

export const InstanceList: React.FC<InstanceListProps> = ({
  instances,
  loading = false,
  onLaunch,
  onEdit,
  onDelete,
  onRepair,
  onInstallModloader,
  onCreateInstance,
}) => {
  const [launchingInstances, setLaunchingInstances] = useState<Set<string>>(new Set());
  const [deletingInstances, setDeletingInstances] = useState<Set<string>>(new Set());
  const [repairingInstances, setRepairingInstances] = useState<Set<string>>(new Set());
  const [installingModloaders, setInstallingModloaders] = useState<Set<string>>(new Set());

  const handleLaunch = async (instanceId: string) => {
    if (launchingInstances.has(instanceId)) return;

    setLaunchingInstances(prev => new Set(prev).add(instanceId));

    try {
      await onLaunch(instanceId);
      toast.success('Instance launched successfully!');
    } catch (error) {
      toast.error('Failed to launch instance');
      console.error('Launch error:', error);
    } finally {
      setLaunchingInstances(prev => {
        const next = new Set(prev);
        next.delete(instanceId);
        return next;
      });
    }
  };

  const handleDelete = async (instanceId: string) => {
    if (deletingInstances.has(instanceId)) return;

    const instance = instances.find(i => i.id === instanceId);
    if (!instance) return;

    // Confirm deletion
    const confirmed = window.confirm(
      `Are you sure you want to delete "${instance.name}"? This action cannot be undone.`
    );

    if (!confirmed) return;

    setDeletingInstances(prev => new Set(prev).add(instanceId));

    try {
      await onDelete(instanceId);
      toast.success('Instance deleted successfully!');
    } catch (error) {
      toast.error('Failed to delete instance');
      console.error('Delete error:', error);
    } finally {
      setDeletingInstances(prev => {
        const next = new Set(prev);
        next.delete(instanceId);
        return next;
      });
    }
  };

  const handleRepair = async (instanceId: string) => {
    if (repairingInstances.has(instanceId)) return;

    setRepairingInstances(prev => new Set(prev).add(instanceId));

    try {
      await onRepair(instanceId);
      toast.success('Instance repair completed!');
    } catch (error) {
      toast.error('Failed to repair instance');
      console.error('Repair error:', error);
    } finally {
      setRepairingInstances(prev => {
        const next = new Set(prev);
        next.delete(instanceId);
        return next;
      });
    }
  };

  const handleInstallModloader = async (instanceId: string) => {
    if (installingModloaders.has(instanceId)) return;

    setInstallingModloaders(prev => new Set(prev).add(instanceId));

    try {
      await onInstallModloader(instanceId);
      toast.success('Modloader installed successfully!');
    } catch (error) {
      toast.error('Failed to install modloader');
      console.error('Modloader install error:', error);
    } finally {
      setInstallingModloaders(prev => {
        const next = new Set(prev);
        next.delete(instanceId);
        return next;
      });
    }
  };

  
  if (loading) {
    return (
      <LoadingOverlay>
        <LoadingSpinner size="lg" />
      </LoadingOverlay>
    );
  }

  if (instances.length === 0) {
    return (
      <EmptyState>
        <EmptyStateIcon>🎮</EmptyStateIcon>
        <EmptyStateTitle>No instances yet</EmptyStateTitle>
        <EmptyStateDescription>
          Create your first instance to start playing Minecraft with modpacks.
        </EmptyStateDescription>
        <Button variant="primary" onClick={onCreateInstance}>
          Create Instance
        </Button>
      </EmptyState>
    );
  }

  return (
    <InstanceListContainer>
      <ListHeader>
        <ListTitle>
          {instances.length} {instances.length === 1 ? 'Instance' : 'Instances'}
        </ListTitle>
        <ListActions>
          <Button variant="primary" onClick={onCreateInstance}>
            Create Instance
          </Button>
        </ListActions>
      </ListHeader>

      <InstanceGrid>
        {instances.map((instance) => (
          <InstanceCard
            key={instance.id}
            instance={instance}
            onLaunch={handleLaunch}
            onEdit={onEdit}
            onDelete={handleDelete}
            onRepair={handleRepair}
            onInstallModloader={handleInstallModloader}
          />
        ))}
      </InstanceGrid>
    </InstanceListContainer>
  );
};