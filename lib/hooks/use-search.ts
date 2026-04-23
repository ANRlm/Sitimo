'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createSavedSearch, deleteSavedSearch, deleteSearchHistory, listSavedSearches, listSearchHistory, searchProblems } from '@/lib/api/search';
import type { SearchCondition } from '@/lib/frontend-contracts';

export function useSearch(keyword: string, formula: string, conditions: SearchCondition[]) {
  return useQuery({
    queryKey: ['search', keyword, formula, conditions],
    queryFn: () => searchProblems(keyword, formula, conditions),
  });
}

export function useSearchHistory(limit = 20) {
  return useQuery({
    queryKey: ['search-history', limit],
    queryFn: () => listSearchHistory(limit),
  });
}

export function useSavedSearches() {
  return useQuery({
    queryKey: ['saved-searches'],
    queryFn: listSavedSearches,
  });
}

export function useCreateSavedSearch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createSavedSearch,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['saved-searches'] });
      toast.success('搜索条件已保存');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteSavedSearch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteSavedSearch,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['saved-searches'] });
      toast.success('已删除保存的搜索');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteSearchHistory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteSearchHistory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['search-history'] });
      toast.success('搜索历史已删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}
