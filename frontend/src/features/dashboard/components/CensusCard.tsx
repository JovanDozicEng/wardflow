/**
 * CensusCard Component
 * Displays census metrics (active encounters, expected discharges)
 */

import { Users, ArrowDown } from 'lucide-react';
import type { CensusMetrics } from '../types';
import { Card } from '../../../shared/components/ui/Card';

interface CensusCardProps {
  census: CensusMetrics;
}

export const CensusCard = ({ census }: CensusCardProps) => {
  return (
    <Card padding="md">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Census</h3>
          <div className="p-2 bg-blue-100 rounded-lg">
            <Users className="w-5 h-5 text-blue-600" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-gray-600">Active Encounters</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">{census.active}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600 flex items-center gap-1">
              <ArrowDown className="w-4 h-4" />
              Expected Discharges
            </p>
            <p className="text-3xl font-bold text-orange-600 mt-1">{census.expectedDischarges}</p>
          </div>
        </div>

        {census.expectedDischarges > 0 && (
          <div className="pt-3 border-t border-gray-200">
            <p className="text-xs text-gray-500">
              {Math.round((census.expectedDischarges / census.active) * 100)}% turnover expected
              today
            </p>
          </div>
        )}
      </div>
    </Card>
  );
};
