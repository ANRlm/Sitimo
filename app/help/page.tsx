"use client"

import { useState } from "react"
import {
  Search,
  BookOpen,
  MessageCircle,
  Mail,
  ExternalLink,
  ChevronRight,
  Download,
  FileText,
  Image as ImageIcon,
  Tag,
  HelpCircle
} from "lucide-react"
import { PageHeader, PagePanel, PageShell } from "@/components/page-shell"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { Kbd } from "@/components/ui/kbd"

const quickStartGuides = [
  {
    icon: FileText,
    title: "创建第一道题目",
    description: "学习如何创建和编辑数学题目",
    href: "#create-problem",
  },
  {
    icon: ImageIcon,
    title: "管理图片资源",
    description: "上传和管理题目中使用的图片",
    href: "#manage-images",
  },
  {
    icon: Download,
    title: "导出试卷",
    description: "在编辑器中生成 PDF 或 LaTeX 导出任务",
    href: "#create-paper",
  },
  {
    icon: Tag,
    title: "使用标签分类",
    description: "为题目添加标签便于检索",
    href: "#use-tags",
  },
]

const faqs = [
  {
    question: "如何在题目中插入数学公式？",
    answer: "系统支持 LaTeX 语法编写数学公式。在编辑器中使用 $ 符号包裹行内公式，使用 $$ 符号包裹独立公式。例如：$x^2 + y^2 = z^2$ 会显示为一个勾股定理公式。",
  },
  {
    question: "如何批量导入题目？",
    answer: "目前支持从 Word、Excel 和 JSON 格式批量导入题目。在题目列表页点击「导入」按钮，选择文件后系统会自动解析并导入。建议先下载导入模板确保格式正确。",
  },
  {
    question: "题篮有什么作用？",
    answer: "题篮类似购物车，您可以在浏览题库时将需要的题目添加到题篮。之后在出卷中心可以快速将题篮中的题目添加到试卷中，提高组卷效率。",
  },
  {
    question: "如何设置题目难度？",
    answer: "在编辑题目时，可以选择难度等级：简单、中等、困难。系统会根据难度分布统计帮助您创建难度均衡的试卷。",
  },
  {
    question: "删除的题目可以恢复吗？",
    answer: "可以。删除的题目会先移动到回收站，保留30天。在这期间您可以随时恢复。30天后会自动永久删除。",
  },
  {
    question: "如何导出试卷为PDF？",
    answer: "在试卷编辑页面，或试卷只读预览页顶部，点击「导出 PDF」按钮即可加入导出队列；如需 Overleaf 可直接导入的 LaTeX 压缩包，可点击「导出 LaTeX 包」。生成后的文件会出现在「导出历史」页面。",
  },
  {
    question: "支持哪些图片格式？",
    answer: "系统支持 JPG、PNG、GIF、SVG 等常见图片格式。单张图片大小限制为 5MB。建议使用清晰度适中的图片，以兼顾预览速度和导出清晰度。",
  },
  {
    question: "如何共享题目给其他老师？",
    answer: "目前可以通过导出题目为 JSON 文件，然后分享给其他老师导入。未来版本将支持团队协作和在线共享功能。",
  },
]

const keyboardShortcuts = [
  { keys: ["Ctrl", "S"], description: "保存当前编辑" },
  { keys: ["Ctrl", "N"], description: "新建题目" },
  { keys: ["Ctrl", "F"], description: "搜索题目" },
  { keys: ["Ctrl", "B"], description: "添加到题篮" },
  { keys: ["Ctrl", "/"], description: "显示快捷键帮助" },
  { keys: ["Esc"], description: "关闭弹窗/取消操作" },
]

export default function HelpPage() {
  const [searchQuery, setSearchQuery] = useState("")

  const filteredFaqs = faqs.filter(faq =>
    faq.question.toLowerCase().includes(searchQuery.toLowerCase()) ||
    faq.answer.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <PageShell>
      <PageHeader
        eyebrow="帮助中心"
        title="查找指引与常见问题"
        description="查找使用指南、常见问题解答和联系支持。"
      >
        <div className="relative max-w-3xl">
          <Search className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索问题..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-12 rounded-2xl pl-11 text-base"
          />
        </div>
      </PageHeader>

      <section className="space-y-4">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-xl font-semibold text-foreground">快速入门</h2>
          <span className="text-sm text-muted-foreground">从高频操作开始</span>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {quickStartGuides.map((guide) => (
            <Card
              key={guide.title}
              className="group cursor-pointer border-border bg-card transition-all hover:border-primary/50 hover:shadow-md"
            >
              <CardContent className="flex items-center gap-4 p-4">
                <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10 transition-colors group-hover:bg-primary/20">
                  <guide.icon className="h-6 w-6 text-primary" />
                </div>
                <div className="min-w-0 flex-1">
                  <h3 className="font-medium text-foreground transition-colors group-hover:text-primary">
                    {guide.title}
                  </h3>
                  <p className="text-sm text-muted-foreground">{guide.description}</p>
                </div>
                <ChevronRight className="h-5 w-5 text-muted-foreground transition-colors group-hover:text-primary" />
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-xl font-semibold text-foreground">常见问题</h2>
        <PagePanel>
          {filteredFaqs.length === 0 ? (
            <Empty className="min-h-[260px] border-none bg-transparent">
              <EmptyMedia variant="icon">
                <HelpCircle className="h-6 w-6" />
              </EmptyMedia>
              <EmptyHeader>
                <EmptyTitle>没有找到相关问题</EmptyTitle>
                <EmptyDescription>试试更短的关键词，或者改搜操作名称。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <Accordion type="single" collapsible className="w-full">
              {filteredFaqs.map((faq, index) => (
                <AccordionItem key={index} value={`item-${index}`} className="px-4">
                  <AccordionTrigger className="py-4 text-left hover:no-underline">
                    <span className="font-medium text-foreground">{faq.question}</span>
                  </AccordionTrigger>
                  <AccordionContent className="pb-4 text-muted-foreground">
                    {faq.answer}
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          )}
        </PagePanel>
      </section>

      <section className="space-y-4">
        <h2 className="text-xl font-semibold text-foreground">快捷键</h2>
        <PagePanel>
          <div className="p-6">
            <div className="grid gap-3 lg:grid-cols-2">
              {keyboardShortcuts.map((shortcut, index) => (
                <div key={index} className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-border/70 px-4 py-3">
                  <span className="text-sm text-muted-foreground">{shortcut.description}</span>
                  <div className="flex items-center gap-1">
                    {shortcut.keys.map((key, keyIndex) => (
                      <span key={keyIndex}>
                        <Kbd>{key}</Kbd>
                        {keyIndex < shortcut.keys.length - 1 && (
                          <span className="mx-1 text-muted-foreground">+</span>
                        )}
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </PagePanel>
      </section>

      <section className="space-y-4">
        <h2 className="text-xl font-semibold text-foreground">联系我们</h2>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <Card className="bg-card border-border">
            <CardHeader>
              <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <BookOpen className="h-6 w-6 text-primary" />
              </div>
              <CardTitle className="text-base">文档中心</CardTitle>
              <CardDescription>查看详细的使用文档</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full">
                查看文档
                <ExternalLink className="ml-2 h-4 w-4" />
              </Button>
            </CardContent>
          </Card>

          <Card className="bg-card border-border">
            <CardHeader>
              <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-emerald-500/10">
                <MessageCircle className="h-6 w-6 text-emerald-500" />
              </div>
              <CardTitle className="text-base">在线客服</CardTitle>
              <CardDescription>工作日 9:00-18:00</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full">
                开始对话
              </Button>
            </CardContent>
          </Card>

          <Card className="bg-card border-border">
            <CardHeader>
              <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-amber-500/10">
                <Mail className="h-6 w-6 text-amber-500" />
              </div>
              <CardTitle className="text-base">邮件支持</CardTitle>
              <CardDescription>24小时内回复</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full">
                发送邮件
              </Button>
            </CardContent>
          </Card>
        </div>
      </section>
    </PageShell>
  )
}
