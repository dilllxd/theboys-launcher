import React from 'react';
import { useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { Sidebar, Header } from './';

const LayoutContainer = styled.div`
  display: flex;
  height: 100vh;
  background-color: var(--color-bg-primary);
`;

const MainContent = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
`;

const ContentArea = styled.div`
  flex: 1;
  padding: var(--spacing-lg);
  overflow-y: auto;
  background-color: var(--color-bg-primary);
`;

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();

  // Get page title based on current route
  const getPageTitle = () => {
    switch (location.pathname) {
      case '/home':
        return 'Home';
      case '/modpacks':
        return 'Modpacks';
      case '/instances':
        return 'Instances';
      case '/settings':
        return 'Settings';
      default:
        return 'TheBoys Launcher';
    }
  };

  return (
    <LayoutContainer>
      <Sidebar />
      <MainContent>
        <Header title={getPageTitle()} />
        <ContentArea>
          {children}
        </ContentArea>
      </MainContent>
    </LayoutContainer>
  );
};

export default Layout;