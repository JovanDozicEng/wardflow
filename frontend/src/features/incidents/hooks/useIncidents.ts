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
      // Backend returns { data: [], total, limit, offset }
      const incidentsData = response.data?.data || response.data || [];
      setIncidents(Array.isArray(incidentsData) ? incidentsData : []);
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.response?.data?.message || 'Failed to fetch incidents');
      console.error('Error fetching incidents:', err);
      setIncidents([]);
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
