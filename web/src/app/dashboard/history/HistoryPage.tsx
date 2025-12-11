/**
 * 📄 History 页面
 *
 * 职责：展示实验室的执行历史和统计
 *
 * 功能：
 * 1. 展示工作流执行历史列表
 * 2. 展示设备事件历史
 * 3. 展示实验室统计数据
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useLabStats, useWorkflowExecutions } from '@/hooks/queries/useHistoryQueries';
import type { ExecutionStatus, WorkflowExecutionHistory } from '@/types/history';
import { ExecutionStatusConfig, formatDateTime, formatDuration } from '@/types/history';
import {
  Activity,
  ArrowRight,
  BarChart3,
  CheckCircle2,
  Clock,
  XCircle,
} from 'lucide-react';
import { useState } from 'react';
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

// 统计卡片组件
function StatsCard({
  title,
  value,
  subtitle,
  icon: Icon,
  color = 'text-primary',
}: {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ElementType;
  color?: string;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className={`h-4 w-4 ${color}`} />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && (
          <p className="text-xs text-muted-foreground">{subtitle}</p>
        )}
      </CardContent>
    </Card>
  );
}

export default function HistoryPage() {
  const { labId } = useParams<{ labId: string }>();
  const navigate = useNavigate();
  const labIdNum = labId ? parseInt(labId, 10) : 0;

  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const pageSize = 10;

  // 获取统计数据
  const { data: stats, isLoading: statsLoading } = useLabStats(labIdNum);

  // 获取执行历史
  const { data: executionsData, isLoading: executionsLoading } = useWorkflowExecutions({
    lab_id: labIdNum,
    status: statusFilter !== 'all' ? (statusFilter as ExecutionStatus) : undefined,
    page,
    page_size: pageSize,
  });

  const handleViewDetail = (execution: WorkflowExecutionHistory) => {
    navigate(`/dashboard/history/execution/${execution.uuid}`);
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">执行历史</h1>
          <p className="text-muted-foreground">查看工作流执行和设备事件历史</p>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-4">
        {statsLoading ? (
          <>
            {[1, 2, 3, 4].map((i) => (
              <Card key={i}>
                <CardHeader className="pb-2">
                  <Skeleton className="h-4 w-24" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-8 w-16" />
                </CardContent>
              </Card>
            ))}
          </>
        ) : stats ? (
          <>
            <StatsCard
              title="总执行次数"
              value={stats.total_executions}
              icon={Activity}
            />
            <StatsCard
              title="成功率"
              value={`${stats.success_rate.toFixed(1)}%`}
              subtitle={`${stats.successful_count} 次成功`}
              icon={CheckCircle2}
              color="text-green-500"
            />
            <StatsCard
              title="失败次数"
              value={stats.failed_count}
              icon={XCircle}
              color="text-red-500"
            />
            <StatsCard
              title="平均耗时"
              value={formatDuration(stats.average_duration_ms)}
              icon={Clock}
            />
          </>
        ) : null}
      </div>

      {/* 标签页 */}
      <Tabs defaultValue="workflows" className="space-y-4">
        <TabsList>
          <TabsTrigger value="workflows" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            工作流执行
          </TabsTrigger>
        </TabsList>

        <TabsContent value="workflows" className="space-y-4">
          {/* 过滤器 */}
          <div className="flex items-center gap-4">
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="筛选状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="success">成功</SelectItem>
                <SelectItem value="failed">失败</SelectItem>
                <SelectItem value="running">运行中</SelectItem>
                <SelectItem value="pending">等待中</SelectItem>
                <SelectItem value="cancelled">已取消</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* 执行历史表格 */}
          <Card>
            <CardHeader>
              <CardTitle>执行记录</CardTitle>
              <CardDescription>
                共 {executionsData?.total || 0} 条记录
              </CardDescription>
            </CardHeader>
            <CardContent>
              {executionsLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3].map((i) => (
                    <Skeleton key={i} className="h-12 w-full" />
                  ))}
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>工作流</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>进度</TableHead>
                      <TableHead>耗时</TableHead>
                      <TableHead>开始时间</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {executionsData?.items?.map((execution) => (
                      <TableRow key={execution.uuid}>
                        <TableCell className="font-medium">
                          {execution.workflow_name}
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={execution.status} />
                        </TableCell>
                        <TableCell>
                          {execution.steps_completed}/{execution.steps_total}
                          {execution.steps_failed > 0 && (
                            <span className="text-red-500 ml-1">
                              ({execution.steps_failed} 失败)
                            </span>
                          )}
                        </TableCell>
                        <TableCell>{formatDuration(execution.duration_ms)}</TableCell>
                        <TableCell>{formatDateTime(execution.started_at)}</TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleViewDetail(execution)}
                          >
                            详情 <ArrowRight className="ml-1 h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                    {(!executionsData?.items || executionsData.items.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                          暂无执行记录
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}

              {/* 分页 */}
              {executionsData && executionsData.total_pages > 1 && (
                <div className="flex items-center justify-end space-x-2 mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    上一页
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    第 {page} / {executionsData.total_pages} 页
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={page >= executionsData.total_pages}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

