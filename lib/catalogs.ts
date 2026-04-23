import { STANDARD_GRADES } from './constants';
import { problemTypeConfig, type Difficulty } from './types';
import type { ImportStep, PaperStatus, ProblemListSort } from './frontend-contracts';

export const SUBJECT_OPTIONS = ['全部', '数学', '物理', '化学', '生物'] as const;

export const GRADE_OPTIONS = ['全部', ...STANDARD_GRADES] as const;

export const PROBLEM_TYPE_FILTER_OPTIONS = [
  { value: 'all', label: '全部' },
  { value: 'multiple_choice', label: '选择' },
  { value: 'fill_blank', label: '填空' },
  { value: 'solve', label: '解答' },
  { value: 'proof', label: '证明' },
] as const;

export const PROBLEM_SORT_OPTIONS: Array<{ value: ProblemListSort; label: string }> = [
  { value: 'updatedAt-desc', label: '更新时间↓' },
  { value: 'createdAt-desc', label: '创建时间↓' },
  { value: 'code-asc', label: '编号↑' },
  { value: 'subjectiveScore-asc', label: '主观难度↑' },
];

export const DIFFICULTY_FILTER_OPTIONS: Array<{ value: Difficulty; label: string }> = [
  { value: 'easy', label: '容易' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' },
  { value: 'olympiad', label: '竞赛' },
];

export const IMPORT_STEP_TITLES = {
  1: '上传',
  2: '解析预览',
  3: '批量元信息',
} satisfies Record<ImportStep, string>;

export const TAG_PRESET_COLORS = [
  '#0F766E',
  '#0891B2',
  '#2563EB',
  '#7C3AED',
  '#C026D3',
  '#DC2626',
  '#EA580C',
  '#F59E0B',
  '#84CC16',
  '#10B981',
  '#6366F1',
  '#EC4899',
] as const;

export const TAG_CATEGORY_LABELS = {
  topic: '知识点',
  source: '来源',
  custom: '自定义',
} as const;

export const EXPORT_STATUS_FILTER_OPTIONS = [
  { value: 'pending', label: '排队中' },
  { value: 'processing', label: '生成中' },
  { value: 'done', label: '完成' },
  { value: 'failed', label: '失败' },
] as const;

export const PAPER_STATUS_LABELS: Record<PaperStatus, string> = {
  draft: '草稿',
  completed: '已完成',
  review: '审核中',
};

export const PAPER_STATUS_CLASSNAMES: Record<PaperStatus, string> = {
  draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200',
  completed: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300',
  review: 'bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300',
};

export const SEARCH_OPERATOR_LABELS = {
  eq: '=',
  contains: '包含',
  gt: '大于',
  lt: '小于',
  between: '介于',
} as const;

export const SEARCH_SORT_OPTIONS = PROBLEM_SORT_OPTIONS.filter(
  (option) => option.value !== 'subjectiveScore-asc'
);

export const PROBLEM_TYPE_OPTIONS = Object.entries(problemTypeConfig).map(([value, label]) => ({
  value,
  label,
}));
