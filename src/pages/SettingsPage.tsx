import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import toast from 'react-hot-toast';

// UI Components
import { Card, Button, Input, Checkbox, Select, LoadingSpinner, Modal, Badge } from '../components/ui';
import { api } from '../utils/api';

// Types
import { LauncherSettings, SystemInfo, JavaVersion } from '../types/launcher';

const SettingsContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
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

const SettingsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--spacing-lg);
`;

const SettingsSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;


const SettingsRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;

  &:not(:last-child) {
    border-bottom: 1px solid var(--color-border-primary);
  }
`;

const SettingsLabel = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  flex: 1;
`;

const LabelText = styled.label`
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
`;

const LabelDescription = styled.p`
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
  margin: 0;
  line-height: 1.4;
`;

const SettingsControl = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  min-width: 200px;
`;

const MemorySlider = styled.input.attrs({ type: 'range' })`
  width: 200px;
  height: 6px;
  background: var(--color-bg-tertiary);
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;

  &::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 20px;
    height: 20px;
    background: var(--color-primary);
    cursor: pointer;
    border-radius: 50%;
    border: 2px solid var(--color-bg-secondary);
    box-shadow: var(--shadow-sm);
  }

  &::-moz-range-thumb {
    width: 20px;
    height: 20px;
    background: var(--color-primary);
    cursor: pointer;
    border-radius: 50%;
    border: 2px solid var(--color-bg-secondary);
    box-shadow: var(--shadow-sm);
  }
`;

const MemoryDisplay = styled.div`
  min-width: 80px;
  text-align: center;
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
  font-size: var(--font-size-base);
`;

const PathInputContainer = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
  flex: 1;
`;

const PathInput = styled(Input)`
  flex: 1;
`;

const BrowseButton = styled(Button)`
  white-space: nowrap;
`;

const SystemInfoContainer = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
`;

const SystemInfoItem = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
`;

const SystemInfoLabel = styled.span`
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
  font-weight: var(--font-weight-medium);
`;

const SystemInfoValue = styled.span`
  font-size: var(--font-size-base);
  color: var(--color-text-primary);
  font-weight: var(--font-weight-semibold);
`;

const JavaVersionList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  max-height: 200px;
  overflow-y: auto;
`;

const JavaVersionItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm);
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-primary);
`;

const JavaVersionInfo = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
`;

const JavaVersionPath = styled.span`
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  font-family: monospace;
`;

const ActionsContainer = styled.div`
  display: flex;
  gap: var(--spacing-md);
  justify-content: flex-end;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border-primary);
`;

const LoadingContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
`;

const StatusIndicator = styled.div<{ status: 'success' | 'warning' | 'error' }>`
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);

  ${({ status }) => {
    switch (status) {
      case 'success':
        return `
          background: rgba(34, 197, 94, 0.1);
          color: rgb(34, 197, 94);
          border: 1px solid rgba(34, 197, 94, 0.2);
        `;
      case 'warning':
        return `
          background: rgba(251, 191, 36, 0.1);
          color: rgb(251, 191, 36);
          border: 1px solid rgba(251, 191, 36, 0.2);
        `;
      case 'error':
        return `
          background: rgba(239, 68, 68, 0.1);
          color: rgb(239, 68, 68);
          border: 1px solid rgba(239, 68, 68, 0.2);
        `;
    }
  }}
`;

export const SettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<LauncherSettings | null>(null);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [javaVersions, setJavaVersions] = useState<JavaVersion[]>([]);
  const [showResetModal, setShowResetModal] = useState(false);

  // Load settings and system info on mount
  useEffect(() => {
    loadSettings();
    loadSystemInfo();
  }, []);

  const loadSettings = async () => {
    try {
      const loadedSettings = await api.getSettings();
      setSettings(loadedSettings);
    } catch (error) {
      toast.error('Failed to load settings');
      console.error('Failed to load settings:', error);
    }
  };

  const loadSystemInfo = async () => {
    try {
      const info = await api.detectSystemInfo();
      setSystemInfo(info);
      setJavaVersions(info.javaVersions || []);
    } catch (error) {
      toast.error('Failed to load system information');
      console.error('Failed to load system info:', error);
      setJavaVersions([]);
    } finally {
      setIsLoading(false);
    }
  };

  const updateSettings = (updates: Partial<LauncherSettings>) => {
    if (!settings) return;

    const newSettings = { ...settings, ...updates };
    setSettings(newSettings);
    setHasChanges(true);
  };

  const saveSettings = async () => {
    if (!settings || !hasChanges) return;

    setIsSaving(true);
    try {
      await api.saveSettings(settings);
      setHasChanges(false);
      toast.success('Settings saved successfully');
    } catch (error) {
      toast.error('Failed to save settings');
      console.error('Failed to save settings:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const resetSettings = async () => {
    try {
      const defaultSettings = await api.resetSettings();
      setSettings(defaultSettings);
      setHasChanges(false);
      setShowResetModal(false);
      toast.success('Settings reset to defaults');
    } catch (error) {
      toast.error('Failed to reset settings');
      console.error('Failed to reset settings:', error);
    }
  };

  const browseForPath = async (type: 'java' | 'prism' | 'instances') => {
    try {
      let path: string | null = null;

      switch (type) {
        case 'java':
          path = await api.browseForJava();
          if (path) updateSettings({ javaPath: path });
          break;
        case 'prism':
          path = await api.browseForPrism();
          if (path) updateSettings({ prismPath: path });
          break;
        case 'instances':
          path = await api.browseForInstancesDir();
          if (path) updateSettings({ instancesDir: path });
          break;
      }

      if (path) {
        toast.success(`Selected ${type} path`);
      }
    } catch (error) {
      toast.error(`Failed to browse for ${type} path`);
      console.error(`Failed to browse for ${type}:`, error);
    }
  };

  
  if (isLoading) {
    return (
      <LoadingContainer>
        <LoadingSpinner size="lg" />
      </LoadingContainer>
    );
  }

  if (!settings || !systemInfo) {
    return (
      <SettingsContainer>
        <PageTitle>Settings</PageTitle>
        <p>Failed to load settings.</p>
      </SettingsContainer>
    );
  }

  const recommendedMemory = Math.floor(systemInfo.totalMemoryMb / 2 / 1024); // 50% of total RAM in GB

  return (
    <SettingsContainer>
      <PageHeader>
        <PageTitle>Settings</PageTitle>
        {hasChanges && (
          <Badge variant="warning">Unsaved Changes</Badge>
        )}
      </PageHeader>

      <SettingsGrid>
        {/* Memory Settings */}
        <Card title="Memory Settings" subtitle="Configure Minecraft memory allocation">
          <SettingsSection>
            <SettingsRow>
              <SettingsLabel>
                <LabelText>Memory Allocation</LabelText>
                <LabelDescription>
                  Allocate memory for Minecraft. Recommended: {recommendedMemory}GB based on your system
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <MemorySlider
                  min="2"
                  max="32"
                  value={settings.memoryMb / 1024}
                  onChange={(e) => updateSettings({ memoryMb: parseInt(e.target.value) * 1024 })}
                />
                <MemoryDisplay>{settings.memoryMb / 1024}GB</MemoryDisplay>
              </SettingsControl>
            </SettingsRow>
          </SettingsSection>
        </Card>

        {/* Java Settings */}
        <Card title="Java Settings" subtitle="Configure Java installation">
          <SettingsSection>
            <SettingsRow>
              <SettingsLabel>
                <LabelText>Java Path</LabelText>
                <LabelDescription>
                  Path to Java executable. Leave empty to use auto-detection.
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <PathInputContainer>
                  <PathInput
                    value={settings.javaPath || ''}
                    onChange={(e) => updateSettings({ javaPath: e.target.value || undefined })}
                    placeholder="Auto-detect Java"
                  />
                  <BrowseButton
                    variant="outline"
                    size="sm"
                    onClick={() => browseForPath('java')}
                  >
                    Browse
                  </BrowseButton>
                </PathInputContainer>
              </SettingsControl>
            </SettingsRow>

            {javaVersions.length > 0 && (
              <SettingsRow>
                <SettingsLabel>
                  <LabelText>Detected Java Versions</LabelText>
                </SettingsLabel>
              </SettingsRow>
            )}

            <JavaVersionList>
              {javaVersions.map((java, index) => (
                <JavaVersionItem key={index}>
                  <JavaVersionInfo>
                    <span>Java {java.version} ({java.is64bit ? '64-bit' : '32-bit'})</span>
                    <JavaVersionPath>{java.path}</JavaVersionPath>
                  </JavaVersionInfo>
                  <StatusIndicator
                    status={java.is64bit && java.majorVersion >= 17 ? 'success' : 'warning'}
                  >
                    {java.is64bit && java.majorVersion >= 17 ? 'Recommended' : 'May not work'}
                  </StatusIndicator>
                </JavaVersionItem>
              ))}
            </JavaVersionList>
          </SettingsSection>
        </Card>

        {/* Prism Launcher Settings */}
        <Card title="Prism Launcher" subtitle="Configure Prism Launcher installation">
          <SettingsSection>
            <SettingsRow>
              <SettingsLabel>
                <LabelText>Prism Launcher Path</LabelText>
                <LabelDescription>
                  Path to Prism Launcher executable. Leave empty to use auto-detection.
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <PathInputContainer>
                  <PathInput
                    value={settings.prismPath || ''}
                    onChange={(e) => updateSettings({ prismPath: e.target.value || undefined })}
                    placeholder="Auto-detect Prism Launcher"
                  />
                  <BrowseButton
                    variant="outline"
                    size="sm"
                    onClick={() => browseForPath('prism')}
                  >
                    Browse
                  </BrowseButton>
                </PathInputContainer>
              </SettingsControl>
            </SettingsRow>

            {settings.prismPath && (
              <SettingsRow>
                <SettingsLabel>
                  <LabelText>Status</LabelText>
                </SettingsLabel>
                <StatusIndicator status="success">
                  Prism Launcher configured
                </StatusIndicator>
              </SettingsRow>
            )}
          </SettingsSection>
        </Card>

        {/* Theme Settings */}
        <Card title="Appearance" subtitle="Configure the launcher appearance">
          <SettingsSection>
            <SettingsRow>
              <SettingsLabel>
                <LabelText>Theme</LabelText>
                <LabelDescription>
                  Choose your preferred color theme
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <Select
                  value={settings.theme}
                  onChange={(e) => updateSettings({ theme: e.target.value as any })}
                >
                  <option value="light">Light</option>
                  <option value="dark">Dark</option>
                  <option value="system">System</option>
                </Select>
              </SettingsControl>
            </SettingsRow>
          </SettingsSection>
        </Card>

        {/* Advanced Settings */}
        <Card title="Advanced Settings" subtitle="Configure advanced launcher options">
          <SettingsSection>
            <SettingsRow>
              <SettingsLabel>
                <LabelText>Auto-update</LabelText>
                <LabelDescription>
                  Automatically check for and install launcher updates
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <Checkbox
                  checked={settings.autoUpdate}
                  onChange={(e) => updateSettings({ autoUpdate: e.target.checked })}
                />
              </SettingsControl>
            </SettingsRow>

            <SettingsRow>
              <SettingsLabel>
                <LabelText>Instances Directory</LabelText>
                <LabelDescription>
                  Custom directory for Minecraft instances. Leave empty for default.
                </LabelDescription>
              </SettingsLabel>
              <SettingsControl>
                <PathInputContainer>
                  <PathInput
                    value={settings.instancesDir || ''}
                    onChange={(e) => updateSettings({ instancesDir: e.target.value || undefined })}
                    placeholder="Default instances directory"
                  />
                  <BrowseButton
                    variant="outline"
                    size="sm"
                    onClick={() => browseForPath('instances')}
                  >
                    Browse
                  </BrowseButton>
                </PathInputContainer>
              </SettingsControl>
            </SettingsRow>
          </SettingsSection>
        </Card>

        {/* System Information */}
        <Card title="System Information" subtitle="Your system specifications">
          <SystemInfoContainer>
            <SystemInfoItem>
              <SystemInfoLabel>Operating System</SystemInfoLabel>
              <SystemInfoValue>{systemInfo.os} ({systemInfo.arch})</SystemInfoValue>
            </SystemInfoItem>
            <SystemInfoItem>
              <SystemInfoLabel>Total Memory</SystemInfoLabel>
              <SystemInfoValue>{Math.round(systemInfo.totalMemoryMb / 1024)}GB</SystemInfoValue>
            </SystemInfoItem>
            <SystemInfoItem>
              <SystemInfoLabel>Available Memory</SystemInfoLabel>
              <SystemInfoValue>{Math.round(systemInfo.availableMemoryMb / 1024)}GB</SystemInfoValue>
            </SystemInfoItem>
            <SystemInfoItem>
              <SystemInfoLabel>CPU Cores</SystemInfoLabel>
              <SystemInfoValue>{systemInfo.cpuCores}</SystemInfoValue>
            </SystemInfoItem>
            <SystemInfoItem>
              <SystemInfoLabel>Java Installed</SystemInfoLabel>
              <SystemInfoValue>
                <StatusIndicator status={systemInfo.javaInstalled ? 'success' : 'error'}>
                  {systemInfo.javaInstalled ? 'Yes' : 'No'}
                </StatusIndicator>
              </SystemInfoValue>
            </SystemInfoItem>
            <SystemInfoItem>
              <SystemInfoLabel>Java Versions</SystemInfoLabel>
              <SystemInfoValue>{javaVersions.length} found</SystemInfoValue>
            </SystemInfoItem>
          </SystemInfoContainer>
        </Card>
      </SettingsGrid>

      {/* Action Buttons */}
      <ActionsContainer>
        <Button
          variant="outline"
          onClick={() => setShowResetModal(true)}
          disabled={isSaving}
        >
          Reset to Defaults
        </Button>
        <Button
          variant="primary"
          onClick={saveSettings}
          disabled={!hasChanges || isSaving}
          loading={isSaving}
        >
          {isSaving ? 'Saving...' : 'Save Settings'}
        </Button>
      </ActionsContainer>

      {/* Reset Confirmation Modal */}
      <Modal
        isOpen={showResetModal}
        onClose={() => setShowResetModal(false)}
        title="Reset Settings"
        size="sm"
      >
        <p>Are you sure you want to reset all settings to their default values? This action cannot be undone.</p>

        <ActionsContainer>
          <Button
            variant="outline"
            onClick={() => setShowResetModal(false)}
          >
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={resetSettings}
          >
            Reset Settings
          </Button>
        </ActionsContainer>
      </Modal>
    </SettingsContainer>
  );
};

export default SettingsPage;