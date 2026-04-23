# 按钮图标对齐修复实施计划

## 执行策略
按顺序执行，每步验证后再继续

## Phase 2: 执行修复

### 任务 2.1: 添加 Eye 图标导入
**优先级**: 高  
**执行者**: Haiku  
**文件**: app/problems/page.tsx (行 5)

**修改内容**:
在 lucide-react 导入语句中添加 `Eye` 图标

**当前代码**:
```tsx
import { Check, Filter, Image as ImageIcon, LayoutGrid, List, Pencil, Plus, Search, ShoppingBasket, Trash2, X } from 'lucide-react';
```

**修改后**:
```tsx
import { Check, Eye, Filter, Image as ImageIcon, LayoutGrid, List, Pencil, Plus, Search, ShoppingBasket, Trash2, X } from 'lucide-react';
```

**验收**: Eye 图标成功导入，无 TypeScript 错误

### 任务 2.2: 为"查看"按钮添加图标
**优先级**: 高  
**执行者**: Haiku  
**文件**: app/problems/page.tsx (行 308-310)

**修改内容**:
在"查看"按钮的 Link 组件内添加 Eye 图标

**当前代码**:
```tsx
<Button variant="outline" size="sm" asChild className="w-full">
  <Link href={`/problems/${problem.id}`}>查看</Link>
</Button>
```

**修改后**:
```tsx
<Button variant="outline" size="sm" asChild className="w-full">
  <Link href={`/problems/${problem.id}`}>
    <Eye className="mr-2 h-4 w-4" />
    查看
  </Link>
</Button>
```

**验收**: 按钮显示图标，与其他按钮视觉对齐

## Phase 3: QA 验证

### QA 循环 1: 构建检查
1. 运行 `pnpm build` 确保无编译错误
2. 运行 `pnpm lint` 确保无 lint 错误

### QA 循环 2: 视觉验证
1. 启动开发服务器 `pnpm dev`
2. 访问 /problems 页面
3. 检查三个按钮图标是否对齐
4. 测试响应式布局（sm、md、lg 断点）

### QA 循环 3: 功能测试
1. 点击"查看"按钮，确认正常跳转到题目详情页
2. 点击"编辑"按钮，确认正常跳转到编辑页
3. 点击"入篮"按钮，确认功能正常

## 依赖关系
- 任务 2.2 依赖 2.1（必须先导入图标）
- Phase 3 依赖 Phase 2 完成

## 回滚计划
如果修改后出现问题，可以通过 git 恢复到之前的版本

## 预期时间
- Phase 2: 2 分钟
- Phase 3: 3 分钟
- **总计**: 5 分钟
