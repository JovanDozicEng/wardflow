/**
 * useIncidents hook - Fetch and manage incidents
 */

import { useEffect } from 'react';
import { useIncidentStore } from '../store/incidentStore';
import { incidentService } from '../services/incidentService';

export const useIncidents = (filters?: { status?: string; type?: string }) => {
  const { incidents, loading, error, setIncidents, setLoading, setError } = useIncidentStore();

  const fetchIncidents = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await incidentService.list(filters);
      setIncidents(response.data);
    } catch (err: any) {
      setError(err?.response?.data?.message || 'Failed to fetch incidents');
      console.error('Error fetching incidents:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchIncidents();
  }, [filters?.status, filters?.type]);

  return {
    incidents,
    loading,
    error,
    refetch: fetchIncidents,
  };
};
