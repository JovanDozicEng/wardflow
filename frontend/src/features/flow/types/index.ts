/**
 * Flow Tracking types
 * Matches backend API for patient flow state transitions
 */

// Flow state enum (matches backend FlowState)
export type FlowState =
  | 'arrived'
  | 'triage'
  | 'provider_eval'
  | 'diagnostics'
  | 'admitted'
  | 'discharge_ready'
  | 'discharged';

// Actor type enum (matches backend ActorType)
export type ActorType = 'user' | 'system';

// Flow state transition entity (matches backend)
export interface FlowStateTransition {
  id: string;
  encounterId: string;
  fromState: FlowState | null; // Null for initial state
  toState: FlowState;
  transitionedAt: string; // ISO timestamp
  actorType: ActorType;
  actorUserId: string | null;
  reason: string | null;
  sourceEventId: string | null;
  isOverride: boolean; // Marks privileged invalid transitions
  createdAt: string;
  // Populated fields (when withActors=true)
  actorName?: string;
  actorEmail?: string;
}

// Request to record a transition (matches backend CreateTransitionRequest)
export interface CreateTransitionRequest {
  toState: FlowState;
  transitionedAt?: string; // Defaults to now if omitted
  reason?: string;
}

// Request to override a transition (matches backend OverrideTransitionRequest)
export interface OverrideTransitionRequest {
  fromState?: FlowState | null; // Explicit from state for override
  toState: FlowState;
  reason: string; // Mandatory for overrides
  transitionedAt?: string;
}

// Flow timeline response (matches backend FlowTimelineResponse)
export interface FlowTimelineResponse {
  encounterId: string;
  currentState: FlowState | null;
  transitions: FlowStateTransition[];
  total: number;
}

// Current state response
export interface CurrentStateResponse {
  encounterId: string;
  currentState: FlowState;
}

// Valid state transitions map (for client-side validation)
export const ValidTransitions: Record<FlowState, FlowState[]> = {
  arrived: ['triage', 'discharge_ready'], // Can skip to discharge if needed
  triage: ['provider_eval', 'discharge_ready'],
  provider_eval: ['diagnostics', 'admitted', 'discharge_ready'],
  diagnostics: ['admitted', 'provider_eval', 'discharge_ready'],
  admitted: ['discharge_ready'],
  discharge_ready: ['discharged', 'admitted'], // Can return to admitted if needed
  discharged: [], // Terminal state
};

// Helper function to check if transition is valid
export const isValidTransition = (from: FlowState | null, to: FlowState): boolean => {
  if (!from) return true; // First transition is always valid
  return ValidTransitions[from]?.includes(to) ?? false;
};

// Helper to get next valid states
export const getNextValidStates = (currentState: FlowState | null): FlowState[] => {
  if (!currentState) return ['arrived']; // Initial state
  return ValidTransitions[currentState] || [];
};

// Flow state display labels
export const FlowStateLabels: Record<FlowState, string> = {
  arrived: 'Arrived',
  triage: 'In Triage',
  provider_eval: 'Provider Evaluation',
  diagnostics: 'Diagnostics',
  admitted: 'Admitted',
  discharge_ready: 'Discharge Ready',
  discharged: 'Discharged',
};

// Flow state colors for UI
export const FlowStateColors: Record<FlowState, string> = {
  arrived: 'gray',
  triage: 'yellow',
  provider_eval: 'blue',
  diagnostics: 'purple',
  admitted: 'green',
  discharge_ready: 'orange',
  discharged: 'slate',
};
