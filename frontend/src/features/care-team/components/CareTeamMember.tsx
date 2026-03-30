/**
 * CareTeamMember - Display card for a single care team assignment
 */

import { Button } from '@/shared/components/ui/Button';
import type { CareTeamAssignment } from '../types/careTeam.types';

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

const ROLE_COLORS: Record<string, string> = {
  primary_nurse: 'bg-blue-100 text-blue-800',
  attending_provider: 'bg-purple-100 text-purple-800',
  consulting_provider: 'bg-indigo-100 text-indigo-800',
  resident: 'bg-cyan-100 text-cyan-800',
  respiratory_therapist: 'bg-teal-100 text-teal-800',
  case_manager: 'bg-orange-100 text-orange-800',
  social_worker: 'bg-pink-100 text-pink-800',
  other: 'bg-gray-100 text-gray-800',
};

interface CareTeamMemberProps {
  assignment: CareTeamAssignment;
  onHandoff?: (assignment: CareTeamAssignment) => void;
  canTransfer?: boolean;
}

export const CareTeamMember = ({
  assignment,
  onHandoff,
  canTransfer = false,
}: CareTeamMemberProps) => {
  const roleLabel = ROLE_LABELS[assignment.roleType] ?? assignment.roleType;
  const roleColor = ROLE_COLORS[assignment.roleType] ?? ROLE_COLORS.other;
  const displayName = assignment.user?.name ?? assignment.userId;

  return (
    <div className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg">
      <div className="flex items-center gap-3 min-w-0">
        {/* Avatar placeholder */}
        <div className="w-9 h-9 rounded-full bg-gray-200 flex items-center justify-center text-sm font-semibold text-gray-600 shrink-0">
          {displayName.charAt(0).toUpperCase()}
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-gray-900 truncate">{displayName}</p>
          {assignment.user?.email && (
            <p className="text-xs text-gray-500 truncate">{assignment.user.email}</p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${roleColor}`}>
          {roleLabel}
        </span>
        {canTransfer && onHandoff && (
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onHandoff(assignment)}
            className="text-xs px-2 py-1"
          >
            Handoff
          </Button>
        )}
      </div>
    </div>
  );
};
