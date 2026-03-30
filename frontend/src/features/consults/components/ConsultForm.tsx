/**
 * ConsultForm - Modal form to create new consult
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Input } from '@/shared/components/ui/Input';
import { Button } from '@/shared/components/ui/Button';
import { Select } from '@/shared/components/ui/Select';
import { EncounterAutocomplete } from '@/shared/components/ui/EncounterAutocomplete';
import type { CreateConsultRequest, ConsultUrgency } from '../types/consult.types';

interface ConsultFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateConsultRequest) => Promise<void>;
  loading?: boolean;
}

export const ConsultForm = ({ isOpen, onClose, onSubmit, loading = false }: ConsultFormProps) => {
  const [formData, setFormData] = useState<CreateConsultRequest>({
    encounterId: '',
    targetService: '',
    reason: '',
    urgency: 'routine',
  });
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Basic validation
    if (!formData.encounterId.trim()) {
      setError('Encounter ID is required');
      return;
    }
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
      // Reset form on success
      setFormData({
        encounterId: '',
        targetService: '',
        reason: '',
        urgency: 'routine',
      });
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to create consult');
    }
  };

  const handleClose = () => {
    setError(null);
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="New Consult Request">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <EncounterAutocomplete
          label="Encounter"
          value={formData.encounterId}
          onChange={(id) => setFormData({ ...formData, encounterId: id })}
          required
          disabled={loading}
        />

        <div>
          <label htmlFor="targetService" className="block text-sm font-medium text-gray-700 mb-1">
            Target Service *
          </label>
          <Input
            id="targetService"
            value={formData.targetService}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, targetService: e.target.value })}
            placeholder="e.g., Cardiology, Neurology"
            disabled={loading}
          />
        </div>

        <div>
          <label htmlFor="urgency" className="block text-sm font-medium text-gray-700 mb-1">
            Urgency *
          </label>
          <Select
            id="urgency"
            value={formData.urgency}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setFormData({ ...formData, urgency: e.target.value as ConsultUrgency })}
            disabled={loading}
            options={[
              { value: 'routine', label: 'Routine' },
              { value: 'urgent', label: 'Urgent' },
              { value: 'emergent', label: 'Emergent' },
            ]}
          />
        </div>

        <div>
          <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-1">
            Reason *
          </label>
          <textarea
            id="reason"
            value={formData.reason}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setFormData({ ...formData, reason: e.target.value })}
            placeholder="Describe the reason for this consult"
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
            {loading ? 'Creating...' : 'Create Consult'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
