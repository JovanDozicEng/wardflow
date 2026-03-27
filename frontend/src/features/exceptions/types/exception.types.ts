/**
 * Exception types and interfaces
 */

export type ExceptionStatus = 'draft' | 'finalized' | 'corrected';

export interface ExceptionEvent {
  id: string;
  encounterId: string;
  type: string;
  status: ExceptionStatus;
  requiredFields: Record<string, boolean>;
  data: Record<string, any>;
  initiatedBy: string;
  initiatedAt: string;
  finalizedBy?: string;
  finalizedAt?: string;
  correctedByEventId?: string;
  correctionReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateExceptionRequest {
  encounterId: string;
  type: string;
  data: Record<string, any>;
}

export interface UpdateExceptionRequest {
  data: Record<string, any>;
}

export interface CorrectExceptionRequest {
  reason: string;
  data: Record<string, any>;
}
