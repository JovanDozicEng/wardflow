/**
 * IncidentDetail - View incident with status history
 */

import { useEffect } from 'react';
import { Badge } from '@/shared/components/ui/Badge';
import { Card } from '@/shared/components/ui/Card';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useIncidentStore } from '../store/incidentStore';
import { useIncidentActions } from '../hooks/useIncidentActions';
import type { Incident, IncidentStatus } from '../types/incident.types';
import { formatDistanceToNow } from 'date-fns';

interface IncidentDetailProps {
  incident: Incident;
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

export const IncidentDetail = ({ incident }: IncidentDetailProps) => {
  const { statusHistory } = useIncidentStore();
  const { fetchStatusHistory, loading } = useIncidentActions();

  useEffect(() => {
    fetchStatusHistory(incident.id);
  }, [incident.id]);

  return (
    <div className="space-y-6">
      {/* Main Info */}
      <Card className="p-6">
        <div className="space-y-4">
          <div className="flex items-start justify-between">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">{incident.type}</h2>
              <p className="text-sm text-gray-500 mt-1">ID: {incident.id}</p>
            </div>
            <div className="flex gap-2">
              <Badge variant="default" className={getStatusColor(incident.status)}>
                {incident.status.replace('_', ' ').toUpperCase()}
              </Badge>
              {incident.severity && (
                <Badge variant="default" className={getSeverityColor(incident.severity)}>
                  {incident.severity.toUpperCase()}
                </Badge>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            {incident.encounterId && (
              <div>
                <span className="font-medium text-gray-700">Encounter ID:</span>
                <p className="text-gray-900">{incident.encounterId}</p>
              </div>
            )}
            <div>
              <span className="font-medium text-gray-700">Event Time:</span>
              <p className="text-gray-900">{new Date(incident.eventTime).toLocaleString()}</p>
            </div>
            <div>
              <span className="font-medium text-gray-700">Reported By:</span>
              <p className="text-gray-900">{incident.reportedBy}</p>
            </div>
            <div>
              <span className="font-medium text-gray-700">Reported At:</span>
              <p className="text-gray-900">
                {new Date(incident.reportedAt).toLocaleString()}
                {' · '}
                {formatDistanceToNow(new Date(incident.reportedAt), { addSuffix: true })}
              </p>
            </div>
          </div>

          {incident.harmIndicators && Object.keys(incident.harmIndicators).length > 0 && (
            <div>
              <p className="font-medium text-gray-700 mb-2">Harm Indicators:</p>
              <div className="bg-gray-50 rounded p-3 text-sm font-mono overflow-x-auto">
                <pre>{JSON.stringify(incident.harmIndicators, null, 2)}</pre>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Status History */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Status History</h3>
        
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Spinner size="md" />
          </div>
        ) : statusHistory.length === 0 ? (
          <p className="text-gray-500 text-sm">No status changes yet</p>
        ) : (
          <div className="space-y-4">
            {statusHistory.map((event, index) => (
              <div
                key={event.id}
                className={`flex gap-4 ${index !== statusHistory.length - 1 ? 'pb-4 border-b border-gray-200' : ''}`}
              >
                {/* Timeline indicator */}
                <div className="flex flex-col items-center">
                  <div className={`w-3 h-3 rounded-full ${index === 0 ? 'bg-blue-600' : 'bg-gray-400'}`} />
                  {index !== statusHistory.length - 1 && (
                    <div className="w-0.5 h-full bg-gray-300 mt-1" />
                  )}
                </div>

                {/* Event details */}
                <div className="flex-1 pt-0.5">
                  <div className="flex items-center gap-2 mb-1">
                    {event.fromStatus && (
                      <>
                        <Badge variant="default" className={getStatusColor(event.fromStatus)}>
                          {event.fromStatus.replace('_', ' ').toUpperCase()}
                        </Badge>
                        <span className="text-gray-400">→</span>
                      </>
                    )}
                    <Badge variant="default" className={getStatusColor(event.toStatus)}>
                      {event.toStatus.replace('_', ' ').toUpperCase()}
                    </Badge>
                  </div>
                  
                  <p className="text-sm text-gray-700">
                    Changed by <span className="font-medium">{event.changedBy}</span>
                  </p>
                  
                  <p className="text-xs text-gray-500">
                    {new Date(event.changedAt).toLocaleString()}
                    {' · '}
                    {formatDistanceToNow(new Date(event.changedAt), { addSuffix: true })}
                  </p>
                  
                  {event.note && (
                    <p className="text-sm text-gray-600 mt-2 bg-gray-50 rounded p-2">
                      <span className="font-medium">Note:</span> {event.note}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
};
