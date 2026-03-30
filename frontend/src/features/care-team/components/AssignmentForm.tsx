/**
 * AssignmentForm - Assign a user to a care team role on an encounter
 */

import { useState, useEffect, useRef } from 'react';
import { Modal } from '@/shared/components/ui/Modal';
import { Button } from '@/shared/components/ui/Button';
import { Select } from '@/shared/components/ui/Select';
import type { AssignRoleRequest, CareTeamRole } from '../types/careTeam.types';
import api from '@/shared/utils/api';

const ROLE_OPTIONS = [
  { value: 'primary_nurse', label: 'Primary Nurse' },
  { value: 'attending_provider', label: 'Attending Provider' },
  { value: 'consulting_provider', label: 'Consulting Provider' },
  { value: 'resident', label: 'Resident' },
  { value: 'respiratory_therapist', label: 'Respiratory Therapist' },
  { value: 'case_manager', label: 'Case Manager' },
  { value: 'social_worker', label: 'Social Worker' },
  { value: 'other', label: 'Other' },
];

interface UserSummary {
  id: string;
  name: string;
  email: string;
  role: string;
}

interface AssignmentFormProps {
  encounterId: string;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (encounterId: string, data: AssignRoleRequest) => Promise<void>;
  loading?: boolean;
}

export const AssignmentForm = ({
  encounterId,
  isOpen,
  onClose,
  onSubmit,
  loading = false,
}: AssignmentFormProps) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<UserSummary[]>([]);
  const [selectedUser, setSelectedUser] = useState<UserSummary | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [searching, setSearching] = useState(false);
  const [roleType, setRoleType] = useState<CareTeamRole>('primary_nurse');
  const [startsAt, setStartsAt] = useState('');
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (isOpen) {
      setQuery(''); setResults([]); setSelectedUser(null);
      setShowDropdown(false); setRoleType('primary_nurse');
      setStartsAt(''); setError(null);
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
  }, [query, selectedUser]);

  const handleSelect = (user: UserSummary) => {
    setSelectedUser(user); setQuery(user.name); setShowDropdown(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!selectedUser) { setError('Please search and select a user'); return; }
    const normalizeTs = (dt: string) => {
      if (!dt) return undefined;
      if (/T\d{2}:\d{2}:\d{2}/.test(dt)) return dt.endsWith('Z') ? dt : dt + 'Z';
      return dt + ':00Z';
    };
    try {
      await onSubmit(encounterId, { userId: selectedUser.id, roleType, startsAt: normalizeTs(startsAt) });
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || 'Failed to assign role');
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Assign Care Team Role">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">{error}</div>
        )}

        {/* User search */}
        <div className="relative">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Search User <span className="text-red-500">*</span>
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

        <Select
          label="Role"
          value={roleType}
          onChange={(e) => setRoleType(e.target.value as CareTeamRole)}
          options={ROLE_OPTIONS}
        />

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Starts At <span className="text-gray-400 text-xs">(defaults to now)</span>
          </label>
          <input
            type="datetime-local"
            value={startsAt}
            onChange={(e) => setStartsAt(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
          <Button type="submit" variant="primary" disabled={loading || !selectedUser}>
            {loading ? 'Assigning…' : 'Assign Role'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};