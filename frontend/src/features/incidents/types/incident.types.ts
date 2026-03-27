/**
 * Incident types and interfaces
 */

export type IncidentStatus = 'submitted' | 'under_review' | 'closed';

export interface Incident {
  id: string;
  encounterId?: string;
  type: string;
  severity?: string;
  harmIndicators?: Record<string, any>;
  eventTime: string;
  reportedBy: string;
  reportedAt: string;
  status: IncidentStatus;
  createdAt: string;
  updatedAt: string;
}

export interface CreateIncidentRequest {
  encounterId?: string;
  type: string;
  severity?: string;
  harmIndicators?: Record<string, any>;
  eventTime: string;
}

export interface UpdateIncidentStatusRequest {
  status: IncidentStatus;
  note?: string;
}

export interface IncidentStatusEvent {
  id: string;
  incidentId: string;
  fromStatus?: IncidentStatus;
  toStatus: IncidentStatus;
  changedBy: string;
  changedAt: string;
  note?: string;
}
