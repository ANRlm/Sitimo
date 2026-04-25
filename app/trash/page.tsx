'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Image as ImageIcon, RotateCcw, Trash2 } from 'lucide-react';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatAbsoluteDateTime } from '@/lib/format';
import { useHardDeleteImage, useImages, useRestoreImage } from '@/lib/hooks/use-images';
import { useHardDeleteProblem, useProblems, useRestoreProblem } from '@/lib/hooks/use-problems';

type TrashTab = 'problem' | 'image';

export default function TrashPage() {
  const [activeTab, setActiveTab] = useState<TrashTab>('problem');

  const deletedProblemsQuery = useProblems({ deleted: true, page: 1, pageSize: 100 });
  const deletedImagesQuery = useImages({ deleted: true, page: 1, pageSize: 100 });
  const restoreProblemMutation = useRestoreProblem();
  const hardDeleteProblemMutation = useHardDeleteProblem();
  const restoreImageMutation = useRestoreImage();
  const hardDeleteImageMutation = useHardDeleteImage();

  const problemItems = useMemo(
    () => (deletedProblemsQuery.data?.items ?? []).filter((item) => item.isDeleted),
    [deletedProblemsQuery.data?.items]
  );
  const imageItems = useMemo(
    () => (deletedImagesQuery.data?.items ?? []).filter((item) => item.isDeleted),
    [deletedImagesQuery.data?.items]
  );

  const clearTab = async (type: TrashTab) => {
    if (type === 'problem') {
      await Promise.all(problemItems.map((item) => hardDeleteProblemMutation.mutateAsync(item.id)));
      return;
    }
    await Promise.all(imageItems.map((item) => hardDeleteImageMutation.mutateAsync(item.id)));
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="回收站"
        title="已删除项目"
        description="删除的题目和图像会暂存在这里，支持恢复或彻底删除；批量清空仍需显式确认。"
        actions={
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" disabled={activeTab === 'problem' ? problemItems.length === 0 : imageItems.length === 0}>
                <Trash2 className="mr-2 h-4 w-4" />
                清空当前标签页
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>确定清空当前标签页吗？</AlertDialogTitle>
                <AlertDialogDescription>这会永久删除当前标签页中的所有项目，且无法恢复。</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>取消</AlertDialogCancel>
                <AlertDialogAction onClick={() => clearTab(activeTab)} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                  确认清空
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        }
      />

      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as TrashTab)}>
        <TabsList className="grid w-fit grid-cols-2">
          <TabsTrigger value="problem">题目</TabsTrigger>
          <TabsTrigger value="image">图像</TabsTrigger>
        </TabsList>

        <TabsContent value="problem" className="mt-4 space-y-4">
          {problemItems.length > 0 ? (
            <div className="grid gap-4">
              {problemItems.map((problem) => (
                <PagePanel key={problem.id}>
                  <div className="flex items-start justify-between gap-4 p-5">
                    <div className="min-w-0 flex-1 space-y-2">
                      <p className="font-mono text-xs text-muted-foreground">{problem.code}</p>
                      <p className="line-clamp-3 text-sm leading-7">{problem.latex}</p>
                      <p className="text-xs text-muted-foreground">最后更新时间 {formatAbsoluteDateTime(problem.updatedAt)}</p>
                    </div>
                    <div className="flex shrink-0 gap-2">
                      <Button variant="outline" size="sm" onClick={() => restoreProblemMutation.mutate(problem.id)} disabled={restoreProblemMutation.isPending}>
                        <RotateCcw className="mr-2 h-4 w-4" />
                        恢复
                      </Button>
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="destructive" size="sm">彻底删除</Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>彻底删除这道题目？</AlertDialogTitle>
                            <AlertDialogDescription>删除后无法恢复。</AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction onClick={() => hardDeleteProblemMutation.mutate(problem.id)} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                              删除
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </div>
                </PagePanel>
              ))}
            </div>
          ) : (
            <EmptyTrashState label="已删除的题目会显示在这里" />
          )}
        </TabsContent>

        <TabsContent value="image" className="mt-4 space-y-4">
          {imageItems.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {imageItems.map((image) => (
                <PagePanel key={image.id}>
                  <div className="space-y-3 p-5">
                    <Link href={`/images/${image.id}`}>
                      <img src={image.thumbnailUrl} alt={image.filename} className="aspect-video w-full rounded-xl object-cover" />
                    </Link>
                    <div>
                      <p className="truncate font-medium">{image.filename}</p>
                      <p className="mt-1 text-xs text-muted-foreground">最后更新时间 {formatAbsoluteDateTime(image.updatedAt ?? image.createdAt)}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" size="sm" className="flex-1" onClick={() => restoreImageMutation.mutate(image.id)} disabled={restoreImageMutation.isPending}>
                        <RotateCcw className="mr-2 h-4 w-4" />
                        恢复
                      </Button>
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="destructive" size="sm" className="flex-1">彻底删除</Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>彻底删除这张图像？</AlertDialogTitle>
                            <AlertDialogDescription>删除后无法恢复。</AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction onClick={() => hardDeleteImageMutation.mutate(image.id)} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                              删除
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </div>
                </PagePanel>
              ))}
            </div>
          ) : (
            <EmptyTrashState label="已删除的图像会显示在这里" />
          )}
        </TabsContent>
      </Tabs>
    </PageShell>
  );
}

function EmptyTrashState({ label }: { label: string }) {
  return (
    <PagePanel>
      <Empty className="min-h-[320px] border-none bg-transparent">
        <EmptyMedia variant="icon">
          <ImageIcon className="h-6 w-6" />
        </EmptyMedia>
        <EmptyHeader>
          <EmptyTitle>{label}</EmptyTitle>
          <EmptyDescription>恢复或彻底删除操作会显示在这里。</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </PagePanel>
  );
}
