import { z } from 'zod';

export const problemImportSchema = z.object({
  latexSource: z.string().min(1, '请粘贴或上传 LaTeX 源码'),
  separatorStart: z.string().min(1, '请输入开始标记'),
  separatorEnd: z.string().min(1, '请输入结束标记'),
  subject: z.string().min(1, '请输入学科'),
  grade: z.string().min(1, '请输入年级'),
  source: z.string().optional(),
  difficulty: z.enum(['easy', 'medium', 'hard', 'olympiad'], {
    required_error: '请选择默认难度',
  }),
  tagNames: z.string().optional(),
});

export type ProblemImportFormValues = z.infer<typeof problemImportSchema>;
