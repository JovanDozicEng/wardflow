/**
 * EncountersPage - Active patient encounters list
 */

import { useEffect, useState, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { Input } from '../shared/components/ui/Input';
import { ROUTES, buildRoute } from '../shared/config/routes';
import api from '../shared/utils/api';
import { searchPatients } from '../features/patients/services/patientService';
import type { Patient } from '../features/patients/types';

interface Encounter {
  id: string;
  patientId: string;
  unitId: string;
  departmentId: string;
  status: 'active' | 'discharged' | 'cancelled';
  startedAt: string;
  endedAt?: string;
}

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-100 text-green-800',
  discharged: 'bg-gray-100 text-gray-700',
  cancelled: 'bg-red-100 text-red-700',
};

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

const CreateEncounterModal = ({
  isOpen,
  onClose,
  onCreated,
}: {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [patientQuery, setPatientQuery] = useState('');
  const [patientResults, setPatientResults] = useState<Patient[]>([]);
  const [selectedPatient, setSelectedPatient] = useState<Patient | null>(null);
  const [loadingPatients, setLoadingPatients] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [unitId, setUnitId] = useState('');
  const [departmentId, setDepartmentId] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Debounce patient search
  useEffect(() => {
    if (!patientQuery.trim() || patientQuery.length < 2) {
      setPatientResults([]);
      setShowDropdown(false);
      return;
    }

    const timer = setTimeout(async () => {
      setLoadingPatients(true);
      try {
        const results = await searchPatients(patientQuery);
        setPatientResults(results);
        setShowDropdown(results.length > 0);
      } catch (err) {
        setPatientResults([]);
        setShowDropdown(false);
      } finally {
        setLoadingPatients(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [patientQuery]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handlePatientSelect = (patient: Patient) => {
    setSelectedPatient(patient);
    setPatientQuery('');
    setShowDropdown(false);
    setPatientResults([]);
  };

  const handleClearPatient = () => {
    setSelectedPatient(null);
    setPatientQuery('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!selectedPatient) {
      setError('Please search and select a patient');
      return;
    }
    if (!unitId.trim() || !departmentId.trim()) {
      setError('Unit ID and Department ID are required');
      return;
    }
    setSaving(true);
    try {
      await api.post('/encounters', {
        patientId: selectedPatient.id,
        unitId: unitId.trim(),
        departmentId: departmentId.trim(),
      });
      setSelectedPatient(null);
      setPatientQuery('');
      setUnitId('');
      setDepartmentId('');
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create encounter');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Encounter">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        {/* Patient Search/Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Patient <span className="text-red-500">*</span>
          </label>
          {selectedPatient ? (
            <div className="flex items-center gap-2 px-3 py-2 bg-blue-50 border border-blue-200 rounded-md">
              <div className="flex-1">
                <p className="text-sm font-medium text-blue-900">
                  {selectedPatient.firstName} {selectedPatient.lastName}
                </p>
                <p className="text-xs text-blue-600">MRN: {selectedPatient.mrn}</p>
              </div>
              <button
                type="button"
                onClick={handleClearPatient}
                className="text-sm text-blue-600 hover:text-blue-800 font-medium"
              >
                Change
              </button>
            </div>
          ) : (
            <div className="relative" ref={dropdownRef}>
              <input
                type="text"
                value={patientQuery}
                onChange={(e) => setPatientQuery(e.target.value)}
                placeholder="Search by name or MRN..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {loadingPatients && (
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  <div className="w-4 h-4 border-2 border-gray-300 border-t-blue-500 rounded-full animate-spin" />
                </div>
              )}
              {showDropdown && patientResults.length > 0 && (
                <ul className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
                  {patientResults.map((patient) => (
                    <li key={patient.id}>
                      <button
                        type="button"
                        onClick={() => handlePatientSelect(patient)}
                        className="w-full px-3 py-2 text-left hover:bg-gray-50 transition-colors"
                      >
                        <p className="text-sm font-medium text-gray-900">
                          {patient.firstName} {patient.lastName}
                        </p>
                        <p className="text-xs text-gray-500">MRN: {patient.mrn}</p>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>

        <Input
          label="Unit ID"
          value={unitId}
          onChange={(e) => setUnitId(e.target.value)}
          placeholder="e.g. unit-icu"
          required
        />
        <Input
          label="Department ID"
          value={departmentId}
          onChange={(e) => setDepartmentId(e.target.value)}
          placeholder="e.g. dept-emergency"
          required
        />
        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create Encounter'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export const EncountersPage = () => {
  const [encounters, setEncounters] = useState<Encounter[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [filterUnit, setFilterUnit] = useState('');
  const [filterStatus, setFilterStatus] = useState('');

  const fetchEncounters = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params: Record<string, string | number> = { limit: 50, offset: 0 };
      if (filterUnit.trim()) params.unitId = filterUnit.trim();
      if (filterStatus) params.status = filterStatus;
      const res = await api.get<{ data: Encounter[]; total: number }>(
        '/encounters',
        { params }
      );
      const data = Array.isArray(res.data?.data) ? res.data.data : [];
      setEncounters(data);
      setTotal(res.data?.total ?? data.length);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load encounters');
    } finally {
      setLoading(false);
    }
  }, [filterUnit, filterStatus]);

  useEffect(() => {
    fetchEncounters();
  }, [fetchEncounters]);

  return (
    <Layout>
      <div className="flex items-center justify-between px-4 pt-6 pb-2">
        <PageHeader
          title="Encounters"
          subtitle={`Active patient encounters${total > 0 ? ` — ${total} total` : ''}`}
        />
        <Button
          variant="primary"
          onClick={() => setShowCreate(true)}
          className="shrink-0 px-4 py-2"
        >
          + New Encounter
        </Button>
      </div>

      {/* Filter bar */}
      <div className="px-4 pb-4 flex flex-wrap gap-3 items-center">
        <Input
          placeholder="Filter by Unit ID"
          value={filterUnit}
          onChange={(e) => setFilterUnit(e.target.value)}
          className="w-48"
        />
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value)}
          className="border border-gray-300 rounded-md px-3 py-2 text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">All Statuses</option>
          <option value="active">Active</option>
          <option value="discharged">Discharged</option>
          <option value="cancelled">Cancelled</option>
        </select>
        {(filterUnit || filterStatus) && (
          <button
            onClick={() => { setFilterUnit(''); setFilterStatus(''); }}
            className="text-sm text-gray-500 hover:text-gray-700 underline"
          >
            Clear filters
          </button>
        )}
      </div>

      <div className="px-4 pb-8">
        {loading && (
          <div className="flex justify-center py-12">
            <Spinner />
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
            {error}
          </div>
        )}

        {!loading && !error && encounters.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">No encounters found</p>
            <p className="text-sm">
              Click <strong>+ New Encounter</strong> to create one.
            </p>
          </div>
        )}

        {!loading && encounters.length > 0 && (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Patient ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Unit
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Started
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-100">
                {encounters.map((enc) => (
                  <tr key={enc.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-sm font-medium text-gray-900 font-mono">
                        {enc.patientId}
                      </span>
                      <p className="text-xs text-gray-400 mt-0.5 font-mono">{enc.id}</p>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {enc.unitId}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize ${
                          STATUS_COLORS[enc.status] ?? 'bg-gray-100 text-gray-700'
                        }`}
                      >
                        {enc.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {formatDate(enc.startedAt)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        to={buildRoute(ROUTES.ENCOUNTER_DETAIL, { id: enc.id })}
                        className="inline-flex items-center px-3 py-1.5 text-xs font-medium rounded-md bg-blue-50 text-blue-700 hover:bg-blue-100 transition-colors"
                      >
                        View →
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreateEncounterModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={fetchEncounters}
      />
    </Layout>
  );
};

export default EncountersPage;

