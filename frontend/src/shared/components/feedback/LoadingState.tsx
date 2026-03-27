/**
 * LoadingState component - Generic loading indicator
 * TODO: Add size variants
 * TODO: Add message customization
 */

import { Spinner } from '../ui';

interface LoadingStateProps {
  message?: string;
  size?: 'sm' | 'md' | 'lg';
}

export const LoadingState = ({ message = 'Loading...', size = 'md' }: LoadingStateProps) => {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      <Spinner size={size} />
      <p className="mt-4 text-sm text-gray-600">{message}</p>
    </div>
  );
};
