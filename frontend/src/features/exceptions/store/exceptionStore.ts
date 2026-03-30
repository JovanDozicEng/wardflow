/**
 * Exception Zustand store
 */

import { create } from 'zustand';
import type { ExceptionEvent } from '../types/exception.types';

interface ExceptionStore {
  exceptions: ExceptionEvent[];
  loading: boolean;
  error: string | null;
  setExceptions: (exceptions: ExceptionEvent[]) => void;
  addException: (exception: ExceptionEvent) => void;
  updateException: (id: string, updates: Partial<ExceptionEvent>) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useExceptionStore = create<ExceptionStore>((set) => ({
  exceptions: [],
  loading: false,
  error: null,
  setExceptions: (exceptions) => set({ exceptions }),
  addException: (exception) => set((state) => ({ exceptions: [exception, ...state.exceptions] })),
  updateException: (id, updates) =>
    set((state) => ({
      exceptions: state.exceptions.map((e) => (e.id === id ? { ...e, ...updates } : e)),
    })),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
}));
