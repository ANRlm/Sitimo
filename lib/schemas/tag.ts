import { z } from "zod"

export const tagSchema = z.object({
  name: z.string().min(1, "标签名不能为空").max(100),
  category: z.enum(["topic", "source", "custom"], {
    required_error: "请选择分类",
  }),
  color: z.string().min(1, "请选择颜色"),
  description: z.string().optional(),
})

export type TagFormValues = z.infer<typeof tagSchema>
