/**
 * Flow Tracking API service
 * Handles all flow state transition API calls
 */

import api from '../../../shared/utils/api';
import type {
  FlowTimelineResponse,
  CurrentStateResponse,
  CreateTransitionRequest,
  OverrideTransitionRequest,
  FlowStateTransition,
} from '../types';

export const flowService = {
  /**
   * Get flow state timeline for an encounter
   * @param encounterId - The encounter ID
   * @param withActors - Include actor details (default: false)
   * @param paginated - Use pagination (default: false)
   * @param limit - Max transitions to return
   * @param offset - Number of transitions to skip
   */
  getTimeline: async (
    encounterId: string,
    options?: {
      withActors?: boolean;
      paginated?: boolean;
      limit?: number;
      offset?: number;
    }
  ): Promise<FlowTimelineResponse> => {
    const response = await api.get<FlowTimelineResponse>(
      `/api/v1/encounters/${encounterId}/flow`,
      {
        params: {
          withActors: options?.withActors ?? false,
          paginated: options?.paginated ?? false,
          limit: options?.limit,
          offset: options?.offset,
        },
      }
    );
    return response.data;
  },

  /**
   * Get current flow state only
   * @param encounterId - The encounter ID
   */
  getCurrentState: async (encounterId: string): Promise<CurrentStateResponse> => {
    const response = await api.get<CurrentStateResponse>(
      `/api/v1/encounters/${encounterId}/flow/current`
    );
    return response.data;
  },

  /**
   * Record a flow state transition
   * @param encounterId - The encounter ID
   * @param data - Transition request data
   */
  recordTransition: async (
    encounterId: string,
    data: CreateTransitionRequest
  ): Promise<FlowStateTransition> => {
    const response = await api.post<FlowStateTransition>(
      `/api/v1/encounters/${encounterId}/flow/transitions`,
      data
    );
    return response.data;
  },

  /**
   * Override a flow state transition (privileged operation)
   * Requires admin or operations role
   * @param encounterId - The encounter ID
   * @param data - Override request with mandatory reason
   */
  overrideTransition: async (
    encounterId: string,
    data: OverrideTransitionRequest
  ): Promise<FlowStateTransition> => {
    const response = await api.post<FlowStateTransition>(
      `/api/v1/encounters/${encounterId}/flow/override`,
      data
    );
    return response.data;
  },
};
