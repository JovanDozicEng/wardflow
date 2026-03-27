/**
 * Spinner component - Loading indicator
 * TODO: Implement with animation
 * TODO: Add size variants
 * TODO: Add color variants
 */

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const Spinner = ({ className = '' }: SpinnerProps) => {
  // TODO: Implement animated spinner with Tailwind
  // TODO: Add size variants
  
  return (
    <div className={`animate-spin rounded-full border-2 border-gray-300 border-t-blue-600 ${className}`} />
  );
};
