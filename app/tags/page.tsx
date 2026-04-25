'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Merge, Pencil, Plus, Search, Trash2, FileText } from 'lucide-react';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { Empty, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { TAG_CATEGORY_LABELS, TAG_PRESET_COLORS } from '@/lib/catalogs';
import { applyValidationErrors } from '@/lib/forms';
import type { TagCategoryFilter } from '@/lib/frontend-contracts';
import { useProblems } from '@/lib/hooks/use-problems';
import { useCreateTag, useDeleteTag, useMergeTag, useTags, useUpdateTag } from '@/lib/hooks/use-tags';
import { tagSchema, type TagFormValues } from '@/lib/schemas/tag';
import { cn } from '@/lib/utils';

const defaultTagValues: TagFormValues = {
  name: '',
  category: 'topic',
  color: '#0F766E',
  description: '',
};

export default function TagsPage() {
  const tagsQuery = useTags();
  const createTagMutation = useCreateTag();
  const updateTagMutation = useUpdateTag();
  const deleteTagMutation = useDeleteTag();
  const mergeTagMutation = useMergeTag();

  const tags = useMemo(() => tagsQuery.data ?? [], [tagsQuery.data]);
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<TagCategoryFilter>('all');
  const [selectedTagId, setSelectedTagId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [mergeTargetId, setMergeTargetId] = useState('');

  const form = useForm<TagFormValues>({
    resolver: zodResolver(tagSchema),
    defaultValues: defaultTagValues,
  });
  const effectiveSelectedTagId = selectedTagId ?? (!isCreating ? tags[0]?.id ?? null : null);

  const relatedProblemsQuery = useProblems({
    tagIds: effectiveSelectedTagId ?? undefined,
    page: 1,
    pageSize: 8,
  });

  const filteredTags = useMemo(
    () =>
      tags.filter((tag) => {
        const matchesCategory = categoryFilter === 'all' || tag.category === categoryFilter;
        const matchesQuery = !searchQuery.trim() || tag.name.toLowerCase().includes(searchQuery.trim().toLowerCase());
        return matchesCategory && matchesQuery;
      }),
    [categoryFilter, searchQuery, tags]
  );

  const selectedTag = useMemo(
    () => tags.find((tag) => tag.id === effectiveSelectedTagId) ?? null,
    [effectiveSelectedTagId, tags]
  );

  useEffect(() => {
    if (selectedTag) {
      form.reset({
        name: selectedTag.name,
        category: selectedTag.category,
        color: selectedTag.color,
        description: selectedTag.description ?? '',
      });
    } else if (isCreating || tags.length === 0) {
      form.reset(defaultTagValues);
    }
  }, [form, isCreating, selectedTag, tags.length]);

  const resetForm = () => {
    setIsCreating(true);
    setSelectedTagId(null);
    form.reset(defaultTagValues);
    setMergeTargetId('');
  };

  const handleSaveTag = async (values: TagFormValues) => {
    const payload = {
      name: values.name.trim(),
      category: values.category,
      color: values.color.trim(),
      description: values.description?.trim() || undefined,
    };

    try {
      if (selectedTag) {
        await updateTagMutation.mutateAsync({
          id: selectedTag.id,
          input: payload,
        });
        return;
      }

      const created = await createTagMutation.mutateAsync(payload);
      setIsCreating(false);
      setSelectedTagId(created.id);
    } catch (error) {
      applyValidationErrors(form, error);
    }
  };

  const category = useWatch({ control: form.control, name: 'category' });
  const color = useWatch({ control: form.control, name: 'color' });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSaveTag)}>
        <PageShell wide>
          <PageHeader
            eyebrow="标签管理"
            title="标签与分类"
            description="保持现有双栏编辑流程，同时把删除和合并等高风险动作收进确认流程。"
            badges={<Badge variant="secondary">共 {filteredTags.length} 个可见标签</Badge>}
            actions={
              <Button type="button" onClick={resetForm}>
                <Plus className="mr-2 h-4 w-4" />
                新建标签
              </Button>
            }
          />

          <div className="grid min-h-[calc(100vh-16rem)] gap-4 xl:grid-cols-[320px_minmax(0,1fr)]">
        <PagePanel className="flex flex-col overflow-hidden">
          <CardHeader className="shrink-0 space-y-4 pb-4 pt-6">
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input placeholder="搜索标签..." value={searchQuery} onChange={(event) => setSearchQuery(event.target.value)} aria-label="搜索标签" className="pl-9" />
              </div>
              <Button type="button" variant="outline" onClick={resetForm}>
                <Plus className="mr-2 h-4 w-4" />
                新建
              </Button>
            </div>

            <Tabs value={categoryFilter} onValueChange={(value) => setCategoryFilter(value as TagCategoryFilter)}>
              <TabsList className="w-full">
                <TabsTrigger value="all" className="flex-1">全部</TabsTrigger>
                <TabsTrigger value="topic" className="flex-1">知识点</TabsTrigger>
                <TabsTrigger value="source" className="flex-1">来源</TabsTrigger>
                <TabsTrigger value="custom" className="flex-1">自定义</TabsTrigger>
              </TabsList>
            </Tabs>
          </CardHeader>

          <ScrollArea className="flex-1">
            <div className="space-y-1 px-5 pb-3">
              {filteredTags.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => {
                    setIsCreating(false);
                    setSelectedTagId(tag.id);
                  }}
                  className={cn(
                    'flex w-full items-center justify-between rounded-md px-3 py-2 text-left transition-colors hover:bg-muted',
                    effectiveSelectedTagId === tag.id && 'bg-primary/10'
                  )}
                >
                  <div className="flex min-w-0 items-center gap-2">
                    <span className="h-3 w-3 rounded-full" style={{ backgroundColor: tag.color }} />
                    <span className="truncate text-sm">{tag.name}</span>
                  </div>
                  <Badge variant="secondary" className="h-5 text-xs">{tag.problemCount}</Badge>
                </button>
              ))}
            </div>
          </ScrollArea>
        </PagePanel>

        <PagePanel className="flex flex-col overflow-hidden">
          <CardHeader className="shrink-0 border-b pb-4 pt-6">
            <CardTitle className="flex items-center gap-2">
              {selectedTag ? <Pencil className="h-5 w-5" /> : <Plus className="h-5 w-5" />}
              {selectedTag ? '编辑标签' : '新建标签'}
            </CardTitle>
          </CardHeader>

          <ScrollArea className="flex-1">
            <CardContent className="space-y-6 p-6">
              <div className="grid gap-6 xl:grid-cols-[1fr_360px]">
                <div className="space-y-6">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>名称</FormLabel>
                        <FormControl>
                          <Input {...field} placeholder="标签名称" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="category"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>分类</FormLabel>
                        <Select value={field.value} onValueChange={field.onChange}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="topic">知识点</SelectItem>
                            <SelectItem value="source">来源</SelectItem>
                            <SelectItem value="custom">自定义</SelectItem>
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="color"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>颜色</FormLabel>
                        <div className="flex flex-wrap gap-2">
                          {TAG_PRESET_COLORS.map((preset) => (
                            <button
                              key={preset}
                              type="button"
                              onClick={() => field.onChange(preset)}
                              className={cn('h-8 w-8 rounded-md transition-transform hover:scale-110', field.value === preset && 'ring-2 ring-offset-2 ring-primary')}
                              style={{ backgroundColor: preset }}
                            />
                          ))}
                        </div>
                        <FormControl>
                          <Input {...field} className="mt-3 max-w-[220px]" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>说明</FormLabel>
                        <FormControl>
                          <Textarea {...field} value={field.value ?? ''} rows={5} placeholder="标签用途说明" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" disabled={createTagMutation.isPending || updateTagMutation.isPending}>
                      保存标签
                    </Button>
                    {selectedTag ? (
                      <ConfirmActionButton
                        variant="destructive"
                        pending={deleteTagMutation.isPending}
                        onConfirm={async () => {
                          try {
                            await deleteTagMutation.mutateAsync(selectedTag.id);
                            setIsCreating(false);
                            setSelectedTagId(null);
                            form.reset(defaultTagValues);
                          } catch {
                            // toast handled by hook
                          }
                        }}
                        title="确认删除标签"
                        description={`确定要删除标签“${selectedTag.name}”吗？已绑定的题目将失去这个标签。`}
                        confirmLabel="确认删除"
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        删除
                      </ConfirmActionButton>
                    ) : null}
                  </div>
                </div>

                <div className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">标签摘要</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3 text-sm">
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">分类</span>
                        <span>{TAG_CATEGORY_LABELS[category]}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">题目数</span>
                        <span>{selectedTag?.problemCount ?? 0}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">颜色</span>
                        <span className="font-mono">{color}</span>
                      </div>
                    </CardContent>
                  </Card>

                  {selectedTag ? (
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base">标签合并</CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-3">
                        <Select value={mergeTargetId} onValueChange={setMergeTargetId}>
                          <SelectTrigger>
                            <SelectValue placeholder="选择目标标签" />
                          </SelectTrigger>
                          <SelectContent>
                            {tags.filter((tag) => tag.id !== selectedTag.id).map((tag) => (
                              <SelectItem key={tag.id} value={tag.id}>
                                {tag.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <ConfirmActionButton
                          variant="outline"
                          className="w-full"
                          disabled={!mergeTargetId || mergeTagMutation.isPending}
                          pending={mergeTagMutation.isPending}
                          onConfirm={async () => {
                            try {
                              await mergeTagMutation.mutateAsync({ id: selectedTag.id, targetId: mergeTargetId });
                              setIsCreating(false);
                              setMergeTargetId('');
                              setSelectedTagId(mergeTargetId);
                            } catch {
                              // toast handled by hook
                            }
                          }}
                          title="确认合并标签"
                          description={`确定要将“${selectedTag.name}”合并到目标标签吗？合并后当前标签会被移除。`}
                          confirmLabel="确认合并"
                        >
                          <Merge className="mr-2 h-4 w-4" />
                          合并到目标标签
                        </ConfirmActionButton>
                      </CardContent>
                    </Card>
                  ) : null}
                </div>
              </div>

              <div className="space-y-3">
                <h3 className="font-medium">使用此标签的题目</h3>
                <div className="grid gap-3">
                  {(relatedProblemsQuery.data?.items ?? []).map((problem) => (
                    <Link key={problem.id} href={`/problems/${problem.id}`} className="rounded-2xl border p-4 transition-colors hover:border-primary">
                      <p className="mb-2 font-mono text-xs text-muted-foreground">{problem.code}</p>
                      <p className="line-clamp-2 text-sm leading-7">{problem.latex}</p>
                    </Link>
                  ))}
                  {selectedTag && (relatedProblemsQuery.data?.items?.length ?? 0) === 0 ? (
                    <Empty>
                      <EmptyMedia variant="icon">
                        <FileText className="h-6 w-6" />
                      </EmptyMedia>
                      <EmptyHeader>
                        <EmptyTitle>当前标签下还没有题目</EmptyTitle>
                      </EmptyHeader>
                    </Empty>
                  ) : null}
                </div>
              </div>
            </CardContent>
          </ScrollArea>
        </PagePanel>
          </div>
        </PageShell>
      </form>
    </Form>
  );
}
