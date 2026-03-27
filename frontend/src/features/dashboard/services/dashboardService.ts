/**
 * Dashboard API service
 * Handles huddle dashboard metrics API calls
 */

import api from '../../../shared/utils/api';
import type { HuddleMetrics, DashboardFilterParams } from '../types';

export const dashboardService = {
  /**
   * Get huddle dashboard metrics
   * @param filters - Optional unit/department filters (RBAC enforced on backend)
   */
  getHuddleMetrics: async (filters?: DashboardFilterParams): Promise<HuddleMetrics> => {
    const response = await api.get<HuddleMetrics>('/api/v1/dashboard/huddle', {
      params: filters,
    });
    return response.data;
  },
};
