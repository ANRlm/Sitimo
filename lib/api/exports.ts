'use client';

import type { ExportJob, Paginated } from '@/lib/types';
import { API_BASE, apiRequest } from './client';

export async function createExport(input: { paperId: string; format: ExportJob['format']; variant: ExportJob['variant'] }) {
  return apiRequest<ExportJob>('/exports', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function listExports(query: { status?: string; page?: number; pageSize?: number }) {
  return apiRequest<Paginated<ExportJob>>('/exports', { query });
}

export async function getExport(id: string) {
  return apiRequest<ExportJob>(`/exports/${id}`);
}

export async function deleteExport(id: string) {
  return apiRequest<{ ok: boolean }>(`/exports/${id}`, { method: 'DELETE' });
}

export function getExportDownloadUrl(id: string) {
  return `${API_BASE}/exports/${id}/download`;
}

export function getExportStreamUrl() {
  return `${API_BASE}/exports/stream`;
}
