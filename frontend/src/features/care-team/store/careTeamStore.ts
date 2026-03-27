/**
 * Care Team Zustand store
 * State management for care team assignments
 */

import { create } from 'zustand';
import type { CareTeamState, AssignmentRequest, TransferRequest } from '../types/careTeam.types';
import { careTeamService } from '../services/careTeamService';

export const useCareTeamStore = create<CareTeamState>((set) => ({
  assignments: [],
  history: [],
  handoffs: [],
  isLoading: false,
  error: null,

  fetchByEncounter: async (encounterId: string) => {
    set({ isLoading: true, error: null });
    try {
      const assignments = await careTeamService.getByEncounter(encounterId);
      set({ assignments, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to fetch care team',
        isLoading: false,
      });
    }
  },

  assign: async (data: AssignmentRequest) => {
    set({ isLoading: true, error: null });
    try {
      await careTeamService.assign(data);
      // Refetch assignments
      const assignments = await careTeamService.getByEncounter(data.encounterId);
      set({ assignments, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error?.message || 'Failed to assign role',
        isLoading: false,
      });
      throw error;
    }
  },

  transfer: async (assignmentId: string, data: TransferRequest) => {
    set({ isLoading: true, error: null });
    try {
      await careTeamService.transfer(assignmentId, data);
      // Note: Need encounterId to refetch, should be passed or stored
      set({ isLoading: false });
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

  clearError: () => {
    set({ error: null });
  },
}));
