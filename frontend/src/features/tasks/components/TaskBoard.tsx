/**
 * TaskBoard Component
 * Kanban-style task board with columns for different statuses
 * Features: Filtering, SLA indicators, drag-and-drop (future)
 */

import { useState, useMemo } from 'react';
import { Clock, User, AlertCircle, Plus, Filter } from 'lucide-react';
import { format } from 'date-fns';
import type { Task, TaskStatus, TaskPriority, ScopeType, TaskFilterParams } from '../types';
import {
  TaskStatusLabels,
  TaskPriorityLabels,
  TaskPriorityColors,
  isTaskOverdue,
} from '../types';
import { Card } from '../../../shared/components/ui/Card';
import { Badge } from '../../../shared/components/ui/Badge';
import { Button } from '../../../shared/components/ui/Button';

interface TaskBoardProps {
  tasks: Task[];
  onTaskClick: (task: Task) => void;
  onCreateTask: () => void;
  onFilterChange?: (filters: TaskFilterParams) => void;
  isLoading?: boolean;
}

export const TaskBoard = ({
  tasks,
  onTaskClick,
  onCreateTask,
  onFilterChange,
  isLoading,
}: TaskBoardProps) => {
  const [filters, setFilters] = useState<TaskFilterParams>({
    overdue: undefined,
    priority: undefined,
    scopeType: undefined,
  });
  const [showFilters, setShowFilters] = useState(false);

  // Group tasks by status
  const tasksByStatus = useMemo(() => {
    const grouped: Record<TaskStatus, Task[]> = {
      open: [],
      in_progress: [],
      completed: [],
      cancelled: [],
      escalated: [],
    };

    tasks.forEach((task) => {
      grouped[task.status].push(task);
    });

    return grouped;
  }, [tasks]);

  const columns: Array<{ status: TaskStatus; color: string }> = [
    { status: 'open', color: 'blue' },
    { status: 'in_progress', color: 'yellow' },
    { status: 'escalated', color: 'red' },
    { status: 'completed', color: 'green' },
  ];

  const handleFilterChange = (newFilters: Partial<TaskFilterParams>) => {
    const updated = { ...filters, ...newFilters };
    setFilters(updated);
    onFilterChange?.(updated);
  };

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} padding="md">
            <div className="animate-pulse space-y-3">
              <div className="h-4 bg-gray-200 rounded w-1/2" />
              <div className="h-20 bg-gray-200 rounded" />
              <div className="h-20 bg-gray-200 rounded" />
            </div>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header with Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h2 className="text-xl font-semibold">Task Board</h2>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2"
          >
            <Filter className="w-4 h-4" />
            Filters
          </Button>
        </div>
        <Button onClick={onCreateTask} className="flex items-center gap-2">
          <Plus className="w-4 h-4" />
          New Task
        </Button>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <Card padding="md">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Priority
              </label>
              <select
                value={filters.priority || ''}
                onChange={(e) =>
                  handleFilterChange({
                    priority: e.target.value ? (e.target.value as TaskPriority) : undefined,
                  })
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
              >
                <option value="">All Priorities</option>
                <option value="urgent">Urgent</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Scope</label>
              <select
                value={filters.scopeType || ''}
                onChange={(e) =>
                  handleFilterChange({
                    scopeType: e.target.value ? (e.target.value as ScopeType) : undefined,
                  })
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
              >
                <option value="">All Scopes</option>
                <option value="encounter">Encounter</option>
                <option value="patient">Patient</option>
                <option value="unit">Unit</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Status</label>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="overdue-filter"
                  checked={filters.overdue || false}
                  onChange={(e) => handleFilterChange({ overdue: e.target.checked || undefined })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <label htmlFor="overdue-filter" className="text-sm text-gray-700">
                  Show only overdue
                </label>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Kanban Board */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {columns.map(({ status, color }) => (
          <TaskColumn
            key={status}
            status={status}
            tasks={tasksByStatus[status]}
            color={color}
            onTaskClick={onTaskClick}
          />
        ))}
      </div>
    </div>
  );
};

interface TaskColumnProps {
  status: TaskStatus;
  tasks: Task[];
  color: string;
  onTaskClick: (task: Task) => void;
}

const TaskColumn = ({ status, tasks, color, onTaskClick }: TaskColumnProps) => {
  const colorClasses: Record<string, string> = {
    blue: 'bg-blue-50 border-blue-200',
    yellow: 'bg-yellow-50 border-yellow-200',
    red: 'bg-red-50 border-red-200',
    green: 'bg-green-50 border-green-200',
  };

  return (
    <div className="space-y-3">
      {/* Column Header */}
      <div className={`p-3 rounded-lg border-2 ${colorClasses[color]}`}>
        <div className="flex items-center justify-between">
          <h3 className="font-semibold">{TaskStatusLabels[status]}</h3>
          <Badge variant="default" className="bg-white">
            {tasks.length}
          </Badge>
        </div>
      </div>

      {/* Tasks */}
      <div className="space-y-2">
        {tasks.length === 0 ? (
          <div className="text-center py-8 text-gray-400 text-sm">No tasks</div>
        ) : (
          tasks.map((task) => <TaskCard key={task.id} task={task} onClick={onTaskClick} />)
        )}
      </div>
    </div>
  );
};

interface TaskCardProps {
  task: Task;
  onClick: (task: Task) => void;
}

const TaskCard = ({ task, onClick }: TaskCardProps) => {
  const overdue = isTaskOverdue(task);
  const priorityColor = TaskPriorityColors[task.priority];

  const priorityColorClasses: Record<string, string> = {
    gray: 'bg-gray-100 text-gray-800',
    blue: 'bg-blue-100 text-blue-800',
    orange: 'bg-orange-100 text-orange-800',
    red: 'bg-red-100 text-red-800',
  };

  return (
    <Card
      padding="none"
      hover
      className="cursor-pointer hover:shadow-md transition-shadow"
      onClick={() => onClick(task)}
    >
      <div className="p-3 space-y-2">
        {/* Priority Badge */}
        <div className="flex items-start justify-between gap-2">
          <span
            className={`px-2 py-0.5 rounded text-xs font-medium ${priorityColorClasses[priorityColor]}`}
          >
            {TaskPriorityLabels[task.priority]}
          </span>
          {overdue && (
            <Badge variant="danger" className="bg-red-100 text-red-800 flex items-center gap-1">
              <AlertCircle className="w-3 h-3" />
              Overdue
            </Badge>
          )}
        </div>

        {/* Title */}
        <h4 className="font-medium text-sm line-clamp-2">{task.title}</h4>

        {/* Details */}
        {task.details && (
          <p className="text-xs text-gray-600 line-clamp-2">{task.details}</p>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between text-xs text-gray-500 pt-2 border-t border-gray-100">
          {task.slaDueAt && (
            <div className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {format(new Date(task.slaDueAt), 'MMM d, h:mm a')}
            </div>
          )}
          {task.ownerName && (
            <div className="flex items-center gap-1">
              <User className="w-3 h-3" />
              <span className="truncate max-w-[100px]">{task.ownerName}</span>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
};
