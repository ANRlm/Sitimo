'use client';

import { useMemo, useRef, useState } from 'react';
import { BookType, Database, Download, Info, Palette, Save, Settings2, Trash2, Upload, User } from 'lucide-react';
import { useTheme } from 'next-themes';
import { PageHeader, PageShell } from '@/components/page-shell';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useClearDemoData, useDemoDataStatus, useExportAllData, useImportAllData, useLoadDemoData, useResetDemoData, useSettings, useSweepOrphans, useUpdateSettings } from '@/lib/hooks/use-settings';

type SettingsData = {
  profile?: { name?: string; avatar?: string; notes?: string };
  preferences?: { defaultView?: string; defaultSort?: string; showAnswerVersion?: boolean };
  export?: { columns?: number; lineHeight?: number; paperSize?: string; mathFont?: string; zhFont?: string };
  separators?: { problemStart?: string; problemEnd?: string };
};

type SettingsDraft = {
  defaultView: string;
  defaultSort: string;
  showAnswerVersion: boolean;
  defaultColumns: string;
  defaultLineHeight: string;
  paperSize: string;
  mathFont: string;
  zhFont: string;
  separatorStart: string;
  separatorEnd: string;
  userName: string;
  userAvatar: string;
  notes: string;
};

export default function SettingsPage() {
  const { theme, setTheme } = useTheme();
  const settingsQuery = useSettings();
  const updateSettingsMutation = useUpdateSettings();
  const resetDemoDataMutation = useResetDemoData();
  const exportAllDataMutation = useExportAllData();
  const importAllDataMutation = useImportAllData();
  const sweepOrphansMutation = useSweepOrphans();
  const demoDataStatusQuery = useDemoDataStatus();
  const loadDemoDataMutation = useLoadDemoData();
  const clearDemoDataMutation = useClearDemoData();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [showClearConfirm, setShowClearConfirm] = useState(false);

  const baseDraft = useMemo(
    () => createSettingsDraft((settingsQuery.data as SettingsData | undefined) ?? undefined),
    [settingsQuery.data]
  );
  const [draftOverrides, setDraftOverrides] = useState<Partial<SettingsDraft>>({});
  const draft = { ...baseDraft, ...draftOverrides };

  const setDraftField = <K extends keyof SettingsDraft>(field: K, value: SettingsDraft[K]) => {
    setDraftOverrides((current) => ({ ...current, [field]: value }));
  };

  const handleSave = async () => {
    await updateSettingsMutation.mutateAsync({
      profile: {
        name: draft.userName,
        avatar: draft.userAvatar || undefined,
        notes: draft.notes,
      },
      preferences: {
        defaultView: draft.defaultView,
        defaultSort: draft.defaultSort,
        showAnswerVersion: draft.showAnswerVersion,
      },
      export: {
        columns: Number(draft.defaultColumns) || 1,
        lineHeight: Number(draft.defaultLineHeight) || 1.3,
        paperSize: draft.paperSize,
        mathFont: draft.mathFont,
        zhFont: draft.zhFont,
      },
      separators: {
        problemStart: draft.separatorStart,
        problemEnd: draft.separatorEnd,
      },
    });
  };

  const handleExportAll = async () => {
    const payload = await exportAllDataMutation.mutateAsync();
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `mathlib-export-${new Date().toISOString().slice(0, 10)}.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const handleImportFile = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    const text = await file.text();
    await importAllDataMutation.mutateAsync(JSON.parse(text) as Record<string, unknown>);
    event.target.value = '';
  };

  const handleClearDemoData = async () => {
    await clearDemoDataMutation.mutateAsync();
    setShowClearConfirm(false);
  };

  const demoDataLoaded = demoDataStatusQuery.data?.loaded ?? false;

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="设置"
        title="系统与偏好"
        description="管理 MathLib 的账户、界面偏好、导出模板与本地数据行为。"
        badges={
          <>
            <Badge variant="secondary">当前主题：{theme === 'system' ? '跟随系统' : theme === 'dark' ? '深色' : '浅色'}</Badge>
            <Badge variant="secondary">默认视图：{draft.defaultView === 'grid' ? '网格' : '列表'}</Badge>
            <Badge variant="secondary">导出双版本：{draft.showAnswerVersion ? '开启' : '关闭'}</Badge>
          </>
        }
        actions={
          <Button onClick={handleSave} disabled={updateSettingsMutation.isPending}>
            <Save className="mr-2 h-4 w-4" />
            保存更改
          </Button>
        }
        className="border-primary/15 bg-[linear-gradient(135deg,hsl(var(--primary)/0.08),transparent_52%),linear-gradient(180deg,hsl(var(--background)),hsl(var(--muted)/0.45))]"
      >
        <div className="flex items-center gap-4">
          <Avatar className="h-16 w-16 border border-primary/15">
            <AvatarImage src={draft.userAvatar} />
            <AvatarFallback className="bg-primary text-xl text-primary-foreground">{draft.userName.charAt(0)}</AvatarFallback>
          </Avatar>
          <div className="text-sm text-muted-foreground">{draft.notes || '本地单用户部署，无远程同步。'}</div>
        </div>
      </PageHeader>

      <div className="grid gap-4 xl:grid-cols-[1fr_360px]">
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                账户信息
              </CardTitle>
              <CardDescription>本地管理员的头像、昵称与备注信息。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <Label className="mb-2 block">昵称</Label>
                  <Input value={draft.userName} onChange={(event) => setDraftField('userName', event.target.value)} />
                </div>
                <div>
                  <Label className="mb-2 block">头像链接</Label>
                  <Input value={draft.userAvatar} onChange={(event) => setDraftField('userAvatar', event.target.value)} placeholder="可留空" />
                </div>
              </div>
              <div>
                <Label className="mb-2 block">备注</Label>
                <Textarea value={draft.notes} onChange={(event) => setDraftField('notes', event.target.value)} rows={3} />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings2 className="h-5 w-5" />
                偏好设置
              </CardTitle>
              <CardDescription>统一控制默认主题、视图与排序。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <Label>主题</Label>
                <RadioGroup value={theme} onValueChange={setTheme} className="grid gap-3 md:grid-cols-3">
                  <Label htmlFor="theme-light" className="flex cursor-pointer items-center gap-3 rounded-2xl border p-4 [&:has([data-state=checked])]:border-primary [&:has([data-state=checked])]:bg-primary/5">
                    <RadioGroupItem value="light" id="theme-light" />
                    <div>
                      <p className="font-medium">浅色</p>
                      <p className="text-sm text-muted-foreground">适合白天录题和审校</p>
                    </div>
                  </Label>
                  <Label htmlFor="theme-dark" className="flex cursor-pointer items-center gap-3 rounded-2xl border p-4 [&:has([data-state=checked])]:border-primary [&:has([data-state=checked])]:bg-primary/5">
                    <RadioGroupItem value="dark" id="theme-dark" />
                    <div>
                      <p className="font-medium">深色</p>
                      <p className="text-sm text-muted-foreground">适合长时间夜间编辑</p>
                    </div>
                  </Label>
                  <Label htmlFor="theme-system" className="flex cursor-pointer items-center gap-3 rounded-2xl border p-4 [&:has([data-state=checked])]:border-primary [&:has([data-state=checked])]:bg-primary/5">
                    <RadioGroupItem value="system" id="theme-system" />
                    <div>
                      <p className="font-medium">跟随系统</p>
                      <p className="text-sm text-muted-foreground">自动跟随桌面外观</p>
                    </div>
                  </Label>
                </RadioGroup>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <Label className="mb-2 block">默认视图</Label>
                  <Select value={draft.defaultView} onValueChange={(value) => setDraftField('defaultView', value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="grid">网格</SelectItem>
                      <SelectItem value="list">列表</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label className="mb-2 block">默认排序</Label>
                  <Select value={draft.defaultSort} onValueChange={(value) => setDraftField('defaultSort', value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="updatedAt-desc">最近更新</SelectItem>
                      <SelectItem value="createdAt-desc">最近创建</SelectItem>
                      <SelectItem value="code-asc">题号升序</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Palette className="h-5 w-5" />
                导出默认值
              </CardTitle>
              <CardDescription>试卷排版与 LaTeX/PDF 输出偏好。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-xl border p-3">
                <div>
                  <p className="font-medium">默认导出答案版</p>
                  <p className="text-sm text-muted-foreground">新建试卷时默认启用双版本导出</p>
                </div>
                <Switch checked={draft.showAnswerVersion} onCheckedChange={(value) => setDraftField('showAnswerVersion', value)} />
              </div>
              <div className="grid gap-4 md:grid-cols-3">
                <div>
                  <Label className="mb-2 block">分栏</Label>
                  <Select value={draft.defaultColumns} onValueChange={(value) => setDraftField('defaultColumns', value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">单栏</SelectItem>
                      <SelectItem value="2">双栏</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label className="mb-2 block">行距</Label>
                  <Input value={draft.defaultLineHeight} onChange={(event) => setDraftField('defaultLineHeight', event.target.value)} />
                </div>
                <div>
                  <Label className="mb-2 block">纸张</Label>
                  <Select value={draft.paperSize} onValueChange={(value) => setDraftField('paperSize', value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="A4">A4</SelectItem>
                      <SelectItem value="B5">B5</SelectItem>
                      <SelectItem value="Letter">Letter</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BookType className="h-5 w-5" />
                字体与分隔符
              </CardTitle>
              <CardDescription>数学字体、中文字体和批量导入解析规则。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <Label className="mb-2 block">数学字体</Label>
                  <Input value={draft.mathFont} onChange={(event) => setDraftField('mathFont', event.target.value)} />
                </div>
                <div>
                  <Label className="mb-2 block">中文字体</Label>
                  <Input value={draft.zhFont} onChange={(event) => setDraftField('zhFont', event.target.value)} />
                </div>
              </div>
              <Separator />
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <Label className="mb-2 block">题目开始标记</Label>
                  <Input value={draft.separatorStart} onChange={(event) => setDraftField('separatorStart', event.target.value)} />
                </div>
                <div>
                  <Label className="mb-2 block">题目结束标记</Label>
                  <Input value={draft.separatorEnd} onChange={(event) => setDraftField('separatorEnd', event.target.value)} />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="h-5 w-5" />
                数据管理
              </CardTitle>
              <CardDescription>导入、导出和重载示例数据。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button variant="outline" className="w-full justify-start" onClick={handleExportAll} disabled={exportAllDataMutation.isPending}>
                <Download className="mr-2 h-4 w-4" />
                导出全部数据
              </Button>
              <input ref={fileInputRef} type="file" accept="application/json" className="hidden" onChange={handleImportFile} />
              <Button variant="outline" className="w-full justify-start" onClick={() => fileInputRef.current?.click()} disabled={importAllDataMutation.isPending}>
                <Upload className="mr-2 h-4 w-4" />
                导入 JSON 数据
              </Button>
              <Button variant="outline" className="w-full justify-start" onClick={() => sweepOrphansMutation.mutate()} disabled={sweepOrphansMutation.isPending}>
                <Palette className="mr-2 h-4 w-4" />
                清理孤儿图片
              </Button>
              <Button className="w-full justify-start" onClick={() => resetDemoDataMutation.mutate()} disabled={resetDemoDataMutation.isPending}>
                <Database className="mr-2 h-4 w-4" />
                重载示例数据
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="h-5 w-5" />
                示例数据管理
              </CardTitle>
              <CardDescription>加载或删除示例数据。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between rounded-xl border p-3">
                <div>
                  <p className="font-medium">示例数据状态</p>
                  <p className="text-sm text-muted-foreground">{demoDataLoaded ? '已加载' : '未加载'}</p>
                </div>
                <Badge variant={demoDataLoaded ? 'default' : 'secondary'}>{demoDataLoaded ? '已加载' : '未加载'}</Badge>
              </div>
              <Button variant="outline" className="w-full justify-start" onClick={() => loadDemoDataMutation.mutate()} disabled={loadDemoDataMutation.isPending}>
                <Database className="mr-2 h-4 w-4" />
                加载示例数据
              </Button>
              <Button
                variant="outline"
                className="w-full justify-start"
                onClick={() => setShowClearConfirm(true)}
                disabled={!demoDataLoaded || clearDemoDataMutation.isPending}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                删除示例数据
              </Button>
            </CardContent>
          </Card>

          <AlertDialog open={showClearConfirm} onOpenChange={setShowClearConfirm}>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>确认删除示例数据</AlertDialogTitle>
                <AlertDialogDescription>此操作将删除所有示例数据，且无法恢复。确定要继续吗？</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>取消</AlertDialogCancel>
                <AlertDialogAction onClick={handleClearDemoData}>确认删除</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Info className="h-5 w-5" />
                关于
              </CardTitle>
              <CardDescription>当前本地部署的技术摘要。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">前端</span>
                <span>Next.js 16 / React 19</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">后端</span>
                <span>Go + PostgreSQL</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">导出引擎</span>
                <span>XeLaTeX / Docker</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </PageShell>
  );
}

function createSettingsDraft(settings?: SettingsData): SettingsDraft {
  return {
    defaultView: settings?.preferences?.defaultView ?? 'grid',
    defaultSort: settings?.preferences?.defaultSort ?? 'updatedAt-desc',
    showAnswerVersion: settings?.preferences?.showAnswerVersion ?? true,
    defaultColumns: String(settings?.export?.columns ?? 1),
    defaultLineHeight: String(settings?.export?.lineHeight ?? 1.3),
    paperSize: settings?.export?.paperSize ?? 'A4',
    mathFont: settings?.export?.mathFont ?? 'Latin Modern Math',
    zhFont: settings?.export?.zhFont ?? 'Source Han Sans SC',
    separatorStart: settings?.separators?.problemStart ?? '\\begin{problem}',
    separatorEnd: settings?.separators?.problemEnd ?? '\\end{problem}',
    userName: settings?.profile?.name ?? '张老师',
    userAvatar: settings?.profile?.avatar ?? '',
    notes: settings?.profile?.notes ?? '本地单用户部署，无远程同步。',
  };
}
