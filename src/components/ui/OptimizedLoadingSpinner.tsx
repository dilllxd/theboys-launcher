import React, { memo, useEffect, useRef } from 'react';
import styled, { keyframes } from 'styled-components';

// Performance-optimized keyframes
const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
`;

const bounce = keyframes`
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
`;

const Container = styled.div<{ size: 'sm' | 'md' | 'lg' }>`
  display: flex;
  align-items: center;
  justify-content: center;
  width: ${props => {
    switch (props.size) {
      case 'sm': return '1rem';
      case 'md': return '2rem';
      case 'lg': return '3rem';
      default: return '2rem';
    }
  }};
  height: ${props => {
    switch (props.size) {
      case 'sm': return '1rem';
      case 'md': return '2rem';
      case 'lg': return '3rem';
      default: return '2rem';
    }
  }};
`;

// Optimized spinner using CSS transforms for better performance
const SpinnerContainer = styled.div<{ variant: 'spinner' | 'dots' | 'pulse' }>`
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
`;

const SpinnerCircle = styled.div`
  width: 100%;
  height: 100%;
  border: 2px solid transparent;
  border-top: 2px solid var(--color-primary);
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
  will-change: transform;
  transform: translateZ(0);
`;

const DotsContainer = styled.div`
  display: flex;
  gap: 4px;
  align-items: center;
  justify-content: center;
`;

const Dot = styled.div<{ delay: number }>`
  width: 6px;
  height: 6px;
  background-color: var(--color-primary);
  border-radius: 50%;
  animation: ${bounce} 1.4s ease-in-out infinite both;
  animation-delay: ${props => props.delay}s;
  will-change: transform;
`;

const PulseCircle = styled.div`
  width: 100%;
  height: 100%;
  background-color: var(--color-primary);
  border-radius: 50%;
  animation: ${pulse} 2s ease-in-out infinite;
  will-change: opacity;
`;

interface OptimizedLoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  variant?: 'spinner' | 'dots' | 'pulse';
  className?: string;
  'aria-label'?: string;
}

export const OptimizedLoadingSpinner = memo<OptimizedLoadingSpinnerProps>(({
  size = 'md',
  variant = 'spinner',
  className,
  'aria-label': ariaLabel = 'Loading...'
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  // Use Intersection Observer to pause animations when not visible
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          container.style.animationPlayState = 'running';
        } else {
          container.style.animationPlayState = 'paused';
        }
      },
      { threshold: 0.1 }
    );

    observer.observe(container);
    return () => observer.unobserve(container);
  }, []);

  // Render the appropriate variant
  const renderVariant = () => {
    switch (variant) {
      case 'dots':
        return (
          <DotsContainer>
            <Dot delay={0} />
            <Dot delay={0.2} />
            <Dot delay={0.4} />
          </DotsContainer>
        );

      case 'pulse':
        return <PulseCircle />;

      default:
        return <SpinnerCircle />;
    }
  };

  return (
    <Container
      ref={containerRef}
      size={size}
      className={className}
      role="status"
      aria-label={ariaLabel}
    >
      <SpinnerContainer variant={variant}>
        {renderVariant()}
      </SpinnerContainer>
    </Container>
  );
});

OptimizedLoadingSpinner.displayName = 'OptimizedLoadingSpinner';