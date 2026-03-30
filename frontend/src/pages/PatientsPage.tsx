/**
 * PatientsPage - Patient list with search and create
 */

import { useEffect, useState, useCallback } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { Input } from '../shared/components/ui/Input';
import { listPatients, createPatient } from '../features/patients/services/patientService';
import type { Patient, CreatePatientRequest } from '../features/patients/types';

const formatDate = (iso?: string) => {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
};

const CreatePatientModal = ({
  isOpen,
  onClose,
  onCreated,
}: {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [mrn, setMrn] = useState('');
  const [dateOfBirth, setDateOfBirth] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!firstName.trim() || !lastName.trim() || !mrn.trim()) {
      setError('First Name, Last Name, and MRN are required');
      return;
    }
    setSaving(true);
    try {
      const data: CreatePatientRequest = {
        firstName: firstName.trim(),
        lastName: lastName.trim(),
        mrn: mrn.trim(),
      };
      if (dateOfBirth) {
        data.dateOfBirth = dateOfBirth;
      }
      await createPatient(data);
      setFirstName('');
      setLastName('');
      setMrn('');
      setDateOfBirth('');
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create patient');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Patient">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}
        <Input
          label="First Name"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
          placeholder="e.g. John"
          required
        />
        <Input
          label="Last Name"
          value={lastName}
          onChange={(e) => setLastName(e.target.value)}
          placeholder="e.g. Smith"
          required
        />
        <Input
          label="Medical Record Number (MRN)"
          value={mrn}
          onChange={(e) => setMrn(e.target.value)}
          placeholder="e.g. MRN-12345"
          required
        />
        <Input
          label="Date of Birth (optional)"
          type="date"
          value={dateOfBirth}
          onChange={(e) => setDateOfBirth(e.target.value)}
        />
        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create Patient'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export const PatientsPage = () => {
  const [patients, setPatients] = useState<Patient[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const fetchPatients = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await listPatients(debouncedSearch, 50, 0);
      const data = Array.isArray(response.data) ? response.data : [];
      setPatients(data);
      setTotal(response.total ?? data.length);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load patients');
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch]);

  useEffect(() => {
    fetchPatients();
  }, [fetchPatients]);

  return (
    <Layout>
      <div className="flex items-center justify-between px-4 pt-6 pb-2">
        <PageHeader
          title="Patients"
          subtitle={`Patient records${total > 0 ? ` — ${total} total` : ''}`}
        />
        <Button
          variant="primary"
          onClick={() => setShowCreate(true)}
          className="shrink-0 px-4 py-2"
        >
          + New Patient
        </Button>
      </div>

      {/* Search bar */}
      <div className="px-4 pb-4 flex flex-wrap gap-3 items-center">
        <Input
          placeholder="Search by name or MRN..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-64"
        />
        {searchQuery && (
          <button
            onClick={() => setSearchQuery('')}
            className="text-sm text-gray-500 hover:text-gray-700 underline"
          >
            Clear search
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

        {!loading && !error && patients.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">No patients found</p>
            <p className="text-sm">
              {searchQuery
                ? 'Try a different search term.'
                : 'Click + New Patient to create one.'}
            </p>
          </div>
        )}

        {!loading && patients.length > 0 && (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    MRN
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Date of Birth
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-100">
                {patients.map((patient) => (
                  <tr key={patient.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-sm font-medium text-gray-900">
                        {patient.firstName} {patient.lastName}
                      </span>
                      <p className="text-xs text-gray-400 mt-0.5 font-mono">{patient.id}</p>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 font-mono">
                      {patient.mrn}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {formatDate(patient.dateOfBirth)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreatePatientModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={fetchPatients}
      />
    </Layout>
  );
};

export default PatientsPage;
