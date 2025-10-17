import React, { memo, forwardRef, useCallback } from 'react';
import styled, { css } from 'styled-components';
import { OptimizedLoadingSpinner } from './OptimizedLoadingSpinner';

// Base button styles with performance optimizations
const ButtonBase = styled.button<{
  variant: 'primary' | 'secondary' | 'outline' | 'ghost';
  size: 'sm' | 'md' | 'lg';
  disabled: boolean;
  loading: boolean;
}>`
  /* Reset styles */
  border: none;
  margin: 0;
  padding: 0;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
  cursor: pointer;

  /* Layout */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  position: relative;
  overflow: hidden;

  /* Size variations */
  ${props => {
    switch (props.size) {
      case 'sm':
        return css`
          padding: var(--spacing-xs) var(--spacing-sm);
          font-size: var(--font-size-sm);
          border-radius: var(--radius-sm);
          min-height: 2rem;
        `;
      case 'lg':
        return css`
          padding: var(--spacing-md) var(--spacing-lg);
          font-size: var(--font-size-lg);
          border-radius: var(--radius-lg);
          min-height: 3rem;
        `;
      default:
        return css`
          padding: var(--spacing-sm) var(--spacing-md);
          font-size: var(--font-size-base);
          border-radius: var(--radius-md);
          min-height: 2.5rem;
        `;
    }
  }}

  /* Variant styles */
  ${props => {
    switch (props.variant) {
      case 'primary':
        return css`
          background-color: var(--color-primary);
          color: var(--color-text-primary);
          font-weight: var(--font-weight-medium);

          &:hover:not(:disabled) {
            background-color: var(--color-primary-dark);
          }

          &:active:not(:disabled) {
            background-color: var(--color-primary-light);
          }
        `;
      case 'secondary':
        return css`
          background-color: var(--color-bg-secondary);
          color: var(--color-text-primary);
          border: 1px solid var(--color-border-primary);

          &:hover:not(:disabled) {
            background-color: var(--color-bg-tertiary);
            border-color: var(--color-border-secondary);
          }
        `;
      case 'outline':
        return css`
          background-color: transparent;
          color: var(--color-primary);
          border: 1px solid var(--color-primary);

          &:hover:not(:disabled) {
            background-color: var(--color-primary);
            color: var(--color-text-primary);
          }
        `;
      case 'ghost':
        return css`
          background-color: transparent;
          color: var(--color-text-secondary);

          &:hover:not(:disabled) {
            background-color: var(--color-bg-secondary);
            color: var(--color-text-primary);
          }
        `;
    }
  }}

  /* State styles */
  ${props => props.disabled && css`
    opacity: 0.6;
    cursor: not-allowed;
    pointer-events: none;
  `}

  ${props => props.loading && css`
    cursor: wait;
    pointer-events: none;
  `}

  /* Performance optimizations */
  will-change: transform, background-color;
  transform: translateZ(0);
  contain: layout style paint;

  /* Smooth transitions */
  transition: all var(--transition-fast) ease-in-out;

  /* Focus styles for accessibility */
  &:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  /* Ripple effect for better UX */
  &::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.3);
    transform: translate(-50%, -50%);
    transition: width 0.3s ease, height 0.3s ease;
    pointer-events: none;
  }

  &:active::after {
    width: 300px;
    height: 300px;
  }
`;

const LoadingContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
`;

const TextContent = styled.span<{ loading: boolean }>`
  transition: opacity var(--transition-fast) ease-in-out;
  opacity: ${props => props.loading ? 0.7 : 1};
`;

interface OptimizedButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  className?: string;
  type?: 'button' | 'submit' | 'reset';
  onClick?: (event?: React.MouseEvent) => void;
  style?: React.CSSProperties;
  'aria-label'?: string;
  'aria-describedby'?: string;
}

export const OptimizedButton = memo(forwardRef<HTMLButtonElement, OptimizedButtonProps>(({
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  className,
  type = 'button',
  onClick,
  style,
  'aria-label': ariaLabel,
  'aria-describedby': ariaDescribedby,
}, ref) => {
  // Memoize click handler to prevent unnecessary re-renders
  const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
    if (disabled || loading) {
      event.preventDefault();
      return;
    }
    onClick?.(event);
  }, [disabled, loading, onClick]);

  // Generate appropriate aria-label
  const computedAriaLabel = ariaLabel || (loading ? 'Loading...' : undefined);

  return (
    <ButtonBase
      ref={ref}
      variant={variant}
      size={size}
      disabled={disabled || loading}
      loading={loading}
      className={className}
      type={type}
      onClick={handleClick}
      style={style}
      aria-label={computedAriaLabel}
      aria-busy={loading}
      aria-describedby={ariaDescribedby}
    >
      {loading && (
        <LoadingContainer>
          <OptimizedLoadingSpinner size="sm" variant="spinner" />
        </LoadingContainer>
      )}
      <TextContent loading={loading}>
        {children}
      </TextContent>
    </ButtonBase>
  );
}));

OptimizedButton.displayName = 'OptimizedButton';