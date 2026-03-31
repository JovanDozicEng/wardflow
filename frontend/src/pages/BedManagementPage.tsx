/**
 * Bed Management Page - Full bed board implementation
 * Shows bed grid with status filtering and inline status changes
 */

import { useEffect, useState, useCallback } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { Input } from '../shared/components/ui/Input';
import { UnitAutocomplete } from '../shared/components/ui/UnitAutocomplete';
import { bedService } from '../features/beds/services/bedService';
import { useAuthStore } from '../features/auth/store/authStore';
import type { Bed, BedStatus } from '../features/beds/types';

const BED_STATUS_COLORS: Record<BedStatus, string> = {
  available: 'bg-green-50 border-green-300',
  cleaning: 'bg-blue-50 border-blue-300',
  occupied: 'bg-red-50 border-red-300',
  blocked: 'bg-red-50 border-red-300',
  maintenance: 'bg-gray-100 border-gray-300',
};

const BED_STATUS_BADGE_COLORS: Record<BedStatus, string> = {
  available: 'bg-green-100 text-green-700',
  cleaning: 'bg-blue-100 text-blue-700',
  occupied: 'bg-red-100 text-red-700',
  blocked: 'bg-red-100 text-red-700',
  maintenance: 'bg-gray-200 text-gray-700',
};

const ALL_STATUSES: BedStatus[] = ['available', 'occupied', 'blocked', 'cleaning', 'maintenance'];

const StatusChangeModal = ({
  bed,
  isOpen,
  onClose,
  onUpdated,
}: {
  bed: Bed;
  isOpen: boolean;
  onClose: () => void;
  onUpdated: () => void;
}) => {
  const [newStatus, setNewStatus] = useState<BedStatus>(bed.currentStatus);
  const [reason, setReason] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newStatus === bed.currentStatus) {
      onClose();
      return;
    }
    setError(null);
    setSaving(true);
    try {
      await bedService.updateBedStatus(bed.id, { status: newStatus, reason: reason.trim() || undefined });
      onUpdated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to update bed status');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Change Bed Status">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div>
          <p className="text-sm font-medium text-gray-700 mb-2">
            Bed: <span className="font-bold">{bed.room} - {bed.label}</span>
          </p>
          <p className="text-sm text-gray-600">
            Current Status:{' '}
            <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${BED_STATUS_BADGE_COLORS[bed.currentStatus]}`}>
              {bed.currentStatus}
            </span>
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            New Status <span className="text-red-500">*</span>
          </label>
          <select
            value={newStatus}
            onChange={(e) => setNewStatus(e.target.value as BedStatus)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          >
            {ALL_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Reason (optional)
          </label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="e.g. Patient admitted, Room cleaning completed"
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Updating…' : 'Update Status'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

const CreateBedModal = ({
  isOpen,
  onClose,
  onCreated,
}: {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [unitId, setUnitId] = useState('');
  const [room, setRoom] = useState('');
  const [label, setLabel] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!unitId || !room.trim() || !label.trim()) {
      setError('Unit, Room, and Label are required');
      return;
    }
    setSaving(true);
    try {
      await bedService.createBed({ unitId, room: room.trim(), label: label.trim() });
      setUnitId('');
      setRoom('');
      setLabel('');
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to create bed');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="New Bed">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <UnitAutocomplete
          label="Unit"
          value={unitId}
          onChange={setUnitId}
          required
          placeholder="Select unit..."
        />

        <Input
          label="Room"
          value={room}
          onChange={(e) => setRoom(e.target.value)}
          placeholder="e.g. 101"
          required
        />

        <Input
          label="Bed Label"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          placeholder="e.g. A, B, 1, 2"
          required
        />

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create Bed'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export const BedManagementPage = () => {
  const { user } = useAuthStore();
  const [beds, setBeds] = useState<Bed[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [unitFilter, setUnitFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showCreate, setShowCreate] = useState(false);
  const [selectedBed, setSelectedBed] = useState<Bed | null>(null);

  const isAdmin = user?.role === 'admin';

  const fetchBeds = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params: { unitId?: string; status?: string; limit: number } = { limit: 100 };
      if (unitFilter) params.unitId = unitFilter;
      if (statusFilter && statusFilter !== 'all') params.status = statusFilter;

      const response = await bedService.listBeds(params);
      const data = Array.isArray(response.data) ? response.data : [];
      setBeds(data);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load beds');
    } finally {
      setLoading(false);
    }
  }, [unitFilter, statusFilter]);

  useEffect(() => {
    fetchBeds();
  }, [fetchBeds]);

  return (
    <Layout>
      <div className="flex items-center justify-between px-4 pt-6 pb-2">
        <PageHeader
          title="Bed Management"
          subtitle={`Bed status board${beds.length > 0 ? ` — ${beds.length} beds` : ''}`}
        />
        {isAdmin && (
          <Button
            variant="primary"
            onClick={() => setShowCreate(true)}
            className="shrink-0 px-4 py-2"
          >
            + New Bed
          </Button>
        )}
      </div>

      {/* Filter bar */}
      <div className="px-4 pb-4 flex flex-wrap gap-3 items-center">
        <div className="w-64">
          <UnitAutocomplete
            value={unitFilter}
            onChange={setUnitFilter}
            placeholder="Filter by unit..."
          />
        </div>

        <div className="w-48">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Statuses</option>
            {ALL_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </option>
            ))}
          </select>
        </div>

        {(unitFilter || statusFilter !== 'all') && (
          <button
            onClick={() => {
              setUnitFilter('');
              setStatusFilter('all');
            }}
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

        {!loading && !error && beds.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">No beds found</p>
            <p className="text-sm">
              {unitFilter || statusFilter !== 'all'
                ? 'Try adjusting your filters.'
                : isAdmin
                ? 'Click + New Bed to create one.'
                : 'Contact an administrator to create beds.'}
            </p>
          </div>
        )}

        {!loading && beds.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {beds.map((bed) => (
              <button
                key={bed.id}
                onClick={() => setSelectedBed(bed)}
                className={`p-4 border-2 rounded-lg text-left transition-all hover:shadow-md ${BED_STATUS_COLORS[bed.currentStatus]}`}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="font-medium text-gray-900">
                    {bed.room} - {bed.label}
                  </div>
                  <span
                    className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${BED_STATUS_BADGE_COLORS[bed.currentStatus]}`}
                  >
                    {bed.currentStatus}
                  </span>
                </div>
                <div className="text-xs text-gray-500 font-mono truncate">{bed.id}</div>
                {bed.capabilities && bed.capabilities.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {bed.capabilities.slice(0, 2).map((cap) => (
                      <span
                        key={cap}
                        className="inline-flex px-1.5 py-0.5 bg-white rounded text-xs text-gray-600"
                      >
                        {cap}
                      </span>
                    ))}
                    {bed.capabilities.length > 2 && (
                      <span className="inline-flex px-1.5 py-0.5 bg-white rounded text-xs text-gray-600">
                        +{bed.capabilities.length - 2}
                      </span>
                    )}
                  </div>
                )}
              </button>
            ))}
          </div>
        )}
      </div>

      {selectedBed && (
        <StatusChangeModal
          bed={selectedBed}
          isOpen={!!selectedBed}
          onClose={() => setSelectedBed(null)}
          onUpdated={() => {
            fetchBeds();
            setSelectedBed(null);
          }}
        />
      )}

      <CreateBedModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={fetchBeds}
      />
    </Layout>
  );
};

export default BedManagementPage;
