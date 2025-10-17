import { memo } from 'react';
import styled, { keyframes, css } from 'styled-components';

// Enhanced animations
const pulse = keyframes`
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.1);
  }
`;

const rotate = keyframes`
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
`;

const statusConfig = {
  online: {
    color: '#10b981',
    label: 'Online',
    icon: '●',
  },
  offline: {
    color: '#ef4444',
    label: 'Offline',
    icon: '●',
  },
  loading: {
    color: '#3b82f6',
    label: 'Loading',
    icon: '⟳',
  },
  warning: {
    color: '#f59e0b',
    label: 'Warning',
    icon: '⚠',
  },
  success: {
    color: '#10b981',
    label: 'Success',
    icon: '✓',
  },
  error: {
    color: '#ef4444',
    label: 'Error',
    icon: '✕',
  },
  info: {
    color: '#3b82f6',
    label: 'Info',
    icon: 'ℹ',
  },
};

type StatusType = keyof typeof statusConfig;

const StatusContainer = styled.div<{
  status: StatusType;
  size: 'sm' | 'md' | 'lg';
  animated: boolean;
  showLabel: boolean;
}>`
  display: flex;
  align-items: center;
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

const StatusIndicator = styled.span<{
  status: StatusType;
  size: 'sm' | 'md' | 'lg';
  animated: boolean;
}>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: ${props => statusConfig[props.status].color};
  font-weight: var(--font-weight-bold);

  /* Size variations */
  ${props => {
    switch (props.size) {
      case 'sm':
        return css`
          width: 12px;
          height: 12px;
          font-size: 8px;
        `;
      case 'lg':
        return css`
          width: 24px;
          height: 24px;
          font-size: 16px;
        `;
      default:
        return css`
          width: 16px;
          height: 16px;
          font-size: 12px;
        `;
    }
  }}

  /* Background for better visibility */
  background-color: ${props => `${statusConfig[props.status].color}20`};
  border: 2px solid ${props => statusConfig[props.status].color};

  /* Animations */
  ${props => props.animated && (
    props.status === 'loading' ? css`
      animation: ${rotate} 1s linear infinite;
    ` : css`
      animation: ${pulse} 2s ease-in-out infinite;
    `
  )}

  /* Performance optimizations */
  will-change: transform, opacity;
  transform: translateZ(0);
`;

const StatusLabel = styled.span<{
  status: StatusType;
}>`
  color: ${props => statusConfig[props.status].color};
  font-weight: var(--font-weight-medium);
  text-transform: capitalize;
  white-space: nowrap;
`;

const StatusTooltip = styled.div`
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-primary);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all var(--transition-fast) ease-in-out;
  z-index: 1000;
  margin-bottom: var(--spacing-xs);
  box-shadow: var(--shadow-md);

  &::after {
    content: '';
    position: absolute;
    top: 100%;
    left: 50%;
    transform: translateX(-50%);
    border: 4px solid transparent;
    border-top-color: var(--color-bg-tertiary);
  }
`;

const StatusWrapper = styled.div`
  position: relative;
  display: inline-flex;

  &:hover ${StatusTooltip} {
    opacity: 1;
    visibility: visible;
  }
`;

interface EnhancedStatusIndicatorProps {
  status: StatusType;
  size?: 'sm' | 'md' | 'lg';
  animated?: boolean;
  showLabel?: boolean;
  label?: string;
  tooltip?: string;
  className?: string;
  'aria-label'?: string;
}

export const EnhancedStatusIndicator = memo<EnhancedStatusIndicatorProps>(({
  status,
  size = 'md',
  animated = false,
  showLabel = false,
  label,
  tooltip,
  className,
  'aria-label': ariaLabel,
}) => {
  const config = statusConfig[status];
  const displayLabel = label || config.label;
  const tooltipText = tooltip || `${config.label} Status`;

  return (
    <StatusWrapper className={className}>
      <StatusContainer
        status={status}
        size={size}
        animated={animated}
        showLabel={showLabel}
      >
        <StatusIndicator
          status={status}
          size={size}
          animated={animated}
          role="status"
          aria-label={ariaLabel || displayLabel}
        >
          {config.icon}
        </StatusIndicator>

        {showLabel && (
          <StatusLabel status={status}>
            {displayLabel}
          </StatusLabel>
        )}
      </StatusContainer>

      <StatusTooltip role="tooltip">
        {tooltipText}
      </StatusTooltip>
    </StatusWrapper>
  );
});

EnhancedStatusIndicator.displayName = 'EnhancedStatusIndicator';

// Preset configurations for common use cases
export const ConnectionStatus = memo((props: Omit<EnhancedStatusIndicatorProps, 'status'>) => (
  <EnhancedStatusIndicator status="online" {...props} />
));

export const LoadingStatus = memo((props: Omit<EnhancedStatusIndicatorProps, 'status' | 'animated'>) => (
  <EnhancedStatusIndicator status="loading" animated {...props} />
));

export const ErrorStatus = memo((props: Omit<EnhancedStatusIndicatorProps, 'status'>) => (
  <EnhancedStatusIndicator status="error" {...props} />
));

export const SuccessStatus = memo((props: Omit<EnhancedStatusIndicatorProps, 'status'>) => (
  <EnhancedStatusIndicator status="success" {...props} />
));

ConnectionStatus.displayName = 'ConnectionStatus';
LoadingStatus.displayName = 'LoadingStatus';
ErrorStatus.displayName = 'ErrorStatus';
SuccessStatus.displayName = 'SuccessStatus';