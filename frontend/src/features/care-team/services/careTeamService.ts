/**
 * Care Team API service
 * Handles all care team-related API calls
 */

import api from '../../../shared/utils/api';
import type { CareTeamAssignment, HandoffNote, AssignmentRequest, TransferRequest } from '../types/careTeam.types';

export const careTeamService = {
  /**
   * Get current care team for an encounter
   */
  getByEncounter: async (encounterId: string): Promise<CareTeamAssignment[]> => {
    const response = await api.get<CareTeamAssignment[]>(`/encounters/${encounterId}/care-team`, {
      params: { activeOnly: true },
    });
    return response.data;
  },

  /**
   * Assign a new role to the care team
   */
  assign: async (data: AssignmentRequest): Promise<CareTeamAssignment> => {
    const response = await api.post<CareTeamAssignment>(
      `/encounters/${data.encounterId}/care-team/assignments`,
      data
    );
    return response.data;
  },

  /**
   * Transfer role with handoff note
   */
  transfer: async (assignmentId: string, data: TransferRequest): Promise<CareTeamAssignment> => {
    const response = await api.post<CareTeamAssignment>(
      `/care-team/assignments/${assignmentId}/transfer`,
      data
    );
    return response.data;
  },

  /**
   * Get assignment history for an encounter
   */
  getHistory: async (encounterId: string): Promise<CareTeamAssignment[]> => {
    const response = await api.get<CareTeamAssignment[]>(`/encounters/${encounterId}/care-team`);
    return response.data;
  },

  /**
   * Get handoff notes for an encounter
   */
  getHandoffs: async (encounterId: string): Promise<HandoffNote[]> => {
    const response = await api.get<HandoffNote[]>(`/encounters/${encounterId}/handoffs`);
    return response.data;
  },
};
