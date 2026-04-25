package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mathlib/server/internal/domain"
)

func readTestBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "parser", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBuildImportPreviewPatternAWithAnswers(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "pattern_a.tex", Content: readTestBytes(t, "pattern_a.tex")},
		{Filename: "pattern_a_answers.tex", Content: readTestBytes(t, "pattern_a_answers.tex")},
	}
	defaults := map[string]any{
		"difficulty": "medium",
		"subject":    "数学",
		"grade":      "高三",
		"source":     "母版导入",
	}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) != 3 {
		t.Fatalf("expected 3 parsed problems, got %d", len(result.Parsed))
	}

	// Problem 1: multiple choice with tasks env
	p1 := result.Parsed[0]
	if p1.Latex == "" {
		t.Error("problem 1 latex should not be empty")
	}
	if !strings.Contains(p1.Latex, `\begin{tasks}`) {
		t.Error("problem 1 body should contain tasks environment")
	}
	if p1.InferredType != domain.ProblemTypeMultipleChoice {
		t.Errorf("problem 1 inferredType: want multiple_choice, got %q", p1.InferredType)
	}
	if p1.Difficulty != domain.DifficultyMedium {
		t.Errorf("problem 1 difficulty: want medium, got %q", p1.Difficulty)
	}
	if p1.Subject == nil || *p1.Subject != "数学" {
		t.Errorf("problem 1 subject: want 数学, got %v", p1.Subject)
	}
	if p1.Grade == nil || *p1.Grade != "高三" {
		t.Errorf("problem 1 grade: want 高三, got %v", p1.Grade)
	}
	if p1.Source == nil || *p1.Source != "母版导入" {
		t.Errorf("problem 1 source: want 母版导入, got %v", p1.Source)
	}
	if p1.AnswerLatex == nil {
		t.Error("problem 1 should have answerLatex from paired answer file")
	} else if !strings.Contains(*p1.AnswerLatex, "C") {
		t.Errorf("problem 1 answer should contain C, got %q", *p1.AnswerLatex)
	}

	// Problem 3: fill blank
	p3 := result.Parsed[2]
	if p3.InferredType != domain.ProblemTypeFillBlank {
		t.Errorf("problem 3 inferredType: want fill_blank, got %q", p3.InferredType)
	}
	if !strings.Contains(p3.Latex, `\underline`) {
		t.Error("problem 3 should contain underline")
	}
	if p3.AnswerLatex == nil {
		t.Error("problem 3 should have answerLatex")
	} else if !strings.Contains(*p3.AnswerLatex, `-2`) {
		t.Errorf("problem 3 answer should contain domain notation, got %q", *p3.AnswerLatex)
	}

	// Status should be success for all
	for i, d := range result.Parsed {
		if d.Status != "success" {
			t.Errorf("problem %d status: want success, got %q", i+1, d.Status)
		}
	}

	// Paired answer file
	if len(result.PairedAnswerFiles) == 0 {
		t.Error("PairedAnswerFiles should be non-empty")
	}

	// Section tags from \section*{集合与不等式}
	hasSectionTag := false
	for _, d := range result.Parsed {
		for _, tag := range d.SectionTags {
			if tag == "集合与不等式" {
				hasSectionTag = true
			}
		}
	}
	if !hasSectionTag {
		t.Error("section tags should contain '集合与不等式'")
	}
}

func TestBuildImportPreviewPatternAWithoutAnswers(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "pattern_a.tex", Content: readTestBytes(t, "pattern_a.tex")},
	}
	defaults := map[string]any{
		"difficulty": "easy",
	}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) != 3 {
		t.Fatalf("expected 3 parsed problems, got %d", len(result.Parsed))
	}

	// No answer file should mean no paired answers
	for i, d := range result.Parsed {
		if d.AnswerLatex != nil {
			t.Errorf("problem %d should not have answerLatex (no answer file provided)", i+1)
		}
	}
	if len(result.PairedAnswerFiles) != 0 {
		t.Errorf("PairedAnswerFiles should be empty, got %v", result.PairedAnswerFiles)
	}

	// Should have a warning about missing answer file
	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(strings.ToLower(w), "未找到") || strings.Contains(strings.ToLower(w), "未匹配") {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected warning about missing answer file")
	}
}

func TestBuildImportPreviewPatternD(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "pattern_d.tex", Content: readTestBytes(t, "pattern_d.tex")},
	}
	defaults := map[string]any{
		"difficulty": "hard",
	}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) != 2 {
		t.Fatalf("expected 2 parsed problems from pattern_d, got %d", len(result.Parsed))
	}

	// First problem: multiple choice (单选题)
	p1 := result.Parsed[0]
	if p1.InferredType != domain.ProblemTypeMultipleChoice {
		t.Errorf("problem 1 inferredType: want multiple_choice, got %q", p1.InferredType)
	}
	if p1.Difficulty != domain.DifficultyHard {
		t.Errorf("problem 1 difficulty: want hard, got %q", p1.Difficulty)
	}

	// Mybox title should be in section tags
	hasTitleTag := false
	for _, tag := range p1.SectionTags {
		if strings.Contains(tag, "集合的交集运算") {
			hasTitleTag = true
		}
	}
	if !hasTitleTag {
		t.Errorf("problem 1 section tags should contain mybox title, got %v", p1.SectionTags)
	}

	// Second problem: fill blank (填空题)
	p2 := result.Parsed[1]
	if p2.InferredType != domain.ProblemTypeFillBlank {
		t.Errorf("problem 2 inferredType: want fill_blank, got %q", p2.InferredType)
	}
}

func TestBuildImportPreviewEmptyFiles(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "edge_empty.tex", Content: readTestBytes(t, "edge_empty.tex")},
	}
	defaults := map[string]any{}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) != 0 {
		t.Errorf("expected 0 parsed problems from empty file, got %d", len(result.Parsed))
	}
}

func TestBuildImportPreviewDefaults(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "pattern_a.tex", Content: readTestBytes(t, "pattern_a.tex")},
	}
	defaults := map[string]any{
		"difficulty": "olympiad",
		"tagNames":   "高考,集合,不等式",
	}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) == 0 {
		t.Fatal("expected at least 1 parsed problem")
	}

	p := result.Parsed[0]
	if p.Difficulty != domain.DifficultyOlympiad {
		t.Errorf("difficulty: want olympiad, got %q", p.Difficulty)
	}
	if len(p.TagNames) != 3 {
		t.Errorf("tagNames: want 3, got %d: %v", len(p.TagNames), p.TagNames)
	}
}

func TestBuildImportPreviewAnswerOnly(t *testing.T) {
	files := []domain.UploadedFile{
		{Filename: "pattern_a_answers.tex", Content: readTestBytes(t, "pattern_a_answers.tex")},
	}
	defaults := map[string]any{}

	result := BuildImportPreview(files, defaults)

	if len(result.Parsed) != 0 {
		t.Errorf("answer-only files should produce 0 problem drafts, got %d", len(result.Parsed))
	}
}
