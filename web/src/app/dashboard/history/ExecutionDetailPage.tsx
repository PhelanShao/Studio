/**
 * 📄 执行详情页面
 *
 * 职责：展示单次工作流执行的详细信息
 *
 * 功能：
 * 1. 展示执行概要信息
 * 2. 展示动作执行时间线
 * 3. 展示错误信息（如有）
 */

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useWorkflowExecutionDetail } from '@/hooks/queries/useHistoryQueries';
import type { ActionExecutionHistory, ExecutionStatus } from '@/types/history';
import { ExecutionStatusConfig, formatDateTime, formatDuration } from '@/types/history';
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  Clock,
  Cpu,
  Play,
  XCircle,
} from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';

// 状态徽章组件
function StatusBadge({ status }: { status: ExecutionStatus }) {
  const config = ExecutionStatusConfig[status];
  return (
    <Badge variant="outline" className={`${config.color} ${config.bgColor} border-0`}>
      {config.label}
    </Badge>
  );
}

// 状态图标
function StatusIcon({ status }: { status: ExecutionStatus }) {
  const iconClass = 'h-5 w-5';
  switch (status) {
    case 'success':
      return <CheckCircle2 className={`${iconClass} text-green-500`} />;
    case 'failed':
      return <XCircle className={`${iconClass} text-red-500`} />;
    case 'running':
      return <Play className={`${iconClass} text-blue-500`} />;
    case 'pending':
      return <Clock className={`${iconClass} text-yellow-500`} />;
    default:
      return <AlertCircle className={`${iconClass} text-gray-500`} />;
  }
}

// 动作时间线项
function ActionTimelineItem({ action, index }: { action: ActionExecutionHistory; index: number }) {
  return (
    <div className="flex gap-4">
      <div className="flex flex-col items-center">
        <div className="flex h-10 w-10 items-center justify-center rounded-full border-2 bg-background">
          <StatusIcon status={action.status} />
        </div>
        {index < 999 && <div className="w-px flex-1 bg-border" />}
      </div>
      <div className="flex-1 pb-8">
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Cpu className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-base">{action.action_name}</CardTitle>
              </div>
              <StatusBadge status={action.status} />
            </div>
            <CardDescription>
              设备: {action.device_name} | 类型: {action.action_type}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDuration(action.duration_ms)}
              </span>
              <span>{formatDateTime(action.created_at)}</span>
            </div>
            {action.error_message && (
              <div className="mt-2 p-2 bg-red-50 dark:bg-red-950 rounded text-sm text-red-600 dark:text-red-400">
                {action.error_message}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function ExecutionDetailPage() {
  const { executionId } = useParams<{ executionId: string }>();
  const navigate = useNavigate();

  const { data: execution, isLoading, error } = useWorkflowExecutionDetail(
    executionId || ''
  );

  if (isLoading) {
    return (
      <div className="container mx-auto py-6 space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error || !execution) {
    return (
      <div className="container mx-auto py-6">
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            执行记录不存在或加载失败
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* 返回按钮 */}
      <Button variant="ghost" onClick={() => navigate(-1)}>
        <ArrowLeft className="mr-2 h-4 w-4" />
        返回
      </Button>

      {/* 执行概要 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-xl">{execution.workflow_name}</CardTitle>
              <CardDescription>执行 ID: {execution.uuid}</CardDescription>
            </div>
            <StatusBadge status={execution.status} />
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div>
              <p className="text-sm text-muted-foreground">开始时间</p>
              <p className="font-medium">{formatDateTime(execution.started_at)}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">完成时间</p>
              <p className="font-medium">
                {execution.completed_at ? formatDateTime(execution.completed_at) : '-'}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">执行耗时</p>
              <p className="font-medium">{formatDuration(execution.duration_ms)}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">步骤进度</p>
              <p className="font-medium">
                {execution.steps_completed} / {execution.steps_total}
                {execution.steps_failed > 0 && (
                  <span className="text-red-500 ml-1">
                    ({execution.steps_failed} 失败)
                  </span>
                )}
              </p>
            </div>
          </div>

          {/* 错误信息 */}
          {execution.error_message && (
            <div className="mt-4 p-4 bg-red-50 dark:bg-red-950 rounded-lg">
              <div className="flex items-center gap-2 text-red-600 dark:text-red-400 font-medium mb-1">
                <AlertCircle className="h-4 w-4" />
                错误信息
              </div>
              <p className="text-sm text-red-600 dark:text-red-400">
                {execution.error_message}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 动作执行时间线 */}
      <Card>
        <CardHeader>
          <CardTitle>动作执行时间线</CardTitle>
          <CardDescription>
            共 {execution.actions?.length || 0} 个动作
          </CardDescription>
        </CardHeader>
        <CardContent>
          {execution.actions && execution.actions.length > 0 ? (
            <div className="space-y-0">
              {execution.actions.map((action, index) => (
                <ActionTimelineItem
                  key={action.uuid}
                  action={action}
                  index={index}
                />
              ))}
            </div>
          ) : (
            <p className="text-center text-muted-foreground py-8">
              暂无动作执行记录
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

