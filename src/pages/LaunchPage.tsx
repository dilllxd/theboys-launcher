import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { Layout } from '../components/layout/Layout';
import { LaunchManager } from '../components/launch/LaunchManager';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Instance } from '../types/launcher';
import { invoke } from '@tauri-apps/api/core';
import { toast } from 'react-hot-toast';

const PageContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  padding: var(--spacing-lg);
`;

const PageHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
`;

const PageTitle = styled.h1`
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
`;

const PageActions = styled.div`
  display: flex;
  gap: var(--spacing-md);
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
`;

const StatCard = styled(Card)`
  padding: var(--spacing-lg);
  text-align: center;
`;

const StatValue = styled.div`
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary-600);
  margin-bottom: var(--spacing-xs);
`;

const StatLabel = styled.div`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
`;

const SectionContainer = styled(motion.div)`
  margin-bottom: var(--spacing-xl);
`;

const SectionTitle = styled.h2`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-md) 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
`;

const LaunchPage: React.FC = () => {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [stats, setStats] = useState({
    totalInstances: 0,
    runningInstances: 0,
    readyInstances: 0,
    needsAttention: 0,
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadInstances();
    loadStats();

    // Set up periodic updates
    const interval = setInterval(() => {
      loadInstances();
      loadStats();
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  const loadInstances = async () => {
    try {
      const instanceList = await invoke<Instance[]>('get_instances');
      setInstances(instanceList);
    } catch (error) {
      console.error('Failed to load instances:', error);
      toast.error('Failed to load instances');
    }
  };

  const loadStats = async () => {
    try {
      const instanceList = await invoke<Instance[]>('get_instances');
      const total = instanceList.length;
      const running = instanceList.filter(i => i.status === 'running').length;
      const ready = instanceList.filter(i => i.status === 'ready').length;
      const needsAttention = instanceList.filter(i =>
        i.status === 'broken' || i.status === 'needsUpdate'
      ).length;

      setStats({
        totalInstances: total,
        runningInstances: running,
        readyInstances: ready,
        needsAttention: needsAttention,
      });
    } catch (error) {
      console.error('Failed to load stats:', error);
    }
  };

  const handleCleanupAll = async () => {
    setLoading(true);
    try {
      const cleanedCount = await invoke<number>('cleanup_finished_processes');
      if (cleanedCount > 0) {
        toast.success(`Cleaned up ${cleanedCount} finished processes`);
      } else {
        toast('No finished processes to clean up');
      }
    } catch (error) {
      console.error('Failed to cleanup processes:', error);
      toast.error('Failed to cleanup processes');
    } finally {
      setLoading(false);
    }
  };

  const handleInitializeLaunchManager = async () => {
    setLoading(true);
    try {
      await invoke('initialize_launch_manager');
      toast.success('Launch manager initialized successfully');
      await loadInstances();
    } catch (error) {
      console.error('Failed to initialize launch manager:', error);
      toast.error('Failed to initialize launch manager');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout>
      <PageContainer>
        <PageHeader>
          <PageTitle>
            🎮 Game Launch Management
          </PageTitle>
          <PageActions>
            <Button
              variant="outline"
              onClick={handleCleanupAll}
              loading={loading}
            >
              🧹 Cleanup Finished
            </Button>
            <Button
              variant="outline"
              onClick={handleInitializeLaunchManager}
              loading={loading}
            >
              🔄 Refresh
            </Button>
          </PageActions>
        </PageHeader>

        <StatsGrid>
          <StatCard>
            <StatValue>{stats.totalInstances}</StatValue>
            <StatLabel>Total Instances</StatLabel>
          </StatCard>
          <StatCard>
            <StatValue>{stats.runningInstances}</StatValue>
            <StatLabel>Running Games</StatLabel>
          </StatCard>
          <StatCard>
            <StatValue>{stats.readyInstances}</StatValue>
            <StatLabel>Ready to Launch</StatLabel>
          </StatCard>
          <StatCard>
            <StatValue>{stats.needsAttention}</StatValue>
            <StatLabel>Needs Attention</StatLabel>
          </StatCard>
        </StatsGrid>

        <SectionContainer
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <LaunchManager
            instances={instances}
            onInstanceUpdate={loadInstances}
          />
        </SectionContainer>

        {/* Quick Actions Section */}
        <SectionContainer
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <Card>
            <div style={{ padding: 'var(--spacing-lg)' }}>
              <SectionTitle style={{ marginBottom: 'var(--spacing-md)' }}>
                ⚡ Quick Actions
              </SectionTitle>
              <div style={{ display: 'flex', gap: 'var(--spacing-md)', flexWrap: 'wrap' }}>
                <Button
                  variant="outline"
                  onClick={handleCleanupAll}
                  disabled={loading}
                >
                  🧹 Cleanup All Finished Processes
                </Button>
                <Button
                  variant="outline"
                  onClick={handleInitializeLaunchManager}
                  disabled={loading}
                >
                  🔄 Reinitialize Launch Manager
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => {
                    // This would navigate to instances page
                    window.location.hash = '/instances';
                  }}
                >
                  📦 Manage Instances
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => {
                    // This would navigate to settings page
                    window.location.hash = '/settings';
                  }}
                >
                  ⚙️ Launcher Settings
                </Button>
              </div>
            </div>
          </Card>
        </SectionContainer>

        {/* Help Section */}
        <SectionContainer
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
        >
          <Card>
            <div style={{ padding: 'var(--spacing-lg)' }}>
              <SectionTitle style={{ marginBottom: 'var(--spacing-md)' }}>
                💡 Launch Management Tips
              </SectionTitle>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 'var(--spacing-md)' }}>
                <div>
                  <h4 style={{ margin: '0 0 var(--spacing-xs) 0', color: 'var(--color-text-primary)' }}>
                    🎮 Launching Games
                  </h4>
                  <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    Use the Launch button on instance cards or visit the Instances page to start games.
                  </p>
                </div>
                <div>
                  <h4 style={{ margin: '0 0 var(--spacing-xs) 0', color: 'var(--color-text-primary)' }}>
                    📊 Monitoring Status
                  </h4>
                  <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    Active games appear here with real-time status updates and performance metrics.
                  </p>
                </div>
                <div>
                  <h4 style={{ margin: '0 0 var(--spacing-xs) 0', color: 'var(--color-text-primary)' }}>
                    🔧 Troubleshooting
                  </h4>
                  <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    If a game fails to launch, check the error logs and ensure Prism Launcher is properly installed.
                  </p>
                </div>
              </div>
            </div>
          </Card>
        </SectionContainer>
      </PageContainer>
    </Layout>
  );
};

export default LaunchPage;