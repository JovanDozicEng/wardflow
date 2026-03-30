/**
 * Discharge Page - Discharge planning and checklists
 * Standalone discharge management page with encounter lookup
 */

import { useState, useCallback } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { EncounterAutocomplete } from '../shared/components/ui/EncounterAutocomplete';
import { dischargeService } from '../features/discharge/services/dischargeService';
import { useAuthStore } from '../features/auth/store/authStore';
import type { DischargeChecklist, DischargeType, ChecklistStatus, ItemStatus } from '../features/discharge/types';

const STATUS_COLORS: Record<ChecklistStatus, string> = {
  in_progress: 'bg-yellow-100 text-yellow-700',
  complete: 'bg-green-100 text-green-700',
  override_complete: 'bg-orange-100 text-orange-700',
};

const ITEM_STATUS_COLORS: Record<ItemStatus, string> = {
  open: 'bg-gray-100 text-gray-600',
  done: 'bg-green-100 text-green-700',
  waived: 'bg-blue-100 text-blue-700',
};

const InitChecklistModal = ({
  encounterId,
  isOpen,
  onClose,
  onCreated,
}: {
  encounterId: string;
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}) => {
  const [dischargeType, setDischargeType] = useState<DischargeType>('standard');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaving(true);
    try {
      await dischargeService.initChecklist(encounterId, { dischargeType });
      onCreated();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to initialize checklist');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Initialize Discharge Checklist">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Discharge Type <span className="text-red-500">*</span>
          </label>
          <select
            value={dischargeType}
            onChange={(e) => setDischargeType(e.target.value as DischargeType)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          >
            <option value="standard">Standard</option>
            <option value="ama">Against Medical Advice (AMA)</option>
            <option value="lwbs">Left Without Being Seen (LWBS)</option>
          </select>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Initializing…' : 'Initialize'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

const OverrideDischargeModal = ({
  encounterId,
  isOpen,
  onClose,
  onCompleted,
}: {
  encounterId: string;
  isOpen: boolean;
  onClose: () => void;
  onCompleted: () => void;
}) => {
  const [reason, setReason] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!reason.trim()) {
      setError('Override reason is required');
      return;
    }
    setSaving(true);
    try {
      await dischargeService.completeDischarge(encounterId, { override: true, reason: reason.trim() });
      onCompleted();
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to complete discharge');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Override Discharge Completion">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="bg-orange-50 border border-orange-200 rounded p-3 text-sm text-orange-800">
          ⚠️ You are overriding incomplete required items. This action requires a documented reason.
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Override Reason <span className="text-red-500">*</span>
          </label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="e.g. Patient left against medical advice, items not applicable"
            rows={4}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {saving ? 'Completing…' : 'Override & Complete'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

const ChecklistView = ({
  checklist,
  onRefresh,
}: {
  checklist: DischargeChecklist;
  onRefresh: () => void;
}) => {
  const { user } = useAuthStore();
  const [itemLoading, setItemLoading] = useState<string | null>(null);
  const [showOverride, setShowOverride] = useState(false);

  const items = checklist.items ?? [];
  const requiredItems = items.filter((i) => i.required);
  const optionalItems = items.filter((i) => !i.required);
  const completedRequired = requiredItems.filter((i) => i.status === 'done' || i.status === 'waived').length;
  const totalRequired = requiredItems.length;
  const progressPercent = totalRequired > 0 ? Math.round((completedRequired / totalRequired) * 100) : 100;

  const canOverride = user?.role === 'admin' || user?.role === 'charge_nurse';
  const allRequiredComplete = completedRequired === totalRequired;

  const handleCompleteItem = async (itemId: string) => {
    setItemLoading(itemId);
    try {
      await dischargeService.completeItem(itemId);
      onRefresh();
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Failed to complete item');
    } finally {
      setItemLoading(null);
    }
  };

  const handleCompleteDischarge = async () => {
    if (allRequiredComplete) {
      try {
        await dischargeService.completeDischarge(checklist.encounterId);
        onRefresh();
      } catch (err: any) {
        alert(err.response?.data?.error?.message || 'Failed to complete discharge');
      }
    } else if (canOverride) {
      setShowOverride(true);
    }
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Discharge Checklist</h3>
            <p className="text-sm text-gray-600 mt-1">Encounter: {checklist.encounterId}</p>
          </div>
          <div className="flex items-center gap-3">
            <span className="inline-flex px-3 py-1 rounded text-sm font-medium capitalize bg-blue-100 text-blue-700">
              {checklist.dischargeType.replace('_', ' ')}
            </span>
            <span
              className={`inline-flex px-3 py-1 rounded text-sm font-medium capitalize ${STATUS_COLORS[checklist.status]}`}
            >
              {checklist.status.replace('_', ' ')}
            </span>
          </div>
        </div>

        {/* Progress bar */}
        <div className="mb-2">
          <div className="flex items-center justify-between text-sm text-gray-600 mb-1">
            <span>Required Items Progress</span>
            <span className="font-medium">
              {completedRequired} / {totalRequired} ({progressPercent}%)
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-green-500 h-2 rounded-full transition-all"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>
      </div>

      {/* Required Items */}
      {requiredItems.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6">
          <h4 className="text-md font-semibold text-gray-900 mb-3">Required Items</h4>
          <div className="space-y-2">
            {requiredItems.map((item) => (
              <div key={item.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                <input
                  type="checkbox"
                  checked={item.status === 'done' || item.status === 'waived'}
                  disabled={item.status !== 'open' || itemLoading === item.id}
                  onChange={() => handleCompleteItem(item.id)}
                  className="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                />
                <div className="flex-1">
                  <span className="text-sm font-medium text-gray-900">{item.label}</span>
                </div>
                <span
                  className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${ITEM_STATUS_COLORS[item.status]}`}
                >
                  {item.status}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Optional Items */}
      {optionalItems.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6">
          <h4 className="text-md font-semibold text-gray-900 mb-3">Optional Items</h4>
          <div className="space-y-2">
            {optionalItems.map((item) => (
              <div key={item.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                <input
                  type="checkbox"
                  checked={item.status === 'done' || item.status === 'waived'}
                  disabled={item.status !== 'open' || itemLoading === item.id}
                  onChange={() => handleCompleteItem(item.id)}
                  className="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                />
                <div className="flex-1">
                  <span className="text-sm font-medium text-gray-900">{item.label}</span>
                </div>
                <span
                  className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${ITEM_STATUS_COLORS[item.status]}`}
                >
                  {item.status}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Complete Discharge Button */}
      {checklist.status === 'in_progress' && (
        <div className="flex justify-end">
          <Button
            variant="primary"
            onClick={handleCompleteDischarge}
            disabled={!allRequiredComplete && !canOverride}
            title={
              !allRequiredComplete && !canOverride
                ? 'Complete all required items or contact an admin/charge nurse'
                : undefined
            }
          >
            Complete Discharge
          </Button>
        </div>
      )}

      <OverrideDischargeModal
        encounterId={checklist.encounterId}
        isOpen={showOverride}
        onClose={() => setShowOverride(false)}
        onCompleted={onRefresh}
      />
    </div>
  );
};

export const DischargePage = () => {
  const [encounterId, setEncounterId] = useState('');
  const [checklist, setChecklist] = useState<DischargeChecklist | null>(null);
  const [loading, setLoading] = useState(false);
  const [notFound, setNotFound] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showInit, setShowInit] = useState(false);

  const fetchChecklist = useCallback(async () => {
    if (!encounterId) return;
    setLoading(true);
    setError(null);
    setNotFound(false);
    setChecklist(null);
    try {
      const data = await dischargeService.getChecklist(encounterId);
      setChecklist(data);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setNotFound(true);
      } else {
        setError(err.response?.data?.error?.message || 'Failed to load checklist');
      }
    } finally {
      setLoading(false);
    }
  }, [encounterId]);

  const handleEncounterChange = (id: string) => {
    setEncounterId(id);
    setChecklist(null);
    setNotFound(false);
    setError(null);
  };

  const handleInitChecklist = () => {
    setShowInit(true);
  };

  return (
    <Layout>
      <div className="px-4 pt-6 pb-2">
        <PageHeader
          title="Discharge Planning"
          subtitle="Discharge checklists and coordination"
        />
      </div>

      <div className="px-4 pb-8 max-w-4xl">
        {/* Encounter Selection */}
        <div className="mb-6">
          <EncounterAutocomplete
            label="Select Encounter"
            value={encounterId}
            onChange={handleEncounterChange}
            placeholder="Search encounters..."
          />
          {encounterId && (
            <div className="mt-2">
              <Button variant="secondary" onClick={fetchChecklist} disabled={loading}>
                {loading ? 'Loading…' : 'Load Checklist'}
              </Button>
            </div>
          )}
        </div>

        {/* Loading */}
        {loading && (
          <div className="flex justify-center py-12">
            <Spinner />
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
            {error}
          </div>
        )}

        {/* Not Found */}
        {notFound && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
            <p className="text-lg font-medium text-yellow-900 mb-2">No checklist found</p>
            <p className="text-sm text-yellow-700 mb-4">
              This encounter does not have a discharge checklist yet.
            </p>
            <Button variant="primary" onClick={handleInitChecklist}>
              Initialize Checklist
            </Button>
          </div>
        )}

        {/* Checklist */}
        {checklist && <ChecklistView checklist={checklist} onRefresh={fetchChecklist} />}

        {/* Empty State */}
        {!loading && !error && !notFound && !checklist && encounterId === '' && (
          <div className="text-center py-12 text-gray-500">
            <p className="text-lg mb-1">Select an encounter to view discharge checklist</p>
            <p className="text-sm">Use the search above to look up an encounter.</p>
          </div>
        )}
      </div>

      <InitChecklistModal
        encounterId={encounterId}
        isOpen={showInit}
        onClose={() => setShowInit(false)}
        onCreated={fetchChecklist}
      />
    </Layout>
  );
};

export default DischargePage;
