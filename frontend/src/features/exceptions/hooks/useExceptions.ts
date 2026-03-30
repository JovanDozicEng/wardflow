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
      // Backend returns { data: [], total, limit, offset }
      const exceptionsData = response.data?.data || response.data || [];
      setExceptions(Array.isArray(exceptionsData) ? exceptionsData : []);
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.response?.data?.message || 'Failed to fetch exceptions');
      console.error('Error fetching exceptions:', err);
      setExceptions([]);
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
