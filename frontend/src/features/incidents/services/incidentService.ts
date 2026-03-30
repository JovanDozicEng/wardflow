/**
 * Incident API service
 */

import { api } from '@/shared/utils/api';
import type {
  Incident,
  CreateIncidentRequest,
  UpdateIncidentStatusRequest,
  IncidentStatusEvent,
} from '../types/incident.types';

export const incidentService = {
  list: (params?: { unitId?: string; status?: string; type?: string }) =>
    api.get<{ data: Incident[]; total: number; limit: number; offset: number }>('/incidents', { params }),
  
  create: (data: CreateIncidentRequest) =>
    api.post<Incident>('/incidents', data),
  
  get: (id: string) =>
    api.get<Incident>(`/incidents/${id}`),
  
  updateStatus: (id: string, data: UpdateIncidentStatusRequest) =>
    api.post<Incident>(`/incidents/${id}/status`, data),
  
  getStatusHistory: (id: string) =>
    api.get<IncidentStatusEvent[]>(`/incidents/${id}/status-history`),
};
