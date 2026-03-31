export type ChecklistStatus = 'in_progress' | 'complete' | 'override_complete';
export type ItemStatus = 'open' | 'done' | 'waived';
export type DischargeType = 'standard' | 'ama' | 'lwbs';

export interface DischargeChecklist {
  id: string;
  encounterId: string;
  dischargeType: DischargeType;
  status: ChecklistStatus;
  completedBy?: string;
  completedAt?: string;
  overrideReason?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  items?: DischargeChecklistItem[];
}

export interface DischargeChecklistItem {
  id: string;
  checklistId: string;
  code: string;
  label: string;
  required: boolean;
  status: ItemStatus;
  completedBy?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface InitChecklistRequest {
  dischargeType: DischargeType;
}

export interface CompleteDischargeRequest {
  override?: boolean;
  reason?: string;
}
