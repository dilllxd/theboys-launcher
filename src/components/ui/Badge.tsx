import React from 'react';
import styled from 'styled-components';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'outline';
  size?: 'sm' | 'md';
  className?: string;
}

const StyledBadge = styled.span<{
  variant: BadgeProps['variant'];
  size: BadgeProps['size'];
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-full);
  white-space: nowrap;

  ${({ size }) => {
    switch (size) {
      case 'sm':
        return `
          padding: var(--spacing-xs) var(--spacing-sm);
          font-size: var(--font-size-xs);
        `;
      default: // md
        return `
          padding: var(--spacing-xs) var(--spacing-md);
          font-size: var(--font-size-sm);
        `;
    }
  }}

  ${({ variant }) => {
    switch (variant) {
      case 'secondary':
        return `
          background-color: var(--color-bg-tertiary);
          color: var(--color-text-secondary);
          border: 1px solid var(--color-border-primary);
        `;
      case 'success':
        return `
          background-color: var(--color-success);
          color: var(--color-text-primary);
        `;
      case 'warning':
        return `
          background-color: var(--color-warning);
          color: var(--color-text-primary);
        `;
      case 'error':
        return `
          background-color: var(--color-error);
          color: var(--color-text-primary);
        `;
      case 'outline':
        return `
          background-color: transparent;
          color: var(--color-text-secondary);
          border: 1px solid var(--color-border-primary);
        `;
      default: // primary
        return `
          background-color: var(--color-primary);
          color: var(--color-text-primary);
        `;
    }
  }}
`;

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  className,
}) => {
  return (
    <StyledBadge variant={variant} size={size} className={className}>
      {children}
    </StyledBadge>
  );
};

export default Badge;