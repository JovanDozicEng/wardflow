/**
 * Care Team Zustand store
 * State management for care team assignments
 */

import { create } from 'zustand';
import type { CareTeamState, AssignRoleRequest, TransferRoleRequest } from '../types/careTeam.types';
import { careTeamService } from '../services/careTeamService';

export const useCareTeamStore = create<CareTeamState>((set, _get) => ({
  assignments: [],
  history: [],
  handoffs: [],
  isLoading: false,
  error: null,

  fetchByEncounter: async (encounterId: string) => {
    set({ isLoading: true, error: null });
    try {
      const assignments = await careTeamService.getByEncounter(encounterId, true, true);
      set({ assignments, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to fetch care team',
        isLoading: false,
      });
    }
  },

  assign: async (encounterId: string, data: AssignRoleRequest) => {
    set({ isLoading: true, error: null });
    try {
      await careTeamService.assign(encounterId, data);
      // Refetch assignments
      const assignments = await careTeamService.getByEncounter(encounterId, true, true);
      set({ assignments, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to assign role',
        isLoading: false,
      });
      throw error;
    }
  },

  transfer: async (assignmentId: string, data: TransferRoleRequest) => {
    set({ isLoading: true, error: null });
    try {
      const result = await careTeamService.transfer(assignmentId, data);
      // Refetch assignments for the encounter
      if (result.encounterId) {
        const assignments = await careTeamService.getByEncounter(result.encounterId, true, true);
        set({ assignments, isLoading: false });
      } else {
        set({ isLoading: false });
      }
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to transfer role',
        isLoading: false,
      });
      throw error;
    }
  },

  fetchHistory: async (encounterId: string) => {
    set({ isLoading: true, error: null });
    try {
      const history = await careTeamService.getHistory(encounterId);
      const handoffs = await careTeamService.getHandoffs(encounterId);
      set({ history, handoffs, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to fetch history',
        isLoading: false,
      });
    }
  },

  fetchHandoffs: async (encounterId: string) => {
    set({ isLoading: true, error: null });
    try {
      const handoffs = await careTeamService.getHandoffs(encounterId);
      set({ handoffs, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to fetch handoffs',
        isLoading: false,
      });
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
