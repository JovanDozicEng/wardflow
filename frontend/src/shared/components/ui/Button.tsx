/**
 * Button component - Base UI primitive
 * TODO: Implement button with variants (primary, secondary, danger, ghost)
 * TODO: Add size prop (sm, md, lg)
 * TODO: Add loading state with spinner
 * TODO: Add disabled state styling
 */

import type { ButtonHTMLAttributes, ReactNode } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}

export const Button = ({
  children,
  variant = 'primary',
  size = 'md',
  isLoading = false,
  className = '',
  disabled,
  ...props
}: ButtonProps) => {
  // TODO: Implement Tailwind classes for variants and sizes
  // TODO: Add loading spinner when isLoading is true
  // TODO: Disable button when isLoading or disabled
  
  return (
    <button
      className={`inline-flex items-center justify-center font-medium rounded-lg transition-colors ${className}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {children}
    </button>
  );
};
