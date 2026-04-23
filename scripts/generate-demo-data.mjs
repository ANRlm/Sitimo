import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { mockExportJobs, mockImages, mockProblems, mockTags } from '../lib/mock-data.ts'

const now = new Date()
const makeDate = (days = 0, hours = 0) =>
  new Date(now.getTime() - days * 24 * 60 * 60 * 1000 - hours * 60 * 60 * 1000).toISOString()

const papers = [
  {
    id: 'paper1',
    title: '2024 年高三数学期中考试',
    subtitle: '函数与导数综合卷',
    schoolName: '北京市第一中学',
    examName: '高三数学期中考试',
    subject: '数学',
    duration: 120,
    totalScore: 120,
    description: '覆盖函数、数列、立体几何与概率统计。',
    status: 'completed',
    instructions: '本卷满分 150 分，考试时间 120 分钟。请将答案写在答题纸指定区域内。',
    footerText: '— 第 1 页 共 4 页 —',
    createdAt: makeDate(14),
    updatedAt: makeDate(0, 1),
    layout: { columns: 1, fontSize: 12, lineHeight: 1.3, paperSize: 'A4', showAnswerVersion: true },
    items: [
      ['p1', 8, 'below'], ['p2', 10, 'below'], ['p3', 12, 'below'], ['p4', 12, 'below'],
      ['p5', 8, 'below'], ['p6', 12, 'right'], ['p7', 14, 'below'], ['p8', 12, 'below'],
      ['p9', 10, 'below'], ['p10', 8, 'below'], ['p11', 6, 'below'], ['p12', 8, 'below'],
    ].map(([problemId, score, imagePosition], orderIndex) => ({
      id: `item-${orderIndex + 1}-${problemId}`,
      problemId,
      score,
      orderIndex,
      imagePosition,
    })),
  },
  {
    id: 'paper2',
    title: '高二数学周测（第 3 周）',
    subtitle: '数列专题',
    schoolName: '北京市第一中学',
    examName: '阶段测试',
    subject: '数学',
    duration: 90,
    totalScore: 60,
    description: '用于高二数列和递推关系课堂检测。',
    status: 'completed',
    instructions: '请将答案写在答题纸指定区域内。',
    footerText: '— 第 1 页 共 2 页 —',
    createdAt: makeDate(10),
    updatedAt: makeDate(8),
    layout: { columns: 1, fontSize: 12, lineHeight: 1.3, paperSize: 'A4', showAnswerVersion: true },
    items: [
      ['p13', 10], ['p18', 8], ['p20', 12], ['p21', 12], ['p4', 10], ['p9', 8],
    ].map(([problemId, score], orderIndex) => ({
      id: `item-${orderIndex + 1}-${problemId}`,
      problemId,
      score,
      orderIndex,
      imagePosition: 'below',
    })),
  },
  {
    id: 'paper3',
    title: '函数专题练习',
    subtitle: '单元滚动训练',
    schoolName: '北京市第一中学',
    examName: '阶段测试',
    subject: '数学',
    duration: 75,
    totalScore: 48,
    description: '草稿版练习卷，强调函数图像与性质。',
    status: 'draft',
    instructions: '请独立完成。',
    footerText: '— 第 1 页 共 2 页 —',
    createdAt: makeDate(3),
    updatedAt: makeDate(1),
    layout: { columns: 1, fontSize: 12, lineHeight: 1.3, paperSize: 'A4', showAnswerVersion: true },
    items: [
      ['p4', 10], ['p8', 12], ['p14', 10], ['p16', 8], ['p19', 8],
    ].map(([problemId, score], orderIndex) => ({
      id: `item-${orderIndex + 1}-${problemId}`,
      problemId,
      score,
      orderIndex,
      imagePosition: 'below',
    })),
  },
  {
    id: 'paper4',
    title: '解析几何综合测试',
    subtitle: '圆锥曲线专项',
    schoolName: '北京市第一中学',
    examName: '阶段测试',
    subject: '数学',
    duration: 100,
    totalScore: 64,
    description: '待审核版本，包含 2 张图像题。',
    status: 'review',
    instructions: '请在规定时间内完成。',
    footerText: '— 第 1 页 共 3 页 —',
    createdAt: makeDate(7),
    updatedAt: makeDate(4),
    layout: { columns: 1, fontSize: 12, lineHeight: 1.3, paperSize: 'A4', showAnswerVersion: true },
    items: [
      ['p6', 12, 'right'], ['p17', 12, 'below'], ['p15', 8, 'below'], ['p22', 14, 'below'], ['p24', 18, 'below'],
    ].map(([problemId, score, imagePosition], orderIndex) => ({
      id: `item-${orderIndex + 1}-${problemId}`,
      problemId,
      score,
      orderIndex,
      imagePosition,
    })),
  },
  {
    id: 'paper5',
    title: '三角函数单元测试',
    subtitle: '基础巩固',
    schoolName: '北京市第一中学',
    examName: '阶段测试',
    subject: '数学',
    duration: 60,
    totalScore: 40,
    description: '三角函数基础练习。',
    status: 'draft',
    instructions: '请在 60 分钟内完成。',
    footerText: '— 第 1 页 共 1 页 —',
    createdAt: makeDate(1),
    updatedAt: makeDate(0, 12),
    layout: { columns: 1, fontSize: 12, lineHeight: 1.3, paperSize: 'A4', showAnswerVersion: true },
    items: [
      ['p2', 10], ['p14', 10], ['p22', 12], ['p9', 8],
    ].map(([problemId, score], orderIndex) => ({
      id: `item-${orderIndex + 1}-${problemId}`,
      problemId,
      score,
      orderIndex,
      imagePosition: 'below',
    })),
  },
]

const payload = {
  tags: mockTags.map(({ problemCount, ...tag }) => tag),
  problems: mockProblems,
  images: mockImages.map((image) => ({
    ...image,
    updatedAt: image.createdAt,
  })),
  papers,
  exportJobs: mockExportJobs.map((job) => ({
    ...job,
    progress: job.progress ?? (job.status === 'done' ? 100 : job.status === 'failed' ? 100 : 0),
  })),
  settings: {
    preferences: {
      defaultView: 'grid',
      defaultSort: 'updatedAt-desc',
      showAnswerVersion: true,
    },
    export: {
      columns: 1,
      lineHeight: 1.3,
      paperSize: 'A4',
    },
    separators: {
      problemStart: '\\begin{problem}',
      problemEnd: '\\end{problem}',
    },
  },
}

const rootDir = path.dirname(fileURLToPath(import.meta.url))
const target = path.resolve(rootDir, '../server/testdata/demo-data.json')
await fs.mkdir(path.dirname(target), { recursive: true })
await fs.writeFile(target, JSON.stringify(payload, null, 2) + '\n')
console.log(`wrote ${target}`)
