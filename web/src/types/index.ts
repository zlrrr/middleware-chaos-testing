// Task types
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface Task {
  id: string;
  middleware: string;
  status: TaskStatus;
  start_time: string;
  end_time?: string;
  config: string;
  result_path?: string;
  error_msg?: string;
  created_at: string;
  updated_at: string;
}

export interface TaskRequest {
  middleware: string;
  config: Record<string, string>;
  duration: string;
  operations?: number;
  concurrency?: number;
}

export interface TaskFilter {
  middleware?: string;
  status?: TaskStatus;
  page?: number;
  page_size?: number;
}

export interface TaskListResponse {
  tasks: Task[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Test result types
export interface DimensionScores {
  availability: number;
  performance: number;
  reliability: number;
  resilience: number;
}

export interface StabilityMetrics {
  total_operations: number;
  successful_operations: number;
  failed_operations: number;
  availability: number;
  total_connection_attempts: number;
  successful_connection_attempts: number;
  connection_success_rate: number;
  p50_latency: number;
  p95_latency: number;
  p99_latency: number;
  avg_latency: number;
  max_latency: number;
  min_latency: number;
  throughput: number;
  error_rate: number;
  data_loss_rate: number;
  data_consistency: number;
  duplicate_rate: number;
  mtbf: number;
  mttr: number;
  total_reconnect_attempts: number;
  successful_reconnects: number;
  reconnect_success_rate: number;
}

export interface Issue {
  type: string;
  severity: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';
  metric: string;
  current: number;
  expected: number;
  message: string;
}

export interface Recommendation {
  priority: 'HIGH' | 'MEDIUM' | 'LOW';
  category: 'CONFIGURATION' | 'SCALING' | 'OPTIMIZATION';
  title: string;
  message: string;
  actions: string[];
  references: string[];
}

export interface TestResult {
  task_id: string;
  middleware: string;
  score: number;
  grade: string;
  status: string;
  metrics: StabilityMetrics;
  issues: Issue[];
  recommendations: Recommendation[];
  rationale: string;
  start_time: string;
  end_time: string;
  duration: number;
  scores: DimensionScores;
}

// Middleware info types
export interface ConfigField {
  type: string;
  required: boolean;
  default?: any;
  description: string;
  example?: string;
}

export interface MiddlewareInfo {
  name: string;
  display_name: string;
  description: string;
  config_spec: Record<string, ConfigField>;
}

// API response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}
