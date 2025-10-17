import React from 'react';
import styled, { keyframes } from 'styled-components';
import { LoadingSpinnerProps } from '../../types/launcher';

const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
`;

const SpinnerContainer = styled.div<{
  size: LoadingSpinnerProps['size'];
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;

  ${({ size }) => {
    switch (size) {
      case 'sm':
        return 'width: 1rem; height: 1rem;';
      case 'lg':
        return 'width: 2rem; height: 2rem;';
      default: // md
        return 'width: 1.5rem; height: 1.5rem;';
    }
  }}
`;

const SpinnerCircle = styled.div<{
  size: LoadingSpinnerProps['size'];
}>`
  border: ${({ size }) => (size === 'sm' ? '2px' : '3px')} solid transparent;
  border-top: ${({ size }) => (size === 'sm' ? '2px' : '3px')} solid var(--color-primary);
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
  width: 100%;
  height: 100%;

  /* Optional: Add subtle pulse effect */
  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border: ${({ size }) => (size === 'sm' ? '2px' : '3px')} solid transparent;
    border-top: ${({ size }) => (size === 'sm' ? '2px' : '3px')} solid var(--color-primary-light);
    border-radius: 50%;
    animation: ${pulse} 2s ease-in-out infinite;
    opacity: 0.3;
  }
`;

const DotsContainer = styled.div`
  display: flex;
  gap: var(--spacing-xs);
  align-items: center;
`;

const Dot = styled.div<{ delay: number }>`
  width: 0.5rem;
  height: 0.5rem;
  background-color: var(--color-primary);
  border-radius: 50%;
  animation: ${pulse} 1.4s ease-in-out infinite both;
  animation-delay: ${({ delay }) => delay}s;
`;

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  className,
  variant = 'spinner',
}) => {
  if (variant === 'dots') {
    return (
      <DotsContainer className={className}>
        <Dot delay={0} />
        <Dot delay={0.2} />
        <Dot delay={0.4} />
      </DotsContainer>
    );
  }

  return (
    <SpinnerContainer size={size} className={className}>
      <SpinnerCircle size={size} />
    </SpinnerContainer>
  );
};

export default LoadingSpinner;