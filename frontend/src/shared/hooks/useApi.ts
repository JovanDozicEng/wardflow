/**
 * useApi hook - Generic API call with loading/error states
 * TODO: Implement with loading, error, data states
 * TODO: Add automatic refetch capability
 * TODO: Add caching support (consider React Query for production)
 * TODO: Add abort controller for cleanup
 */

import { useState, useEffect, useCallback } from 'react';
import type { AxiosResponse } from 'axios';

interface UseApiOptions {
  immediate?: boolean; // Execute immediately on mount
}

interface UseApiReturn<T> {
  data: T | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  reset: () => void;
}

export const useApi = <T,>(
  apiFunc: () => Promise<AxiosResponse<T>>,
  options: UseApiOptions = { immediate: true }
): UseApiReturn<T> => {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await apiFunc();
      setData(response.data);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  }, [apiFunc]);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setIsLoading(false);
  }, []);

  useEffect(() => {
    if (options.immediate) {
      execute();
    }
  }, [execute, options.immediate]);

  return { data, isLoading, error, refetch: execute, reset };
};
