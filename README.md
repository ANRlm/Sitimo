# Sitimo

> 面向中文教师的数学题库管理系统，支持 LaTeX 公式编辑、试卷组排与 PDF 导出。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-compose-blue)](docker-compose.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8)](server/go.mod)
[![Next.js](https://img.shields.io/badge/Next.js-16-black)](package.json)

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [架构概览](#架构概览)
- [快速开始](#快速开始)
- [开发环境](#开发环境)
- [配置参考](#配置参考)
- [LaTeX 导入格式](#latex-导入格式)
- [试卷导出](#试卷导出)
- [常用命令](#常用命令)
- [API 接口](#api-接口)
- [排障指南](#排障指南)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

---

## 功能特性

- **题库管理** — 创建、编辑、标签分类、难度分级，支持多张图片附件
- **LaTeX 公式** — 基于 MathJax 的实时渲染预览，CodeMirror 编辑器提供语法高亮
- **批量导入** — 上传 `.tex` 文件，自动解析 `enumerate` / `mybox` / 文本标记三种结构，配套解析文件自动匹配
- **全文搜索** — 关键词 + 公式双索引（PostgreSQL `tsvector`），搜索结果高亮片段
- **试卷组卷** — 拖拽排序、自定义分值、两栏/单栏版式
- **PDF / LaTeX 导出** — XeLaTeX 编译生成高质量 PDF；亦可导出完整 LaTeX 压缩包在 Overleaf 中继续编辑
- **图片管理** — 上传、裁剪、软删除，孤儿图片自动清理
- **深色/浅色主题** — 跟随系统偏好，支持手动切换

---

## 技术栈

| 层次 | 技术 |
|------|------|
| 前端 | Next.js 16 · React 19 · TailwindCSS v4 · better-react-mathjax · CodeMirror 6 |
| 后端 | Go 1.26 · chi · pgx v5 · zerolog |
| 数据库 | PostgreSQL 16（全文索引 `tsvector`，`LISTEN/NOTIFY` 实时推送） |
| 排版引擎 | XeLaTeX（ctex 中文支持）· ImageMagick |
| 容器 | Docker Compose · Node 20 Alpine · golang:1.26-bookworm |

---

## 架构概览

```
┌─────────────────────────────────────────────────────┐
│                    浏览器 (Next.js)                   │
│  题库管理  试卷组卷  导入预览  搜索  导出进度(SSE)    │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP / SSE
┌──────────────────────▼──────────────────────────────┐
│               Go API Server (:8080)                  │
│  ┌───────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ LaTeX     │  │  Export      │  │  Search      │  │
│  │ Parser    │  │  Manager     │  │  Engine      │  │
│  │ (import)  │  │  (XeLaTeX)   │  │  (tsvector)  │  │
│  └───────────┘  └──────────────┘  └──────────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│              PostgreSQL 16 (:5432)                   │
│   problems · papers · exports · images · tags        │
└─────────────────────────────────────────────────────┘
         │
┌────────▼──────────────┐
│  本地文件存储 (volume) │
│  original/ derived/   │
│  thumbnails/          │
└───────────────────────┘
```

---

## 快速开始

### 前置要求

- Docker Engine 20.10+
- Docker Compose V2

### 部署步骤

```bash
# 1. 复制环境变量模板
cp .env.example .env

# 2. 启动所有服务（首次启动会自动执行数据库迁移）
docker compose up -d

# 3. （可选）填充演示数据
docker compose exec server go run ./cmd/mathlib seed
```

访问地址：

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080/api/v1 |
| PostgreSQL | localhost:5432 |

---

## 开发环境

### 前端（Next.js）

```bash
# 安装依赖
pnpm install

# 启动开发服务器（需后端已运行）
pnpm dev
```

### 后端（Go）

```bash
cd server

# 启动数据库（仅 postgres 服务）
docker compose up -d postgres

# 复制并修改环境变量
cp .env.example .env

# 执行数据库迁移
go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 \
  -dir migrations postgres "$DATABASE_URL" up

# 启动 Go 服务
go run ./cmd/mathlib serve
```

### 运行测试

```bash
# Go 后端测试
cd server && go test ./...

# 构建验证
cd server && go build ./...
```

---

## 配置参考

所有配置通过根目录 `.env` 文件（或容器环境变量）管理：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DATABASE_URL` | — | PostgreSQL 连接字符串 |
| `PUBLIC_BASE_URL` | `http://localhost:8080` | 后端对外访问地址（用于生成文件下载链接） |
| `STORAGE_ROOT` | `./storage` | 文件存储根目录（容器内路径） |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS 允许的前端地址（逗号分隔） |
| `AUTO_SEED` | `false` | 首次启动是否自动填充示例数据 |
| `NEXT_PUBLIC_API_BASE_URL` | `/api/v1` | 前端 API 地址（相对路径即可） |
| `XELATEX_PATH` | `xelatex` | XeLaTeX 可执行文件路径 |
| `MAGICK_PATH` | `convert` | ImageMagick convert 路径 |
| `LOG_LEVEL` | `info` | 日志级别（`debug` / `info` / `warn` / `error`） |

### 数据持久化卷

| 卷名 | 挂载路径 | 内容 |
|------|----------|------|
| `pgdata` | postgres 数据目录 | PostgreSQL 数据库文件 |
| `storage` | `/app/storage`（容器内） | 原始图片、导出文件 |

---

## LaTeX 导入格式

访问 `/problems/import` 页面，上传 `.tex` 文件即可批量导入题目。

### 支持的题目结构

| 结构 | 示例标签 | 说明 |
|------|----------|------|
| `enumerate` A 型 | `label=\textbf{题 \arabic*}` | 编号含"题"字 |
| `enumerate` B 型 | `label=\textbf{例\arabic*.}` | 编号含"例"字 |
| `enumerate` C 型 | `label=\arabic*.` | 纯数字编号 |
| `mybox` 环境 | `\begin{mybox}{标题}` | 自定义盒子格式 |
| 文本标记 | `\textbf{例1.}` | 粗体例题标记 |

### 支持的数学分隔符

```
$...$       行内公式
$$...$$     显示公式
\(...\)     行内公式（推荐）
\[...\]     显示公式（推荐）
```

### 章节过滤规则

解析器会跳过标题含以下关键词的 `\section{}` 或 `\section*{}` 节：

`答案` · `解析` · `解答` · `参考` · `简析`

### 配套解析文件命名规则

题目文件 `foo.tex` 的配套解析文件应命名为：
- `foo配套解析.tex`（精确匹配，优先）
- `foo_answers.tex`（后缀匹配）
- 模糊匹配：文件名 Levenshtein 距离 ≤ 3

### 示例片段

```latex
\begin{enumerate}[label=\textbf{题 \arabic*}]
  \item 已知 $f(x) = x^2 - 2x + 1$，求函数的最小值。

  \item 设 $a, b > 0$，证明：
    \[\frac{a+b}{2} \geq \sqrt{ab}\]

  \item 解方程组：
    \[\begin{cases} x + y = 5 \\ x - y = 1 \end{cases}\]
\end{enumerate}
```

---

## 试卷导出

在试卷编辑页面点击「导出」按钮，可生成以下两种格式：

### PDF 导出

- 使用 XeLaTeX 编译，自动处理中文（`ctex` 宏包）
- 支持单栏 / 双栏版式
- 可配置纸张大小（A4 / B5 / Letter）、字号、行距
- 图片自动嵌入，缺失图片显示占位框

### LaTeX 压缩包导出

导出内容包含：
- `main.tex` — 可直接编译的 LaTeX 源文件
- `images/` — 所有引用图片
- `latexmkrc` — 预配置 XeLaTeX 编译选项
- `README.txt` — Overleaf 使用说明

内置宏包：`amsmath` · `amssymb` · `graphicx` · `geometry` · `enumitem`

> 导出内容仅包含题目正文，不包含答案与解析。

---

## 常用命令

### 服务管理

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f web
docker compose logs -f server

# 重启服务
docker compose restart

# 重新构建镜像
docker compose build --no-cache
```

### 数据库操作

```bash
# 手动执行数据库迁移
docker compose exec server sh -c \
  'go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 \
   -dir migrations postgres "$DATABASE_URL" up'

# 填充示例数据
docker compose exec server go run ./cmd/mathlib seed

# 进入 PostgreSQL 交互式终端
docker compose exec postgres psql -U mathlib -d mathlib
```

### 开发调试

```bash
# 进入后端容器
docker compose exec server sh

# 进入前端容器
docker compose exec web sh

# 运行后端测试
docker compose exec server go test ./...

# 手动清理孤儿图片
curl -X POST http://localhost:8080/api/v1/settings/sweep-orphans
```

---

## API 接口

### 题目相关

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/problems/batch-import/preview` | 解析 `.tex` 文件，返回预览草稿 |
| `POST` | `/api/v1/problems/batch-import` | 提交导入（持久化到数据库） |
| `GET` | `/api/v1/problems` | 题目列表（支持分页、筛选） |
| `POST` | `/api/v1/problems` | 新建单道题目 |
| `GET` | `/api/v1/problems/:id` | 题目详情 |
| `PUT` | `/api/v1/problems/:id` | 更新题目 |
| `DELETE` | `/api/v1/problems/:id` | 软删除题目 |

### 搜索

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/search` | 关键词 + 公式搜索，返回高亮片段 |

### 试卷相关

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/papers` | 试卷列表 |
| `POST` | `/api/v1/papers` | 新建试卷 |
| `GET` | `/api/v1/papers/:id` | 试卷详情（含所有题目） |
| `PUT` | `/api/v1/papers/:id` | 更新试卷元数据 |
| `PUT` | `/api/v1/papers/:id/items` | 更新题目顺序与分值 |
| `DELETE` | `/api/v1/papers/:id` | 删除试卷 |

### 导出相关

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/papers/:id/exports` | 创建导出任务（`format`: `pdf`/`latex`；`variant`: `student`/`answer`/`both`） |
| `GET` | `/api/v1/exports/stream` | SSE 流，实时推送导出进度（每 15 秒心跳） |
| `GET` | `/api/v1/exports/:id/download` | 下载导出文件（PDF 或 ZIP） |
| `DELETE` | `/api/v1/exports/:id` | 取消 / 删除导出任务 |

### 运维接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/settings/sweep-orphans` | 手动清理孤儿图片 |

---

## 排障指南

**搜索结果为空**
> 确认已执行数据库迁移，`search_tsv` / `formula_tsv` 触发器是否存在。

**导出任务长时间卡在 `processing`**
> 检查 `xelatex` 是否可执行（`which xelatex`），以及 `STORAGE_ROOT/derived/exports/` 目录是否可写。

**PDF 编译报错 `missing \item`**
> 确认 `.tex` 源文件中 `description` 环境使用了 `enumitem` 宏包选项（如 `style=nextline`），该宏包已内置于导出模板。

**导入预览中 `\section*{答案}` 未被过滤**
> 升级到最新版本，旧版解析器仅匹配 `\section{...}`，不匹配 `\section*{...}`。

**导出 PDF 含有 `°` 度数符号编译失败**
> 源题目中应使用 `^\circ` 代替 `°`，或使用 `\(30°\)` / `\[30°\]` 分隔符（系统会自动转换）；`$30°$` 格式同样支持自动转换（已在最新版修复）。

**多行数学环境（`cases`/`align`/`pmatrix`）显示错误**
> 升级到最新版本，旧版会将数学环境中的 `\\` 换行符全局替换为空格。

**多实例下导出进度不更新**
> 确认 PostgreSQL `LISTEN/NOTIFY` 可用，且所有实例连接到同一数据库。

**图片 404**
> 检查 `PUBLIC_BASE_URL` 与 `STORAGE_ROOT` 是否指向同一套文件系统。

**数学公式在导入预览中显示异常**
> 确认 `.tex` 文件使用标准数学分隔符（`$`、`\(`、`\[`）。

**导出 PDF 进度条卡住**
> 检查 SSE 流端点 `/api/v1/exports/stream` 是否可达，以及反向代理是否关闭了流式响应缓冲（Nginx 需设置 `proxy_buffering off`）。

---

## 贡献指南

欢迎提交 Issue 与 Pull Request！

### 分支命名

```
feat/<简短描述>      新功能
fix/<简短描述>       Bug 修复
docs/<简短描述>      文档更新
refactor/<简短描述>  重构
```

### 提交信息

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat(parser): support \section* in section filter
fix(export): escape LaTeX special chars in paper title
docs: update README import format section
```

### 开发流程

1. Fork 本仓库，基于 `main` 创建功能分支
2. 修改代码并确保 Go 测试全部通过：`cd server && go test ./...`
3. 提交并推送分支
4. 在 GitHub 上开启 Pull Request，描述改动的原因与测试方法

---

## 许可证

本项目采用 [MIT License](LICENSE) 授权。
