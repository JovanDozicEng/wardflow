/**
 * AssignTaskModal Component
 * Modal for assigning or reassigning tasks to users
 */

import { useState } from 'react';
import { User, UserX } from 'lucide-react';
import type { AssignTaskRequest } from '../types';
import { Button } from '../../../shared/components/ui/Button';
import { Modal } from '../../../shared/components/ui/Modal';

interface AssignTaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: AssignTaskRequest) => Promise<void>;
  taskTitle: string;
  currentOwnerId: string | null;
  currentOwnerName?: string | null;
}

export const AssignTaskModal = ({
  isOpen,
  onClose,
  onSubmit,
  taskTitle,
  currentOwnerId,
  currentOwnerName,
}: AssignTaskModalProps) => {
  const [toOwnerId, setToOwnerId] = useState('');
  const [reason, setReason] = useState('');
  const [isUnassigning, setIsUnassigning] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!isUnassigning && !toOwnerId.trim()) {
      setError('Please enter a user ID or select unassign');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      await onSubmit({
        toOwnerId: isUnassigning ? null : toOwnerId.trim(),
        reason: reason.trim() || undefined,
      });
      handleClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to assign task');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setToOwnerId('');
      setReason('');
      setIsUnassigning(false);
      setError(null);
      onClose();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={isUnassigning ? 'Unassign Task' : 'Assign Task'}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Task Info */}
        <div className="p-3 bg-gray-50 rounded-lg">
          <p className="text-sm text-gray-600">Task</p>
          <p className="font-medium mt-1">{taskTitle}</p>
          {currentOwnerName && (
            <p className="text-sm text-gray-600 mt-2">
              Current Owner: <span className="font-medium">{currentOwnerName}</span>
            </p>
          )}
        </div>

        {/* Action Toggle */}
        <div className="flex gap-2">
          <Button
            type="button"
            variant={!isUnassigning ? 'primary' : 'secondary'}
            onClick={() => setIsUnassigning(false)}
            className="flex-1 flex items-center justify-center gap-2"
          >
            <User className="w-4 h-4" />
            Assign
          </Button>
          <Button
            type="button"
            variant={isUnassigning ? 'primary' : 'secondary'}
            onClick={() => setIsUnassigning(true)}
            className="flex-1 flex items-center justify-center gap-2"
          >
            <UserX className="w-4 h-4" />
            Unassign
          </Button>
        </div>

        {/* User ID Input (only when assigning) */}
        {!isUnassigning && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              User ID <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={toOwnerId}
              onChange={(e) => setToOwnerId(e.target.value)}
              placeholder="Enter user ID"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            <p className="mt-1 text-xs text-gray-500">
              Enter the UUID of the user to assign this task to
            </p>
          </div>
        )}

        {/* Reason */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Reason (Optional)
          </label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            rows={3}
            placeholder={
              isUnassigning
                ? 'Optional reason for unassigning'
                : 'Optional reason for assignment'
            }
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        {/* Error Display */}
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {error}
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
            {isUnassigning ? 'Unassign Task' : 'Assign Task'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
