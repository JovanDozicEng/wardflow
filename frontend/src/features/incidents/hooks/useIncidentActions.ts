/**
 * useIncidentActions hook - Actions for incidents
 */

import { useState } from 'react';
import { incidentService } from '../services/incidentService';
import { useIncidentStore } from '../store/incidentStore';
import type { CreateIncidentRequest, UpdateIncidentStatusRequest } from '../types/incident.types';

export const useIncidentActions = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { addIncident, updateIncident, setSelectedIncident, setStatusHistory } = useIncidentStore();

  const createIncident = async (data: CreateIncidentRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await incidentService.create(data);
      addIncident(response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to create incident';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const fetchIncident = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await incidentService.get(id);
      setSelectedIncident(response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to fetch incident';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const updateStatus = async (id: string, data: UpdateIncidentStatusRequest) => {
    try {
      setLoading(true);
      setError(null);
      const response = await incidentService.updateStatus(id, data);
      updateIncident(id, response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to update incident status';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const fetchStatusHistory = async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await incidentService.getStatusHistory(id);
      setStatusHistory(response.data);
      return response.data;
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || 'Failed to fetch status history';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return {
    createIncident,
    fetchIncident,
    updateStatus,
    fetchStatusHistory,
    loading,
    error,
  };
};
