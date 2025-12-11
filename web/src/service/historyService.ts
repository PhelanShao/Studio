import { config } from '@/configs';
import apiClient from '@/service/http/client';
import type {
  ApiResponse,
  DeviceEventHistory,
  DeviceEventQueryParams,
  LabStats,
  LabStatsQueryParams,
  PaginatedResponse,
  WorkflowExecutionDetail,
  WorkflowExecutionHistory,
  WorkflowExecutionQueryParams,
} from '@/types/history';

// Event History 相关服务
export const historyService = {
  // ========== 工作流执行历史 ==========

  /**
   * 获取工作流执行历史列表
   */
  async getWorkflowExecutions(
    params: WorkflowExecutionQueryParams
  ): Promise<ApiResponse<PaginatedResponse<WorkflowExecutionHistory>>> {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/history/workflow`,
      { params }
    );
    return res.data;
  },

  /**
   * 获取工作流执行详情（包含动作列表）
   */
  async getWorkflowExecutionDetail(
    executionUuid: string
  ): Promise<ApiResponse<WorkflowExecutionDetail>> {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/history/workflow/execution/${executionUuid}`
    );
    return res.data;
  },

  // ========== 设备事件历史 ==========

  /**
   * 获取设备事件历史列表
   */
  async getDeviceEvents(
    params: DeviceEventQueryParams
  ): Promise<ApiResponse<PaginatedResponse<DeviceEventHistory>>> {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/history/device`,
      { params }
    );
    return res.data;
  },

  // ========== 实验室统计 ==========

  /**
   * 获取实验室使用统计
   */
  async getLabStats(
    labId: number,
    params?: LabStatsQueryParams
  ): Promise<ApiResponse<LabStats>> {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/${labId}/stats`,
      { params }
    );
    return res.data;
  },
};

export default historyService;

