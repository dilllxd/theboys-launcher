import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { invoke } from '../types/mock-tauri';
import {
  Modpack,
  InstalledModpack,
  ModpackUpdate,
  Modloader
} from '../types';
import { Button, Card, LoadingSpinner, Badge, Modal } from '../components/ui';
import { formatBytes } from '../utils/format';

const ModpacksContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  padding: var(--spacing-lg);
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

const ControlsContainer = styled.div`
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
`;

const ViewToggle = styled.div`
  display: flex;
  background: var(--color-surface);
  border-radius: var(--radius-md);
  padding: 4px;
`;

const ViewButton = styled.button<{ active: boolean }>`
  padding: var(--spacing-sm) var(--spacing-md);
  border: none;
  background: ${props => props.active ? 'var(--color-primary)' : 'transparent'};
  color: ${props => props.active ? 'var(--color-primary-foreground)' : 'var(--color-text-secondary)'};
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: ${props => props.active ? 'var(--color-primary)' : 'var(--color-surface-hover)'};
  }
`;

const SearchContainer = styled.div`
  position: relative;
  flex: 1;
  max-width: 400px;
`;

const SearchInput = styled.input`
  width: 100%;
  padding: var(--spacing-md) var(--spacing-lg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: var(--font-size-md);

  &::placeholder {
    color: var(--color-text-placeholder);
  }

  &:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(var(--color-primary-rgb), 0.2);
  }
`;

const FilterContainer = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
`;

const FilterChip = styled.button<{ active: boolean }>`
  padding: var(--spacing-xs) var(--spacing-md);
  border: 1px solid var(--color-border);
  background: ${props => props.active ? 'var(--color-primary)' : 'var(--color-surface)'};
  color: ${props => props.active ? 'var(--color-primary-foreground)' : 'var(--color-text-secondary)'};
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    border-color: var(--color-primary);
    background: ${props => props.active ? 'var(--color-primary)' : 'var(--color-surface-hover)'};
  }
`;

const ModpacksGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-lg);
`;

const ModpacksList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
`;

const ModpackCard = styled(Card)`
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  }

  &.installed {
    border-left: 4px solid var(--color-success);
  }

  &.has-update {
    border-left: 4px solid var(--color-warning);
  }
`;

const ModpackHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-md);
`;

const ModpackTitle = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
`;

const ModpackDescription = styled.p`
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin: 0;
  line-height: 1.5;
`;

const ModpackBadges = styled.div`
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
  margin-top: var(--spacing-sm);
`;

const ModpackDetails = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  margin: var(--spacing-md) 0;
`;

const DetailRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--font-size-sm);
`;

const DetailLabel = styled.span`
  color: var(--color-text-secondary);
`;

const DetailValue = styled.span`
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
`;

const ModpackActions = styled.div`
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
`;

const StatusIndicator = styled.div<{ status: 'installed' | 'update' | 'available' }>`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: ${props => {
    switch (props.status) {
      case 'installed': return 'var(--color-success)';
      case 'update': return 'var(--color-warning)';
      case 'available': return 'var(--color-primary)';
    }
  }};
`;

const LoadingContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
`;

const UpdateBadge = styled(Badge)`
  position: absolute;
  top: var(--spacing-md);
  right: var(--spacing-md);
  background: var(--color-warning);
  color: var(--color-warning-foreground);
  font-size: var(--font-size-xs);
  padding: 4px 8px;
`;

const DefaultBadge = styled(Badge)`
  background: var(--color-primary);
  color: var(--color-primary-foreground);
  font-size: var(--font-size-xs);
  padding: 4px 8px;
`;

type ViewMode = 'grid' | 'list';
type FilterModloader = 'all' | Modloader;

export const ModpacksPage: React.FC = () => {
  const [modpacks, setModpacks] = useState<Modpack[]>([]);
  const [installedModpacks, setInstalledModpacks] = useState<InstalledModpack[]>([]);
  const [updates, setUpdates] = useState<ModpackUpdate[]>([]);
  const [defaultModpack, setDefaultModpack] = useState<Modpack | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedModloader, setSelectedModloader] = useState<FilterModloader>('all');
  const [selectedModpack, setSelectedModpack] = useState<Modpack | null>(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  // Load data on component mount
  useEffect(() => {
    loadModpacks();
  }, []);

  const loadModpacks = async () => {
    try {
      setLoading(true);
      setError(null);

      const [modpacksData, installedData, defaultData] = await Promise.all([
        invoke('get_available_modpacks'),
        invoke('get_installed_modpacks'),
        invoke('get_default_modpack')
      ]);

      setModpacks(modpacksData);
      setInstalledModpacks(installedData);
      setDefaultModpack(defaultData);

      // Check for updates
      if (installedData.length > 0) {
        const updatesData = await invoke('check_all_modpack_updates');
        setUpdates(updatesData);
      }
    } catch (err) {
      setError(err as string);
    } finally {
      setLoading(false);
    }
  };

  const isModpackInstalled = (modpackId: string) => {
    return installedModpacks.some(installed => installed.modpack.id === modpackId);
  };

  const hasUpdate = (modpackId: string) => {
    return updates.some(update => update.modpackId === modpackId && update.updateAvailable);
  };

  const getInstalledModpack = (modpackId: string) => {
    return installedModpacks.find(installed => installed.modpack.id === modpackId);
  };

  const filteredModpacks = modpacks.filter(modpack => {
    const matchesSearch = searchQuery === '' ||
      modpack.displayName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      modpack.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      modpack.id.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesModloader = selectedModloader === 'all' || modpack.modloader === selectedModloader;

    return matchesSearch && matchesModloader;
  });

  const handleInstallModpack = async (modpack: Modpack) => {
    try {
      const downloadId = await invoke('download_modpack', {
        modpackId: modpack.id
      });

      // TODO: Show download progress dialog
      console.log('Download started:', downloadId);

      // Refresh modpacks after a delay
      setTimeout(loadModpacks, 2000);
    } catch (err) {
      console.error('Failed to install modpack:', err);
      setError(err as string);
    }
  };

  const handleUpdateModpack = async (modpack: Modpack) => {
    try {
      const downloadId = await invoke('download_modpack', {
        modpackId: modpack.id
      });

      console.log('Update started:', downloadId);
      setTimeout(loadModpacks, 2000);
    } catch (err) {
      console.error('Failed to update modpack:', err);
      setError(err as string);
    }
  };

  const handleSetDefault = async (modpackId: string) => {
    try {
      await invoke('select_default_modpack', { modpackId });
      await loadModpacks();
    } catch (err) {
      console.error('Failed to set default modpack:', err);
      setError(err as string);
    }
  };

  const getModpackStatus = (modpack: Modpack) => {
    if (hasUpdate(modpack.id)) return 'update';
    if (isModpackInstalled(modpack.id)) return 'installed';
    return 'available';
  };

  const renderModpackCard = (modpack: Modpack) => {
    const installed = getInstalledModpack(modpack.id);
    const status = getModpackStatus(modpack);
    const isDefault = defaultModpack?.id === modpack.id;

    return (
      <ModpackCard
        key={modpack.id}
        className={`${isModpackInstalled(modpack.id) ? 'installed' : ''} ${hasUpdate(modpack.id) ? 'has-update' : ''}`}
        onClick={() => {
          setSelectedModpack(modpack);
          setShowDetailsModal(true);
        }}
      >
        <ModpackHeader>
          <StatusIndicator status={status} />
          <div style={{ flex: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-xs)' }}>
              <ModpackTitle>{modpack.displayName}</ModpackTitle>
              {isDefault && <DefaultBadge>Default</DefaultBadge>}
              {hasUpdate(modpack.id) && <UpdateBadge>Update Available</UpdateBadge>}
            </div>
            <ModpackDescription>{modpack.description}</ModpackDescription>
          </div>
        </ModpackHeader>

        <ModpackBadges>
          <Badge variant="outline">{modpack.minecraftVersion}</Badge>
          <Badge variant="outline">{modpack.modloader}</Badge>
          <Badge variant="outline">{modpack.version}</Badge>
        </ModpackBadges>

        <ModpackDetails>
          <DetailRow>
            <DetailLabel>Instance Name:</DetailLabel>
            <DetailValue>{modpack.instanceName}</DetailValue>
          </DetailRow>
          {installed && (
            <>
              <DetailRow>
                <DetailLabel>Installed Version:</DetailLabel>
                <DetailValue>{installed.installedVersion}</DetailValue>
              </DetailRow>
              <DetailRow>
                <DetailLabel>Size:</DetailLabel>
                <DetailValue>{formatBytes(installed.sizeBytes)}</DetailValue>
              </DetailRow>
            </>
          )}
        </ModpackDetails>

        <ModpackActions>
          {status === 'available' && (
            <Button
              onClick={(e?: React.MouseEvent) => {
                e?.stopPropagation();
                handleInstallModpack(modpack);
              }}
            >
              Install
            </Button>
          )}
          {status === 'installed' && (
            <>
              <Button
                variant="outline"
                onClick={(e?: React.MouseEvent) => {
                  e?.stopPropagation();
                  handleSetDefault(modpack.id);
                }}
                disabled={isDefault}
              >
                {isDefault ? 'Default' : 'Set Default'}
              </Button>
              <Button
                variant="secondary"
                onClick={(e?: React.MouseEvent) => {
                  e?.stopPropagation();
                  // TODO: Launch modpack
                }}
              >
                Launch
              </Button>
            </>
          )}
          {status === 'update' && (
            <Button
              variant="primary"
              onClick={(e?: React.MouseEvent) => {
                e?.stopPropagation();
                handleUpdateModpack(modpack);
              }}
            >
              Update
            </Button>
          )}
        </ModpackActions>
      </ModpackCard>
    );
  };

  if (loading) {
    return (
      <LoadingContainer>
        <LoadingSpinner size="lg" />
      </LoadingContainer>
    );
  }

  if (error) {
    return (
      <ModpacksContainer>
        <PageTitle>Modpacks</PageTitle>
        <EmptyState>
          <p>Error loading modpacks: {error}</p>
          <Button onClick={loadModpacks}>Retry</Button>
        </EmptyState>
      </ModpacksContainer>
    );
  }

  return (
    <ModpacksContainer>
      <PageHeader>
        <PageTitle>Modpacks</PageTitle>
        <ControlsContainer>
          <ViewToggle>
            <ViewButton
              active={viewMode === 'grid'}
              onClick={() => setViewMode('grid')}
            >
              Grid
            </ViewButton>
            <ViewButton
              active={viewMode === 'list'}
              onClick={() => setViewMode('list')}
            >
              List
            </ViewButton>
          </ViewToggle>
          <Button
            variant="outline"
            onClick={loadModpacks}
          >
            Refresh
          </Button>
        </ControlsContainer>
      </PageHeader>

      <ControlsContainer>
        <SearchContainer>
          <SearchInput
            type="text"
            placeholder="Search modpacks..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </SearchContainer>

        <FilterContainer>
          <FilterChip
            active={selectedModloader === 'all'}
            onClick={() => setSelectedModloader('all')}
          >
            All
          </FilterChip>
          {(['vanilla', 'forge', 'fabric', 'quilt', 'neoforge'] as Modloader[]).map(modloader => (
            <FilterChip
              key={modloader}
              active={selectedModloader === modloader}
              onClick={() => setSelectedModloader(modloader)}
            >
              {modloader.charAt(0).toUpperCase() + modloader.slice(1)}
            </FilterChip>
          ))}
        </FilterContainer>
      </ControlsContainer>

      {filteredModpacks.length === 0 ? (
        <EmptyState>
          <p>No modpacks found matching your criteria.</p>
        </EmptyState>
      ) : (
        viewMode === 'grid' ? (
          <ModpacksGrid>
            {filteredModpacks.map(renderModpackCard)}
          </ModpacksGrid>
        ) : (
          <ModpacksList>
            {filteredModpacks.map(renderModpackCard)}
          </ModpacksList>
        )
      )}

      {/* Modpack Details Modal */}
      <Modal
        isOpen={showDetailsModal}
        onClose={() => setShowDetailsModal(false)}
        title={selectedModpack?.displayName || 'Modpack Details'}
        size="md"
      >
        {selectedModpack && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-md)' }}>
            <p>{selectedModpack.description}</p>

            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: 'var(--spacing-sm)' }}>
              <strong>ID:</strong> <span>{selectedModpack.id}</span>
              <strong>Version:</strong> <span>{selectedModpack.version}</span>
              <strong>Minecraft:</strong> <span>{selectedModpack.minecraftVersion}</span>
              <strong>Modloader:</strong> <span>{selectedModpack.modloader} {selectedModpack.loaderVersion}</span>
              <strong>Instance:</strong> <span>{selectedModpack.instanceName}</span>
            </div>

            <div style={{ display: 'flex', gap: 'var(--spacing-sm)', justifyContent: 'flex-end' }}>
              <Button
                variant="outline"
                onClick={() => setShowDetailsModal(false)}
              >
                Close
              </Button>
              {getModpackStatus(selectedModpack) === 'available' && (
                <Button
                  onClick={() => {
                    handleInstallModpack(selectedModpack);
                    setShowDetailsModal(false);
                  }}
                >
                  Install
                </Button>
              )}
              {getModpackStatus(selectedModpack) === 'update' && (
                <Button
                  onClick={() => {
                    handleUpdateModpack(selectedModpack);
                    setShowDetailsModal(false);
                  }}
                >
                  Update
                </Button>
              )}
            </div>
          </div>
        )}
      </Modal>
    </ModpacksContainer>
  );
};

export default ModpacksPage;