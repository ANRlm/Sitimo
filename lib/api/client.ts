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

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/api/v1';

type RequestOptions = RequestInit & {
  query?: Record<string, string | number | boolean | undefined | null>;
};

function withQuery(path: string, query?: RequestOptions['query']) {
  const url = new URL(`${API_BASE}${path}`);

  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') {
        return;
      }
      url.searchParams.set(key, String(value));
    });
  }

  return url.toString();
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
