import { z } from "zod"

export const problemSchema = z.object({
  latex: z.string().min(1, "题干不能为空").max(50000, "题干超过最大长度"),
  answerLatex: z.string().optional(),
  solutionLatex: z.string().optional(),
  type: z.enum(["multiple_choice", "fill_blank", "solve", "proof", "other"], {
    required_error: "请选择题型",
  }),
  difficulty: z.enum(["easy", "medium", "hard", "olympiad"], {
    required_error: "请选择难度",
  }),
  subjectiveScore: z.number().min(0).max(10).optional(),
  subject: z.string().optional(),
  grade: z.string().optional(),
  source: z.string().optional(),
  tagIds: z.array(z.string()).default([]),
  imageIds: z.array(z.string()).default([]),
  notes: z.string().optional(),
})

export type ProblemFormValues = z.infer<typeof problemSchema>
