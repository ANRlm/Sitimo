'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { clearDemoData, exportAllData, getDemoDataStatus, getSettings, importAllData, loadDemoData, resetDemoData, sweepOrphans, updateSettings } from '@/lib/api/settings';

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      toast.success('设置已保存');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useResetDemoData() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: resetDemoData,
    onSuccess: () => {
      queryClient.invalidateQueries();
      toast.success('示例数据已重置');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useExportAllData() {
  return useMutation({
    mutationFn: exportAllData,
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useImportAllData() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: importAllData,
    onSuccess: () => {
      queryClient.invalidateQueries();
      toast.success('数据已导入');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useSweepOrphans() {
  return useMutation({
    mutationFn: sweepOrphans,
    onSuccess: (result) => {
      toast.success(result.deleted > 0 ? `已清理 ${result.deleted} 张孤儿图片` : '没有可清理的孤儿图片');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDemoDataStatus() {
  return useQuery({
    queryKey: ['demo-data-status'],
    queryFn: getDemoDataStatus,
  });
}

export function useLoadDemoData() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: loadDemoData,
    onSuccess: () => {
      queryClient.invalidateQueries();
      toast.success('示例数据已加载');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useClearDemoData() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: clearDemoData,
    onSuccess: () => {
      queryClient.invalidateQueries();
      toast.success('示例数据已删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}
