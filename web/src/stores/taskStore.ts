import { create } from 'zustand';
import type { Task, TaskListResponse, TaskFilter, TaskRequest } from '../types';
import { api } from '../services/api';

interface TaskStore {
  tasks: Task[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  loading: boolean;
  error: string | null;
  selectedTask: Task | null;

  // Actions
  fetchTasks: (filter?: TaskFilter) => Promise<void>;
  fetchTask: (taskId: string) => Promise<void>;
  createTask: (request: TaskRequest) => Promise<Task>;
  deleteTask: (taskId: string) => Promise<void>;
  runTask: (taskId: string) => Promise<void>;
  setSelectedTask: (task: Task | null) => void;
  clearError: () => void;
}

export const useTaskStore = create<TaskStore>((set) => ({
  tasks: [],
  total: 0,
  page: 1,
  pageSize: 10,
  totalPages: 0,
  loading: false,
  error: null,
  selectedTask: null,

  fetchTasks: async (filter?: TaskFilter) => {
    set({ loading: true, error: null });
    try {
      const response: TaskListResponse = await api.getTasks(filter);
      set({
        tasks: response.tasks,
        total: response.total,
        page: response.page,
        pageSize: response.page_size,
        totalPages: response.total_pages,
        loading: false,
      });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch tasks',
        loading: false
      });
    }
  },

  fetchTask: async (taskId: string) => {
    set({ loading: true, error: null });
    try {
      const task = await api.getTask(taskId);
      set({ selectedTask: task, loading: false });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch task',
        loading: false
      });
    }
  },

  createTask: async (request: TaskRequest) => {
    set({ loading: true, error: null });
    try {
      const task = await api.createTask(request);
      set((state) => ({
        tasks: [task, ...state.tasks],
        total: state.total + 1,
        loading: false,
      }));
      return task;
    } catch (error: any) {
      set({
        error: error.message || 'Failed to create task',
        loading: false
      });
      throw error;
    }
  },

  deleteTask: async (taskId: string) => {
    set({ loading: true, error: null });
    try {
      await api.deleteTask(taskId);
      set((state) => ({
        tasks: state.tasks.filter((t) => t.id !== taskId),
        total: state.total - 1,
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to delete task',
        loading: false
      });
      throw error;
    }
  },

  runTask: async (taskId: string) => {
    set({ loading: true, error: null });
    try {
      await api.runTask(taskId);
      // Update task status to running
      set((state) => ({
        tasks: state.tasks.map((t) =>
          t.id === taskId ? { ...t, status: 'running' as const } : t
        ),
        loading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to run task',
        loading: false
      });
      throw error;
    }
  },

  setSelectedTask: (task) => set({ selectedTask: task }),
  clearError: () => set({ error: null }),
}));
