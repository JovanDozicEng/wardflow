/**
 * Exception API service
 */

import { api } from '@/shared/utils/api';
import type {
  ExceptionEvent,
  CreateExceptionRequest,
  UpdateExceptionRequest,
  CorrectExceptionRequest,
} from '../types/exception.types';

export const exceptionService = {
  list: (params?: { encounterId?: string; type?: string; status?: string }) =>
    api.get<ExceptionEvent[]>('/exceptions', { params }),
  
  create: (data: CreateExceptionRequest) =>
    api.post<ExceptionEvent>('/exceptions', data),
  
  update: (id: string, data: UpdateExceptionRequest) =>
    api.patch<ExceptionEvent>(`/exceptions/${id}`, data),
  
  finalize: (id: string) =>
    api.post<ExceptionEvent>(`/exceptions/${id}/finalize`),
  
  correct: (id: string, data: CorrectExceptionRequest) =>
    api.post<ExceptionEvent>(`/exceptions/${id}/correct`, data),
};
