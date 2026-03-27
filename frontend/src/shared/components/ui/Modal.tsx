/**
 * Modal component - Dialog overlay
 * TODO: Implement with portal
 * TODO: Add backdrop click to close
 * TODO: Add ESC key to close
 * TODO: Add animation (fade in/out)
 * TODO: Consider using headlessui Dialog for better accessibility
 */

import type { ReactNode } from 'react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

export const Modal = ({
  isOpen,
  onClose,
  title,
  children,
}: ModalProps) => {
  if (!isOpen) return null;
  
  // TODO: Use React Portal to render at document body level
  // TODO: Add backdrop overlay
  // TODO: Add close button
  // TODO: Implement size variants
  // TODO: Add focus trap for accessibility
  
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black bg-opacity-50"
        onClick={onClose}
      />
      
      {/* Modal content */}
      <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {title && (
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-semibold text-gray-900">{title}</h2>
          </div>
        )}
        <div className="px-6 py-4">
          {children}
        </div>
      </div>
    </div>
  );
};
