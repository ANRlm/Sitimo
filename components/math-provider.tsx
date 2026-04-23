'use client';

import type { ReactNode } from 'react';
import { MathJaxContext } from 'better-react-mathjax';

const mathJaxConfig = {
  loader: { load: ['[tex]/ams'] },
  tex: {
    packages: { '[+]': ['ams'] },
    inlineMath: [['\\(', '\\)']],
    displayMath: [['\\[', '\\]']],
  },
};

export function MathProvider({ children }: { children: ReactNode }) {
  return <MathJaxContext config={mathJaxConfig}>{children}</MathJaxContext>;
}
