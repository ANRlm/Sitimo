'use client';

import { MathJax } from 'better-react-mathjax';
import { normalizeLatexForDisplay } from '@/lib/latex';
import { cn } from '@/lib/utils';

type MathTextProps = {
  latex: string;
  inline?: boolean;
  className?: string;
};

export function MathText({ latex, inline = false, className }: MathTextProps) {
  const normalized = normalizeLatexForDisplay(latex);

  return (
    <div className={cn('math-text max-w-full', className)}>
      <MathJax inline={inline}>{normalized}</MathJax>
    </div>
  );
}
