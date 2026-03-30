/**
 * ExceptionList - List exceptions with status badges
 */

import { useState } from 'react';
import { Badge } from '@/shared/components/ui/Badge';
import { Card } from '@/shared/components/ui/Card';
import { Button } from '@/shared/components/ui/Button';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useExceptions } from '../hooks/useExceptions';
import { usePermissions } from '@/features/auth/hooks/usePermissions';
import type { ExceptionEvent, ExceptionStatus } from '../types/exception.types';
import { formatDistanceToNow } from 'date-fns';

interface ExceptionListProps {
  onEdit: (exception: ExceptionEvent) => void;
  onFinalize: (exception: ExceptionEvent) => void;
  onCorrect: (exception: ExceptionEvent) => void;
  actionLoading?: boolean;
}

const getStatusColor = (status: ExceptionStatus) => {
  switch (status) {
    case 'draft':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'finalized':
      return 'bg-green-100 text-green-800 border-green-200';
    case 'corrected':
      return 'bg-blue-100 text-blue-800 border-blue-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

export const ExceptionList = ({
  onEdit,
  onFinalize,
  onCorrect,
  actionLoading = false,
}: ExceptionListProps) => {
  const [statusFilter, setStatusFilter] = useState<ExceptionStatus | 'all'>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  
  const { exceptions, loading, error } = useExceptions({
    status: statusFilter === 'all' ? undefined : statusFilter,
    type: typeFilter === 'all' ? undefined : typeFilter,
  });
  
  const { hasAnyRole } = usePermissions();
  const canCorrectExceptions = hasAnyRole(['quality_safety', 'admin']);

  // Get unique types for filter
  const types = Array.from(new Set(exceptions.map((e) => e.type)));

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
      <div className="space-y-3">
        <div className="flex gap-2 overflow-x-auto pb-2">
          <span className="text-sm font-medium text-gray-700 py-2">Status:</span>
          {(['all', 'draft', 'finalized', 'corrected'] as const).map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                statusFilter === status
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {status === 'all' ? 'All' : status.charAt(0).toUpperCase() + status.slice(1)}
            </button>
          ))}
        </div>
        
        {types.length > 0 && (
          <div className="flex gap-2 overflow-x-auto pb-2">
            <span className="text-sm font-medium text-gray-700 py-2">Type:</span>
            <button
              onClick={() => setTypeFilter('all')}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                typeFilter === 'all'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              All
            </button>
            {types.map((type) => (
              <button
                key={type}
                onClick={() => setTypeFilter(type)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                  typeFilter === type
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {type}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Exceptions List */}
      {exceptions.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">No exceptions found</p>
        </div>
      ) : (
        <div className="space-y-3">
          {exceptions.map((exception) => (
            <Card key={exception.id} className="p-4 hover:shadow-md transition-shadow">
              <div className="space-y-3">
                {/* Header */}
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-semibold text-gray-900">{exception.type}</h3>
                      <Badge variant="default" className={getStatusColor(exception.status)}>
                        {exception.status.toUpperCase()}
                      </Badge>
                    </div>
                    <p className="text-sm text-gray-500 mt-1">
                      Encounter: {exception.encounterId}
                    </p>
                  </div>
                </div>

                {/* Data preview */}
                <div className="text-sm">
                  <p className="font-medium text-gray-900 mb-1">Exception Data:</p>
                  <div className="bg-gray-50 rounded p-2 text-xs font-mono overflow-x-auto">
                    <pre>{JSON.stringify(exception.data, null, 2)}</pre>
                  </div>
                </div>

                {/* Metadata */}
                <div className="text-xs text-gray-500 space-y-1">
                  <div>
                    <span className="font-medium">Initiated by:</span> {exception.initiatedBy}
                    {' · '}
                    {formatDistanceToNow(new Date(exception.initiatedAt), { addSuffix: true })}
                  </div>
                  
                  {exception.finalizedBy && (
                    <div>
                      <span className="font-medium">Finalized by:</span> {exception.finalizedBy}
                      {exception.finalizedAt && (
                        <span> · {formatDistanceToNow(new Date(exception.finalizedAt), { addSuffix: true })}</span>
                      )}
                    </div>
                  )}
                  
                  {exception.correctedByEventId && (
                    <div>
                      <span className="font-medium">Corrected by event:</span> {exception.correctedByEventId}
                      {exception.correctionReason && (
                        <span> · Reason: {exception.correctionReason}</span>
                      )}
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="flex gap-2 pt-2 border-t border-gray-200">
                  {exception.status === 'draft' && (
                    <>
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => onEdit(exception)}
                        disabled={actionLoading}
                        className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700"
                      >
                        Edit
                      </Button>
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => onFinalize(exception)}
                        disabled={actionLoading}
                        className="px-3 py-1.5 text-sm bg-green-600 hover:bg-green-700 text-white"
                      >
                        Finalize
                      </Button>
                    </>
                  )}
                  
                  {exception.status === 'finalized' && canCorrectExceptions && (
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() => onCorrect(exception)}
                      disabled={actionLoading}
                      className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white"
                    >
                      Correct
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};
