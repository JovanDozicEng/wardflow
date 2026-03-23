/**
 * Auth-related TypeScript types
 * Matches backend API contracts for authentication
 */

import type { User, Role } from '../../../shared/types';

// Login credentials
export interface LoginCredentials {
  email: string;
  password: string;
}

// Registration request
export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  role: Role;
  unitIds?: string[];
  departmentIds?: string[];
}

// Login response from API
export interface LoginResponse {
  user: User;
  token: string;
  expiresAt: number; // Unix timestamp
}

// User info response (from /auth/me endpoint)
export interface UserInfo {
  user: User;
}

// Change password request
export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

// Auth state for Zustand store
export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => void;
  loadUser: () => Promise<void>;
  clearError: () => void;
  
  // Permission helpers
  hasRole: (role: Role) => boolean;
  hasUnitAccess: (unitId: string) => boolean;
  hasDepartmentAccess: (departmentId: string) => boolean;
}
