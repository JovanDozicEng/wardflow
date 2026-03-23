/**
 * Card component - Container with shadow and border
 * TODO: Implement with padding variants
 * TODO: Add hover state option
 * TODO: Add clickable variant
 */

import type { HTMLAttributes, ReactNode } from 'react';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
}

export const Card = ({
  children,
  padding = 'md',
  hover = false,
  className = '',
  ...props
}: CardProps) => {
  // TODO: Implement padding variants
  // TODO: Add hover shadow effect when hover=true
  
  return (
    <div
      className={`bg-white rounded-lg shadow border border-gray-200 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
};
