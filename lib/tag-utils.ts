import type { TagCategoryFilter } from './frontend-contracts';
import type { Problem, Tag } from './types';

export function filterTags(
  tags: readonly Tag[],
  searchQuery: string,
  categoryFilter: TagCategoryFilter
) {
  return tags.filter((tag) => {
    if (searchQuery && !tag.name.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }

    if (categoryFilter !== 'all' && tag.category !== categoryFilter) {
      return false;
    }

    return true;
  });
}

export function listTagUsageProblems(
  problems: readonly Problem[],
  tagId: string,
  limit = 10
) {
  return problems.filter((problem) => problem.tagIds.includes(tagId)).slice(0, limit);
}
