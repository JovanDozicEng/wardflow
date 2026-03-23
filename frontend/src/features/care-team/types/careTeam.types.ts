/**
 * Care Team types
 * Matches backend API for care team assignments and handoffs
 */

import type { User } from '../../../shared/types';

// Care team role types
export type CareTeamRole = 'nurse' | 'provider' | 'charge_nurse' | 'specialist';

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

// Handoff note entity
export interface HandoffNote {
  id: string;
  encounterId: string;
  fromUserId: string;
  toUserId: string;
  roleType: CareTeamRole;
  note: string;
  structuredFields?: Record<string, any>; // JSON structured data
  createdAt: string;
}

// Request to assign new role to encounter
export interface AssignmentRequest {
  encounterId: string;
  userId: string;
  roleType: CareTeamRole;
  startsAt?: string; // Defaults to now
  note?: string;
}

// Request to transfer role (with handoff)
export interface TransferRequest {
  toUserId: string;
  handoffNote: string;
  structuredFields?: Record<string, any>;
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
  assign: (data: AssignmentRequest) => Promise<void>;
  transfer: (assignmentId: string, data: TransferRequest) => Promise<void>;
  fetchHistory: (encounterId: string) => Promise<void>;
  clearError: () => void;
}
