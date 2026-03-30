/**
 * Patient service - API interactions for patient management
 */

import api from '../../../shared/utils/api';
import type { Patient, CreatePatientRequest, ListPatientsResponse } from '../types';

/**
 * List patients with optional search and pagination
 */
export const listPatients = async (
  q?: string,
  limit: number = 50,
  offset: number = 0
): Promise<ListPatientsResponse> => {
  const params: Record<string, string | number> = { limit, offset };
  if (q?.trim()) {
    params.q = q.trim();
  }

  const response = await api.get<ListPatientsResponse>('/patients', { params });
  return response.data;
};

/**
 * Get a single patient by ID
 */
export const getPatient = async (id: string): Promise<Patient> => {
  const response = await api.get<Patient>(`/patients/${id}`);
  return response.data;
};

/**
 * Create a new patient
 */
export const createPatient = async (data: CreatePatientRequest): Promise<Patient> => {
  const response = await api.post<Patient>('/patients', data);
  return response.data;
};

/**
 * Search patients (wrapper around listPatients for autocomplete)
 */
export const searchPatients = async (q: string): Promise<Patient[]> => {
  const response = await listPatients(q, 20, 0);
  return response.data;
};
