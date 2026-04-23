'use client';

import type { ExportJob, MetaStats, Problem } from '@/lib/types';
import { apiRequest } from './client';

export async function getMetaStats() {
  return apiRequest<MetaStats>('/meta/stats');
}

export async function getRecentProblems(limit = 5) {
  return apiRequest<Problem[]>('/meta/recent-problems', { query: { limit } });
}

export async function getRecentExports(limit = 5) {
  return apiRequest<ExportJob[]>('/meta/recent-exports', { query: { limit } });
}

export async function getGrades() {
  return apiRequest<string[]>('/meta/grades');
}
