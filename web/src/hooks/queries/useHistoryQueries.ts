/**
 * 🎣 Query Hook Layer - History 数据查询
 *
 * 职责：
 * 1. 封装 historyService 的 HTTP 请求
 * 2. 提供 TanStack Query 缓存策略
 * 3. 管理服务器状态（执行历史、事件、统计等）
 */

import { historyService } from '@/service/historyService';
import type {
  DeviceEventQueryParams,
  LabStatsQueryParams,
  WorkflowExecutionQueryParams,
} from '@/types/history';
import { useQuery } from '@tanstack/react-query';

// ============= Query Keys =============
export const historyKeys = {
  all: ['history'] as const,
  
  // 工作流执行历史
  workflowExecutions: () => [...historyKeys.all, 'workflow-executions'] as const,
  workflowExecutionList: (params: WorkflowExecutionQueryParams) =>
    [...historyKeys.workflowExecutions(), params] as const,
  workflowExecutionDetail: (uuid: string) =>
    [...historyKeys.workflowExecutions(), 'detail', uuid] as const,
  
  // 设备事件历史
  deviceEvents: () => [...historyKeys.all, 'device-events'] as const,
  deviceEventList: (params: DeviceEventQueryParams) =>
    [...historyKeys.deviceEvents(), params] as const,
  
  // 实验室统计
  labStats: (labId: number, params?: LabStatsQueryParams) =>
    [...historyKeys.all, 'stats', labId, params] as const,
};

// ============= Query Hooks =============

/**
 * 获取工作流执行历史列表
 */
export function useWorkflowExecutions(
  params: WorkflowExecutionQueryParams,
  enabled = true
) {
  return useQuery({
    queryKey: historyKeys.workflowExecutionList(params),
    queryFn: () => historyService.getWorkflowExecutions(params),
    enabled: !!params.lab_id && enabled,
    staleTime: 30000, // 30秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后垃圾回收
    select: (data) => data?.data,
  });
}

/**
 * 获取工作流执行详情（包含动作列表）
 */
export function useWorkflowExecutionDetail(executionUuid: string, enabled = true) {
  return useQuery({
    queryKey: historyKeys.workflowExecutionDetail(executionUuid),
    queryFn: () => historyService.getWorkflowExecutionDetail(executionUuid),
    enabled: !!executionUuid && enabled,
    staleTime: 60000, // 1分钟
    select: (data) => data?.data,
  });
}

/**
 * 获取设备事件历史列表
 */
export function useDeviceEvents(
  params: DeviceEventQueryParams,
  enabled = true
) {
  return useQuery({
    queryKey: historyKeys.deviceEventList(params),
    queryFn: () => historyService.getDeviceEvents(params),
    enabled: !!params.lab_id && enabled,
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
    select: (data) => data?.data,
  });
}

/**
 * 获取实验室使用统计
 */
export function useLabStats(
  labId: number,
  params?: LabStatsQueryParams,
  enabled = true
) {
  return useQuery({
    queryKey: historyKeys.labStats(labId, params),
    queryFn: () => historyService.getLabStats(labId, params),
    enabled: !!labId && enabled,
    staleTime: 60000, // 1分钟
    gcTime: 10 * 60 * 1000, // 10分钟后垃圾回收
    select: (data) => data?.data,
  });
}

/**
 * 获取工作流执行历史（带自动刷新）
 * 用于实时监控场景
 */
export function useWorkflowExecutionsLive(
  params: WorkflowExecutionQueryParams,
  refetchInterval = 10000 // 默认10秒刷新
) {
  return useQuery({
    queryKey: historyKeys.workflowExecutionList(params),
    queryFn: () => historyService.getWorkflowExecutions(params),
    enabled: !!params.lab_id,
    staleTime: 5000,
    refetchInterval,
    select: (data) => data?.data,
  });
}

