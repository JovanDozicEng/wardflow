/**
 * Dashboard API service
 * Handles huddle dashboard metrics API calls
 */

import api from '../../../shared/utils/api';
import type { HuddleMetrics, DashboardFilterParams } from '../types';

// Raw backend response shape (flat structure from Go backend)
interface RawHuddleMetrics {
  generatedAt: string;
  activeEncounters: number;
  expectedDischarges: number;
  expectedAdmissions?: number;
  flowDistribution: HuddleMetrics['flowDistribution'];
  taskMetrics: HuddleMetrics['taskMetrics'];
  riskIndicators: HuddleMetrics['riskIndicators'];
  overdueTasks?: HuddleMetrics['overdueTasks'];
  longStayPatients?: HuddleMetrics['longStayPatients'];
  pendingDischarges?: HuddleMetrics['pendingDischarges'];
  unitId?: string | null;
  departmentId?: string | null;
}

export const dashboardService = {
  /**
   * Get huddle dashboard metrics
   * @param filters - Optional unit/department filters (RBAC enforced on backend)
   */
  getHuddleMetrics: async (filters?: DashboardFilterParams): Promise<HuddleMetrics> => {
    const response = await api.get<RawHuddleMetrics>('/dashboard/huddle', {
      params: filters,
    });

    const raw = response.data;

    // Map flat backend response to nested frontend type
    return {
      unitId: raw.unitId ?? null,
      departmentId: raw.departmentId ?? null,
      generatedAt: raw.generatedAt,
      census: {
        active: raw.activeEncounters ?? 0,
        expectedDischarges: raw.expectedDischarges ?? 0,
      },
      flowDistribution: raw.flowDistribution ?? {},
      taskMetrics: raw.taskMetrics ?? {
        totalOpen: 0,
        totalOverdue: 0,
        highPriority: 0,
        urgent: 0,
        unassigned: 0,
        completedToday: 0,
      },
      riskIndicators: raw.riskIndicators ?? {
        patientsInTriageOver2hrs: 0,
        patientsWaitingForBedOver4hrs: 0,
        overdueHighPriorityTasks: 0,
        unassignedUrgentTasks: 0,
        encountersWithoutCareTeam: 0,
      },
      overdueTasks: raw.overdueTasks ?? [],
      longStayPatients: raw.longStayPatients ?? [],
      pendingDischarges: raw.pendingDischarges ?? [],
    };
  },
};
