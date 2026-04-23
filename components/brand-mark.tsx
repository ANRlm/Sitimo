'use client';

import { cn } from '@/lib/utils';

type BrandMarkProps = {
  className?: string;
  iconClassName?: string;
};

export function BrandMark({ className, iconClassName }: BrandMarkProps) {
  return (
    <div
      className={cn(
        'flex h-8 w-8 items-center justify-center rounded-lg bg-linear-to-br from-primary via-primary to-primary/75 text-primary-foreground shadow-[0_10px_30px_-18px_hsl(var(--primary))]',
        className
      )}
      aria-hidden="true"
    >
      <svg
        viewBox="0 0 32 32"
        className={cn('h-5 w-5', iconClassName)}
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M7 24.5L7 8.5"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
        <path
          d="M7 24.5H24.5"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
        <path
          d="M10 21C11.2 18.1 12.9 16.2 15 16.2C17.6 16.2 17.8 22.3 20.6 22.3C22.1 22.3 23.4 20.4 25 16.8"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M17.8 7.6C16.6 8.6 15.9 10.3 15.9 12.4V23.4"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
        <path
          d="M15.1 10.5H19.8"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        />
      </svg>
    </div>
  );
}
