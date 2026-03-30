/**
 * Care Team types
 * Matches backend API for care team assignments and handoffs
 */

import type { User } from '../../../shared/types';

// Care team role types (matches backend RoleType enum)
export type CareTeamRole = 
  | 'primary_nurse'
  | 'attending_provider'
  | 'consulting_provider'
  | 'resident'
  | 'respiratory_therapist'
  | 'case_manager'
  | 'social_worker'
  | 'other';

// Care team assignment entity
export interface CareTeamAssignment {
  id: string;
  encounterId: string;
  userId: string;
  user?: User; // Populated from API
  roleType: CareTeamRole;
  startsAt: string; // ISO timestamp
  endsAt?: string; // ISO timestamp, null if currently active
  createdAt: string;
  createdBy: string;
  handoffNoteId?: string;
}

// Handoff note entity (updated to match backend)
export interface HandoffNote {
  id: string;
  encounterId: string;
  assignmentId: string;
  fromUserId: string;
  toUserId: string;
  roleType: CareTeamRole;
  note: string;
  structuredFields?: Record<string, any> | null;
  createdAt: string;
}

// Request to assign new role to encounter (matches backend AssignRoleRequest)
export interface AssignRoleRequest {
  userId: string;
  roleType: CareTeamRole;
  startsAt?: string; // Defaults to now if omitted
}

// Request to transfer role (with handoff) (matches backend TransferRoleRequest)
export interface TransferRoleRequest {
  toUserId: string;
  handoffNote?: string; // Required for critical roles
  structuredFields?: Record<string, any>;
}

// API response for care team list
export interface CareTeamResponse {
  encounterId: string;
  assignments?: CareTeamAssignment[]; // Simple response
  members?: CareTeamMember[]; // Response with user details
}

// Care team member with populated user details
export interface CareTeamMember {
  assignment: CareTeamAssignment;
  userName: string;
  userEmail: string;
}

// Care team state for store
export interface CareTeamState {
  assignments: CareTeamAssignment[];
  history: CareTeamAssignment[];
  handoffs: HandoffNote[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchByEncounter: (encounterId: string) => Promise<void>;
  assign: (encounterId: string, data: AssignRoleRequest) => Promise<void>;
  transfer: (assignmentId: string, data: TransferRoleRequest) => Promise<void>;
  fetchHistory: (encounterId: string) => Promise<void>;
  fetchHandoffs: (encounterId: string) => Promise<void>;
  clearError: () => void;
}
