/**
 * Transport Page - Transport request management
 * Dispatch queue with status tracking and request creation
 */

import { useEffect, useState, useCallback } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { Input } from '../shared/components/ui/Input';
import { EncounterAutocomplete } from '../shared/components/ui/EncounterAutocomplete';
import { transportService } from '../features/transport/services/transportService';
import { useAuthStore } from '../features/auth/store/authStore';
import type { TransportRequest, TransportStatus, TransportPriority } from '../features/transport/types';

const PRIORITY_COLORS: Record<TransportPriority, string> = {
  routine: 'bg-gray-100 text-gray-700',
  urgent: 'bg-yellow-100 text-yellow-700',
  emergent: 'bg-red-100 text-red-700',
};

const STATUS_COLORS: Record<TransportStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  assigned: 'bg-blue-100 text-blue-700',
  in_transit: 'bg-purple-100 text-purple-700',
  completed: 'bg-green-100 text-green-700',
  cancelled: 'bg-gray-100 text-gray-600',
};

const formatDateTime = (iso: string) => {
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const shortId = (id: string) => id.slice(0, 8) + '…';

const CreateTransportModal = ({
  isOpen,
  onClose,
  onCreated,
}: {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [encounterId, setEncounterId] = useState('');
  const [origin, setOrigin] = useState('');
  const [destination, setDestination] = useState('');
  const [priority, setPriority] = useState<TransportPriority>('routine');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!encounterId || !origin.trim() || !destination.trim()) {
      setError('Encounter, Origin, and Destination are required');
      return;
    }
    setSaving(true);
    try {
      await transportService.createRequest({
        encounterId,
        origin: origin.trim(),
        destination: destination.trim(),
        priority,
      });
      setEncounterId('');
      setOrigin('');
      setDestination('');
      setPriority('routine');
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create transport request');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Transport Request">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <EncounterAutocomplete
          label="Encounter"
          value={encounterId}
          onChange={setEncounterId}
          required
          placeholder="Search encounters..."
        />

        <Input
          label="Origin"
          value={origin}
          onChange={(e) => setOrigin(e.target.value)}
          placeholder="e.g. ICU North - Room 12"
          required
        />

        <Input
          label="Destination"
          value={destination}
          onChange={(e) => setDestination(e.target.value)}
          placeholder="e.g. Radiology - CT Suite 2"
          required
        />

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Priority
          </label>
          <select
            value={priority}
            onChange={(e) => setPriority(e.target.value as TransportPriority)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="routine">Routine</option>
            <option value="urgent">Urgent</option>
            <option value="emergent">Emergent</option>
          </select>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create Request'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export const TransportPage = () => {
  const { user } = useAuthStore();
  const [requests, setRequests] = useState<TransportRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showCreate, setShowCreate] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchRequests = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params: { status?: string; limit: number } = { limit: 50 };
      if (statusFilter && statusFilter !== 'all') params.status = statusFilter;

      const response = await transportService.listRequests(params);
      const data = Array.isArray(response.data) ? response.data : [];
      setRequests(data);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load transport requests');
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    fetchRequests();
  }, [fetchRequests]);

  const handleAccept = async (id: string) => {
    if (!user) return;
    setActionLoading(id);
    try {
      await transportService.acceptRequest(id, { assignedTo: user.id });
      fetchRequests();
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Failed to accept request');
    } finally {
      setActionLoading(null);
    }
  };

  const handleComplete = async (id: string) => {
    setActionLoading(id);
    try {
      await transportService.completeRequest(id);
      fetchRequests();
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Failed to complete request');
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <Layout>
      <div className="flex items-center justify-between px-4 pt-6 pb-2">
        <PageHeader
          title="Transport"
          subtitle={`Patient transport dispatch${requests.length > 0 ? ` — ${requests.length} requests` : ''}`}
        />
        <Button
          variant="primary"
          onClick={() => setShowCreate(true)}
          className="shrink-0 px-4 py-2"
        >
          + New Transport Request
        </Button>
      </div>

      {/* Filter bar */}
      <div className="px-4 pb-4 flex flex-wrap gap-3 items-center">
        <div className="w-48">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Statuses</option>
            <option value="pending">Pending</option>
            <option value="assigned">Assigned</option>
            <option value="in_transit">In Transit</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>

        {statusFilter !== 'all' && (
          <button
            onClick={() => setStatusFilter('all')}
            className="text-sm text-gray-500 hover:text-gray-700 underline"
          >
            Clear filter
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

        {!loading && !error && requests.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">No transport requests found</p>
            <p className="text-sm">
              {statusFilter !== 'all'
                ? 'Try adjusting your filter.'
                : 'Click + New Transport Request to create one.'}
            </p>
          </div>
        )}

        {!loading && requests.length > 0 && (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Encounter
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Route
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Priority
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Assigned To
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
                {requests.map((req) => (
                  <tr key={req.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-sm font-mono text-gray-600">
                        {shortId(req.encounterId)}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900">
                        {req.origin} → {req.destination}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${PRIORITY_COLORS[req.priority]}`}
                      >
                        {req.priority}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${STATUS_COLORS[req.status]}`}
                      >
                        {req.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {req.assignedTo ? shortId(req.assignedTo) : '—'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {formatDateTime(req.createdAt)}
                    </td>
                    <td className="px-6 py-4">
                      {req.status === 'pending' && (
                        <Button
                          variant="primary"
                          onClick={() => handleAccept(req.id)}
                          disabled={actionLoading === req.id}
                          className="px-3 py-1 text-xs"
                        >
                          {actionLoading === req.id ? 'Accepting…' : 'Accept'}
                        </Button>
                      )}
                      {req.status === 'assigned' && (
                        <Button
                          variant="primary"
                          onClick={() => handleComplete(req.id)}
                          disabled={actionLoading === req.id}
                          className="px-3 py-1 text-xs"
                        >
                          {actionLoading === req.id ? 'Completing…' : 'Complete'}
                        </Button>
                      )}
                      {(req.status === 'completed' || req.status === 'cancelled') && (
                        <span className="text-xs text-gray-400">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CreateTransportModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={fetchRequests}
      />
    </Layout>
  );
};

export default TransportPage;
