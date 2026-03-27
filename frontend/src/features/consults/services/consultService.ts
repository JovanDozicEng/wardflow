/**
 * Consult API service
 */

import { api } from '@/shared/utils/api';
import type {
  ConsultRequest,
  CreateConsultRequest,
  DeclineConsultRequest,
  RedirectConsultRequest,
} from '../types/consult.types';

export const consultService = {
  list: (params?: { unitId?: string; status?: string; targetService?: string }) =>
    api.get<ConsultRequest[]>('/consults', { params }),
  
  create: (data: CreateConsultRequest) =>
    api.post<ConsultRequest>('/consults', data),
  
  accept: (id: string) =>
    api.post<ConsultRequest>(`/consults/${id}/accept`),
  
  decline: (id: string, data: DeclineConsultRequest) =>
    api.post<ConsultRequest>(`/consults/${id}/decline`, data),
  
  redirect: (id: string, data: RedirectConsultRequest) =>
    api.post<ConsultRequest>(`/consults/${id}/redirect`, data),
  
  complete: (id: string) =>
    api.post<ConsultRequest>(`/consults/${id}/complete`),
};
