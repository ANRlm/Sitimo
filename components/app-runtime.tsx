'use client';

import { useEffect } from 'react';
import { useExports, useExportStream } from '@/lib/hooks/use-exports';
import { useExportStore } from '@/lib/store';

export function AppRuntime() {
  const syncFromJobs = useExportStore((state) => state.syncFromJobs);
  const exportsQuery = useExports({ page: 1, pageSize: 20 });

  useExportStream();

  useEffect(() => {
    if (exportsQuery.data?.items) {
      syncFromJobs(exportsQuery.data.items);
    }
  }, [exportsQuery.data?.items, syncFromJobs]);

  return null;
}
