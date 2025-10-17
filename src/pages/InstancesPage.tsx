import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Instance } from '../types/launcher';
import { InstanceList } from '../components/instances/InstanceList';
import { CreateInstanceWizard } from '../components/instances/CreateInstanceWizard';
import { InstanceDetails } from '../components/instances/InstanceDetails';
import { api } from '../utils/api';
import toast from 'react-hot-toast';

const InstancesContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
`;

const PageTitle = styled.h1`
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0;
`;

const PageActions = styled.div`
  display: flex;
  gap: var(--spacing-md);
`;

export const InstancesPage: React.FC = () => {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateWizard, setShowCreateWizard] = useState(false);
  const [selectedInstance, setSelectedInstance] = useState<Instance | null>(null);
  const [showDetails, setShowDetails] = useState(false);

  // Load instances on mount
  useEffect(() => {
    loadInstances();
  }, []);

  const loadInstances = async () => {
    setLoading(true);
    try {
      const instances = await api.getInstances();
      setInstances(instances);
    } catch (error) {
      toast.error('Failed to load instances');
      console.error('Load instances error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLaunch = async (instanceId: string) => {
    try {
      await api.launchInstance(instanceId);
      toast.success('Instance launched successfully!');
      // Optionally refresh instances to update status
      setTimeout(loadInstances, 1000);
    } catch (error) {
      // Error is handled by the InstanceList component
      throw error;
    }
  };

  const handleEdit = (instanceId: string) => {
    const instance = instances.find(i => i.id === instanceId);
    if (instance) {
      setSelectedInstance(instance);
      setShowDetails(true);
    }
  };

  const handleDelete = async (instanceId: string) => {
    try {
      await api.deleteInstance(instanceId);
      // Remove from local state immediately for better UX
      setInstances(prev => prev.filter(i => i.id !== instanceId));
    } catch (error) {
      // Error is handled by the InstanceList component
      throw error;
    }
  };

  const handleRepair = async (instanceId: string) => {
    try {
      await api.repairInstance(instanceId);
      // Refresh instances after repair
      setTimeout(loadInstances, 2000);
    } catch (error) {
      // Error is handled by the InstanceList component
      throw error;
    }
  };

  const handleInstallModloader = async (instanceId: string) => {
    try {
      await api.installModloader(instanceId);
      // Refresh instances after modloader installation
      setTimeout(loadInstances, 2000);
    } catch (error) {
      // Error is handled by the InstanceList component
      throw error;
    }
  };

  const handleCreateInstance = () => {
    setShowCreateWizard(true);
  };

  const handleCreateSuccess = () => {
    setShowCreateWizard(false);
    loadInstances(); // Refresh the list
  };

  const handleDetailsUpdate = () => {
    loadInstances(); // Refresh to show updated data
  };

  return (
    <InstancesContainer>
      <PageHeader>
        <PageTitle>Instances</PageTitle>
        <PageActions>
          {/* Additional page-level actions can be added here */}
        </PageActions>
      </PageHeader>

      <InstanceList
        instances={instances}
        loading={loading}
        onLaunch={handleLaunch}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onRepair={handleRepair}
        onInstallModloader={handleInstallModloader}
        onCreateInstance={handleCreateInstance}
      />

      <CreateInstanceWizard
        isOpen={showCreateWizard}
        onClose={() => setShowCreateWizard(false)}
        onSuccess={handleCreateSuccess}
      />

      <InstanceDetails
        instance={selectedInstance}
        isOpen={showDetails}
        onClose={() => {
          setShowDetails(false);
          setSelectedInstance(null);
        }}
        onUpdate={handleDetailsUpdate}
      />
    </InstancesContainer>
  );
};

export default InstancesPage;