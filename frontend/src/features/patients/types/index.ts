/**
 * Patient module type definitions
 */

export interface Patient {
  id: string;
  firstName: string;
  lastName: string;
  dateOfBirth?: string; // ISO date e.g. "1985-04-12"
  mrn: string; // Medical Record Number — unique identifier
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePatientRequest {
  firstName: string;
  lastName: string;
  mrn: string;
  dateOfBirth?: string;
}

export interface ListPatientsResponse {
  data: Patient[];
  total: number;
  limit: number;
  offset: number;
}
