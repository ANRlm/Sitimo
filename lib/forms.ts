'use client';

import type { FieldValues, Path, UseFormReturn } from 'react-hook-form';
import { AppError } from '@/lib/api/client';

type ValidationDetail = {
  field: string;
  message: string;
};

export function applyValidationErrors<TFieldValues extends FieldValues>(
  form: UseFormReturn<TFieldValues>,
  error: unknown
) {
  if (!(error instanceof AppError) || error.code !== 'validation_failed' || !Array.isArray(error.details)) {
    return;
  }

  for (const detail of error.details as ValidationDetail[]) {
    if (!detail?.field || !detail?.message) {
      continue;
    }
    form.setError(detail.field as Path<TFieldValues>, { message: detail.message });
  }
}
