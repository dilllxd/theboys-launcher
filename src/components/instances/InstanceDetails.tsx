import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Instance, InstanceValidation, InstanceStatistics } from '../../types/launcher';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { api } from '../../utils/api';
import toast from 'react-hot-toast';

interface InstanceDetailsProps {
  instance: Instance | null;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: () => void;
}

const DetailsContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  max-height: 80vh;
  overflow-y: auto;
`;

const DetailsHeader = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
`;

const InstanceIcon = styled.div`
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
  border-radius: var(--border-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
`;

const InstanceInfo = styled.div`
  flex: 1;
`;

const InstanceName = styled.h2`
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
`;

const InstanceDetailsContainer = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
`;

const TabContainer = styled.div`
  display: flex;
  border-bottom: 1px solid var(--color-border-light);
  margin-bottom: var(--spacing-lg);
`;

const Tab = styled.button<{ active: boolean }>`
  padding: var(--spacing-md) var(--spacing-lg);
  background: none;
  border: none;
  border-bottom: 2px solid ${props => props.active ? 'var(--color-primary-500)' : 'transparent'};
  color: ${props => props.active ? 'var(--color-primary-600)' : 'var(--color-text-secondary)'};
  font-weight: ${props => props.active ? 'var(--font-weight-semibold)' : 'var(--font-weight-normal)'};
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: var(--color-primary-600);
  }
`;

const TabContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
`;

const Label = styled.label`
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
`;

const Input = styled.input`
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-base);
  transition: border-color 0.2s ease;

  &:focus {
    outline: none;
    border-color: var(--color-primary-500);
  }
`;

const Textarea = styled.textarea`
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-base);
  font-family: monospace;
  resize: vertical;
  min-height: 100px;
  transition: border-color 0.2s ease;

  &:focus {
    outline: none;
    border-color: var(--color-primary-500);
  }
`;

const Slider = styled.input`
  width: 100%;
  margin: var(--spacing-sm) 0;
`;


const Actions = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  justify-content: flex-end;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
`;

const StatCard = styled(Card)`
  padding: var(--spacing-lg);
  text-align: center;
`;

const StatValue = styled.div`
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary-600);
  margin-bottom: var(--spacing-xs);
`;

const StatLabel = styled.div`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
`;

const ValidationResults = styled(Card)`
  padding: var(--spacing-lg);
`;

const ValidationHeader = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-md) 0;
`;

const ValidationStatus = styled.div<{ valid: boolean }>`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-md);
  margin-bottom: var(--spacing-md);
  background: ${props => props.valid ? 'var(--color-success-50)' : 'var(--color-error-50)'};
  color: ${props => props.valid ? 'var(--color-success-700)' : 'var(--color-error-700)'};
  font-weight: var(--font-weight-semibold);
`;

const ValidationList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
`;

const ValidationItem = styled.div<{ type: 'issue' | 'recommendation' }>`
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  border-radius: var(--border-radius-sm);
  background: ${props => props.type === 'issue' ? 'var(--color-error-50)' : 'var(--color-warning-50)'};
  color: ${props => props.type === 'issue' ? 'var(--color-error-700)' : 'var(--color-warning-700)'};
`;

const LogContainer = styled.div`
  background: var(--color-gray-900);
  color: var(--color-gray-100);
  padding: var(--spacing-lg);
  border-radius: var(--border-radius-md);
  font-family: monospace;
  font-size: var(--font-size-sm);
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
`;

const LogActions = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
`;

type TabType = 'settings' | 'statistics' | 'validation' | 'logs';

export const InstanceDetails: React.FC<InstanceDetailsProps> = ({
  instance,
  isOpen,
  onClose,
  onUpdate,
}) => {
  const [activeTab, setActiveTab] = useState<TabType>('settings');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // Form state
  const [instanceName, setInstanceName] = useState('');
  const [memoryMb, setMemoryMb] = useState(4096);
  const [javaPath, setJavaPath] = useState('');
  const [jvmArgs, setJvmArgs] = useState('');

  // Data state
  const [validation, setValidation] = useState<InstanceValidation | null>(null);
  const [statistics, setStatistics] = useState<InstanceStatistics | null>(null);
  const [logs, setLogs] = useState<string[]>([]);

  // Load data when instance changes
  useEffect(() => {
    if (instance && isOpen) {
      loadInstanceData();
    }
  }, [instance, isOpen]);

  // Load tab-specific data
  useEffect(() => {
    if (instance && isOpen) {
      switch (activeTab) {
        case 'validation':
          loadValidation();
          break;
        case 'statistics':
          loadStatistics();
          break;
        case 'logs':
          loadLogs();
          break;
      }
    }
  }, [activeTab, instance, isOpen]);

  const loadInstanceData = () => {
    if (!instance) return;

    setInstanceName(instance.name);
    setMemoryMb(instance.memoryMb);
    setJavaPath(instance.javaPath);
    setJvmArgs(instance.jvmArgs || '');
  };

  const loadValidation = async () => {
    if (!instance) return;

    setLoading(true);
    try {
      const result = await api.validateInstance(instance.id);
      setValidation(result);
    } catch (error) {
      toast.error('Failed to validate instance');
      console.error('Validation error:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadStatistics = async () => {
    if (!instance) return;

    setLoading(true);
    try {
      const result = await api.getInstanceStatistics(instance.id);
      setStatistics(result);
    } catch (error) {
      toast.error('Failed to load statistics');
      console.error('Statistics error:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadLogs = async () => {
    if (!instance) return;

    setLoading(true);
    try {
      const result = await api.getInstanceLogs(instance.id);
      setLogs(result);
    } catch (error) {
      toast.error('Failed to load logs');
      console.error('Logs error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!instance) return;

    setSaving(true);
    try {
      const updatedInstance = {
        ...instance,
        name: instanceName,
        memoryMb,
        javaPath,
        jvmArgs: jvmArgs || undefined,
      };

      await api.updateInstance(updatedInstance);
      toast.success('Instance settings saved!');
      onUpdate();
    } catch (error) {
      toast.error('Failed to save settings');
      console.error('Save error:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleRepair = async () => {
    if (!instance) return;

    try {
      await api.repairInstance(instance.id);
      toast.success('Instance repair completed!');
      loadValidation(); // Reload validation after repair
      onUpdate();
    } catch (error) {
      toast.error('Failed to repair instance');
      console.error('Repair error:', error);
    }
  };

  const handleInstallModloader = async () => {
    if (!instance) return;

    try {
      await api.installModloader(instance.id);
      toast.success('Modloader installed successfully!');
      onUpdate();
    } catch (error) {
      toast.error('Failed to install modloader');
      console.error('Modloader install error:', error);
    }
  };

  const handleClearLogs = async () => {
    if (!instance) return;

    try {
      await api.clearInstanceLogs(instance.id);
      toast.success('Logs cleared!');
      setLogs([]);
    } catch (error) {
      toast.error('Failed to clear logs');
      console.error('Clear logs error:', error);
    }
  };

  const formatBytes = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatPlayTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours === 0) return `${minutes}m`;
    return `${hours}h ${minutes}m`;
  };

  if (!instance) return null;

  const tabs = [
    { id: 'settings', label: 'Settings' },
    { id: 'statistics', label: 'Statistics' },
    { id: 'validation', label: 'Validation' },
    { id: 'logs', label: 'Logs' },
  ] as const;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Instance Details"
      size="xl"
    >
      <DetailsContainer>
        <DetailsHeader>
          <InstanceIcon>
            {instance.name.charAt(0).toUpperCase()}
          </InstanceIcon>
          <InstanceInfo>
            <InstanceName>{instance.name}</InstanceName>
            <InstanceDetailsContainer>
              <span>📦 {instance.minecraftVersion}</span>
              <span>⚙️ {instance.loaderType}</span>
              <span>💾 {instance.memoryMb}MB</span>
              <span>📅 Created {new Date(instance.createdAt).toLocaleDateString()}</span>
            </InstanceDetailsContainer>
          </InstanceInfo>
        </DetailsHeader>

        <TabContainer>
          {tabs.map((tab) => (
            <Tab
              key={tab.id}
              active={activeTab === tab.id}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </Tab>
          ))}
        </TabContainer>

        <TabContent>
          {activeTab === 'settings' && (
            <Form>
              <FormGroup>
                <Label htmlFor="instanceName">Instance Name</Label>
                <Input
                  id="instanceName"
                  type="text"
                  value={instanceName}
                  onChange={(e) => setInstanceName(e.target.value)}
                />
              </FormGroup>

              <FormGroup>
                <Label htmlFor="memory">Memory Allocation: {memoryMb}MB</Label>
                <Slider
                  id="memory"
                  type="range"
                  min="1024"
                  max="16384"
                  step="512"
                  value={memoryMb}
                  onChange={(e) => setMemoryMb(Number(e.target.value))}
                />
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  <span>1GB</span>
                  <span>16GB</span>
                </div>
              </FormGroup>

              <FormGroup>
                <Label htmlFor="javaPath">Java Path</Label>
                <Input
                  id="javaPath"
                  type="text"
                  value={javaPath}
                  onChange={(e) => setJavaPath(e.target.value)}
                />
              </FormGroup>

              <FormGroup>
                <Label htmlFor="jvmArgs">JVM Arguments (Optional)</Label>
                <Textarea
                  id="jvmArgs"
                  value={jvmArgs}
                  onChange={(e) => setJvmArgs(e.target.value)}
                  placeholder="Additional JVM arguments..."
                />
              </FormGroup>
            </Form>
          )}

          {activeTab === 'statistics' && (
            <>
              {loading ? (
                <LoadingSpinner size="lg" />
              ) : statistics ? (
                <StatsGrid>
                  <StatCard>
                    <StatValue>{formatBytes(statistics.totalSizeBytes)}</StatValue>
                    <StatLabel>Total Size</StatLabel>
                  </StatCard>
                  <StatCard>
                    <StatValue>{statistics.modsCount}</StatValue>
                    <StatLabel>Mods</StatLabel>
                  </StatCard>
                  <StatCard>
                    <StatValue>{statistics.resourcePacksCount}</StatValue>
                    <StatLabel>Resource Packs</StatLabel>
                  </StatCard>
                  <StatCard>
                    <StatValue>{statistics.screenshotsCount}</StatValue>
                    <StatLabel>Screenshots</StatLabel>
                  </StatCard>
                  <StatCard>
                    <StatValue>{formatPlayTime(statistics.totalPlaytimeSeconds)}</StatValue>
                    <StatLabel>Total Playtime</StatLabel>
                  </StatCard>
                  <StatCard>
                    <StatValue>
                      {statistics.lastPlayed
                        ? new Date(statistics.lastPlayed).toLocaleDateString()
                        : 'Never'
                      }
                    </StatValue>
                    <StatLabel>Last Played</StatLabel>
                  </StatCard>
                </StatsGrid>
              ) : (
                <p>No statistics available</p>
              )}
            </>
          )}

          {activeTab === 'validation' && (
            <>
              {loading ? (
                <LoadingSpinner size="lg" />
              ) : validation ? (
                <ValidationResults>
                  <ValidationHeader>Instance Validation</ValidationHeader>
                  <ValidationStatus valid={validation.isValid}>
                    {validation.isValid ? '✅ Instance is valid' : '❌ Instance has issues'}
                  </ValidationStatus>

                  {validation.issues.length > 0 && (
                    <div>
                      <h4 style={{ margin: '0 0 var(--spacing-sm) 0' }}>Issues:</h4>
                      <ValidationList>
                        {validation.issues.map((issue, index) => (
                          <ValidationItem key={index} type="issue">
                            <span>❌</span>
                            <span>{issue}</span>
                          </ValidationItem>
                        ))}
                      </ValidationList>
                    </div>
                  )}

                  {validation.recommendations.length > 0 && (
                    <div>
                      <h4 style={{ margin: 'var(--spacing-md) 0 var(--spacing-sm) 0' }}>Recommendations:</h4>
                      <ValidationList>
                        {validation.recommendations.map((recommendation, index) => (
                          <ValidationItem key={index} type="recommendation">
                            <span>💡</span>
                            <span>{recommendation}</span>
                          </ValidationItem>
                        ))}
                      </ValidationList>
                    </div>
                  )}

                  {!validation.isValid && (
                    <div style={{ display: 'flex', gap: 'var(--spacing-sm)', marginTop: 'var(--spacing-lg)' }}>
                      <Button variant="primary" onClick={handleRepair}>
                        Repair Instance
                      </Button>
                      <Button variant="outline" onClick={handleInstallModloader}>
                        Reinstall Modloader
                      </Button>
                    </div>
                  )}
                </ValidationResults>
              ) : (
                <p>No validation data available</p>
              )}
            </>
          )}

          {activeTab === 'logs' && (
            <>
              <LogActions>
                <Button variant="outline" onClick={loadLogs}>
                  Refresh
                </Button>
                <Button variant="outline" onClick={handleClearLogs}>
                  Clear Logs
                </Button>
              </LogActions>
              {loading ? (
                <LoadingSpinner size="lg" />
              ) : logs.length > 0 ? (
                <LogContainer>
                  {logs.join('\n')}
                </LogContainer>
              ) : (
                <p>No logs available</p>
              )}
            </>
          )}
        </TabContent>

        <Actions>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
          {activeTab === 'settings' && (
            <Button variant="primary" onClick={handleSave} loading={saving}>
              Save Changes
            </Button>
          )}
        </Actions>
      </DetailsContainer>
    </Modal>
  );
};