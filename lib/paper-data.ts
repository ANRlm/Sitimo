import { mockImages, mockProblems } from '@/lib/mock-data';
import type { PaperStatus } from '@/lib/frontend-contracts';
import type { ImageAsset, Paper, PaperItem, Problem } from '@/lib/types';

export type PaperRecord = Paper & {
  description?: string;
  subject: string;
  status: PaperStatus;
  instructions: string;
  footerText: string;
};

type PaperItemSeed = {
  problemId: string;
  score: number;
  imagePosition?: PaperItem['imagePosition'];
};

function createItems(seed: PaperItemSeed[]): PaperItem[] {
  return seed.map((item, index) => ({
    id: `item-${index + 1}-${item.problemId}`,
    orderIndex: index,
    ...item,
  }));
}

function createPaperRecord(
  id: string,
  title: string,
  subtitle: string,
  status: PaperStatus,
  itemSeed: PaperItemSeed[],
  overrides?: Partial<Omit<PaperRecord, 'id' | 'title' | 'subtitle' | 'items' | 'status'>>
): PaperRecord {
  const now = new Date();
  const createdAt = new Date(now.getTime() - 1000 * 60 * 60 * 24 * 14).toISOString();
  const updatedAt = new Date(now.getTime() - 1000 * 60 * 45).toISOString();
  const items = createItems(itemSeed);

  return {
    id,
    title,
    subtitle,
    schoolName: overrides?.schoolName ?? '北京市第一中学',
    examName: overrides?.examName ?? '阶段测试',
    duration: overrides?.duration ?? 120,
    totalScore:
      overrides?.totalScore ?? items.reduce((sum, item) => sum + item.score, 0),
    items,
    layout: overrides?.layout ?? {
      columns: 1,
      fontSize: 12,
      lineHeight: 1.3,
      paperSize: 'A4',
      showAnswerVersion: true,
    },
    createdAt,
    updatedAt,
    description:
      overrides?.description ?? '用于课堂测验与阶段复习的标准化试卷模板。',
    subject: overrides?.subject ?? '数学',
    status,
    instructions:
      overrides?.instructions ??
      '本卷满分 150 分，考试时间 120 分钟。请将答案写在答题纸指定区域内。',
    footerText:
      overrides?.footerText ??
      '— 第 1 页 共 4 页 —',
  };
}

export const mockPaperRecords: PaperRecord[] = [
  createPaperRecord(
    'paper1',
    '2024 年高三数学期中考试',
    '函数与导数综合卷',
    'completed',
    [
      { problemId: 'p1', score: 8 },
      { problemId: 'p2', score: 10 },
      { problemId: 'p3', score: 12, imagePosition: 'below' },
      { problemId: 'p4', score: 12 },
      { problemId: 'p5', score: 8 },
      { problemId: 'p6', score: 12, imagePosition: 'right' },
      { problemId: 'p7', score: 14, imagePosition: 'below' },
      { problemId: 'p8', score: 12 },
      { problemId: 'p9', score: 10 },
      { problemId: 'p10', score: 8 },
      { problemId: 'p11', score: 6 },
      { problemId: 'p12', score: 8 },
    ],
    {
      examName: '高三数学期中考试',
      description: '覆盖函数、数列、立体几何与概率统计。',
    }
  ),
  createPaperRecord(
    'paper2',
    '高二数学周测（第 3 周）',
    '数列专题',
    'completed',
    [
      { problemId: 'p13', score: 10 },
      { problemId: 'p18', score: 8 },
      { problemId: 'p20', score: 12 },
      { problemId: 'p21', score: 12 },
      { problemId: 'p4', score: 10 },
      { problemId: 'p9', score: 8 },
    ],
    {
      duration: 90,
      totalScore: 60,
      description: '用于高二数列和递推关系课堂检测。',
    }
  ),
  createPaperRecord(
    'paper3',
    '函数专题练习',
    '单元滚动训练',
    'draft',
    [
      { problemId: 'p4', score: 10 },
      { problemId: 'p8', score: 12 },
      { problemId: 'p14', score: 10 },
      { problemId: 'p16', score: 8 },
      { problemId: 'p19', score: 8 },
    ],
    {
      duration: 75,
      totalScore: 48,
      description: '草稿版练习卷，强调函数图像与性质。',
    }
  ),
  createPaperRecord(
    'paper4',
    '解析几何综合测试',
    '圆锥曲线专项',
    'review',
    [
      { problemId: 'p6', score: 12, imagePosition: 'right' },
      { problemId: 'p17', score: 12 },
      { problemId: 'p15', score: 8 },
      { problemId: 'p22', score: 14 },
      { problemId: 'p24', score: 18 },
    ],
    {
      duration: 100,
      totalScore: 64,
      description: '待审核版本，包含 2 张图像题。',
    }
  ),
  createPaperRecord(
    'paper5',
    '三角函数单元测试',
    '基础巩固',
    'draft',
    [
      { problemId: 'p2', score: 10 },
      { problemId: 'p14', score: 10 },
      { problemId: 'p22', score: 12 },
      { problemId: 'p9', score: 8 },
    ],
    {
      duration: 60,
      totalScore: 40,
    }
  ),
  createPaperRecord(
    'paper6',
    '高一数学月考',
    '集合与函数基础',
    'completed',
    [
      { problemId: 'p10', score: 8 },
      { problemId: 'p14', score: 10 },
      { problemId: 'p15', score: 10 },
      { problemId: 'p16', score: 10 },
      { problemId: 'p19', score: 8 },
      { problemId: 'p21', score: 14 },
    ],
    {
      duration: 100,
      totalScore: 60,
      schoolName: '上海市第二中学',
    }
  ),
  createPaperRecord(
    'paper7',
    '概率统计专项训练',
    '选填压轴混编',
    'review',
    [
      { problemId: 'p5', score: 10 },
      { problemId: 'p12', score: 8 },
      { problemId: 'p18', score: 8 },
      { problemId: 'p20', score: 12 },
      { problemId: 'p24', score: 18 },
    ],
    {
      duration: 90,
      totalScore: 56,
    }
  ),
  createPaperRecord(
    'paper8',
    '竞赛压轴讲评卷',
    '不等式与数论',
    'draft',
    [
      { problemId: 'p23', score: 20 },
      { problemId: 'p24', score: 20 },
      { problemId: 'p3', score: 18 },
    ],
    {
      duration: 120,
      totalScore: 58,
      description: '面向拔高训练的讲评版本。',
    }
  ),
];

export function getPaperRecord(id: string): PaperRecord | undefined {
  return mockPaperRecords.find((paper) => paper.id === id);
}

export function getPaperItemProblem(
  problemId: string
): Problem | undefined {
  return mockProblems.find((problem) => problem.id === problemId);
}

export function getPaperRenderItems(paper: PaperRecord): Array<{
  item: PaperItem;
  problem: Problem;
  images: ImageAsset[];
}> {
  return paper.items
    .map((item) => {
      const problem = getPaperItemProblem(item.problemId);

      if (!problem) {
        return null;
      }

      const images = mockImages.filter((image) =>
        problem.imageIds.includes(image.id)
      );

      return { item, problem, images };
    })
    .filter((item): item is { item: PaperItem; problem: Problem; images: ImageAsset[] } =>
      Boolean(item)
    );
}

export function createNewPaperId(): string {
  return `paper-${Date.now()}`;
}
