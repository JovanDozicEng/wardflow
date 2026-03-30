/**
 * Dashboard types
 * Matches backend API for daily huddle metrics
 */

import type { FlowState } from '../../flow/types';
import type { TaskPriority, ScopeType } from '../../tasks/types';

// Huddle metrics top-level response (matches backend HuddleMetrics)
export interface HuddleMetrics {
  unitId: string | null;
  departmentId: string | null;
  generatedAt: string; // ISO timestamp
  census: CensusMetrics;
  flowDistribution: FlowDistribution;
  taskMetrics: TaskMetrics;
  riskIndicators: RiskIndicators;
  overdueTasks: TaskSummary[];
  longStayPatients: EncounterSummary[];
  pendingDischarges: EncounterSummary[];
}

// Census metrics
export interface CensusMetrics {
  active: number;
  expectedDischarges: number;
}

// Flow state distribution (matches backend FlowDistribution)
export interface FlowDistribution {
  arrived: number;
  triage: number;
  providerEval: number;
  diagnostics: number;
  admitted: number;
  dischargeReady: number;
  discharged: number;
}

// Task metrics overview (matches backend TaskMetrics)
export interface TaskMetrics {
  totalOpen: number;
  totalOverdue: number;
  highPriority: number;
  urgent: number;
  unassigned: number;
  completedToday: number;
}

// Risk indicators (matches backend RiskIndicators)
export interface RiskIndicators {
  patientsInTriageOver2hrs: number;
  patientsWaitingForBedOver4hrs: number;
  overdueHighPriorityTasks: number;
  unassignedUrgentTasks: number;
  encountersWithoutCareTeam: number;
}

// Task summary for drill-down (matches backend TaskSummary)
export interface TaskSummary {
  id: string;
  title: string;
  priority: TaskPriority;
  slaDueAt: string | null;
  ownerName: string | null;
  scopeType: ScopeType;
  scopeId: string;
}

// Encounter summary for drill-down (matches backend EncounterSummary)
export interface EncounterSummary {
  id: string;
  patientId: string;
  unitId: string;
  departmentId: string;
  currentState: FlowState | null;
  startedAt: string; // ISO timestamp
  lengthOfStay: string; // Human-readable (e.g., '2d 3h')
}

// Dashboard filter parameters
export interface DashboardFilterParams {
  unitId?: string;
  departmentId?: string;
}

// Helper to parse flow distribution into array for charts
export const flowDistributionToArray = (
  dist: FlowDistribution
): Array<{ state: FlowState; count: number; label: string }> => {
  return [
    { state: 'arrived', count: dist.arrived, label: 'Arrived' },
    { state: 'triage', count: dist.triage, label: 'Triage' },
    { state: 'provider_eval', count: dist.providerEval, label: 'Provider Eval' },
    { state: 'diagnostics', count: dist.diagnostics, label: 'Diagnostics' },
    { state: 'admitted', count: dist.admitted, label: 'Admitted' },
    { state: 'discharge_ready', count: dist.dischargeReady, label: 'Discharge Ready' },
    { state: 'discharged', count: dist.discharged, label: 'Discharged' },
  ];
};

// Helper to identify high-risk indicators
export const getHighRiskIndicators = (
  risks: RiskIndicators
): Array<{ key: string; value: number; label: string; threshold: number }> => {
  return [
    {
      key: 'triageOver2hrs',
      value: risks.patientsInTriageOver2hrs,
      label: 'Triage >2hrs',
      threshold: 3,
    },
    {
      key: 'waitingForBed',
      value: risks.patientsWaitingForBedOver4hrs,
      label: 'Waiting for Bed >4hrs',
      threshold: 2,
    },
    {
      key: 'overdueHighPriority',
      value: risks.overdueHighPriorityTasks,
      label: 'Overdue High Priority Tasks',
      threshold: 5,
    },
    {
      key: 'unassignedUrgent',
      value: risks.unassignedUrgentTasks,
      label: 'Unassigned Urgent Tasks',
      threshold: 1,
    },
    {
      key: 'noCareTeam',
      value: risks.encountersWithoutCareTeam,
      label: 'No Care Team Assigned',
      threshold: 0,
    },
  ].filter((indicator) => indicator.value > indicator.threshold);
};
