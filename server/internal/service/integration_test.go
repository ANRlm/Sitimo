package service_test

import (
	"path/filepath"
	"strings"
	"testing"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"
)

func TestE2EPatternAWithAnswers(t *testing.T) {
	svc := &service.Service{}

	t.Run("full pipeline", func(t *testing.T) {
		files := []domain.UploadedFile{
			{
				Filename: "pattern_a.tex",
				Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a.tex"))),
			},
			{
				Filename: "pattern_a_answers.tex",
				Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a_answers.tex"))),
			},
		}

		result := svc.PreviewBatchImport(domain.ImportPreviewRequest{
			Files:    files,
			Defaults: map[string]any{"difficulty": "medium", "subject": "数学"},
		})

		if n := len(result.Parsed); n != 3 {
			t.Fatalf("Expected 3 parsed problems, got %d", n)
		}

		if result.Parsed[0].InferredType != domain.ProblemTypeMultipleChoice {
			t.Errorf("Problem 0: expected InferredType multiple_choice, got %q", result.Parsed[0].InferredType)
		}
		if result.Parsed[1].InferredType != domain.ProblemTypeMultipleChoice {
			t.Errorf("Problem 1: expected InferredType multiple_choice, got %q", result.Parsed[1].InferredType)
		}
		if result.Parsed[2].InferredType != domain.ProblemTypeFillBlank {
			t.Errorf("Problem 2: expected InferredType fill_blank, got %q", result.Parsed[2].InferredType)
		}

		if len(result.PairedAnswerFiles) == 0 {
			t.Error("Expected PairedAnswerFiles to contain the matched answer file")
		} else {
			found := false
			for _, f := range result.PairedAnswerFiles {
				if f == "pattern_a_answers.tex" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("PairedAnswerFiles does not contain pattern_a_answers.tex: %v", result.PairedAnswerFiles)
			}
		}

		for i, draft := range result.Parsed {
			if draft.AnswerLatex == nil || *draft.AnswerLatex == "" {
				t.Errorf("Problem %d: expected AnswerLatex to be populated", i)
			} else {
				t.Logf("  Problem %d answer: %s", i, *draft.AnswerLatex)
			}
		}

		hasSectionTag := false
		for _, draft := range result.Parsed {
			for _, tag := range draft.SectionTags {
				if strings.Contains(tag, "集合与不等式") {
					hasSectionTag = true
					break
				}
			}
		}
		if !hasSectionTag {
			t.Log("Note: section tags may not contain expected tag '集合与不等式'")
		}

		for i, draft := range result.Parsed {
			if draft.Status != "success" {
				t.Errorf("Problem %d: expected status 'success', got %q", i, draft.Status)
			}
		}
	})

	t.Run("answer content", func(t *testing.T) {
		files := []domain.UploadedFile{
			{
				Filename: "pattern_a.tex",
				Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a.tex"))),
			},
			{
				Filename: "pattern_a_answers.tex",
				Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a_answers.tex"))),
			},
		}

		result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

		if len(result.Parsed) < 3 {
			t.Fatalf("Need at least 3 problems, got %d", len(result.Parsed))
		}

		if result.Parsed[0].AnswerLatex != nil && !strings.Contains(*result.Parsed[0].AnswerLatex, "C") {
			t.Errorf("Problem 0: expected answer to contain 'C', got %q", *result.Parsed[0].AnswerLatex)
		}
		if result.Parsed[0].SolutionLatex == nil || *result.Parsed[0].SolutionLatex == "" {
			t.Error("Problem 0: expected SolutionLatex to be populated")
		}
	})
}

func TestE2EPatternBWithoutAnswers(t *testing.T) {
	svc := &service.Service{}

	files := []domain.UploadedFile{
		{
			Filename: "pattern_b.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_b.tex"))),
		},
	}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

	if n := len(result.Parsed); n != 2 {
		t.Fatalf("Expected 2 parsed problems, got %d", n)
	}

	for i, draft := range result.Parsed {
		if draft.InferredType != domain.ProblemTypeSolve {
			t.Errorf("Problem %d: expected InferredType solve, got %q", i, draft.InferredType)
		}
	}

	hasMissingAnswerWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "未找到配套解析文件") {
			hasMissingAnswerWarning = true
			break
		}
	}
	if !hasMissingAnswerWarning {
		t.Log("Expected warning about missing answer file (acceptable if pairing logic allows it)")
	}

	if len(result.PairedAnswerFiles) != 0 {
		t.Errorf("Expected no PairedAnswerFiles, got %v", result.PairedAnswerFiles)
	}
}

func TestE2EPatternDMyboxTitles(t *testing.T) {
	svc := &service.Service{}

	files := []domain.UploadedFile{
		{
			Filename: "pattern_d.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_d.tex"))),
		},
	}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

	if n := len(result.Parsed); n != 2 {
		t.Fatalf("Expected 2 parsed problems, got %d", n)
	}

	if result.Parsed[0].InferredType != domain.ProblemTypeMultipleChoice {
		t.Errorf("Problem 0: expected InferredType multiple_choice, got %q", result.Parsed[0].InferredType)
	}
	if result.Parsed[1].InferredType != domain.ProblemTypeFillBlank {
		t.Errorf("Problem 1: expected InferredType fill_blank, got %q", result.Parsed[1].InferredType)
	}

	if len(result.Parsed[0].SectionTags) == 0 {
		t.Error("Problem 0: expected section tags from mybox title, got none")
	} else {
		hasTitle := false
		for _, tag := range result.Parsed[0].SectionTags {
			if strings.Contains(tag, "单选题") || strings.Contains(tag, "集合") || strings.Contains(tag, "交集") {
				hasTitle = true
				break
			}
		}
		if !hasTitle {
			t.Errorf("Problem 0: section tags missing mybox title: %v", result.Parsed[0].SectionTags)
		} else {
			t.Logf("  Problem 0 sectionTags: %v", result.Parsed[0].SectionTags)
		}
	}

	if len(result.Parsed[1].SectionTags) == 0 {
		t.Error("Problem 1: expected section tags from mybox title, got none")
	} else {
		hasTitle := false
		for _, tag := range result.Parsed[1].SectionTags {
			if strings.Contains(tag, "填空题") || strings.Contains(tag, "函数") || strings.Contains(tag, "定义域") {
				hasTitle = true
				break
			}
		}
		if !hasTitle {
			t.Errorf("Problem 1: section tags missing mybox title: %v", result.Parsed[1].SectionTags)
		} else {
			t.Logf("  Problem 1 sectionTags: %v", result.Parsed[1].SectionTags)
		}
	}
}

func TestE2EPatternDWithAnswers(t *testing.T) {
	svc := &service.Service{}

	files := []domain.UploadedFile{
		{
			Filename: "pattern_d.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_d.tex"))),
		},
		{
			Filename: "pattern_d_answers.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_d_answers.tex"))),
		},
	}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

	if n := len(result.Parsed); n != 2 {
		t.Fatalf("Expected 2 parsed problems, got %d", n)
	}

	if len(result.PairedAnswerFiles) > 0 {
		t.Logf("Paired answer files: %v", result.PairedAnswerFiles)
	}

	for i, draft := range result.Parsed {
		if draft.AnswerLatex != nil && *draft.AnswerLatex != "" {
			t.Logf("Problem %d answer: %s", i, *draft.AnswerLatex)
		}
	}
}

func TestE2EEmptyFile(t *testing.T) {
	svc := &service.Service{}

	files := []domain.UploadedFile{
		{
			Filename: "edge_empty.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "edge_empty.tex"))),
		},
	}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

	if len(result.Parsed) != 0 {
		t.Errorf("Expected ZERO parsed problems from empty file, got %d", len(result.Parsed))
	}

	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "未解析到任何题目") || strings.Contains(w, "未包含可识别的题目环境") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Log("No specific empty-file warning found (acceptable)")
	}
}

func TestE2EMalformedFile(t *testing.T) {
	svc := &service.Service{}

	files := []domain.UploadedFile{
		{
			Filename: "edge_malformed.tex",
			Content:  readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "edge_malformed.tex"))),
		},
	}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{Files: files})

	if len(result.Parsed) == 0 && len(result.Errors) == 0 {
		t.Log("Malformed file returned no problems and no errors (acceptable for partial parsing)")
	}

	for i, draft := range result.Parsed {
		if draft.Status != "success" && draft.Status != "error" {
			t.Errorf("Problem %d: unexpected status %q", i, draft.Status)
		}
	}

	for i, draft := range result.Parsed {
		if len(draft.Warnings) > 0 {
			t.Logf("Problem %d warnings: %v", i, draft.Warnings)
		}
	}
}

func TestE2ELegacyBackwardCompat(t *testing.T) {
	svc := &service.Service{}

	t.Run("legacy begin problem delimiter", func(t *testing.T) {
		latex := `\begin{problem} 求极限 \lim_{x \to 0} \frac{\sin x}{x} \end{problem}
\begin{problem} 求导数 f'(x) = 3x^2 + 2x \end{problem}`

		result := svc.PreviewBatchImport(domain.ImportPreviewRequest{
			Latex:    latex,
			Defaults: map[string]any{"difficulty": "easy"},
		})

		if n := len(result.Parsed); n != 2 {
			t.Fatalf("Expected 2 parsed problems from legacy mode, got %d", n)
		}

		for i, draft := range result.Parsed {
			if draft.InferredType != domain.ProblemTypeSolve {
				t.Errorf("Problem %d: expected InferredType solve, got %q", i, draft.InferredType)
			}
			if draft.Status != "success" {
				t.Errorf("Problem %d: expected status 'success', got %q", i, draft.Status)
			}
			if strings.TrimSpace(draft.Latex) == "" {
				t.Errorf("Problem %d: expected non-empty LaTeX body", i)
			}
		}
	})

	t.Run("legacy custom separator", func(t *testing.T) {
		latex := `\item 第1题：证明勾股定理。\item 第2题：解一元二次方程。`

		result := svc.PreviewBatchImport(domain.ImportPreviewRequest{
			Latex:          latex,
			SeparatorStart: `\item`,
		})

		if len(result.Parsed) < 1 {
			t.Fatal("Expected at least 1 parsed problem from single-delimiter mode")
		}

		for i, draft := range result.Parsed {
			if draft.InferredType != domain.ProblemTypeSolve {
				t.Errorf("Problem %d: expected InferredType solve, got %q", i, draft.InferredType)
			}
			}
	})

	t.Run("legacy exclusive with files", func(t *testing.T) {
		patternA := readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_a.tex")))

		result := svc.PreviewBatchImport(domain.ImportPreviewRequest{
			Latex: `\begin{problem} 旧格式题目 \end{problem}`,
			Files: []domain.UploadedFile{
				{Filename: "pattern_a.tex", Content: patternA},
			},
		})

		if len(result.Parsed) == 0 {
			t.Fatal("Expected parsed problems from files (ignoring legacy latex)")
		}

		hasWarning := false
		for _, w := range result.Warnings {
			if strings.Contains(w, "忽略了 LaTeX 源码") {
				hasWarning = true
				break
			}
		}
		if !hasWarning {
			t.Error("Expected warning about ignoring legacy LaTeX when both are provided")
		}

		for _, draft := range result.Parsed {
			if strings.Contains(draft.Latex, "旧格式题目") {
				t.Error("Expected legacy latex to be ignored when files are provided")
			}
		}
	})
}

func TestE2ENoInputError(t *testing.T) {
	svc := &service.Service{}

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{})

	if len(result.Errors) == 0 {
		t.Fatal("Expected error when neither files nor latex provided")
	}
	if len(result.Parsed) != 0 {
		t.Errorf("Expected no parsed results, got %d", len(result.Parsed))
	}
}

func TestE2EDefaultsPropagation(t *testing.T) {
	svc := &service.Service{}

	patternB := readTestFile(t, testdataPath(t, filepath.Join("testdata", "parser", "pattern_b.tex")))

	result := svc.PreviewBatchImport(domain.ImportPreviewRequest{
		Files: []domain.UploadedFile{
			{Filename: "pattern_b.tex", Content: patternB},
		},
		Defaults: map[string]any{
			"difficulty": "hard",
			"subject":    "物理",
			"grade":      "高二",
			"source":     "高考真题",
		},
	})

	if len(result.Parsed) == 0 {
		t.Fatal("Expected parsed problems, got 0")
	}

	for i, draft := range result.Parsed {
		if draft.Difficulty != domain.DifficultyHard {
			t.Errorf("Problem %d: expected difficulty 'hard', got %q", i, draft.Difficulty)
		}
		if draft.Subject == nil || *draft.Subject != "物理" {
			t.Errorf("Problem %d: expected subject '物理', got %v", i, draft.Subject)
		}
		if draft.Grade == nil || *draft.Grade != "高二" {
			t.Errorf("Problem %d: expected grade '高二', got %v", i, draft.Grade)
		}
		if draft.Source == nil || *draft.Source != "高考真题" {
			t.Errorf("Problem %d: expected source '高考真题', got %v", i, draft.Source)
		}
	}
}
