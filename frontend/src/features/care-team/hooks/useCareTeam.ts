/**
 * useCareTeam hook
 * Convenience hook for care team operations
 */

import { useCareTeamStore } from '../store/careTeamStore';

export const useCareTeam = () => {
  const {
    assignments,
    history,
    handoffs,
    isLoading,
    error,
    fetchByEncounter,
    assign,
    transfer,
    fetchHistory,
    clearError,
  } = useCareTeamStore();

  return {
    assignments,
    history,
    handoffs,
    isLoading,
    error,
    fetchByEncounter,
    assign,
    transfer,
    fetchHistory,
    clearError,
  };
};
