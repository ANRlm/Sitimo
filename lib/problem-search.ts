import { GRADE_OPTIONS, SEARCH_OPERATOR_LABELS } from './catalogs';
import type {
  SavedSearch,
  SearchCondition,
  SearchField,
  SearchFieldConfig,
  SearchOperator,
} from './frontend-contracts';
import { difficultyConfig, problemTypeConfig, type Difficulty, type Problem } from './types';
import type { Tag } from './types';

export function createSearchFieldConfig(tags: readonly Tag[]): Record<SearchField, SearchFieldConfig> {
  return {
    subject: {
      label: '学科',
      type: 'select',
      operators: ['eq'],
      options: [{ label: '数学', value: '数学' }],
    },
    grade: {
      label: '年级',
      type: 'select',
      operators: ['eq'],
      options: GRADE_OPTIONS.filter((grade) => grade !== '全部').map((grade) => ({
        label: grade,
        value: grade,
      })),
    },
    difficulty: {
      label: '难度',
      type: 'select',
      operators: ['eq'],
      options: (Object.keys(difficultyConfig) as Difficulty[]).map((difficulty) => ({
        label: difficultyConfig[difficulty].label,
        value: difficulty,
      })),
    },
    subjectiveScore: {
      label: '主观难度',
      type: 'number',
      operators: ['eq', 'gt', 'lt', 'between'],
    },
    type: {
      label: '题型',
      type: 'select',
      operators: ['eq'],
      options: Object.entries(problemTypeConfig).map(([value, label]) => ({
        value,
        label,
      })),
    },
    hasImage: {
      label: '有无图像',
      type: 'select',
      operators: ['eq'],
      options: [
        { label: '有图', value: 'yes' },
        { label: '无图', value: 'no' },
      ],
    },
    tag: {
      label: '标签',
      type: 'select',
      operators: ['eq', 'contains'],
      options: tags.map((tag) => ({ label: tag.name, value: tag.id })),
    },
    source: {
      label: '来源',
      type: 'text',
      operators: ['eq', 'contains'],
    },
    date: {
      label: '日期',
      type: 'date',
      operators: ['eq', 'gt', 'lt', 'between'],
    },
  };
}

export const INITIAL_SAVED_SEARCHES: SavedSearch[] = [
  {
    id: 'saved-1',
    name: '我的高三二模筛选',
    query: '',
    conditions: [
      { id: 'saved-1-condition-1', field: 'grade', operator: 'eq', value: '高三' },
      { id: 'saved-1-condition-2', field: 'source', operator: 'contains', value: '二模' },
    ],
  },
  {
    id: 'saved-2',
    name: '三角函数困难题',
    query: '三角函数',
    conditions: [
      { id: 'saved-2-condition-1', field: 'tag', operator: 'eq', value: 't3' },
      { id: 'saved-2-condition-2', field: 'difficulty', operator: 'eq', value: 'hard' },
    ],
  },
  {
    id: 'saved-3',
    name: '概率统计压轴',
    query: '概率',
    conditions: [
      { id: 'saved-3-condition-1', field: 'tag', operator: 'eq', value: 't7' },
      { id: 'saved-3-condition-2', field: 'subjectiveScore', operator: 'gt', value: '7' },
    ],
  },
];

export const INITIAL_SEARCH_HISTORY = [
  '积分 \\int_0^1 x^2 dx',
  '高二 概率统计',
  'P-2024-0010',
  '三角函数 困难',
  '数学归纳法 证明',
];

export function searchProblems(
  problems: readonly Problem[],
  query: string,
  conditions: readonly SearchCondition[],
  tags: readonly Tag[]
): Problem[] {
  return problems.filter((problem) => {
    const matchesQuery =
      !query.trim() ||
      problem.code.toLowerCase().includes(query.trim().toLowerCase()) ||
      problem.latex.toLowerCase().includes(query.trim().toLowerCase());

    if (!matchesQuery) {
      return false;
    }

    return conditions.every((condition) => matchSearchCondition(problem, condition, tags));
  });
}

export function formatSearchConditionLabel(
  condition: SearchCondition,
  fieldConfig: Record<SearchField, SearchFieldConfig>
) {
  const config = fieldConfig[condition.field];
  const optionLabel = config.options?.find((item) => item.value === condition.value)?.label ?? condition.value;
  const secondOptionLabel =
    condition.secondValue
      ? config.options?.find((item) => item.value === condition.secondValue)?.label ?? condition.secondValue
      : undefined;

  return `${config.label} ${SEARCH_OPERATOR_LABELS[condition.operator]} ${
    condition.operator === 'between' ? `${optionLabel} - ${secondOptionLabel}` : optionLabel
  }`;
}

export function matchSearchCondition(
  problem: Problem,
  condition: SearchCondition,
  tags: readonly Tag[]
) {
  switch (condition.field) {
    case 'subject':
      return compareText(problem.subject ?? '', condition);
    case 'grade':
      return compareText(problem.grade ?? '', condition);
    case 'difficulty':
      return compareText(problem.difficulty, condition);
    case 'subjectiveScore':
      return compareNumber(problem.subjectiveScore ?? 0, condition);
    case 'type':
      return compareText(problem.type, condition);
    case 'hasImage':
      return condition.value === 'yes' ? problem.imageIds.length > 0 : problem.imageIds.length === 0;
    case 'tag': {
      const tagNames = tags
        .filter((tag) => problem.tagIds.includes(tag.id))
        .map((tag) => tag.name)
        .join(' ');
      const selectedTagName = tags.find((tag) => tag.id === condition.value)?.name ?? '';
      return compareText(tagNames, { ...condition, value: selectedTagName });
    }
    case 'source':
      return compareText(problem.source ?? '', condition);
    case 'date':
      return compareDate(problem.createdAt, condition);
    default:
      return true;
  }
}

function compareText(value: string, condition: SearchCondition) {
  if (condition.operator === 'contains') {
    return value.toLowerCase().includes(condition.value.toLowerCase());
  }

  return value === condition.value;
}

function compareNumber(value: number, condition: SearchCondition) {
  const target = Number(condition.value);
  const second = Number(condition.secondValue);

  if (condition.operator === 'gt') {
    return value > target;
  }
  if (condition.operator === 'lt') {
    return value < target;
  }
  if (condition.operator === 'between') {
    return value >= target && value <= second;
  }

  return value === target;
}

function compareDate(dateString: string, condition: SearchCondition) {
  const value = new Date(dateString).getTime();
  const target = new Date(condition.value).getTime();
  const second = condition.secondValue ? new Date(condition.secondValue).getTime() : target;

  if (condition.operator === 'gt') {
    return value > target;
  }
  if (condition.operator === 'lt') {
    return value < target;
  }
  if (condition.operator === 'between') {
    return value >= target && value <= second;
  }

  return value === target;
}

export function createSearchCondition(
  field: SearchField,
  operator: SearchOperator,
  value: string,
  secondValue?: string
): SearchCondition {
  return {
    id: `condition-${Date.now()}`,
    field,
    operator,
    value,
    secondValue: operator === 'between' ? secondValue : undefined,
  };
}
