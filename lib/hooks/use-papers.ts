'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createPaper, deletePaper, duplicatePaper, getPaper, listPapers, updatePaper, updatePaperItems, type PaperWriteInput } from '@/lib/api/papers';
import type { PaperItem } from '@/lib/types';

export function usePapers(query: { keyword?: string; page?: number; pageSize?: number }) {
  return useQuery({
    queryKey: ['papers', query],
    queryFn: () => listPapers(query),
  });
}

export function usePaper(id?: string) {
  return useQuery({
    queryKey: ['papers', id],
    queryFn: () => getPaper(id!),
    enabled: Boolean(id),
  });
}

export function useCreatePaper() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPaper,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['papers'] });
      toast.success('试卷已创建');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useUpdatePaper(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PaperWriteInput) => updatePaper(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['papers'] });
      toast.success('试卷已更新');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useUpdatePaperItems(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (items: PaperItem[]) => updatePaperItems(id, items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['papers'] });
      toast.success('试卷排序已保存');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeletePaper() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deletePaper,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['papers'] });
      toast.success('试卷已删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDuplicatePaper() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: duplicatePaper,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['papers'] });
      toast.success('试卷已复制');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}
