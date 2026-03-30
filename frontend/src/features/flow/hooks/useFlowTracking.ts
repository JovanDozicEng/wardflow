/**
 * useFlowTracking hook
 * Custom hook for managing flow state transitions
 */

import { useState, useEffect, useCallback } from 'react';
import { flowService } from '../services/flowService';
import type {
  FlowState,
  FlowStateTransition,
  CreateTransitionRequest,
  OverrideTransitionRequest,
} from '../types';

interface UseFlowTrackingOptions {
  encounterId: string;
  autoRefresh?: boolean;
  refreshInterval?: number; // milliseconds
}

interface UseFlowTrackingReturn {
  currentState: FlowState | null;
  transitions: FlowStateTransition[];
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  recordTransition: (data: CreateTransitionRequest) => Promise<void>;
  overrideTransition: (data: OverrideTransitionRequest) => Promise<void>;
}

export const useFlowTracking = ({
  encounterId,
  autoRefresh = false,
  refreshInterval = 30000, // 30 seconds default
}: UseFlowTrackingOptions): UseFlowTrackingReturn => {
  const [currentState, setCurrentState] = useState<FlowState | null>(null);
  const [transitions, setTransitions] = useState<FlowStateTransition[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTimeline = useCallback(async () => {
    try {
      setError(null);
      const data = await flowService.getTimeline(encounterId, {
        withActors: true,
      });
      setCurrentState(data.currentState);
      setTransitions(data.transitions.reverse()); // Reverse to show newest first
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to fetch flow timeline');
      console.error('Flow timeline fetch error:', err);
    } finally {
      setIsLoading(false);
    }
  }, [encounterId]);

  const recordTransition = useCallback(
    async (data: CreateTransitionRequest) => {
      try {
        await flowService.recordTransition(encounterId, data);
        // Refresh timeline after successful transition
        await fetchTimeline();
      } catch (err: any) {
        const message = err.response?.data?.error?.message || 'Failed to record transition';
        throw new Error(message);
      }
    },
    [encounterId, fetchTimeline]
  );

  const overrideTransition = useCallback(
    async (data: OverrideTransitionRequest) => {
      try {
        await flowService.overrideTransition(encounterId, data);
        // Refresh timeline after successful override
        await fetchTimeline();
      } catch (err: any) {
        const message = err.response?.data?.error?.message || 'Failed to override transition';
        throw new Error(message);
      }
    },
    [encounterId, fetchTimeline]
  );

  // Initial fetch
  useEffect(() => {
    fetchTimeline();
  }, [fetchTimeline]);

  // Auto-refresh setup
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      fetchTimeline();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, fetchTimeline]);

  return {
    currentState,
    transitions,
    isLoading,
    error,
    refresh: fetchTimeline,
    recordTransition,
    overrideTransition,
  };
};
