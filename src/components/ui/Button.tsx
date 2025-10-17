import React from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { ButtonProps } from '../../types/launcher';

const StyledButton = styled(motion.button)<{
  variant: ButtonProps['variant'];
  size: ButtonProps['size'];
  disabled: boolean;
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  font-family: var(--font-family-sans);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;

  /* Size variants */
  ${({ size }) => {
    switch (size) {
      case 'sm':
        return `
          padding: var(--spacing-sm) var(--spacing-md);
          font-size: var(--font-size-sm);
          min-height: 2rem;
        `;
      case 'lg':
        return `
          padding: var(--spacing-md) var(--spacing-xl);
          font-size: var(--font-size-lg);
          min-height: 3rem;
        `;
      default: // md
        return `
          padding: var(--spacing-sm) var(--spacing-lg);
          font-size: var(--font-size-base);
          min-height: 2.5rem;
        `;
    }
  }}

  /* Variant styles */
  ${({ variant, disabled }) => {
    if (disabled) {
      return `
        background-color: var(--color-bg-quaternary);
        color: var(--color-text-disabled);
        border-color: var(--color-border-secondary);
        cursor: not-allowed;
        opacity: 0.6;
      `;
    }

    switch (variant) {
      case 'secondary':
        return `
          background-color: var(--color-bg-tertiary);
          color: var(--color-text-primary);
          border-color: var(--color-border-primary);

          &:hover:not(:disabled) {
            background-color: var(--color-bg-quaternary);
            border-color: var(--color-border-secondary);
          }

          &:active:not(:disabled) {
            background-color: var(--color-bg-quaternary);
            transform: translateY(1px);
          }
        `;
      case 'outline':
        return `
          background-color: transparent;
          color: var(--color-primary);
          border-color: var(--color-primary);

          &:hover:not(:disabled) {
            background-color: var(--color-primary);
            color: var(--color-text-primary);
          }

          &:active:not(:disabled) {
            transform: translateY(1px);
          }
        `;
      case 'ghost':
        return `
          background-color: transparent;
          color: var(--color-text-secondary);

          &:hover:not(:disabled) {
            background-color: var(--color-bg-tertiary);
            color: var(--color-text-primary);
          }

          &:active:not(:disabled) {
            transform: translateY(1px);
          }
        `;
      default: // primary
        return `
          background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
          color: var(--color-text-primary);
          border-color: var(--color-primary);
          box-shadow: 0 2px 4px rgba(255, 107, 53, 0.2);

          &:hover:not(:disabled) {
            background: linear-gradient(135deg, var(--color-primary-light) 0%, var(--color-primary) 100%);
            box-shadow: 0 4px 8px rgba(255, 107, 53, 0.3);
            transform: translateY(-1px);
          }

          &:active:not(:disabled) {
            transform: translateY(1px);
            box-shadow: 0 2px 4px rgba(255, 107, 53, 0.2);
          }
        `;
    }
  }}

  &:focus-visible {
    outline: 2px solid var(--color-border-focus);
    outline-offset: 2px;
  }

  /* Loading state */
  &[data-loading="true"] {
    pointer-events: none;
    opacity: 0.8;
  }
`;

const LoadingIcon = styled.div`
  width: 1em;
  height: 1em;
  border: 2px solid transparent;
  border-top: 2px solid currentColor;
  border-radius: 50%;
  animation: spin 1s linear infinite;

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
`;

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  onClick,
  className,
  type = 'button',
  style,
  ...props
}) => {
  return (
    <StyledButton
      variant={variant}
      size={size}
      disabled={disabled || loading}
      onClick={onClick}
      type={type}
      className={className}
      style={style}
      data-loading={loading}
      whileHover={!disabled && !loading ? { scale: 1.02 } : {}}
      whileTap={!disabled && !loading ? { scale: 0.98 } : {}}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
      {...props}
    >
      {loading && <LoadingIcon />}
      {children}
    </StyledButton>
  );
};

export default Button;