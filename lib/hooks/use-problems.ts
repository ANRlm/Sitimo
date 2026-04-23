'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  batchDeleteProblems,
  batchTagProblems,
  commitBatchImport,
  createProblem,
  deleteProblem,
  getProblem,
  hardDeleteProblem,
  listProblems,
  listProblemVersions,
  previewBatchImport,
  restoreProblem,
  rollbackProblemVersion,
  updateProblem,
  type ProblemListQuery,
  type ProblemWriteInput,
} from '@/lib/api/problems';

export function useProblems(query: ProblemListQuery) {
  return useQuery({
    queryKey: ['problems', query],
    queryFn: () => listProblems(query),
  });
}

export function useProblem(id?: string) {
  return useQuery({
    queryKey: ['problems', id],
    queryFn: () => getProblem(id!),
    enabled: Boolean(id),
  });
}

export function useProblemVersions(id?: string) {
  return useQuery({
    queryKey: ['problem-versions', id],
    queryFn: () => listProblemVersions(id!),
    enabled: Boolean(id),
  });
}

export function useCreateProblem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProblemWriteInput) => createProblem(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      queryClient.invalidateQueries({ queryKey: ['meta'] });
      toast.success('题目已保存');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useUpdateProblem(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProblemWriteInput) => updateProblem(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      queryClient.invalidateQueries({ queryKey: ['problem-versions', id] });
      toast.success('题目已更新');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteProblem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteProblem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('题目已移入回收站');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useRestoreProblem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: restoreProblem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('题目已恢复');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useHardDeleteProblem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: hardDeleteProblem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('题目已彻底删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function usePreviewBatchImport() {
  return useMutation({
    mutationFn: previewBatchImport,
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useCommitBatchImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: commitBatchImport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('批量导入完成');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useBatchTagProblems() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ problemIds, tagIds, replace }: { problemIds: string[]; tagIds: string[]; replace: boolean }) =>
      batchTagProblems(problemIds, tagIds, replace),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('标签已批量更新');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useBatchDeleteProblems() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: batchDeleteProblems,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('批量删除完成');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useRollbackProblemVersion(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (version: number) => rollbackProblemVersion(id, version),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['problems', id] });
      queryClient.invalidateQueries({ queryKey: ['problem-versions', id] });
      toast.success('题目已回滚');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}
