import React, { memo, useEffect, useRef, useCallback } from 'react';
import styled, { keyframes } from 'styled-components';
import { createPortal } from 'react-dom';

// Enhanced animations
const fadeIn = keyframes`
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
`;

const Backdrop = styled.div<{ open: boolean }>`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  opacity: ${props => props.open ? 1 : 0};
  visibility: ${props => props.open ? 'visible' : 'hidden'};
  transition: all var(--transition-normal) ease-in-out;
  padding: var(--spacing-lg);
`;

const ModalContainer = styled.div<{
  size: 'sm' | 'md' | 'lg' | 'xl';
  open: boolean;
}>`
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-lg);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  max-width: ${props => {
    switch (props.size) {
      case 'sm': return '400px';
      case 'lg': return '800px';
      case 'xl': return '1200px';
      default: return '600px';
    }
  }};
  max-height: 90vh;
  width: 100%;
  overflow: hidden;
  transform: ${props => props.open ? 'scale(1)' : 'scale(0.95)'};
  opacity: ${props => props.open ? 1 : 0};
  animation: ${props => props.open ? fadeIn : 'none'} var(--transition-normal) ease-in-out;
  will-change: transform, opacity;
  contain: layout style paint;
`;

const ModalHeader = styled.header`
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, var(--color-bg-secondary) 0%, var(--color-bg-tertiary) 100%);
`;

const ModalTitle = styled.h2`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
  line-height: 1.3;
`;

const ModalCloseButton = styled.button`
  background: none;
  border: none;
  padding: var(--spacing-sm);
  border-radius: var(--radius-full);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast) ease-in-out;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;

  &:hover {
    background-color: var(--color-bg-tertiary);
    color: var(--color-text-primary);
  }

  &:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }
`;

const ModalContent = styled.main`
  padding: var(--spacing-lg);
  overflow-y: auto;
  max-height: calc(90vh - 140px);

  /* Custom scrollbar */
  scrollbar-width: thin;
  scrollbar-color: var(--color-bg-quaternary) var(--color-bg-secondary);

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: var(--color-bg-secondary);
  }

  &::-webkit-scrollbar-thumb {
    background: var(--color-bg-quaternary);
    border-radius: var(--radius-full);
  }
`;

const ModalFooter = styled.footer`
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border-primary);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--spacing-md);
  background: var(--color-bg-tertiary);
`;

interface EnhancedModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showCloseButton?: boolean;
  closeOnBackdropClick?: boolean;
  closeOnEscape?: boolean;
  footer?: React.ReactNode;
  'aria-label'?: string;
}

export const EnhancedModal = memo<EnhancedModalProps>(({
  isOpen,
  onClose,
  title,
  children,
  size = 'md',
  showCloseButton = true,
  closeOnBackdropClick = true,
  closeOnEscape = true,
  footer,
  'aria-label': ariaLabel,
}) => {
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  // Handle escape key
  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    if (event.key === 'Escape' && closeOnEscape) {
      event.preventDefault();
      onClose();
    }
  }, [closeOnEscape, onClose]);

  // Handle backdrop click
  const handleBackdropClick = useCallback((event: React.MouseEvent) => {
    if (event.target === event.currentTarget && closeOnBackdropClick) {
      onClose();
    }
  }, [closeOnBackdropClick, onClose]);

  // Focus management
  const trapFocus = useCallback((element: HTMLElement) => {
    const focusableElements = element.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    ) as NodeListOf<HTMLElement>;

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key === 'Tab') {
        if (e.shiftKey) {
          if (document.activeElement === firstElement) {
            lastElement.focus();
            e.preventDefault();
          }
        } else {
          if (document.activeElement === lastElement) {
            firstElement.focus();
            e.preventDefault();
          }
        }
      }
    };

    element.addEventListener('keydown', handleTabKey);
    firstElement?.focus();

    return () => {
      element.removeEventListener('keydown', handleTabKey);
    };
  }, []);

  // Handle modal open/close
  useEffect(() => {
    if (isOpen) {
      // Store previous focus
      previousFocusRef.current = document.activeElement as HTMLElement;

      // Prevent body scroll
      document.body.style.overflow = 'hidden';

      // Add escape key listener
      document.addEventListener('keydown', handleKeyDown);

      // Focus trap after a short delay to ensure DOM is ready
      const timer = setTimeout(() => {
        if (modalRef.current) {
          trapFocus(modalRef.current);
        }
      }, 100);

      return () => {
        clearTimeout(timer);
        document.body.style.overflow = '';
        document.removeEventListener('keydown', handleKeyDown);

        // Restore previous focus
        if (previousFocusRef.current) {
          previousFocusRef.current.focus();
        }
      };
    }
  }, [isOpen, handleKeyDown, trapFocus]);

  // Prevent click-through when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.pointerEvents = 'none';
      return () => {
        document.body.style.pointerEvents = '';
      };
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const modal = (
    <Backdrop open={isOpen} onClick={handleBackdropClick}>
      <ModalContainer
        ref={modalRef}
        size={size}
        open={isOpen}
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel || title}
        onClick={(e) => e.stopPropagation()}
      >
        <ModalHeader>
          <ModalTitle>{title}</ModalTitle>
          {showCloseButton && (
            <ModalCloseButton
              onClick={onClose}
              aria-label="Close modal"
              type="button"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </ModalCloseButton>
          )}
        </ModalHeader>

        <ModalContent>
          {children}
        </ModalContent>

        {footer && (
          <ModalFooter>
            {footer}
          </ModalFooter>
        )}
      </ModalContainer>
    </Backdrop>
  );

  return createPortal(modal, document.body);
});

EnhancedModal.displayName = 'EnhancedModal';