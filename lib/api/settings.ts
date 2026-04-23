'use client';

import type { Paginated } from '@/lib/types';
import { apiRequest } from './client';

export async function getSettings() {
  return apiRequest<Record<string, unknown>>('/settings');
}

export async function updateSettings(input: Record<string, unknown>) {
  return apiRequest<{ ok: boolean }>('/settings', {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export async function resetDemoData() {
  return apiRequest<{ ok: boolean }>('/settings/demo-data', { method: 'POST' });
}

export async function exportAllData() {
  return apiRequest<Record<string, unknown>>('/settings/export-all', { method: 'POST' });
}

export async function importAllData(payload: Record<string, unknown>) {
  return apiRequest<{ ok: boolean }>('/settings/import-all', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function sweepOrphans() {
  return apiRequest<{ deleted: number; bytesFreed: number; imageIds: string[] }>('/settings/sweep-orphans', {
    method: 'POST',
  });
}

export async function loadDemoData() {
  return apiRequest<{ ok: boolean }>('/settings/demo-data/load', { method: 'POST' });
}

export async function clearDemoData() {
  return apiRequest<{ ok: boolean }>('/settings/demo-data/clear', { method: 'POST' });
}

export async function getDemoDataStatus() {
  return apiRequest<{ loaded: boolean }>('/settings/demo-data/status');
}
