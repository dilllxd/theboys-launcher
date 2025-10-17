import React from 'react';
import styled from 'styled-components';

const TooltipContainer = styled.div`
  position: relative;
  display: inline-block;
`;

const TooltipContent = styled.div`
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-primary);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  white-space: nowrap;
  z-index: 1000;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--transition-fast);
  border: 1px solid var(--color-border-primary);
  box-shadow: var(--shadow-md);
  margin-bottom: var(--spacing-xs);

  ${TooltipContainer}:hover & {
    opacity: 1;
  }
`;

interface TooltipProps {
  content: string;
  children: React.ReactNode;
}

export const Tooltip: React.FC<TooltipProps> = ({ content, children }) => {
  return (
    <TooltipContainer>
      {children}
      <TooltipContent>{content}</TooltipContent>
    </TooltipContainer>
  );
};

export default Tooltip;