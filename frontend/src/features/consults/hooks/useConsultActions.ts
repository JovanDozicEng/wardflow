/**
 * useConsultActions hook - Actions for consults
 */

import { useState } from 'react';
import { consultService } from '../services/consultService';
import { useConsultStore } from '../store/consultStore';
import type {
  CreateConsultRequest,
  DeclineConsultRequest,
  RedirectConsultRequest,
} from '../types/consult.types';

export const useConsultActions = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { addConsult, updateConsult } = useConsultStore();

  const createConsult = async (data: CreateConsultRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.create(data);
      addConsult(response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to create consult';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const acceptConsult = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.accept(id);
      updateConsult(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to accept consult';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const declineConsult = async (id: string, data: DeclineConsultRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.decline(id, data);
      updateConsult(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to decline consult';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const redirectConsult = async (id: string, data: RedirectConsultRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.redirect(id, data);
      updateConsult(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to redirect consult';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const completeConsult = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await consultService.complete(id);
      updateConsult(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to complete consult';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return {
    createConsult,
    acceptConsult,
    declineConsult,
    redirectConsult,
    completeConsult,
    loading,
    error,
  };
};
