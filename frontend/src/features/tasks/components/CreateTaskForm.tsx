/**
 * CreateTaskForm Component
 * Form for creating new tasks with validation
 */

import { useState } from 'react';
import type { CreateTaskRequest, ScopeType, TaskPriority } from '../types';
import { TaskPriorityLabels, ScopeTypeLabels } from '../types';
import { Button } from '../../../shared/components/ui/Button';
import { Modal } from '../../../shared/components/ui/Modal';
import { EncounterAutocomplete } from '../../../shared/components/ui/EncounterAutocomplete';

interface CreateTaskFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateTaskRequest) => Promise<void>;
  defaultScopeType?: ScopeType;
  defaultScopeId?: string;
}

export const CreateTaskForm = ({
  isOpen,
  onClose,
  onSubmit,
  defaultScopeType,
  defaultScopeId,
}: CreateTaskFormProps) => {
  const [formData, setFormData] = useState<CreateTaskRequest>({
    scopeType: defaultScopeType || 'encounter',
    scopeId: defaultScopeId || '',
    title: '',
    details: '',
    priority: 'medium',
    slaDueAt: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (field: keyof CreateTaskRequest, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error for this field
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.scopeType) {
      newErrors.scopeType = 'Scope type is required';
    }
    if (!formData.scopeId.trim()) {
      newErrors.scopeId = 'Scope ID is required';
    }
    if (!formData.title.trim()) {
      newErrors.title = 'Title is required';
    }
    if (formData.title.length > 200) {
      newErrors.title = 'Title must be less than 200 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    const normalizeDatetime = (val?: string) => {
      if (!val) return undefined;
      // datetime-local gives "YYYY-MM-DDTHH:MM" — append seconds+Z for RFC3339
      return /T\d{2}:\d{2}$/.test(val) ? `${val}:00Z` : val;
    };

    setIsSubmitting(true);
    try {
      await onSubmit({
        ...formData,
        title: formData.title.trim(),
        details: formData.details?.trim() || undefined,
        slaDueAt: normalizeDatetime(formData.slaDueAt),
      });
      handleClose();
    } catch (error: any) {
      setErrors({
        submit: error.response?.data?.error?.message || 'Failed to create task',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setFormData({
        scopeType: defaultScopeType || 'encounter',
        scopeId: defaultScopeId || '',
        title: '',
        details: '',
        priority: 'medium',
        slaDueAt: '',
      });
      setErrors({});
      onClose();
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Create New Task">
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Scope Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Scope <span className="text-red-500">*</span>
          </label>
          <select
            value={formData.scopeType}
            onChange={(e) => handleChange('scopeType', e.target.value as ScopeType)}
            disabled={!!defaultScopeType}
            className={`w-full px-3 py-2 border rounded-lg ${
              errors.scopeType ? 'border-red-500' : 'border-gray-300'
            }`}
          >
            <option value="encounter">{ScopeTypeLabels.encounter}</option>
            <option value="patient">{ScopeTypeLabels.patient}</option>
            <option value="unit">{ScopeTypeLabels.unit}</option>
          </select>
          {errors.scopeType && (
            <p className="mt-1 text-sm text-red-600">{errors.scopeType}</p>
          )}
        </div>

        {/* Scope ID */}
        <div>
          {formData.scopeType === 'encounter' && !defaultScopeId ? (
            <EncounterAutocomplete
              label={`${ScopeTypeLabels[formData.scopeType]} ID`}
              value={formData.scopeId}
              onChange={(id) => handleChange('scopeId', id)}
              required
              error={errors.scopeId}
            />
          ) : (
            <>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {ScopeTypeLabels[formData.scopeType]} ID <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.scopeId}
                onChange={(e) => handleChange('scopeId', e.target.value)}
                disabled={!!defaultScopeId}
                placeholder={`Enter ${ScopeTypeLabels[formData.scopeType]} ID`}
                className={`w-full px-3 py-2 border rounded-lg ${
                  errors.scopeId ? 'border-red-500' : 'border-gray-300'
                }`}
              />
              {errors.scopeId && <p className="mt-1 text-sm text-red-600">{errors.scopeId}</p>}
            </>
          )}
        </div>

        {/* Title */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Title <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={formData.title}
            onChange={(e) => handleChange('title', e.target.value)}
            placeholder="Enter task title"
            maxLength={200}
            className={`w-full px-3 py-2 border rounded-lg ${
              errors.title ? 'border-red-500' : 'border-gray-300'
            }`}
          />
          {errors.title && <p className="mt-1 text-sm text-red-600">{errors.title}</p>}
          <p className="mt-1 text-xs text-gray-500">{formData.title.length}/200 characters</p>
        </div>

        {/* Details */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Details</label>
          <textarea
            value={formData.details}
            onChange={(e) => handleChange('details', e.target.value)}
            rows={3}
            placeholder="Optional task details"
            className="w-full px-3 py-2 border border-gray-300 rounded-lg"
          />
        </div>

        {/* Priority */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
          <select
            value={formData.priority}
            onChange={(e) => handleChange('priority', e.target.value as TaskPriority)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg"
          >
            <option value="low">{TaskPriorityLabels.low}</option>
            <option value="medium">{TaskPriorityLabels.medium}</option>
            <option value="high">{TaskPriorityLabels.high}</option>
            <option value="urgent">{TaskPriorityLabels.urgent}</option>
          </select>
        </div>

        {/* SLA Due Date */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            SLA Due Date (Optional)
          </label>
          <input
            type="datetime-local"
            value={formData.slaDueAt}
            onChange={(e) => handleChange('slaDueAt', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg"
          />
          <p className="mt-1 text-xs text-gray-500">
            Leave empty if no SLA deadline is required
          </p>
        </div>

        {/* Submit Error */}
        {errors.submit && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {errors.submit}
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-3 pt-2">
          <Button
            type="button"
            variant="secondary"
            onClick={handleClose}
            disabled={isSubmitting}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting} isLoading={isSubmitting} className="flex-1">
            Create Task
          </Button>
        </div>
      </form>
    </Modal>
  );
};
