/**
 * RedirectModal - Modal to redirect consult to another service
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Input } from '@/shared/components/ui/Input';
import { Button } from '@/shared/components/ui/Button';
import type { RedirectConsultRequest } from '../types/consult.types';

interface RedirectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: RedirectConsultRequest) => Promise<void>;
  loading?: boolean;
}

export const RedirectModal = ({ isOpen, onClose, onSubmit, loading = false }: RedirectModalProps) => {
  const [formData, setFormData] = useState<RedirectConsultRequest>({
    targetService: '',
    reason: '',
  });
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!formData.targetService.trim()) {
      setError('Target service is required');
      return;
    }
    if (!formData.reason.trim()) {
      setError('Reason is required');
      return;
    }

    try {
      await onSubmit(formData);
      setFormData({ targetService: '', reason: '' });
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to redirect consult');
    }
  };

  const handleClose = () => {
    setError(null);
    setFormData({ targetService: '', reason: '' });
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Redirect Consult">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
          <p className="font-medium mb-1">ℹ️ Redirecting consult</p>
          <p>This consult will be sent to a different service for review.</p>
        </div>

        <div>
          <label htmlFor="targetService" className="block text-sm font-medium text-gray-700 mb-1">
            New Target Service *
          </label>
          <Input
            id="targetService"
            value={formData.targetService}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, targetService: e.target.value })}
            placeholder="e.g., Internal Medicine, Surgery"
            disabled={loading}
            autoFocus
          />
        </div>

        <div>
          <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-1">
            Reason for Redirect *
          </label>
          <textarea
            id="reason"
            value={formData.reason}
            onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
            placeholder="Explain why this consult is being redirected"
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
            className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white"
          >
            {loading ? 'Redirecting...' : 'Redirect Consult'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
