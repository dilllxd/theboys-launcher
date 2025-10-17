import React from 'react';
import styled from 'styled-components';

const CheckboxContainer = styled.label`
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
`;

const HiddenCheckbox = styled.input.attrs({ type: 'checkbox' })`
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
`;

const StyledCheckbox = styled.div<{ checked: boolean; disabled: boolean }>`
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid var(--color-border-primary);
  border-radius: var(--radius-sm);
  background-color: ${({ checked }) =>
    checked ? 'var(--color-primary)' : 'var(--color-bg-tertiary)'
  };
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  cursor: ${({ disabled }) => (disabled ? 'not-allowed' : 'pointer')};
  opacity: ${({ disabled }) => (disabled ? '0.6' : '1')};

  ${HiddenCheckbox}:focus + & {
    border-color: var(--color-border-focus);
    box-shadow: 0 0 0 3px rgba(255, 107, 53, 0.1);
  }

  &::after {
    content: '✓';
    color: var(--color-text-primary);
    font-size: 0.75rem;
    font-weight: bold;
    opacity: ${({ checked }) => (checked ? 1 : 0)};
    transform: ${({ checked }) => (checked ? 'scale(1)' : 'scale(0)')};
    transition: all var(--transition-fast);
  }
`;

interface CheckboxProps extends Omit<React.ComponentProps<'input'>, 'type'> {
  label?: string;
}

const Checkbox = React.forwardRef<HTMLInputElement, CheckboxProps>(
  ({ label, disabled, checked, onChange, ...props }, ref) => {
    return (
      <CheckboxContainer>
        <HiddenCheckbox
          ref={ref}
          disabled={disabled}
          checked={checked}
          onChange={onChange}
          {...props}
        />
        <StyledCheckbox checked={!!checked} disabled={!!disabled} />
        {label && <span>{label}</span>}
      </CheckboxContainer>
    );
  }
);

Checkbox.displayName = 'Checkbox';

export { Checkbox };
export default Checkbox;