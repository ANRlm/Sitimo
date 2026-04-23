'use client';

import { Monitor } from 'lucide-react';

export function MobileBlocker() {
  return (
    <div className="fixed inset-0 z-[100] flex flex-col items-center justify-center bg-background p-8 md:hidden">
      <Monitor className="mb-6 h-20 w-20 text-primary" />
      <h1 className="mb-3 text-center text-2xl font-semibold text-foreground">
        请在电脑上使用 MathLib
      </h1>
      <p className="max-w-sm text-center leading-relaxed text-muted-foreground">
        本应用需要较大屏幕来编辑公式和组排试卷，请使用笔记本或台式机浏览器访问。
      </p>
    </div>
  );
}
