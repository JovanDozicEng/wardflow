/**
 * CorrectionModal - Create correction for finalized exception
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import type { ExceptionEvent, CorrectExceptionRequest } from '../types/exception.types';

interface CorrectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CorrectExceptionRequest) => Promise<void>;
  exception: ExceptionEvent | null;
  loading?: boolean;
}

export const CorrectionModal = ({
  isOpen,
  onClose,
  onSubmit,
  exception,
  loading = false,
}: CorrectionModalProps) => {
  const [reason, setReason] = useState('');
  const [dataJson, setDataJson] = useState(
    exception?.data ? JSON.stringify(exception.data, null, 2) : '{}'
  );
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!reason.trim()) {
      setError('Correction reason is required');
      return;
    }

    // Validate JSON
    let parsedData: Record<string, any>;
    try {
      parsedData = JSON.parse(dataJson);
    } catch (err) {
      setError('Invalid JSON format');
      return;
    }

    try {
      await onSubmit({ reason, data: parsedData });
      setReason('');
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to create correction');
    }
  };

  const handleClose = () => {
    setError(null);
    setReason('');
    onClose();
  };

  if (!exception) return null;

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Correct Exception">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
          <p className="font-medium mb-1">ℹ️ Creating a correction</p>
          <p>This will create a new exception event that corrects the data in the original finalized exception.</p>
        </div>

        <div className="bg-gray-50 border border-gray-200 rounded p-3 text-sm">
          <p className="font-medium text-gray-900 mb-2">Original Exception:</p>
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

        <div>
          <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-1">
            Reason for Correction *
          </label>
          <textarea
            id="reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Explain why this correction is needed"
            rows={3}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
            autoFocus
          />
        </div>

        <div>
          <label htmlFor="data" className="block text-sm font-medium text-gray-700 mb-1">
            Corrected Data (JSON) *
          </label>
          <textarea
            id="data"
            value={dataJson}
            onChange={(e) => setDataJson(e.target.value)}
            placeholder='{"key": "value"}'
            rows={8}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed font-mono text-sm"
          />
          <p className="text-xs text-gray-500 mt-1">Update the data with correct values</p>
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
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
          >
            {loading ? 'Creating...' : 'Create Correction'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
