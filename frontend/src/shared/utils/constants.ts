/**
 * Application-wide constants
 * Single source of truth for configuration values
 */

import { Role } from '../types';

// API Configuration
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  TIMEOUT: 10000, // 10 seconds
} as const;

// Role display names
export const ROLE_LABELS: Record<Role, string> = {
  [Role.NURSE]: 'Nurse',
  [Role.PROVIDER]: 'Provider',
  [Role.CHARGE_NURSE]: 'Charge Nurse',
  [Role.OPERATIONS]: 'Operations',
  [Role.CONSULT]: 'Consult Service',
  [Role.TRANSPORT]: 'Transport',
  [Role.QUALITY_SAFETY]: 'Quality & Safety',
  [Role.ADMIN]: 'Administrator',
} as const;

// Task statuses
export const TASK_STATUSES = {
  OPEN: 'open',
  IN_PROGRESS: 'in_progress',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
  ESCALATED: 'escalated',
} as const;

export const TASK_STATUS_LABELS: Record<string, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
  escalated: 'Escalated',
} as const;

// Priority levels
export const PRIORITY_LEVELS = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high',
  URGENT: 'urgent',
} as const;

export const PRIORITY_LABELS: Record<string, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
} as const;

// Consult statuses
export const CONSULT_STATUSES = {
  PENDING: 'pending',
  ACCEPTED: 'accepted',
  DECLINED: 'declined',
  COMPLETED: 'completed',
  REDIRECTED: 'redirected',
  CANCELLED: 'cancelled',
} as const;

export const CONSULT_STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  accepted: 'Accepted',
  declined: 'Declined',
  completed: 'Completed',
  redirected: 'Redirected',
  cancelled: 'Cancelled',
} as const;

// Bed statuses
export const BED_STATUSES = {
  AVAILABLE: 'available',
  OCCUPIED: 'occupied',
  BLOCKED: 'blocked',
  MAINTENANCE: 'maintenance',
  CLEANING: 'cleaning',
} as const;

export const BED_STATUS_LABELS: Record<string, string> = {
  available: 'Available',
  occupied: 'Occupied',
  blocked: 'Blocked',
  maintenance: 'Maintenance',
  cleaning: 'Cleaning',
} as const;

// Transport request statuses
export const TRANSPORT_STATUSES = {
  PENDING: 'pending',
  ASSIGNED: 'assigned',
  IN_TRANSIT: 'in_transit',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const;

export const TRANSPORT_STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  assigned: 'Assigned',
  in_transit: 'In Transit',
  completed: 'Completed',
  cancelled: 'Cancelled',
} as const;

// Pagination defaults
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
} as const;

// Local storage keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_PREFERENCES: 'user_preferences',
  SIDEBAR_COLLAPSED: 'sidebar_collapsed',
} as const;

// Date/time formats
export const DATE_FORMATS = {
  DISPLAY: 'MMM dd, yyyy',
  DISPLAY_WITH_TIME: 'MMM dd, yyyy \'at\' h:mm a',
  TIME_ONLY: 'h:mm a',
  ISO: 'yyyy-MM-dd\'T\'HH:mm:ss.SSSxxx',
} as const;

// Validation constraints
export const VALIDATION = {
  PASSWORD_MIN_LENGTH: 8,
  EMAIL_MAX_LENGTH: 255,
  NAME_MAX_LENGTH: 100,
  NOTE_MAX_LENGTH: 2000,
} as const;
