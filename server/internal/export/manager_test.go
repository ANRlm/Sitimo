package export

import (
	"strings"
	"testing"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
)

func TestNormalizeLatexForExport(t *testing.T) {
	input := "题目选项（\\text{A}）为 \\(60°\\)。\\n1. 当 \\(n=1\\) 时成立。\\n2. 结论成立，且 \\neq 0, \\nabla f(x)=0, \\left[0,1\\right] 上成立。"
	got := normalizeLatexForExport(input)

	if strings.Contains(got, `\n1.`) || strings.Contains(got, `\n2.`) {
		t.Fatalf("expected literal escaped newlines to be normalized, got %q", got)
	}
	if !strings.Contains(got, "\n\n1. 当") || !strings.Contains(got, "\n\n2. 结论") {
		t.Fatalf("expected normalized paragraph breaks, got %q", got)
	}
	if !strings.Contains(got, `\neq 0`) {
		t.Fatalf("expected \\neq command to remain intact, got %q", got)
	}
	if !strings.Contains(got, `\nabla f(x)=0`) {
		t.Fatalf("expected \\nabla command to remain intact, got %q", got)
	}
	if !strings.Contains(got, `\left[0,1\right]`) {
		t.Fatalf("expected \\right command to remain intact, got %q", got)
	}
	if !strings.Contains(got, "题目选项（A）为") {
		t.Fatalf("expected bare \\text command outside math to become plain text, got %q", got)
	}
	if !strings.Contains(got, `\(60^\circ\)`) {
		t.Fatalf("expected degree symbol in math to become ^\\circ, got %q", got)
	}
}

func TestRenderLatexNormalizesAnswerAndSolution(t *testing.T) {
	answer := "第一步\\n第二步"
	solution := "证明如下：\\n1. 先设 \\(a=b\\)。\\n2. 再证 \\neq 0 且 \\nabla f(x)=0，\\left(x+1\\right)^2 > 0。"

	manager := &Manager{}
	rendered, _, err := manager.renderLatex(domain.PaperDetail{
		Paper: domain.Paper{
			Title: "导出规范化测试",
			Layout: domain.PaperLayout{
				Columns:           1,
				FontSize:          12,
				LineHeight:        1.4,
				PaperSize:         "A4",
				ShowAnswerVersion: true,
			},
		},
		ItemDetails: []domain.PaperItemDetail{
			{
				PaperItem: domain.PaperItem{
					Score:      10,
					OrderIndex: 0,
				},
				Problem: &domain.ProblemDetail{
					Problem: domain.Problem{
						Latex:         "已知 \\(f(x)=x^2\\)，求极值。",
						AnswerLatex:   &answer,
						SolutionLatex: &solution,
					},
				},
			},
		},
	}, domain.ExportVariantAnswer)
	if err != nil {
		t.Fatalf("render latex: %v", err)
	}

	if strings.Contains(rendered, `\n1. 先设`) || strings.Contains(rendered, `\n第二步`) {
		t.Fatalf("expected escaped newlines to be normalized in rendered latex, got %q", rendered)
	}
	if !strings.Contains(rendered, "第一步\n\n第二步") {
		t.Fatalf("expected normalized answer paragraphs, got %q", rendered)
	}
	if !strings.Contains(rendered, `\neq 0`) || !strings.Contains(rendered, `\nabla f(x)=0`) {
		t.Fatalf("expected valid latex commands to be preserved, got %q", rendered)
	}
	if !strings.Contains(rendered, `\left(x+1\right)^2 > 0`) {
		t.Fatalf("expected \\right command to be preserved, got %q", rendered)
	}
}

func TestRenderLatexSupportsOverleafDefaultCompiler(t *testing.T) {
	manager := &Manager{}
	rendered, _, err := manager.renderLatex(domain.PaperDetail{
		Paper: domain.Paper{
			Title: "编译兼容测试",
			Layout: domain.PaperLayout{
				Columns:    1,
				FontSize:   12,
				LineHeight: 1.3,
				PaperSize:  "A4",
			},
		},
	}, domain.ExportVariantStudent)
	if err != nil {
		t.Fatalf("render latex: %v", err)
	}

	if !strings.Contains(rendered, `\usepackage{CJKutf8}`) {
		t.Fatalf("expected pdfLaTeX-compatible Chinese fallback, got %q", rendered)
	}
	if !strings.Contains(rendered, `\usepackage[UTF8]{ctex}`) {
		t.Fatalf("expected XeLaTeX Chinese path to remain available, got %q", rendered)
	}
	if strings.Contains(rendered, `MathLib exported bundle requires XeLaTeX`) {
		t.Fatalf("expected latex bundle to avoid hard failure on pdfLaTeX, got %q", rendered)
	}
}

func TestLatexImageFilename(t *testing.T) {
	info := &store.StoredImageInfo{
		ID:          "demo-img5",
		Filename:    "triangle-source.jpeg",
		StoragePath: "original/de/demo-img5.png",
	}

	if got := latexImageFilename(info); got != "demo-img5.png" {
		t.Fatalf("unexpected latex image filename: %q", got)
	}
}
