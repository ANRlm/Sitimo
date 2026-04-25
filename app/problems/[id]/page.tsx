'use client';

import Link from 'next/link';
import { ArrowLeft, Copy, Link2, Pencil, ShoppingBasket, Trash2 } from 'lucide-react';
import { notFound } from 'next/navigation';
import { toast } from 'sonner';
import { use } from 'react';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { LatexCodeEditor } from '@/components/latex-code-editor';
import { MathText } from '@/components/math-text';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatAbsoluteDateTime } from '@/lib/format';
import { useDeleteProblem, useProblem, useProblemVersions, useRollbackProblemVersion } from '@/lib/hooks/use-problems';
import { useBasketStore } from '@/lib/store';
import { difficultyConfig } from '@/lib/types';

export default function ProblemDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const problemQuery = useProblem(id);
  const versionsQuery = useProblemVersions(id);
  const rollbackMutation = useRollbackProblemVersion(id);
  const deleteMutation = useDeleteProblem();
  const addItem = useBasketStore((state) => state.addItem);
  const removeItem = useBasketStore((state) => state.removeItem);
  const hasItem = useBasketStore((state) => state.hasItem);

  const problem = problemQuery.data;

  if (!problemQuery.isLoading && !problem) {
    notFound();
  }

  if (!problem) {
    return (
      <div className="space-y-4">
        <PagePanel>
          <div className="space-y-4 p-6">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-32 w-full" />
          </div>
        </PagePanel>
      </div>
    );
  }

  const difficulty = difficultyConfig[problem.difficulty];
  const isInBasket = hasItem(problem.id);

  return (
    <PageShell>
      <PageHeader
        eyebrow="题目详情"
        title={problem.code}
        description={problem.source ?? '查看题干、答案、解析、关联图像与版本历史。'}
        badges={
          <>
            <Badge variant="outline" style={{ borderColor: difficulty.color, color: difficulty.color }}>
              {difficulty.label}
            </Badge>
            <Badge variant="secondary">难度分 {problem.subjectiveScore ?? 0}/10</Badge>
            {problem.subject ? <Badge variant="secondary">{problem.subject}</Badge> : null}
            {problem.grade ? <Badge variant="secondary">{problem.grade}</Badge> : null}
          </>
        }
        actions={
          <>
            <Button variant="ghost" asChild>
              <Link href="/problems">
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回题库
              </Link>
            </Button>
            <Button asChild>
              <Link href={`/problems/${problem.id}/edit`}>
                <Pencil className="mr-2 h-4 w-4" />
                编辑
              </Link>
            </Button>
            <Button
              variant={isInBasket ? 'outline' : 'secondary'}
              onClick={() => {
                if (isInBasket) {
                  removeItem(problem.id);
                  toast.success('已从题目篮子移除');
                  return;
                }
                addItem({
                  id: problem.id,
                  code: problem.code,
                  latex: problem.latex,
                  difficulty: problem.difficulty,
                });
                toast.success('已加入题目篮子');
              }}
            >
              <ShoppingBasket className="mr-2 h-4 w-4" />
              {isInBasket ? '已在篮子中' : '加入篮子'}
            </Button>
            <Button
              variant="outline"
              onClick={async () => {
                await navigator.clipboard.writeText(window.location.href);
                toast.success('链接已复制');
              }}
            >
              <Link2 className="mr-2 h-4 w-4" />
              复制链接
            </Button>
            <ConfirmActionButton
              variant="ghost"
              className="text-destructive hover:text-destructive"
              onConfirm={() => deleteMutation.mutateAsync(problem.id)}
              pending={deleteMutation.isPending}
              title="确认删除题目"
              description={`确定要删除题目 ${problem.code} 吗？删除后会移入回收站。`}
              confirmLabel="确认删除"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              删除
            </ConfirmActionButton>
          </>
        }
      >
        {problem.tags.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {problem.tags.map((tag) => (
              <Badge key={tag.id} variant="outline" style={{ borderColor: tag.color, color: tag.color }}>
                {tag.name}
              </Badge>
            ))}
          </div>
        ) : null}
      </PageHeader>

      <PagePanel>
        <div className="min-h-[260px] p-8">
          <p className="mb-4 text-xs font-medium uppercase tracking-[0.24em] text-muted-foreground">题目</p>
          <MathText latex={problem.latex} className="text-lg leading-loose" />
        </div>
      </PagePanel>

      <Tabs defaultValue="answer">
        <TabsList className="mb-3 flex h-auto flex-wrap justify-start gap-2 rounded-xl bg-transparent p-0">
          <TabsTrigger value="answer">答案</TabsTrigger>
          <TabsTrigger value="solution">解析</TabsTrigger>
          <TabsTrigger value="source">源码</TabsTrigger>
          <TabsTrigger value="images">关联图像</TabsTrigger>
          <TabsTrigger value="history">版本历史</TabsTrigger>
          <TabsTrigger value="metadata">元数据</TabsTrigger>
        </TabsList>

        <TabsContent value="answer">
          <PagePanel>
            <div className="p-6">{problem.answerLatex ? <MathText latex={problem.answerLatex} className="leading-loose" /> : <div className="text-sm text-muted-foreground">暂无答案</div>}</div>
          </PagePanel>
        </TabsContent>

        <TabsContent value="solution">
          <PagePanel>
            <div className="p-6">{problem.solutionLatex ? <MathText latex={problem.solutionLatex} className="leading-loose" /> : <div className="text-sm text-muted-foreground">暂无解析</div>}</div>
          </PagePanel>
        </TabsContent>

        <TabsContent value="source">
          <PagePanel>
            <div className="flex items-center justify-between border-b border-border/70 px-6 py-4">
              <h2 className="text-base font-semibold">LaTeX 源码</h2>
              <Button variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(problem.latex)}>
                <Copy className="mr-2 h-4 w-4" />
                复制
              </Button>
            </div>
            <div className="p-6">
              <LatexCodeEditor value={problem.latex} readOnly minHeight={240} />
            </div>
          </PagePanel>
        </TabsContent>

        <TabsContent value="images">
          <PagePanel>
            <div className="grid gap-4 p-6 md:grid-cols-3">
              {problem.images.length > 0 ? (
                problem.images.map((image) => (
                  <Link key={image.id} href={`/images/${image.id}`} className="overflow-hidden rounded-2xl border border-border/70 transition-colors hover:border-primary">
                    <img src={image.thumbnailUrl} alt={image.description ?? image.filename} className="aspect-video w-full object-cover" />
                    <div className="space-y-1 p-4">
                      <p className="truncate text-sm font-medium">{image.filename}</p>
                      <p className="text-xs text-muted-foreground">{image.description ?? '未填写描述'}</p>
                    </div>
                  </Link>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">暂无关联图像</div>
              )}
            </div>
          </PagePanel>
        </TabsContent>

        <TabsContent value="history">
          <PagePanel>
            <div className="p-6">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>版本号</TableHead>
                    <TableHead>时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(versionsQuery.data ?? []).map((entry) => (
                    <TableRow key={entry.id}>
                      <TableCell className="font-medium">v{entry.version}</TableCell>
                      <TableCell>{formatAbsoluteDateTime(entry.createdAt)}</TableCell>
                      <TableCell className="text-right">
                        <Button variant="outline" size="sm" onClick={() => rollbackMutation.mutate(entry.version)} disabled={rollbackMutation.isPending}>
                          恢复此版本
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </PagePanel>
        </TabsContent>

        <TabsContent value="metadata">
          <PagePanel>
            <div className="grid gap-3 p-6 text-sm sm:grid-cols-[96px_1fr]">
              <span className="text-muted-foreground">创建时间</span>
              <span>{formatAbsoluteDateTime(problem.createdAt)}</span>
              <span className="text-muted-foreground">修改时间</span>
              <span>{formatAbsoluteDateTime(problem.updatedAt)}</span>
              <span className="text-muted-foreground">版本</span>
              <span>{problem.version}</span>
            </div>
          </PagePanel>
        </TabsContent>
      </Tabs>
    </PageShell>
  );
}
