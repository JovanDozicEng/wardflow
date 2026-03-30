/**
 * TaskMetricsCard Component
 * Displays task board overview metrics
 */

import { CheckSquare, AlertCircle, TrendingUp } from 'lucide-react';
import type { TaskMetrics } from '../types';
import { Card } from '../../../shared/components/ui/Card';

interface TaskMetricsCardProps {
  metrics: TaskMetrics;
}

export const TaskMetricsCard = ({ metrics }: TaskMetricsCardProps) => {
  const criticalCount = metrics.urgent + metrics.totalOverdue;

  return (
    <Card padding="md">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Task Metrics</h3>
          <div className="p-2 bg-green-100 rounded-lg">
            <CheckSquare className="w-5 h-5 text-green-600" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-3">
            <div>
              <p className="text-sm text-gray-600">Total Open</p>
              <p className="text-2xl font-bold text-gray-900">{metrics.totalOpen}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600 flex items-center gap-1">
                <AlertCircle className="w-4 h-4 text-red-500" />
                Overdue
              </p>
              <p className="text-2xl font-bold text-red-600">{metrics.totalOverdue}</p>
            </div>
          </div>

          <div className="space-y-3">
            <div>
              <p className="text-sm text-gray-600">High Priority</p>
              <p className="text-2xl font-bold text-orange-600">{metrics.highPriority}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Urgent</p>
              <p className="text-2xl font-bold text-red-600">{metrics.urgent}</p>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 pt-3 border-t border-gray-200">
          <div>
            <p className="text-sm text-gray-600">Unassigned</p>
            <p className="text-lg font-semibold text-gray-900">{metrics.unassigned}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600 flex items-center gap-1">
              <TrendingUp className="w-4 h-4 text-green-500" />
              Completed Today
            </p>
            <p className="text-lg font-semibold text-green-600">{metrics.completedToday}</p>
          </div>
        </div>

        {criticalCount > 0 && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium">
              ⚠️ {criticalCount} tasks need immediate attention
            </p>
          </div>
        )}
      </div>
    </Card>
  );
};
