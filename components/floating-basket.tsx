'use client';

import { ShoppingBasket, X, Trash2 } from 'lucide-react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { ScrollArea } from '@/components/ui/scroll-area';
import { MathText } from '@/components/math-text';
import { useBasketStore } from '@/lib/store';
import { difficultyConfig } from '@/lib/types';

interface FloatingBasketProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function FloatingBasket({ open, onOpenChange }: FloatingBasketProps) {
  const { items, removeItem, clearBasket } = useBasketStore();

  return (
    <>
      {/* Floating button */}
      <button
        onClick={() => onOpenChange(true)}
        className="fixed bottom-6 right-6 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-accent text-accent-foreground shadow-lg transition-transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2"
      >
        <ShoppingBasket className="h-6 w-6" />
        {items.length > 0 && (
          <span className="absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center rounded-full bg-background text-xs font-semibold text-primary shadow">
            {items.length}
          </span>
        )}
        <span className="sr-only">打开题目篮子</span>
      </button>

      {/* Sheet */}
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="flex w-96 flex-col p-0">
          <SheetHeader className="border-b px-4 py-3">
            <SheetTitle className="flex items-center gap-2">
              <ShoppingBasket className="h-5 w-5" />
              题目篮子 · 已选 {items.length} 项
            </SheetTitle>
          </SheetHeader>

          {items.length === 0 ? (
            <div className="flex flex-1 flex-col items-center justify-center gap-3 p-6 text-muted-foreground">
              <ShoppingBasket className="h-12 w-12 opacity-50" />
              <p>篮子是空的</p>
              <p className="text-sm">从题库中选择题目添加到这里</p>
            </div>
          ) : (
            <>
              <ScrollArea className="flex-1">
                <div className="divide-y">
                  {items.map((item) => {
                    const config = difficultyConfig[item.difficulty];

                    return (
                      <div
                        key={item.id}
                        className="group flex items-start gap-3 p-4 hover:bg-muted/50"
                      >
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-mono text-xs text-muted-foreground">
                              {item.code}
                            </span>
                            <span
                              className="flex items-center gap-1 text-xs"
                              style={{ color: config.color }}
                            >
                              <span
                                className="h-2 w-2 rounded-full"
                                style={{ backgroundColor: config.color }}
                              />
                              {config.label}
                            </span>
                          </div>
                          <MathText latex={item.latex} className="line-clamp-2 text-sm leading-relaxed" />
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6 shrink-0 opacity-0 group-hover:opacity-100"
                          onClick={() => removeItem(item.id)}
                        >
                          <X className="h-4 w-4" />
                          <span className="sr-only">移除</span>
                        </Button>
                      </div>
                    );
                  })}
                </div>
              </ScrollArea>

              <div className="flex gap-2 border-t p-4">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={clearBasket}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  清空
                </Button>
                <Button asChild className="flex-1">
                  <Link href="/papers/new">前往试卷编辑器</Link>
                </Button>
              </div>
            </>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
