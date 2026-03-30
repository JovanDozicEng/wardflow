/**
 * ConsultInbox - List of consults with filters
 */

import { useState } from 'react';
import { ConsultCard } from './ConsultCard';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useConsults } from '../hooks/useConsults';
import type { ConsultStatus } from '../types/consult.types';

interface ConsultInboxProps {
  onAccept: (id: string) => void;
  onDecline: (id: string) => void;
  onRedirect: (id: string) => void;
  onComplete: (id: string) => void;
  actionLoading?: boolean;
}

export const ConsultInbox = ({
  onAccept,
  onDecline,
  onRedirect,
  onComplete,
  actionLoading = false,
}: ConsultInboxProps) => {
  const [statusFilter, setStatusFilter] = useState<ConsultStatus | 'all'>('all');
  const { consults, loading, error } = useConsults({
    status: statusFilter === 'all' ? undefined : statusFilter,
  });

  const filterOptions: { value: ConsultStatus | 'all'; label: string }[] = [
    { value: 'all', label: 'All' },
    { value: 'pending', label: 'Pending' },
    { value: 'accepted', label: 'Accepted' },
    { value: 'completed', label: 'Completed' },
    { value: 'declined', label: 'Declined' },
    { value: 'redirected', label: 'Redirected' },
  ];

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-800">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex gap-2 overflow-x-auto pb-2">
        {filterOptions.map((option) => (
          <button
            key={option.value}
            onClick={() => setStatusFilter(option.value)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
              statusFilter === option.value
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {option.label}
          </button>
        ))}
      </div>

      {/* Consults List */}
      {consults.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">No consults found</p>
          {statusFilter !== 'all' && (
            <button
              onClick={() => setStatusFilter('all')}
              className="text-blue-600 hover:text-blue-700 mt-2 text-sm"
            >
              Clear filters
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {consults.map((consult) => (
            <ConsultCard
              key={consult.id}
              consult={consult}
              onAccept={onAccept}
              onDecline={onDecline}
              onRedirect={onRedirect}
              onComplete={onComplete}
              loading={actionLoading}
            />
          ))}
        </div>
      )}
    </div>
  );
};
