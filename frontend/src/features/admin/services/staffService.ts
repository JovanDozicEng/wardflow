import { api } from '../../../shared/utils/api';
import type { Role } from '../../../shared/types/common.types';

export interface StaffProfile {
  id: string;
  name: string;
  email: string;
  role: Role;
  isActive: boolean;
  unitIds: string[];
  departmentIds: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ListStaffParams {
  q?: string;
  role?: string;
  limit?: number;
  offset?: number;
}

export interface ListStaffResponse {
  data: StaffProfile[];
  total: number;
  limit: number;
  offset: number;
}

export interface UpdateStaffRequest {
  role?: Role;
  isActive?: boolean;
  unitIds?: string[];
  departmentIds?: string[];
}

export const staffService = {
  async listStaff(params: ListStaffParams = {}): Promise<ListStaffResponse> {
    const response = await api.get<ListStaffResponse>('/admin/staff', { params });
    return response.data;
  },

  async updateStaff(userId: string, patch: UpdateStaffRequest): Promise<StaffProfile> {
    const response = await api.patch<StaffProfile>(`/admin/staff/${userId}`, patch);
    return response.data;
  },
};
