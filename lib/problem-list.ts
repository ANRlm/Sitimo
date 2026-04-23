import type { ReadonlyURLSearchParams } from 'next/navigation';
import { PROBLEM_TYPE_FILTER_OPTIONS } from './catalogs';
import type {
  ProblemListActiveFilter,
  ProblemListFilters,
  ProblemListSort,
} from './frontend-contracts';
import { difficultyConfig, type Difficulty, type Problem } from './types';
import type { Tag } from './types';

export const DEFAULT_PROBLEM_LIST_FILTERS: ProblemListFilters = {
  search: '',
  subject: '全部',
  grade: '全部',
  difficulties: [],
  type: 'all',
  scoreRange: [1, 10],
  hasImage: 'all',
  tagIds: [],
  startDate: null,
  endDate: null,
};

export function parseProblemListFilters(
  searchParams: ReadonlyURLSearchParams | URLSearchParams
): ProblemListFilters {
  const nextFilters: ProblemListFilters = {
    ...DEFAULT_PROBLEM_LIST_FILTERS,
  };

  const search = searchParams.get('search');
  const subject = searchParams.get('subject');
  const grade = searchParams.get('grade');
  const type = searchParams.get('type');
  const hasImage = searchParams.get('hasImage');
  const startDate = searchParams.get('startDate');
  const endDate = searchParams.get('endDate');
  const tagId = searchParams.get('tag');
  const tagIds = searchParams.get('tags');
  const scoreMin = searchParams.get('scoreMin');
  const scoreMax = searchParams.get('scoreMax');
  const difficulties = searchParams.get('difficulty');

  if (search) {
    nextFilters.search = search;
  }

  if (subject) {
    nextFilters.subject = subject;
  }

  if (grade) {
    nextFilters.grade = grade;
  }

  if (type === 'multiple_choice' || type === 'fill_blank' || type === 'solve' || type === 'proof') {
    nextFilters.type = type;
  }

  if (hasImage === 'yes' || hasImage === 'no') {
    nextFilters.hasImage = hasImage;
  }

  if (startDate) {
    nextFilters.startDate = startDate;
  }

  if (endDate) {
    nextFilters.endDate = endDate;
  }

  if (tagId) {
    nextFilters.tagIds = [tagId];
  } else if (tagIds) {
    nextFilters.tagIds = tagIds
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean);
  }

  if (difficulties) {
    nextFilters.difficulties = difficulties
      .split(',')
      .filter(isDifficulty);
  }

  const parsedMin = scoreMin ? Number(scoreMin) : 1;
  const parsedMax = scoreMax ? Number(scoreMax) : 10;
  if (Number.isFinite(parsedMin) && Number.isFinite(parsedMax)) {
    nextFilters.scoreRange = [
      Math.max(1, Math.min(parsedMin, 10)),
      Math.max(1, Math.min(parsedMax, 10)),
    ];
  }

  return nextFilters;
}

export function queryProblems(
  problems: readonly Problem[],
  filters: ProblemListFilters,
  sortBy: ProblemListSort
): Problem[] {
  return sortProblems(filterProblems(problems, filters), sortBy);
}

export function filterProblems(
  problems: readonly Problem[],
  filters: ProblemListFilters
): Problem[] {
  return problems.filter((problem) => {
    if (filters.search.trim()) {
      const query = filters.search.trim().toLowerCase();
      const matchesSearch =
        problem.code.toLowerCase().includes(query) ||
        problem.latex.toLowerCase().includes(query);

      if (!matchesSearch) {
        return false;
      }
    }

    if (filters.subject !== '全部' && problem.subject !== filters.subject) {
      return false;
    }

    if (filters.grade !== '全部' && problem.grade !== filters.grade) {
      return false;
    }

    if (filters.difficulties.length > 0 && !filters.difficulties.includes(problem.difficulty)) {
      return false;
    }

    if (filters.type !== 'all' && problem.type !== filters.type) {
      return false;
    }

    if (
      problem.subjectiveScore &&
      (problem.subjectiveScore < filters.scoreRange[0] ||
        problem.subjectiveScore > filters.scoreRange[1])
    ) {
      return false;
    }

    if (filters.hasImage === 'yes' && problem.imageIds.length === 0) {
      return false;
    }

    if (filters.hasImage === 'no' && problem.imageIds.length > 0) {
      return false;
    }

    if (filters.tagIds.length > 0 && !filters.tagIds.every((tagId) => problem.tagIds.includes(tagId))) {
      return false;
    }

    if (filters.startDate) {
      const startTime = new Date(`${filters.startDate}T00:00:00`).getTime();
      if (new Date(problem.createdAt).getTime() < startTime) {
        return false;
      }
    }

    if (filters.endDate) {
      const endTime = new Date(`${filters.endDate}T23:59:59`).getTime();
      if (new Date(problem.createdAt).getTime() > endTime) {
        return false;
      }
    }

    return true;
  });
}

export function sortProblems(
  problems: readonly Problem[],
  sortBy: ProblemListSort
): Problem[] {
  const [sortField, sortOrder] = sortBy.split('-');

  return [...problems].sort((left, right) => {
    let comparison = 0;

    if (sortField === 'updatedAt' || sortField === 'createdAt') {
      comparison = new Date(right[sortField]).getTime() - new Date(left[sortField]).getTime();
    }

    if (sortField === 'code') {
      comparison = left.code.localeCompare(right.code);
    }

    if (sortField === 'subjectiveScore') {
      comparison = (left.subjectiveScore ?? 0) - (right.subjectiveScore ?? 0);
    }

    return sortOrder === 'desc' ? comparison : -comparison;
  });
}

export function buildProblemActiveFilters(
  filters: ProblemListFilters,
  tags: readonly Tag[]
): ProblemListActiveFilter[] {
  const activeFilters: ProblemListActiveFilter[] = [];

  if (filters.search) {
    activeFilters.push({ key: 'search', label: `关键词: ${filters.search}` });
  }

  if (filters.subject !== '全部') {
    activeFilters.push({ key: 'subject', label: `学科: ${filters.subject}` });
  }

  if (filters.grade !== '全部') {
    activeFilters.push({ key: 'grade', label: `年级: ${filters.grade}` });
  }

  filters.difficulties.forEach((difficulty) => {
    activeFilters.push({
      key: `difficulty-${difficulty}`,
      label: `难度: ${difficultyConfig[difficulty].label}`,
    });
  });

  if (filters.type !== 'all') {
    const label =
      PROBLEM_TYPE_FILTER_OPTIONS.find((option) => option.value === filters.type)?.label ?? '全部';
    activeFilters.push({ key: 'type', label: `题型: ${label}` });
  }

  if (filters.scoreRange[0] !== 1 || filters.scoreRange[1] !== 10) {
    activeFilters.push({
      key: 'scoreRange',
      label: `主观难度: ${filters.scoreRange[0]}-${filters.scoreRange[1]}`,
    });
  }

  if (filters.hasImage !== 'all') {
    activeFilters.push({
      key: 'hasImage',
      label: `图像: ${filters.hasImage === 'yes' ? '有图' : '无图'}`,
    });
  }

  filters.tagIds.forEach((tagId) => {
    const tag = tags.find((item) => item.id === tagId);
    if (tag) {
      activeFilters.push({ key: `tag-${tag.id}`, label: `标签: ${tag.name}` });
    }
  });

  if (filters.startDate || filters.endDate) {
    activeFilters.push({
      key: 'createdAt',
      label: `创建日期: ${filters.startDate ?? '不限'} ~ ${filters.endDate ?? '不限'}`,
    });
  }

  return activeFilters;
}

export function clearProblemFilter(filters: ProblemListFilters, key: string): ProblemListFilters {
  if (key === 'search') {
    return { ...filters, search: '' };
  }
  if (key === 'subject') {
    return { ...filters, subject: '全部' };
  }
  if (key === 'grade') {
    return { ...filters, grade: '全部' };
  }
  if (key === 'type') {
    return { ...filters, type: 'all' };
  }
  if (key === 'scoreRange') {
    return { ...filters, scoreRange: [1, 10] };
  }
  if (key === 'hasImage') {
    return { ...filters, hasImage: 'all' };
  }
  if (key === 'createdAt') {
    return { ...filters, startDate: null, endDate: null };
  }
  if (key.startsWith('difficulty-')) {
    const difficulty = key.replace('difficulty-', '') as Difficulty;
    return {
      ...filters,
      difficulties: filters.difficulties.filter((item) => item !== difficulty),
    };
  }
  if (key.startsWith('tag-')) {
    const tagId = key.replace('tag-', '');
    return {
      ...filters,
      tagIds: filters.tagIds.filter((item) => item !== tagId),
    };
  }

  return filters;
}

export function toggleDifficultyFilter(
  currentValues: readonly Difficulty[],
  difficulty: Difficulty
): Difficulty[] {
  return currentValues.includes(difficulty)
    ? currentValues.filter((item) => item !== difficulty)
    : [...currentValues, difficulty];
}

function isDifficulty(value: string): value is Difficulty {
  return value === 'easy' || value === 'medium' || value === 'hard' || value === 'olympiad';
}
