/**
 * IncidentReviewQueue - For quality_safety role
 */

import { Badge } from '@/shared/components/ui/Badge';
import { Card } from '@/shared/components/ui/Card';
import { Button } from '@/shared/components/ui/Button';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useIncidents } from '../hooks/useIncidents';
import type { Incident } from '../types/incident.types';
import { formatDistanceToNow } from 'date-fns';

interface IncidentReviewQueueProps {
  onReview: (incident: Incident) => void;
}

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

export const IncidentReviewQueue = ({ onReview }: IncidentReviewQueueProps) => {
  // Show only submitted incidents (pending review)
  const { incidents, loading, error } = useIncidents({ status: 'submitted' });

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
      {incidents.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
          <p className="text-gray-600">No incidents pending review</p>
        </div>
      ) : (
        <>
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <p className="text-yellow-900 font-medium">
              {incidents.length} incident{incidents.length !== 1 ? 's' : ''} awaiting review
            </p>
          </div>

          <div className="space-y-3">
            {incidents.map((incident) => (
              <Card key={incident.id} className="p-4 hover:shadow-md transition-shadow">
                <div className="space-y-3">
                  {/* Header */}
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="font-semibold text-gray-900">{incident.type}</h3>
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

                  {/* Harm Indicators */}
                  {incident.harmIndicators && Object.keys(incident.harmIndicators).length > 0 && (
                    <div className="text-sm">
                      <p className="font-medium text-gray-900 mb-1">Harm Indicators:</p>
                      <div className="bg-gray-50 rounded p-2 text-xs font-mono overflow-x-auto">
                        <pre>{JSON.stringify(incident.harmIndicators, null, 2)}</pre>
                      </div>
                    </div>
                  )}

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

                  {/* Actions */}
                  <div className="flex gap-2 pt-2 border-t border-gray-200">
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() => onReview(incident)}
                      className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white"
                    >
                      Review Incident
                    </Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  );
};
