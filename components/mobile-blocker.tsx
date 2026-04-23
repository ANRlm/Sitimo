'use client';

import { Monitor } from 'lucide-react';
import { BrandMark } from '@/components/brand-mark';

export function MobileBlocker() {
  return (
    <div className="fixed inset-0 z-[100] flex flex-col items-center justify-center bg-background p-8 md:hidden">
      <div className="mb-6 flex items-center gap-4">
        <BrandMark className="h-14 w-14 rounded-2xl" iconClassName="h-8 w-8" />
        <Monitor className="h-12 w-12 text-primary/85" />
      </div>
      <h1 className="mb-3 text-center text-2xl font-semibold text-foreground">
        请在电脑上使用 Sitimo
      </h1>
      <p className="max-w-sm text-center leading-relaxed text-muted-foreground">
        本应用需要较大屏幕来编辑公式和组排试卷，请使用笔记本或台式机浏览器访问。
      </p>
    </div>
  );
}
