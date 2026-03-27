/**
 * Common types shared across the application
 * Matches backend models from Go API
 */

// Role enum matching backend RBAC
export const Role = {
  NURSE: 'nurse',
  PROVIDER: 'provider',
  CHARGE_NURSE: 'charge_nurse',
  OPERATIONS: 'operations',
  CONSULT: 'consult',
  TRANSPORT: 'transport',
  QUALITY_SAFETY: 'quality_safety',
  ADMIN: 'admin',
} as const;

export type Role = typeof Role[keyof typeof Role];

// User entity matching backend User model
export interface User {
  id: string; // UUID
  email: string;
  name: string;
  role: Role;
  unitIds: string[]; // Units this user has access to
  departmentIds: string[]; // Departments this user has access to
  isActive: boolean;
  createdAt: string; // ISO timestamp
  updatedAt: string; // ISO timestamp
}

// Encounter entity - central anchor for care coordination
export interface Encounter {
  id: string; // UUID
  patientId: string;
  unitId: string;
  departmentId: string;
  status: EncounterStatus;
  startedAt: string; // ISO timestamp
  endedAt?: string; // ISO timestamp, optional
  createdAt: string;
  updatedAt: string;
  createdBy: string; // User ID
  updatedBy: string; // User ID
}

export const EncounterStatus = {
  ACTIVE: 'active',
  DISCHARGED: 'discharged',
  TRANSFERRED: 'transferred',
  CANCELLED: 'cancelled',
} as const;

export type EncounterStatus = typeof EncounterStatus[keyof typeof EncounterStatus];

// Audit fields mixin for entities requiring audit trail
export interface AuditFields {
  createdAt: string; // ISO timestamp
  updatedAt: string; // ISO timestamp
  createdBy: string; // User ID
  updatedBy?: string; // User ID
}

// Soft delete support
export interface SoftDeletable {
  deletedAt?: string; // ISO timestamp
}

// Common metadata for timeline/history entities
export interface TimelineEvent {
  id: string;
  timestamp: string; // ISO timestamp
  actorType: 'user' | 'system';
  actorUserId?: string;
  reason?: string;
  source?: string; // 'user_action' | 'system_event'
}
