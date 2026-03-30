/**
 * Unit and Department service - API interactions
 */

import api from '../../../shared/utils/api';
import type {
  Unit,
  Department,
  CreateUnitRequest,
  CreateDepartmentRequest,
} from '../types';

/**
 * List all departments with optional search query
 */
export const listDepartments = async (q?: string): Promise<Department[]> => {
  const params: Record<string, string> = {};
  if (q?.trim()) {
    params.q = q.trim();
  }

  const response = await api.get<Department[]>('/departments', { params });
  return response.data;
};

/**
 * List all units with optional search query and department filter
 */
export const listUnits = async (q?: string, departmentId?: string): Promise<Unit[]> => {
  const params: Record<string, string> = {};
  if (q?.trim()) {
    params.q = q.trim();
  }
  if (departmentId?.trim()) {
    params.departmentId = departmentId.trim();
  }

  const response = await api.get<Unit[]>('/units', { params });
  return response.data;
};

/**
 * Create a new department (admin only)
 */
export const createDepartment = async (data: CreateDepartmentRequest): Promise<Department> => {
  const response = await api.post<Department>('/departments', data);
  return response.data;
};

/**
 * Create a new unit (admin only)
 */
export const createUnit = async (data: CreateUnitRequest): Promise<Unit> => {
  const response = await api.post<Unit>('/units', data);
  return response.data;
};
