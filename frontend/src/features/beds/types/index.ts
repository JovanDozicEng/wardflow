export type BedStatus = 'available' | 'occupied' | 'blocked' | 'cleaning' | 'maintenance';
export type BedRequestStatus = 'pending' | 'assigned' | 'cancelled';
export type BedPriority = 'routine' | 'urgent' | 'emergent';

export interface Bed {
  id: string;
  unitId: string;
  room: string;
  label: string;
  capabilities: string[];
  currentStatus: BedStatus;
  currentEncounterId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface BedStatusEvent {
  id: string;
  bedId: string;
  fromStatus?: BedStatus;
  toStatus: BedStatus;
  reason?: string;
  changedBy: string;
  changedAt: string;
  createdAt: string;
}

export interface BedRequest {
  id: string;
  encounterId: string;
  requiredCapabilities: string[];
  priority: BedPriority;
  status: BedRequestStatus;
  assignedBedId?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBedRequest {
  unitId: string;
  room: string;
  label: string;
  capabilities?: string[];
}

export interface UpdateBedStatusRequest {
  status: BedStatus;
  reason?: string;
}

export interface CreateBedRequestRequest {
  requiredCapabilities?: string[];
  priority?: BedPriority;
}

export interface AssignBedRequest {
  bedId: string;
}

export interface ListBedsResponse {
  data: Bed[];
  total: number;
  limit: number;
  offset: number;
}
