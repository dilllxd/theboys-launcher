import React from 'react';
import styled from 'styled-components';
import { LoadingSpinner } from './ui';

const LoadingContainer = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, var(--color-bg-primary) 0%, var(--color-bg-secondary) 100%);
  gap: var(--spacing-lg);
`;

const LoadingMessage = styled.p`
  color: var(--color-text-secondary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-medium);
  text-align: center;
  max-width: 300px;
`;

const Logo = styled.div`
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  margin-bottom: var(--spacing-md);
  text-align: center;

  /* TheBoys logo styling */
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 20px rgba(255, 107, 53, 0.3);
`;

export const LoadingScreen: React.FC<{
  message?: string;
}> = ({ message = "Loading..." }) => {
  return (
    <LoadingContainer>
      <Logo>TheBoys</Logo>
      <LoadingSpinner size="lg" />
      <LoadingMessage>{message}</LoadingMessage>
    </LoadingContainer>
  );
};

export default LoadingScreen;