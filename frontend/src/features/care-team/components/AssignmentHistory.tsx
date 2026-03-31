/**
 * AssignmentHistory - Timeline of care team assignments and handoff notes for an encounter
 */

import { useEffect } from 'react';
import { Spinner } from '@/shared/components/ui/Spinner';
import { useCareTeam } from '../hooks/useCareTeam';
import type { CareTeamAssignment, HandoffNote } from '../types/careTeam.types';

const ROLE_LABELS: Record<string, string> = {
  primary_nurse: 'Primary Nurse',
  attending_provider: 'Attending Provider',
  consulting_provider: 'Consulting Provider',
  resident: 'Resident',
  respiratory_therapist: 'Respiratory Therapist',
  case_manager: 'Case Manager',
  social_worker: 'Social Worker',
  other: 'Other',
};

const formatDateTime = (iso: string) => {
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
};

interface AssignmentHistoryProps {
  encounterId: string;
}

const AssignmentCard = ({
  assignment,
  handoffs,
}: {
  assignment: CareTeamAssignment;
  handoffs: HandoffNote[];
}) => {
  const relatedHandoff = (handoffs ?? []).find((h) => h.assignmentId === assignment.id);
  const isActive = !assignment.endsAt;
  const roleLabel = ROLE_LABELS[assignment.roleType] ?? assignment.roleType;

  return (
    <div className="relative pl-8">
      {/* Timeline dot */}
      <div
        className={`absolute left-0 top-1.5 w-3 h-3 rounded-full border-2 ${
          isActive
            ? 'bg-green-500 border-green-600'
            : 'bg-gray-300 border-gray-400'
        }`}
      />

      <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
        <div className="flex items-start justify-between gap-2">
          <div>
            <span className="font-semibold text-gray-900 text-sm">{roleLabel}</span>
            {isActive && (
              <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                Active
              </span>
            )}
          </div>
          <span className="text-xs text-gray-500 whitespace-nowrap">
            {formatDateTime(assignment.startsAt)}
            {assignment.endsAt && ` → ${formatDateTime(assignment.endsAt)}`}
          </span>
        </div>

        <p className="text-xs text-gray-500 mt-1">User: {assignment.userId}</p>

        {relatedHandoff && (
          <div className="mt-3 pt-3 border-t border-gray-100">
            <p className="text-xs font-medium text-gray-700 mb-1">Handoff Note</p>
            <p className="text-sm text-gray-700 italic">"{relatedHandoff.note}"</p>
            {relatedHandoff.structuredFields && Object.keys(relatedHandoff.structuredFields).length > 0 && (
              <div className="mt-2 space-y-1">
                {Object.entries(relatedHandoff.structuredFields).map(([key, val]) => (
                  <div key={key} className="text-xs">
                    <span className="font-medium text-gray-600 capitalize">{key}: </span>
                    <span className="text-gray-700">{String(val)}</span>
                  </div>
                ))}
              </div>
            )}
            <p className="text-xs text-gray-400 mt-2">
              {formatDateTime(relatedHandoff.createdAt)} · from {relatedHandoff.fromUserId} → {relatedHandoff.toUserId}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export const AssignmentHistory = ({ encounterId }: AssignmentHistoryProps) => {
  const { history, handoffs, isLoading, error, fetchHistory } = useCareTeam();

  useEffect(() => {
    if (encounterId) {
      fetchHistory(encounterId);
    }
  }, [encounterId]);

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <Spinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
        {error}
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 text-sm">
        No assignment history found.
      </div>
    );
  }

  // Sort newest first
  const sorted = [...(history ?? [])].sort(
    (a, b) => new Date(b.startsAt).getTime() - new Date(a.startsAt).getTime()
  );

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
        Assignment History
      </h3>

      {/* Vertical timeline */}
      <div className="relative border-l-2 border-gray-200 ml-1.5 space-y-4">
        {sorted.map((assignment) => (
          <AssignmentCard
            key={assignment.id}
            assignment={assignment}
            handoffs={handoffs}
          />
        ))}
      </div>
    </div>
  );
};
