/**
 * IncidentList - List incidents with filters
 */

import { useState } from 'react';
import { Badge } from '@/shared/components/ui/Badge';
import { Card } from '@/shared/components/ui/Card';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useIncidents } from '../hooks/useIncidents';
import type { Incident, IncidentStatus } from '../types/incident.types';
import { formatDistanceToNow } from 'date-fns';

interface IncidentListProps {
  onSelectIncident: (incident: Incident) => void;
}

const getStatusColor = (status: IncidentStatus) => {
  switch (status) {
    case 'submitted':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'under_review':
      return 'bg-blue-100 text-blue-800 border-blue-200';
    case 'closed':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

const getSeverityColor = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return 'bg-red-100 text-red-800 border-red-200';
    case 'severe':
      return 'bg-orange-100 text-orange-800 border-orange-200';
    case 'moderate':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'minor':
      return 'bg-green-100 text-green-800 border-green-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

export const IncidentList = ({ onSelectIncident }: IncidentListProps) => {
  const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'all'>('all');
  
  const { incidents, loading, error } = useIncidents({
    status: statusFilter === 'all' ? undefined : statusFilter,
  });

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
        {(['all', 'submitted', 'under_review', 'closed'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setStatusFilter(status)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
              statusFilter === status
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {status === 'all' ? 'All' : status.replace('_', ' ').charAt(0).toUpperCase() + status.slice(1).replace('_', ' ')}
          </button>
        ))}
      </div>

      {/* Incidents List */}
      {incidents.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">No incidents found</p>
        </div>
      ) : (
        <div className="space-y-3">
          {incidents.map((incident) => (
            <Card
              key={incident.id}
              className="p-4 hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => onSelectIncident(incident)}
            >
              <div className="space-y-3">
                {/* Header */}
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-semibold text-gray-900">{incident.type}</h3>
                      <Badge variant="default" className={getStatusColor(incident.status)}>
                        {incident.status.replace('_', ' ').toUpperCase()}
                      </Badge>
                      {incident.severity && (
                        <Badge variant="default" className={getSeverityColor(incident.severity)}>
                          {incident.severity.toUpperCase()}
                        </Badge>
                      )}
                    </div>
                    {incident.encounterId && (
                      <p className="text-sm text-gray-500 mt-1">
                        Encounter: {incident.encounterId}
                      </p>
                    )}
                  </div>
                </div>

                {/* Metadata */}
                <div className="text-xs text-gray-500 space-y-1">
                  <div>
                    <span className="font-medium">Reported by:</span> {incident.reportedBy}
                    {' · '}
                    {formatDistanceToNow(new Date(incident.reportedAt), { addSuffix: true })}
                  </div>
                  <div>
                    <span className="font-medium">Event time:</span>{' '}
                    {new Date(incident.eventTime).toLocaleString()}
                  </div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};
