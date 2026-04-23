import { MathText } from '@/components/math-text';
import type { PaperDetail } from '@/lib/types';

type PaperDocumentProps = {
  paper: PaperDetail;
  showAnswers?: boolean;
  scaleClassName?: string;
};

function AnswerSpace({
  blankLines,
  lineHeight,
}: {
  blankLines: number;
  lineHeight: number;
}) {
  if (blankLines <= 0) {
    return null;
  }

  return (
    <div
      className="rounded-2xl border border-dashed border-slate-200 bg-slate-50/80 px-4 py-3"
      style={{ minHeight: `${blankLines * lineHeight}em` }}
    >
      <p className="text-xs font-medium tracking-[0.18em] text-slate-400">作答区</p>
    </div>
  );
}

export function PaperDocument({ paper, showAnswers = false, scaleClassName }: PaperDocumentProps) {
  const items = (paper.itemDetails?.length
    ? paper.itemDetails
    : paper.items.map((item) => ({
        ...item,
        problem: undefined,
      })))
    .slice()
    .sort((left, right) => left.orderIndex - right.orderIndex);

  return (
    <div
      className={`mx-auto my-6 w-full max-w-[210mm] rounded-[28px] border border-slate-200 bg-white p-10 text-slate-900 shadow-xl ${scaleClassName ?? ''}`}
    >
      <header className="border-b border-slate-200 pb-8 text-center">
        {paper.schoolName ? (
          <p className="text-sm tracking-[0.25em] text-slate-500">{paper.schoolName}</p>
        ) : null}
        <h1 className="mt-3 text-3xl font-semibold tracking-[0.08em]">{paper.title}</h1>
        {paper.subtitle ? <p className="mt-2 text-sm text-slate-500">{paper.subtitle}</p> : null}
        <div className="mt-5 grid grid-cols-3 gap-4 text-sm text-slate-600">
          <div className="rounded-full border border-slate-200 px-4 py-2">姓名 __________</div>
          <div className="rounded-full border border-slate-200 px-4 py-2">班级 __________</div>
          <div className="rounded-full border border-slate-200 px-4 py-2">考号 __________</div>
        </div>
        {paper.instructions ? <p className="mt-4 text-sm leading-6 text-slate-500">{paper.instructions}</p> : null}
      </header>

      <section
        className="mt-8 space-y-7"
        style={{
          fontSize: `${paper.layout.fontSize}pt`,
          lineHeight: String(paper.layout.lineHeight),
        }}
      >
        {items.map((item, index) => {
          const problem = item.problem;
          const image = problem?.images?.[0];
          const imagePosition = item.imagePosition ?? 'below';
          const blankLines = item.blankLines ?? 0;

          return (
            <article key={item.id} className="group rounded-3xl border border-slate-200 px-5 py-4 transition-colors hover:border-emerald-300">
              <div className="flex items-start gap-4">
                <div className="mt-1 flex shrink-0 items-center gap-2 text-base font-semibold">
                  <span>{index + 1}.</span>
                  <span className="rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-700">
                    {item.score} 分
                  </span>
                </div>
                <div className="min-w-0 flex-1 space-y-4">
                  {problem ? (
                    image && imagePosition === 'right' ? (
                      <>
                        <div className="grid gap-4 md:grid-cols-[1fr_180px]">
                          <MathText latex={problem.latex} className="leading-8" />
                          <img
                            src={image.url}
                            alt={image.description ?? image.filename}
                            className="max-h-48 rounded-2xl border border-slate-200 object-contain"
                          />
                        </div>

                        <AnswerSpace blankLines={blankLines} lineHeight={paper.layout.lineHeight} />

                        {showAnswers && problem.answerLatex ? (
                          <div className="rounded-2xl border border-emerald-200 bg-emerald-50/80 px-4 py-3 text-sm">
                            <p className="mb-2 font-medium text-emerald-800">答案</p>
                            <MathText latex={problem.answerLatex} />
                          </div>
                        ) : null}
                      </>
                    ) : (
                      <>
                        <MathText latex={problem.latex} className="leading-8" />

                        {image && imagePosition === 'inline' ? (
                          <img
                            src={image.url}
                            alt={image.description ?? image.filename}
                            className="max-h-56 rounded-2xl border border-slate-200 object-contain"
                          />
                        ) : null}

                        {image && imagePosition === 'below' ? (
                          <img
                            src={image.url}
                            alt={image.description ?? image.filename}
                            className="max-h-64 rounded-2xl border border-slate-200 object-contain"
                          />
                        ) : null}

                        <AnswerSpace blankLines={blankLines} lineHeight={paper.layout.lineHeight} />

                        {showAnswers && problem.answerLatex ? (
                          <div className="rounded-2xl border border-emerald-200 bg-emerald-50/80 px-4 py-3 text-sm">
                            <p className="mb-2 font-medium text-emerald-800">答案</p>
                            <MathText latex={problem.answerLatex} />
                          </div>
                        ) : null}
                      </>
                    )
                  ) : (
                    <div className="rounded-2xl border border-dashed p-4 text-sm text-slate-500">
                      题目数据缺失，导出时请先重新保存试卷。
                    </div>
                  )}
                </div>
              </div>
            </article>
          );
        })}
      </section>

      {paper.footerText ? (
        <footer className="mt-10 border-t border-slate-200 pt-4 text-center text-sm text-slate-500">
          {paper.footerText}
        </footer>
      ) : null}
    </div>
  );
}
