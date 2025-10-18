import React from 'react';
import styled from 'styled-components';

const HeaderContainer = styled.header`
  height: 60px;
  background-color: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-lg);
  flex-shrink: 0;
`;

const HeaderTitle = styled.h2`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
`;

const HeaderActions = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
`;

const ThemeToggle = styled.button`
  background: none;
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: var(--font-size-lg);

  &:hover {
    background-color: var(--color-bg-tertiary);
    color: var(--color-text-primary);
    border-color: var(--color-border-secondary);
  }
`;

interface HeaderProps {
  title?: string;
  onSidebarToggle?: () => void;
  sidebarCollapsed?: boolean;
  isScrolled?: boolean;
}

export const Header: React.FC<HeaderProps> = ({
  title = "TheBoys Launcher",
  onSidebarToggle,
  sidebarCollapsed,
  isScrolled
}) => {
  // Use isScrolled for potential future styling (suppressing unused warning)
  void isScrolled;
  const [currentTheme, setCurrentTheme] = React.useState<'light' | 'dark'>('light');

  React.useEffect(() => {
    // Load theme from localStorage or system preference
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null;
    if (savedTheme) {
      setCurrentTheme(savedTheme);
      document.documentElement.setAttribute('data-theme', savedTheme);
    } else {
      // Check system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      const theme = prefersDark ? 'dark' : 'light';
      setCurrentTheme(theme);
      document.documentElement.setAttribute('data-theme', theme);
    }
  }, []);

  const handleThemeToggle = () => {
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    setCurrentTheme(newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
  };

  return (
    <HeaderContainer>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-md)' }}>
        {onSidebarToggle && (
          <ThemeToggle onClick={onSidebarToggle} title="Toggle sidebar">
            {sidebarCollapsed ? '☰' : '☰'}
          </ThemeToggle>
        )}
        <HeaderTitle>{title}</HeaderTitle>
      </div>
      <HeaderActions>
        <ThemeToggle onClick={handleThemeToggle} title="Toggle theme">
          {currentTheme === 'dark' ? '☀️' : '🌙'}
        </ThemeToggle>
      </HeaderActions>
    </HeaderContainer>
  );
};

export default Header;