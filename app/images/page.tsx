'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Eye, ImageOff, LayoutGrid, Pencil, Rows3, Search, Trash2, Upload } from 'lucide-react';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { useBatchDeleteImages, useImages, useUploadImage } from '@/lib/hooks/use-images';
import { useTags } from '@/lib/hooks/use-tags';

const NO_UPLOAD_TAG = '__none__';

export default function ImagesPage() {
  const [search, setSearch] = useState('');
  const [tagId, setTagId] = useState('all');
  const [mime, setMime] = useState('all');
  const [viewMode, setViewMode] = useState<'masonry' | 'grid'>('masonry');
  const [page, setPage] = useState(1);
  const [selectedImages, setSelectedImages] = useState<string[]>([]);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadDescription, setUploadDescription] = useState('');
  const [uploadTagId, setUploadTagId] = useState<string>(NO_UPLOAD_TAG);

  const tagsQuery = useTags();
  const imagesQuery = useImages({
    keyword: search,
    tagIds: tagId === 'all' ? undefined : tagId,
    mime: mime === 'all' ? undefined : mime,
    page,
    pageSize: viewMode === 'grid' ? 12 : 8,
  });
  const uploadMutation = useUploadImage();
  const batchDeleteMutation = useBatchDeleteImages();

  const images = imagesQuery.data?.items ?? [];
  const total = imagesQuery.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / (viewMode === 'grid' ? 12 : 8)));
  const tagMap = useMemo(() => new Map((tagsQuery.data ?? []).map((tag) => [tag.id, tag])), [tagsQuery.data]);
  const activeFilterCount = [Boolean(search.trim()), tagId !== 'all', mime !== 'all'].filter(Boolean).length;

  const upload = async () => {
    if (!uploadFile) {
      return;
    }

    await uploadMutation.mutateAsync({
      file: uploadFile,
      tagIds: uploadTagId === NO_UPLOAD_TAG ? [] : [uploadTagId],
      description: uploadDescription || undefined,
    });

    setUploadOpen(false);
    setUploadFile(null);
    setUploadDescription('');
    setUploadTagId(NO_UPLOAD_TAG);
  };

  const handleBatchDelete = async () => {
    await batchDeleteMutation.mutateAsync(selectedImages);
    setSelectedImages([]);
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="图库"
        title="图像资源"
        description="统一管理题目关联图像、示意图和素材，筛选与上传分离，避免列表状态污染新内容。"
        badges={
          <>
            <Badge variant="secondary">共 {total} 张图像</Badge>
            <Badge variant="secondary">启用筛选 {activeFilterCount}</Badge>
          </>
        }
        actions={
          <Dialog
            open={uploadOpen}
            onOpenChange={(open) => {
              setUploadOpen(open);
              if (!open) {
                setUploadFile(null);
                setUploadDescription('');
                setUploadTagId(NO_UPLOAD_TAG);
              }
            }}
          >
            <DialogTrigger asChild>
              <Button>
                <Upload className="mr-2 h-4 w-4" />
                上传图像
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>上传图像</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div className="rounded-2xl border border-dashed p-4 text-sm leading-6 text-muted-foreground">
                  当前对话框中的标签仅作用于这次上传，不会继承图库列表当前的筛选状态。
                </div>
                <Input type="file" accept="image/png,image/jpeg,image/heic,image/heif,image/webp" onChange={(event) => setUploadFile(event.target.files?.[0] ?? null)} />
                <Select value={uploadTagId} onValueChange={setUploadTagId}>
                  <SelectTrigger>
                    <SelectValue placeholder="上传时附加标签（可选）" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={NO_UPLOAD_TAG}>不设置标签</SelectItem>
                    {(tagsQuery.data ?? []).map((tag) => (
                      <SelectItem key={tag.id} value={tag.id}>
                        {tag.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Input value={uploadDescription} onChange={(event) => setUploadDescription(event.target.value)} placeholder="图像描述（可选）" />
                <Button onClick={upload} disabled={!uploadFile || uploadMutation.isPending} className="w-full">
                  确认上传
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        }
      />

      <PageToolbar className="space-y-4">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex flex-1 flex-col gap-3 xl:flex-row xl:items-center">
            <div className="relative w-full xl:max-w-sm">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => {
                  setSearch(event.target.value);
                  setPage(1);
                }}
                placeholder="搜图片名、描述..."
                aria-label="搜索图片"
                className="pl-9"
              />
            </div>

            <div className="flex flex-1 flex-wrap items-center gap-2">
              <Select
                value={tagId}
                onValueChange={(value) => {
                  setTagId(value);
                  setPage(1);
                }}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="标签筛选" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部标签</SelectItem>
                  {(tagsQuery.data ?? []).map((tag) => (
                    <SelectItem key={tag.id} value={tag.id}>
                      {tag.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={mime}
                onValueChange={(value) => {
                  setMime(value);
                  setPage(1);
                }}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="格式" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部格式</SelectItem>
                  <SelectItem value="image/png">PNG</SelectItem>
                  <SelectItem value="image/jpeg">JPG</SelectItem>
                  <SelectItem value="image/webp">WebP</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {selectedImages.length > 0 ? (
              <>
                <Badge variant="secondary">已选 {selectedImages.length}</Badge>
                <Button variant="ghost" size="sm" onClick={() => setSelectedImages([])}>
                  清空选择
                </Button>
                <ConfirmActionButton
                  variant="destructive"
                  size="sm"
                  onConfirm={handleBatchDelete}
                  pending={batchDeleteMutation.isPending}
                  title="确认批量删除图像"
                  description={`确定要删除选中的 ${selectedImages.length} 张图像吗？它们会移入回收站。`}
                  confirmLabel="确认删除"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  删除所选
                </ConfirmActionButton>
              </>
            ) : null}

            <ToggleGroup type="single" value={viewMode} onValueChange={(value) => value && setViewMode(value as typeof viewMode)}>
              <ToggleGroupItem value="masonry" aria-label="网格视图">
                <Rows3 className="h-4 w-4" />
              </ToggleGroupItem>
              <ToggleGroupItem value="grid" aria-label="列表视图">
                <LayoutGrid className="h-4 w-4" />
              </ToggleGroupItem>
            </ToggleGroup>
          </div>
        </div>

        <p className="text-sm text-muted-foreground">
          {selectedImages.length > 0 ? `已选 ${selectedImages.length} 张图像，可直接批量删除。` : `当前共有 ${total} 张图像，支持瀑布流与规则网格两种浏览方式。`}
        </p>
      </PageToolbar>

      {images.length > 0 ? (
        <div className={viewMode === 'grid' ? 'grid gap-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4' : 'columns-1 gap-4 md:columns-2 xl:columns-3'}>
          {images.map((image) => (
            <PagePanel key={image.id} className="group mb-4 break-inside-avoid overflow-hidden transition-colors hover:border-primary/40">
              <div className="relative border-b border-border/70 bg-background/20">
                <div
                  className={viewMode === 'grid' ? 'relative aspect-[4/3] w-full' : 'relative w-full'}
                  style={viewMode === 'masonry' ? { aspectRatio: `${image.width} / ${image.height}` } : undefined}
                >
                  <Image
                    src={image.thumbnailUrl}
                    alt={image.description ?? image.filename}
                    fill
                    unoptimized
                    sizes={viewMode === 'grid' ? '(min-width: 1536px) 25vw, (min-width: 1280px) 33vw, (min-width: 768px) 50vw, 100vw' : '(min-width: 1280px) 33vw, (min-width: 768px) 50vw, 100vw'}
                    className="object-cover"
                  />
                </div>
                <div className="absolute left-3 top-3 rounded-md bg-background/90 p-1 shadow-sm">
                  <Checkbox
                    checked={selectedImages.includes(image.id)}
                    onCheckedChange={() =>
                      setSelectedImages((current) =>
                        current.includes(image.id) ? current.filter((item) => item !== image.id) : [...current, image.id]
                      )
                    }
                  />
                </div>
              </div>
              <div className="flex min-h-[170px] flex-col gap-3 p-4">
                <div className="min-h-[2.75rem]">
                  <p className="truncate font-medium">{image.filename}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {image.width} × {image.height}
                  </p>
                </div>

                <div className="flex min-h-[2.75rem] flex-wrap content-start gap-2 overflow-hidden">
                  {image.tagIds.slice(0, 3).map((item) => {
                    const tag = tagMap.get(item);
                    return tag ? (
                      <Badge key={tag.id} variant="outline" style={{ borderColor: tag.color, color: tag.color }}>
                        {tag.name}
                      </Badge>
                    ) : null;
                  })}
                </div>

                <div className="mt-auto grid gap-2 sm:grid-cols-2">
                  <Button size="sm" variant="outline" asChild className="w-full">
                    <Link href={`/images/${image.id}`}>
                      <Eye className="mr-2 h-4 w-4" />
                      查看
                    </Link>
                  </Button>
                  <Button size="sm" variant="outline" asChild className="w-full">
                    <Link href={`/images/${image.id}/edit`}>
                      <Pencil className="mr-2 h-4 w-4" />
                      编辑
                    </Link>
                  </Button>
                </div>
              </div>
            </PagePanel>
          ))}
        </div>
      ) : (
        <PagePanel>
          <Empty className="min-h-[340px] border-none bg-transparent">
            <EmptyMedia variant="icon">
              <ImageOff className="h-6 w-6" />
            </EmptyMedia>
            <EmptyHeader>
              <EmptyTitle>当前没有匹配图像</EmptyTitle>
              <EmptyDescription>试试清空筛选条件，或者直接上传新的图像素材。</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </PagePanel>
      )}

      {totalPages > 1 ? (
        <PageToolbar className="flex justify-center">
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious href="#" onClick={(event) => { event.preventDefault(); setPage((value) => Math.max(1, value - 1)); }} />
              </PaginationItem>
              <PaginationItem>
                <span className="px-3 text-sm text-muted-foreground">
                  第 {page} / {totalPages} 页
                </span>
              </PaginationItem>
              <PaginationItem>
                <PaginationNext href="#" onClick={(event) => { event.preventDefault(); setPage((value) => Math.min(totalPages, value + 1)); }} />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </PageToolbar>
      ) : null}
    </PageShell>
  );
}
