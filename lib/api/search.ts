'use client';

import type { SearchCondition } from '@/lib/frontend-contracts';
import type { SavedSearchEntry, SearchHistoryEntry, SearchResult } from '@/lib/types';
import { apiRequest } from './client';

function normalizeSearchResult(problem: SearchResult): SearchResult {
  return {
    ...problem,
    tagIds: problem.tagIds ?? [],
    imageIds: problem.imageIds ?? [],
    warnings: problem.warnings ?? [],
    tags: problem.tags ?? [],
    images: problem.images ?? [],
  };
}

export async function searchProblems(keyword: string, formula: string, conditions: SearchCondition[]) {
  const data = await apiRequest<SearchResult[]>('/search', {
    query: {
      keyword,
      formula,
      conditions: conditions.length > 0 ? JSON.stringify(conditions) : undefined,
    },
  });
  return data.map(normalizeSearchResult);
}

export async function listSearchHistory(limit = 20) {
  return apiRequest<SearchHistoryEntry[]>('/search/history', { query: { limit } });
}

export async function deleteSearchHistory(id: string) {
  return apiRequest<{ ok: boolean }>(`/search/history/${id}`, { method: 'DELETE' });
}

export async function listSavedSearches() {
  return apiRequest<SavedSearchEntry[]>('/search/saved');
}

export async function createSavedSearch(input: { name: string; query: string; filters: Record<string, unknown> }) {
  return apiRequest<SavedSearchEntry>('/search/saved', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function deleteSavedSearch(id: string) {
  return apiRequest<{ ok: boolean }>(`/search/saved/${id}`, { method: 'DELETE' });
}
