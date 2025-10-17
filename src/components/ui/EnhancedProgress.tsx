import { memo, useEffect, useState } from 'react';
import styled, { keyframes, css } from 'styled-components';

// Enhanced animations
const shimmer = keyframes`
  0% {
    background-position: -200px 0;
  }
  100% {
    background-position: calc(200px + 100%) 0;
  }
`;

const glow = keyframes`
  0%, 100% {
    box-shadow: 0 0 5px rgba(255, 107, 53, 0.5);
  }
  50% {
    box-shadow: 0 0 20px rgba(255, 107, 53, 0.8);
  }
`;

const ProgressContainer = styled.div<{
  size: 'sm' | 'md' | 'lg';
  variant: 'primary' | 'success' | 'warning' | 'error';
  animated: boolean;
}>`
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);

  /* Size variations */
  ${props => {
    switch (props.size) {
      case 'sm':
        return css`
          font-size: var(--font-size-xs);
        `;
      case 'lg':
        return css`
          font-size: var(--font-size-lg);
        `;
      default:
        return css`
          font-size: var(--font-size-sm);
        `;
    }
  }}
`;

const ProgressBar = styled.div<{
  size: 'sm' | 'md' | 'lg';
  variant: 'primary' | 'success' | 'warning' | 'error';
  animated: boolean;
}>`
  position: relative;
  width: 100%;
  background-color: var(--color-bg-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);

  /* Size variations */
  ${props => {
    switch (props.size) {
      case 'sm':
        return css`
          height: 6px;
        `;
      case 'lg':
        return css`
          height: 16px;
        `;
      default:
        return css`
          height: 10px;
        `;
    }
  }}

  /* Animation */
  ${props => props.animated && css`
    animation: ${glow} 2s ease-in-out infinite;
  `}
`;

const ProgressFill = styled.div<{
  progress: number;
  variant: 'primary' | 'success' | 'warning' | 'error';
  animated: boolean;
  striped: boolean;
}>`
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--transition-normal) ease-in-out;
  position: relative;
  overflow: hidden;

  /* Color variants */
  ${props => {
    switch (props.variant) {
      case 'success':
        return css`
          background-color: var(--color-success);
        `;
      case 'warning':
        return css`
          background-color: var(--color-warning);
        `;
      case 'error':
        return css`
          background-color: var(--color-error);
        `;
      default:
        return css`
          background-color: var(--color-primary);
        `;
    }
  }}

  /* Animated fill */
  ${props => props.animated && css`
    &::after {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 255, 255, 0.3),
        transparent
      );
      animation: ${shimmer} 2s infinite;
      background-size: 200px 100%;
    }
  `}

  /* Striped pattern */
  ${props => props.striped && css`
    background-image: linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.15) 25%,
      transparent 25%,
      transparent 50%,
      rgba(255, 255, 255, 0.15) 50%,
      rgba(255, 255, 255, 0.15) 75%,
      transparent 75%,
      transparent
    );
    background-size: 1rem 1rem;
  `}

  /* Animated stripes */
  ${props => props.striped && props.animated && css`
    animation: progress-bar-stripes 1s linear infinite;
  `}

  @keyframes progress-bar-stripes {
    0% {
      background-position: 1rem 0;
    }
    100% {
      background-position: 0 0;
    }
  }
`;

const ProgressLabel = styled.div<{
  showPercentage: boolean;
  showLabel: boolean;
  position: 'top' | 'bottom' | 'inline';
}>`
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);

  ${props => props.position === 'inline' && css`
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: var(--color-text-primary);
    font-weight: var(--font-weight-medium);
    z-index: 1;
  `}
`;

const ProgressText = styled.span`
  font-weight: var(--font-weight-medium);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ProgressPercentage = styled.span`
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  min-width: 45px;
  text-align: right;
`;

interface EnhancedProgressProps {
  value: number;
  max?: number;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'success' | 'warning' | 'error';
  animated?: boolean;
  striped?: boolean;
  showLabel?: boolean;
  showPercentage?: boolean;
  label?: string;
  position?: 'top' | 'bottom' | 'inline';
  className?: string;
  'aria-label'?: string;
  'aria-valuenow'?: number;
  'aria-valuemin'?: number;
  'aria-valuemax'?: number;
}

export const EnhancedProgress = memo<EnhancedProgressProps>(({
  value,
  max = 100,
  size = 'md',
  variant = 'primary',
  animated = false,
  striped = false,
  showLabel = false,
  showPercentage = true,
  label,
  position = 'top',
  className,
  'aria-label': ariaLabel,
  'aria-valuenow': ariaValueNow,
  'aria-valuemin': ariaValueMin = 0,
  'aria-valuemax': ariaValueMax = max,
}) => {
  const [displayValue, setDisplayValue] = useState(0);
  const [isComplete, setIsComplete] = useState(false);

  // Smooth animation of progress value
  useEffect(() => {
    const targetValue = Math.min(Math.max(value, 0), max);
    const step = (targetValue - displayValue) / 20;

    if (Math.abs(targetValue - displayValue) > 0.5) {
      const timer = setTimeout(() => {
        setDisplayValue(prev => prev + step);
      }, 16);
      return () => clearTimeout(timer);
    } else {
      setDisplayValue(targetValue);
    }
  }, [value, displayValue, max]);

  // Check if progress is complete
  useEffect(() => {
    setIsComplete(displayValue >= max);
  }, [displayValue, max]);

  // Calculate percentage
  const percentage = Math.round((displayValue / max) * 100);

  // Generate accessibility label
  const accessibilityLabel = ariaLabel ||
    (label ? `${label}: ${percentage}%` : `Progress: ${percentage}%`);

  // Render label content
  const renderLabel = () => {
    if (!showLabel && !showPercentage) return null;
    if (position === 'inline') {
      return (
        <ProgressLabel
          showPercentage={showPercentage}
          showLabel={showLabel}
          position={position}
        >
          {showPercentage && `${percentage}%`}
        </ProgressLabel>
      );
    }

    return (
      <ProgressLabel
        showPercentage={showPercentage}
        showLabel={showLabel}
        position={position}
      >
        <ProgressText>{label}</ProgressText>
        {showPercentage && <ProgressPercentage>{percentage}%</ProgressPercentage>}
      </ProgressLabel>
    );
  };

  return (
    <ProgressContainer
      size={size}
      variant={variant}
      animated={animated && !isComplete}
      className={className}
    >
      {position === 'top' && renderLabel()}

      <ProgressBar
        size={size}
        variant={variant}
        animated={animated && !isComplete}
        role="progressbar"
        aria-label={accessibilityLabel}
        aria-valuenow={ariaValueNow || displayValue}
        aria-valuemin={ariaValueMin}
        aria-valuemax={ariaValueMax}
      >
        <ProgressFill
          progress={percentage}
          variant={variant}
          animated={animated && !isComplete}
          striped={striped}
        >
          {position === 'inline' && renderLabel()}
        </ProgressFill>
      </ProgressBar>

      {position === 'bottom' && renderLabel()}
    </ProgressContainer>
  );
});

EnhancedProgress.displayName = 'EnhancedProgress';

// Preset configurations for common use cases
export const CircularProgress = memo((props: Omit<EnhancedProgressProps, 'variant'>) => (
  <EnhancedProgress variant="primary" {...props} />
));

export const SuccessProgress = memo((props: Omit<EnhancedProgressProps, 'variant'>) => (
  <EnhancedProgress variant="success" {...props} />
));

export const WarningProgress = memo((props: Omit<EnhancedProgressProps, 'variant'>) => (
  <EnhancedProgress variant="warning" {...props} />
));

export const ErrorProgress = memo((props: Omit<EnhancedProgressProps, 'variant'>) => (
  <EnhancedProgress variant="error" {...props} />
));

CircularProgress.displayName = 'CircularProgress';
SuccessProgress.displayName = 'SuccessProgress';
WarningProgress.displayName = 'WarningProgress';
ErrorProgress.displayName = 'ErrorProgress';