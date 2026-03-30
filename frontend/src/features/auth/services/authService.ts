/**
 * Auth API service
 * Handles all authentication-related API calls
 */

import api from '../../../shared/utils/api';
import type { LoginCredentials, RegisterRequest, LoginResponse } from '../types/auth.types';

export const authService = {
  /**
   * Login with email and password
   */
  login: async (credentials: LoginCredentials): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/login', credentials);
    return response.data;
  },

  /**
   * Register new user
   */
  register: async (data: RegisterRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/register', data);
    return response.data;
  },

  /**
   * Get current user info
   * Backend returns the user object directly (flat), not wrapped in { user: ... }
   */
  me: async (): Promise<import('../../../shared/types').User> => {
    const response = await api.get<import('../../../shared/types').User>('/auth/me');
    return response.data;
  },

  /**
   * Logout (optional - mainly clears client-side state)
   */
  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
  },

  /**
   * Change password
   */
  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    await api.post('/auth/change-password', {
      currentPassword,
      newPassword,
    });
  },

  /**
   * Refresh token (if refresh token flow is implemented)
   * TODO: Implement when backend adds refresh token endpoint
   */
  refresh: async (): Promise<{ token: string }> => {
    const response = await api.post<{ token: string }>('/auth/refresh');
    return response.data;
  },
};
