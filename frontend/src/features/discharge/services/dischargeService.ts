import api from '../../../shared/utils/api';
import type { DischargeChecklist, InitChecklistRequest, CompleteDischargeRequest } from '../types';

export const dischargeService = {
  initChecklist: (encounterId: string, data: InitChecklistRequest): Promise<DischargeChecklist> =>
    api.post(`/encounters/${encounterId}/discharge-checklist/init`, data).then(r => r.data),

  getChecklist: (encounterId: string): Promise<DischargeChecklist> =>
    api.get(`/encounters/${encounterId}/discharge-checklist`).then(r => r.data),

  completeItem: (itemId: string): Promise<DischargeChecklist> =>
    api.post(`/discharge-checklist/items/${itemId}/complete`, {}).then(r => r.data),

  completeDischarge: (encounterId: string, data?: CompleteDischargeRequest): Promise<DischargeChecklist> =>
    api.post(`/encounters/${encounterId}/discharge/complete`, data ?? {}).then(r => r.data),
};
