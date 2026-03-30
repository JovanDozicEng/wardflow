/**
 * HandoffForm - Transfer a care team role with structured handoff note
 */

import { useState } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import { Input } from '@/shared/components/ui/Input';
import type { CareTeamAssignment, TransferRoleRequest } from '../types/careTeam.types';

const ROLE_LABELS: Record<string, string> = {
  primary_nurse: 'Primary Nurse',
  attending_provider: 'Attending Provider',
  consulting_provider: 'Consulting Provider',
  resident: 'Resident',
  respiratory_therapist: 'Respiratory Therapist',
  case_manager: 'Case Manager',
  social_worker: 'Social Worker',
  other: 'Other',
};

interface HandoffFormProps {
  assignment: CareTeamAssignment | null;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (assignmentId: string, data: TransferRoleRequest) => Promise<void>;
  loading?: boolean;
}

export const HandoffForm = ({
  assignment,
  isOpen,
  onClose,
  onSubmit,
  loading = false,
}: HandoffFormProps) => {
  const [toUserId, setToUserId] = useState('');
  const [handoffNote, setHandoffNote] = useState('');
  const [pendingTasks, setPendingTasks] = useState('');
  const [patientStatus, setPatientStatus] = useState('');
  const [priorities, setPriorities] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!assignment) return;
    if (!toUserId.trim()) {
      setError('Incoming user ID is required');
      return;
    }
    if (!handoffNote.trim()) {
      setError('Handoff note is required');
      return;
    }

    const structuredFields: Record<string, any> = {};
    if (patientStatus.trim()) structuredFields.patientStatus = patientStatus.trim();
    if (pendingTasks.trim()) structuredFields.pendingTasks = pendingTasks.trim();
    if (priorities.trim()) structuredFields.priorities = priorities.trim();

    try {
      await onSubmit(assignment.id, {
        toUserId: toUserId.trim(),
        handoffNote: handoffNote.trim(),
        structuredFields: Object.keys(structuredFields).length > 0 ? structuredFields : undefined,
      });
      setToUserId('');
      setHandoffNote('');
      setPendingTasks('');
      setPatientStatus('');
      setPriorities('');
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to transfer role');
    }
  };

  const roleLabel = assignment ? (ROLE_LABELS[assignment.roleType] ?? assignment.roleType) : '';

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`Handoff — ${roleLabel}`}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div className="bg-amber-50 border border-amber-200 rounded p-3 text-sm text-amber-800">
          Transferring <strong>{roleLabel}</strong> role. The incoming clinician will receive a handoff note.
        </div>

        <Input
          label="Incoming User ID"
          value={toUserId}
          onChange={(e) => setToUserId(e.target.value)}
          placeholder="Enter user ID of incoming clinician"
          required
        />

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Patient Status <span className="text-gray-400 text-xs">(optional)</span>
          </label>
          <textarea
            value={patientStatus}
            onChange={(e) => setPatientStatus(e.target.value)}
            rows={2}
            placeholder="Current patient status and clinical summary…"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Pending Tasks <span className="text-gray-400 text-xs">(optional)</span>
          </label>
          <textarea
            value={pendingTasks}
            onChange={(e) => setPendingTasks(e.target.value)}
            rows={2}
            placeholder="Outstanding tasks, orders, follow-ups…"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Priorities <span className="text-gray-400 text-xs">(optional)</span>
          </label>
          <textarea
            value={priorities}
            onChange={(e) => setPriorities(e.target.value)}
            rows={2}
            placeholder="Key priorities for incoming shift…"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Handoff Note <span className="text-red-500">*</span>
          </label>
          <textarea
            value={handoffNote}
            onChange={(e) => setHandoffNote(e.target.value)}
            rows={3}
            placeholder="Summary handoff note for the record…"
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={loading}>
            {loading ? 'Transferring…' : 'Complete Handoff'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
