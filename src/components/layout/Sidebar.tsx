import React from 'react';
import styled from 'styled-components';
import { Link, useLocation } from 'react-router-dom';

const SidebarContainer = styled.aside`
  width: 250px;
  background-color: var(--color-bg-secondary);
  border-right: 1px solid var(--color-border-primary);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
`;

const Logo = styled.div`
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-primary);
  text-align: center;

  h1 {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    color: var(--color-primary);
    margin: 0;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  p {
    font-size: var(--font-size-sm);
    color: var(--color-text-tertiary);
    margin: var(--spacing-xs) 0 0 0;
  }
`;

const Navigation = styled.nav`
  flex: 1;
  padding: var(--spacing-md);
`;

const NavigationList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
`;

const NavigationItem = styled.li`
  margin-bottom: var(--spacing-xs);
`;

const NavigationLink = styled(Link)<{
  active: boolean;
}>`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  text-decoration: none;
  color: ${({ active }) =>
    active ? 'var(--color-primary)' : 'var(--color-text-secondary)'
  };
  background-color: ${({ active }) =>
    active ? 'var(--color-bg-tertiary)' : 'transparent'
  };
  transition: all var(--transition-fast);
  font-weight: var(--font-weight-medium);

  &:hover {
    background-color: var(--color-bg-tertiary);
    color: var(--color-text-primary);
  }

  svg {
    width: 1.25rem;
    height: 1.25rem;
  }
`;

export const Sidebar: React.FC = () => {
  const location = useLocation();

  const navigationItems = [
    { id: 'home', label: 'Home', icon: '🏠', path: '/home' },
    { id: 'modpacks', label: 'Modpacks', icon: '📦', path: '/modpacks' },
    { id: 'instances', label: 'Instances', icon: '🎮', path: '/instances' },
    { id: 'launch', label: 'Launch Manager', icon: '🚀', path: '/launch' },
    { id: 'downloads', label: 'Downloads', icon: '⬇️', path: '/downloads' },
    { id: 'settings', label: 'Settings', icon: '⚙️', path: '/settings' },
  ];

  return (
    <SidebarContainer>
      <Logo>
        <h1>TheBoys</h1>
        <p>Launcher</p>
      </Logo>

      <Navigation>
        <NavigationList>
          {navigationItems.map((item) => (
            <NavigationItem key={item.id}>
              <NavigationLink
                to={item.path}
                active={location.pathname === item.path}
              >
                <span>{item.icon}</span>
                <span>{item.label}</span>
              </NavigationLink>
            </NavigationItem>
          ))}
        </NavigationList>
      </Navigation>
    </SidebarContainer>
  );
};

export default Sidebar;