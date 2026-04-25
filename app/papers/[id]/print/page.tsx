'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { ArrowLeft, Printer } from 'lucide-react';
import { notFound } from 'next/navigation';
import { PaperDocument } from '@/components/paper-document';
import { Button } from '@/components/ui/button';
import { usePaper } from '@/lib/hooks/use-papers';

export default function PrintPreviewPage({ params }: { params: { id: string } }) {
  const { id } = params;
  const paperQuery = usePaper(id);

  useEffect(() => {
    document.title = '打印预览';
    return () => {
      document.title = 'Sitimo';
    };
  }, []);

  if (!paperQuery.isLoading && !paperQuery.data) {
    notFound();
  }

  if (!paperQuery.data) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-sm text-muted-foreground">正在加载试卷...</div>
      </div>
    );
  }

  const paper = paperQuery.data;

  const handlePrint = () => {
    window.print();
  };

  return (
    <>
      <style jsx global>{`
        @media print {
          body {
            background: white !important;
          }

          .no-print {
            display: none !important;
          }

          .print-container {
            padding: 0 !important;
            background: white !important;
          }

          .paper-document {
            box-shadow: none !important;
            border: none !important;
            border-radius: 0 !important;
            margin: 0 !important;
            padding: 20mm !important;
            max-width: none !important;
          }
        }
      `}</style>

      <div className="min-h-screen bg-slate-100/50 print:bg-white">
        <header className="no-print sticky top-0 z-10 border-b bg-white/95 px-6 py-4 backdrop-blur supports-[backdrop-filter]:bg-white/80">
          <div className="mx-auto flex max-w-[210mm] items-center justify-between">
            <Button variant="ghost" size="sm" asChild>
              <Link href={`/papers/${id}`}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回
              </Link>
            </Button>
            <div className="flex items-center gap-3">
              <h1 className="text-lg font-medium">{paper.title}</h1>
            </div>
            <Button onClick={handlePrint}>
              <Printer className="mr-2 h-4 w-4" />
              打印
            </Button>
          </div>
        </header>

        <main className="print-container p-6">
          <PaperDocument paper={paper} showAnswers={false} />
        </main>
      </div>
    </>
  );
}