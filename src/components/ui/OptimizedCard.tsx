import React, { memo, useRef, useEffect, useCallback, useMemo, useState } from 'react';
import styled, { css } from 'styled-components';
import { createIntersectionObserver, VirtualScrollManager } from '../../utils/performance';

const CardContainer = styled.div<{
  interactive: boolean;
  elevated: boolean;
}>`
  background-color: var(--color-bg-secondary);
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  position: relative;
  overflow: hidden;

  /* Performance optimizations */
  contain: layout style paint;
  will-change: transform;
  transform: translateZ(0);

  /* Interactive styles */
  ${props => props.interactive && css`
    cursor: pointer;
    transition: all var(--transition-fast) ease-in-out;

    &:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-lg);
      border-color: var(--color-border-secondary);
    }

    &:active {
      transform: translateY(0);
    }
  `}

  /* Elevated styles */
  ${props => props.elevated && css`
    box-shadow: var(--shadow-md);
  `}

  /* Focus styles for accessibility */
  &:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }
`;

const CardHeader = styled.div`
  margin-bottom: var(--spacing-md);
`;

const CardTitle = styled.h3`
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
  line-height: 1.3;
`;

const CardSubtitle = styled.p`
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
  line-height: 1.4;
`;


const CardFooter = styled.div`
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const LazyContent = styled.div<{ loaded: boolean }>`
  opacity: ${props => props.loaded ? 1 : 0};
  transform: ${props => props.loaded ? 'translateY(0)' : 'translateY(10px)'};
  transition: opacity var(--transition-normal) ease-in-out,
              transform var(--transition-normal) ease-in-out;
`;

interface OptimizedCardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  footer?: React.ReactNode;
  interactive?: boolean;
  elevated?: boolean;
  lazy?: boolean;
  className?: string;
  onClick?: () => void;
  'aria-label'?: string;
  role?: string;
}

export const OptimizedCard = memo<OptimizedCardProps>(({
  children,
  title,
  subtitle,
  footer,
  interactive = false,
  elevated = false,
  lazy = false,
  className,
  onClick,
  'aria-label': ariaLabel,
  role = 'article',
}) => {
  const cardRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(!lazy);

  // Intersection Observer for lazy loading
  useEffect(() => {
    if (!lazy || isLoaded) return;

    const card = cardRef.current;
    if (!card) return;

    const observer = createIntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting && !isLoaded) {
            // Small delay for better UX
            setTimeout(() => setIsLoaded(true), 100);
          }
        });
      },
      { rootMargin: '50px' }
    );

    if (observer) {
      observer.observe(card);
      return () => observer.unobserve(card);
    }
  }, [lazy, isLoaded]);

  // Memoize click handler
  const handleClick = useCallback(() => {
    if (interactive && onClick) {
      onClick();
    }
  }, [interactive, onClick]);

  // Memoize header content
  const headerContent = useMemo(() => {
    if (!title && !subtitle) return null;

    return (
      <CardHeader>
        {title && <CardTitle>{title}</CardTitle>}
        {subtitle && <CardSubtitle>{subtitle}</CardSubtitle>}
      </CardHeader>
    );
  }, [title, subtitle]);

  // Memoize footer content
  const footerContent = useMemo(() => {
    if (!footer) return null;
    return <CardFooter>{footer}</CardFooter>;
  }, [footer]);

  // Computed accessibility props
  const accessibilityProps = useMemo(() => ({
    role,
    'aria-label': ariaLabel || title,
    tabIndex: interactive ? 0 : undefined,
  }), [role, ariaLabel, title, interactive]);

  return (
    <CardContainer
      ref={cardRef}
      interactive={interactive}
      elevated={elevated}
      className={className}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (interactive && (e.key === 'Enter' || e.key === ' ') && onClick) {
          e.preventDefault();
          onClick();
        }
      }}
      {...accessibilityProps}
    >
      {headerContent}

      <LazyContent loaded={isLoaded}>
        {children}
      </LazyContent>

      {footerContent}
    </CardContainer>
  );
});

OptimizedCard.displayName = 'OptimizedCard';

// Virtualized Card List Component
interface VirtualizedCardListProps {
  items: Array<{
    id: string;
    title?: string;
    subtitle?: string;
    content: React.ReactNode;
    height?: number;
  }>;
  itemHeight?: number;
  containerHeight?: number;
}

export const VirtualizedCardList = memo<VirtualizedCardListProps>(({
  items,
  itemHeight = 200,
  containerHeight = 600,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const scrollElementRef = useRef<HTMLDivElement>(null);

  const virtualManager = useMemo(() => {
    return new VirtualScrollManager(containerHeight, itemHeight);
  }, [containerHeight, itemHeight]);

  const [visibleItems, setVisibleItems] = useState(() =>
    virtualManager.getVisibleItems()
  );

  // Update virtual manager when items change
  useEffect(() => {
    const itemsWithHeight = items.map(item => ({
      ...item,
      height: item.height || itemHeight,
    }));
    virtualManager.setItems(itemsWithHeight);
    setVisibleItems(virtualManager.getVisibleItems());
  }, [items, itemHeight, virtualManager]);

  // Handle scroll events
  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const scrollTop = e.currentTarget.scrollTop;
      virtualManager.setScrollTop(scrollTop);
      setVisibleItems(virtualManager.getVisibleItems());
    },
    [virtualManager]
  );

  return (
    <div
      ref={containerRef}
      style={{ height: containerHeight, overflow: 'hidden' }}
    >
      <div
        ref={scrollElementRef}
        style={{
          height: containerHeight,
          overflowY: 'auto',
          position: 'relative',
        }}
        onScroll={handleScroll}
      >
        {/* Spacer for total height */}
        <div
          style={{
            height: visibleItems.totalHeight,
            position: 'relative',
          }}
        >
          {/* Visible items */}
          {visibleItems.items.map((item, index) => {
            const actualIndex = visibleItems.startIndex + index;
            const itemData = items[actualIndex];

            return (
              <div
                key={item.id}
                style={{
                  position: 'absolute',
                  top: visibleItems.offsetY + (index * itemHeight),
                  left: 0,
                  right: 0,
                  height: item.height || itemHeight,
                  padding: 'var(--spacing-sm)',
                }}
              >
                <OptimizedCard
                  interactive
                  elevated
                  title={itemData.title}
                  subtitle={itemData.subtitle}
                >
                  {itemData.content}
                </OptimizedCard>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
});

VirtualizedCardList.displayName = 'VirtualizedCardList';