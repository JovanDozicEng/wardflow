/**
 * RBAC (Role-Based Access Control) types
 */

import { Role } from './common.types';

// Permission action types
export const Permission = {
  // Care team permissions
  VIEW_CARE_TEAM: 'view_care_team',
  ASSIGN_CARE_TEAM: 'assign_care_team',
  TRANSFER_CARE_TEAM: 'transfer_care_team',
  
  // Task permissions
  VIEW_TASKS: 'view_tasks',
  CREATE_TASK: 'create_task',
  ASSIGN_TASK: 'assign_task',
  COMPLETE_TASK: 'complete_task',
  
  // Flow permissions
  VIEW_FLOW: 'view_flow',
  UPDATE_FLOW: 'update_flow',
  OVERRIDE_FLOW: 'override_flow', // Privileged
  
  // Consult permissions
  VIEW_CONSULTS: 'view_consults',
  CREATE_CONSULT: 'create_consult',
  ACCEPT_CONSULT: 'accept_consult',
  DECLINE_CONSULT: 'decline_consult',
  
  // Bed management
  VIEW_BEDS: 'view_beds',
  MANAGE_BEDS: 'manage_beds',
  
  // Transport
  VIEW_TRANSPORT: 'view_transport',
  CREATE_TRANSPORT: 'create_transport',
  ASSIGN_TRANSPORT: 'assign_transport',
  
  // Incidents
  VIEW_INCIDENTS: 'view_incidents',
  CREATE_INCIDENT: 'create_incident',
  REVIEW_INCIDENT: 'review_incident',
  CLOSE_INCIDENT: 'close_incident',
  
  // Admin
  MANAGE_USERS: 'manage_users',
  VIEW_AUDIT_LOG: 'view_audit_log',
} as const;

export type Permission = typeof Permission[keyof typeof Permission];

// Role to permissions mapping (client-side reference)
// Note: Backend is source of truth, this is for UI hints only
export const ROLE_PERMISSIONS: Record<Role, Permission[]> = {
  [Role.NURSE]: [
    Permission.VIEW_CARE_TEAM,
    Permission.VIEW_TASKS,
    Permission.CREATE_TASK,
    Permission.COMPLETE_TASK,
    Permission.VIEW_FLOW,
    Permission.VIEW_CONSULTS,
    Permission.CREATE_CONSULT,
    Permission.VIEW_TRANSPORT,
    Permission.CREATE_TRANSPORT,
    Permission.VIEW_INCIDENTS,
    Permission.CREATE_INCIDENT,
  ],
  [Role.PROVIDER]: [
    Permission.VIEW_CARE_TEAM,
    Permission.ASSIGN_CARE_TEAM,
    Permission.VIEW_TASKS,
    Permission.CREATE_TASK,
    Permission.ASSIGN_TASK,
    Permission.VIEW_FLOW,
    Permission.UPDATE_FLOW,
    Permission.VIEW_CONSULTS,
    Permission.CREATE_CONSULT,
    Permission.ACCEPT_CONSULT,
    Permission.DECLINE_CONSULT,
    Permission.VIEW_INCIDENTS,
    Permission.CREATE_INCIDENT,
  ],
  [Role.CHARGE_NURSE]: [
    Permission.VIEW_CARE_TEAM,
    Permission.ASSIGN_CARE_TEAM,
    Permission.TRANSFER_CARE_TEAM,
    Permission.VIEW_TASKS,
    Permission.CREATE_TASK,
    Permission.ASSIGN_TASK,
    Permission.VIEW_FLOW,
    Permission.UPDATE_FLOW,
    Permission.VIEW_CONSULTS,
    Permission.VIEW_BEDS,
    Permission.MANAGE_BEDS,
    Permission.VIEW_TRANSPORT,
    Permission.VIEW_INCIDENTS,
    Permission.CREATE_INCIDENT,
  ],
  [Role.OPERATIONS]: [
    Permission.VIEW_CARE_TEAM,
    Permission.VIEW_TASKS,
    Permission.VIEW_FLOW,
    Permission.OVERRIDE_FLOW,
    Permission.VIEW_CONSULTS,
    Permission.VIEW_BEDS,
    Permission.MANAGE_BEDS,
    Permission.VIEW_TRANSPORT,
    Permission.ASSIGN_TRANSPORT,
    Permission.VIEW_INCIDENTS,
  ],
  [Role.CONSULT]: [
    Permission.VIEW_CONSULTS,
    Permission.ACCEPT_CONSULT,
    Permission.DECLINE_CONSULT,
    Permission.VIEW_TASKS,
  ],
  [Role.TRANSPORT]: [
    Permission.VIEW_TRANSPORT,
    Permission.ASSIGN_TRANSPORT,
    Permission.VIEW_TASKS,
  ],
  [Role.QUALITY_SAFETY]: [
    Permission.VIEW_INCIDENTS,
    Permission.CREATE_INCIDENT,
    Permission.REVIEW_INCIDENT,
    Permission.CLOSE_INCIDENT,
    Permission.VIEW_AUDIT_LOG,
  ],
  [Role.ADMIN]: Object.values(Permission), // All permissions
};

// User context available throughout the app
export interface UserContext {
  user: {
    id: string;
    email: string;
    name: string;
    role: Role;
    unitIds: string[];
    departmentIds: string[];
  };
  hasRole: (role: Role) => boolean;
  hasPermission: (permission: Permission) => boolean;
  hasUnitAccess: (unitId: string) => boolean;
  hasDepartmentAccess: (departmentId: string) => boolean;
}
