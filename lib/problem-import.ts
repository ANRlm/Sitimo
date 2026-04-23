import { IMPORT_STEP_TITLES } from './catalogs';
import type {
  ImportBatchMetadata,
  ImportStep,
  NormalizedParsedProblemDraft,
  ParsedProblemDraft,
} from './frontend-contracts';

export const MOCK_IMPORT_SOURCE = `\\begin{problem}
求定积分 \\(\\int_0^1 x^2\\,dx\\) 的值。
\\end{problem}

\\begin{problem}
在 \\(\\triangle ABC\\) 中，已知 \\(a=2\\)，\\(b=\\sqrt{3}\\)，\\(\\angle A=60^\\circ\\)，求 \\(\\angle B\\)。
\\end{problem}

\\begin{problem}
已知函数 \\(f(x)=\\ln x-\\dfrac12x^2+x\\)，求其极值。
\\end{problem}

\\begin{problem}
证明：若 \\(a+b=1\\)，则 \\(\\dfrac{1}{a}+\\dfrac{4}{b}\\) 存在最小值。
% 这一题故意缺失 \\end{problem}，用于展示错误状态`;

export const INITIAL_PARSED_PROBLEM_DRAFTS: ParsedProblemDraft[] = [
  {
    id: 'draft-1',
    title: '题目 #1',
    latex: '求定积分 \\(\\int_0^1 x^2\\,dx\\) 的值。',
    difficulty: 'easy',
    status: 'success',
    tagNames: ['积分', '定积分'],
  },
  {
    id: 'draft-2',
    title: '题目 #2',
    latex: '在 \\(\\triangle ABC\\) 中，已知 \\(a=2\\)，\\(b=\\sqrt{3}\\)，\\(\\angle A=60^\\circ\\)，求 \\(\\angle B\\)。',
    difficulty: 'medium',
    status: 'success',
    tagNames: ['三角函数'],
  },
  {
    id: 'draft-3',
    title: '题目 #3',
    latex: '已知函数 \\(f(x)=\\ln x-\\dfrac12x^2+x\\)，求其极值。',
    difficulty: 'medium',
    status: 'success',
    tagNames: ['函数', '导数'],
  },
  {
    id: 'draft-4',
    title: '题目 #4',
    latex: '证明：若 \\(a+b=1\\)，则 \\(\\dfrac{1}{a}+\\dfrac{4}{b}\\) 存在最小值。',
    difficulty: 'hard',
    status: 'error',
    error: '缺少 \\end{problem}，解析在第 17 行终止。',
    tagNames: ['不等式'],
  },
];

export const PROBLEM_IMPORT_STEP_TITLES = IMPORT_STEP_TITLES satisfies Record<ImportStep, string>;

export function normalizeProblemImportDrafts(
  drafts: readonly ParsedProblemDraft[],
  defaults: Pick<ImportBatchMetadata, 'subject' | 'grade' | 'source'>
): NormalizedParsedProblemDraft[] {
  return drafts.map((draft, index) => ({
    ...draft,
    code: `P-2026-${String(index + 1).padStart(4, '0')}`,
    subject: draft.subject ?? defaults.subject,
    grade: draft.grade ?? defaults.grade,
    source: draft.source ?? defaults.source,
  }));
}

export function applyImportBatchMetadata(
  drafts: readonly ParsedProblemDraft[],
  metadata: ImportBatchMetadata
): ParsedProblemDraft[] {
  return drafts.map((draft) => ({
    ...draft,
    subject: metadata.subject,
    grade: metadata.grade,
    source: metadata.source,
    difficulty: metadata.difficulty,
    tagNames: Array.from(new Set([...draft.tagNames, ...metadata.tagNames])),
  }));
}
