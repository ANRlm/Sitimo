'use client';

import React, { use, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useForm, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  closestCenter,
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { ArrowLeft, Download, FileText, GripVertical, Plus, Save, ShoppingBasket, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { MathText } from '@/components/math-text';
import { PaperDocument } from '@/components/paper-document';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Slider } from '@/components/ui/slider';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { useCreateExport } from '@/lib/hooks/use-exports';
import { useCreatePaper, usePaper, useUpdatePaper } from '@/lib/hooks/use-papers';
import { useProblems } from '@/lib/hooks/use-problems';
import { applyValidationErrors } from '@/lib/forms';
import { paperSchema, type PaperFormValues } from '@/lib/schemas/paper';
import { useBasketStore } from '@/lib/store';
import type { PaperDetail, PaperItemDetail, Problem, ProblemDetail } from '@/lib/types';

type EditorMode = 'edit' | 'preview';
type EditorItem = PaperItemDetail;

const defaultLayout = {
  columns: 1 as 1 | 2,
  fontSize: 12,
  lineHeight: 1.4,
  paperSize: 'A4' as const,
  showAnswerVersion: true,
};

const defaultPaperValues: PaperFormValues = {
  title: '未命名试卷',
  subtitle: '',
  schoolName: '北京市第一中学',
  examName: '阶段测试',
  duration: '120',
  description: '',
  status: 'draft',
  instructions: '请认真作答，将答案写在答题纸上。',
  footerText: 'Sitimo 自动组卷',
  columns: '1',
  fontSize: '12',
  lineHeight: '1.4',
  paperSize: 'A4',
  showAnswerVersion: true,
};

function ensureProblemDetail(problem: Problem | ProblemDetail): ProblemDetail {
  if ('tags' in problem && 'images' in problem) {
    return problem;
  }
  return {
    ...problem,
    tags: [],
    images: [],
  };
}

function createEditorItem(problem: Problem | ProblemDetail, orderIndex: number): EditorItem {
  return {
    id: `item-${problem.id}-${Date.now()}-${orderIndex}`,
    problemId: problem.id,
    score: 10,
    orderIndex,
    imagePosition: 'below',
    blankLines: 0,
    problem: ensureProblemDetail(problem),
  };
}

function basketProblemToProblem(item: { id: string; code: string; latex: string; difficulty: Problem['difficulty'] }): Problem {
  return {
    id: item.id,
    code: item.code,
    latex: item.latex,
    difficulty: item.difficulty,
    type: 'solve',
    tagIds: [],
    imageIds: [],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    version: 1,
    isDeleted: false,
  };
}

function emptyToUndefined(value: string | undefined) {
  const normalized = value?.trim();
  return normalized ? normalized : undefined;
}

function parseOptionalInt(value: string | undefined) {
  const normalized = value?.trim();
  if (!normalized) {
    return undefined;
  }
  const parsed = Number.parseInt(normalized, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function parseOptionalFloat(value: string | undefined) {
  const normalized = value?.trim();
  if (!normalized) {
    return undefined;
  }
  const parsed = Number.parseFloat(normalized);
  return Number.isFinite(parsed) ? parsed : undefined;
}

const SortableItemCard = React.memo(function SortableItemCard({
  item,
  index,
  onScoreChange,
  onPositionChange,
  onBlankLinesChange,
  onRemove,
}: {
  item: EditorItem;
  index: number;
  onScoreChange: (value: number) => void;
  onPositionChange: (value: 'inline' | 'below' | 'right') => void;
  onBlankLinesChange: (value: number) => void;
  onRemove: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: item.id });

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={`rounded-2xl border bg-card p-4 ${isDragging ? 'shadow-lg ring-2 ring-primary/20' : ''}`}
    >
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Button type="button" variant="ghost" size="icon" className="h-8 w-8 cursor-grab" {...attributes} {...listeners}>
            <GripVertical className="h-4 w-4" />
          </Button>
          <div>
            <p className="font-medium">{index + 1}. {item.problem?.code ?? item.problemId}</p>
            <p className="text-xs text-muted-foreground">拖拽可调整题序</p>
          </div>
        </div>
        <Button type="button" variant="ghost" size="icon" className="text-destructive" onClick={onRemove}>
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <div className="space-y-4">
        {item.problem ? <MathText latex={item.problem.latex} className="leading-7" /> : <div className="text-sm text-muted-foreground">题目内容缺失</div>}
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <p className="mb-2 text-sm font-medium">分值</p>
            <Input value={String(item.score)} onChange={(event) => onScoreChange(Number(event.target.value) || 0)} />
          </div>
          <div>
            <p className="mb-2 text-sm font-medium">图像位置</p>
            <Select value={item.imagePosition ?? 'below'} onValueChange={(value) => onPositionChange(value as 'inline' | 'below' | 'right')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="inline">内联</SelectItem>
                <SelectItem value="below">下方</SelectItem>
                <SelectItem value="right">右侧</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div>
          <div className="mb-2 flex items-center justify-between">
            <p className="text-sm font-medium">留空行</p>
            <span className="text-sm text-muted-foreground">{item.blankLines ?? 0} 行</span>
          </div>
          <Slider
            value={[item.blankLines ?? 0]}
            onValueChange={(values) => onBlankLinesChange(values[0])}
            min={0}
            max={10}
            step={1}
          />
        </div>
      </div>
    </div>
  );
});

export default function PaperEditorPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const isNew = id === 'new';
  const router = useRouter();
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  const paperQuery = usePaper(isNew ? undefined : id);
  const createPaperMutation = useCreatePaper();
  const updatePaperMutation = useUpdatePaper(id);
  const createExportMutation = useCreateExport();

  const basketItems = useBasketStore((state) => state.items);
  const clearBasket = useBasketStore((state) => state.clearBasket);

  const [initializedKey, setInitializedKey] = useState<string | null>(() => (isNew ? 'new' : null));
  const [mode, setMode] = useState<EditorMode>('edit');
  const [searchDialogOpen, setSearchDialogOpen] = useState(false);
  const [problemKeyword, setProblemKeyword] = useState('');
  const [items, setItems] = useState<EditorItem[]>(() =>
    isNew ? basketItems.map((item, index) => createEditorItem(basketProblemToProblem(item), index)) : []
  );

  const form = useForm<PaperFormValues>({
    resolver: zodResolver(paperSchema),
    defaultValues: defaultPaperValues,
  });

  const liveSearchQuery = useProblems({ keyword: problemKeyword, page: 1, pageSize: 24 });

  useEffect(() => {
    if (paperQuery.data && initializedKey !== paperQuery.data.id) {
      const paper = paperQuery.data;
      form.reset({
        title: paper.title,
        subtitle: paper.subtitle ?? '',
        schoolName: paper.schoolName ?? '',
        examName: paper.examName ?? '',
        duration: paper.duration ? String(paper.duration) : '',
        description: paper.description ?? '',
        instructions: paper.instructions ?? '',
        footerText: paper.footerText ?? '',
        status: paper.status,
        columns: paper.layout.columns === 2 ? '2' : '1',
        fontSize: String(paper.layout.fontSize),
        lineHeight: String(paper.layout.lineHeight),
        paperSize: paper.layout.paperSize,
        showAnswerVersion: paper.layout.showAnswerVersion,
      });
      const normalizedItems = (paper.itemDetails.length
        ? paper.itemDetails
        : paper.items.map((item) => ({
            ...item,
            problem: undefined,
          })))
        .slice()
        .sort((left, right) => left.orderIndex - right.orderIndex)
        .map((item) => ({
          ...item,
          problem: item.problem ? ensureProblemDetail(item.problem) : undefined,
        }));
      setItems(normalizedItems);
      setInitializedKey(paper.id);
    }
  }, [form, initializedKey, paperQuery.data]);

  const availableProblems = useMemo(() => liveSearchQuery.data?.items ?? [], [liveSearchQuery.data?.items]);
  const totalScore = useMemo(() => items.reduce((sum, item) => sum + Number(item.score || 0), 0), [items]);

  const watched = useWatch({ control: form.control });
  const title = watched.title;
  const subtitle = watched.subtitle;
  const schoolName = watched.schoolName;
  const examName = watched.examName;
  const duration = watched.duration;
  const description = watched.description;
  const instructions = watched.instructions;
  const footerText = watched.footerText;
  const status = watched.status;
  const columns = watched.columns;
  const fontSize = watched.fontSize;
  const lineHeight = watched.lineHeight;
  const paperSize = watched.paperSize;
  const showAnswerVersion = watched.showAnswerVersion;

  const previewLayout = useMemo(
    () => ({
      columns: (columns === '2' ? 2 : 1) as 1 | 2,
      fontSize: parseOptionalInt(fontSize) ?? defaultLayout.fontSize,
      lineHeight: parseOptionalFloat(lineHeight) ?? defaultLayout.lineHeight,
      paperSize: paperSize ?? defaultLayout.paperSize,
      showAnswerVersion: showAnswerVersion ?? defaultLayout.showAnswerVersion,
    }),
    [columns, fontSize, lineHeight, paperSize, showAnswerVersion]
  );

  const previewPaper: PaperDetail = useMemo(
    () => ({
      id: paperQuery.data?.id ?? 'draft-paper',
      title: title ?? '未命名试卷',
      subtitle: emptyToUndefined(subtitle),
      schoolName: emptyToUndefined(schoolName),
      examName: emptyToUndefined(examName),
      subject: undefined,
      duration: parseOptionalInt(duration),
      totalScore,
      description: emptyToUndefined(description),
      status: status ?? 'draft',
      instructions: emptyToUndefined(instructions),
      footerText: emptyToUndefined(footerText),
      header: {},
      items: items.map((item, index) => ({
        id: item.id,
        problemId: item.problemId,
        score: Number(item.score || 0),
        orderIndex: index,
        imagePosition: item.imagePosition,
        blankLines: item.blankLines,
      })),
      itemDetails: items.map((item, index) => ({
        ...item,
        score: Number(item.score || 0),
        orderIndex: index,
      })),
      layout: previewLayout,
      createdAt: paperQuery.data?.createdAt ?? new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }),
    [
      description,
      duration,
      examName,
      footerText,
      instructions,
      items,
      paperQuery.data?.createdAt,
      paperQuery.data?.id,
      previewLayout,
      schoolName,
      status,
      subtitle,
      title,
      totalScore,
    ]
  );

  if (!isNew && !paperQuery.isLoading && !paperQuery.data) {
    router.replace('/papers');
    return null;
  }

  const savePaper = async (values: PaperFormValues) => {
    const payload = {
      title: values.title.trim(),
      subtitle: emptyToUndefined(values.subtitle),
      schoolName: emptyToUndefined(values.schoolName),
      examName: emptyToUndefined(values.examName),
      subject: undefined,
      duration: parseOptionalInt(values.duration),
      totalScore,
      description: emptyToUndefined(values.description),
      status: values.status,
      instructions: emptyToUndefined(values.instructions),
      footerText: emptyToUndefined(values.footerText),
      items: items.map((item, index) => ({
        id: item.id,
        problemId: item.problemId,
        score: Number(item.score || 0),
        orderIndex: index,
        imagePosition: item.imagePosition,
        blankLines: item.blankLines,
      })),
      layout: {
        columns: (values.columns === '2' ? 2 : 1) as 1 | 2,
        fontSize: parseOptionalInt(values.fontSize) ?? defaultLayout.fontSize,
        lineHeight: parseOptionalFloat(values.lineHeight) ?? defaultLayout.lineHeight,
        paperSize: values.paperSize,
        showAnswerVersion: values.showAnswerVersion,
      },
    };

    try {
      const result = isNew
        ? await createPaperMutation.mutateAsync(payload)
        : await updatePaperMutation.mutateAsync(payload);
      if (isNew) {
        router.replace(`/papers/${result.id}/editor`);
      }
    } catch (error) {
      applyValidationErrors(form, error);
    }
  };

  const queueExport = (format: 'pdf' | 'latex') => {
    if (isNew || !paperQuery.data) {
      toast.error('请先保存试卷后再导出');
      return;
    }
    createExportMutation.mutate({
      paperId: paperQuery.data.id,
      format,
      variant: showAnswerVersion ? 'both' : 'student',
    });
  };

  const appendProblem = (problem: Problem | ProblemDetail) => {
    setItems((current) => {
      if (current.some((item) => item.problemId === problem.id)) {
        toast.info('该题目已在当前试卷中');
        return current;
      }
      return [...current, createEditorItem(problem, current.length)];
    });
    setSearchDialogOpen(false);
    setProblemKeyword('');
  };

  const appendBasketItems = () => {
    if (basketItems.length === 0) {
      toast.info('题目篮子为空');
      return;
    }

    let appended = 0;
    setItems((current) => {
      const next = [...current];
      basketItems.forEach((item) => {
        if (next.some((entry) => entry.problemId === item.id)) {
          return;
        }
        appended += 1;
        next.push(
          createEditorItem(basketProblemToProblem(item), next.length)
        );
      });
      return next;
    });

    if (appended === 0) {
      toast.info('篮子中的题目已经全部在试卷里');
      return;
    }

    toast.success(`已加入 ${appended} 道题目`);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }

    setItems((current) => {
      const oldIndex = current.findIndex((item) => item.id === active.id);
      const newIndex = current.findIndex((item) => item.id === over.id);
      return arrayMove(current, oldIndex, newIndex).map((item, index) => ({
        ...item,
        orderIndex: index,
      }));
    });
  };

  return (
    <Form {...form}>
      <div className="flex h-[calc(100vh-3.5rem)] flex-col">
        <div className="border-b border-border/70 bg-background/95 px-4 py-5 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-start">
            <div className="flex min-w-0 items-start gap-3">
              <Button
                variant="ghost"
                size="sm"
                asChild
                className="mt-2 h-10 rounded-xl border border-border/60 bg-card/70 px-3 text-foreground shadow-xs hover:bg-accent/80"
              >
                <Link href="/papers">
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  返回
                </Link>
              </Button>
              <div className="min-w-0 flex-1 rounded-3xl border border-border/70 bg-card/80 p-4 shadow-sm">
                <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span className="rounded-full border border-border/70 bg-background/80 px-2.5 py-1 font-medium text-foreground/80">
                    试卷编辑器
                  </span>
                  <span className="rounded-full border border-border/70 px-2.5 py-1">
                    {items.length} 题
                  </span>
                  <span className="rounded-full border border-border/70 px-2.5 py-1">
                    总分 {totalScore}
                  </span>
                  <span className="rounded-full border border-border/70 px-2.5 py-1">
                    {previewLayout.columns === 2 ? '双栏' : '单栏'} · {previewLayout.paperSize}
                  </span>
                </div>
                <div className="mt-3 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                  <FormField
                    control={form.control}
                    name="title"
                    render={({ field }) => (
                      <FormItem className="w-full max-w-[38rem] gap-1">
                        <FormControl>
                          <Input
                            {...field}
                            className="h-14 w-full rounded-2xl border-border/60 bg-background/90 px-5 text-xl font-semibold tracking-tight shadow-none focus-visible:border-primary/40 focus-visible:ring-primary/20 md:text-[1.75rem]"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="flex items-center gap-3 lg:justify-end">
                    <div className="text-right">
                      <p className="text-xs font-medium tracking-[0.18em] text-muted-foreground/80">
                        当前状态
                      </p>
                      <p className="mt-1 text-sm text-muted-foreground">
                        管理试卷内容、版式与导出设置
                      </p>
                    </div>
                    <Badge
                      variant="outline"
                      className="h-10 rounded-full border-amber-300/70 bg-amber-100/90 px-4 text-xs font-semibold text-amber-800 shadow-sm"
                    >
                      {isNew ? '新建中' : '编辑中'}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-3 rounded-3xl border border-border/70 bg-card/80 p-3 shadow-sm xl:min-w-[24rem]">
              <div className="flex flex-wrap items-center gap-2 xl:justify-end">
                <Tabs value={mode} onValueChange={(value) => setMode(value as EditorMode)}>
                  <TabsList className="rounded-xl border border-border/60 bg-background/80 p-1">
                    <TabsTrigger value="edit">编辑</TabsTrigger>
                    <TabsTrigger value="preview">即时预览</TabsTrigger>
                  </TabsList>
                </Tabs>

                {!isNew ? (
                  <Button variant="outline" className="rounded-xl bg-background/80" asChild>
                    <Link href={`/papers/${id}`}>已保存预览</Link>
                  </Button>
                ) : null}
              </div>

              <p className="text-xs text-muted-foreground xl:text-right">
                {isNew ? '即时预览展示当前未保存改动；保存后即可打开已保存预览。' : '即时预览展示当前未保存改动；已保存预览展示数据库中的当前版本。'}
              </p>

              <div className="flex flex-wrap items-center gap-2 xl:justify-end">
                <Button type="button" variant="outline" className="rounded-xl bg-background/80" onClick={() => queueExport('latex')} disabled={createExportMutation.isPending}>
                  <Download className="mr-2 h-4 w-4" />
                  导出 LaTeX 包
                </Button>
                <Button type="button" variant="outline" className="rounded-xl bg-background/80" onClick={() => queueExport('pdf')} disabled={createExportMutation.isPending}>
                  <Download className="mr-2 h-4 w-4" />
                  导出 PDF
                </Button>
                <Button type="button" className="rounded-xl" onClick={form.handleSubmit(savePaper)} disabled={createPaperMutation.isPending || updatePaperMutation.isPending}>
                  <Save className="mr-2 h-4 w-4" />
                  保存
                </Button>
              </div>
            </div>
          </div>
        </div>

        {mode === 'preview' ? (
          <div className="overflow-auto p-4">
            <PaperDocument paper={previewPaper} showAnswers={showAnswerVersion} />
          </div>
        ) : (
          <div className="grid min-h-0 flex-1 gap-4 p-6 xl:grid-cols-[360px_1fr]">
            <ScrollArea className="min-h-0">
              <div className="space-y-4 pr-4">
                <Card>
                  <CardHeader>
                    <CardTitle>基本信息</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <FormField
                      control={form.control}
                      name="subtitle"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>副标题</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="schoolName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>学校</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="examName"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>考试名</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="duration"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>时长（分钟）</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} inputMode="numeric" />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="status"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>状态</FormLabel>
                          <Select value={field.value} onValueChange={field.onChange}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="draft">草稿</SelectItem>
                              <SelectItem value="review">审核中</SelectItem>
                              <SelectItem value="completed">已完成</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="description"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>描述</FormLabel>
                          <FormControl>
                            <Textarea {...field} value={field.value ?? ''} rows={3} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="instructions"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>考试说明</FormLabel>
                          <FormControl>
                            <Textarea {...field} value={field.value ?? ''} rows={4} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="footerText"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>页脚</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>版式</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <FormField
                        control={form.control}
                        name="columns"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>分栏</FormLabel>
                            <Select value={field.value} onValueChange={field.onChange}>
                              <FormControl>
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                <SelectItem value="1">单栏</SelectItem>
                                <SelectItem value="2">双栏</SelectItem>
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name="paperSize"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>纸张</FormLabel>
                            <Select value={field.value} onValueChange={field.onChange}>
                              <FormControl>
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                <SelectItem value="A4">A4</SelectItem>
                                <SelectItem value="B5">B5</SelectItem>
                                <SelectItem value="Letter">Letter</SelectItem>
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                    <div className="grid gap-4 md:grid-cols-2">
                      <FormField
                        control={form.control}
                        name="fontSize"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>字号</FormLabel>
                            <FormControl>
                              <Input {...field} value={field.value ?? ''} inputMode="decimal" />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name="lineHeight"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>行距</FormLabel>
                            <FormControl>
                              <Input {...field} value={field.value ?? ''} inputMode="decimal" />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                    <FormField
                      control={form.control}
                      name="showAnswerVersion"
                      render={({ field }) => (
                        <FormItem className="gap-3 rounded-xl border p-3">
                          <div className="flex items-center justify-between">
                            <div>
                              <FormLabel>双版本导出</FormLabel>
                              <p className="text-sm text-muted-foreground">开启后导出时同时生成学生版和答案版</p>
                            </div>
                            <FormControl>
                              <Switch checked={field.value} onCheckedChange={field.onChange} />
                            </FormControl>
                          </div>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </CardContent>
                </Card>
              </div>
            </ScrollArea>

            <div className="flex min-h-0 flex-col rounded-3xl border bg-card">
              <div className="flex items-center justify-between gap-3 border-b px-5 py-4">
                <div>
                  <h2 className="font-semibold">题目列表</h2>
                  <p className="text-sm text-muted-foreground">共 {items.length} 题，当前总分 {totalScore}</p>
                </div>
                <div className="flex items-center gap-2">
                  <Button type="button" variant="outline" onClick={appendBasketItems}>
                    <ShoppingBasket className="mr-2 h-4 w-4" />
                    从篮子导入
                  </Button>
                  <Button type="button" variant="outline" onClick={clearBasket} disabled={basketItems.length === 0}>
                    清空篮子
                  </Button>
                  <Dialog open={searchDialogOpen} onOpenChange={setSearchDialogOpen}>
                    <DialogTrigger asChild>
                      <Button type="button">
                        <Plus className="mr-2 h-4 w-4" />
                        添加题目
                      </Button>
                    </DialogTrigger>
                    <DialogContent className="max-w-3xl">
                      <DialogHeader>
                        <DialogTitle>添加题目</DialogTitle>
                      </DialogHeader>
                      <div className="space-y-4">
                        <Input value={problemKeyword} onChange={(event) => setProblemKeyword(event.target.value)} placeholder="搜索题目编号、题干关键词或公式..." />
                        <div className="grid max-h-[420px] gap-3 overflow-auto">
                          {availableProblems.map((problem) => (
                            <button
                              key={problem.id}
                              type="button"
                              onClick={() => appendProblem(problem)}
                              className="rounded-2xl border p-4 text-left transition-colors hover:border-primary"
                            >
                              <p className="mb-2 font-mono text-xs text-muted-foreground">{problem.code}</p>
                              <MathText latex={problem.latex} className="leading-7" />
                            </button>
                          ))}
                        </div>
                      </div>
                    </DialogContent>
                  </Dialog>
                </div>
              </div>

              <ScrollArea className="min-h-0 flex-1">
                <div className="space-y-4 p-5">
                  {items.length === 0 ? (
                    <Empty>
                      <EmptyMedia variant="icon">
                        <FileText className="h-6 w-6" />
                      </EmptyMedia>
                      <EmptyHeader>
                        <EmptyTitle>还没有添加题目</EmptyTitle>
                        <EmptyDescription>请点击上方按钮添加。</EmptyDescription>
                      </EmptyHeader>
                    </Empty>
                  ) : (
                    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
                      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
                        <div className="space-y-4">
                          {items.map((item, index) => (
                            <SortableItemCard
                              key={item.id}
                              item={item}
                              index={index}
                              onScoreChange={(value) =>
                                setItems((current) =>
                                  current.map((entry) => (entry.id === item.id ? { ...entry, score: value } : entry))
                                )
                              }
                              onPositionChange={(value) =>
                                setItems((current) =>
                                  current.map((entry) => (entry.id === item.id ? { ...entry, imagePosition: value } : entry))
                                )
                              }
                              onBlankLinesChange={(value) =>
                                setItems((current) =>
                                  current.map((entry) => (entry.id === item.id ? { ...entry, blankLines: value } : entry))
                                )
                              }
                              onRemove={() =>
                                setItems((current) =>
                                  current
                                    .filter((entry) => entry.id !== item.id)
                                    .map((entry, nextIndex) => ({ ...entry, orderIndex: nextIndex }))
                                )
                              }
                            />
                          ))}
                        </div>
                      </SortableContext>
                    </DndContext>
                  )}
                </div>
              </ScrollArea>

              <Separator />
              <div className="flex items-center justify-between px-5 py-4 text-sm text-muted-foreground">
                <span>拖拽题目可调整顺序，保存后会同步到后端。</span>
                <span>导出默认使用 {showAnswerVersion ? '双版本' : '学生版'}。</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </Form>
  );
}
