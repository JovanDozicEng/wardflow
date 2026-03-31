/**
 * CareTeamList - List of active care team members for an encounter
 */

import { useEffect, useState } from 'react';
import { Spinner } from '@/shared/components/ui/Spinner';
import { Button } from '@/shared/components/ui/Button';
import { useCareTeam } from '../hooks/useCareTeam';
import { CareTeamMember } from './CareTeamMember';
import { AssignmentForm } from './AssignmentForm';
import { HandoffForm } from './HandoffForm';
import { AssignmentHistory } from './AssignmentHistory';
import type { CareTeamAssignment } from '../types/careTeam.types';

interface CareTeamListProps {
  encounterId: string;
  canAssign?: boolean;
  canTransfer?: boolean;
}

export const CareTeamList = ({
  encounterId,
  canAssign = false,
  canTransfer = false,
}: CareTeamListProps) => {
  const { assignments, isLoading, error, fetchByEncounter, assign, transfer } = useCareTeam();
  const [showAssignForm, setShowAssignForm] = useState(false);
  const [handoffTarget, setHandoffTarget] = useState<CareTeamAssignment | null>(null);
  const [showHistory, setShowHistory] = useState(false);

  useEffect(() => {
    if (encounterId) {
      fetchByEncounter(encounterId);
    }
  }, [encounterId]);

  return (
    <div className="space-y-3">
      {/* Header — always visible */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
          Care Team ({assignments.length})
        </h3>
        <div className="flex gap-2">
          <button
            onClick={() => setShowHistory((v) => !v)}
            className="text-xs text-blue-600 hover:underline"
          >
            {showHistory ? 'Hide History' : 'View History'}
          </button>
          {canAssign && (
            <Button
              variant="primary"
              size="sm"
              onClick={() => setShowAssignForm(true)}
              className="text-xs px-2 py-1"
            >
              + Assign Member
            </Button>
          )}
        </div>
      </div>

      {/* Body */}
      {isLoading && assignments.length === 0 ? (
        <div className="flex justify-center py-6">
          <Spinner />
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
          {error}
        </div>
      ) : assignments.length === 0 ? (
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-6 text-center">
          <p className="text-sm text-gray-500">No active care team members.</p>
          {canAssign && (
            <button
              onClick={() => setShowAssignForm(true)}
              className="mt-2 text-sm text-blue-600 hover:underline"
            >
              Assign the first member →
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-2">
          {assignments.map((a) => (
            <CareTeamMember
              key={a.id}
              assignment={a}
              canTransfer={canTransfer}
              onHandoff={setHandoffTarget}
            />
          ))}
        </div>
      )}

      {/* History panel */}
      {showHistory && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <AssignmentHistory encounterId={encounterId} />
        </div>
      )}

      {/* Assign Form Modal */}
      <AssignmentForm
        encounterId={encounterId}
        isOpen={showAssignForm}
        onClose={() => setShowAssignForm(false)}
        onSubmit={assign}
        loading={isLoading}
      />

      {/* Handoff Form Modal */}
      <HandoffForm
        assignment={handoffTarget}
        isOpen={handoffTarget !== null}
        onClose={() => setHandoffTarget(null)}
        onSubmit={transfer}
        loading={isLoading}
      />
    </div>
  );
};
