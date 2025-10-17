import React, { memo, useEffect, useRef, useState } from 'react';
import styled, { keyframes } from 'styled-components';
import { OptimizedButton } from './OptimizedButton';

// Enhanced animations
const slideIn = keyframes`
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
`;

const slideOut = keyframes`
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(100%);
    opacity: 0;
  }
`;

const bounceIn = keyframes`
  0% {
    transform: translateX(100%) scale(0.3);
    opacity: 0;
  }
  50% {
    transform: translateX(-10%) scale(1.05);
  }
  70% {
    transform: translateX(2%) scale(0.98);
  }
  100% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
`;

const typeConfig = {
  success: {
    background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    icon: '✓',
    duration: 5000,
  },
  error: {
    background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
    icon: '✕',
    duration: 8000,
  },
  warning: {
    background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
    icon: '⚠',
    duration: 6000,
  },
  info: {
    background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
    icon: 'ℹ',
    duration: 4000,
  },
};

type NotificationType = keyof typeof typeConfig;

interface NotificationProps {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  duration?: number;
  persistent?: boolean;
  action?: {
    label: string;
    onClick: () => void;
  };
  onClose: (id: string) => void;
}

const NotificationContainer = styled.div<{
  type: NotificationType;
  isExiting: boolean;
}>`
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  border-radius: var(--radius-lg);
  color: white;
  min-width: 320px;
  max-width: 480px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
  position: relative;
  overflow: hidden;

  /* Dynamic background based on type */
  background: ${props => typeConfig[props.type].background};

  /* Animations */
  animation: ${props => props.isExiting ? slideOut : bounceIn} 0.3s ease-in-out;
  will-change: transform, opacity;

  /* Hover effect */
  &:hover {
    transform: translateX(-5px);
    box-shadow: 0 15px 35px rgba(0, 0, 0, 0.3);
  }

  /* Progress bar indicator */
  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    height: 3px;
    background: rgba(255, 255, 255, 0.7);
    border-radius: 0 0 0 var(--radius-lg);
    animation: progress-slide ${props => typeConfig[props.type].duration}ms linear;
  }

  @keyframes progress-slide {
    from {
      width: 100%;
    }
    to {
      width: 0%;
    }
  }
`;

const NotificationIcon = styled.div`
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  font-weight: bold;
  font-size: 14px;
`;

const NotificationContent = styled.div`
  flex: 1;
  min-width: 0;
`;

const NotificationTitle = styled.h4`
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  margin: 0 0 var(--spacing-xs) 0;
  line-height: 1.3;
`;

const NotificationMessage = styled.p`
  font-size: var(--font-size-xs);
  margin: 0;
  line-height: 1.4;
  opacity: 0.9;
`;

const NotificationActions = styled.div`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-sm);
`;

const NotificationClose = styled.button`
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  cursor: pointer;
  transition: all var(--transition-fast) ease-in-out;
  font-size: 12px;

  &:hover {
    background: rgba(255, 255, 255, 0.3);
    transform: scale(1.1);
  }

  &:focus-visible {
    outline: 2px solid white;
    outline-offset: 2px;
  }
`;

const ActionButton = styled(OptimizedButton)`
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  font-size: var(--font-size-xs);
  padding: var(--spacing-xs) var(--spacing-sm);

  &:hover {
    background: rgba(255, 255, 255, 0.3);
  }
`;

export const EnhancedNotification: React.FC<NotificationProps> = memo(({
  id,
  type,
  title,
  message,
  duration,
  persistent = false,
  action,
  onClose,
}) => {
  const [isExiting, setIsExiting] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const timerRef = useRef<NodeJS.Timeout>();
  const progressTimerRef = useRef<NodeJS.Timeout>();

  const handleMouseEnter = () => {
    setIsPaused(true);
  };

  const handleMouseLeave = () => {
    setIsPaused(false);
  };

  const handleClose = () => {
    if (!isExiting) {
      setIsExiting(true);
      setTimeout(() => onClose(id), 300);
    }
  };

  const handleAction = () => {
    action?.onClick();
    handleClose();
  };

  // Auto-dismiss timer
  useEffect(() => {
    if (!persistent && !isPaused && !isExiting) {
      const notificationDuration = duration || typeConfig[type].duration;

      timerRef.current = setTimeout(() => {
        handleClose();
      }, notificationDuration);
    }

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, [persistent, isPaused, isExiting, type, duration, onClose]);

  // Keyboard accessibility
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        handleClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  return (
    <NotificationContainer
      type={type}
      isExiting={isExiting}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      role="alert"
      aria-live="polite"
      aria-atomic="true"
    >
      <NotificationIcon>
        {typeConfig[type].icon}
      </NotificationIcon>

      <NotificationContent>
        <NotificationTitle>{title}</NotificationTitle>
        {message && <NotificationMessage>{message}</NotificationMessage>}

        {action && (
          <NotificationActions>
            <ActionButton
              variant="outline"
              size="sm"
              onClick={handleAction}
            >
              {action.label}
            </ActionButton>
          </NotificationActions>
        )}
      </NotificationContent>

      <NotificationClose
        onClick={handleClose}
        aria-label="Close notification"
      >
        ×
      </NotificationClose>
    </NotificationContainer>
  );
});

EnhancedNotification.displayName = 'EnhancedNotification';

// Notification Container Component
const NotificationsWrapper = styled.div`
  position: fixed;
  top: var(--spacing-lg);
  right: var(--spacing-lg);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  pointer-events: none;

  > * {
    pointer-events: auto;
  }
`;

interface NotificationData {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  duration?: number;
  persistent?: boolean;
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface EnhancedNotificationContainerProps {
  notifications: NotificationData[];
  onClose: (id: string) => void;
}

export const EnhancedNotificationContainer: React.FC<EnhancedNotificationContainerProps> = memo(({
  notifications,
  onClose,
}) => {
  return (
    <NotificationsWrapper role="region" aria-label="Notifications">
      {notifications.map((notification) => (
        <EnhancedNotification
          key={notification.id}
          {...notification}
          onClose={onClose}
        />
      ))}
    </NotificationsWrapper>
  );
});

EnhancedNotificationContainer.displayName = 'EnhancedNotificationContainer';