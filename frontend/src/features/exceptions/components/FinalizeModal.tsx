/**
 * FinalizeModal - Confirm finalization of exception
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import type { ExceptionEvent } from '../types/exception.types';

interface FinalizeModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: () => Promise<void>;
  exception: ExceptionEvent | null;
  loading?: boolean;
}

export const FinalizeModal = ({
  isOpen,
  onClose,
  onSubmit,
  exception,
  loading = false,
}: FinalizeModalProps) => {
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      await onSubmit();
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to finalize exception');
    }
  };

  const handleClose = () => {
    setError(null);
    onClose();
  };

  if (!exception) return null;

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Finalize Exception">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-yellow-50 border border-yellow-200 rounded p-4 space-y-2">
          <p className="font-medium text-yellow-900 flex items-center gap-2">
            <span className="text-xl">⚠️</span>
            Warning: Finalizing this exception
          </p>
          <ul className="text-sm text-yellow-800 list-disc list-inside space-y-1 ml-6">
            <li>This exception will be <strong>immutable</strong></li>
            <li>You will <strong>not be able to edit</strong> the data</li>
            <li>Only corrections (new events) will be allowed</li>
            <li>This action <strong>cannot be undone</strong></li>
          </ul>
        </div>

        <div className="bg-gray-50 border border-gray-200 rounded p-3 text-sm">
          <p className="font-medium text-gray-900 mb-2">Exception Details:</p>
          <div className="space-y-1 text-gray-700">
            <p>
              <span className="font-medium">Type:</span> {exception.type}
            </p>
            <p>
              <span className="font-medium">Encounter:</span> {exception.encounterId}
            </p>
            <p>
              <span className="font-medium">Status:</span> {exception.status}
            </p>
          </div>
        </div>

        <div className="flex gap-3 pt-2">
          <Button
            type="button"
            variant="secondary"
            onClick={handleClose}
            disabled={loading}
            className="flex-1 px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={loading}
            isLoading={loading}
            className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 text-white"
          >
            {loading ? 'Finalizing...' : 'Finalize Exception'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
