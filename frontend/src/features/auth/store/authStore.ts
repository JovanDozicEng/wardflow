/**
 * Auth Zustand store
 * Global authentication state management
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthState, LoginCredentials, RegisterRequest } from '../types/auth.types';
import { authService } from '../services/authService';
import type { Role } from '../../../shared/types';
import { STORAGE_KEYS } from '../../../shared/utils/constants';

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (credentials: LoginCredentials) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authService.login(credentials);
          
          // Store token in localStorage (also persisted by zustand)
          localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, response.token);
          
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error: any) {
          set({
            error: error.response?.data?.error?.message || 'Login failed',
            isLoading: false,
          });
          throw error;
        }
      },

      register: async (data: RegisterRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authService.register(data);
          
          // Store token in localStorage
          localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, response.token);
          
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error: any) {
          set({
            error: error.response?.data?.error?.message || 'Registration failed',
            isLoading: false,
          });
          throw error;
        }
      },

      logout: () => {
        // Call logout endpoint (optional, mainly for server-side cleanup)
        authService.logout().catch(console.error);
        
        // Clear token from localStorage
        localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
        
        // Reset state
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          error: null,
        });
      },

      loadUser: async () => {
        const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
        if (!token) {
          set({ isAuthenticated: false });
          return;
        }

        set({ isLoading: true });
        try {
          // authService.me() returns the User object directly
          const user = await authService.me();
          set({
            user,
            token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          // Token invalid or expired
          localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },

      clearError: () => {
        set({ error: null });
      },

      hasRole: (role: Role): boolean => {
        const { user } = get();
        return user?.role === role;
      },

      hasUnitAccess: (unitId: string): boolean => {
        const { user } = get();
        return user?.unitIds?.includes(unitId) ?? false;
      },

      hasDepartmentAccess: (departmentId: string): boolean => {
        const { user } = get();
        return user?.departmentIds?.includes(departmentId) ?? false;
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
