'use client';

import { MathJax } from 'better-react-mathjax';
import { cn } from '@/lib/utils';

type MathTextProps = {
  latex: string;
  inline?: boolean;
  className?: string;
};

export function MathText({ latex, inline = false, className }: MathTextProps) {
  return (
    <div className={cn('math-text max-w-full', className)}>
      <MathJax inline={inline}>{latex}</MathJax>
    </div>
  );
}
