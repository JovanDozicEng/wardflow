/**
 * DepartmentsPage - Admin-only department management
 */

import { useEffect, useState, useCallback } from 'react';
import { Navigate } from 'react-router-dom';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { Input } from '../shared/components/ui/Input';
import { listDepartments, createDepartment } from '../features/units/services/unitService';
import { useAuthStore } from '../features/auth/store/authStore';
import type { Department, CreateDepartmentRequest } from '../features/units/types';
import { ROUTES } from '../shared/config/routes';

const formatDate = (iso?: string) => {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
};

const CreateDepartmentModal = ({
  isOpen,
  onClose,
  onCreated,
}: {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!name.trim() || !code.trim()) {
      setError('Name and Code are required');
      return;
    }
    setSaving(true);
    try {
      const data: CreateDepartmentRequest = {
        name: name.trim(),
        code: code.trim(),
      };
      await createDepartment(data);
      setName('');
      setCode('');
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create department');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Department">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}
        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Emergency Medicine"
          required
        />
        <Input
          label="Code"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          placeholder="e.g. EMERGENCY"
          required
        />
        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create Department'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export const DepartmentsPage = () => {
  const { user } = useAuthStore();
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Admin guard
  if (user?.role !== 'admin') {
    return <Navigate to={ROUTES.UNAUTHORIZED} replace />;
  }

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const fetchDepartments = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listDepartments(debouncedSearch);
      setDepartments(Array.isArray(data) ? data : []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load departments');
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch]);

  useEffect(() => {
    fetchDepartments();
  }, [fetchDepartments]);

  return (
    <Layout>
      <div className="flex items-center justify-between px-4 pt-6 pb-2">
        <PageHeader
          title="Departments"
          subtitle={`Manage departments${departments.length > 0 ? ` — ${departments.length} total` : ''}`}
        />
        <Button
          variant="primary"
          onClick={() => setShowCreate(true)}
          className="shrink-0 px-4 py-2"
        >
          + New Department
        </Button>
      </div>

      {/* Search bar */}
      <div className="px-4 pb-4 flex flex-wrap gap-3 items-center">
        <Input
          placeholder="Search by name or code..."
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

        {!loading && !error && departments.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">No departments found</p>
            <p className="text-sm">
              {searchQuery
                ? 'Try a different search term.'
                : 'Click + New Department to create one.'}
            </p>
          </div>
        )}

        {!loading && departments.length > 0 && (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-100">
                {departments.map((dept) => (
                  <tr key={dept.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-sm font-medium text-gray-900">
                        {dept.name}
                      </span>
                      <p className="text-xs text-gray-400 mt-0.5 font-mono">{dept.id}</p>
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                        {dept.code}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {formatDate(dept.createdAt)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-400">
                      —
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreateDepartmentModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={fetchDepartments}
      />
    </Layout>
  );
};

export default DepartmentsPage;
