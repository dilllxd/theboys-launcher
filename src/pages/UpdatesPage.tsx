import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { Layout } from '../components/layout/Layout';
import { UpdateManager } from '../components/updates/UpdateManager';
import { LauncherUpdateManager } from '../components/updates/LauncherUpdateManager';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Select } from '../components/ui/Select';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';
import { Badge } from '../components/ui/Badge';
import { Instance } from '../types/launcher';

interface ModpackUpdate {
  modpack_id: string;
  current_version: string;
  latest_version: string;
  update_available: boolean;
  changelog_url?: string;
  download_url: string;
  size_bytes: number;
}

export const UpdatesPage: React.FC = () => {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [selectedInstanceId, setSelectedInstanceId] = useState<string>('');
  const [allUpdates, setAllUpdates] = useState<ModpackUpdate[]>([]);
  const [_isLoading, _setIsLoading] = useState(false);
  const [checkingAll, setCheckingAll] = useState(false);
  const [activeTab, setActiveTab] = useState<'launcher' | 'instances'>('launcher');

  useEffect(() => {
    loadInstances();
    checkAllUpdates();
  }, []);

  const loadInstances = async () => {
    try {
      const instances = await invoke<Instance[]>('get_instances');
      setInstances(instances);
      if (instances.length > 0 && !selectedInstanceId) {
        setSelectedInstanceId(instances[0].id);
      }
    } catch (error) {
      console.error('Failed to load instances:', error);
    }
  };

  const checkAllUpdates = async () => {
    setCheckingAll(true);
    try {
      const updates = await invoke<ModpackUpdate[]>('check_all_modpack_updates');
      setAllUpdates(updates);
    } catch (error) {
      console.error('Failed to check all updates:', error);
    } finally {
      setCheckingAll(false);
    }
  };

  const updateAllInstances = async () => {
    // This would update all instances that have updates available
    // Implementation would depend on specific requirements
    console.log('Updating all instances with available updates');
  };

  const selectedInstance = instances.find(i => i.id === selectedInstanceId);

  const getInstanceUpdateStatus = (instanceId: string) => {
    const update = allUpdates.find(u => u.modpack_id === instanceId);
    return update;
  };

  return (
    <Layout>
      <div className="updates-page">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-3xl font-bold">Updates</h1>
            <p className="text-gray-600 mt-1">
              Manage launcher updates and modpack updates for your instances
            </p>
          </div>
          <div className="flex gap-2">
            {activeTab === 'instances' && (
              <>
                <Button
                  variant="outline"
                  onClick={checkAllUpdates}
                  disabled={checkingAll}
                >
                  {checkingAll ? <LoadingSpinner size="sm" /> : '🔄'} Check All Updates
                </Button>
                {allUpdates.length > 0 && (
                  <Button onClick={updateAllInstances}>
                    ⬆️ Update All ({allUpdates.length})
                  </Button>
                )}
              </>
            )}
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex gap-1 mb-6 bg-gray-100 p-1 rounded-lg w-fit">
          <button
            className={`px-4 py-2 rounded-md font-medium text-sm transition-colors ${
              activeTab === 'launcher'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
            onClick={() => setActiveTab('launcher')}
          >
            🚀 Launcher Updates
          </button>
          <button
            className={`px-4 py-2 rounded-md font-medium text-sm transition-colors ${
              activeTab === 'instances'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
            onClick={() => setActiveTab('instances')}
          >
            📦 Instance Updates
            {allUpdates.length > 0 && (
              <Badge variant="success" className="ml-2">
                {allUpdates.length}
              </Badge>
            )}
          </button>
        </div>

        {/* Tab Content */}
        {activeTab === 'launcher' && (
          <LauncherUpdateManager />
        )}

        {activeTab === 'instances' && (
          <>
            {/* Global Update Status */}
            <Card className="mb-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold mb-2">Global Update Status</h2>
                  <div className="flex items-center gap-4 text-sm">
                    <span>
                      Total Instances: <Badge variant="outline">{instances.length}</Badge>
                    </span>
                    <span>
                      Updates Available: <Badge variant="success">{allUpdates.length}</Badge>
                    </span>
                    <span>
                      Up to Date: <Badge variant="outline">{instances.length - allUpdates.length}</Badge>
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-4xl mb-2">
                    {allUpdates.length > 0 ? '⬆️' : '✅'}
                  </div>
                  <div className="text-sm text-gray-600">
                    {allUpdates.length > 0 ? 'Updates Available' : 'All Up to Date'}
                  </div>
                </div>
              </div>

              {allUpdates.length > 0 && (
                <div className="mt-4 pt-4 border-t">
                  <h3 className="font-medium mb-3">Instances with Updates:</h3>
                  <div className="space-y-2">
                    {allUpdates.map((update, index) => {
                      const instance = instances.find(i => i.id === update.modpack_id);
                      return (
                        <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{instance?.name || update.modpack_id}</span>
                            <Badge variant="success">
                              {update.current_version} → {update.latest_version}
                            </Badge>
                          </div>
                          <Button size="sm" variant="outline">
                            Update
                          </Button>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </Card>

            {/* Instance Selector */}
            {instances.length > 0 && (
              <Card className="mb-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-xl font-semibold mb-2">Instance Management</h2>
                    <p className="text-sm text-gray-600">
                      Select an instance to manage updates, backups, and manual downloads
                    </p>
                  </div>
                  <div className="w-64">
                    <Select
                      value={selectedInstanceId}
                      onChange={(e) => setSelectedInstanceId(e.target.value)}
                    >
                      {instances.map(instance => (
                        <option key={instance.id} value={instance.id}>
                          {instance.name}
                          {getInstanceUpdateStatus(instance.id) && ' (Update Available)'}
                        </option>
                      ))}
                    </Select>
                  </div>
                </div>
              </Card>
            )}

            {/* Selected Instance Update Manager */}
            {selectedInstance && (
              <div>
                <div className="flex items-center gap-2 mb-4">
                  <h2 className="text-xl font-semibold">
                    {selectedInstance.name}
                  </h2>
                  {getInstanceUpdateStatus(selectedInstance.id) && (
                    <Badge variant="success">Update Available</Badge>
                  )}
                  <Badge variant="outline">
                    {selectedInstance.minecraftVersion} • {selectedInstance.loaderType}
                  </Badge>
                </div>
                <UpdateManager instanceId={selectedInstanceId} />
              </div>
            )}

            {/* Empty State */}
            {instances.length === 0 && !_isLoading && (
              <Card>
                <div className="text-center py-12">
                  <div className="text-6xl mb-4">📦</div>
                  <h3 className="text-xl font-semibold mb-2">No Instances Found</h3>
                  <p className="text-gray-600 mb-6">
                    Create an instance to start managing updates and backups
                  </p>
                  <Button onClick={() => window.location.href = '/instances'}>
                    Create Instance
                  </Button>
                </div>
              </Card>
            )}

            {_isLoading && (
              <Card>
                <div className="flex items-center justify-center py-12">
                  <LoadingSpinner size="lg" />
                  <span className="ml-3">Loading instances...</span>
                </div>
              </Card>
            )}
          </>
        )}
      </div>
    </Layout>
  );
};