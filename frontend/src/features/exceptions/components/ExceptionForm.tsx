/**
 * ExceptionForm - Create or edit exception
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Input } from '@/shared/components/ui/Input';
import { Button } from '@/shared/components/ui/Button';
import { EncounterAutocomplete } from '@/shared/components/ui/EncounterAutocomplete';
import type { CreateExceptionRequest, ExceptionEvent } from '../types/exception.types';

interface ExceptionFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateExceptionRequest | { data: Record<string, any> }) => Promise<void>;
  editMode?: boolean;
  exception?: ExceptionEvent;
  loading?: boolean;
}

export const ExceptionForm = ({
  isOpen,
  onClose,
  onSubmit,
  editMode = false,
  exception,
  loading = false,
}: ExceptionFormProps) => {
  const [encounterId, setEncounterId] = useState(exception?.encounterId || '');
  const [type, setType] = useState(exception?.type || '');
  const [dataJson, setDataJson] = useState(
    exception?.data ? JSON.stringify(exception.data, null, 2) : '{}'
  );
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate JSON
    let parsedData: Record<string, any>;
    try {
      parsedData = JSON.parse(dataJson);
    } catch (err) {
      setError('Invalid JSON format');
      return;
    }

    if (!editMode) {
      if (!encounterId.trim()) {
        setError('Encounter ID is required');
        return;
      }
      if (!type.trim()) {
        setError('Exception type is required');
        return;
      }
    }

    try {
      if (editMode) {
        await onSubmit({ data: parsedData });
      } else {
        await onSubmit({
          encounterId,
          type,
          data: parsedData,
        });
      }
      
      // Reset form on success
      if (!editMode) {
        setEncounterId('');
        setType('');
        setDataJson('{}');
      }
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to save exception');
    }
  };

  const handleClose = () => {
    setError(null);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={editMode ? 'Edit Exception' : 'New Exception'}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        {!editMode && (
          <>
            <EncounterAutocomplete
              label="Encounter"
              value={encounterId}
              onChange={setEncounterId}
              required
              disabled={loading}
            />

            <div>
              <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-1">
                Exception Type *
              </label>
              <Input
                id="type"
                value={type}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setType(e.target.value)}
                placeholder="e.g., medication_error, documentation_missing"
                disabled={loading}
              />
            </div>
          </>
        )}

        {editMode && exception && (
          <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
            <p>
              <span className="font-medium">Encounter:</span> {exception.encounterId}
            </p>
            <p>
              <span className="font-medium">Type:</span> {exception.type}
            </p>
          </div>
        )}

        <div>
          <label htmlFor="data" className="block text-sm font-medium text-gray-700 mb-1">
            Exception Data (JSON) *
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
          <p className="text-xs text-gray-500 mt-1">Enter valid JSON format</p>
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
            {loading ? 'Saving...' : editMode ? 'Update Exception' : 'Create Exception'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
