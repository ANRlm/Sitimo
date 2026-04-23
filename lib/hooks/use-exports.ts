'use client';

import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createExport, deleteExport, getExportStreamUrl, listExports } from '@/lib/api/exports';
import { useExportStore } from '@/lib/store';
import type { ExportJob } from '@/lib/types';

export function useExports(query: { status?: string; page?: number; pageSize?: number }) {
  return useQuery({
    queryKey: ['exports', query],
    queryFn: () => listExports(query),
  });
}

export function useCreateExport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createExport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['exports'] });
      toast.success('导出任务已加入队列');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteExport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteExport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['exports'] });
      toast.success('导出任务请求已提交');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useExportStream() {
  const queryClient = useQueryClient();
  const syncFromJobs = useExportStore((state) => state.syncFromJobs);

  useEffect(() => {
    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let closed = false;

    const applyJob = (job: ExportJob) => {
      queryClient.setQueriesData<{ items: ExportJob[]; total: number; page: number; pageSize: number } | undefined>(
        { queryKey: ['exports'] },
        (current) => {
          if (!current) {
            return current;
          }
          const existing = current.items.filter((item) => item.id !== job.id);
          const items = [job, ...existing].sort((left, right) => right.createdAt.localeCompare(left.createdAt));
          return { ...current, items };
        }
      );
      const latest = queryClient.getQueriesData<{ items: ExportJob[]; total: number; page: number; pageSize: number }>({ queryKey: ['exports'] });
      const firstPage = latest.find((entry) => Array.isArray(entry[1]?.items))?.[1];
      if (firstPage?.items) {
        syncFromJobs(firstPage.items);
      } else {
        syncFromJobs([job]);
      }
    };

    const connect = () => {
      if (closed) {
        return;
      }

      eventSource = new EventSource(getExportStreamUrl());
      eventSource.onopen = () => {
        retryCount = 0;
      };

      eventSource.onmessage = (event) => {
        try {
          const job = JSON.parse(event.data) as ExportJob;
          if (!job?.id) {
            return;
          }
          applyJob(job);
        } catch {
          // Ignore non-job payloads pushed by manual notifications.
        }
      };

      eventSource.onerror = () => {
        eventSource?.close();
        if (closed) {
          return;
        }
        const delay = Math.min(10000, 500 * 2 ** retryCount);
        retryCount += 1;
        reconnectTimer = setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      closed = true;
      eventSource?.close();
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
    };
  }, [queryClient, syncFromJobs]);
}
