# Sitimo

Sitimo 是一个面向中文教师的数学题库管理系统。

## 快速开始

### 前置要求

- Docker Engine 20.10+
- Docker Compose V2

### 部署步骤

1. 复制环境变量模板：
   ```bash
   cp .env.example .env
   ```

2. 启动所有服务：
   ```bash
   docker compose up -d
   ```

3. 访问应用：
   - 前端：http://localhost:3000
   - 后端 API：http://localhost:8080/api/v1

首次启动会自动执行数据库迁移。

### 填充示例数据（可选）

```bash
docker compose exec server go run ./cmd/mathlib seed
```

## 常用命令

### 服务管理

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 查看服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f web
docker compose logs -f server

# 重启服务
docker compose restart

# 重新构建镜像
docker compose build
```

### 数据库操作

```bash
# 手动执行数据库迁移
docker compose exec server sh -c 'go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 -dir migrations postgres "$DATABASE_URL" up'

# 填充示例数据
docker compose exec server go run ./cmd/mathlib seed
```

### 开发调试

```bash
# 进入后端容器
docker compose exec server sh

# 进入前端容器
docker compose exec web sh

# 运行后端测试
docker compose exec server go test ./...
```

## 环境变量

所有配置通过根目录 `.env` 文件管理。关键变量：

- `DATABASE_URL`：数据库连接字符串
- `PUBLIC_BASE_URL`：后端公开访问地址
- `STORAGE_ROOT`：文件存储路径（容器内）
- `ALLOWED_ORIGINS`：CORS 允许的前端地址
- `AUTO_SEED`：首次启动是否自动填充示例数据
- `NEXT_PUBLIC_API_BASE_URL`：前端 API 地址

## 数据持久化

项目使用 Docker 卷持久化数据：

- `pgdata`：PostgreSQL 数据库数据
- `storage`：后端文件存储（图片、导出文件）

## 端口说明

- `3000`：前端服务
- `8080`：后端 API 服务
- `5432`：PostgreSQL 数据库

## 运维接口

- `GET /api/v1/exports/stream`：导出任务 SSE，服务端每 15 秒发送心跳
- `DELETE /api/v1/exports/:id`：`pending` 任务直接删除，`processing` 任务改为请求取消
- `POST /api/v1/settings/sweep-orphans`：手动清理已软删除且未被题目引用的孤儿图片
- 服务启动时会自动执行一次孤儿图片清理

## 导入试卷

访问 `/problems/import` 页面，上传 `.tex` 文件即可批量导入题目。

支持的 LaTeX 格式：
- `enumerate` 环境（`label=\textbf{题 \arabic*}` / `label=\textbf{例\arabic*.}` / `label=\arabic*.`）
- `tasks` 环境（选择题选项，自动识别为 A/B/C/D）
- `mybox` 自定义环境
- 数学分隔符：`$...$`、`$$...$$`、`\(...\)`、`\[...\]` 均支持

配套解析文件命名规则：题目文件 `foo.tex` 对应解析文件 `foo配套解析.tex` 或 `foo_answers.tex`。

## 排障

- 搜索结果为空时，先确认已执行 `make migrate`，尤其是 `search_tsv`/`formula_tsv` 触发器是否存在
- 导出任务长时间停在 `processing`，先检查 `xelatex` 是否可执行，以及 `STORAGE_ROOT/derived/exports` 是否可写
- 多实例下导出流没有更新时，确认 PostgreSQL `LISTEN/NOTIFY` 可用且实例都连到同一个数据库
- 图片文件存在但详情页 404 时，检查 `PUBLIC_BASE_URL` 和 `STORAGE_ROOT` 是否对应同一套文件
- 导入预览中数学公式显示异常时，确认 `.tex` 文件使用标准数学分隔符（`$`、`\(`、`\[`）

## 性能基线

- 题目列表、搜索、图片列表、试卷列表、导出列表、详情查询已经切到定向仓储查询，不再依赖运行时 `LoadSnapshot`
- 关键词搜索走 `search_tsv + plainto_tsquery`，公式搜索走 `formula_tsv + to_tsquery`
- 搜索高亮片段由 PostgreSQL `ts_headline` 生成，前端直接渲染 `<mark>` 片段
- 导出事件源改为 PostgreSQL `LISTEN/NOTIFY`，进程内 `Broadcaster` 只负责实例内 fan-out
