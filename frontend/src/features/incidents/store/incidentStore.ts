/**
 * Incident Zustand store
 */

import { create } from 'zustand';
import type { Incident, IncidentStatusEvent } from '../types/incident.types';

interface IncidentStore {
  incidents: Incident[];
  selectedIncident: Incident | null;
  statusHistory: IncidentStatusEvent[];
  loading: boolean;
  error: string | null;
  setIncidents: (incidents: Incident[]) => void;
  addIncident: (incident: Incident) => void;
  updateIncident: (id: string, updates: Partial<Incident>) => void;
  setSelectedIncident: (incident: Incident | null) => void;
  setStatusHistory: (history: IncidentStatusEvent[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useIncidentStore = create<IncidentStore>((set) => ({
  incidents: [],
  selectedIncident: null,
  statusHistory: [],
  loading: false,
  error: null,
  setIncidents: (incidents) => set({ incidents }),
  addIncident: (incident) => set((state) => ({ incidents: [incident, ...state.incidents] })),
  updateIncident: (id, updates) =>
    set((state) => ({
      incidents: state.incidents.map((i) => (i.id === id ? { ...i, ...updates } : i)),
      selectedIncident:
        state.selectedIncident?.id === id
          ? { ...state.selectedIncident, ...updates }
          : state.selectedIncident,
    })),
  setSelectedIncident: (incident) => set({ selectedIncident: incident }),
  setStatusHistory: (history) => set({ statusHistory: history }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
}));
