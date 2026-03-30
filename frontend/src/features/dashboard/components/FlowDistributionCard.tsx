/**
 * FlowDistributionCard Component
 * Displays flow state distribution with visual bar chart
 */

import { Activity } from 'lucide-react';
import type { FlowDistribution } from '../types';
import { flowDistributionToArray } from '../types';
import { FlowStateColors } from '../../flow/types';
import { Card } from '../../../shared/components/ui/Card';

interface FlowDistributionCardProps {
  distribution: FlowDistribution;
}

export const FlowDistributionCard = ({ distribution }: FlowDistributionCardProps) => {
  const data = flowDistributionToArray(distribution);
  const total = data.reduce((sum, item) => sum + item.count, 0);

  const colorClasses: Record<string, string> = {
    gray: 'bg-gray-500',
    yellow: 'bg-yellow-500',
    blue: 'bg-blue-500',
    purple: 'bg-purple-500',
    green: 'bg-green-500',
    orange: 'bg-orange-500',
    slate: 'bg-slate-500',
  };

  return (
    <Card padding="md">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Flow Distribution</h3>
          <div className="p-2 bg-purple-100 rounded-lg">
            <Activity className="w-5 h-5 text-purple-600" />
          </div>
        </div>

        <div className="space-y-3">
          {data
            .filter((item) => item.count > 0)
            .map((item) => {
              const percentage = total > 0 ? (item.count / total) * 100 : 0;
              const color = FlowStateColors[item.state];

              return (
                <div key={item.state}>
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="font-medium text-gray-700">{item.label}</span>
                    <span className="text-gray-900 font-semibold">
                      {item.count} ({Math.round(percentage)}%)
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${colorClasses[color]}`}
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              );
            })}
        </div>

        {total === 0 && (
          <div className="text-center py-8 text-gray-400 text-sm">No active encounters</div>
        )}
      </div>
    </Card>
  );
};
