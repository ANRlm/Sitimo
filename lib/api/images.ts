'use client';

import type { ImageAsset, Paginated, ProblemDetail, Tag } from '@/lib/types';
import { apiRequest } from './client';

export type ImageListQuery = {
  keyword?: string;
  tagIds?: string;
  mime?: string;
  page?: number;
  pageSize?: number;
  deleted?: boolean;
};

export type ImageDetailResponse = {
  image: ImageAsset;
  linkedProblems: ProblemDetail[];
  tags: Tag[];
};

export type ImageEditInput = {
  crop?: { x: number; y: number; w: number; h: number };
  rotate?: number;
  resize?: { w: number; h: number };
};

function normalizeImage(image: ImageAsset): ImageAsset {
  return {
    ...image,
    tagIds: image.tagIds ?? [],
    linkedProblemIds: image.linkedProblemIds ?? [],
  };
}

function normalizeProblemDetail(problem: ProblemDetail): ProblemDetail {
  return {
    ...problem,
    tagIds: problem.tagIds ?? [],
    imageIds: problem.imageIds ?? [],
    warnings: problem.warnings ?? [],
    tags: problem.tags ?? [],
    images: problem.images ?? [],
  };
}

export async function listImages(query: ImageListQuery) {
  const data = await apiRequest<Paginated<ImageAsset>>('/images', { query });
  return {
    ...data,
    items: data.items.map(normalizeImage),
  };
}

export async function getImage(id: string) {
  const data = await apiRequest<ImageDetailResponse>(`/images/${id}`);
  return {
    image: normalizeImage(data.image),
    linkedProblems: (data.linkedProblems ?? []).map(normalizeProblemDetail),
    tags: data.tags ?? [],
  };
}

export async function uploadImage(input: { file: File; tagIds: string[]; description?: string }) {
  const formData = new FormData();
  formData.append('file', input.file);
  if (input.description) {
    formData.append('description', input.description);
  }
  if (input.tagIds.length > 0) {
    formData.append('tagIds', input.tagIds.join(','));
  }
  return normalizeImage(await apiRequest<ImageAsset>('/images', {
    method: 'POST',
    body: formData,
  }));
}

export async function updateImage(id: string, input: { tagIds: string[]; description?: string }) {
  return normalizeImage(await apiRequest<ImageAsset>(`/images/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }));
}

export async function deleteImage(id: string) {
  return apiRequest<{ ok: boolean }>(`/images/${id}`, { method: 'DELETE' });
}

export async function hardDeleteImage(id: string) {
  return apiRequest<{ ok: boolean }>(`/images/${id}/hard`, { method: 'DELETE' });
}

export async function restoreImage(id: string) {
  return apiRequest<{ ok: boolean }>(`/images/${id}/restore`, { method: 'POST' });
}

export async function editImage(id: string, input: ImageEditInput, problemId?: string) {
  return normalizeImage(await apiRequest<ImageAsset>(`/images/${id}/edit`, {
    method: 'POST',
    query: problemId ? { problemId } : undefined,
    body: JSON.stringify(input),
  }));
}

export async function batchDeleteImages(ids: string[]) {
  return apiRequest<{ ok: boolean; deleted: number }>('/images/batch-delete', {
    method: 'POST',
    body: JSON.stringify({ imageIds: ids }),
  });
}
