/**
 * Care Team API service
 * Handles all care team-related API calls
 */

import api from '../../../shared/utils/api';
import type { User } from '../../../shared/types';
import type { 
  CareTeamAssignment, 
  HandoffNote, 
  AssignRoleRequest, 
  TransferRoleRequest,
  CareTeamResponse 
} from '../types/careTeam.types';

export const careTeamService = {
  /**
   * Get current care team for an encounter
   * @param encounterId - The encounter ID
   * @param activeOnly - Return only active assignments (default: true)
   * @param withDetails - Include user details (default: false)
   */
  getByEncounter: async (
    encounterId: string, 
    activeOnly = true,
    withDetails = false
  ): Promise<CareTeamAssignment[]> => {
    const response = await api.get<CareTeamResponse>(
      `/encounters/${encounterId}/care-team/assignments`,
      {
        params: { activeOnly, withDetails },
      }
    );
    
    // Backend returns either { assignments } or { members } based on withDetails
    if (withDetails && response.data.members) {
      // Convert members to assignments with partial user data from API
      return response.data.members.map(m => ({
        ...m.assignment,
        user: {
          id: m.assignment.userId,
          name: m.userName,
          email: m.userEmail,
        } as User,
      }));
    }
    
    return response.data.assignments || [];
  },

  /**
   * Assign a new role to the care team
   * @param encounterId - The encounter ID
   * @param data - Assignment request data
   */
  assign: async (encounterId: string, data: AssignRoleRequest): Promise<CareTeamAssignment> => {
    const response = await api.post<CareTeamAssignment>(
      `/encounters/${encounterId}/care-team/assignments`,
      data
    );
    return response.data;
  },

  /**
   * Transfer role with handoff note
   * @param assignmentId - The assignment ID to transfer
   * @param data - Transfer request with handoff note
   */
  transfer: async (assignmentId: string, data: TransferRoleRequest): Promise<CareTeamAssignment> => {
    const response = await api.post<CareTeamAssignment>(
      `/care-team/assignments/${assignmentId}/transfer`,
      data
    );
    return response.data;
  },

  /**
   * Get assignment history for an encounter (includes ended assignments)
   * @param encounterId - The encounter ID
   */
  getHistory: async (encounterId: string): Promise<CareTeamAssignment[]> => {
    const response = await api.get<CareTeamResponse>(
      `/encounters/${encounterId}/care-team/assignments`,
      {
        params: { activeOnly: false },
      }
    );
    return response.data.assignments || [];
  },

  /**
   * Get handoff notes for an encounter
   * @param encounterId - The encounter ID
   */
  getHandoffs: async (encounterId: string): Promise<HandoffNote[]> => {
    const response = await api.get<{ data: HandoffNote[] }>(
      `/encounters/${encounterId}/handoffs`
    );
    return response.data.data;
  },
};
