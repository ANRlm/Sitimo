'use client';

import { useEffect, useMemo, useState } from 'react';
import { AlertCircle, Download, FileCode2, FileText, LoaderCircle, Search, Trash2, X } from 'lucide-react';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { EXPORT_STATUS_FILTER_OPTIONS } from '@/lib/catalogs';
import { getExportDownloadUrl } from '@/lib/api/exports';
import { formatAbsoluteDateTime } from '@/lib/format';
import { useDeleteExport, useExports } from '@/lib/hooks/use-exports';
import type { ExportJob } from '@/lib/types';

export default function ExportsPage() {
  const [query, setQuery] = useState('');
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [formatFilter, setFormatFilter] = useState('all');
  const [dateFilter, setDateFilter] = useState('all');
  const now = useCurrentTimestamp();

  const exportsQuery = useExports({ page: 1, pageSize: 100 });
  const deleteExportMutation = useDeleteExport();

  const rows = useMemo(() => exportsQuery.data?.items ?? [], [exportsQuery.data?.items]);

  const filteredRows = useMemo(
    () =>
      rows.filter((row) => {
        const matchesQuery = !query.trim() || row.paperTitle.toLowerCase().includes(query.trim().toLowerCase());
        const matchesStatus = selectedStatuses.length === 0 || selectedStatuses.includes(row.status);
        const matchesFormat = formatFilter === 'all' || row.format === formatFilter;
        const age = now - new Date(row.createdAt).getTime();
        const matchesDate =
          dateFilter === 'all' ||
          (dateFilter === 'today' && age <= 24 * 60 * 60 * 1000) ||
          (dateFilter === 'week' && age <= 7 * 24 * 60 * 60 * 1000) ||
          (dateFilter === 'month' && age <= 30 * 24 * 60 * 60 * 1000);

        return matchesQuery && matchesStatus && matchesFormat && matchesDate;
      }),
    [dateFilter, formatFilter, now, query, rows, selectedStatuses]
  );

  const activeRow = filteredRows.find((row) => row.status === 'processing') ?? filteredRows.find((row) => row.status === 'pending');

  const toggleStatus = (status: string) => {
    setSelectedStatuses((current) => (current.includes(status) ? current.filter((item) => item !== status) : [...current, status]));
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="导出历史"
        title="导出任务"
        description="统一查看 PDF 与 LaTeX 导出状态，失败原因、取消和清理动作都在明确的确认或详情视图中完成。"
        badges={<Badge variant="secondary">共 {filteredRows.length} 条结果</Badge>}
      />

      {activeRow ? (
        <PagePanel className="border-primary/20 bg-primary/5">
          <div className="flex items-center justify-between gap-4 p-5">
            <div className="flex-1">
              <p className="font-medium">
                {activeRow.status === 'processing' ? '正在生成' : '正在排队'}《{activeRow.paperTitle}》
              </p>
              <div className="mt-3 flex items-center gap-3">
                <Progress value={activeRow.progress ?? 0} className="flex-1" />
                <span className="text-sm text-muted-foreground">{activeRow.progress ?? 0}%</span>
              </div>
              <p className="mt-2 text-xs text-muted-foreground">已用时 {formatElapsed(activeRow, now)}</p>
            </div>
            <ConfirmActionButton
              variant="ghost"
              size="icon"
              onConfirm={() => deleteExportMutation.mutateAsync(activeRow.id)}
              pending={deleteExportMutation.isPending}
              title={activeRow.status === 'processing' ? '确认取消导出任务' : '确认移除排队任务'}
              description={`确定要终止《${activeRow.paperTitle}》的当前导出任务吗？`}
              confirmLabel="确认终止"
            >
              <X className="h-4 w-4" />
            </ConfirmActionButton>
          </div>
        </PagePanel>
      ) : null}

      <PageToolbar className="flex flex-col gap-3">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center">
          <div className="relative min-w-[260px] flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索任务标题..." className="pl-9" />
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Select value={formatFilter} onValueChange={setFormatFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="格式" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部格式</SelectItem>
                <SelectItem value="pdf">PDF</SelectItem>
                <SelectItem value="latex">LaTeX</SelectItem>
              </SelectContent>
            </Select>

            <Select value={dateFilter} onValueChange={setDateFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="日期" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部日期</SelectItem>
                <SelectItem value="today">今天</SelectItem>
                <SelectItem value="week">近 7 天</SelectItem>
                <SelectItem value="month">近 30 天</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {EXPORT_STATUS_FILTER_OPTIONS.map((status) => (
            <button
              key={status.value}
              type="button"
              onClick={() => toggleStatus(status.value)}
              className={`rounded-full border px-3 py-1.5 text-sm transition-colors ${
                selectedStatuses.includes(status.value) ? 'border-primary bg-primary/8 text-primary' : 'border-border text-muted-foreground'
              }`}
            >
              {status.label}
            </button>
          ))}
        </div>
      </PageToolbar>

      <PagePanel className="overflow-hidden">
        {filteredRows.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>任务标题</TableHead>
                <TableHead>格式</TableHead>
                <TableHead>版本</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>耗时</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredRows.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      {row.format === 'pdf' ? <FileText className="h-4 w-4 text-primary" /> : <FileCode2 className="h-4 w-4 text-primary" />}
                      <span>{row.paperTitle}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{row.format.toUpperCase()}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{row.variant === 'student' ? '学生' : row.variant === 'answer' ? '答案' : '双版本'}</Badge>
                  </TableCell>
                  <TableCell>{renderStatus(row)}</TableCell>
                  <TableCell>{formatAbsoluteDateTime(row.createdAt)}</TableCell>
                  <TableCell>{formatElapsed(row, now)}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      {row.status === 'done' ? (
                        <Button variant="ghost" size="icon" asChild>
                          <a href={getExportDownloadUrl(row.id)} download>
                            <Download className="h-4 w-4" />
                          </a>
                        </Button>
                      ) : null}

                      {row.status === 'failed' ? (
                        <Dialog>
                          <DialogTrigger asChild>
                            <Button variant="ghost" size="icon">
                              <AlertCircle className="h-4 w-4 text-destructive" />
                            </Button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>错误详情</DialogTitle>
                            </DialogHeader>
                            <pre className="whitespace-pre-wrap rounded-xl bg-muted p-4 text-sm">
                              {row.errorMessage ?? '未知错误'}
                            </pre>
                          </DialogContent>
                        </Dialog>
                      ) : null}

                      <ConfirmActionButton
                        variant="ghost"
                        size="icon"
                        onConfirm={() => deleteExportMutation.mutateAsync(row.id)}
                        pending={deleteExportMutation.isPending}
                        title={row.status === 'pending' || row.status === 'processing' ? '确认终止导出任务' : '确认删除导出记录'}
                        description={
                          row.status === 'pending' || row.status === 'processing'
                            ? `确定要终止《${row.paperTitle}》的导出任务吗？`
                            : `确定要删除《${row.paperTitle}》的导出记录吗？`
                        }
                        confirmLabel={row.status === 'pending' || row.status === 'processing' ? '确认终止' : '确认删除'}
                      >
                        {row.status === 'pending' || row.status === 'processing' ? <X className="h-4 w-4" /> : <Trash2 className="h-4 w-4" />}
                      </ConfirmActionButton>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Empty className="min-h-[320px] border-none bg-transparent">
            <EmptyMedia variant="icon">
              <FileText className="h-6 w-6" />
            </EmptyMedia>
            <EmptyHeader>
              <EmptyTitle>没有匹配的导出记录</EmptyTitle>
              <EmptyDescription>试试调整状态、日期或格式筛选条件。</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </PagePanel>
    </PageShell>
  );
}

function renderStatus(row: ExportJob) {
  if (row.status === 'processing') {
    return (
      <Badge variant="outline" className="gap-2 border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300">
        <LoaderCircle className="h-3 w-3 animate-spin" />
        生成中
      </Badge>
    );
  }

  if (row.status === 'done') {
    return <Badge variant="outline" className="border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">完成</Badge>;
  }

  if (row.status === 'failed') {
    return <Badge variant="outline" className="border-rose-500/40 bg-rose-500/10 text-rose-700 dark:text-rose-300">失败</Badge>;
  }

  return <Badge variant="outline" className="border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300">排队中</Badge>;
}

function formatElapsed(row: ExportJob, now: number) {
  const start = row.startedAt ? new Date(row.startedAt).getTime() : new Date(row.createdAt).getTime();
  const end = row.completedAt ? new Date(row.completedAt).getTime() : now;
  const minutes = Math.max(0, Math.round((end - start) / 60000));
  if (minutes < 1) {
    return '少于 1 分钟';
  }
  if (minutes < 60) {
    return `${minutes} 分钟`;
  }
  const hours = Math.floor(minutes / 60);
  const remain = minutes % 60;
  return remain > 0 ? `${hours} 小时 ${remain} 分钟` : `${hours} 小时`;
}

function useCurrentTimestamp(intervalMs = 60_000) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, intervalMs);

    return () => window.clearInterval(timer);
  }, [intervalMs]);

  return now;
}
