import api from '../../../shared/utils/api';
import type { TransportRequest, CreateTransportRequest, UpdateTransportRequest, AcceptTransportRequest, ListTransportResponse } from '../types';

export const transportService = {
  listRequests: (params?: { status?: string; unitId?: string; limit?: number; offset?: number }): Promise<ListTransportResponse> =>
    api.get('/transport/requests', { params }).then(r => r.data),

  createRequest: (data: CreateTransportRequest): Promise<TransportRequest> =>
    api.post('/transport/requests', data).then(r => r.data),

  acceptRequest: (id: string, data: AcceptTransportRequest): Promise<TransportRequest> =>
    api.post(`/transport/requests/${id}/accept`, data).then(r => r.data),

  updateRequest: (id: string, data: UpdateTransportRequest): Promise<TransportRequest> =>
    api.patch(`/transport/requests/${id}`, data).then(r => r.data),

  completeRequest: (id: string): Promise<TransportRequest> =>
    api.post(`/transport/requests/${id}/complete`, {}).then(r => r.data),
};
