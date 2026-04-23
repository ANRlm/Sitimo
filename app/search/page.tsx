'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Clock3, LayoutGrid, List, Plus, Save, Search, Trash2, X } from 'lucide-react';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { SEARCH_OPERATOR_LABELS, SEARCH_SORT_OPTIONS } from '@/lib/catalogs';
import { buildSearchPreview, formatRelativeTime } from '@/lib/format';
import type { ProblemListSort, SearchCondition, SearchField, SearchFieldConfig, SearchOperator } from '@/lib/frontend-contracts';
import { useCreateSavedSearch, useDeleteSavedSearch, useDeleteSearchHistory, useSavedSearches, useSearch, useSearchHistory } from '@/lib/hooks/use-search';
import { useTags } from '@/lib/hooks/use-tags';
import { createSearchCondition, createSearchFieldConfig, formatSearchConditionLabel } from '@/lib/problem-search';
import type { SavedSearchEntry, SearchHistoryEntry, SearchResult } from '@/lib/types';
import { difficultyConfig, problemTypeConfig } from '@/lib/types';
import { cn } from '@/lib/utils';

type DraftState = {
  field: SearchField;
  operator: SearchOperator;
  value: string;
  secondValue: string;
};

type SearchEntrySummary = {
  title: string;
  detail: string;
};

type SearchEntryFilters = {
  formula?: string;
  conditions?: SearchCondition[];
};

export default function SearchPage() {
  const tagsQuery = useTags();
  const savedSearchesQuery = useSavedSearches();
  const searchHistoryQuery = useSearchHistory();
  const createSavedSearchMutation = useCreateSavedSearch();
  const deleteSavedSearchMutation = useDeleteSavedSearch();
  const deleteSearchHistoryMutation = useDeleteSearchHistory();

  const [query, setQuery] = useState('');
  const [formula, setFormula] = useState('');
  const [sortBy, setSortBy] = useState<ProblemListSort>('updatedAt-desc');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [conditions, setConditions] = useState<SearchCondition[]>([]);
  const [conditionBuilderOpen, setConditionBuilderOpen] = useState(false);
  const [draft, setDraft] = useState<DraftState>({
    field: 'subject',
    operator: 'eq',
    value: '数学',
    secondValue: '',
  });

  const fieldConfig = useMemo<Record<SearchField, SearchFieldConfig>>(
    () => createSearchFieldConfig(tagsQuery.data ?? []),
    [tagsQuery.data]
  );

  const resultsQuery = useSearch(query, formula, conditions);
  const activeSearchCount = [Boolean(query.trim()), Boolean(formula.trim()), conditions.length > 0].filter(Boolean).length;
  const hasActiveSearch = activeSearchCount > 0;

  const results = useMemo(() => {
    const filtered = [...(resultsQuery.data ?? [])];
    const [field, direction] = sortBy.split('-');

    filtered.sort((left, right) => {
      let comparison = 0;
      if (field === 'code') {
        comparison = left.code.localeCompare(right.code);
      }
      if (field === 'updatedAt' || field === 'createdAt') {
        comparison = new Date(right[field]).getTime() - new Date(left[field]).getTime();
      }
      return direction === 'desc' ? comparison : -comparison;
    });

    return filtered;
  }, [resultsQuery.data, sortBy]);

  const visibleSearchHistory = useMemo(
    () => (searchHistoryQuery.data ?? []).filter((item) => isMeaningfulSearchEntry(item)),
    [searchHistoryQuery.data]
  );

  const resetDraft = (nextField: SearchField = 'subject') => {
    const config = fieldConfig[nextField];
    setDraft({
      field: nextField,
      operator: config.operators[0],
      value: config.options?.[0]?.value ?? (nextField === 'subject' ? '数学' : ''),
      secondValue: '',
    });
  };

  const addCondition = () => {
    if (!draft.value.trim()) {
      return;
    }
    if (draft.operator === 'between' && !draft.secondValue.trim()) {
      return;
    }
    setConditions((current) => [...current, createSearchCondition(draft.field, draft.operator, draft.value, draft.secondValue)]);
    setConditionBuilderOpen(false);
    resetDraft();
  };

  const restoreSearch = (searchEntry: SavedSearchEntry | SearchHistoryEntry) => {
    setQuery(searchEntry.query);
    const filters = readSearchEntryFilters(searchEntry);
    setFormula(filters.formula ?? '');
    setConditions(Array.isArray(filters.conditions) ? filters.conditions : []);
  };

  const clearSearch = () => {
    setQuery('');
    setFormula('');
    setConditions([]);
    resetDraft();
  };

  const saveCurrentSearch = async () => {
    if (!hasActiveSearch) {
      return;
    }
    await createSavedSearchMutation.mutateAsync({
      name: query.trim() || formula.trim() || `筛选 ${Date.now()}`,
      query,
      filters: {
        formula,
        conditions,
      },
    });
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="高级搜索"
        title="条件检索"
        description="支持关键词、公式和结构化筛选的组合搜索，首屏默认展示全库结果，不再预置会导致空结果的条件。"
        actions={
          <Button variant="outline" onClick={saveCurrentSearch} disabled={!hasActiveSearch || createSavedSearchMutation.isPending}>
            <Save className="mr-2 h-4 w-4" />
            保存此搜索
          </Button>
        }
      >
        <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_320px]">
          <div className="relative">
            <Search className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="输入关键词、题号或题干片段..." className="h-12 pl-11 text-base" />
          </div>
          <Input value={formula} onChange={(event) => setFormula(event.target.value)} placeholder="公式搜索，如：\\int_0^1 x^2 dx" className="h-12" />
        </div>
      </PageHeader>

      <PageToolbar className="space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          {conditions.map((condition) => (
            <Badge key={condition.id} variant="secondary" className="max-w-full gap-2 px-3 py-1">
              <span className="truncate">{formatSearchConditionLabel(condition, fieldConfig)}</span>
              <button type="button" onClick={() => setConditions((current) => current.filter((item) => item.id !== condition.id))}>
                <X className="h-3.5 w-3.5" />
              </button>
            </Badge>
          ))}

          <Popover
            open={conditionBuilderOpen}
            onOpenChange={(open) => {
              setConditionBuilderOpen(open);
              if (!open) {
                resetDraft();
              }
            }}
          >
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm">
                <Plus className="mr-2 h-4 w-4" />
                添加筛选条件
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[420px] space-y-3">
              <div className="grid gap-3">
                <div className="grid gap-3 md:grid-cols-3">
                  <Select
                    value={draft.field}
                    onValueChange={(value) => {
                      resetDraft(value as SearchField);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {(Object.keys(fieldConfig) as SearchField[]).map((field) => (
                        <SelectItem key={field} value={field}>
                          {fieldConfig[field].label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>

                  <Select value={draft.operator} onValueChange={(value) => setDraft((current) => ({ ...current, operator: value as SearchOperator, secondValue: '' }))}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {fieldConfig[draft.field].operators.map((operator) => (
                        <SelectItem key={operator} value={operator}>
                          {SEARCH_OPERATOR_LABELS[operator]}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>

                  <DynamicConditionInput
                    fieldConfig={fieldConfig}
                    field={draft.field}
                    operator={draft.operator}
                    value={draft.value}
                    secondValue={draft.secondValue}
                    onChange={(value) => setDraft((current) => ({ ...current, value }))}
                    onSecondChange={(value) => setDraft((current) => ({ ...current, secondValue: value }))}
                  />
                </div>

                <div className="flex justify-end gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setConditionBuilderOpen(false);
                      resetDraft();
                    }}
                  >
                    取消
                  </Button>
                  <Button size="sm" onClick={addCondition}>
                    确认
                  </Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>

          {hasActiveSearch ? (
            <Button variant="ghost" size="sm" onClick={clearSearch}>
              <X className="mr-2 h-4 w-4" />
              清空当前搜索
            </Button>
          ) : null}
        </div>

        <div className="flex flex-col gap-2 text-sm text-muted-foreground lg:flex-row lg:items-center lg:justify-between">
          <p>支持关键词 + LaTeX 公式 token 混合搜索，筛选条件之间是 AND 关系。</p>
          <span>{hasActiveSearch ? `当前启用 ${activeSearchCount} 组搜索条件。` : '当前处于默认全库浏览状态。'}</span>
        </div>
      </PageToolbar>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="space-y-4">
          <PageToolbar className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="space-y-1">
              <p className="text-sm font-medium text-foreground">搜索结果</p>
              <p className="text-sm text-muted-foreground">共 {results.length} 条结果，当前条件会即时参与排序与结果预览。</p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Select value={sortBy} onValueChange={(value) => setSortBy(value as ProblemListSort)}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {SEARCH_SORT_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <ToggleGroup type="single" value={viewMode} onValueChange={(value) => value && setViewMode(value as 'grid' | 'list')}>
                <ToggleGroupItem value="grid" size="sm">
                  <LayoutGrid className="h-4 w-4" />
                </ToggleGroupItem>
                <ToggleGroupItem value="list" size="sm">
                  <List className="h-4 w-4" />
                </ToggleGroupItem>
              </ToggleGroup>
            </div>
          </PageToolbar>

          {results.length === 0 && !resultsQuery.isLoading ? (
            <PagePanel>
              <Empty className="min-h-[320px] border-none bg-transparent">
                <EmptyMedia variant="icon">
                  <Search className="h-6 w-6" />
                </EmptyMedia>
                <EmptyHeader>
                  <EmptyTitle>当前条件下没有匹配结果</EmptyTitle>
                  <EmptyDescription>试试减少筛选条件、改短关键词，或者清空公式搜索。</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </PagePanel>
          ) : (
            <div className={cn('grid gap-4', viewMode === 'grid' ? 'xl:grid-cols-2' : 'grid-cols-1')}>
              {results.map((problem) => (
                <SearchResultCard key={problem.id} problem={problem} query={query || formula} compact={viewMode === 'list'} />
              ))}
            </div>
          )}
        </div>

        <div className="space-y-4 lg:sticky lg:top-20 lg:self-start">
          <PagePanel>
            <div className="border-b border-border/70 px-5 py-4">
              <h2 className="text-base font-semibold">已保存的搜索</h2>
              <p className="mt-1 text-sm text-muted-foreground">保存常用组合条件，后续可一键恢复。</p>
            </div>
            <div className="space-y-3 p-5">
              {(savedSearchesQuery.data ?? []).length > 0 ? (
                (savedSearchesQuery.data ?? []).map((item) => {
                  const summary = buildSearchEntrySummary(item, fieldConfig);

                  return (
                    <div key={item.id} className="rounded-2xl border border-border/70 bg-background/20 p-4">
                      <div className="space-y-1">
                        <p className="font-medium">{item.name}</p>
                        <p className="text-sm text-foreground">{summary.title}</p>
                        <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">{summary.detail}</p>
                      </div>
                      <div className="mt-3 flex items-center justify-between gap-3">
                        <span className="text-xs text-muted-foreground">{formatRelativeTime(item.createdAt)}</span>
                        <div className="flex gap-2">
                          <Button variant="outline" size="sm" onClick={() => restoreSearch(item)}>
                            恢复
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => deleteSavedSearchMutation.mutate(item.id)} disabled={deleteSavedSearchMutation.isPending}>
                            <Trash2 className="mr-2 h-4 w-4" />
                            删除
                          </Button>
                        </div>
                      </div>
                    </div>
                  );
                })
              ) : (
                <Empty className="border-none bg-transparent p-0">
                  <EmptyHeader>
                    <EmptyTitle className="text-base">还没有保存的搜索</EmptyTitle>
                    <EmptyDescription>配置好一组常用条件后，点击页面顶部的“保存此搜索”即可。</EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </div>
          </PagePanel>

          <PagePanel>
            <div className="border-b border-border/70 px-5 py-4">
              <h2 className="text-base font-semibold">搜索历史</h2>
              <p className="mt-1 text-sm text-muted-foreground">最近执行过的搜索会显示在这里。</p>
            </div>
            <div className="space-y-3 p-5">
              {visibleSearchHistory.length > 0 ? (
                visibleSearchHistory.map((item) => {
                  const summary = buildSearchEntrySummary(item, fieldConfig);

                  return (
                    <div key={item.id} className="rounded-2xl border border-border/70 bg-background/20 p-4">
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5 rounded-full border border-primary/20 bg-primary/10 p-2 text-primary">
                          <Clock3 className="h-4 w-4" />
                        </div>
                        <div className="min-w-0 flex-1 space-y-1">
                          <p className="truncate text-sm font-medium text-foreground">{summary.title}</p>
                          <p className="line-clamp-2 text-sm leading-6 text-muted-foreground">{summary.detail}</p>
                        </div>
                      </div>
                      <div className="mt-3 flex items-center justify-between gap-3">
                        <div className="space-y-1">
                          <p className="text-xs text-muted-foreground">{formatRelativeTime(item.createdAt)}</p>
                          <p className="text-xs text-muted-foreground">{item.resultCount} 条结果</p>
                        </div>
                        <div className="flex gap-2">
                          <Button variant="outline" size="sm" onClick={() => restoreSearch(item)}>
                            恢复
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => deleteSearchHistoryMutation.mutate(item.id)}>
                            <X className="mr-2 h-4 w-4" />
                            删除
                          </Button>
                        </div>
                      </div>
                    </div>
                  );
                })
              ) : (
                <Empty className="border-none bg-transparent p-0">
                  <EmptyHeader>
                    <EmptyTitle className="text-base">还没有搜索历史</EmptyTitle>
                    <EmptyDescription>执行关键词、公式或结构化筛选后，这里会保留最近的条件组合。</EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </div>
          </PagePanel>
        </div>
      </div>
    </PageShell>
  );
}

function SearchResultCard({ problem, query, compact }: { problem: SearchResult; query: string; compact: boolean }) {
  const highlights = buildSearchPreview(problem.snippet || problem.latex, query);
  const image = problem.images[0];

  return (
    <Link href={`/problems/${problem.id}`} className="block h-full">
      <PagePanel className="h-full overflow-hidden transition-colors hover:border-primary/50">
        <div className={cn('flex h-full flex-col gap-4', compact ? 'p-4' : 'p-5')}>
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-sm text-muted-foreground">{problem.code}</span>
            <Badge variant="outline">{difficultyConfig[problem.difficulty].label}</Badge>
            <Badge variant="secondary">{problemTypeConfig[problem.type]}</Badge>
            {problem.subject ? <Badge variant="secondary">{problem.subject}</Badge> : null}
            {problem.grade ? <Badge variant="secondary">{problem.grade}</Badge> : null}
            <span className="ml-auto text-xs text-muted-foreground">{formatRelativeTime(problem.updatedAt)}</span>
          </div>

          <div className={cn('grid gap-4', image ? 'grid-cols-[minmax(0,1fr)_84px]' : 'grid-cols-1')}>
            <div className="min-w-0 space-y-3">
              <div className={cn('overflow-hidden text-sm leading-7 text-foreground', compact ? 'min-h-[72px] max-h-[96px]' : 'min-h-[92px] max-h-[120px]')}>
                {highlights.map((part, index) =>
                  part.highlighted ? (
                    <mark key={`${part.text}-${index}`} className="rounded bg-accent/30 px-1">{part.text}</mark>
                  ) : (
                    <span key={`${part.text}-${index}`}>{part.text}</span>
                  )
                )}
              </div>
              {problem.tags.length > 0 ? (
                <div className="flex min-h-[2.5rem] flex-wrap content-start gap-2 overflow-hidden">
                  {problem.tags.slice(0, 3).map((tag) => (
                    <Badge key={tag.id} variant="secondary" className="text-xs">{tag.name}</Badge>
                  ))}
                </div>
              ) : null}
            </div>

            {image ? (
              <div className="overflow-hidden rounded-2xl border border-border/70 bg-background/20">
                <Image
                  src={image.thumbnailUrl}
                  alt={image.filename}
                  width={image.width}
                  height={image.height}
                  unoptimized
                  className="h-[84px] w-[84px] object-cover"
                />
              </div>
            ) : null}
          </div>
        </div>
      </PagePanel>
    </Link>
  );
}

function DynamicConditionInput({
  fieldConfig,
  field,
  operator,
  value,
  secondValue,
  onChange,
  onSecondChange,
}: {
  fieldConfig: Record<SearchField, SearchFieldConfig>;
  field: SearchField;
  operator: SearchOperator;
  value: string;
  secondValue: string;
  onChange: (value: string) => void;
  onSecondChange: (value: string) => void;
}) {
  const config = fieldConfig[field];
  const inputType = config.type === 'number' ? 'number' : config.type === 'date' ? 'date' : 'text';

  if (config.type === 'select' && config.options) {
    return (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder="值" />
        </SelectTrigger>
        <SelectContent>
          {config.options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }

  if (operator === 'between') {
    return (
      <div className="grid gap-2">
        <Input value={value} onChange={(event) => onChange(event.target.value)} type={inputType} placeholder="起始值" />
        <Input value={secondValue} onChange={(event) => onSecondChange(event.target.value)} type={inputType} placeholder="结束值" />
      </div>
    );
  }

  return <Input value={value} onChange={(event) => onChange(event.target.value)} type={inputType} placeholder="值" />;
}

function readSearchEntryFilters(entry: SavedSearchEntry | SearchHistoryEntry): SearchEntryFilters {
  return (entry.filters ?? {}) as SearchEntryFilters;
}

function isMeaningfulSearchEntry(entry: SearchHistoryEntry) {
  const filters = readSearchEntryFilters(entry);
  const conditionCount = Array.isArray(filters.conditions) ? filters.conditions.filter((condition) => condition.value.trim()).length : 0;

  return Boolean(entry.query.trim() || filters.formula?.trim() || conditionCount > 0);
}

function buildSearchEntrySummary(
  entry: SavedSearchEntry | SearchHistoryEntry,
  fieldConfig: Record<SearchField, SearchFieldConfig>
): SearchEntrySummary {
  const filters = readSearchEntryFilters(entry);
  const keyword = entry.query.trim();
  const formula = filters.formula?.trim() ?? '';
  const conditionLabels = (filters.conditions ?? [])
    .filter((condition) => condition.value.trim())
    .map((condition) => formatSearchConditionLabel(condition, fieldConfig));

  const title = keyword || formula || conditionLabels[0] || '搜索条件';
  const detailParts: string[] = [];

  if (keyword && title !== keyword) {
    detailParts.push(`关键词：${keyword}`);
  }
  if (formula && title !== formula) {
    detailParts.push(`公式：${formula}`);
  }
  if (conditionLabels.length > 0) {
    const remaining = conditionLabels.filter((label) => label !== title);
    if (remaining.length > 0) {
      detailParts.push(remaining.slice(0, 2).join(' / '));
    }
    if (remaining.length > 2) {
      detailParts.push(`+${remaining.length - 2} 条条件`);
    }
  }

  return {
    title,
    detail: detailParts.join(' / ') || '仅包含当前主条件。',
  };
}
