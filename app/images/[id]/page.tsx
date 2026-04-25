'use client';

import Image from 'next/image';
import { use } from 'react';
import Link from 'next/link';
import { ArrowLeft, Pencil } from 'lucide-react';
import { notFound } from 'next/navigation';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useImage } from '@/lib/hooks/use-images';

export default function ImageDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const imageQuery = useImage(id);
  const payload = imageQuery.data;

  if (!imageQuery.isLoading && !payload) {
    notFound();
  }

  if (!payload) {
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

  const { image, linkedProblems, tags } = payload;

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="图像详情"
        title={image.filename}
        description={image.description ?? '查看图像尺寸、标签和关联题目。'}
        actions={
          <>
            <Button variant="ghost" asChild>
              <Link href="/images">
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回图库
              </Link>
            </Button>
            <Button asChild>
              <Link href={`/images/${image.id}/edit`}>
                <Pencil className="mr-2 h-4 w-4" />
                编辑图像
              </Link>
            </Button>
          </>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <PagePanel>
          <div className="p-6">
            <Image
              src={image.url}
              alt={image.description ?? image.filename}
              width={image.width}
              height={image.height}
              unoptimized
              className="h-auto w-full rounded-2xl border border-border/70 object-contain"
            />
          </div>
        </PagePanel>

        <div className="space-y-4">
          <PagePanel>
            <div className="space-y-4 p-5 text-sm">
              <div className="grid gap-3 sm:grid-cols-2">
                <div>格式：{image.mime}</div>
                <div>尺寸：{image.width} × {image.height}</div>
                <div>大小：{Math.round(image.size / 1024)} KB</div>
                <div>关联题目：{linkedProblems.length}</div>
              </div>
              <div className="flex flex-wrap gap-2">
                {tags.map((tag) => (
                  <Badge key={tag.id} variant="outline" style={{ borderColor: tag.color, color: tag.color }}>
                    {tag.name}
                  </Badge>
                ))}
              </div>
            </div>
          </PagePanel>

          <PagePanel>
            <div className="border-b border-border/70 px-5 py-4">
              <h2 className="text-base font-semibold">关联题目</h2>
            </div>
            <div className="space-y-3 p-5">
              {linkedProblems.length > 0 ? (
                linkedProblems.map((problem) => (
                  <Link key={problem.id} href={`/problems/${problem.id}`} className="block rounded-xl border border-border/70 p-3 transition-colors hover:bg-muted/50">
                    <p className="font-mono text-xs text-muted-foreground">{problem.code}</p>
                    <div className="mt-1 line-clamp-2 text-sm">{problem.latex}</div>
                  </Link>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">暂无关联题目</div>
              )}
            </div>
          </PagePanel>
        </div>
      </div>
    </PageShell>
  );
}
