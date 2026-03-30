export type TransportStatus = 'pending' | 'assigned' | 'in_transit' | 'completed' | 'cancelled';
export type TransportPriority = 'routine' | 'urgent' | 'emergent';

export interface TransportRequest {
  id: string;
  encounterId: string;
  origin: string;
  destination: string;
  priority: TransportPriority;
  status: TransportStatus;
  assignedTo?: string;
  assignedAt?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface TransportChangeEvent {
  id: string;
  requestId: string;
  changedFields: Record<string, unknown>;
  changedBy: string;
  reason?: string;
  changedAt: string;
  createdAt: string;
}

export interface CreateTransportRequest {
  encounterId: string;
  origin: string;
  destination: string;
  priority?: TransportPriority;
}

export interface UpdateTransportRequest {
  origin?: string;
  destination?: string;
  priority?: TransportPriority;
  reason?: string;
}

export interface AcceptTransportRequest {
  assignedTo: string;
}

export interface ListTransportResponse {
  data: TransportRequest[];
  total: number;
  limit: number;
  offset: number;
}
