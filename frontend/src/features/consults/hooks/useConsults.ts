/**
 * useConsults hook - Fetch and manage consults
 */

import { useEffect } from 'react';
import { useConsultStore } from '../store/consultStore';
import { consultService } from '../services/consultService';

export const useConsults = (filters?: { status?: string; targetService?: string }) => {
  const { consults, loading, error, setConsults, setLoading, setError } = useConsultStore();

  const fetchConsults = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.list(filters);
      // Backend returns { data: [], total, limit, offset }
      // Extract the data array from the response
      const consultsData = response.data?.data || response.data || [];
      setConsults(Array.isArray(consultsData) ? consultsData : []);
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.response?.data?.message || 'Failed to fetch consults');
      console.error('Error fetching consults:', err);
      // Set empty array on error to prevent map() errors
      setConsults([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConsults();
  }, [filters?.status, filters?.targetService]);

  return {
    consults,
    loading,
    error,
    refetch: fetchConsults,
  };
};
