/**
 * DrillDownList Component
 * Displays drill-down lists for tasks or encounters
 */

import { Clock, User, MapPin } from 'lucide-react';
import type { TaskSummary, EncounterSummary } from '../types';
import { Card } from '../../../shared/components/ui/Card';
import { format } from 'date-fns';

interface DrillDownListProps {
  title: string;
  items: TaskSummary[] | EncounterSummary[];
  type: 'task' | 'encounter';
  emptyMessage: string;
}

export const DrillDownList = ({ title, items, type, emptyMessage }: DrillDownListProps) => {
  return (
    <Card padding="md">
      <div className="space-y-3">
        <div className="flex items-center justify-between pb-2 border-b border-gray-200">
          <h3 className="font-semibold text-gray-900">{title}</h3>
          <span className="text-sm text-gray-500">{items.length}</span>
        </div>

        {items.length === 0 ? (
          <div className="text-center py-8 text-gray-400 text-sm">{emptyMessage}</div>
        ) : (
          <div className="space-y-2 max-h-[400px] overflow-y-auto">
            {type === 'task'
              ? (items as TaskSummary[]).map((task) => (
                  <TaskItem key={task.id} task={task} />
                ))
              : (items as EncounterSummary[]).map((encounter) => (
                  <EncounterItem key={encounter.id} encounter={encounter} />
                ))}
          </div>
        )}
      </div>
    </Card>
  );
};

interface TaskItemProps {
  task: TaskSummary;
}

const TaskItem = ({ task }: TaskItemProps) => {
  const priorityColors: Record<string, string> = {
    low: 'bg-gray-100 text-gray-700',
    medium: 'bg-blue-100 text-blue-700',
    high: 'bg-orange-100 text-orange-700',
    urgent: 'bg-red-100 text-red-700',
  };

  return (
    <div className="p-3 bg-gray-50 rounded-lg hover:bg-gray-100 cursor-pointer transition-colors">
      <div className="flex items-start justify-between gap-2 mb-2">
        <p className="text-sm font-medium text-gray-900 line-clamp-2">{task.title}</p>
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${priorityColors[task.priority]}`}>
          {task.priority}
        </span>
      </div>

      <div className="flex items-center gap-3 text-xs text-gray-600">
        {task.slaDueAt && (
          <span className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {format(new Date(task.slaDueAt), 'MMM d, h:mm a')}
          </span>
        )}
        {task.ownerName && (
          <span className="flex items-center gap-1">
            <User className="w-3 h-3" />
            {task.ownerName}
          </span>
        )}
      </div>
    </div>
  );
};

interface EncounterItemProps {
  encounter: EncounterSummary;
}

const EncounterItem = ({ encounter }: EncounterItemProps) => {
  return (
    <div className="p-3 bg-gray-50 rounded-lg hover:bg-gray-100 cursor-pointer transition-colors">
      <div className="flex items-start justify-between gap-2 mb-2">
        <div>
          <p className="text-sm font-medium text-gray-900">Patient {encounter.patientId}</p>
          <p className="text-xs text-gray-600">Encounter {encounter.id.slice(0, 8)}</p>
        </div>
        <span className="px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700">
          {encounter.lengthOfStay}
        </span>
      </div>

      <div className="flex items-center gap-3 text-xs text-gray-600">
        <span className="flex items-center gap-1">
          <MapPin className="w-3 h-3" />
          {encounter.unitId}
        </span>
        {encounter.currentState && (
          <span className="text-gray-500">• {encounter.currentState.replace('_', ' ')}</span>
        )}
      </div>
    </div>
  );
};
