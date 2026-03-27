/**
 * ConsultCard - Display single consult with actions
 */

import { Badge } from '@/shared/components/ui/Badge';
import { Button } from '@/shared/components/ui/Button';
import { Card } from '@/shared/components/ui/Card';
import { usePermissions } from '@/features/auth/hooks/usePermissions';
import type { ConsultRequest, ConsultUrgency, ConsultStatus } from '../types/consult.types';
import { formatDistanceToNow } from 'date-fns';

interface ConsultCardProps {
  consult: ConsultRequest;
  onAccept: (id: string) => void;
  onDecline: (id: string) => void;
  onRedirect: (id: string) => void;
  onComplete: (id: string) => void;
  loading?: boolean;
}

const getUrgencyColor = (urgency: ConsultUrgency) => {
  switch (urgency) {
    case 'emergent':
      return 'bg-red-100 text-red-800 border-red-200';
    case 'urgent':
      return 'bg-orange-100 text-orange-800 border-orange-200';
    case 'routine':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

const getStatusColor = (status: ConsultStatus) => {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'accepted':
      return 'bg-blue-100 text-blue-800 border-blue-200';
    case 'completed':
      return 'bg-green-100 text-green-800 border-green-200';
    case 'declined':
      return 'bg-red-100 text-red-800 border-red-200';
    case 'redirected':
      return 'bg-purple-100 text-purple-800 border-purple-200';
    case 'cancelled':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

export const ConsultCard = ({
  consult,
  onAccept,
  onDecline,
  onRedirect,
  onComplete,
  loading = false,
}: ConsultCardProps) => {
  const { hasAnyRole } = usePermissions();
  const canManageConsults = hasAnyRole(['provider', 'consult', 'admin']);

  const showAcceptDeclineButtons = consult.status === 'pending' && canManageConsults;
  const showCompleteButton = consult.status === 'accepted' && canManageConsults;
  const showRedirectButton = (consult.status === 'pending' || consult.status === 'accepted') && canManageConsults;

  return (
    <Card className="p-4 hover:shadow-md transition-shadow">
      <div className="space-y-3">
        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <h3 className="font-semibold text-gray-900 truncate">
                {consult.targetService}
              </h3>
              <Badge variant="default" className={getUrgencyColor(consult.urgency)}>
                {consult.urgency.toUpperCase()}
              </Badge>
              <Badge variant="default" className={getStatusColor(consult.status)}>
                {consult.status.toUpperCase()}
              </Badge>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              Encounter: {consult.encounterId}
            </p>
          </div>
        </div>

        {/* Reason */}
        <div className="text-sm text-gray-700">
          <p className="font-medium text-gray-900 mb-1">Reason:</p>
          <p>{consult.reason}</p>
        </div>

        {/* Metadata */}
        <div className="text-xs text-gray-500 space-y-1">
          <div>
            <span className="font-medium">Created by:</span> {consult.createdBy}
            {' · '}
            {formatDistanceToNow(new Date(consult.createdAt), { addSuffix: true })}
          </div>
          
          {consult.acceptedBy && (
            <div>
              <span className="font-medium">Accepted by:</span> {consult.acceptedBy}
              {consult.acceptedAt && (
                <span> · {formatDistanceToNow(new Date(consult.acceptedAt), { addSuffix: true })}</span>
              )}
            </div>
          )}
          
          {consult.closeReason && (
            <div>
              <span className="font-medium">Close reason:</span> {consult.closeReason}
            </div>
          )}
          
          {consult.redirectedTo && (
            <div>
              <span className="font-medium">Redirected to:</span> {consult.redirectedTo}
            </div>
          )}
        </div>

        {/* Actions */}
        {(showAcceptDeclineButtons || showCompleteButton || showRedirectButton) && (
          <div className="flex gap-2 pt-2 border-t border-gray-200">
            {showAcceptDeclineButtons && (
              <>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => onAccept(consult.id)}
                  disabled={loading}
                  className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white"
                >
                  Accept
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() => onDecline(consult.id)}
                  disabled={loading}
                  className="px-3 py-1.5 text-sm bg-red-600 hover:bg-red-700 text-white"
                >
                  Decline
                </Button>
              </>
            )}
            
            {showCompleteButton && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => onComplete(consult.id)}
                disabled={loading}
                className="px-3 py-1.5 text-sm bg-green-600 hover:bg-green-700 text-white"
              >
                Complete
              </Button>
            )}
            
            {showRedirectButton && (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => onRedirect(consult.id)}
                disabled={loading}
                className="px-3 py-1.5 text-sm bg-purple-600 hover:bg-purple-700 text-white"
              >
                Redirect
              </Button>
            )}
          </div>
        )}
      </div>
    </Card>
  );
};
