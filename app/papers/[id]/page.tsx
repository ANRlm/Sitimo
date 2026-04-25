'use client';

import { use, useState } from 'react';
import Link from 'next/link';
import { ArrowLeft, Download, Eye, Pencil, Settings2 } from 'lucide-react';
import { notFound } from 'next/navigation';
import { PaperDocument } from '@/components/paper-document';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatAbsoluteDateTime } from '@/lib/format';
import { useCreateExport } from '@/lib/hooks/use-exports';
import { usePaper } from '@/lib/hooks/use-papers';

export default function PaperPreviewPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const paperQuery = usePaper(id);
  const createExportMutation = useCreateExport();
  const [variant, setVariant] = useState<'student' | 'answer'>('student');

  if (!paperQuery.isLoading && !paperQuery.data) {
    notFound();
  }

  if (!paperQuery.data) {
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

  const paper = paperQuery.data;
  const exportVariant = variant === 'answer' ? 'answer' : 'student';

  const queueExport = (format: 'pdf' | 'latex') => {
    createExportMutation.mutate({ paperId: paper.id, format, variant: exportVariant });
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="试卷预览"
        title={paper.title}
        description={`${paper.subtitle ?? '未填写副标题'} · 共 ${paper.items.length} 题 · ${paper.totalScore ?? 0} 分`}
        actions={
          <>
            <Button variant="ghost" asChild>
              <Link href="/papers">
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回
              </Link>
            </Button>
            <Tabs value={variant} onValueChange={(value) => setVariant(value as 'student' | 'answer')}>
              <TabsList>
                <TabsTrigger value="student">学生版</TabsTrigger>
                <TabsTrigger value="answer">答案版</TabsTrigger>
              </TabsList>
            </Tabs>
            <Button asChild>
              <Link href={`/papers/${paper.id}/editor`}>
                <Pencil className="mr-2 h-4 w-4" />
                编辑
              </Link>
            </Button>
            <Button variant="outline" onClick={() => queueExport('latex')} disabled={createExportMutation.isPending}>
              <Download className="mr-2 h-4 w-4" />
              导出 LaTeX 包
            </Button>
            <Button variant="outline" onClick={() => queueExport('pdf')} disabled={createExportMutation.isPending}>
              <Download className="mr-2 h-4 w-4" />
              导出 PDF
            </Button>
            <Sheet>
              <SheetTrigger asChild>
                <Button variant="outline">
                  <Settings2 className="mr-2 h-4 w-4" />
                  设置
                </Button>
              </SheetTrigger>
              <SheetContent className="w-[360px]">
                <SheetHeader>
                  <SheetTitle>试卷设置</SheetTitle>
                </SheetHeader>
                <div className="space-y-4 py-6">
                  <Card>
                    <CardContent className="space-y-3 p-4 text-sm">
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">学校</span>
                        <span>{paper.schoolName ?? '未设置'}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">考试名</span>
                        <span>{paper.examName ?? '未设置'}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">纸张</span>
                        <span>{paper.layout.paperSize}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">分栏</span>
                        <span>{paper.layout.columns === 1 ? '单栏' : '双栏'}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">字号</span>
                        <span>{paper.layout.fontSize} pt</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">行距</span>
                        <span>{paper.layout.lineHeight}</span>
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="space-y-3 p-4 text-sm">
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">创建时间</span>
                        <span>{formatAbsoluteDateTime(paper.createdAt)}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">最近修改</span>
                        <span>{formatAbsoluteDateTime(paper.updatedAt)}</span>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </SheetContent>
            </Sheet>
          </>
        }
      />

      <PagePanel className="rounded-[32px] bg-[radial-gradient(circle_at_top,_hsl(var(--primary)/0.08),_transparent_45%),linear-gradient(180deg,_transparent,_hsl(var(--muted)))] p-4">
        <PaperDocument paper={paper} showAnswers={variant === 'answer'} />
      </PagePanel>

      <PagePanel className="border-dashed bg-muted/40">
        <div className="flex items-start gap-3 p-4 text-sm text-muted-foreground">
          <Eye className="mt-0.5 h-4 w-4" />
          <p>当前为只读预览。若需调整题序、分值或版式，请进入编辑器；若需生成文件，可直接使用顶部导出按钮。</p>
        </div>
      </PagePanel>
    </PageShell>
  );
}
