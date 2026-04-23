# MathLib 功能增强实施计划

**创建时间**: 2026-04-22  
**基于文档**: 
- `.omc/autopilot/requirements.md`
- `.omc/autopilot/spec.md`

**预估总时间**: 12 小时（开发 9h + 测试 3h）

---

## 执行摘要

实现 5 个核心功能增强：批量删除、示例数据系统、学科字段移除、年级标准化、试卷留空行。采用 3 阶段执行策略，优先完成数据库和后端基础，再实现前端功能，最后集成测试。

---

## 1. 任务分解

### Feature 1: 批量删除功能

#### Task 1.1: 实现图片批量删除 API
- **文件**: `server/internal/api/images.go`, `server/internal/service/images.go`, `server/internal/store/images.go`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 45min
- **优先级**: P0
- **关键点**:
  - 新增 `POST /api/images/batch-delete` 端点
  - 请求体: `{"imageIds": ["id1", "id2"]}`
  - 响应: `{"ok": true, "deleted": 2}`
  - 使用事务确保原子性
  - 软删除实现（设置 deleted_at）
- **测试验证**:
  - 批量删除 10 张图片 < 500ms
  - 事务回滚测试（部分 ID 无效）
  - 被引用图片删除后题目显示占位符

#### Task 1.2: 实现题目批量删除 API
- **文件**: `server/internal/api/problems.go`, `server/internal/service/problems.go`, `server/internal/store/problems.go`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 45min
- **优先级**: P0
- **关键点**:
  - 新增 `POST /api/problems/batch-delete` 端点
  - 请求体: `{"problemIds": ["id1", "id2"]}`
  - 级联删除关联的 paper_items
  - 软删除实现
- **测试验证**:
  - 批量删除 100 题 < 2s
  - 试卷中的题目被删除后正确处理

#### Task 1.3: 图库批量删除 UI
- **文件**: `app/images/page.tsx`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 60min
- **优先级**: P0
- **依赖**: Task 1.1
- **关键点**:
  - 添加复选框列
  - 全选/取消全选功能
  - 批量删除按钮（选中 > 0 时显示）
  - 删除确认对话框
  - 使用 TanStack Query mutation
- **测试验证**:
  - 选中状态正确切换
  - 删除后列表自动刷新
  - 错误提示正确显示

#### Task 1.4: 题库批量删除 UI
- **文件**: `app/problems/page.tsx`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 60min
- **优先级**: P0
- **依赖**: Task 1.2
- **关键点**:
  - 同 Task 1.3
  - 支持筛选后的批量删除
- **测试验证**:
  - 同 Task 1.3

---

### Feature 2: 示例数据系统

#### Task 2.1: 创建示例数据服务
- **文件**: `server/internal/service/demo.go` (新建)
- **复杂度**: 复杂
- **代理**: Opus
- **时间**: 90min
- **优先级**: P0
- **关键点**:
  - 实现 `LoadDemoData()` 方法
  - 实现 `ClearDemoData()` 方法
  - 实现 `GetDemoDataStatus()` 方法
  - 使用 ID 前缀 `demo-` 标识示例数据
  - 创建完整示例数据集：
    - 10 个题目（覆盖各年级、难度）
    - 5 张图片
    - 2 份试卷
    - 5 个标签
  - 事务保证原子性
- **测试验证**:
  - 加载示例数据后可正常使用
  - 清除示例数据不影响用户数据
  - 重复加载幂等性

#### Task 2.2: 示例数据管理 API
- **文件**: `server/internal/api/settings.go`
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 30min
- **优先级**: P0
- **依赖**: Task 2.1
- **关键点**:
  - `POST /api/settings/demo-data/load`
  - `POST /api/settings/demo-data/clear`
  - `GET /api/settings/demo-data/status`
- **测试验证**:
  - API 端点正确调用服务层
  - 错误处理正确

#### Task 2.3: 示例数据管理 UI
- **文件**: `app/settings/page.tsx`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 45min
- **优先级**: P0
- **依赖**: Task 2.2
- **关键点**:
  - 新增"示例数据"设置区域
  - "加载示例数据"按钮
  - "清除示例数据"按钮（仅在有示例数据时显示）
  - 示例数据状态显示
  - 操作确认对话框
- **测试验证**:
  - 按钮状态正确切换
  - 操作成功后显示提示
  - 页面其他功能不受影响

---

### Feature 3: 移除学科字段

#### Task 3.1: 移除学科 API 端点
- **文件**: `server/internal/api/meta.go`, `server/internal/service/meta.go`
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 15min
- **优先级**: P1
- **关键点**:
  - 删除 `GET /api/meta/subjects` 端点
  - 清理相关服务层代码
- **测试验证**:
  - 端点返回 404
  - 其他 meta 端点正常工作

#### Task 3.2: 移除学科前端代码
- **文件**: `app/problems/page.tsx`, `app/problems/[id]/edit/page.tsx`, `lib/hooks/use-meta.ts`
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 30min
- **优先级**: P1
- **依赖**: Task 3.1
- **关键点**:
  - 删除学科选择下拉框
  - 删除 `useSubjects` hook
  - 清理相关类型定义
  - 移除学科筛选功能
- **测试验证**:
  - 题目编辑页面正常显示
  - 题库列表页面正常工作
  - 无 TypeScript 错误

---

### Feature 4: 年级标准化

#### Task 4.1: 创建年级常量
- **文件**: `lib/constants.ts` (新建)
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 10min
- **优先级**: P1
- **关键点**:
  - 定义标准年级列表：`["初一", "初二", "初三", "高一", "高二", "高三"]`
  - 导出类型定义
- **测试验证**:
  - 常量可正确导入

#### Task 4.2: 更新后端年级类型
- **文件**: `server/internal/domain/types.go`
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 15min
- **优先级**: P1
- **关键点**:
  - 添加年级常量定义
  - 更新验证逻辑（允许标准值 + 自定义值）
- **测试验证**:
  - API 接受标准年级值
  - API 接受历史自定义值

#### Task 4.3: 年级标准化 UI
- **文件**: `app/problems/[id]/edit/page.tsx`
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 30min
- **优先级**: P1
- **依赖**: Task 4.1, Task 4.2
- **关键点**:
  - 年级输入改为下拉选择
  - 使用 `lib/constants.ts` 中的标准列表
  - 历史自定义值显示为"自定义: {value}"
  - 保留自定义输入选项（高级用户）
- **测试验证**:
  - 下拉列表显示 6 个标准年级
  - 历史数据正确显示
  - 新建题目使用标准年级

---

### Feature 5: 试卷留空行功能

#### Task 5.1: 数据库迁移
- **文件**: `server/migrations/006_add_blank_lines_to_paper_items.sql` (新建)
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 15min
- **优先级**: P0
- **关键点**:
  - 添加 `blank_lines INTEGER NOT NULL DEFAULT 0`
  - 添加约束 `CHECK (blank_lines >= 0 AND blank_lines <= 10)`
  - 提供回滚脚本
- **测试验证**:
  - 迁移成功执行
  - 回滚成功执行
  - 现有数据不受影响

#### Task 5.2: 更新后端类型和 API
- **文件**: `server/internal/domain/types.go`, `server/internal/store/sqlc/models.go`, `server/internal/api/papers.go`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 30min
- **优先级**: P0
- **依赖**: Task 5.1
- **关键点**:
  - 更新 `PaperItem` 结构体添加 `BlankLines` 字段
  - 更新 sqlc 查询
  - 更新 API 验证逻辑
- **测试验证**:
  - API 接受 blankLines 字段
  - 验证范围 0-10
  - 数据库正确存储

#### Task 5.3: 更新前端类型
- **文件**: `lib/types.ts` (或相关类型文件)
- **复杂度**: 简单
- **代理**: Haiku
- **时间**: 10min
- **优先级**: P0
- **依赖**: Task 5.2
- **关键点**:
  - 更新 `PaperItem` 接口添加 `blankLines?: number`
- **测试验证**:
  - 无 TypeScript 错误

#### Task 5.4: 试卷编辑器留空行 UI
- **文件**: `app/papers/[id]/editor/page.tsx`
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 60min
- **优先级**: P0
- **依赖**: Task 5.3
- **关键点**:
  - 每个题目项添加"留空行"控制
  - 使用数字输入框（范围 0-10）
  - 实时预览留空行效果
  - 保存时包含 blankLines 字段
- **测试验证**:
  - 留空行数量可调整
  - 预览正确显示
  - 保存后刷新数据正确

#### Task 5.5: 更新 LaTeX 导出逻辑
- **文件**: `server/internal/export/latex.go` (或相关导出文件)
- **复杂度**: 标准
- **代理**: Sonnet
- **时间**: 45min
- **优先级**: P0
- **依赖**: Task 5.2
- **关键点**:
  - 在题目后插入 `\vspace{Xem}`
  - X = blankLines 值
  - 仅在 blankLines > 0 时插入
- **测试验证**:
  - PDF 导出包含留空行
  - 留空行高度正确
  - 不影响其他导出格式

---

## 2. 依赖关系图

```
Phase 1: 数据库和后端基础
├── Task 1.1 (图片批量删除 API) ─┐
├── Task 1.2 (题目批量删除 API) ─┤
├── Task 2.1 (示例数据服务) ─────┤ 可并行
├── Task 3.1 (移除学科 API) ─────┤
├── Task 4.1 (年级常量) ──────────┤
├── Task 4.2 (后端年级类型) ─────┤
└── Task 5.1 (数据库迁移) ────────┘
     │
     ├─→ Task 5.2 (后端类型和 API) ──→ Task 5.3 (前端类型)
     │
     └─→ Task 2.2 (示例数据 API)

Phase 2: 前端核心功能
├── Task 1.3 (图库批量删除 UI) ──┐ 依赖 Task 1.1
├── Task 1.4 (题库批量删除 UI) ──┤ 依赖 Task 1.2
├── Task 2.3 (示例数据 UI) ───────┤ 依赖 Task 2.2  可并行
├── Task 3.2 (移除学科前端) ──────┤ 依赖 Task 3.1
└── Task 4.3 (年级标准化 UI) ─────┘ 依赖 Task 4.1, 4.2

Phase 3: 集成和优化
├── Task 5.4 (留空行 UI) ─────────┐ 依赖 Task 5.3
└── Task 5.5 (LaTeX 导出) ────────┘ 依赖 Task 5.2
```

---

## 3. 实现步骤详解

### Phase 1: 数据库和后端基础 (串行关键路径，部分并行)

**并行组 1** (可同时执行):
- Task 1.1: 图片批量删除 API
- Task 1.2: 题目批量删除 API
- Task 2.1: 示例数据服务
- Task 3.1: 移除学科 API
- Task 4.1: 年级常量
- Task 4.2: 后端年级类型

**串行组 1**:
- Task 5.1: 数据库迁移 → Task 5.2: 后端类型和 API → Task 5.3: 前端类型

**串行组 2**:
- Task 2.1: 示例数据服务 → Task 2.2: 示例数据 API

**预估时间**: 3.5 小时（并行执行）

---

### Phase 2: 前端核心功能 (部分并行)

**并行组 2** (可同时执行):
- Task 1.3: 图库批量删除 UI
- Task 1.4: 题库批量删除 UI
- Task 2.3: 示例数据 UI
- Task 3.2: 移除学科前端
- Task 4.3: 年级标准化 UI

**预估时间**: 1 小时（并行执行）

---

### Phase 3: 集成和优化 (串行)

**串行组 3**:
- Task 5.4: 留空行 UI
- Task 5.5: LaTeX 导出

**预估时间**: 1.75 小时

---

## 4. 风险评估与缓解策略

### 高风险任务

#### Task 1.1 & 1.2: 批量删除 API
**风险**:
- 大量数据删除可能超时
- 事务锁定导致性能问题
- 级联删除逻辑复杂

**缓解策略**:
- 限制单次删除数量（最多 100 项）
- 使用批量 SQL 操作而非循环
- 添加超时保护（30s）
- 充分测试事务回滚

#### Task 2.1: 示例数据服务
**风险**:
- 示例数据与用户数据冲突（ID 重复）
- 示例数据质量不足
- 清除逻辑误删用户数据

**缓解策略**:
- 使用唯一前缀 `demo-` 确保 ID 不冲突
- 精心设计示例数据覆盖所有功能
- 清除时严格检查 ID 前缀
- 添加二次确认机制

#### Task 5.5: LaTeX 导出逻辑
**风险**:
- LaTeX 语法错误导致 PDF 生成失败
- 留空行高度不符合预期
- 影响现有导出功能

**缓解策略**:
- 使用标准 LaTeX 命令 `\vspace{Xem}`
- 充分测试各种留空行数量
- 保持向后兼容（blankLines = 0 时无变化）
- 添加导出错误日志

---

### 中风险任务

#### Task 3.2: 移除学科前端代码
**风险**:
- 遗漏部分学科相关代码
- 影响其他功能

**缓解策略**:
- 全局搜索 "subject" 关键词
- 检查所有筛选和表单组件
- 回归测试题目 CRUD 功能

#### Task 5.4: 留空行 UI
**风险**:
- UI 交互不直观
- 预览效果不准确

**缓解策略**:
- 参考常见试卷编辑器设计
- 提供实时预览
- 添加使用提示

---

### 低风险任务

- Task 4.1, 4.2, 4.3: 年级标准化（向后兼容）
- Task 5.1: 数据库迁移（简单字段添加）
- Task 2.2, 2.3: 示例数据 UI（独立功能）

---

## 5. 执行策略

### 推荐执行顺序

**Day 1 (4 小时)**: Phase 1 完成
1. 启动 6 个并行任务（Task 1.1, 1.2, 2.1, 3.1, 4.1, 4.2）
2. 完成后执行 Task 5.1 → 5.2 → 5.3
3. 完成后执行 Task 2.2

**Day 2 (3 小时)**: Phase 2 完成
1. 启动 5 个并行任务（Task 1.3, 1.4, 2.3, 3.2, 4.3）
2. 完成后进入 Phase 3

**Day 2 (2 小时)**: Phase 3 完成
1. 执行 Task 5.4
2. 执行 Task 5.5

**Day 3 (3 小时)**: 测试和优化
1. 端到端测试所有功能
2. 性能测试批量删除
3. 回归测试现有功能
4. 修复发现的问题

---

## 6. 验收标准

### 功能验收

#### Feature 1: 批量删除
- [ ] 图库页面显示复选框和批量删除按钮
- [ ] 题库页面显示复选框和批量删除按钮
- [ ] 批量删除 10 项 < 500ms
- [ ] 批量删除 100 项 < 2s
- [ ] 删除确认对话框正确显示
- [ ] 删除后列表自动刷新

#### Feature 2: 示例数据
- [ ] 设置页面显示示例数据管理区域
- [ ] 可成功加载示例数据
- [ ] 示例数据包含题目、图片、试卷、标签
- [ ] 示例数据可正常使用（查看、编辑、删除）
- [ ] 可一键清除所有示例数据
- [ ] 清除示例数据不影响用户数据

#### Feature 3: 学科移除
- [ ] 题目编辑页面无学科选择
- [ ] 题库列表页面无学科筛选
- [ ] `/api/meta/subjects` 返回 404
- [ ] 现有题目数据不受影响

#### Feature 4: 年级标准化
- [ ] 题目编辑页面年级为下拉选择
- [ ] 下拉列表包含 6 个标准年级
- [ ] 历史自定义年级显示为"自定义: {value}"
- [ ] 新建题目默认使用标准年级

#### Feature 5: 留空行
- [ ] 试卷编辑器每个题目显示留空行控制
- [ ] 留空行数量可调整（0-10）
- [ ] 预览正确显示留空行
- [ ] PDF 导出包含留空行
- [ ] 留空行高度符合预期（1em 单位）

---

### 技术验收

#### 数据库
- [ ] 迁移脚本无错误执行
- [ ] 回滚脚本无错误执行
- [ ] 新增字段约束正确
- [ ] 现有数据完整性保持

#### 后端
- [ ] 所有新增 API 端点正常工作
- [ ] API 响应格式符合规范
- [ ] 错误处理完整
- [ ] 日志记录充分
- [ ] 无 Go 编译错误

#### 前端
- [ ] 无 TypeScript 错误
- [ ] 无 ESLint 警告
- [ ] 所有组件正确渲染
- [ ] 状态管理正确
- [ ] 错误提示友好

#### 性能
- [ ] 批量删除 100 项 < 2s
- [ ] 示例数据加载 < 3s
- [ ] 页面加载时间无明显增加
- [ ] 内存使用无异常增长

---

## 7. 回滚计划

### 数据库回滚
```sql
-- 如果 Task 5.1 出现问题
ALTER TABLE paper_items DROP COLUMN blank_lines;
```

### API 回滚
- 删除新增的 API 端点
- 恢复 `/api/meta/subjects` 端点（如果 Task 3.1 有问题）

### 前端回滚
- 使用 Git 回滚到功能开发前的提交
- 逐个功能回滚（通过 feature branch）

---

## 8. 成功指标

- **功能完整性**: 5/5 功能全部实现并通过验收
- **性能达标**: 批量操作 < 2s
- **零回归**: 现有功能无破坏
- **代码质量**: 无 TypeScript/Go 错误，无 ESLint 警告
- **测试覆盖**: 所有新增 API 有测试用例
- **文档完整**: README 更新，API 文档更新

---

## 9. 后续优化建议

### 短期 (1-2 周)
- 添加批量操作的撤销功能
- 优化批量删除性能（使用后台任务）
- 增加更多示例数据

### 中期 (1-2 月)
- 批量编辑功能（批量修改年级、标签）
- 示例数据自定义（用户可选择加载哪些示例）
- 留空行模板（预设常用留空行配置）

### 长期 (3-6 月)
- 数据导入导出（Excel/CSV）
- 版本历史和恢复
- 协作功能（多用户）

---

## 附录：文件变更清单

### 新建文件
- `server/migrations/006_add_blank_lines_to_paper_items.sql`
- `server/internal/service/demo.go`
- `lib/constants.ts`

### 修改文件

#### 后端 (Go)
- `server/internal/api/images.go` - 批量删除端点
- `server/internal/api/problems.go` - 批量删除端点
- `server/internal/api/papers.go` - 留空行字段
- `server/internal/api/settings.go` - 示例数据管理端点
- `server/internal/api/meta.go` - 移除学科端点
- `server/internal/service/images.go` - 批量删除逻辑
- `server/internal/service/problems.go` - 批量删除逻辑
- `server/internal/service/meta.go` - 移除学科逻辑
- `server/internal/store/images.go` - 批量删除查询
- `server/internal/store/problems.go` - 批量删除查询
- `server/internal/domain/types.go` - 年级类型、留空行字段
- `server/internal/export/latex.go` - 留空行导出

#### 前端 (TypeScript/React)
- `app/images/page.tsx` - 批量删除 UI
- `app/problems/page.tsx` - 批量删除 UI
- `app/settings/page.tsx` - 示例数据管理 UI
- `app/problems/[id]/edit/page.tsx` - 移除学科、年级标准化
- `app/papers/[id]/editor/page.tsx` - 留空行 UI
- `lib/types.ts` - 类型定义更新
- `lib/hooks/use-meta.ts` - 移除学科 hook

---

**计划状态**: Ready for Execution  
**下一步**: 启动 Phase 1 并行任务组
