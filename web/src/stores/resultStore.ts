import { create } from 'zustand';
import type { TestResult } from '../types';
import { api } from '../services/api';

interface ResultStore {
  results: Record<string, TestResult>;
  loading: boolean;
  error: string | null;

  // Actions
  fetchResult: (taskId: string) => Promise<void>;
  getResult: (taskId: string) => TestResult | null;
  clearError: () => void;
}

export const useResultStore = create<ResultStore>((set, get) => ({
  results: {},
  loading: false,
  error: null,

  fetchResult: async (taskId: string) => {
    set({ loading: true, error: null });
    try {
      const result = await api.getTaskResult(taskId);
      set((state) => ({
        results: { ...state.results, [taskId]: result },
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch result',
        loading: false
      });
    }
  },

  getResult: (taskId: string) => {
    return get().results[taskId] || null;
  },

  clearError: () => set({ error: null }),
}));
