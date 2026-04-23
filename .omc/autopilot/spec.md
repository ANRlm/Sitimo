# 按钮图标对齐问题修复规格

## 问题描述

在题目管理页面（`app/problems/page.tsx`）的题目卡片底部，存在三个操作按钮的视觉对齐问题：

- **查看**按钮：仅包含文字，无图标
- **编辑**按钮：包含铅笔图标 + 文字
- **入篮**按钮：包含购物篮图标 + 文字

这种不一致导致按钮在视觉上不对齐，用户体验不佳。

## 当前实现

位置：`app/problems/page.tsx:306-338`

```tsx
<div className="shrink-0 p-5 pt-4">
  <div className="grid gap-2 sm:grid-cols-3">
    <Button variant="outline" size="sm" asChild className="w-full">
      <Link href={`/problems/${problem.id}`}>查看</Link>
    </Button>
    <Button variant="outline" size="sm" asChild className="w-full">
      <Link href={`/problems/${problem.id}/edit`}>
        <Pencil className="mr-2 h-4 w-4" />
        编辑
      </Link>
    </Button>
    <Button
      size="sm"
      className="w-full"
      variant={inBasket ? 'outline' : 'default'}
      onClick={() => { /* ... */ }}
    >
      {inBasket ? <Check className="mr-2 h-4 w-4" /> : <ShoppingBasket className="mr-2 h-4 w-4" />}
      {inBasket ? '已选' : '入篮'}
    </Button>
  </div>
</div>
```

## 解决方案

为"查看"按钮添加合适的图标，使三个按钮保持视觉一致性。

### 图标选择

使用 `Eye` 图标（来自 lucide-react），因为：
1. 语义明确：表示"查看/预览"操作
2. 已在项目中使用 lucide-react 图标库
3. 与其他按钮图标风格一致

### 修改内容

1. 在文件顶部导入 `Eye` 图标：
   ```tsx
   import { Check, Eye, Filter, Image as ImageIcon, LayoutGrid, List, Pencil, Plus, Search, ShoppingBasket, Trash2, X } from 'lucide-react';
   ```

2. 修改"查看"按钮，添加图标：
   ```tsx
   <Button variant="outline" size="sm" asChild className="w-full">
     <Link href={`/problems/${problem.id}`}>
       <Eye className="mr-2 h-4 w-4" />
       查看
     </Link>
   </Button>
   ```

## 实施步骤

1. 在第 5 行的 import 语句中添加 `Eye` 图标
2. 在第 308-310 行的"查看"按钮中添加 `Eye` 图标

## 预期效果

修复后，三个按钮将保持一致的视觉样式：
- 所有按钮都包含图标 + 文字
- 图标大小统一（h-4 w-4）
- 图标与文字间距统一（mr-2）
- 整体视觉对齐，用户体验更佳

## 影响范围

- 文件：`app/problems/page.tsx`
- 影响行：第 5 行（导入）、第 308-310 行（按钮实现）
- 风险：低（仅添加图标，不改变功能逻辑）

## 测试要点

1. 视觉检查：三个按钮图标对齐
2. 功能检查：点击"查看"按钮正常跳转
3. 响应式检查：不同屏幕尺寸下按钮布局正常
4. 无障碍检查：图标不影响屏幕阅读器
