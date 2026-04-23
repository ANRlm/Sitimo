'use client';

import { useQuery } from '@tanstack/react-query';
import { getGrades, getMetaStats, getRecentExports, getRecentProblems } from '@/lib/api/meta';

export function useMetaStats() {
  return useQuery({
    queryKey: ['meta', 'stats'],
    queryFn: getMetaStats,
  });
}

export function useRecentProblems(limit = 5) {
  return useQuery({
    queryKey: ['meta', 'recent-problems', limit],
    queryFn: () => getRecentProblems(limit),
  });
}

export function useRecentExports(limit = 5) {
  return useQuery({
    queryKey: ['meta', 'recent-exports', limit],
    queryFn: () => getRecentExports(limit),
  });
}

export function useGrades() {
  return useQuery({
    queryKey: ['meta', 'grades'],
    queryFn: getGrades,
  });
}
