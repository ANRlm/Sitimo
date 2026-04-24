'use client';

export class AppError extends Error {
  code: string;
  status: number;
  details: unknown;

  constructor(message: string, code = 'unknown_error', status = 500, details: unknown = null) {
    super(message);
    this.name = 'AppError';
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

type ApiEnvelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
};

const API_BASE = '/api/v1';

type RequestOptions = RequestInit & {
  query?: Record<string, string | number | boolean | undefined | null>;
};

function withQuery(path: string, query?: RequestOptions['query']) {
  let url = `${API_BASE}${path}`;
  if (query) {
    const params = new URLSearchParams();
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') {
        return;
      }
      params.set(key, String(value));
    });
    const qs = params.toString();
    if (qs) {
      url = `${url}${url.includes('?') ? '&' : '?'}${qs}`;
    }
  }
  return url;
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { query, headers, body, ...rest } = options;
  const isFormData = typeof FormData !== 'undefined' && body instanceof FormData;

  const response = await fetch(withQuery(path, query), {
    ...rest,
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...headers,
    },
    body,
    cache: 'no-store',
  });

  const payload = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || payload.error) {
    throw new AppError(payload.error?.message ?? `请求失败: ${response.status}`, payload.error?.code, response.status, payload.data);
  }

  return payload.data;
}

export { API_BASE };
