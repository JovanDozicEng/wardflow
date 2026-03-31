import api from '../../../shared/utils/api';
import type { Bed, BedRequest, BedStatusEvent, CreateBedRequest, UpdateBedStatusRequest, CreateBedRequestRequest, AssignBedRequest, ListBedsResponse } from '../types';

export const bedService = {
  listBeds: (params?: { unitId?: string; status?: string; limit?: number; offset?: number }): Promise<ListBedsResponse> =>
    api.get('/beds', { params }).then(r => r.data),

  getBed: (id: string): Promise<Bed> =>
    api.get(`/beds/${id}`).then(r => r.data),

  createBed: (data: CreateBedRequest): Promise<Bed> =>
    api.post('/beds', data).then(r => r.data),

  updateBedStatus: (bedId: string, data: UpdateBedStatusRequest): Promise<BedStatusEvent> =>
    api.post(`/beds/${bedId}/status`, data).then(r => r.data),

  createBedRequest: (encounterId: string, data: CreateBedRequestRequest): Promise<BedRequest> =>
    api.post(`/encounters/${encounterId}/bed-requests`, data).then(r => r.data),

  assignBed: (requestId: string, data: AssignBedRequest): Promise<BedRequest> =>
    api.post(`/bed-requests/${requestId}/assign`, data).then(r => r.data),
};
