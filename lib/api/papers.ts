'use client';

import type { Paginated, PaperDetail, PaperItem, PaperStatus } from '@/lib/types';
import { apiRequest } from './client';

export type PaperWriteInput = {
  title: string;
  subtitle?: string;
  schoolName?: string;
  examName?: string;
  subject?: string;
  duration?: number;
  totalScore?: number;
  description?: string;
  status: PaperStatus;
  instructions?: string;
  footerText?: string;
  items: PaperItem[];
  layout: PaperDetail['layout'];
};

function normalizePaperDetail(paper: PaperDetail): PaperDetail {
  return {
    ...paper,
    items: paper.items ?? [],
    itemDetails: (paper.itemDetails ?? []).map((item) => ({
      ...item,
      problem: item.problem
        ? {
            ...item.problem,
            tagIds: item.problem.tagIds ?? [],
            imageIds: item.problem.imageIds ?? [],
            warnings: item.problem.warnings ?? [],
            tags: item.problem.tags ?? [],
            images: item.problem.images ?? [],
          }
        : undefined,
    })),
  };
}

export async function listPapers(query: { keyword?: string; page?: number; pageSize?: number }) {
  const data = await apiRequest<Paginated<PaperDetail>>('/papers', { query });
  return {
    ...data,
    items: data.items.map(normalizePaperDetail),
  };
}

export async function getPaper(id: string) {
  return normalizePaperDetail(await apiRequest<PaperDetail>(`/papers/${id}`));
}

export async function createPaper(input: PaperWriteInput) {
  return normalizePaperDetail(await apiRequest<PaperDetail>('/papers', {
    method: 'POST',
    body: JSON.stringify(input),
  }));
}

export async function updatePaper(id: string, input: PaperWriteInput) {
  return normalizePaperDetail(await apiRequest<PaperDetail>(`/papers/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }));
}

export async function updatePaperItems(id: string, items: PaperItem[]) {
  return normalizePaperDetail(await apiRequest<PaperDetail>(`/papers/${id}/items`, {
    method: 'PUT',
    body: JSON.stringify({ items }),
  }));
}

export async function deletePaper(id: string) {
  return apiRequest<{ ok: boolean }>(`/papers/${id}`, { method: 'DELETE' });
}

export async function duplicatePaper(id: string) {
  return normalizePaperDetail(await apiRequest<PaperDetail>(`/papers/${id}/duplicate`, { method: 'POST' }));
}
