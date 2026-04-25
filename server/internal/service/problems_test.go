package service_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"
)

func testdataPath(t *testing.T, relative string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", relative)
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read test file %s: %v", path, err)
	}
	return data
}

func TestPreviewBatchImportStructuralParser(t *testing.T) {
	svc := &service.Service{}

	patternA := readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a.tex")))

	input := domain.ImportPreviewRequest{
		Files: []domain.UploadedFile{
			{Filename: "pattern_a.tex", Content: patternA},
		},
		Defaults: map[string]any{"difficulty": "medium", "subject": "数学"},
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) == 0 {
		t.Fatal("Expected parsed problems from structural parser, got 0")
	}

	for i, draft := range result.Parsed {
		if draft.Difficulty == "" {
			t.Errorf("Parsed[%d]: expected non-empty Difficulty", i)
		}
		if draft.Status == "" {
			t.Errorf("Parsed[%d]: expected non-empty Status", i)
		}
		if draft.InferredType == "" && draft.Status == "success" {
			t.Errorf("Parsed[%d]: expected InferredType to be populated for success draft", i)
		}
		if strings.TrimSpace(draft.Latex) == "" && draft.Status == "success" {
			t.Errorf("Parsed[%d]: expected non-empty Latex for success draft", i)
		}
	}
}

func TestPreviewBatchImportStructuralParserWithDefaults(t *testing.T) {
	svc := &service.Service{}

	patternB := readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_b.tex")))

	input := domain.ImportPreviewRequest{
		Files: []domain.UploadedFile{
			{Filename: "pattern_b.tex", Content: patternB},
		},
		Defaults: map[string]any{
			"difficulty": "hard",
			"subject":    "物理",
			"grade":      "高二",
		},
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) == 0 {
		t.Fatal("Expected parsed problems from structural parser, got 0")
	}

	for i, draft := range result.Parsed {
		if draft.Difficulty != domain.DifficultyHard {
			t.Errorf("Parsed[%d]: expected difficulty 'hard', got %q", i, draft.Difficulty)
		}
		if draft.Subject == nil || *draft.Subject != "物理" {
			t.Errorf("Parsed[%d]: expected subject '物理', got %v", i, draft.Subject)
		}
	}
}

func TestPreviewBatchImportLegacyMode(t *testing.T) {
	svc := &service.Service{}

	input := domain.ImportPreviewRequest{
		Latex: `\begin{problem} 求极限 \lim_{x \to 0} \frac{\sin x}{x} \end{problem}
\begin{problem} 求导数 f'(x) = 3x^2 + 2x \end{problem}`,
		Defaults: map[string]any{"difficulty": "easy"},
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) != 2 {
		t.Fatalf("Expected 2 parsed problems from legacy mode, got %d", len(result.Parsed))
	}

	for i, draft := range result.Parsed {
		if draft.InferredType != domain.ProblemTypeSolve {
			t.Errorf("Parsed[%d]: expected InferredType 'solve', got %q", i, draft.InferredType)
		}
		if draft.NeedsReview {
			t.Errorf("Parsed[%d]: expected NeedsReview=false for legacy drafts", i)
		}
		if draft.Status != "success" {
			t.Errorf("Parsed[%d]: expected status 'success', got %q", i, draft.Status)
		}
	}
}

func TestPreviewBatchImportLegacyModeWithCustomSeparators(t *testing.T) {
	svc := &service.Service{}

	input := domain.ImportPreviewRequest{
		Latex:          `\item 第1题：证明勾股定理。\item 第2题：解一元二次方程。`,
		SeparatorStart: `\item`,
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) < 1 {
		t.Fatal("Expected at least 1 parsed problem from single-delimiter mode")
	}
	for _, draft := range result.Parsed {
		if draft.InferredType != domain.ProblemTypeSolve {
			t.Errorf("Expected InferredType 'solve' for legacy draft, got %q", draft.InferredType)
		}
	}
}

func TestPreviewBatchImportErrorNeitherInput(t *testing.T) {
	svc := &service.Service{}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{})

	if len(result.Errors) == 0 {
		t.Fatal("Expected error when neither files nor latex provided")
	}
	if len(result.Parsed) != 0 {
		t.Errorf("Expected no parsed results, got %d", len(result.Parsed))
	}
}

func TestPreviewBatchImportBothInputsPriority(t *testing.T) {
	svc := &service.Service{}

	patternA := readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a.tex")))

	input := domain.ImportPreviewRequest{
		Latex: `\begin{problem} 旧格式题目 \end{problem}`,
		Files: []domain.UploadedFile{
			{Filename: "pattern_a.tex", Content: patternA},
		},
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) == 0 {
		t.Fatal("Expected parsed problems from files (ignoring legacy latex)")
	}

	hasLegacyWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "忽略了 LaTeX 源码") {
			hasLegacyWarning = true
			break
		}
	}
	if !hasLegacyWarning {
		t.Error("Expected warning about ignoring legacy latex when both are provided")
	}

	// Verify structural parser results — should NOT include legacy format draft
	for _, draft := range result.Parsed {
		if strings.Contains(draft.Latex, "旧格式题目") {
			t.Error("Expected legacy latex to be ignored when files are provided")
		}
	}
}

func TestPreviewBatchImportEmptyFiles(t *testing.T) {
	svc := &service.Service{}

	input := domain.ImportPreviewRequest{
		Files: []domain.UploadedFile{
			{Filename: "edge_empty.tex", Content: []byte{}},
		},
	}

	result := svc.PreviewBatchImport(input)

	if len(result.Parsed) == 0 && len(result.Errors) == 0 {
		// Empty file with no parsed results should produce a warning
		hasWarning := false
		for _, w := range result.Warnings {
			if strings.Contains(w, "未解析到任何题目") {
				hasWarning = true
				break
			}
		}
		if !hasWarning {
			t.Error("Expected warning for empty file with no parseable content")
		}
	}
}
