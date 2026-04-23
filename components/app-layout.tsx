'use client';

import { useState } from 'react';
import { AppSidebar } from './app-sidebar';
import { AppHeader } from './app-header';
import { FloatingBasket } from './floating-basket';
import { MobileBlocker } from './mobile-blocker';
import { useSidebarStore } from '@/lib/store';
import { cn } from '@/lib/utils';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const [basketOpen, setBasketOpen] = useState(false);
  const { isCollapsed } = useSidebarStore();

  return (
    <>
      {/* Mobile blocker - only shown on small screens */}
      <MobileBlocker />

      {/* Main layout - hidden on mobile */}
      <div className="hidden md:block">
        <AppSidebar />
        <div className={cn('min-h-screen transition-all duration-300', isCollapsed ? 'ml-16' : 'ml-60')}>
          <AppHeader onOpenBasket={() => setBasketOpen(true)} />
          <main className="min-h-[calc(100vh-3.5rem)]">{children}</main>
        </div>
        <FloatingBasket open={basketOpen} onOpenChange={setBasketOpen} />
      </div>
    </>
  );
}
