/**
 * Unit and Department type definitions
 */

export interface Department {
  id: string;
  name: string;
  code: string;
  createdAt: string;
  updatedAt: string;
}

export interface Unit {
  id: string;
  name: string;
  code: string;
  departmentId: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateDepartmentRequest {
  name: string;
  code: string;
}

export interface CreateUnitRequest {
  name: string;
  code: string;
  departmentId: string;
}
