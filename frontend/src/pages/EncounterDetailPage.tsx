/**
 * EncounterDetailPage - Tabbed encounter detail view
 * Tabs: Care Team | Flow | Tasks
 */

import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';

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

import { Permission } from '../shared/types';
import { ROUTES } from '../shared/config/routes';

type Tab = 'care-team' | 'flow' | 'tasks';

const TABS: { id: Tab; label: string }[] = [
  { id: 'care-team', label: 'Care Team' },
  { id: 'flow', label: 'Patient Flow' },
  { id: 'tasks', label: 'Tasks' },
];

export const EncounterDetailPage = () => {
  const { id: encounterId } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState<Tab>('care-team');
  const { hasPermission } = usePermissions();

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

  useEffect(() => {
    if (activeTab === 'tasks') {
      fetchTasks();
    }
  }, [activeTab, fetchTasks]);

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

  if (!encounterId) {
    return (
      <Layout>
        <div className="max-w-5xl mx-auto px-4 py-8">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <h2 className="text-xl font-semibold text-red-900">Missing Encounter ID</h2>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="max-w-5xl mx-auto px-4 py-8">
        {/* Breadcrumb */}
        <div className="mb-4">
          <Link to={ROUTES.ENCOUNTER_LIST} className="text-sm text-blue-600 hover:underline">
            ← Back to Encounters
          </Link>
        </div>

        <PageHeader
          title="Encounter"
          subtitle={`ID: ${encounterId}`}
        />

        {/* Tab bar */}
        <div className="mt-6 border-b border-gray-200">
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
        </div>
      </div>
    </Layout>
  );
};

export default EncounterDetailPage;
