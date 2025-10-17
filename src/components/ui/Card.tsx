import React from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { CardProps } from '../../types/launcher';

const StyledCard = styled(motion.div)<{
  interactive: boolean;
}>`
  background-color: var(--color-bg-secondary);
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;

  /* Interactive variant */
  ${({ interactive }) =>
    interactive &&
    `
    cursor: pointer;

    &:hover {
      border-color: var(--color-border-secondary);
      box-shadow: var(--shadow-md);
      transform: translateY(-2px);
    }

    &:active {
      transform: translateY(0);
    }
  `}

  /* Gradient overlay for visual interest */
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
    opacity: 0;
    transition: opacity var(--transition-normal);
  }

  &:hover::before {
    opacity: 1;
  }
`;

const CardHeader = styled.div`
  margin-bottom: var(--spacing-md);

  &:last-child {
    margin-bottom: 0;
  }
`;

const CardTitle = styled.h3`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
  line-height: 1.3;
`;

const CardSubtitle = styled.p`
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
  margin: 0;
  line-height: 1.4;
`;

const CardContent = styled.div`
  color: var(--color-text-secondary);
  line-height: 1.6;

  &:last-child {
    margin-bottom: 0;
  }
`;

const CardFooter = styled.div`
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-md);
`;

export const Card: React.FC<CardProps> & {
  Header: typeof CardHeader;
  Title: typeof CardTitle;
  Subtitle: typeof CardSubtitle;
  Content: typeof CardContent;
  Footer: typeof CardFooter;
} = ({ children, title, subtitle, className, interactive = false, onClick }) => {
  const hasHeader = title || subtitle;
  const content = React.Children.toArray(children);

  // Separate footer content (last child if it's a fragment with footer data-attribute)
  const mainContent = content.filter((child, index) => {
    if (index === content.length - 1 && React.isValidElement(child)) {
      return child.props['data-footer'] !== true;
    }
    return true;
  });

  const footerContent = content.find((child, index) => {
    if (index === content.length - 1 && React.isValidElement(child)) {
      return child.props['data-footer'] === true;
    }
    return false;
  });

  return (
    <StyledCard
      className={className}
      interactive={interactive}
      onClick={interactive ? onClick : undefined}
      whileHover={interactive ? { scale: 1.02 } : {}}
      whileTap={interactive ? { scale: 0.98 } : {}}
      transition={{ type: "spring", stiffness: 300, damping: 20 }}
    >
      {hasHeader && (
        <CardHeader>
          {title && <CardTitle>{title}</CardTitle>}
          {subtitle && <CardSubtitle>{subtitle}</CardSubtitle>}
        </CardHeader>
      )}

      <CardContent>{mainContent}</CardContent>

      {footerContent && <CardFooter>{footerContent}</CardFooter>}
    </StyledCard>
  );
};

// Card sub-components for more explicit structure
Card.Header = CardHeader;
Card.Title = CardTitle;
Card.Subtitle = CardSubtitle;
Card.Content = CardContent;
Card.Footer = CardFooter;

export default Card;