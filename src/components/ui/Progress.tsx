import React from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { ProgressProps } from '../../types/launcher';

const ProgressContainer = styled.div<{
  size: ProgressProps['size'];
}>`
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);

  ${({ size }) => {
    switch (size) {
      case 'sm':
        return 'font-size: var(--font-size-xs);';
      case 'lg':
        return 'font-size: var(--font-size-base);';
      default: // md
        return 'font-size: var(--font-size-sm);';
    }
  }}
`;

const ProgressBar = styled.div<{
  size: ProgressProps['size'];
  color: ProgressProps['color'];
}>`
  background-color: var(--color-bg-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
  position: relative;

  ${({ size }) => {
    switch (size) {
      case 'sm':
        return 'height: 0.25rem;';
      case 'lg':
        return 'height: 0.75rem;';
      default: // md
        return 'height: 0.5rem;';
    }
  }}
`;

const ProgressFill = styled(motion.div)<{
  color: ProgressProps['color'];
}>`
  height: 100%;
  border-radius: inherit;
  background: ${({ color }) => {
    switch (color) {
      case 'success':
        return 'var(--color-success)';
      case 'warning':
        return 'var(--color-warning)';
      case 'error':
        return 'var(--color-error)';
      default: // primary
        return 'linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-light) 100%)';
    }
  }};
  box-shadow: ${({ color }) => {
    switch (color) {
      case 'success':
        return '0 0 8px rgba(56, 161, 105, 0.4)';
      case 'warning':
        return '0 0 8px rgba(214, 158, 46, 0.4)';
      case 'error':
        return '0 0 8px rgba(229, 62, 62, 0.4)';
      default: // primary
        return '0 0 8px rgba(255, 107, 53, 0.4)';
    }
  }};
`;

const ProgressLabel = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-medium);
`;

const ProgressText = styled.span`
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const ProgressPercentage = styled.span`
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  min-width: 3ch;
  text-align: right;
`;

export const Progress: React.FC<ProgressProps> = ({
  value,
  max = 100,
  showLabel = false,
  showPercentage = true,
  color = 'primary',
  size = 'md',
  className,
  label,
}) => {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

  return (
    <ProgressContainer size={size} className={className}>
      {(showLabel || label) && (
        <ProgressLabel>
          <ProgressText>{label || ''}</ProgressText>
          {showPercentage && (
            <ProgressPercentage>{Math.round(percentage)}%</ProgressPercentage>
          )}
        </ProgressLabel>
      )}

      <ProgressBar size={size} color={color}>
        <ProgressFill
          color={color}
          initial={{ width: 0 }}
          animate={{ width: `${percentage}%` }}
          transition={{
            duration: 0.3,
            ease: "easeOut"
          }}
        />
      </ProgressBar>
    </ProgressContainer>
  );
};

export default Progress;