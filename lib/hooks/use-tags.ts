'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createTag, deleteTag, listTags, mergeTag, updateTag } from '@/lib/api/tags';
import type { Tag } from '@/lib/types';

export function useTags() {
  return useQuery({
    queryKey: ['tags'],
    queryFn: listTags,
  });
}

export function useCreateTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: Pick<Tag, 'name' | 'category' | 'color' | 'description'>) => createTag(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      toast.success('标签已创建');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useUpdateTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Pick<Tag, 'name' | 'category' | 'color' | 'description'> }) =>
      updateTag(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('标签已更新');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('标签已删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useMergeTag() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, targetId }: { id: string; targetId: string }) => mergeTag(id, targetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('标签已合并');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}
