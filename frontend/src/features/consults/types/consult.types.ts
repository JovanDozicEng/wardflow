/**
 * Consult types and interfaces
 */

export type ConsultUrgency = 'routine' | 'urgent' | 'emergent';
export type ConsultStatus = 'pending' | 'accepted' | 'declined' | 'completed' | 'redirected' | 'cancelled';

export interface ConsultRequest {
  id: string;
  encounterId: string;
  targetService: string;
  reason: string;
  urgency: ConsultUrgency;
  status: ConsultStatus;
  createdBy: string;
  createdAt: string;
  acceptedBy?: string;
  acceptedAt?: string;
  closedAt?: string;
  closeReason?: string;
  redirectedTo?: string;
  updatedAt: string;
}

export interface CreateConsultRequest {
  encounterId: string;
  targetService: string;
  reason: string;
  urgency: ConsultUrgency;
}

export interface DeclineConsultRequest {
  reason: string;
}

export interface RedirectConsultRequest {
  targetService: string;
  reason: string;
}
