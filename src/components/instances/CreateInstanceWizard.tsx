import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { InstanceConfig, Modloader, Modpack } from '../../types/launcher';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { Card } from '../ui/Card';
import { api } from '../../utils/api';
import toast from 'react-hot-toast';

interface CreateInstanceWizardProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

const WizardContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  min-height: 500px;
`;

const WizardHeader = styled.div`
  text-align: center;
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
`;

const WizardTitle = styled.h2`
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm) 0;
`;

const WizardDescription = styled.p`
  color: var(--color-text-secondary);
  margin: 0;
`;

const StepIndicator = styled.div`
  display: flex;
  justify-content: center;
  gap: var(--spacing-md);
  margin: var(--spacing-lg) 0;
`;

const Step = styled.div<{ active: boolean; completed: boolean }>`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: ${props => {
    if (props.completed) return 'var(--color-success-600)';
    if (props.active) return 'var(--color-primary-600)';
    return 'var(--color-gray-400)';
  }};

  font-weight: ${props => props.active ? 'var(--font-weight-semibold)' : 'var(--font-weight-normal)'};
`;

const StepNumber = styled.div<{ active: boolean; completed: boolean }>`
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: ${props => {
    if (props.completed) return 'var(--color-success-500)';
    if (props.active) return 'var(--color-primary-500)';
    return 'var(--color-gray-200)';
  }};
  color: white;
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
`;

const StepContent = styled.div`
  flex: 1;
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

  &:disabled {
    background: var(--color-gray-100);
    color: var(--color-gray-500);
  }
`;

const Select = styled.select`
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-base);
  background: white;
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


const ModpackGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-md);
  max-height: 400px;
  overflow-y: auto;
  padding: var(--spacing-sm);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-md);
`;

const ModpackCard = styled(Card)<{ selected: boolean }>`
  padding: var(--spacing-md);
  cursor: pointer;
  border: 2px solid ${props => props.selected ? 'var(--color-primary-500)' : 'transparent'};
  background: ${props => props.selected ? 'var(--color-primary-50)' : 'white'};
  transition: all 0.2s ease;

  &:hover {
    border-color: ${props => props.selected ? 'var(--color-primary-600)' : 'var(--color-gray-300)'};
  }
`;

const ModpackName = styled.h4`
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
`;

const ModpackDescription = styled.p`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0 0 var(--spacing-sm) 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
`;

const ModpackInfo = styled.div`
  display: flex;
  gap: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
`;

const WizardActions = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
`;

const LoadingContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--spacing-xl);
`;

const SummaryCard = styled(Card)`
  padding: var(--spacing-lg);
`;

const SummaryTitle = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-md) 0;
`;

const SummaryList = styled.dl`
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--spacing-sm) var(--spacing-md);
  margin: 0;
`;

const SummaryTerm = styled.dt`
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
`;

const SummaryDescription = styled.dd`
  color: var(--color-text-primary);
  margin: 0;
`;

type WizardStep = 'modpack' | 'configuration' | 'summary';

export const CreateInstanceWizard: React.FC<CreateInstanceWizardProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [currentStep, setCurrentStep] = useState<WizardStep>('modpack');
  const [loading, setLoading] = useState(false);
  const [modpacks, setModpacks] = useState<Modpack[]>([]);
  const [selectedModpack, setSelectedModpack] = useState<Modpack | null>(null);

  // Form state
  const [instanceName, setInstanceName] = useState('');
  const [selectedModloader, setSelectedModloader] = useState<Modloader>('vanilla');
  const [loaderVersion, setLoaderVersion] = useState('');
  const [memoryMb, setMemoryMb] = useState(4096);
  const [javaPath, setJavaPath] = useState('');
  const [jvmArgs, setJvmArgs] = useState('');

  // Modloader versions
  const [availableModloaderVersions, setAvailableModloaderVersions] = useState<string[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

  // Load modpacks on mount
  useEffect(() => {
    if (isOpen && modpacks.length === 0) {
      loadModpacks();
    }
  }, [isOpen]);

  // Load Java path on mount
  useEffect(() => {
    if (isOpen && !javaPath) {
      loadJavaPath();
    }
  }, [isOpen]);

  // Load modloader versions when selection changes
  useEffect(() => {
    if (selectedModpack && selectedModloader !== 'vanilla') {
      loadModloaderVersions();
    }
  }, [selectedModpack, selectedModloader]);

  const loadModpacks = async () => {
    try {
      const modpacks = await api.getAvailableModpacks();
      setModpacks(modpacks);
    } catch (error) {
      toast.error('Failed to load modpacks');
      console.error('Load modpacks error:', error);
    }
  };

  const loadJavaPath = async () => {
    try {
      const settings = await api.getSettings();
      setJavaPath(settings.javaPath || '');
    } catch (error) {
      console.error('Load settings error:', error);
    }
  };

  const loadModloaderVersions = async () => {
    if (!selectedModpack) return;

    setLoadingVersions(true);
    try {
      const versions = await api.getModloaderVersions(
        selectedModloader,
        selectedModpack.minecraftVersion
      );
      setAvailableModloaderVersions(versions);
      if (versions.length > 0 && !loaderVersion) {
        setLoaderVersion(versions[0]);
      }
    } catch (error) {
      toast.error('Failed to load modloader versions');
      console.error('Load modloader versions error:', error);
    } finally {
      setLoadingVersions(false);
    }
  };

  const handleModpackSelect = (modpack: Modpack) => {
    setSelectedModpack(modpack);
    setInstanceName(modpack.instanceName);
    setSelectedModloader(modpack.modloader);
    setLoaderVersion(modpack.loaderVersion);
  };

  const handleNext = () => {
    switch (currentStep) {
      case 'modpack':
        if (selectedModpack) {
          setCurrentStep('configuration');
        }
        break;
      case 'configuration':
        setCurrentStep('summary');
        break;
      case 'summary':
        handleCreate();
        break;
    }
  };

  const handlePrevious = () => {
    switch (currentStep) {
      case 'configuration':
        setCurrentStep('modpack');
        break;
      case 'summary':
        setCurrentStep('configuration');
        break;
    }
  };

  const handleCreate = async () => {
    if (!selectedModpack) return;

    setLoading(true);
    try {
      const config: InstanceConfig = {
        name: instanceName,
        modpackId: selectedModpack.id,
        minecraftVersion: selectedModpack.minecraftVersion,
        loaderType: selectedModloader,
        loaderVersion: selectedModloader === 'vanilla' ? '' : loaderVersion,
        memoryMb,
        javaPath,
        jvmArgs: jvmArgs || undefined,
      };

      await api.createInstance(config);
      toast.success('Instance created successfully!');
      onSuccess();
      handleClose();
    } catch (error) {
      toast.error('Failed to create instance');
      console.error('Create instance error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    // Reset state
    setCurrentStep('modpack');
    setSelectedModpack(null);
    setInstanceName('');
    setSelectedModloader('vanilla');
    setLoaderVersion('');
    setMemoryMb(4096);
    setJvmArgs('');
    setAvailableModloaderVersions([]);

    onClose();
  };

  const canProceed = () => {
    switch (currentStep) {
      case 'modpack':
        return selectedModpack !== null;
      case 'configuration':
        return instanceName.trim() !== '' && javaPath.trim() !== '';
      case 'summary':
        return true;
      default:
        return false;
    }
  };

  const steps = [
    { id: 'modpack', label: 'Select Modpack' },
    { id: 'configuration', label: 'Configure' },
    { id: 'summary', label: 'Summary' }
  ] as const;

  const renderStepContent = () => {
    switch (currentStep) {
      case 'modpack':
        return (
          <StepContent>
            <Label>Choose a modpack to install</Label>
            <ModpackGrid>
              {modpacks.map((modpack) => (
                <ModpackCard
                  key={modpack.id}
                  selected={selectedModpack?.id === modpack.id}
                  onClick={() => handleModpackSelect(modpack)}
                >
                  <ModpackName>{modpack.displayName}</ModpackName>
                  <ModpackDescription>{modpack.description}</ModpackDescription>
                  <ModpackInfo>
                    <span>🎮 {modpack.minecraftVersion}</span>
                    <span>⚙️ {modpack.modloader}</span>
                    {modpack.default && <span>⭐ Default</span>}
                  </ModpackInfo>
                </ModpackCard>
              ))}
            </ModpackGrid>
          </StepContent>
        );

      case 'configuration':
        return (
          <StepContent>
            <Form>
              <FormGroup>
                <Label htmlFor="instanceName">Instance Name</Label>
                <Input
                  id="instanceName"
                  type="text"
                  value={instanceName}
                  onChange={(e) => setInstanceName(e.target.value)}
                  placeholder="Enter instance name"
                  required
                />
              </FormGroup>

              <FormGroup>
                <Label htmlFor="modloader">Modloader</Label>
                <Select
                  id="modloader"
                  value={selectedModloader}
                  onChange={(e) => setSelectedModloader(e.target.value as Modloader)}
                  disabled={selectedModpack?.modloader !== 'vanilla'}
                >
                  <option value="vanilla">Vanilla</option>
                  <option value="forge">Forge</option>
                  <option value="fabric">Fabric</option>
                  <option value="quilt">Quilt</option>
                  <option value="neoforge">NeoForge</option>
                </Select>
                {selectedModpack?.modloader !== 'vanilla' && (
                  <small style={{ color: 'var(--color-text-secondary)' }}>
                    Modloader is predetermined by the modpack
                  </small>
                )}
              </FormGroup>

              {selectedModloader !== 'vanilla' && (
                <FormGroup>
                  <Label htmlFor="loaderVersion">
                    Modloader Version
                    {loadingVersions && <LoadingSpinner size="sm" />}
                  </Label>
                  <Select
                    id="loaderVersion"
                    value={loaderVersion}
                    onChange={(e) => setLoaderVersion(e.target.value)}
                    disabled={loadingVersions || availableModloaderVersions.length === 0}
                  >
                    {availableModloaderVersions.map((version) => (
                      <option key={version} value={version}>
                        {version}
                      </option>
                    ))}
                  </Select>
                </FormGroup>
              )}

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
                  placeholder="Path to Java executable"
                  required
                />
              </FormGroup>

              <FormGroup>
                <Label htmlFor="jvmArgs">JVM Arguments (Optional)</Label>
                <Input
                  id="jvmArgs"
                  type="text"
                  value={jvmArgs}
                  onChange={(e) => setJvmArgs(e.target.value)}
                  placeholder="Additional JVM arguments"
                />
              </FormGroup>
            </Form>
          </StepContent>
        );

      case 'summary':
        return (
          <StepContent>
            <SummaryCard>
              <SummaryTitle>Instance Summary</SummaryTitle>
              <SummaryList>
                <SummaryTerm>Name:</SummaryTerm>
                <SummaryDescription>{instanceName}</SummaryDescription>

                <SummaryTerm>Modpack:</SummaryTerm>
                <SummaryDescription>{selectedModpack?.displayName}</SummaryDescription>

                <SummaryTerm>Minecraft:</SummaryTerm>
                <SummaryDescription>{selectedModpack?.minecraftVersion}</SummaryDescription>

                <SummaryTerm>Modloader:</SummaryTerm>
                <SummaryDescription>
                  {selectedModloader === 'vanilla'
                    ? 'Vanilla'
                    : `${selectedModloader} ${loaderVersion}`
                  }
                </SummaryDescription>

                <SummaryTerm>Memory:</SummaryTerm>
                <SummaryDescription>{memoryMb}MB</SummaryDescription>

                <SummaryTerm>Java:</SummaryTerm>
                <SummaryDescription>{javaPath}</SummaryDescription>

                {jvmArgs && (
                  <>
                    <SummaryTerm>JVM Args:</SummaryTerm>
                    <SummaryDescription>{jvmArgs}</SummaryDescription>
                  </>
                )}
              </SummaryList>
            </SummaryCard>
          </StepContent>
        );

      default:
        return null;
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create New Instance"
      size="xl"
    >
      <WizardContainer>
        <WizardHeader>
          <WizardTitle>Create New Instance</WizardTitle>
          <WizardDescription>
            Set up a new Minecraft instance with your preferred modpack and settings
          </WizardDescription>
        </WizardHeader>

        <StepIndicator>
          {steps.map((step, index) => (
            <Step
              key={step.id}
              active={currentStep === step.id}
              completed={steps.findIndex(s => s.id === currentStep) > index}
            >
              <StepNumber
                active={currentStep === step.id}
                completed={steps.findIndex(s => s.id === currentStep) > index}
              >
                {steps.findIndex(s => s.id === currentStep) > index ? '✓' : index + 1}
              </StepNumber>
              <span>{step.label}</span>
            </Step>
          ))}
        </StepIndicator>

        {loading ? (
          <LoadingContainer>
            <LoadingSpinner size="lg" />
          </LoadingContainer>
        ) : (
          renderStepContent()
        )}

        <WizardActions>
          <Button
            variant="outline"
            onClick={handlePrevious}
            disabled={currentStep === 'modpack'}
          >
            Previous
          </Button>

          <Button
            variant="primary"
            onClick={handleNext}
            disabled={!canProceed()}
            loading={loading}
          >
            {currentStep === 'summary' ? 'Create Instance' : 'Next'}
          </Button>
        </WizardActions>
      </WizardContainer>
    </Modal>
  );
};