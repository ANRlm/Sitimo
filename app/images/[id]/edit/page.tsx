'use client';

import Image from 'next/image';
import { use, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft, Save } from 'lucide-react';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { useEditImage, useImage } from '@/lib/hooks/use-images';

export default function ImageEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const searchParams = useSearchParams();
  const router = useRouter();
  const imageQuery = useImage(id);
  const editMutation = useEditImage();
  const problemId = searchParams.get('problemId') ?? undefined;

  const image = imageQuery.data?.image;
  const [draft, setDraft] = useState<{
    imageId: string;
    rotation: number[];
    width: string;
    height: string;
  } | null>(null);
  const editorState = useMemo(() => {
    if (!image) {
      return null;
    }

    if (draft?.imageId === image.id) {
      return draft;
    }

    return {
      imageId: image.id,
      rotation: [0],
      width: String(image.width),
      height: String(image.height),
    };
  }, [draft, image]);

  const apply = async () => {
    if (!editorState) {
      return;
    }

    await editMutation.mutateAsync({
      id,
      problemId,
      input: {
        rotate: editorState.rotation[0],
        resize: {
          w: Number(editorState.width),
          h: Number(editorState.height),
        },
      },
    });
    router.push(problemId ? `/problems/${problemId}/edit` : `/images/${id}`);
  };

  if (!image) {
    return <div className="p-6 text-sm text-muted-foreground">正在加载图像...</div>;
  }

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="图像编辑器"
        title={image.filename}
        description={problemId ? '当前从题目上下文进入，保存后会生成只影响当前题目的图像版本。' : '当前从图库直接编辑，保存后会更新这张图像的内容。'}
        actions={
          <>
            <Button variant="ghost" asChild>
              <Link href={problemId ? `/problems/${problemId}/edit` : `/images/${id}`}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回
              </Link>
            </Button>
            <Button onClick={apply} disabled={editMutation.isPending}>
              <Save className="mr-2 h-4 w-4" />
              应用修改
            </Button>
          </>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
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

        <PagePanel>
          <div className="space-y-5 p-5">
            <div className="rounded-xl border border-amber-400/30 bg-amber-500/10 p-4 text-sm text-amber-900 dark:text-amber-200">
              {problemId ? '保存后会生成新的图像版本并回填到当前题目，原图库图像不会被这次操作直接覆盖。' : '保存后会直接更新当前图库图像，请在确认尺寸和旋转参数后提交。'}
            </div>

            <div className="rounded-xl border border-border/70 bg-muted/30 p-4 text-sm">
              当前尺寸：{image.width} × {image.height}
            </div>

            <div>
              <Label className="mb-2 block">旋转 {editorState?.rotation[0] ?? 0}°</Label>
              <Slider
                min={0}
                max={360}
                step={90}
                value={editorState?.rotation ?? [0]}
                onValueChange={(value) =>
                  setDraft((current) => ({
                    imageId: image.id,
                    rotation: value,
                    width: current?.imageId === image.id ? current.width : String(image.width),
                    height: current?.imageId === image.id ? current.height : String(image.height),
                  }))
                }
              />
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <Label className="mb-2 block">宽度</Label>
                <Input
                  value={editorState?.width ?? String(image.width)}
                  onChange={(event) =>
                    setDraft((current) => ({
                      imageId: image.id,
                      rotation: current?.imageId === image.id ? current.rotation : [0],
                      width: event.target.value,
                      height: current?.imageId === image.id ? current.height : String(image.height),
                    }))
                  }
                />
              </div>
              <div>
                <Label className="mb-2 block">高度</Label>
                <Input
                  value={editorState?.height ?? String(image.height)}
                  onChange={(event) =>
                    setDraft((current) => ({
                      imageId: image.id,
                      rotation: current?.imageId === image.id ? current.rotation : [0],
                      width: current?.imageId === image.id ? current.width : String(image.width),
                      height: event.target.value,
                    }))
                  }
                />
              </div>
            </div>
          </div>
        </PagePanel>
      </div>
    </PageShell>
  );
}
