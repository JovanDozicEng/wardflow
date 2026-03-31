/**
 * HandoffForm - Transfer a care team role with structured handoff note
 */

import { useState, useEffect, useRef } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import type { CareTeamAssignment, TransferRoleRequest } from '../types/careTeam.types';
import api from '@/shared/utils/api';

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

interface UserSummary {
  id: string;
  name: string;
  email: string;
  role: string;
}

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
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<UserSummary[]>([]);
  const [selectedUser, setSelectedUser] = useState<UserSummary | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [searching, setSearching] = useState(false);
  const [handoffNote, setHandoffNote] = useState('');
  const [pendingTasks, setPendingTasks] = useState('');
  const [patientStatus, setPatientStatus] = useState('');
  const [priorities, setPriorities] = useState('');
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (isOpen) {
      setQuery(''); setResults([]); setSelectedUser(null);
      setShowDropdown(false); setHandoffNote('');
      setPendingTasks(''); setPatientStatus(''); setPriorities(''); setError(null);
    }
  }, [isOpen]);

  useEffect(() => {
    if (selectedUser) return;
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (!query.trim()) { setResults([]); setShowDropdown(false); return; }
    debounceRef.current = setTimeout(async () => {
      setSearching(true);
      try {
        const res = await api.get<UserSummary[]>('/users', { params: { q: query } });
        setResults(Array.isArray(res.data) ? res.data : []);
        setShowDropdown(true);
      } catch {
        setResults([]);
      } finally {
        setSearching(false);
      }
    }, 300);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
        debounceRef.current = null;
      }
    };
  }, [query, selectedUser]);

  const handleSelect = (user: UserSummary) => {
    setSelectedUser(user); setQuery(user.name); setShowDropdown(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!assignment) return;
    if (!selectedUser) {
      setError('Please search and select the incoming clinician');
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
        toUserId: selectedUser.id,
        handoffNote: handoffNote.trim(),
        structuredFields: Object.keys(structuredFields).length > 0 ? structuredFields : undefined,
      });
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || err.message || 'Failed to transfer role');
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

        {/* Incoming clinician search */}
        <div className="relative">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Incoming Clinician <span className="text-red-500">*</span>
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              value={query}
              onChange={(e) => { setQuery(e.target.value); if (selectedUser) setSelectedUser(null); }}
              onFocus={() => results.length > 0 && setShowDropdown(true)}
              placeholder="Type name or email…"
              autoComplete="off"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            {selectedUser && (
              <button type="button" onClick={() => { setSelectedUser(null); setQuery(''); setResults([]); }}
                className="text-gray-400 hover:text-gray-600 px-2 text-lg leading-none">✕</button>
            )}
          </div>

          {selectedUser && (
            <div className="mt-1.5 flex items-center gap-2 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-md text-sm">
              <span className="font-medium text-blue-900">{selectedUser.name}</span>
              <span className="text-blue-500 text-xs">{selectedUser.email}</span>
              <span className="ml-auto text-xs text-gray-400 capitalize">{selectedUser.role.replace('_', ' ')}</span>
            </div>
          )}

          {searching && <p className="text-xs text-gray-400 mt-1">Searching…</p>}

          {showDropdown && results.length > 0 && (
            <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg max-h-52 overflow-y-auto">
              {results.map((u) => (
                <button key={u.id} type="button" onClick={() => handleSelect(u)}
                  className="w-full text-left px-4 py-2.5 hover:bg-blue-50 transition-colors">
                  <div className="text-sm font-medium text-gray-900">{u.name}</div>
                  <div className="text-xs text-gray-500">{u.email} · <span className="capitalize">{u.role.replace('_', ' ')}</span></div>
                </button>
              ))}
            </div>
          )}
          {showDropdown && !searching && results.length === 0 && query.trim() && (
            <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-md shadow p-3 text-sm text-gray-500">
              No users found for "{query}"
            </div>
          )}
        </div>

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
          <Button type="submit" variant="primary" disabled={loading || !selectedUser}>
            {loading ? 'Transferring…' : 'Complete Handoff'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
