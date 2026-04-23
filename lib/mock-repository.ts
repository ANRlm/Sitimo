import { mockExportJobs, mockImages, mockProblems, mockTags } from './mock-data';
import { getPaperRecord, mockPaperRecords } from './paper-data';
import type {
  ExportHistoryRow,
  PaperSummary,
} from './frontend-contracts';
import type { ImageAsset, Problem, Tag } from './types';

export function listProblems(): Problem[] {
  return mockProblems;
}

export function getProblemById(id: string): Problem | undefined {
  return mockProblems.find((problem) => problem.id === id);
}

export function listTags(): Tag[] {
  return mockTags;
}

export function getTagById(id: string): Tag | undefined {
  return mockTags.find((tag) => tag.id === id);
}

export function listImages(): ImageAsset[] {
  return mockImages;
}

export function getImageById(id: string): ImageAsset | undefined {
  return mockImages.find((image) => image.id === id);
}

export function listProblemTags(problem: Pick<Problem, 'tagIds'>): Tag[] {
  return mockTags.filter((tag) => problem.tagIds.includes(tag.id));
}

export function listProblemImages(problem: Pick<Problem, 'imageIds'>): ImageAsset[] {
  return mockImages.filter((image) => problem.imageIds.includes(image.id));
}

export function getProblemPrimaryImage(problem: Pick<Problem, 'imageIds'>): ImageAsset | undefined {
  return mockImages.find((image) => problem.imageIds.includes(image.id));
}

export function listProblemsByTagId(tagId: string, limit?: number): Problem[] {
  const problems = mockProblems.filter((problem) => problem.tagIds.includes(tagId));

  if (typeof limit === 'number') {
    return problems.slice(0, limit);
  }

  return problems;
}

export function listLinkedProblemsForImage(
  image: Pick<ImageAsset, 'linkedProblemIds'>
): Problem[] {
  return mockProblems.filter((problem) => image.linkedProblemIds.includes(problem.id));
}

export function listPaperRecords() {
  return mockPaperRecords;
}

export function listPaperSummaries(): PaperSummary[] {
  return mockPaperRecords.map((paper) => ({
    id: paper.id,
    title: paper.title,
    subtitle: paper.subtitle,
    description: paper.description,
    status: paper.status,
    problemCount: paper.items.length,
    totalScore: paper.totalScore ?? paper.items.reduce((sum, item) => sum + item.score, 0),
    createdAt: paper.createdAt,
    updatedAt: paper.updatedAt,
  }));
}

export function getPaperSummaryById(id: string): PaperSummary | undefined {
  const paper = getPaperRecord(id);

  if (!paper) {
    return undefined;
  }

  return {
    id: paper.id,
    title: paper.title,
    subtitle: paper.subtitle,
    description: paper.description,
    status: paper.status,
    problemCount: paper.items.length,
    totalScore: paper.totalScore ?? paper.items.reduce((sum, item) => sum + item.score, 0),
    createdAt: paper.createdAt,
    updatedAt: paper.updatedAt,
  };
}

export function listExportHistoryRows(): ExportHistoryRow[] {
  const paperSummaries = listPaperSummaries();
  const paperSummaryMap = new Map(paperSummaries.map((paper) => [paper.id, paper]));

  const seededRows = mockExportJobs.map<ExportHistoryRow>((job) => {
    const summary = paperSummaryMap.get(job.paperId);

    return {
      ...job,
      itemCount: summary?.problemCount ?? 12,
      elapsed: formatMockElapsed(job.createdAt, job.completedAt),
    };
  });

  const syntheticRows = Array.from({ length: 8 }, (_, index) => {
    const summary = paperSummaries[index % paperSummaries.length];
    const createdAt = new Date(Date.now() - (index + 4) * 1000 * 60 * 60).toISOString();
    const completedAt =
      index === 1 || index === 4 || index === 6
        ? undefined
        : new Date(Date.now() - (index + 4) * 1000 * 60 * 40).toISOString();
    const format: ExportHistoryRow['format'] = index % 2 === 0 ? 'pdf' : 'latex';
    const variant: ExportHistoryRow['variant'] =
      index % 3 === 0 ? 'both' : index % 3 === 1 ? 'student' : 'answer';
    const status: ExportHistoryRow['status'] =
      index === 1 ? 'processing' : index === 4 ? 'failed' : index === 6 ? 'pending' : 'done';

    return {
      id: `history-${index + 1}`,
      paperId: summary?.id ?? `paper-${index + 10}`,
      paperTitle: `阶段测试导出任务 ${index + 1}`,
      format,
      variant,
      status,
      progress: index === 1 ? 58 : undefined,
      errorMessage: index === 4 ? 'LaTeX 编译失败：第 42 行缺少 \\end{align}' : undefined,
      downloadUrl: '#',
      createdAt,
      completedAt,
      itemCount: summary?.problemCount ?? 10 + index,
      elapsed: formatMockElapsed(createdAt, completedAt),
    };
  });

  return [...seededRows, ...syntheticRows];
}

function formatMockElapsed(createdAt: string, completedAt?: string) {
  const start = new Date(createdAt).getTime();
  const end = completedAt ? new Date(completedAt).getTime() : Date.now();
  const seconds = Math.max(1, Math.round((end - start) / 1000));

  return `${seconds}s`;
}
