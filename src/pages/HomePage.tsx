import React from 'react';
import styled from 'styled-components';
import { Card, Button } from '../components/ui';

const HomeContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
`;

const WelcomeSection = styled.div`
  text-align: center;
  padding: var(--spacing-xl) 0;
`;

const WelcomeTitle = styled.h1`
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-md);

  /* Gradient text effect */
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
`;

const WelcomeSubtitle = styled.p`
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-xl);
  line-height: 1.6;
`;

const QuickActionsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-lg);
`;

const StatusCard = styled(Card)`
  text-align: center;
`;

const StatusIcon = styled.div`
  font-size: var(--font-size-3xl);
  margin-bottom: var(--spacing-md);
`;

const StatusTitle = styled.h3`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-sm);
`;

const StatusDescription = styled.p`
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-lg);
`;

export const HomePage: React.FC = () => {
  const [isLoading, setIsLoading] = React.useState(false);

  const handleLaunchMinecraft = async () => {
    setIsLoading(true);
    try {
      // TODO: Implement Minecraft launch logic
      await new Promise(resolve => setTimeout(resolve, 2000));
      console.log('Minecraft launched!');
    } catch (error) {
      console.error('Failed to launch Minecraft:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <HomeContainer>
      <WelcomeSection>
        <WelcomeTitle>Welcome to TheBoys Launcher</WelcomeTitle>
        <WelcomeSubtitle>
          Your gateway to amazing Minecraft modpack experiences.
          Browse modpacks, manage instances, and launch your adventures.
        </WelcomeSubtitle>

        <Button
          variant="primary"
          size="lg"
          onClick={handleLaunchMinecraft}
          loading={isLoading}
        >
          {isLoading ? 'Launching...' : 'Quick Launch'}
        </Button>
      </WelcomeSection>

      <QuickActionsGrid>
        <StatusCard interactive>
          <StatusIcon>📦</StatusIcon>
          <StatusTitle>Modpacks Available</StatusTitle>
          <StatusDescription>
            Browse and install from our curated collection of modpacks
          </StatusDescription>
          <Button variant="outline">Browse Modpacks</Button>
        </StatusCard>

        <StatusCard interactive>
          <StatusIcon>🎮</StatusIcon>
          <StatusTitle>Instances</StatusTitle>
          <StatusDescription>
            Manage your Minecraft instances and configurations
          </StatusDescription>
          <Button variant="outline">Manage Instances</Button>
        </StatusCard>

        <StatusCard interactive>
          <StatusIcon>⚙️</StatusIcon>
          <StatusTitle>Settings</StatusTitle>
          <StatusDescription>
            Configure launcher settings and preferences
          </StatusDescription>
          <Button variant="outline">Open Settings</Button>
        </StatusCard>
      </QuickActionsGrid>
    </HomeContainer>
  );
};

export default HomePage;