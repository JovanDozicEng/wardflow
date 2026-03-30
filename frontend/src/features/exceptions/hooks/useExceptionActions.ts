/**
 * useExceptionActions hook - Actions for exceptions
 */

import { useState } from 'react';
import { exceptionService } from '../services/exceptionService';
import { useExceptionStore } from '../store/exceptionStore';
import type {
  CreateExceptionRequest,
  UpdateExceptionRequest,
  CorrectExceptionRequest,
} from '../types/exception.types';

export const useExceptionActions = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { addException, updateException } = useExceptionStore();

  const createException = async (data: CreateExceptionRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await exceptionService.create(data);
      addException(response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to create exception';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const updateExceptionData = async (id: string, data: UpdateExceptionRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await exceptionService.update(id, data);
      updateException(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to update exception';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const finalizeException = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await exceptionService.finalize(id);
      updateException(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to finalize exception';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const correctException = async (id: string, data: CorrectExceptionRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await exceptionService.correct(id, data);
      addException(response.data); // Add the correction as a new exception
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to correct exception';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return {
    createException,
    updateExceptionData,
    finalizeException,
    correctException,
    loading,
    error,
  };
};
