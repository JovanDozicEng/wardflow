/**
 * StatusUpdateModal - Update incident status
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Select } from '@/shared/components/ui/Select';
import { Button } from '@/shared/components/ui/Button';
import type { Incident, IncidentStatus, UpdateIncidentStatusRequest } from '../types/incident.types';

interface StatusUpdateModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: UpdateIncidentStatusRequest) => Promise<void>;
  incident: Incident | null;
  loading?: boolean;
}

export const StatusUpdateModal = ({
  isOpen,
  onClose,
  onSubmit,
  incident,
  loading = false,
}: StatusUpdateModalProps) => {
  const [status, setStatus] = useState<IncidentStatus>('under_review');
  const [note, setNote] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      await onSubmit({
        status,
        note: note.trim() || undefined,
      });
      setNote('');
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to update status');
    }
  };

  const handleClose = () => {
    setError(null);
    setNote('');
    onClose();
  };

  if (!incident) return null;

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Update Incident Status">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-gray-50 border border-gray-200 rounded p-3 text-sm">
          <p className="font-medium text-gray-900 mb-2">Incident Details:</p>
          <div className="space-y-1 text-gray-700">
            <p>
              <span className="font-medium">Type:</span> {incident.type}
            </p>
            <p>
              <span className="font-medium">Current Status:</span>{' '}
              <span className="font-semibold">{incident.status.replace('_', ' ').toUpperCase()}</span>
            </p>
          </div>
        </div>

        <div>
          <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
            New Status *
          </label>
          <Select
            id="status"
            value={status}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setStatus(e.target.value as IncidentStatus)}
            disabled={loading}
            options={[
              { value: 'submitted', label: 'Submitted' },
              { value: 'under_review', label: 'Under Review' },
              { value: 'closed', label: 'Closed' },
            ]}
          />
        </div>

        <div>
          <label htmlFor="note" className="block text-sm font-medium text-gray-700 mb-1">
            Note (Optional)
          </label>
          <textarea
            id="note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Add a note about this status change"
            rows={4}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
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
            variant="primary"
            disabled={loading}
            isLoading={loading}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
          >
            {loading ? 'Updating...' : 'Update Status'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
