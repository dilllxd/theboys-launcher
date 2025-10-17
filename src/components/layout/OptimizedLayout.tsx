import React, { memo, useMemo, useCallback, useState, useEffect } from 'react';
import styled, { ThemeProvider } from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { throttle, performance } from '../../utils/performance';

const LayoutContainer = styled.div`
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--color-bg-primary);
  overflow: hidden;
`;

const MainContent = styled.main<{ sidebarCollapsed: boolean }>`
  display: flex;
  flex: 1;
  overflow: hidden;
  transition: all var(--transition-normal) ease-in-out;
`;

const ContentArea = styled.div`
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--spacing-lg);

  /* Custom scrollbar for better performance */
  scrollbar-width: thin;
  scrollbar-color: var(--color-bg-quaternary) var(--color-bg-secondary);

  &::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    background: var(--color-bg-secondary);
  }

  &::-webkit-scrollbar-thumb {
    background: var(--color-bg-quaternary);
    border-radius: var(--radius-full);
  }

  /* Optimize scrolling performance */
  contain: layout style paint;
  transform: translateZ(0);
  will-change: scroll-position;
`;

interface OptimizedLayoutProps {
  children: React.ReactNode;
  className?: string;
}

export const OptimizedLayout = memo<OptimizedLayoutProps>(({ children, className }) => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);

  // Memoize sidebar toggle to prevent unnecessary re-renders
  const handleSidebarToggle = useCallback(() => {
    performance.mark('sidebar-toggle-start');
    setSidebarCollapsed(prev => !prev);
    performance.mark('sidebar-toggle-end');
    const duration = performance.measure('sidebar-toggle', 'sidebar-toggle-start', 'sidebar-toggle-end');
    console.log(`Sidebar toggle took: ${duration}ms`);
  }, []);

  // Optimize scroll handling with throttle
  const handleScroll = useMemo(
    () => throttle(() => {
      const scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
      setIsScrolled(scrollTop > 50);
    }, 16), // ~60fps
    []
  );

  // Add scroll listener with cleanup
  useEffect(() => {
    const contentArea = document.querySelector('[data-scrollable]');
    if (contentArea) {
      contentArea.addEventListener('scroll', handleScroll, { passive: true });
      return () => {
        contentArea.removeEventListener('scroll', handleScroll);
      };
    }
  }, [handleScroll]);

  // Memoize the theme object to prevent unnecessary re-renders
  const theme = useMemo(() => ({
    colors: {
      primary: 'var(--color-primary)',
      secondary: 'var(--color-bg-secondary)',
      text: 'var(--color-text-primary)',
    },
  }), []);

  // Memoize animation variants
  const contentVariants = useMemo(() => ({
    initial: { opacity: 0, x: sidebarCollapsed ? 20 : 0 },
    animate: {
      opacity: 1,
      x: 0,
      transition: {
        duration: 0.3,
        ease: 'easeOut',
        staggerChildren: 0.1
      }
    },
    exit: { opacity: 0, x: -20 }
  }), [sidebarCollapsed]);

  // Performance monitoring
  useEffect(() => {
    performance.mark('layout-mount-start');

    // Collect performance metrics after component mounts
    const timer = setTimeout(() => {
      performance.mark('layout-mount-end');
      const mountTime = performance.measure('layout-mount', 'layout-mount-start', 'layout-mount-end');
      console.log(`Layout mount time: ${mountTime}ms`);
    }, 100);

    return () => clearTimeout(timer);
  }, []);

  return (
    <ThemeProvider theme={theme}>
      <LayoutContainer className={className}>
        <Header
          onSidebarToggle={handleSidebarToggle}
          sidebarCollapsed={sidebarCollapsed}
          isScrolled={isScrolled}
        />

        <MainContent sidebarCollapsed={sidebarCollapsed}>
          <Sidebar collapsed={sidebarCollapsed} />

          <AnimatePresence mode="wait">
            <motion.div
              key={sidebarCollapsed ? 'collapsed' : 'expanded'}
              variants={contentVariants}
              initial="initial"
              animate="animate"
              exit="exit"
              style={{ flex: 1, overflow: 'hidden' }}
            >
              <ContentArea
                data-scrollable
                role="main"
                aria-label="Main content"
              >
                {children}
              </ContentArea>
            </motion.div>
          </AnimatePresence>
        </MainContent>
      </LayoutContainer>
    </ThemeProvider>
  );
});

OptimizedLayout.displayName = 'OptimizedLayout';