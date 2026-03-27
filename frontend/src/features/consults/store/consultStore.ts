/**
 * Consult Zustand store
 */

import { create } from 'zustand';
import type { ConsultRequest } from '../types/consult.types';

interface ConsultStore {
  consults: ConsultRequest[];
  loading: boolean;
  error: string | null;
  setConsults: (consults: ConsultRequest[]) => void;
  addConsult: (consult: ConsultRequest) => void;
  updateConsult: (id: string, updates: Partial<ConsultRequest>) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useConsultStore = create<ConsultStore>((set) => ({
  consults: [],
  loading: false,
  error: null,
  setConsults: (consults) => set({ consults }),
  addConsult: (consult) => set((state) => ({ consults: [consult, ...state.consults] })),
  updateConsult: (id, updates) =>
    set((state) => ({
      consults: state.consults.map((c) => (c.id === id ? { ...c, ...updates } : c)),
    })),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
}));
