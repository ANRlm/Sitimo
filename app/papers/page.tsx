'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Copy, Eye, FileStack, Pencil, Plus, Search, Trash2 } from 'lucide-react';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { PAPER_STATUS_CLASSNAMES, PAPER_STATUS_LABELS } from '@/lib/catalogs';
import { formatRelativeTime } from '@/lib/format';
import { useDeletePaper, useDuplicatePaper, usePapers } from '@/lib/hooks/use-papers';

export default function PapersPage() {
  const [query, setQuery] = useState('');
  const papersQuery = usePapers({ keyword: query, page: 1, pageSize: 100 });
  const duplicateMutation = useDuplicatePaper();
  const deleteMutation = useDeletePaper();

  const papers = useMemo(() => papersQuery.data?.items ?? [], [papersQuery.data?.items]);

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="试卷"
        title="试卷列表"
        description="统一查看试卷状态、题量和导出入口，危险操作统一走确认，避免列表页直接误删。"
        actions={
          <Button asChild>
            <Link href="/papers/new">
              <Plus className="mr-2 h-4 w-4" />
              新建试卷
            </Link>
          </Button>
        }
      />

      <PageToolbar className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="relative w-full lg:max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索试卷标题..." aria-label="搜索试卷标题" className="pl-9" />
        </div>
        <div className="rounded-full bg-muted px-3 py-2 text-sm text-muted-foreground">共 {papersQuery.data?.total ?? 0} 份试卷</div>
      </PageToolbar>

      <PagePanel className="overflow-hidden">
        {papers.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="py-3 px-4">标题</TableHead>
                <TableHead className="py-3 px-4">题目数</TableHead>
                <TableHead className="py-3 px-4">总分</TableHead>
                <TableHead className="py-3 px-4">最后修改</TableHead>
                <TableHead className="py-3 px-4">状态</TableHead>
                <TableHead className="w-[260px] text-right py-3 px-4">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {papers.map((paper) => (
                <TableRow key={paper.id}>
                  <TableCell>
                    <div className="space-y-1">
                      <div className="font-medium">{paper.title}</div>
                      <div className="text-sm text-muted-foreground">{paper.subtitle ?? paper.description ?? '未填写副标题'}</div>
                    </div>
                  </TableCell>
                  <TableCell>{paper.items.length}</TableCell>
                  <TableCell>{paper.totalScore ?? 0}</TableCell>
                  <TableCell>{formatRelativeTime(paper.updatedAt)}</TableCell>
                  <TableCell>
                    <Badge className={PAPER_STATUS_CLASSNAMES[paper.status]}>{PAPER_STATUS_LABELS[paper.status]}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" asChild title="查看" aria-label="查看">
                        <Link href={`/papers/${paper.id}`}>
                          <Eye className="h-4 w-4" />
                        </Link>
                      </Button>
                      <Button variant="ghost" size="icon" asChild title="编辑" aria-label="编辑">
                        <Link href={`/papers/${paper.id}/editor`}>
                          <Pencil className="h-4 w-4" />
                        </Link>
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => duplicateMutation.mutate(paper.id)} disabled={duplicateMutation.isPending} title="复制" aria-label="复制">
                        <Copy className="h-4 w-4" />
                      </Button>
                      <ConfirmActionButton
                        variant="ghost"
                        size="icon"
                        className="text-destructive hover:text-destructive"
                        onConfirm={() => deleteMutation.mutateAsync(paper.id)}
                        pending={deleteMutation.isPending}
                        title="确认删除试卷"
                        description={`确定要删除《${paper.title}》吗？删除后将从试卷列表移除。`}
                        confirmLabel="确认删除"
                      >
                        <Trash2 className="h-4 w-4" />
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
              <FileStack className="h-6 w-6" />
            </EmptyMedia>
            <EmptyHeader>
              <EmptyTitle>没有匹配的试卷</EmptyTitle>
              <EmptyDescription>试试更短的关键词，或者直接新建一份试卷。</EmptyDescription>
            </EmptyHeader>
            <Button asChild>
              <Link href="/papers/new">新建试卷</Link>
            </Button>
          </Empty>
        )}
      </PagePanel>
    </PageShell>
  );
}
