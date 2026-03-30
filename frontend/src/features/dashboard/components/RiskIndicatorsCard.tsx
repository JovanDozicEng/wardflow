/**
 * RiskIndicatorsCard Component
 * Displays risk indicators with alert styling
 */

import { AlertTriangle, Clock, Bed, ListChecks, UserX, Users } from 'lucide-react';
import type { RiskIndicators } from '../types';
import { getHighRiskIndicators } from '../types';
import { Card } from '../../../shared/components/ui/Card';

interface RiskIndicatorsCardProps {
  indicators: RiskIndicators;
}

export const RiskIndicatorsCard = ({ indicators }: RiskIndicatorsCardProps) => {
  const highRiskItems = getHighRiskIndicators(indicators);
  const hasHighRisk = highRiskItems.length > 0;

  const iconMap: Record<string, React.ReactNode> = {
    triageOver2hrs: <Clock className="w-4 h-4" />,
    waitingForBed: <Bed className="w-4 h-4" />,
    overdueHighPriority: <ListChecks className="w-4 h-4" />,
    unassignedUrgent: <UserX className="w-4 h-4" />,
    noCareTeam: <Users className="w-4 h-4" />,
  };

  const allIndicators = [
    {
      key: 'triageOver2hrs',
      label: 'Triage >2hrs',
      value: indicators.patientsInTriageOver2hrs,
      threshold: 3,
    },
    {
      key: 'waitingForBed',
      label: 'Waiting for Bed >4hrs',
      value: indicators.patientsWaitingForBedOver4hrs,
      threshold: 2,
    },
    {
      key: 'overdueHighPriority',
      label: 'Overdue High Priority',
      value: indicators.overdueHighPriorityTasks,
      threshold: 5,
    },
    {
      key: 'unassignedUrgent',
      label: 'Unassigned Urgent',
      value: indicators.unassignedUrgentTasks,
      threshold: 1,
    },
    {
      key: 'noCareTeam',
      label: 'No Care Team',
      value: indicators.encountersWithoutCareTeam,
      threshold: 0,
    },
  ];

  return (
    <Card padding="md">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Risk Indicators</h3>
          <div
            className={`p-2 rounded-lg ${hasHighRisk ? 'bg-red-100' : 'bg-gray-100'}`}
          >
            <AlertTriangle
              className={`w-5 h-5 ${hasHighRisk ? 'text-red-600' : 'text-gray-400'}`}
            />
          </div>
        </div>

        {hasHighRisk ? (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium mb-2">
              ⚠️ {highRiskItems.length} risk{highRiskItems.length > 1 ? 's' : ''} above threshold
            </p>
            <div className="space-y-2">
              {highRiskItems.map((item) => (
                <div
                  key={item.key}
                  className="flex items-center justify-between text-sm text-red-700"
                >
                  <span className="flex items-center gap-2">
                    {iconMap[item.key]}
                    {item.label}
                  </span>
                  <span className="font-semibold">
                    {item.value} (&gt;{item.threshold})
                  </span>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
            <p className="text-sm text-green-800 font-medium">
              ✓ All indicators within normal range
            </p>
          </div>
        )}

        <div className="space-y-2">
          {allIndicators.map((item) => {
            const isAboveThreshold = item.value > item.threshold;

            return (
              <div
                key={item.key}
                className={`flex items-center justify-between text-sm ${
                  isAboveThreshold ? 'font-semibold' : ''
                }`}
              >
                <span className="flex items-center gap-2 text-gray-700">
                  {iconMap[item.key]}
                  {item.label}
                </span>
                <span className={isAboveThreshold ? 'text-red-600' : 'text-gray-900'}>
                  {item.value}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </Card>
  );
};
