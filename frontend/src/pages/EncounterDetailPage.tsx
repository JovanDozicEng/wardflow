/**
 * EncounterDetailPage - Tabbed encounter detail view
 * Tabs: Care Team | Flow | Tasks | Discharge
 */

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';

// Care Team
import { CareTeamList } from '../features/care-team/components/CareTeamList';
import { usePermissions } from '../features/auth/hooks/usePermissions';

// Flow
import { FlowTimeline } from '../features/flow/components/FlowTimeline';
import { TransitionStateButton } from '../features/flow/components/TransitionStateButton';
import { useFlowTracking } from '../features/flow/hooks/useFlowTracking';

// Tasks
import { TaskBoard } from '../features/tasks/components/TaskBoard';
import { CreateTaskForm } from '../features/tasks/components/CreateTaskForm';
import { taskService } from '../features/tasks/services/taskService';
import type { Task, CreateTaskRequest } from '../features/tasks/types';

// Discharge
import { dischargeService } from '../features/discharge/services/dischargeService';
import { useAuthStore } from '../features/auth/store/authStore';
import type { DischargeChecklist, DischargeType, ChecklistStatus, ItemStatus } from '../features/discharge/types';

import { Permission } from '../shared/types';
import { ROUTES } from '../shared/config/routes';

type Tab = 'care-team' | 'flow' | 'tasks' | 'discharge';

const TABS: { id: Tab; label: string }[] = [
  { id: 'care-team', label: 'Care Team' },
  { id: 'flow', label: 'Patient Flow' },
  { id: 'tasks', label: 'Tasks' },
  { id: 'discharge', label: 'Discharge' },
];

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

export const EncounterDetailPage = () => {
  const { id: encounterId } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState<Tab>('care-team');
  const { hasPermission } = usePermissions();
  const { user } = useAuthStore();

  // Care Team permissions
  const canAssign = hasPermission(Permission.ASSIGN_CARE_TEAM);
  const canTransfer = hasPermission(Permission.TRANSFER_CARE_TEAM);

  // Flow permissions
  const canUpdateFlow = hasPermission(Permission.UPDATE_FLOW);
  const canOverrideFlow = hasPermission(Permission.OVERRIDE_FLOW);

  // Flow state
  const { currentState, transitions, isLoading: flowLoading, recordTransition, overrideTransition } =
    useFlowTracking({ encounterId: encounterId ?? '' });

  // Tasks state
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [showCreateTask, setShowCreateTask] = useState(false);

  // Discharge state
  const [checklist, setChecklist] = useState<DischargeChecklist | null>(null);
  const [checklistLoading, setChecklistLoading] = useState(false);
  const [checklistNotFound, setChecklistNotFound] = useState(false);
  const [showInitChecklist, setShowInitChecklist] = useState(false);
  const [showOverrideDischarge, setShowOverrideDischarge] = useState(false);
  const [itemLoading, setItemLoading] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    if (!encounterId) return;
    setTasksLoading(true);
    try {
      const response = await taskService.listTasks({ scopeType: 'encounter', scopeId: encounterId });
      const data = response.tasks ?? [];
      setTasks(Array.isArray(data) ? data : []);
    } catch {
      // silently handle; TaskBoard shows empty state
    } finally {
      setTasksLoading(false);
    }
  }, [encounterId]);

  const fetchChecklist = useCallback(async () => {
    if (!encounterId) return;
    setChecklistLoading(true);
    setChecklistNotFound(false);
    try {
      const data = await dischargeService.getChecklist(encounterId);
      setChecklist(data);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setChecklistNotFound(true);
        setChecklist(null);
      }
    } finally {
      setChecklistLoading(false);
    }
  }, [encounterId]);

  useEffect(() => {
    if (activeTab === 'tasks') {
      fetchTasks();
    }
  }, [activeTab, fetchTasks]);

  useEffect(() => {
    if (activeTab === 'discharge') {
      fetchChecklist();
    }
  }, [activeTab, fetchChecklist]);

  const handleCreateTask = async (data: CreateTaskRequest) => {
    await taskService.createTask({
      ...data,
      scopeType: 'encounter',
      scopeId: encounterId ?? '',
    });
    setShowCreateTask(false);
    fetchTasks();
  };

  const handleStatusChange = async (task: Task, newStatus: import('../features/tasks/types').TaskStatus) => {
    await taskService.updateTask(task.id, { status: newStatus });
    fetchTasks();
  };

  const handleCompleteItem = async (itemId: string) => {
    setItemLoading(itemId);
    try {
      await dischargeService.completeItem(itemId);
      fetchChecklist();
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Failed to complete item');
    } finally {
      setItemLoading(null);
    }
  };

  const handleCompleteDischarge = async () => {
    if (!checklist || !encounterId) return;

    const items = checklist.items ?? [];
    const requiredItems = items.filter((i) => i.required);
    const completedRequired = requiredItems.filter((i) => i.status === 'done' || i.status === 'waived').length;
    const allRequiredComplete = completedRequired === requiredItems.length;

    const canOverride = user?.role === 'admin' || user?.role === 'charge_nurse';

    if (allRequiredComplete) {
      try {
        await dischargeService.completeDischarge(encounterId);
        fetchChecklist();
      } catch (err: any) {
        alert(err.response?.data?.error?.message || 'Failed to complete discharge');
      }
    } else if (canOverride) {
      setShowOverrideDischarge(true);
    }
  };

  if (!encounterId) {
    return (
      <Layout>
        <PageHeader title="Encounter" subtitle="Missing encounter ID" />
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
          <h2 className="text-xl font-semibold text-red-900">Missing Encounter ID</h2>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <PageHeader
        title="Encounter"
        subtitle={`ID: ${encounterId}`}
        breadcrumbs={[
          { label: 'Encounters', href: ROUTES.ENCOUNTER_LIST },
          { label: encounterId ?? '' },
        ]}
      />

      {/* Tab bar */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex gap-6">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      <div className="mt-6">
          {/* --- Care Team Tab --- */}
          {activeTab === 'care-team' && (
            <CareTeamList
              encounterId={encounterId}
              canAssign={canAssign}
              canTransfer={canTransfer}
            />
          )}

          {/* --- Flow Tab --- */}
          {activeTab === 'flow' && (
            <div className="space-y-4">
              {canUpdateFlow && (
                <div className="flex justify-end">
                  <TransitionStateButton
                    currentState={currentState}
                    encounterId={encounterId}
                    onTransition={recordTransition}
                    onOverride={canOverrideFlow ? overrideTransition : undefined}
                    canOverride={canOverrideFlow}
                  />
                </div>
              )}
              <FlowTimeline
                transitions={transitions}
                currentState={currentState}
                isLoading={flowLoading}
              />
            </div>
          )}

          {/* --- Tasks Tab --- */}
          {activeTab === 'tasks' && (
            <div className="space-y-4">
              {tasksLoading && tasks.length === 0 ? (
                <div className="flex justify-center py-8">
                  <Spinner />
                </div>
              ) : (
                <TaskBoard
                  tasks={tasks}
                  onTaskClick={() => {}}
                  onCreateTask={() => setShowCreateTask(true)}
                  onStatusChange={handleStatusChange}
                  isLoading={tasksLoading}
                />
              )}
              {showCreateTask && (
                <CreateTaskForm
                  isOpen={showCreateTask}
                  onClose={() => setShowCreateTask(false)}
                  onSubmit={handleCreateTask}
                  defaultScopeType="encounter"
                  defaultScopeId={encounterId}
                />
              )}
            </div>
          )}

          {/* --- Discharge Tab --- */}
          {activeTab === 'discharge' && (
            <div className="space-y-4">
              {checklistLoading && (
                <div className="flex justify-center py-8">
                  <Spinner />
                </div>
              )}

              {!checklistLoading && checklistNotFound && (
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
                  <p className="text-lg font-medium text-yellow-900 mb-2">No checklist found</p>
                  <p className="text-sm text-yellow-700 mb-4">
                    This encounter does not have a discharge checklist yet.
                  </p>
                  <Button variant="primary" onClick={() => setShowInitChecklist(true)}>
                    Initialize Checklist
                  </Button>
                </div>
              )}

              {!checklistLoading && checklist && (
                <div className="space-y-4">
                  {/* Header */}
                  <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900">Discharge Checklist</h3>
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
                    {(() => {
                      const items = checklist.items ?? [];
                      const requiredItems = items.filter((i) => i.required);
                      const completedRequired = requiredItems.filter(
                        (i) => i.status === 'done' || i.status === 'waived'
                      ).length;
                      const totalRequired = requiredItems.length;
                      const progressPercent =
                        totalRequired > 0 ? Math.round((completedRequired / totalRequired) * 100) : 100;

                      return (
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
                      );
                    })()}
                  </div>

                  {/* Required Items */}
                  {(() => {
                    const items = checklist.items ?? [];
                    const requiredItems = items.filter((i) => i.required);
                    return (
                      requiredItems.length > 0 && (
                        <div className="bg-white rounded-lg shadow p-6">
                          <h4 className="text-md font-semibold text-gray-900 mb-3">Required Items</h4>
                          <div className="space-y-2">
                            {requiredItems.map((item) => (
                              <div key={item.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                                <label className="flex items-center gap-3 flex-1 min-w-0 cursor-pointer">
                                  <input
                                    type="checkbox"
                                    checked={item.status === 'done' || item.status === 'waived'}
                                    disabled={item.status !== 'open' || itemLoading === item.id}
                                    onChange={() => handleCompleteItem(item.id)}
                                    aria-label={item.label}
                                    className="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                                  />
                                  <span className="text-sm font-medium text-gray-900">{item.label}</span>
                                </label>
                                <span
                                  className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${ITEM_STATUS_COLORS[item.status]}`}
                                >
                                  {item.status}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )
                    );
                  })()}

                  {/* Optional Items */}
                  {(() => {
                    const items = checklist.items ?? [];
                    const optionalItems = items.filter((i) => !i.required);
                    return (
                      optionalItems.length > 0 && (
                        <div className="bg-white rounded-lg shadow p-6">
                          <h4 className="text-md font-semibold text-gray-900 mb-3">Optional Items</h4>
                          <div className="space-y-2">
                            {optionalItems.map((item) => (
                              <div key={item.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                                <label className="flex items-center gap-3 flex-1 min-w-0 cursor-pointer">
                                  <input
                                    type="checkbox"
                                    checked={item.status === 'done' || item.status === 'waived'}
                                    disabled={item.status !== 'open' || itemLoading === item.id}
                                    onChange={() => handleCompleteItem(item.id)}
                                    aria-label={item.label}
                                    className="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                                  />
                                  <span className="text-sm font-medium text-gray-900">{item.label}</span>
                                </label>
                                <span
                                  className={`inline-flex px-2 py-0.5 rounded text-xs font-medium capitalize ${ITEM_STATUS_COLORS[item.status]}`}
                                >
                                  {item.status}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )
                    );
                  })()}

                  {/* Complete Discharge Button */}
                  {checklist.status === 'in_progress' && (() => {
                    const items = checklist.items ?? [];
                    const requiredItems = items.filter((i) => i.required);
                    const completedRequired = requiredItems.filter(
                      (i) => i.status === 'done' || i.status === 'waived'
                    ).length;
                    const allRequiredComplete = completedRequired === requiredItems.length;
                    const canOverride = user?.role === 'admin' || user?.role === 'charge_nurse';

                    return (
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
                    );
                  })()}
                </div>
              )}

              {showInitChecklist && (
                <InitChecklistModal
                  encounterId={encounterId ?? ''}
                  isOpen={showInitChecklist}
                  onClose={() => setShowInitChecklist(false)}
                  onCreated={fetchChecklist}
                />
              )}

              {showOverrideDischarge && (
                <OverrideDischargeModal
                  encounterId={encounterId ?? ''}
                  isOpen={showOverrideDischarge}
                  onClose={() => setShowOverrideDischarge(false)}
                  onCompleted={fetchChecklist}
                />
              )}
            </div>
          )}
        </div>
    </Layout>
  );
};

export default EncounterDetailPage;
