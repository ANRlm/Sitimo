import type { Difficulty, ExportJob, ProblemType, Tag } from './types';

export type PaperStatus = 'draft' | 'completed' | 'review';

export type PaperSummary = {
  id: string;
  title: string;
  subtitle?: string;
  description?: string;
  status: PaperStatus;
  problemCount: number;
  totalScore: number;
  createdAt: string;
  updatedAt: string;
};

export type ProblemListViewMode = 'grid' | 'list' | 'compact';
export type ProblemListSort =
  | 'updatedAt-desc'
  | 'createdAt-desc'
  | 'code-asc'
  | 'subjectiveScore-asc';
export type ProblemListTypeFilter = ProblemType | 'all';
export type ProblemListHasImageFilter = 'all' | 'yes' | 'no';

export type ProblemListFilters = {
  search: string;
  subject: string;
  grade: string;
  difficulties: Difficulty[];
  type: ProblemListTypeFilter;
  scoreRange: [number, number];
  hasImage: ProblemListHasImageFilter;
  tagIds: string[];
  startDate: string | null;
  endDate: string | null;
};

export type ProblemListActiveFilter = {
  key: string;
  label: string;
};

export type SearchField =
  | 'subject'
  | 'grade'
  | 'difficulty'
  | 'subjectiveScore'
  | 'type'
  | 'hasImage'
  | 'tag'
  | 'source'
  | 'date';

export type SearchOperator = 'eq' | 'contains' | 'gt' | 'lt' | 'between';
export type SearchFieldInputType = 'select' | 'text' | 'number' | 'date';

export type SearchCondition = {
  id: string;
  field: SearchField;
  operator: SearchOperator;
  value: string;
  secondValue?: string;
};

export type SavedSearch = {
  id: string;
  name: string;
  query: string;
  conditions: SearchCondition[];
};

export type SearchFieldConfig = {
  label: string;
  type: SearchFieldInputType;
  operators: SearchOperator[];
  options?: Array<{ label: string; value: string }>;
};

export type TagCategoryFilter = 'all' | Tag['category'];

export type ImportStep = 1 | 2 | 3;

export type ParsedProblemDraft = {
  id: string;
  title: string;
  latex: string;
  difficulty: Difficulty;
  status: 'success' | 'error';
  error?: string;
  warnings?: string[];
  subject?: string;
  grade?: string;
  source?: string;
  tagNames: string[];
};

export type ImportBatchMetadata = {
  subject: string;
  grade: string;
  source: string;
  difficulty: Difficulty;
  tagNames: string[];
};

export type NormalizedParsedProblemDraft = ParsedProblemDraft & {
  code: string;
  subject: string;
  grade: string;
  source: string;
};

export type ExportHistoryRow = ExportJob & {
  itemCount: number;
  elapsed: string;
};
