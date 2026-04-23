'use client';

import type { Tag } from '@/lib/types';
import { apiRequest } from './client';

export async function listTags() {
  return apiRequest<Tag[]>('/tags');
}

export async function createTag(input: Pick<Tag, 'name' | 'category' | 'color' | 'description'>) {
  return apiRequest<Tag>('/tags', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function updateTag(id: string, input: Pick<Tag, 'name' | 'category' | 'color' | 'description'>) {
  return apiRequest<Tag>(`/tags/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export async function deleteTag(id: string) {
  return apiRequest<{ ok: boolean }>(`/tags/${id}`, { method: 'DELETE' });
}

export async function mergeTag(id: string, targetId: string) {
  return apiRequest<{ ok: boolean }>(`/tags/${id}/merge`, {
    method: 'POST',
    body: JSON.stringify({ targetId }),
  });
}
