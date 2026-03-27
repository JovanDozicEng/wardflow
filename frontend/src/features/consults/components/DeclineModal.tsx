/**
 * DeclineModal - Modal to decline consult with reason
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import type { DeclineConsultRequest } from '../types/consult.types';

interface DeclineModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: DeclineConsultRequest) => Promise<void>;
  loading?: boolean;
}

export const DeclineModal = ({ isOpen, onClose, onSubmit, loading = false }: DeclineModalProps) => {
  const [reason, setReason] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!reason.trim()) {
      setError('Reason is required to decline a consult');
      return;
    }

    try {
      await onSubmit({ reason });
      setReason('');
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to decline consult');
    }
  };

  const handleClose = () => {
    setError(null);
    setReason('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Decline Consult">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-yellow-50 border border-yellow-200 rounded p-3 text-sm text-yellow-800">
          <p className="font-medium mb-1">⚠️ Declining this consult</p>
          <p>Please provide a clear reason for declining this consult request.</p>
        </div>

        <div>
          <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-1">
            Reason for Declining *
          </label>
          <textarea
            id="reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Explain why this consult is being declined"
            rows={4}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            autoFocus
          />
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
            variant="danger"
            disabled={loading}
            isLoading={loading}
            className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 text-white"
          >
            {loading ? 'Declining...' : 'Decline Consult'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
