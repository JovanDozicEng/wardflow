/**
 * useExceptions hook - Fetch and manage exceptions
 */

import { useEffect } from 'react';
import { useExceptionStore } from '../store/exceptionStore';
import { exceptionService } from '../services/exceptionService';

export const useExceptions = (filters?: { type?: string; status?: string; encounterId?: string }) => {
  const { exceptions, loading, error, setExceptions, setLoading, setError } = useExceptionStore();

  const fetchExceptions = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await exceptionService.list(filters);
      setExceptions(response.data);
    } catch (err: any) {
      setError(err?.response?.data?.message || 'Failed to fetch exceptions');
      console.error('Error fetching exceptions:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchExceptions();
  }, [filters?.type, filters?.status, filters?.encounterId]);

  return {
    exceptions,
    loading,
    error,
    refetch: fetchExceptions,
  };
};
