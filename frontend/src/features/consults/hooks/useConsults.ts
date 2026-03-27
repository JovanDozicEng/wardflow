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
      setConsults(response.data);
    } catch (err: any) {
      setError(err?.response?.data?.message || 'Failed to fetch consults');
      console.error('Error fetching consults:', err);
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
