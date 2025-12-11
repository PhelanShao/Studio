// Event History 相关类型定义

// 执行状态枚举
export type ExecutionStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled' | 'timeout';

// 设备事件类型枚举
export type DeviceEventType = 
  | 'status_change' 
  | 'data_received' 
  | 'error' 
  | 'connected' 
  | 'disconnected' 
  | 'command_sent' 
  | 'command_result';

// 工作流执行历史
export interface WorkflowExecutionHistory {
  uuid: string;
  workflow_uuid: string;
  workflow_name: string;
  status: ExecutionStatus;
  steps_total: number;
  steps_completed: number;
  steps_failed: number;
  duration_ms: number;
  error_message?: string;
  started_at: string;
  completed_at?: string;
}

// 工作流执行详情（包含动作列表）
export interface WorkflowExecutionDetail extends WorkflowExecutionHistory {
  actions: ActionExecutionHistory[];
}

// 动作执行历史
export interface ActionExecutionHistory {
  uuid: string;
  device_uuid: string;
  device_name: string;
  action_type: string;
  action_name: string;
  status: ExecutionStatus;
  duration_ms: number;
  error_message?: string;
  created_at: string;
}

// 设备事件历史
export interface DeviceEventHistory {
  uuid: string;
  device_uuid: string;
  event_type: DeviceEventType;
  event_data: Record<string, unknown>;
  timestamp: string;
}

// 实验室统计数据
export interface LabStats {
  total_executions: number;
  successful_count: number;
  failed_count: number;
  success_rate: number;
  average_duration_ms: number;
  total_actions_count: number;
  total_device_events: number;
}

// 分页响应
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// API 响应包装
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 工作流执行历史查询参数
export interface WorkflowExecutionQueryParams {
  lab_id: number;
  workflow_id?: number;
  status?: ExecutionStatus;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

// 设备事件查询参数
export interface DeviceEventQueryParams {
  lab_id: number;
  device_id?: number;
  event_type?: DeviceEventType;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

// 实验室统计查询参数
export interface LabStatsQueryParams {
  start_time?: string;
  end_time?: string;
}

// 状态显示配置
export const ExecutionStatusConfig: Record<ExecutionStatus, { label: string; color: string; bgColor: string }> = {
  pending: { label: '等待中', color: 'text-yellow-600', bgColor: 'bg-yellow-100' },
  running: { label: '运行中', color: 'text-blue-600', bgColor: 'bg-blue-100' },
  success: { label: '成功', color: 'text-green-600', bgColor: 'bg-green-100' },
  failed: { label: '失败', color: 'text-red-600', bgColor: 'bg-red-100' },
  cancelled: { label: '已取消', color: 'text-gray-600', bgColor: 'bg-gray-100' },
  timeout: { label: '超时', color: 'text-orange-600', bgColor: 'bg-orange-100' },
};

// 事件类型显示配置
export const DeviceEventTypeConfig: Record<DeviceEventType, { label: string; icon: string }> = {
  status_change: { label: '状态变更', icon: 'RefreshCw' },
  data_received: { label: '数据接收', icon: 'Download' },
  error: { label: '错误', icon: 'AlertCircle' },
  connected: { label: '已连接', icon: 'Link' },
  disconnected: { label: '已断开', icon: 'Unlink' },
  command_sent: { label: '命令发送', icon: 'Send' },
  command_result: { label: '命令结果', icon: 'CheckCircle' },
};

// 格式化持续时间
export function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

// 格式化时间
export function formatDateTime(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

