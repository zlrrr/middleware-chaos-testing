import axios, { AxiosInstance } from 'axios';
import type {
  ApiResponse,
  Task,
  TaskRequest,
  TaskListResponse,
  TaskFilter,
  TestResult,
  MiddlewareInfo,
} from '../types';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('API Error:', error);
        return Promise.reject(error);
      }
    );
  }

  // Task API
  async createTask(request: TaskRequest): Promise<Task> {
    const { data } = await this.client.post<ApiResponse<Task>>('/tasks', request);
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to create task');
    }
    return data.data;
  }

  async getTasks(filter?: TaskFilter): Promise<TaskListResponse> {
    const params: Record<string, any> = {};
    if (filter) {
      if (filter.middleware) params.middleware = filter.middleware;
      if (filter.status) params.status = filter.status;
      if (filter.page) params.page = filter.page;
      if (filter.page_size) params.page_size = filter.page_size;
    }

    const { data } = await this.client.get<ApiResponse<TaskListResponse>>('/tasks', { params });
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to fetch tasks');
    }
    return data.data;
  }

  async getTask(taskId: string): Promise<Task> {
    const { data } = await this.client.get<ApiResponse<Task>>(`/tasks/${taskId}`);
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to fetch task');
    }
    return data.data;
  }

  async deleteTask(taskId: string): Promise<void> {
    const { data } = await this.client.delete<ApiResponse<void>>(`/tasks/${taskId}`);
    if (!data.success) {
      throw new Error(data.error || 'Failed to delete task');
    }
  }

  async runTask(taskId: string): Promise<void> {
    const { data } = await this.client.post<ApiResponse<void>>(`/tasks/${taskId}/run`);
    if (!data.success) {
      throw new Error(data.error || 'Failed to run task');
    }
  }

  async getTaskResult(taskId: string): Promise<TestResult> {
    const { data } = await this.client.get<ApiResponse<TestResult>>(`/tasks/${taskId}/result`);
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to fetch task result');
    }
    return data.data;
  }

  // Middleware API
  async getMiddlewares(): Promise<MiddlewareInfo[]> {
    const { data } = await this.client.get<ApiResponse<MiddlewareInfo[]>>('/middlewares');
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to fetch middlewares');
    }
    return data.data;
  }

  async getMiddlewareConfig(name: string): Promise<MiddlewareInfo> {
    const { data } = await this.client.get<ApiResponse<MiddlewareInfo>>(`/middlewares/${name}/config`);
    if (!data.success || !data.data) {
      throw new Error(data.error || 'Failed to fetch middleware config');
    }
    return data.data;
  }

  // Health check
  async checkHealth(): Promise<boolean> {
    try {
      const { data } = await this.client.get('/health');
      return data.status === 'healthy';
    } catch {
      return false;
    }
  }
}

export const api = new ApiClient();
