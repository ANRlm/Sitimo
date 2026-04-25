package parser

import (
	"strings"
	"testing"
)

func TestExtractAnswersFromPatternA(t *testing.T) {
	input := readTestFile(t, "pattern_a_answers.tex")
	blocks := ScanBlocks(input)
	entries, warnings := ExtractAnswers(blocks, 3)

	if len(entries) != 3 {
		t.Fatalf("Expected 3 answer entries, got %d", len(entries))
	}
	if len(warnings) > 0 {
		t.Logf("warnings (acceptable): %v", warnings)
	}
	if entries[0].AnswerLatex == "" {
		t.Error("Problem 1 should have an answer")
	}
	if entries[0].SolutionLatex == "" {
		t.Error("Problem 1 should have a solution")
	}

	if !strings.Contains(entries[0].AnswerLatex, "C") {
		t.Errorf("Problem 1 answer should contain 'C', got %q", entries[0].AnswerLatex)
	}
	if !strings.Contains(entries[0].SolutionLatex, `x^2 - 3x - 4`) {
		t.Errorf("Problem 1 solution should contain math, got %q", entries[0].SolutionLatex)
	}

	for i, e := range entries {
		if e.Index != i {
			t.Errorf("Entry %d: expected Index=%d, got %d", i, i, e.Index)
		}
	}
}

func TestExtractAnswersFromPatternD(t *testing.T) {
	input := readTestFile(t, "pattern_d_answers.tex")
	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 2)

	if len(entries) != 2 {
		t.Fatalf("Expected 2 answer entries, got %d", len(entries))
	}

	if entries[0].AnswerLatex == "" {
		t.Error("Problem 1 should have an answer")
	}
	if entries[0].SolutionLatex == "" {
		t.Error("Problem 1 should have a solution")
	}

	if entries[1].AnswerLatex == "" {
		t.Error("Problem 2 should have an answer")
	}
	if entries[1].SolutionLatex == "" {
		t.Error("Problem 2 should have a solution")
	}
}

func TestExtractAnswersCountMismatch(t *testing.T) {
	input := readTestFile(t, "pattern_a_answers.tex")
	blocks := ScanBlocks(input)
	entries, warnings := ExtractAnswers(blocks, 5)

	if len(warnings) == 0 {
		t.Error("Expected warning about count mismatch")
	}
	if len(entries) == 0 {
		t.Error("Should return partial results even with mismatch")
	}
}

func TestExtractAnswersEmptyFile(t *testing.T) {
	blocks := []Block{}
	entries, warnings := ExtractAnswers(blocks, 5)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty input, got %d", len(entries))
	}
	if len(warnings) == 0 {
		t.Error("Expected warning about no answers found")
	}
}

func TestExtractAnswersNoEnumerate(t *testing.T) {
	input := "纯文本内容没有 enumerate 环境"
	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from non-enumerate content, got %d", len(entries))
	}
}

func TestExtractAnswersSingleItem(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} A

\textbf{解析：} Simple explanation.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "A" {
		t.Errorf("Expected answer 'A', got %q", entries[0].AnswerLatex)
	}
	if entries[0].SolutionLatex != "Simple explanation." {
		t.Errorf("Expected solution 'Simple explanation.', got %q", entries[0].SolutionLatex)
	}
}

func TestExtractAnswersSolutionOnly(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{解析：} Explanation without explicit answer.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "" {
		t.Errorf("Expected empty answer, got %q", entries[0].AnswerLatex)
	}
	if entries[0].SolutionLatex == "" {
		t.Error("Expected non-empty solution")
	}
}

func TestExtractAnswersAnswerOnly(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} B
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "B" {
		t.Errorf("Expected answer 'B', got %q", entries[0].AnswerLatex)
	}
	if entries[0].SolutionLatex != "" {
		t.Errorf("Expected empty solution, got %q", entries[0].SolutionLatex)
	}
}

func TestExtractAnswersMultiLineSolution(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} C

\textbf{解析：} Line one.
Line two.
Line three.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if !strings.Contains(entries[0].SolutionLatex, "Line one") {
		t.Errorf("Solution should contain 'Line one', got %q", entries[0].SolutionLatex)
	}
	if !strings.Contains(entries[0].SolutionLatex, "Line three") {
		t.Errorf("Solution should contain 'Line three', got %q", entries[0].SolutionLatex)
	}
}

func TestExtractAnswersMultipleEnumerateBlocks(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} A

\textbf{解析：} First explanation.
\end{enumerate}

Some text between.

\begin{enumerate}[resume]
\item \textbf{答案：} B

\textbf{解析：} Second explanation.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 2)

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries across multiple enumerate blocks, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "A" {
		t.Errorf("Entry 0 answer: expected 'A', got %q", entries[0].AnswerLatex)
	}
	if entries[1].AnswerLatex != "B" {
		t.Errorf("Entry 1 answer: expected 'B', got %q", entries[1].AnswerLatex)
	}
}

func TestExtractAnswersAltSolutionMarker(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} D

\textbf{解：} This uses the short marker.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "D" {
		t.Errorf("Expected answer 'D', got %q", entries[0].AnswerLatex)
	}
	if !strings.Contains(entries[0].SolutionLatex, "short marker") {
		t.Errorf("Solution should contain 'short marker', got %q", entries[0].SolutionLatex)
	}
}

func TestExtractAnswersItemWithContentPrefix(t *testing.T) {
	input := `\begin{enumerate}
\item 题目：已知集合

\textbf{答案：} A

\textbf{解析：} Explanation here.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].AnswerLatex != "A" {
		t.Errorf("Expected answer 'A', got %q", entries[0].AnswerLatex)
	}
}

func TestExtractAnswersNestedEnumerateIgnored(t *testing.T) {
	input := `\begin{enumerate}
\item \textbf{答案：} 42

\textbf{解析：} Outer explanation.
\begin{enumerate}
\item Inner item — should be ignored.
\end{enumerate}
More outer text.
\end{enumerate}`

	blocks := ScanBlocks(input)
	entries, _ := ExtractAnswers(blocks, 1)

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry (inner enumerate ignored), got %d", len(entries))
	}
	if entries[0].AnswerLatex != "42" {
		t.Errorf("Expected answer '42', got %q", entries[0].AnswerLatex)
	}
}
