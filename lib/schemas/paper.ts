import { z } from 'zod';

export const paperSchema = z.object({
  title: z.string().trim().min(1, '试卷标题不能为空').max(200),
  subtitle: z.string().optional(),
  schoolName: z.string().optional(),
  examName: z.string().optional(),
  subject: z.string().optional(),
  duration: z.string().optional().refine((value) => !value || (/^\d+$/.test(value) && Number(value) > 0), {
    message: '时长必须是正整数',
  }),
  description: z.string().optional(),
  status: z.enum(['draft', 'completed', 'review']).default('draft'),
  instructions: z.string().optional(),
  footerText: z.string().optional(),
  columns: z.enum(['1', '2']).default('1'),
  fontSize: z.string().optional().refine((value) => !value || Number(value) > 0, {
    message: '字号必须大于 0',
  }),
  lineHeight: z.string().optional().refine((value) => !value || Number(value) > 0, {
    message: '行距必须大于 0',
  }),
  paperSize: z.enum(['A4', 'B5', 'Letter']).default('A4'),
  showAnswerVersion: z.boolean().default(true),
});

export type PaperFormValues = z.infer<typeof paperSchema>;
