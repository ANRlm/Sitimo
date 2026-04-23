'use client';

import { useMemo, useState } from 'react';
import { Compass, ImageIcon, LibraryBig, School } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

const steps = [
  {
    title: '欢迎使用 MathLib',
    description: 'MathLib 主要由题库、图库和试卷三大模块组成，先看一遍结构再开始录题会更快。',
    icon: Compass,
  },
  {
    title: '高亮 Sidebar · 题库',
    description: '左侧 Sidebar 中的“题库”是核心入口，可筛选、批量导入、加入篮子，再继续下游组卷。',
    icon: LibraryBig,
  },
  {
    title: '高亮 Sidebar · 图库',
    description: '“图库”里可以管理几何图、函数图像和配图，并把图片关联到具体题目。',
    icon: ImageIcon,
  },
  {
    title: '高亮 Sidebar · 试卷',
    description: '进入“试卷”后，从题目篮子挑题、调整分值和版式，再导出 PDF 或 Overleaf 可直接导入的 LaTeX 包。',
    icon: School,
  },
];

type OnboardingTourProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: () => void;
};

export function OnboardingTour({
  open,
  onOpenChange,
  onFinish,
}: OnboardingTourProps) {
  const [step, setStep] = useState(0);
  const activeStep = useMemo(() => steps[step], [step]);
  const Icon = activeStep.icon;

  const handleClose = () => {
    setStep(0);
    onOpenChange(false);
    onFinish();
  };

  const handleNext = () => {
    if (step === steps.length - 1) {
      handleClose();
      return;
    }

    setStep((current) => current + 1);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px]">
        <DialogHeader>
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <Icon className="h-6 w-6" />
          </div>
          <DialogTitle>{activeStep.title}</DialogTitle>
          <DialogDescription className="leading-7">
            {activeStep.description}
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2">
          {steps.map((item, index) => (
            <div
              key={item.title}
              className={`h-2 flex-1 rounded-full ${index <= step ? 'bg-primary' : 'bg-muted'}`}
            />
          ))}
        </div>

        <div className="rounded-2xl border bg-muted/40 p-4 text-sm text-muted-foreground">
          第 {step + 1} 步，共 {steps.length} 步。完成后会写入 localStorage，避免重复弹出。
        </div>

        <DialogFooter className="justify-between sm:justify-between">
          <Button variant="ghost" onClick={handleClose}>
            跳过
          </Button>
          <Button onClick={handleNext}>
            {step === steps.length - 1 ? '完成' : '下一步'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
